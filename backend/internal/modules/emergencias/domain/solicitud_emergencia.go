package domain

import (
	"errors"
	"fmt"
	"sort"
	"time"
)

const (
	EstadoSolicitudRegistrada       = "Registrada"
	EstadoSolicitudBuscandoStock    = "Buscando Stock"
	EstadoSolicitudReservada        = "Reservada"
	EstadoSolicitudEnEsperaDonantes = "En Espera de Donantes"
	EstadoSolicitudAtendida         = "Atendida"
	EstadoSolicitudCancelada        = "Cancelada"

	OrigenAsignacionLocal      = "LOCAL"
	OrigenAsignacionRedExterna = "RED_EXTERNA"
	OrigenAsignacionNinguno    = "NINGUNO"

	UrgenciaAltaVital = "ALTA_VITAL"
	UrgenciaMedia     = "MEDIA"
)

var (
	ErrStockNoDisponibleEnRed       = errors.New("stock no disponible en ningún banco de la red")
	ErrSolicitudNoEncontrada        = errors.New("solicitud de emergencia no encontrada")
	ErrTransicionSolicitudInvalida  = errors.New("transición de estado no permitida para SolicitudEmergencia")
	ErrParametrosEmergenciaInvalidos = errors.New("parámetros de solicitud de emergencia inválidos")
)

type UnidadAsignada struct {
	UnidadID     string `json:"unidadId"`
	CodigoBarras string `json:"codigoBarras"`
}

type CandidatoBanco struct {
	BancoID             string
	Nombre              string
	Coordenadas         string
	Latitud             float64
	Longitud            float64
	Jurisdiccion        string
	EsLegacySoap        bool
	UnidadesDisponibles int
	DistanciaKm         float64
	TiempoEstimadoMin   int
	MismaJurisdiccion   bool
}

type SolicitudEmergencia struct {
	ID                        string           `json:"solicitudId"`
	FechaHora                 time.Time        `json:"fechaHora"`
	HospitalOrigenID          string           `json:"hospitalOrigenId"`
	CoordenadasOrigen         string           `json:"coordenadasOrigen"`
	MedicoID                  uint             `json:"medicoId"`
	TipoSangreReq             string           `json:"tipoSangreReq"`
	FactorRhReq               string           `json:"factorRhReq"`
	CantidadReq               int              `json:"cantidadReq"`
	Urgencia                  string           `json:"urgencia"`
	EstadoSolicitud           string           `json:"estadoSolicitud"`
	OrigenAsignacion          string           `json:"origenAsignacion"`
	BancoAsignadoID           string           `json:"bancoAsignadoId"`
	BancoAsignadoNombre       string           `json:"bancoAsignadoNombre"`
	DistanciaKm               float64          `json:"distanciaKm"`
	TiempoEstimadoMin         int              `json:"tiempoEstimadoMin"`
	TiempoRespuestaMatchingMs int64            `json:"tiempoRespuestaMatchingMs"`
	FallbackJurisdiccionUsado bool             `json:"fallbackJurisdiccionUsado"`
	MotivoCancelacion         string           `json:"motivoCancelacion,omitempty"`
	UnidadesAsignadas         []UnidadAsignada `json:"unidadesAsignadas"`
}

func NuevaSolicitudEmergencia(id, hospitalOrigenID, coordOrigen string, medicoID uint, tipoSangre, factorRh string, cantidad int, urgencia string, ahora time.Time) (*SolicitudEmergencia, error) {
	if hospitalOrigenID == "" || cantidad <= 0 || (tipoSangre != "O" && tipoSangre != "A" && tipoSangre != "B" && tipoSangre != "AB") || (factorRh != "+" && factorRh != "-") {
		return nil, ErrParametrosEmergenciaInvalidos
	}
	if urgencia == "" {
		urgencia = UrgenciaAltaVital
	}
	return &SolicitudEmergencia{
		ID:                id,
		FechaHora:         ahora,
		HospitalOrigenID:  hospitalOrigenID,
		CoordenadasOrigen: coordOrigen,
		MedicoID:          medicoID,
		TipoSangreReq:     tipoSangre,
		FactorRhReq:       factorRh,
		CantidadReq:       cantidad,
		Urgencia:          urgencia,
		EstadoSolicitud:   EstadoSolicitudRegistrada,
		OrigenAsignacion:  OrigenAsignacionNinguno,
		UnidadesAsignadas: []UnidadAsignada{},
	}, nil
}

// GenerarMatching ordena los bancos candidatos con stock suficiente priorizando menor distancia
// o misma jurisdicción cuando se activa la Contingencia E1 (fallbackJurisdiccion).
func (s *SolicitudEmergencia) GenerarMatching(candidatos []CandidatoBanco, fallbackJurisdiccion bool) (*CandidatoBanco, error) {
	if s.EstadoSolicitud == EstadoSolicitudRegistrada {
		s.EstadoSolicitud = EstadoSolicitudBuscandoStock
	}

	filtrados := make([]CandidatoBanco, 0, len(candidatos))
	for _, c := range candidatos {
		if c.UnidadesDisponibles >= s.CantidadReq {
			filtrados = append(filtrados, c)
		}
	}

	if len(filtrados) == 0 {
		s.EstadoSolicitud = EstadoSolicitudEnEsperaDonantes
		s.OrigenAsignacion = OrigenAsignacionNinguno
		return nil, ErrStockNoDisponibleEnRed
	}

	if fallbackJurisdiccion {
		s.FallbackJurisdiccionUsado = true
		sort.SliceStable(filtrados, func(i, j int) bool {
			if filtrados[i].MismaJurisdiccion != filtrados[j].MismaJurisdiccion {
				return filtrados[i].MismaJurisdiccion
			}
			return filtrados[i].UnidadesDisponibles > filtrados[j].UnidadesDisponibles
		})
	} else {
		sort.SliceStable(filtrados, func(i, j int) bool {
			if filtrados[i].DistanciaKm != filtrados[j].DistanciaKm {
				return filtrados[i].DistanciaKm < filtrados[j].DistanciaKm
			}
			return filtrados[i].UnidadesDisponibles > filtrados[j].UnidadesDisponibles
		})
	}

	elegido := filtrados[0]
	return &elegido, nil
}

func (s *SolicitudEmergencia) AsignarUnidades(bancoID, bancoNombre, origen string, distanciaKm float64, tiempoMin int, unidades []UnidadAsignada) error {
	if s.EstadoSolicitud != EstadoSolicitudRegistrada &&
		s.EstadoSolicitud != EstadoSolicitudBuscandoStock &&
		s.EstadoSolicitud != EstadoSolicitudEnEsperaDonantes {
		return fmt.Errorf("%w: no se pueden asignar unidades en estado '%s'", ErrTransicionSolicitudInvalida, s.EstadoSolicitud)
	}
	s.EstadoSolicitud = EstadoSolicitudReservada
	s.OrigenAsignacion = origen
	s.BancoAsignadoID = bancoID
	s.BancoAsignadoNombre = bancoNombre
	s.DistanciaKm = distanciaKm
	s.TiempoEstimadoMin = tiempoMin
	s.UnidadesAsignadas = unidades
	return nil
}

func (s *SolicitudEmergencia) AsignarUnidad(unidadID, codigoBarras string) {
	s.UnidadesAsignadas = append(s.UnidadesAsignadas, UnidadAsignada{
		UnidadID:     unidadID,
		CodigoBarras: codigoBarras,
	})
}

func (s *SolicitudEmergencia) Cancelar(motivo string) error {
	if s.EstadoSolicitud != EstadoSolicitudReservada &&
		s.EstadoSolicitud != EstadoSolicitudEnEsperaDonantes &&
		s.EstadoSolicitud != EstadoSolicitudRegistrada &&
		s.EstadoSolicitud != EstadoSolicitudBuscandoStock {
		return fmt.Errorf("%w: no se puede cancelar una solicitud en estado '%s'", ErrTransicionSolicitudInvalida, s.EstadoSolicitud)
	}
	s.EstadoSolicitud = EstadoSolicitudCancelada
	if motivo == "" {
		motivo = "Cancelada por el médico solicitante"
	}
	s.MotivoCancelacion = motivo
	return nil
}

func (s *SolicitudEmergencia) ConfirmarTransfusion() error {
	if s.EstadoSolicitud != EstadoSolicitudReservada {
		return fmt.Errorf("%w: solo una solicitud 'Reservada' puede pasar a 'Atendida' (estado actual: '%s')", ErrTransicionSolicitudInvalida, s.EstadoSolicitud)
	}
	s.EstadoSolicitud = EstadoSolicitudAtendida
	return nil
}
