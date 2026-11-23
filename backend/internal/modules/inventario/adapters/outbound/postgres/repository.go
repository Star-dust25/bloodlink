package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bloodlink/backend/internal/models"
	"github.com/bloodlink/backend/internal/modules/inventario/domain"
	"github.com/bloodlink/backend/internal/modules/inventario/ports"
	"gorm.io/gorm"
)

type InventarioRepository struct {
	db *gorm.DB
}

func NewInventarioRepository(db *gorm.DB) *InventarioRepository {
	return &InventarioRepository{db: db}
}

func (r *InventarioRepository) ListarBancos(ctx context.Context, hospitalID string) ([]domain.BancoSangre, error) {
	var orms []models.BancoSangreORM
	q := r.db.WithContext(ctx).Model(&models.BancoSangreORM{})
	if hospitalID != "" {
		q = q.Where("id = ?", hospitalID)
	}
	if err := q.Order("id ASC").Find(&orms).Error; err != nil {
		return nil, err
	}

	res := make([]domain.BancoSangre, 0, len(orms))
	for _, o := range orms {
		res = append(res, domain.BancoSangre{
			ID:              o.ID,
			Nombre:          o.Nombre,
			Direccion:       o.Direccion,
			Latitud:         o.Latitud,
			Longitud:        o.Longitud,
			Coordenadas:     o.Coordenadas,
			Jurisdiccion:    o.Jurisdiccion,
			TipoInstitucion: o.TipoInstitucion,
			EsLegacySoap:    o.EsLegacySoap,
		})
	}
	return res, nil
}

func (r *InventarioRepository) ObtenerBancoPorID(ctx context.Context, bancoID string) (*domain.BancoSangre, error) {
	var o models.BancoSangreORM
	if err := r.db.WithContext(ctx).Where("id = ?", bancoID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBancoNoEncontrado
		}
		return nil, err
	}
	return &domain.BancoSangre{
		ID:              o.ID,
		Nombre:          o.Nombre,
		Direccion:       o.Direccion,
		Latitud:         o.Latitud,
		Longitud:        o.Longitud,
		Coordenadas:     o.Coordenadas,
		Jurisdiccion:    o.Jurisdiccion,
		TipoInstitucion: o.TipoInstitucion,
		EsLegacySoap:    o.EsLegacySoap,
	}, nil
}

func (r *InventarioRepository) ListarUnidades(ctx context.Context, filtro ports.FiltroStock) ([]domain.UnidadSangre, error) {
	var orms []models.UnidadSangreORM
	q := r.db.WithContext(ctx).Model(&models.UnidadSangreORM{})
	if filtro.HospitalID != "" {
		q = q.Where("banco_id = ?", filtro.HospitalID)
	}
	if filtro.TipoSangre != "" {
		q = q.Where("tipo_sangre = ?", filtro.TipoSangre)
	}
	if filtro.FactorRh != "" {
		q = q.Where("factor_rh = ?", filtro.FactorRh)
	}
	if filtro.Estado != "" {
		q = q.Where("estado = ?", filtro.Estado)
	}

	if err := q.Order("fecha_caducidad ASC").Find(&orms).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	res := make([]domain.UnidadSangre, 0, len(orms))
	for _, o := range orms {
		u := toDomainUnidad(&o, now)
		res = append(res, *u)
	}
	return res, nil
}

func (r *InventarioRepository) ObtenerUnidadPorID(ctx context.Context, unidadID string) (*domain.UnidadSangre, error) {
	var o models.UnidadSangreORM
	if err := r.db.WithContext(ctx).Where("id = ?", unidadID).First(&o).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUnidadNoEncontrada
		}
		return nil, err
	}
	return toDomainUnidad(&o, time.Now()), nil
}

func (r *InventarioRepository) CrearUnidadConAuditoria(ctx context.Context, u *domain.UnidadSangre, usuarioID, detalle string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orm := models.UnidadSangreORM{
			ID:              u.ID,
			CodigoBarras:    u.CodigoBarras,
			BancoID:         u.BancoID,
			TipoSangre:      u.TipoSangre,
			FactorRh:        u.FactorRh,
			VolumenMl:       u.VolumenMl,
			Origen:          u.Origen,
			FechaExtraccion: u.FechaExtraccion,
			FechaCaducidad:  u.FechaCaducidad,
			Estado:          u.Estado,
		}
		if err := tx.Create(&orm).Error; err != nil {
			return err
		}

		audit := models.AuditLogORM{
			Entidad:        "UnidadSangre",
			EntidadID:      u.ID,
			Accion:         "REGISTRAR_INGRESO_UNIDAD",
			EstadoAnterior: "NINGUNO",
			EstadoNuevo:    u.Estado,
			UsuarioID:      usuarioID,
			Detalle:        detalle,
			Timestamp:      time.Now(),
		}
		return tx.Create(&audit).Error
	})
}

func (r *InventarioRepository) ActualizarEstadoConAuditoria(ctx context.Context, u *domain.UnidadSangre, estadoAnterior, usuarioID, motivo string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.UnidadSangreORM{}).
			Where("id = ?", u.ID).
			Updates(map[string]any{
				"estado":     u.Estado,
				"updated_at": time.Now(),
			}).Error; err != nil {
			return err
		}

		detalle := fmt.Sprintf("Transición %s -> %s", estadoAnterior, u.Estado)
		if motivo != "" {
			detalle = fmt.Sprintf("%s. Motivo: %s", detalle, motivo)
		}
		audit := models.AuditLogORM{
			Entidad:        "UnidadSangre",
			EntidadID:      u.ID,
			Accion:         "ACTUALIZAR_ESTADO_UNIDAD",
			EstadoAnterior: estadoAnterior,
			EstadoNuevo:    u.Estado,
			UsuarioID:      usuarioID,
			Detalle:        detalle,
			Timestamp:      time.Now(),
		}
		return tx.Create(&audit).Error
	})
}

func toDomainUnidad(o *models.UnidadSangreORM, now time.Time) *domain.UnidadSangre {
	u := &domain.UnidadSangre{
		ID:              o.ID,
		CodigoBarras:    o.CodigoBarras,
		BancoID:         o.BancoID,
		TipoSangre:      o.TipoSangre,
		FactorRh:        o.FactorRh,
		VolumenMl:       o.VolumenMl,
		Origen:          o.Origen,
		FechaExtraccion: o.FechaExtraccion,
		FechaCaducidad:  o.FechaCaducidad,
		Estado:          o.Estado,
	}
	u.AlertaCaducidad = u.VerificarCaducidad(now)
	return u
}
