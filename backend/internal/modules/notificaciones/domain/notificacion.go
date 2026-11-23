package domain

import "time"

const (
	CanalSMS         = "SMS"
	CanalEmail       = "EMAIL"
	CanalAlertaPanel = "ALERTA_PANEL"
)

type Notificacion struct {
	ID           uint      `json:"id"`
	Evento       string    `json:"evento"`
	Canal        string    `json:"canal"`
	Destinatario string    `json:"destinatario"`
	Mensaje      string    `json:"mensaje"`
	Timestamp    time.Time `json:"timestamp"`
}

type UnidadPorCaducar struct {
	UnidadID        string `json:"unidadId"`
	CodigoBarras    string `json:"codigoBarras"`
	BancoID         string `json:"bancoId"`
	TipoSangre      string `json:"tipoSangre"`
	FactorRh        string `json:"factorRh"`
	FechaCaducidad  string `json:"fechaCaducidad"`
	DiasRestantes   int    `json:"diasRestantes"`
	AlertaCaducidad string `json:"alertaCaducidad"`
}

type DonanteConvocable struct {
	DonanteID        string
	NombreCompleto   string
	Correo           string
	TipoSangre       string
	FactorRh         string
	BancoPreferidoID string
}
