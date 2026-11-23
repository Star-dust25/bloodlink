package ports

import (
	"context"

	"github.com/bloodlink/backend/internal/modules/notificaciones/domain"
)

type VerificacionCaducidadResponse struct {
	TotalDetectadas   int                       `json:"totalDetectadas"`
	AlertasPublicadas int                       `json:"alertasPublicadas"`
	Unidades          []domain.UnidadPorCaducar `json:"unidades"`
	EjecutadoEn       string                    `json:"ejecutadoEn"`
}

// INotificacionService defines the inbound port for MS-Notificaciones.
type INotificacionService interface {
	RegistrarSuscripcionesEventos()
	ListarNotificaciones(ctx context.Context, canal string, limit int) ([]domain.Notificacion, error)
	EjecutarJobVerificacionCaducidad(ctx context.Context) (*VerificacionCaducidadResponse, error)
}

// INotificacionRepository defines the outbound persistence port for notifications and alerts.
type INotificacionRepository interface {
	GuardarNotificacion(ctx context.Context, n *domain.Notificacion) error
	ListarNotificaciones(ctx context.Context, canal string, limit int) ([]domain.Notificacion, error)
	BuscarDonantesCompatibles(ctx context.Context, tipoSangre, factorRh string) ([]domain.DonanteConvocable, error)
	BuscarUnidadesProximasACaducar7Dias(ctx context.Context) ([]domain.UnidadPorCaducar, error)
}
