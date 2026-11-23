package application

import (
	"context"
	"fmt"
	"time"

	"github.com/bloodlink/backend/internal/events"
	"github.com/bloodlink/backend/internal/modules/notificaciones/domain"
	"github.com/bloodlink/backend/internal/modules/notificaciones/ports"
)

type SrvNotificaciones struct {
	repo     ports.INotificacionRepository
	eventBus events.IEventBus
}

func NewSrvNotificaciones(repo ports.INotificacionRepository, eventBus events.IEventBus) *SrvNotificaciones {
	return &SrvNotificaciones{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (s *SrvNotificaciones) RegistrarSuscripcionesEventos() {
	if s.eventBus == nil {
		return
	}

	s.eventBus.Subscribe(events.EventUnidadRegistrada, func(ctx context.Context, ev events.DomainEvent) {
		msg := fmt.Sprintf("Nueva unidad %v (%v%v) ingresada en banco %v",
			ev.Payload["codigoBarras"], ev.Payload["tipoSangre"], ev.Payload["factorRh"], ev.Payload["bancoId"])
		_ = s.repo.GuardarNotificacion(ctx, &domain.Notificacion{
			Evento:       string(ev.Type),
			Canal:        domain.CanalAlertaPanel,
			Destinatario: fmt.Sprintf("BANCO:%v", ev.Payload["bancoId"]),
			Mensaje:      msg,
			Timestamp:    ev.Timestamp,
		})
	})

	s.eventBus.Subscribe(events.EventEstadoUnidadActualizado, func(ctx context.Context, ev events.DomainEvent) {
		msg := fmt.Sprintf("Unidad %v cambió de estado %v -> %v",
			ev.Payload["unidadId"], ev.Payload["estadoAnterior"], ev.Payload["estadoNuevo"])
		_ = s.repo.GuardarNotificacion(ctx, &domain.Notificacion{
			Evento:       string(ev.Type),
			Canal:        domain.CanalAlertaPanel,
			Destinatario: "PANEL_TRAZABILIDAD",
			Mensaje:      msg,
			Timestamp:    ev.Timestamp,
		})
	})

	s.eventBus.Subscribe(events.EventUnidadProximaACaducar, func(ctx context.Context, ev events.DomainEvent) {
		msg := fmt.Sprintf("ALERTA CADUCIDAD [%v]: Unidad %v (%v%v) en %v vence el %v (%v días restantes)",
			ev.Payload["alertaCaducidad"], ev.Payload["unidadId"], ev.Payload["tipoSangre"],
			ev.Payload["factorRh"], ev.Payload["bancoId"], ev.Payload["fechaCaducidad"], ev.Payload["diasRestantes"])
		_ = s.repo.GuardarNotificacion(ctx, &domain.Notificacion{
			Evento:       string(ev.Type),
			Canal:        domain.CanalAlertaPanel,
			Destinatario: fmt.Sprintf("COORDINADOR_RED:%v", ev.Payload["bancoId"]),
			Mensaje:      msg,
			Timestamp:    ev.Timestamp,
		})
	})

	s.eventBus.Subscribe(events.EventEmergenciaReservada, func(ctx context.Context, ev events.DomainEvent) {
		msg := fmt.Sprintf("Emergencia %v atendida (%v): %v unidades %v%v reservadas en %v para %v",
			ev.Payload["solicitudId"], ev.Payload["origenAsignacion"], ev.Payload["cantidad"],
			ev.Payload["tipoSangre"], ev.Payload["factorRh"], ev.Payload["bancoAsignadoId"], ev.Payload["hospitalOrigenId"])
		_ = s.repo.GuardarNotificacion(ctx, &domain.Notificacion{
			Evento:       string(ev.Type),
			Canal:        domain.CanalAlertaPanel,
			Destinatario: fmt.Sprintf("DESPACHO:%v->%v", ev.Payload["bancoAsignadoId"], ev.Payload["hospitalOrigenId"]),
			Mensaje:      msg,
			Timestamp:    ev.Timestamp,
		})
	})

	s.eventBus.Subscribe(events.EventSolicitudCancelada, func(ctx context.Context, ev events.DomainEvent) {
		msg := fmt.Sprintf("Solicitud %v cancelada. Se liberaron %v unidades en %v. Motivo: %v",
			ev.Payload["solicitudId"], ev.Payload["unidadesLiberadas"], ev.Payload["bancoAsignadoId"], ev.Payload["motivo"])
		_ = s.repo.GuardarNotificacion(ctx, &domain.Notificacion{
			Evento:       string(ev.Type),
			Canal:        domain.CanalAlertaPanel,
			Destinatario: fmt.Sprintf("BANCO:%v", ev.Payload["bancoAsignadoId"]),
			Mensaje:      msg,
			Timestamp:    ev.Timestamp,
		})
	})

	s.eventBus.Subscribe(events.EventStockCriticoDetectado, func(ctx context.Context, ev events.DomainEvent) {
		tipo, _ := ev.Payload["tipoSangre"].(string)
		rh, _ := ev.Payload["factorRh"].(string)
		solID, _ := ev.Payload["solicitudId"].(string)

		alertMsg := fmt.Sprintf("STOCK CRÍTICO DETECTADO para grupo %s%s (Solicitud %s). Activando convocatoria automática de donantes voluntarios.", tipo, rh, solID)
		_ = s.repo.GuardarNotificacion(ctx, &domain.Notificacion{
			Evento:       string(ev.Type),
			Canal:        domain.CanalAlertaPanel,
			Destinatario: "COORDINACION_REGIONAL_PIURA",
			Mensaje:      alertMsg,
			Timestamp:    ev.Timestamp,
		})

		donantes, err := s.repo.BuscarDonantesCompatibles(ctx, tipo, rh)
		if err == nil {
			for _, d := range donantes {
				smsMsg := fmt.Sprintf("[MINSA BloodLink] Estimado(a) %s, requerimos donación urgente de sangre %s%s en %s. Acérquese con su DNI.",
					d.NombreCompleto, d.TipoSangre, d.FactorRh, d.BancoPreferidoID)
				_ = s.repo.GuardarNotificacion(ctx, &domain.Notificacion{
					Evento:       string(ev.Type),
					Canal:        domain.CanalSMS,
					Destinatario: fmt.Sprintf("%s (%s)", d.NombreCompleto, d.DonanteID),
					Mensaje:      smsMsg,
					Timestamp:    time.Now(),
				})
				_ = s.repo.GuardarNotificacion(ctx, &domain.Notificacion{
					Evento:       string(ev.Type),
					Canal:        domain.CanalEmail,
					Destinatario: d.Correo,
					Mensaje:      smsMsg,
					Timestamp:    time.Now(),
				})
			}
		}
	})
}

func (s *SrvNotificaciones) ListarNotificaciones(ctx context.Context, canal string, limit int) ([]domain.Notificacion, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListarNotificaciones(ctx, canal, limit)
}

func (s *SrvNotificaciones) EjecutarJobVerificacionCaducidad(ctx context.Context) (*ports.VerificacionCaducidadResponse, error) {
	unidades, err := s.repo.BuscarUnidadesProximasACaducar7Dias(ctx)
	if err != nil {
		return nil, err
	}

	publicadas := 0
	for _, u := range unidades {
		if s.eventBus != nil {
			s.eventBus.Publish(ctx, events.EventUnidadProximaACaducar, map[string]any{
				"unidadId":        u.UnidadID,
				"codigoBarras":    u.CodigoBarras,
				"bancoId":         u.BancoID,
				"tipoSangre":      u.TipoSangre,
				"factorRh":        u.FactorRh,
				"fechaCaducidad":  u.FechaCaducidad,
				"diasRestantes":   u.DiasRestantes,
				"alertaCaducidad": u.AlertaCaducidad,
			})
			publicadas++
		}
	}

	return &ports.VerificacionCaducidadResponse{
		TotalDetectadas:   len(unidades),
		AlertasPublicadas: publicadas,
		Unidades:          unidades,
		EjecutadoEn:       time.Now().Format(time.RFC3339),
	}, nil
}
