package domain

import (
	"errors"
	"testing"
	"time"
)

func TestSolicitudEmergencia_MatchingYMaquinaDeEstados(t *testing.T) {
	now := time.Now()

	// 1. Validación de parámetros
	if _, err := NuevaSolicitudEmergencia("SOL-1", "", "-5.19,-80.63", 12, "O", "+", 2, "", now); !errors.Is(err, ErrParametrosEmergenciaInvalidos) {
		t.Fatalf("esperaba ErrParametrosEmergenciaInvalidos, obtuvo %v", err)
	}

	sol, err := NuevaSolicitudEmergencia("SOL-2026-0801", "HOSP-MINSA-01", "-5.1945,-80.6328", 12, "O", "+", 2, "", now)
	if err != nil || sol.Urgencia != UrgenciaAltaVital {
		t.Fatalf("error creando solicitud: %v", err)
	}

	candidatos := []CandidatoBanco{
		{BancoID: "HOSP-MINSA-02", Nombre: "Sullana", UnidadesDisponibles: 6, DistanciaKm: 38.5, TiempoEstimadoMin: 55, MismaJurisdiccion: false},
		{BancoID: "HOSP-ESSALUD-04", Nombre: "Cayetano Heredia", UnidadesDisponibles: 10, DistanciaKm: 2.2, TiempoEstimadoMin: 6, MismaJurisdiccion: false},
		{BancoID: "CLIN-PRIV-01", Nombre: "Clínica Miraflores", UnidadesDisponibles: 1, DistanciaKm: 1.5, TiempoEstimadoMin: 5, MismaJurisdiccion: true},
	}

	// 2. Matching normal por distancia mínima (debe elegir HOSP-ESSALUD-04 porque CLIN-PRIV-01 solo tiene 1 unidad)
	ganador, err := sol.GenerarMatching(candidatos, false)
	if err != nil {
		t.Fatalf("GenerarMatching error: %v", err)
	}
	if ganador.BancoID != "HOSP-ESSALUD-04" {
		t.Errorf("esperaba banco ganador HOSP-ESSALUD-04, obtuvo %s", ganador.BancoID)
	}

	// 3. Matching con Contingencia E1 (Fallback Jurisdicción + desempate por stock)
	candidatosE1 := []CandidatoBanco{
		{BancoID: "HOSP-MINSA-02", Nombre: "Sullana", UnidadesDisponibles: 6, DistanciaKm: 5.0, MismaJurisdiccion: false},
		{BancoID: "CLIN-MISMA-JUR-1", Nombre: "Banco Misma Jurisdicción 1", UnidadesDisponibles: 3, DistanciaKm: 10.0, MismaJurisdiccion: true},
		{BancoID: "CLIN-MISMA-JUR-2", Nombre: "Banco Misma Jurisdicción 2", UnidadesDisponibles: 8, DistanciaKm: 12.0, MismaJurisdiccion: true},
	}
	ganadorE1, err := sol.GenerarMatching(candidatosE1, true)
	if err != nil || ganadorE1.BancoID != "CLIN-MISMA-JUR-2" || !sol.FallbackJurisdiccionUsado {
		t.Fatalf("Contingencia E1 falló: ganador=%v, fallback=%t, err=%v", ganadorE1, sol.FallbackJurisdiccionUsado, err)
	}

	// 4. Asignar unidades, confirmar transfusión y validar transiciones inválidas
	sol.AsignarUnidad("UND-PRE", "BL-PRE")
	err = sol.AsignarUnidades(ganador.BancoID, ganador.Nombre, OrigenAsignacionRedExterna, ganador.DistanciaKm, ganador.TiempoEstimadoMin, []UnidadAsignada{
		{UnidadID: "UND-1", CodigoBarras: "BL-1"},
		{UnidadID: "UND-2", CodigoBarras: "BL-2"},
	})
	if err != nil || sol.EstadoSolicitud != EstadoSolicitudReservada {
		t.Fatalf("AsignarUnidades error: %v", err)
	}

	// ConfirmarTransfusion y verificar errores de máquina de estados
	if err := sol.ConfirmarTransfusion(); err != nil || sol.EstadoSolicitud != EstadoSolicitudAtendida {
		t.Fatalf("ConfirmarTransfusion error: %v", err)
	}
	if err := sol.ConfirmarTransfusion(); !errors.Is(err, ErrTransicionSolicitudInvalida) {
		t.Errorf("esperaba ErrTransicionSolicitudInvalida al re-confirmar")
	}
	if err := sol.AsignarUnidades("B1", "B1", OrigenAsignacionLocal, 0, 1, nil); !errors.Is(err, ErrTransicionSolicitudInvalida) {
		t.Errorf("esperaba ErrTransicionSolicitudInvalida al asignar en estado Atendida")
	}
	if err := sol.Cancelar(""); !errors.Is(err, ErrTransicionSolicitudInvalida) {
		t.Errorf("esperaba ErrTransicionSolicitudInvalida al cancelar solicitud Atendida, obtuvo %v", err)
	}

	// 5. Cancelación válida con motivo por defecto
	solCancel, _ := NuevaSolicitudEmergencia("SOL-2026-0803", "HOSP-MINSA-01", "-5.19,-80.63", 12, "O", "+", 1, UrgenciaMedia, now)
	if err := solCancel.Cancelar(""); err != nil || solCancel.EstadoSolicitud != EstadoSolicitudCancelada || solCancel.MotivoCancelacion == "" {
		t.Fatalf("Cancelar válida falló: %v", err)
	}

	// 6. Escenario sin stock en red (cantidad requerida mayor al stock de cualquier banco) -> En Espera de Donantes
	solSinStock, _ := NuevaSolicitudEmergencia("SOL-2026-0802", "HOSP-MINSA-01", "-5.19,-80.63", 12, "AB", "-", 25, UrgenciaAltaVital, now)
	_, errSinStock := solSinStock.GenerarMatching(candidatos, false)
	if !errors.Is(errSinStock, ErrStockNoDisponibleEnRed) || solSinStock.EstadoSolicitud != EstadoSolicitudEnEsperaDonantes {
		t.Errorf("esperaba ErrStockNoDisponibleEnRed y estado 'En Espera de Donantes', obtuvo err=%v estado=%s", errSinStock, solSinStock.EstadoSolicitud)
	}
}
