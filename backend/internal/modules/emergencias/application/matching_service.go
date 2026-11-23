package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bloodlink/backend/internal/events"
	"github.com/bloodlink/backend/internal/modules/emergencias/domain"
	"github.com/bloodlink/backend/internal/modules/emergencias/ports"
	integPorts "github.com/bloodlink/backend/internal/modules/integracion/ports"
	"github.com/google/uuid"
)

type SrvAtenderEmergencia struct {
	repo          ports.IEmergenciaRepository
	mapas         integPorts.IMapas
	legacyGateway integPorts.ILegacySoapGateway
	eventBus      events.IEventBus
}

func NewSrvAtenderEmergencia(
	repo ports.IEmergenciaRepository,
	mapas integPorts.IMapas,
	legacyGateway integPorts.ILegacySoapGateway,
	eventBus events.IEventBus,
) *SrvAtenderEmergencia {
	return &SrvAtenderEmergencia{
		repo:          repo,
		mapas:         mapas,
		legacyGateway: legacyGateway,
		eventBus:      eventBus,
	}
}

func (s *SrvAtenderEmergencia) CrearSolicitudYMatching(
	ctx context.Context,
	req ports.CrearEmergenciaRequest,
	usuarioID string,
) (*ports.EmergenciaMatchingResponseDTO, *ports.EmergenciaSinStockConflictDTO, error) {
	startTime := time.Now()

	bancoOrigen, err := s.repo.ObtenerBanco(ctx, req.HospitalOrigenID)
	if err != nil || bancoOrigen == nil {
		return nil, nil, fmt.Errorf("hospital de origen '%s' no existe", req.HospitalOrigenID)
	}

	coordOrigen := strings.TrimSpace(req.CoordenadasOrigen)
	if coordOrigen == "" {
		coordOrigen = bancoOrigen.Coordenadas
	}
	if req.MedicoID == 0 {
		req.MedicoID = 12
	}

	shortUUID := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:8])
	solicitudID := fmt.Sprintf("SOL-%d-%s", startTime.Year(), shortUUID)

	sol, err := domain.NuevaSolicitudEmergencia(
		solicitudID,
		req.HospitalOrigenID,
		coordOrigen,
		req.MedicoID,
		strings.ToUpper(strings.TrimSpace(req.TipoSangreReq)),
		strings.TrimSpace(req.FactorRhReq),
		req.CantidadReq,
		strings.ToUpper(strings.TrimSpace(req.Urgencia)),
		startTime,
	)
	if err != nil {
		return nil, nil, err
	}

	// PASO 1: Regla de Prioridad Local (CUN-001) con bloqueo pesimista FOR UPDATE SKIP LOCKED
	unidadesLocales, errLocal := s.repo.IntentarReservaAtomica(ctx, sol.HospitalOrigenID, sol.TipoSangreReq, sol.FactorRhReq, sol.CantidadReq, usuarioID)
	if errLocal == nil && len(unidadesLocales) == sol.CantidadReq {
		_ = sol.AsignarUnidades(bancoOrigen.BancoID, bancoOrigen.Nombre, domain.OrigenAsignacionLocal, 0.0, 2, unidadesLocales)
		sol.TiempoRespuestaMatchingMs = maxInt64(1, time.Since(startTime).Milliseconds())

		detalle := fmt.Sprintf("Asignación LOCAL inmediata en %s (%d unidades %s%s)", bancoOrigen.BancoID, sol.CantidadReq, sol.TipoSangreReq, sol.FactorRhReq)
		if err := s.repo.GuardarSolicitud(ctx, sol, usuarioID, detalle); err != nil {
			return nil, nil, err
		}

		s.publicarEventosReserva(ctx, sol)
		dto := toEmergenciaDTO(sol)
		return &dto, nil, nil
	}

	// PASO 2: Búsqueda en Red Externa y Cálculo Geoespacial / Fallback Jurisdicción (CUS-001 / Contingencia E1)
	sol.EstadoSolicitud = domain.EstadoSolicitudBuscandoStock
	candidatos, err := s.repo.ListarCandidatosExternos(ctx, sol.HospitalOrigenID, sol.TipoSangreReq, sol.FactorRhReq)
	if err != nil {
		return nil, nil, err
	}

	fallbackJurisdiccion := req.SimularFalloMapas
	for i := range candidatos {
		candidatos[i].MismaJurisdiccion = strings.EqualFold(candidatos[i].Jurisdiccion, bancoOrigen.Jurisdiccion)
		if s.mapas != nil {
			ruta, errMapa := s.mapas.CalcularDistanciaYTiempo(ctx, coordOrigen, candidatos[i].Latitud, candidatos[i].Longitud, req.SimularFalloMapas)
			if errMapa != nil {
				fallbackJurisdiccion = true
				// En modo contingencia jurisdiccional estimamos distancia referencial por jurisdicción
				if candidatos[i].MismaJurisdiccion {
					candidatos[i].DistanciaKm = 2.5
					candidatos[i].TiempoEstimadoMin = 7
				} else {
					candidatos[i].DistanciaKm = 15.0
					candidatos[i].TiempoEstimadoMin = 25
				}
			} else {
				candidatos[i].DistanciaKm = ruta.DistanciaKm
				candidatos[i].TiempoEstimadoMin = ruta.TiempoEstimadoMin
			}
		}
	}

	// Ordenar candidatos usando el algoritmo de dominio GenerarMatching
	_, errMatch := sol.GenerarMatching(candidatos, fallbackJurisdiccion)
	if errMatch != nil {
		// Escenario sin stock en toda la red -> 409 Conflict + Convocatoria de donantes
		sol.TiempoRespuestaMatchingMs = maxInt64(1, time.Since(startTime).Milliseconds())
		detalle := fmt.Sprintf("Stock insuficiente en red para %d unidades %s%s. Activando convocatoria de donantes", sol.CantidadReq, sol.TipoSangreReq, sol.FactorRhReq)
		_ = s.repo.GuardarSolicitud(ctx, sol, usuarioID, detalle)

		if s.eventBus != nil {
			s.eventBus.Publish(ctx, events.EventStockCriticoDetectado, map[string]any{
				"solicitudId":      sol.ID,
				"hospitalOrigenId": sol.HospitalOrigenID,
				"tipoSangre":       sol.TipoSangreReq,
				"factorRh":         sol.FactorRhReq,
				"cantidadFaltante": sol.CantidadReq,
				"motivo":           "SIN_STOCK_EN_RED_EMERGENCIA",
			})
		}

		conflict := &ports.EmergenciaSinStockConflictDTO{
			SolicitudID:                 sol.ID,
			EstadoSolicitud:             domain.EstadoSolicitudEnEsperaDonantes,
			Mensaje:                     "Stock no disponible en red",
			SugerirConvocatoriaDonantes: true,
			TiempoRespuestaMatchingMs:   sol.TiempoRespuestaMatchingMs,
			FechaHora:                   sol.FechaHora.Format(time.RFC3339),
		}
		return nil, conflict, domain.ErrStockNoDisponibleEnRed
	}

	// Intentar reserva atómica siguiendo el orden de candidatos elegibles (maneja condiciones de carrera concurrentes)
	var unidadesReservadas []domain.UnidadAsignada
	var bancoGanador *domain.CandidatoBanco

	// Re-obtener lista ordenada por el mismo criterio para iterar en caso de empate concurrente
	for len(candidatos) > 0 {
		ganador, errG := sol.GenerarMatching(candidatos, fallbackJurisdiccion)
		if errG != nil || ganador == nil {
			break
		}

		if ganador.EsLegacySoap && s.legacyGateway != nil {
			soapRes, errSoap := s.legacyGateway.ConsultarYReservarStockLegacy(ctx, ganador.BancoID, sol.TipoSangreReq, sol.FactorRhReq, sol.CantidadReq)
			if errSoap == nil && soapRes != nil && soapRes.Exito && len(soapRes.UnidadesIDs) == sol.CantidadReq {
				for idx, uID := range soapRes.UnidadesIDs {
					bc := uID
					if idx < len(soapRes.CodigosBarras) {
						bc = soapRes.CodigosBarras[idx]
					}
					unidadesReservadas = append(unidadesReservadas, domain.UnidadAsignada{
						UnidadID:     uID,
						CodigoBarras: bc,
					})
				}
				bancoGanador = ganador
				break
			}
			// Fallback directo a transacción BD si el servidor HTTP aún está en loopback interno o test unitario
		}

		unidadesDB, errDB := s.repo.IntentarReservaAtomica(ctx, ganador.BancoID, sol.TipoSangreReq, sol.FactorRhReq, sol.CantidadReq, usuarioID)
		if errDB == nil && len(unidadesDB) == sol.CantidadReq {
			unidadesReservadas = unidadesDB
			bancoGanador = ganador
			break
		}

		// Si otro hilo reservó las unidades de este banco milisegundos antes, descartar banco y probar el siguiente más cercano
		restantes := make([]domain.CandidatoBanco, 0, len(candidatos)-1)
		for _, c := range candidatos {
			if c.BancoID != ganador.BancoID {
				restantes = append(restantes, c)
			}
		}
		candidatos = restantes
	}

	if bancoGanador == nil || len(unidadesReservadas) < sol.CantidadReq {
		sol.EstadoSolicitud = domain.EstadoSolicitudEnEsperaDonantes
		sol.OrigenAsignacion = domain.OrigenAsignacionNinguno
		sol.TiempoRespuestaMatchingMs = maxInt64(1, time.Since(startTime).Milliseconds())
		_ = s.repo.GuardarSolicitud(ctx, sol, usuarioID, "Stock agotado por concurrencia; pasa a espera de donantes")

		if s.eventBus != nil {
			s.eventBus.Publish(ctx, events.EventStockCriticoDetectado, map[string]any{
				"solicitudId":      sol.ID,
				"hospitalOrigenId": sol.HospitalOrigenID,
				"tipoSangre":       sol.TipoSangreReq,
				"factorRh":         sol.FactorRhReq,
				"cantidadFaltante": sol.CantidadReq,
			})
		}

		conflict := &ports.EmergenciaSinStockConflictDTO{
			SolicitudID:                 sol.ID,
			EstadoSolicitud:             domain.EstadoSolicitudEnEsperaDonantes,
			Mensaje:                     "Stock no disponible en red",
			SugerirConvocatoriaDonantes: true,
			TiempoRespuestaMatchingMs:   sol.TiempoRespuestaMatchingMs,
			FechaHora:                   sol.FechaHora.Format(time.RFC3339),
		}
		return nil, conflict, domain.ErrStockNoDisponibleEnRed
	}

	_ = sol.AsignarUnidades(
		bancoGanador.BancoID,
		bancoGanador.Nombre,
		domain.OrigenAsignacionRedExterna,
		bancoGanador.DistanciaKm,
		bancoGanador.TiempoEstimadoMin,
		unidadesReservadas,
	)
	sol.TiempoRespuestaMatchingMs = maxInt64(1, time.Since(startTime).Milliseconds())

	detalle := fmt.Sprintf("Matching exitoso en RED_EXTERNA (%s a %.2f km, %d min). FallbackJurisdiccion=%t",
		bancoGanador.BancoID, bancoGanador.DistanciaKm, bancoGanador.TiempoEstimadoMin, sol.FallbackJurisdiccionUsado)
	if err := s.repo.GuardarSolicitud(ctx, sol, usuarioID, detalle); err != nil {
		return nil, nil, err
	}

	s.publicarEventosReserva(ctx, sol)
	dto := toEmergenciaDTO(sol)
	return &dto, nil, nil
}

func (s *SrvAtenderEmergencia) publicarEventosReserva(ctx context.Context, sol *domain.SolicitudEmergencia) {
	if s.eventBus == nil {
		return
	}
	s.eventBus.Publish(ctx, events.EventEmergenciaReservada, map[string]any{
		"solicitudId":      sol.ID,
		"hospitalOrigenId": sol.HospitalOrigenID,
		"bancoAsignadoId":  sol.BancoAsignadoID,
		"origenAsignacion": sol.OrigenAsignacion,
		"tipoSangre":       sol.TipoSangreReq,
		"factorRh":         sol.FactorRhReq,
		"cantidad":         sol.CantidadReq,
		"distanciaKm":      sol.DistanciaKm,
	})

	restante, err := s.repo.ContarStockDisponibleGlobalPorTipo(ctx, sol.TipoSangreReq, sol.FactorRhReq)
	if err == nil && restante < 5 { // <10% de capacidad base (50)
		s.eventBus.Publish(ctx, events.EventStockCriticoDetectado, map[string]any{
			"solicitudId":     sol.ID,
			"tipoSangre":      sol.TipoSangreReq,
			"factorRh":        sol.FactorRhReq,
			"stockDisponible": restante,
			"motivo":          "UMBRAL_CRITICO_MENOR_10_PORCIENTO",
		})
	}
}

func (s *SrvAtenderEmergencia) ListarSolicitudes(ctx context.Context, estado string) ([]ports.EmergenciaMatchingResponseDTO, error) {
	list, err := s.repo.ListarSolicitudes(ctx, estado)
	if err != nil {
		return nil, err
	}
	res := make([]ports.EmergenciaMatchingResponseDTO, 0, len(list))
	for i := range list {
		res = append(res, toEmergenciaDTO(&list[i]))
	}
	return res, nil
}

func (s *SrvAtenderEmergencia) ObtenerSolicitud(ctx context.Context, solicitudID string) (*ports.EmergenciaMatchingResponseDTO, error) {
	sol, err := s.repo.ObtenerSolicitudPorID(ctx, solicitudID)
	if err != nil {
		return nil, err
	}
	dto := toEmergenciaDTO(sol)
	return &dto, nil
}

func (s *SrvAtenderEmergencia) CancelarSolicitud(ctx context.Context, solicitudID, motivo, usuarioID string) (*ports.EmergenciaMatchingResponseDTO, error) {
	sol, err := s.repo.ObtenerSolicitudPorID(ctx, solicitudID)
	if err != nil {
		return nil, err
	}
	if err := sol.Cancelar(motivo); err != nil {
		return nil, err
	}
	if err := s.repo.CancelarSolicitudYLiberarUnidades(ctx, sol, usuarioID); err != nil {
		return nil, err
	}

	if s.eventBus != nil {
		s.eventBus.Publish(ctx, events.EventSolicitudCancelada, map[string]any{
			"solicitudId":       sol.ID,
			"hospitalOrigenId":  sol.HospitalOrigenID,
			"bancoAsignadoId":   sol.BancoAsignadoID,
			"unidadesLiberadas": len(sol.UnidadesAsignadas),
			"motivo":            sol.MotivoCancelacion,
		})
	}

	dto := toEmergenciaDTO(sol)
	return &dto, nil
}

func (s *SrvAtenderEmergencia) ConfirmarTransfusion(ctx context.Context, solicitudID, usuarioID string) (*ports.EmergenciaMatchingResponseDTO, error) {
	sol, err := s.repo.ObtenerSolicitudPorID(ctx, solicitudID)
	if err != nil {
		return nil, err
	}
	if err := sol.ConfirmarTransfusion(); err != nil {
		return nil, err
	}
	if err := s.repo.ConfirmarTransfusionYUnidades(ctx, sol, usuarioID); err != nil {
		return nil, err
	}

	if s.eventBus != nil {
		for _, u := range sol.UnidadesAsignadas {
			s.eventBus.Publish(ctx, events.EventEstadoUnidadActualizado, map[string]any{
				"unidadId":       u.UnidadID,
				"solicitudId":    sol.ID,
				"estadoAnterior": "Reservada",
				"estadoNuevo":    "Transfundida",
			})
		}
	}

	dto := toEmergenciaDTO(sol)
	return &dto, nil
}

func toEmergenciaDTO(sol *domain.SolicitudEmergencia) ports.EmergenciaMatchingResponseDTO {
	var bancoDTO *ports.BancoAsignadoDTO
	if sol.BancoAsignadoID != "" {
		bancoDTO = &ports.BancoAsignadoDTO{
			BancoID:           sol.BancoAsignadoID,
			Nombre:            sol.BancoAsignadoNombre,
			DistanciaKm:       sol.DistanciaKm,
			TiempoEstimadoMin: sol.TiempoEstimadoMin,
		}
	}
	unidades := sol.UnidadesAsignadas
	if unidades == nil {
		unidades = []domain.UnidadAsignada{}
	}
	return ports.EmergenciaMatchingResponseDTO{
		SolicitudID:               sol.ID,
		EstadoSolicitud:           sol.EstadoSolicitud,
		OrigenAsignacion:          sol.OrigenAsignacion,
		HospitalOrigenID:          sol.HospitalOrigenID,
		TipoSangreReq:             sol.TipoSangreReq,
		FactorRhReq:               sol.FactorRhReq,
		CantidadReq:               sol.CantidadReq,
		Urgencia:                  sol.Urgencia,
		BancoAsignado:             bancoDTO,
		UnidadesAsignadas:         unidades,
		TiempoRespuestaMatchingMs: sol.TiempoRespuestaMatchingMs,
		FallbackJurisdiccionUsado: sol.FallbackJurisdiccionUsado,
		MotivoCancelacion:         sol.MotivoCancelacion,
		FechaHora:                 sol.FechaHora.Format(time.RFC3339),
	}
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

var _ = errors.Is
