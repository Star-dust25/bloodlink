package domain

import (
	"errors"
	"time"
)

var ErrTrazabilidadUnidadNoEncontrada = errors.New("unidad de sangre no encontrada para reporte de trazabilidad")

type StockPorGrupoKPI struct {
	GrupoSanguineo      string `json:"grupoSanguineo"`
	TipoSangre          string `json:"tipoSangre"`
	FactorRh            string `json:"factorRh"`
	UnidadesDisponibles int    `json:"unidadesDisponibles"`
	UnidadesReservadas  int    `json:"unidadesReservadas"`
}

type EmergenciaEstadoKPI struct {
	Estado   string `json:"estado"`
	Cantidad int    `json:"cantidad"`
}

type EventoAuditItem struct {
	ID             uint      `json:"id"`
	Entidad        string    `json:"entidad"`
	EntidadID      string    `json:"entidadId"`
	Accion         string    `json:"accion"`
	EstadoAnterior string    `json:"estadoAnterior"`
	EstadoNuevo    string    `json:"estadoNuevo"`
	UsuarioID      string    `json:"usuarioId"`
	Detalle        string    `json:"detalle"`
	Timestamp      time.Time `json:"timestamp"`
}

type NotificacionResumenItem struct {
	ID           uint      `json:"id"`
	Evento       string    `json:"evento"`
	Canal        string    `json:"canal"`
	Destinatario string    `json:"destinatario"`
	Mensaje      string    `json:"mensaje"`
	Timestamp    time.Time `json:"timestamp"`
}

type DashboardMetricas struct {
	TotalBancosRed             int                       `json:"totalBancosRed"`
	TotalUnidadesDisponibles   int                       `json:"totalUnidadesDisponibles"`
	UnidadesProximasACaducar7d int                       `json:"unidadesProximasACaducar7d"`
	UnidadesAlertaRoja3d       int                       `json:"unidadesAlertaRoja3d"`
	TotalEmergencias           int                       `json:"totalEmergencias"`
	EmergenciasActivas         int                       `json:"emergenciasActivas"`
	EmergenciasAtendidas       int                       `json:"emergenciasAtendidas"`
	TiempoPromedioMatchingMs   float64                   `json:"tiempoPromedioMatchingMs"`
	DonantesHabilitados        int                       `json:"donantesHabilitados"`
	StockPorTipoSangre         []StockPorGrupoKPI        `json:"stockPorTipoSangre"`
	EmergenciasPorEstado       []EmergenciaEstadoKPI     `json:"emergenciasPorEstado"`
	UltimasNotificaciones      []NotificacionResumenItem `json:"ultimasNotificaciones"`
	GeneradoEn                 string                    `json:"generadoEn"`
}

type TrazabilidadUnidadReporte struct {
	UnidadID        string            `json:"unidadId"`
	CodigoBarras    string            `json:"codigoBarras"`
	BancoID         string            `json:"bancoId"`
	BancoNombre     string            `json:"bancoNombre"`
	TipoSangre      string            `json:"tipoSangre"`
	FactorRh        string            `json:"factorRh"`
	VolumenMl       int               `json:"volumenMl"`
	Origen          string            `json:"origen"`
	FechaExtraccion string            `json:"fechaExtraccion"`
	FechaCaducidad  string            `json:"fechaCaducidad"`
	EstadoActual    string            `json:"estadoActual"`
	Historial       []EventoAuditItem `json:"historial"`
	GeneradoEn      string            `json:"generadoEn"`
}
