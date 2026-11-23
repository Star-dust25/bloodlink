package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/bloodlink/backend/internal/events"
	"github.com/bloodlink/backend/internal/modules/notificaciones/domain"
)

type mockNotifRepo struct {
	mu    sync.Mutex
	items []domain.Notificacion
}

func (m *mockNotifRepo) GuardarNotificacion(_ context.Context, n *domain.Notificacion) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items = append(m.items, *n)
	return nil
}

func (m *mockNotifRepo) ListarNotificaciones(_ context.Context, _ string, _ int) ([]domain.Notificacion, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]domain.Notificacion, len(m.items))
	copy(res, m.items)
	return res, nil
}

func (m *mockNotifRepo) BuscarDonantesCompatibles(_ context.Context, _, _ string) ([]domain.DonanteConvocable, error) {
	return []domain.DonanteConvocable{
		{DonanteID: "DON-2026-4401", NombreCompleto: "MARIA QUISPE", Correo: "m@q.pe", TipoSangre: "O", FactorRh: "+", BancoPreferidoID: "HOSP-MINSA-01"},
	}, nil
}

func (m *mockNotifRepo) BuscarUnidadesProximasACaducar7Dias(_ context.Context) ([]domain.UnidadPorCaducar, error) {
	return []domain.UnidadPorCaducar{
		{UnidadID: "UND-2026-0106", CodigoBarras: "BL-MINSA01-0106", BancoID: "HOSP-MINSA-01", TipoSangre: "O", FactorRh: "-", FechaCaducidad: "2026-06-19", DiasRestantes: 4, AlertaCaducidad: "AMARILLO"},
	}, nil
}

func TestSrvNotificaciones_EventosYJobCaducidad(t *testing.T) {
	repo := &mockNotifRepo{}
	bus := events.NewInMemoryEventBus(64, 2)
	srv := NewSrvNotificaciones(repo, bus)
	srv.RegistrarSuscripcionesEventos()

	ctx := context.Background()

	// 1. Ejecutar job de caducidad a 7 días (RF-08)
	jobResp, err := srv.EjecutarJobVerificacionCaducidad(ctx)
	if err != nil || jobResp.TotalDetectadas != 1 {
		t.Fatalf("EjecutarJobVerificacionCaducidad error: %v", err)
	}

	// 2. Publicar todos los eventos de dominio soportados
	bus.Publish(ctx, events.EventUnidadRegistrada, map[string]any{"codigoBarras": "BL-1", "tipoSangre": "O", "factorRh": "+", "bancoId": "HOSP-MINSA-01"})
	bus.Publish(ctx, events.EventEstadoUnidadActualizado, map[string]any{"unidadId": "UND-1", "estadoAnterior": "Disponible", "estadoNuevo": "Reservada"})
	bus.Publish(ctx, events.EventEmergenciaReservada, map[string]any{"solicitudId": "SOL-1", "origenAsignacion": "LOCAL", "cantidad": 2, "tipoSangre": "O", "factorRh": "+", "bancoAsignadoId": "HOSP-MINSA-01", "hospitalOrigenId": "HOSP-MINSA-01"})
	bus.Publish(ctx, events.EventSolicitudCancelada, map[string]any{"solicitudId": "SOL-1", "unidadesLiberadas": 2, "bancoAsignadoId": "HOSP-MINSA-01", "motivo": "Cancelada"})
	bus.Publish(ctx, events.EventStockCriticoDetectado, map[string]any{
		"solicitudId": "SOL-2026-0899",
		"tipoSangre":  "O",
		"factorRh":    "+",
	})

	time.Sleep(50 * time.Millisecond)
	bus.Close()

	list, err := srv.ListarNotificaciones(ctx, "", 0)
	if err != nil || len(list) < 7 {
		t.Fatalf("esperaba al menos 7 notificaciones generadas por eventos, obtuvo %d (err=%v)", len(list), err)
	}
}
