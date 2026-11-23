package ports

import (
	"context"

	"github.com/bloodlink/backend/internal/modules/emergencias/domain"
)

type CrearEmergenciaRequest struct {
	HospitalOrigenID  string `json:"hospitalOrigenId" binding:"required"`
	CoordenadasOrigen string `json:"coordenadasOrigen"`
	MedicoID          uint   `json:"medicoId"`
	TipoSangreReq     string `json:"tipoSangreReq" binding:"required"`
	FactorRhReq       string `json:"factorRhReq" binding:"required"`
	CantidadReq       int    `json:"cantidadReq" binding:"required,min=1"`
	Urgencia          string `json:"urgencia"`
	SimularFalloMapas bool   `json:"simularFalloMapas"`
}

type BancoAsignadoDTO struct {
	BancoID           string  `json:"bancoId"`
	Nombre            string  `json:"nombre"`
	DistanciaKm       float64 `json:"distanciaKm"`
	TiempoEstimadoMin int     `json:"tiempoEstimadoMin"`
}

type EmergenciaMatchingResponseDTO struct {
	SolicitudID               string                  `json:"solicitudId"`
	EstadoSolicitud           string                  `json:"estadoSolicitud"`
	OrigenAsignacion          string                  `json:"origenAsignacion"`
	HospitalOrigenID          string                  `json:"hospitalOrigenId,omitempty"`
	TipoSangreReq             string                  `json:"tipoSangreReq,omitempty"`
	FactorRhReq               string                  `json:"factorRhReq,omitempty"`
	CantidadReq               int                     `json:"cantidadReq,omitempty"`
	Urgencia                  string                  `json:"urgencia,omitempty"`
	BancoAsignado             *BancoAsignadoDTO       `json:"bancoAsignado,omitempty"`
	UnidadesAsignadas         []domain.UnidadAsignada `json:"unidadesAsignadas"`
	TiempoRespuestaMatchingMs int64                   `json:"tiempoRespuestaMatchingMs"`
	FallbackJurisdiccionUsado bool                    `json:"fallbackJurisdiccionUsado"`
	MotivoCancelacion         string                  `json:"motivoCancelacion,omitempty"`
	FechaHora                 string                  `json:"fechaHora"`
}

type EmergenciaSinStockConflictDTO struct {
	SolicitudID                 string `json:"solicitudId"`
	EstadoSolicitud             string `json:"estadoSolicitud"`
	Mensaje                     string `json:"mensaje"`
	SugerirConvocatoriaDonantes bool   `json:"sugerirConvocatoriaDonantes"`
	TiempoRespuestaMatchingMs   int64  `json:"tiempoRespuestaMatchingMs"`
	FechaHora                   string `json:"fechaHora"`
}

type CancelarEmergenciaRequest struct {
	Motivo string `json:"motivo"`
}

// IMatching defines the inbound port for MS-Emergencias (Sección 7.4.1 I_Matching / Srv_AtenderEmergencia).
type IMatching interface {
	CrearSolicitudYMatching(ctx context.Context, req CrearEmergenciaRequest, usuarioID string) (*EmergenciaMatchingResponseDTO, *EmergenciaSinStockConflictDTO, error)
	ListarSolicitudes(ctx context.Context, estado string) ([]EmergenciaMatchingResponseDTO, error)
	ObtenerSolicitud(ctx context.Context, solicitudID string) (*EmergenciaMatchingResponseDTO, error)
	CancelarSolicitud(ctx context.Context, solicitudID, motivo, usuarioID string) (*EmergenciaMatchingResponseDTO, error)
	ConfirmarTransfusion(ctx context.Context, solicitudID, usuarioID string) (*EmergenciaMatchingResponseDTO, error)
}

// IEmergenciaRepository defines the outbound persistence port with pessimistic locking FOR UPDATE SKIP LOCKED.
type IEmergenciaRepository interface {
	ObtenerBanco(ctx context.Context, bancoID string) (*domain.CandidatoBanco, error)
	ListarCandidatosExternos(ctx context.Context, hospitalOrigenID, tipoSangre, factorRh string) ([]domain.CandidatoBanco, error)
	IntentarReservaAtomica(ctx context.Context, bancoID, tipoSangre, factorRh string, cantidad int, usuarioID string) ([]domain.UnidadAsignada, error)
	GuardarSolicitud(ctx context.Context, sol *domain.SolicitudEmergencia, usuarioID, detalle string) error
	ObtenerSolicitudPorID(ctx context.Context, solicitudID string) (*domain.SolicitudEmergencia, error)
	ListarSolicitudes(ctx context.Context, estado string) ([]domain.SolicitudEmergencia, error)
	CancelarSolicitudYLiberarUnidades(ctx context.Context, sol *domain.SolicitudEmergencia, usuarioID string) error
	ConfirmarTransfusionYUnidades(ctx context.Context, sol *domain.SolicitudEmergencia, usuarioID string) error
	ContarStockDisponibleGlobalPorTipo(ctx context.Context, tipoSangre, factorRh string) (int, error)
}
