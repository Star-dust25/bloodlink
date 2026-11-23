package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bloodlink/backend/internal/events"
	"github.com/bloodlink/backend/internal/modules/inventario/domain"
	"github.com/bloodlink/backend/internal/modules/inventario/ports"
	"github.com/google/uuid"
)

type SrvGestionarInventario struct {
	repo     ports.IInventarioRepository
	eventBus events.IEventBus
}

func NewSrvGestionarInventario(repo ports.IInventarioRepository, eventBus events.IEventBus) *SrvGestionarInventario {
	return &SrvGestionarInventario{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (s *SrvGestionarInventario) ConsultarStockRed(ctx context.Context, filtro ports.FiltroStock) (*ports.StockRedResponseDTO, error) {
	bancos, err := s.repo.ListarBancos(ctx, filtro.HospitalID)
	if err != nil {
		return nil, err
	}

	estadoFiltro := filtro.Estado
	if estadoFiltro == "" {
		estadoFiltro = domain.EstadoUnidadDisponible
	}

	unidades, err := s.repo.ListarUnidades(ctx, ports.FiltroStock{
		TipoSangre: filtro.TipoSangre,
		FactorRh:   filtro.FactorRh,
		Estado:     estadoFiltro,
		HospitalID: filtro.HospitalID,
	})
	if err != nil {
		return nil, err
	}

	now := time.Now()
	bancosResumen := make([]ports.BancoStockResumenDTO, 0, len(bancos))
	stockGlobal := 0

	for _, b := range bancos {
		disp := 0
		porCaducar7d := 0
		for _, u := range unidades {
			if u.BancoID != b.ID {
				continue
			}
			dias := u.DiasParaCaducar(now)
			if dias > 0 {
				disp++
				stockGlobal++
				if dias <= domain.DiasUmbralAlertaAmarilla {
					porCaducar7d++
				}
			}
		}

		bancosResumen = append(bancosResumen, ports.BancoStockResumenDTO{
			BancoID:               b.ID,
			Nombre:                b.Nombre,
			TipoInstitucion:       b.TipoInstitucion,
			Coordenadas:           b.Coordenadas,
			Jurisdiccion:          b.Jurisdiccion,
			EsLegacySoap:          b.EsLegacySoap,
			UnidadesDisponibles:   disp,
			ProximasACaducar7Dias: porCaducar7d,
		})
	}

	tipoLabel := filtro.TipoSangre
	if tipoLabel == "" {
		tipoLabel = "TODOS"
	}
	rhLabel := filtro.FactorRh
	if rhLabel == "" {
		rhLabel = "TODOS"
	}

	return &ports.StockRedResponseDTO{
		TipoSangre:            tipoLabel,
		FactorRh:              rhLabel,
		StockGlobalDisponible: stockGlobal,
		NivelAbastecimiento:   domain.CalcularNivelAbastecimiento(stockGlobal),
		Bancos:                bancosResumen,
		ConsultadoEn:          now.Format(time.RFC3339),
	}, nil
}

func (s *SrvGestionarInventario) ListarUnidades(ctx context.Context, filtro ports.FiltroStock) ([]ports.UnidadSangreResponseDTO, error) {
	unidades, err := s.repo.ListarUnidades(ctx, filtro)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := make([]ports.UnidadSangreResponseDTO, 0, len(unidades))
	for _, u := range unidades {
		result = append(result, toUnidadDTO(&u, now))
	}
	return result, nil
}

func (s *SrvGestionarInventario) ObtenerUnidad(ctx context.Context, unidadID string) (*ports.UnidadSangreResponseDTO, error) {
	u, err := s.repo.ObtenerUnidadPorID(ctx, unidadID)
	if err != nil {
		return nil, err
	}
	dto := toUnidadDTO(u, time.Now())
	return &dto, nil
}

func (s *SrvGestionarInventario) RegistrarUnidad(ctx context.Context, req ports.RegistrarUnidadRequest, usuarioID string) (*ports.UnidadSangreResponseDTO, error) {
	banco, err := s.repo.ObtenerBancoPorID(ctx, req.BancoID)
	if err != nil || banco == nil {
		return nil, domain.ErrBancoNoEncontrado
	}

	fechaExt, err := time.Parse("2006-01-02", strings.TrimSpace(req.FechaExtraccion))
	if err != nil {
		return nil, fmt.Errorf("formato de fechaExtraccion inválido (use YYYY-MM-DD): %w", err)
	}

	now := time.Now()
	shortUUID := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:8])
	unidadID := fmt.Sprintf("UND-%d-%s", now.Year(), shortUUID)
	cleanBank := strings.ReplaceAll(strings.ReplaceAll(req.BancoID, "HOSP-", ""), "-", "")
	codigoBarras := fmt.Sprintf("BL-%s-%d%s", cleanBank, now.Year(), shortUUID)

	unidad, err := domain.NuevaUnidadSangre(
		unidadID,
		codigoBarras,
		req.BancoID,
		strings.ToUpper(strings.TrimSpace(req.TipoSangre)),
		strings.TrimSpace(req.FactorRh),
		strings.TrimSpace(req.Origen),
		req.VolumenMl,
		fechaExt,
		now,
	)
	if err != nil {
		return nil, err
	}

	detalle := fmt.Sprintf("Registro de unidad %s (%s%s) en %s", unidad.CodigoBarras, unidad.TipoSangre, unidad.FactorRh, unidad.BancoID)
	if err := s.repo.CrearUnidadConAuditoria(ctx, unidad, usuarioID, detalle); err != nil {
		return nil, err
	}

	if s.eventBus != nil {
		s.eventBus.Publish(ctx, events.EventUnidadRegistrada, map[string]any{
			"unidadId":        unidad.ID,
			"codigoBarras":    unidad.CodigoBarras,
			"bancoId":         unidad.BancoID,
			"tipoSangre":      unidad.TipoSangre,
			"factorRh":        unidad.FactorRh,
			"fechaCaducidad":  unidad.FechaCaducidad.Format("2006-01-02"),
			"alertaCaducidad": unidad.AlertaCaducidad,
		})
	}

	dto := toUnidadDTO(unidad, now)
	return &dto, nil
}

func (s *SrvGestionarInventario) ActualizarEstado(ctx context.Context, unidadID string, req ports.ActualizarEstadoRequest, usuarioID string) (*ports.UnidadSangreResponseDTO, error) {
	unidad, err := s.repo.ObtenerUnidadPorID(ctx, unidadID)
	if err != nil {
		return nil, err
	}

	estadoAnterior := unidad.Estado
	if err := unidad.ActualizarEstado(strings.TrimSpace(req.NuevoEstado)); err != nil {
		return nil, err
	}

	if err := s.repo.ActualizarEstadoConAuditoria(ctx, unidad, estadoAnterior, usuarioID, req.Motivo); err != nil {
		return nil, err
	}

	if s.eventBus != nil {
		s.eventBus.Publish(ctx, events.EventEstadoUnidadActualizado, map[string]any{
			"unidadId":       unidad.ID,
			"bancoId":        unidad.BancoID,
			"estadoAnterior": estadoAnterior,
			"estadoNuevo":    unidad.Estado,
			"motivo":         req.Motivo,
		})
	}

	dto := toUnidadDTO(unidad, time.Now())
	return &dto, nil
}

func toUnidadDTO(u *domain.UnidadSangre, now time.Time) ports.UnidadSangreResponseDTO {
	return ports.UnidadSangreResponseDTO{
		UnidadID:        u.ID,
		CodigoBarras:    u.CodigoBarras,
		BancoID:         u.BancoID,
		TipoSangre:      u.TipoSangre,
		FactorRh:        u.FactorRh,
		VolumenMl:       u.VolumenMl,
		Origen:          u.Origen,
		FechaExtraccion: u.FechaExtraccion.Format("2006-01-02"),
		FechaCaducidad:  u.FechaCaducidad.Format("2006-01-02"),
		Estado:          u.Estado,
		AlertaCaducidad: u.VerificarCaducidad(now),
		DiasRestantes:   u.DiasParaCaducar(now),
	}
}
