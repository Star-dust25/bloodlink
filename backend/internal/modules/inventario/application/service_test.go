package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/bloodlink/backend/internal/events"
	"github.com/bloodlink/backend/internal/modules/inventario/domain"
	"github.com/bloodlink/backend/internal/modules/inventario/ports"
)

type mockInventarioRepo struct {
	bancos   []domain.BancoSangre
	unidades map[string]*domain.UnidadSangre
	seq      int64
}

func newMockInventarioRepo() *mockInventarioRepo {
	now := time.Now().Truncate(24 * time.Hour)
	return &mockInventarioRepo{
		bancos: []domain.BancoSangre{
			{ID: "HOSP-MINSA-01", Nombre: "Hospital Santa Rosa", TipoInstitucion: "MINSA", Coordenadas: "-5.1945,-80.6328", Jurisdiccion: "Piura Oeste"},
		},
		unidades: map[string]*domain.UnidadSangre{
			"UND-2026-0101": {
				ID:              "UND-2026-0101",
				CodigoBarras:    "BL-MINSA01-20260101",
				BancoID:         "HOSP-MINSA-01",
				TipoSangre:      "O",
				FactorRh:        "-",
				VolumenMl:       450,
				Origen:          "DONACION",
				FechaExtraccion: now.AddDate(0, 0, -38),
				FechaCaducidad:  now.AddDate(0, 0, 4),
				Estado:          domain.EstadoUnidadDisponible,
			},
		},
		seq: 900,
	}
}

func (m *mockInventarioRepo) ListarBancos(_ context.Context, _ string) ([]domain.BancoSangre, error) {
	return m.bancos, nil
}

func (m *mockInventarioRepo) ObtenerBancoPorID(_ context.Context, bancoID string) (*domain.BancoSangre, error) {
	for _, b := range m.bancos {
		if b.ID == bancoID {
			return &b, nil
		}
	}
	return nil, domain.ErrBancoNoEncontrado
}

func (m *mockInventarioRepo) ListarUnidades(_ context.Context, _ ports.FiltroStock) ([]domain.UnidadSangre, error) {
	res := make([]domain.UnidadSangre, 0, len(m.unidades))
	for _, u := range m.unidades {
		res = append(res, *u)
	}
	return res, nil
}

func (m *mockInventarioRepo) ObtenerUnidadPorID(_ context.Context, unidadID string) (*domain.UnidadSangre, error) {
	u, ok := m.unidades[unidadID]
	if !ok {
		return nil, domain.ErrUnidadNoEncontrada
	}
	copyU := *u
	return &copyU, nil
}

func (m *mockInventarioRepo) CrearUnidadConAuditoria(_ context.Context, u *domain.UnidadSangre, _, _ string) error {
	m.unidades[u.ID] = u
	return nil
}

func (m *mockInventarioRepo) ActualizarEstadoConAuditoria(_ context.Context, u *domain.UnidadSangre, _, _, _ string) error {
	m.unidades[u.ID] = u
	return nil
}

func (m *mockInventarioRepo) NextSequenceNumber(_ context.Context) (int64, error) {
	m.seq++
	return m.seq, nil
}

func TestSrvGestionarInventario_FlujosCompletos(t *testing.T) {
	repo := newMockInventarioRepo()
	bus := events.NewInMemoryEventBus(32, 1)
	defer bus.Close()

	srv := NewSrvGestionarInventario(repo, bus)
	ctx := context.Background()

	// 1. Consultar Stock Red
	stock, err := srv.ConsultarStockRed(ctx, ports.FiltroStock{})
	if err != nil {
		t.Fatalf("ConsultarStockRed error: %v", err)
	}
	if stock.StockGlobalDisponible != 1 || len(stock.Bancos) != 1 {
		t.Errorf("StockGlobalDisponible = %d, esperaba 1", stock.StockGlobalDisponible)
	}

	// 2. Registrar Unidad Válida
	ayer := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	creada, err := srv.RegistrarUnidad(ctx, ports.RegistrarUnidadRequest{
		BancoID:         "HOSP-MINSA-01",
		TipoSangre:      "O",
		FactorRh:        "+",
		VolumenMl:       450,
		FechaExtraccion: ayer,
		Origen:          "DONACION_VOLUNTARIA",
	}, "13")
	if err != nil {
		t.Fatalf("RegistrarUnidad válida error: %v", err)
	}
	if creada.Estado != "Disponible" || creada.AlertaCaducidad != "VERDE" {
		t.Errorf("unidad creada con estado/alerta inesperados: %+v", creada)
	}

	// 3. Registrar Unidad Caducada (> 42 días -> ErrUnidadYaCaducada 422)
	hace50Dias := time.Now().AddDate(0, 0, -50).Format("2006-01-02")
	_, err = srv.RegistrarUnidad(ctx, ports.RegistrarUnidadRequest{
		BancoID:         "HOSP-MINSA-01",
		TipoSangre:      "A",
		FactorRh:        "+",
		VolumenMl:       450,
		FechaExtraccion: hace50Dias,
		Origen:          "DONACION_VOLUNTARIA",
	}, "13")
	if !errors.Is(err, domain.ErrUnidadYaCaducada) {
		t.Fatalf("esperaba ErrUnidadYaCaducada, obtuvo %v", err)
	}

	// 4. Actualizar Estado y Obtener Unidad
	act, err := srv.ActualizarEstado(ctx, creada.UnidadID, ports.ActualizarEstadoRequest{
		NuevoEstado: "Reservada",
		Motivo:      "Reserva manual",
	}, "13")
	if err != nil || act.Estado != "Reservada" {
		t.Fatalf("ActualizarEstado error: %v, estado: %v", err, act)
	}

	obt, err := srv.ObtenerUnidad(ctx, creada.UnidadID)
	if err != nil || obt.Estado != "Reservada" {
		t.Fatalf("ObtenerUnidad error: %v", err)
	}

	list, err := srv.ListarUnidades(ctx, ports.FiltroStock{})
	if err != nil || len(list) != 2 {
		t.Fatalf("ListarUnidades esperaba 2, obtuvo %d (%v)", len(list), err)
	}
}
