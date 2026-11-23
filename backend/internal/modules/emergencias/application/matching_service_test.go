package application

import (
	"context"
	"errors"
	"testing"

	"github.com/bloodlink/backend/internal/events"
	"github.com/bloodlink/backend/internal/modules/emergencias/domain"
	"github.com/bloodlink/backend/internal/modules/emergencias/ports"
	integDomain "github.com/bloodlink/backend/internal/modules/integracion/domain"
	integMaps "github.com/bloodlink/backend/internal/modules/integracion/adapters/outbound/maps"
)

type mockEmergenciaRepo struct {
	bancos      map[string]*domain.CandidatoBanco
	stockPorBco map[string]int
	solicitudes map[string]*domain.SolicitudEmergencia
	seq         int64
}

func newMockEmergenciaRepo() *mockEmergenciaRepo {
	return &mockEmergenciaRepo{
		bancos: map[string]*domain.CandidatoBanco{
			"HOSP-MINSA-01": {
				BancoID:      "HOSP-MINSA-01",
				Nombre:       "Hospital Santa Rosa - Piura",
				Coordenadas:  "-5.1945,-80.6328",
				Latitud:      -5.1945,
				Longitud:     -80.6328,
				Jurisdiccion: "Piura Oeste",
				EsLegacySoap: false,
			},
			"HOSP-ESSALUD-04": {
				BancoID:      "HOSP-ESSALUD-04",
				Nombre:       "Hospital Cayetano Heredia - EsSalud",
				Coordenadas:  "-5.1821,-80.6190",
				Latitud:      -5.1821,
				Longitud:     -80.6190,
				Jurisdiccion: "Castilla",
				EsLegacySoap: true,
			},
			"HOSP-MINSA-02": {
				BancoID:      "HOSP-MINSA-02",
				Nombre:       "Hospital Sullana",
				Coordenadas:  "-4.9039,-80.6853",
				Latitud:      -4.9039,
				Longitud:     -80.6853,
				Jurisdiccion: "Piura Oeste",
				EsLegacySoap: false,
			},
		},
		stockPorBco: map[string]int{
			"HOSP-MINSA-01":   0,
			"HOSP-ESSALUD-04": 10,
			"HOSP-MINSA-02":   6,
		},
		solicitudes: make(map[string]*domain.SolicitudEmergencia),
		seq:         800,
	}
}

func (m *mockEmergenciaRepo) ObtenerBanco(_ context.Context, bancoID string) (*domain.CandidatoBanco, error) {
	b, ok := m.bancos[bancoID]
	if !ok {
		return nil, errors.New("no existe")
	}
	return b, nil
}

func (m *mockEmergenciaRepo) ListarCandidatosExternos(_ context.Context, hospitalOrigenID, _, _ string) ([]domain.CandidatoBanco, error) {
	var list []domain.CandidatoBanco
	for id, b := range m.bancos {
		if id == hospitalOrigenID {
			continue
		}
		if m.stockPorBco[id] > 0 {
			c := *b
			c.UnidadesDisponibles = m.stockPorBco[id]
			list = append(list, c)
		}
	}
	return list, nil
}

func (m *mockEmergenciaRepo) IntentarReservaAtomica(_ context.Context, bancoID, _, _ string, cantidad int, _ string) ([]domain.UnidadAsignada, error) {
	if m.stockPorBco[bancoID] < cantidad {
		return nil, domain.ErrStockNoDisponibleEnRed
	}
	m.stockPorBco[bancoID] -= cantidad
	res := make([]domain.UnidadAsignada, 0, cantidad)
	for i := 1; i <= cantidad; i++ {
		res = append(res, domain.UnidadAsignada{
			UnidadID:     "UND-TEST-01",
			CodigoBarras: "BL-TEST-01",
		})
	}
	return res, nil
}

func (m *mockEmergenciaRepo) GuardarSolicitud(_ context.Context, sol *domain.SolicitudEmergencia, _, _ string) error {
	copySol := *sol
	m.solicitudes[sol.ID] = &copySol
	return nil
}

func (m *mockEmergenciaRepo) ObtenerSolicitudPorID(_ context.Context, solicitudID string) (*domain.SolicitudEmergencia, error) {
	s, ok := m.solicitudes[solicitudID]
	if !ok {
		return nil, domain.ErrSolicitudNoEncontrada
	}
	copySol := *s
	return &copySol, nil
}

func (m *mockEmergenciaRepo) ListarSolicitudes(_ context.Context, _ string) ([]domain.SolicitudEmergencia, error) {
	res := make([]domain.SolicitudEmergencia, 0, len(m.solicitudes))
	for _, s := range m.solicitudes {
		res = append(res, *s)
	}
	return res, nil
}

func (m *mockEmergenciaRepo) CancelarSolicitudYLiberarUnidades(_ context.Context, sol *domain.SolicitudEmergencia, _ string) error {
	m.solicitudes[sol.ID] = sol
	return nil
}

func (m *mockEmergenciaRepo) ConfirmarTransfusionYUnidades(_ context.Context, sol *domain.SolicitudEmergencia, _ string) error {
	m.solicitudes[sol.ID] = sol
	return nil
}

func (m *mockEmergenciaRepo) ContarStockDisponibleGlobalPorTipo(_ context.Context, _, _ string) (int, error) {
	return 2, nil // < 5 para disparar también alerta de umbral crítico <10%
}

func (m *mockEmergenciaRepo) NextSequenceNumber(_ context.Context) (int64, error) {
	m.seq++
	return m.seq, nil
}

type mockLegacySoap struct{}

func (m *mockLegacySoap) ConsultarYReservarStockLegacy(_ context.Context, bancoID, _, _ string, cantidad int) (*integDomain.ReservaLegacyResultado, error) {
	unidades := make([]string, 0, cantidad)
	codigos := make([]string, 0, cantidad)
	for i := 1; i <= cantidad; i++ {
		unidades = append(unidades, "UND-2026-0401")
		codigos = append(codigos, "BL-ESSALUD04-20260401")
	}
	return &integDomain.ReservaLegacyResultado{
		Exito:             true,
		BancoID:           bancoID,
		CodigoTransaccion: "SOAP-TX-01",
		UnidadesIDs:       unidades,
		CodigosBarras:     codigos,
	}, nil
}

func TestSrvAtenderEmergencia_CasosCP01_CP02_ContingenciaE1_Y_409(t *testing.T) {
	repo := newMockEmergenciaRepo()
	bus := events.NewInMemoryEventBus(64, 1)
	defer bus.Close()
	mapas := integMaps.NewHaversineMapAdapter(integMaps.NewInMemoryGeoCache())
	legacy := &mockLegacySoap{}

	srv := NewSrvAtenderEmergencia(repo, mapas, legacy, bus)
	ctx := context.Background()

	// 1. CP-01 / CP-02: Sin stock local en HOSP-MINSA-01 -> Salto a RED_EXTERNA más cercana (HOSP-ESSALUD-04 vía SOAP Legacy)
	resp, conflict, err := srv.CrearSolicitudYMatching(ctx, ports.CrearEmergenciaRequest{
		HospitalOrigenID:  "HOSP-MINSA-01",
		CoordenadasOrigen: "-5.1945,-80.6328",
		MedicoID:          12,
		TipoSangreReq:     "O",
		FactorRhReq:       "+",
		CantidadReq:       2,
		Urgencia:          "ALTA_VITAL",
	}, "12")
	if err != nil || conflict != nil {
		t.Fatalf("esperaba matching exitoso en red externa, obtuvo err=%v conflict=%v", err, conflict)
	}
	if resp.OrigenAsignacion != "RED_EXTERNA" || resp.BancoAsignado.BancoID != "HOSP-ESSALUD-04" {
		t.Errorf("esperaba RED_EXTERNA en HOSP-ESSALUD-04, obtuvo %s en %+v", resp.OrigenAsignacion, resp.BancoAsignado)
	}

	// 2. Prioridad Local cuando sí hay stock local
	repo.stockPorBco["HOSP-MINSA-01"] = 3
	respLocal, _, err := srv.CrearSolicitudYMatching(ctx, ports.CrearEmergenciaRequest{
		HospitalOrigenID: "HOSP-MINSA-01",
		TipoSangreReq:    "A",
		FactorRhReq:      "-",
		CantidadReq:      2,
	}, "12")
	if err != nil || respLocal.OrigenAsignacion != "LOCAL" {
		t.Fatalf("esperaba asignación LOCAL, obtuvo %+v (err=%v)", respLocal, err)
	}

	// 3. Contingencia E1 (SimularFalloMapas = true -> fallbackJurisdiccionUsado = true, elige HOSP-MINSA-02 que comparte jurisdicción)
	repo.stockPorBco["HOSP-MINSA-01"] = 0
	respE1, _, err := srv.CrearSolicitudYMatching(ctx, ports.CrearEmergenciaRequest{
		HospitalOrigenID:  "HOSP-MINSA-01",
		TipoSangreReq:     "O",
		FactorRhReq:       "+",
		CantidadReq:       1,
		SimularFalloMapas: true,
	}, "12")
	if err != nil || !respE1.FallbackJurisdiccionUsado {
		t.Fatalf("esperaba FallbackJurisdiccionUsado=true, obtuvo %+v (err=%v)", respE1, err)
	}

	// 4. Obtener, Listar, Confirmar y Cancelar
	if _, err := srv.ObtenerSolicitud(ctx, resp.SolicitudID); err != nil {
		t.Fatalf("ObtenerSolicitud error: %v", err)
	}
	if list, err := srv.ListarSolicitudes(ctx, ""); err != nil || len(list) < 3 {
		t.Fatalf("ListarSolicitudes error: %v", err)
	}
	if _, err := srv.ConfirmarTransfusion(ctx, resp.SolicitudID, "12"); err != nil {
		t.Fatalf("ConfirmarTransfusion error: %v", err)
	}
	if _, err := srv.CancelarSolicitud(ctx, respLocal.SolicitudID, "Paciente estabilizado", "12"); err != nil {
		t.Fatalf("CancelarSolicitud error: %v", err)
	}

	// 5. Sin Stock en toda la red -> 409 Conflict + Convocatoria
	_, conflict409, err409 := srv.CrearSolicitudYMatching(ctx, ports.CrearEmergenciaRequest{
		HospitalOrigenID: "HOSP-MINSA-01",
		TipoSangreReq:    "AB",
		FactorRhReq:      "-",
		CantidadReq:      50,
	}, "12")
	if !errors.Is(err409, domain.ErrStockNoDisponibleEnRed) || conflict409 == nil || !conflict409.SugerirConvocatoriaDonantes {
		t.Fatalf("esperaba 409 Conflict con SugerirConvocatoriaDonantes=true, obtuvo err=%v conflict=%+v", err409, conflict409)
	}

	// 6. Errores por hospital inexistente y parámetros inválidos
	if _, _, err := srv.CrearSolicitudYMatching(ctx, ports.CrearEmergenciaRequest{HospitalOrigenID: "NO-EXISTE", TipoSangreReq: "O", FactorRhReq: "+", CantidadReq: 1}, "12"); err == nil {
		t.Error("esperaba error para hospital inexistente")
	}
	if _, _, err := srv.CrearSolicitudYMatching(ctx, ports.CrearEmergenciaRequest{HospitalOrigenID: "HOSP-MINSA-01", TipoSangreReq: "Z", FactorRhReq: "+", CantidadReq: 1}, "12"); err == nil {
		t.Error("esperaba error para tipo de sangre inválido")
	}
	if _, err := srv.ObtenerSolicitud(ctx, "NO-EXISTE"); err == nil {
		t.Error("esperaba error al obtener solicitud inexistente")
	}
}
