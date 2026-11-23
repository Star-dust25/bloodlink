package models

import "time"

// BancoSangreORM maps to the PostgreSQL table bancos_sangre.
type BancoSangreORM struct {
	ID              string    `gorm:"primaryKey;size:32" json:"id"`
	Nombre          string    `gorm:"size:160;not null" json:"nombre"`
	Direccion       string    `gorm:"size:220;not null" json:"direccion"`
	Latitud         float64   `gorm:"not null" json:"latitud"`
	Longitud        float64   `gorm:"not null" json:"longitud"`
	Coordenadas     string    `gorm:"size:64;not null" json:"coordenadas"`
	Jurisdiccion    string    `gorm:"size:80;not null;index" json:"jurisdiccion"`
	TipoInstitucion string    `gorm:"size:32;not null" json:"tipoInstitucion"`
	EsLegacySoap    bool      `gorm:"default:false;not null" json:"esLegacySoap"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (BancoSangreORM) TableName() string { return "bancos_sangre" }

// UsuarioPersonalORM maps to the PostgreSQL table usuarios_personal.
type UsuarioPersonalORM struct {
	ID                 uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	NombreCompleto     string     `gorm:"size:150;not null" json:"nombreCompleto"`
	DocumentoIdentidad string     `gorm:"size:20;uniqueIndex;not null" json:"documentoIdentidad"`
	Email              string     `gorm:"size:150;uniqueIndex;not null" json:"email"`
	PasswordHash       string     `gorm:"size:255;not null" json:"-"`
	Rol                string     `gorm:"size:32;not null;index" json:"rol"`
	HospitalID         string     `gorm:"size:32;not null;index" json:"hospitalId"`
	IntentosFallidos   int        `gorm:"default:0;not null" json:"intentosFallidos"`
	BloqueadoHasta     *time.Time `json:"bloqueadoHasta,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

func (UsuarioPersonalORM) TableName() string { return "usuarios_personal" }

// UnidadSangreORM maps to the PostgreSQL table unidades_sangre.
type UnidadSangreORM struct {
	ID              string    `gorm:"primaryKey;size:36" json:"unidadId"`
	CodigoBarras    string    `gorm:"size:64;uniqueIndex;not null" json:"codigoBarras"`
	BancoID         string    `gorm:"size:32;not null;index:idx_banco_tipo_estado" json:"bancoId"`
	TipoSangre      string    `gorm:"size:4;not null;index:idx_banco_tipo_estado" json:"tipoSangre"`
	FactorRh        string    `gorm:"size:2;not null;index:idx_banco_tipo_estado" json:"factorRh"`
	VolumenMl       int       `gorm:"not null;default:450" json:"volumenMl"`
	Origen          string    `gorm:"size:64;not null" json:"origen"`
	FechaExtraccion time.Time `gorm:"type:date;not null" json:"fechaExtraccion"`
	FechaCaducidad  time.Time `gorm:"type:date;not null;index" json:"fechaCaducidad"`
	Estado          string    `gorm:"size:24;not null;index:idx_banco_tipo_estado" json:"estado"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (UnidadSangreORM) TableName() string { return "unidades_sangre" }

// SolicitudEmergenciaORM maps to the PostgreSQL table solicitudes_emergencia.
type SolicitudEmergenciaORM struct {
	ID                         string                  `gorm:"primaryKey;size:36" json:"solicitudId"`
	FechaHora                  time.Time               `gorm:"not null;index" json:"fechaHora"`
	HospitalOrigenID           string                  `gorm:"size:32;not null;index" json:"hospitalOrigenId"`
	CoordenadasOrigen          string                  `gorm:"size:64;not null" json:"coordenadasOrigen"`
	MedicoID                   uint                    `gorm:"not null;index" json:"medicoId"`
	TipoSangreReq              string                  `gorm:"size:4;not null" json:"tipoSangreReq"`
	FactorRhReq                string                  `gorm:"size:2;not null" json:"factorRhReq"`
	CantidadReq                int                     `gorm:"not null" json:"cantidadReq"`
	Urgencia                   string                  `gorm:"size:24;not null" json:"urgencia"`
	EstadoSolicitud            string                  `gorm:"size:32;not null;index" json:"estadoSolicitud"`
	OrigenAsignacion           string                  `gorm:"size:24;not null" json:"origenAsignacion"`
	BancoAsignadoID            string                  `gorm:"size:32" json:"bancoAsignadoId"`
	DistanciaKm                float64                 `json:"distanciaKm"`
	TiempoEstimadoMin          int                     `json:"tiempoEstimadoMin"`
	TiempoRespuestaMatchingMs  int64                   `json:"tiempoRespuestaMatchingMs"`
	FallbackJurisdiccionUsado  bool                    `gorm:"default:false;not null" json:"fallbackJurisdiccionUsado"`
	MotivoCancelacion          string                  `gorm:"size:255" json:"motivoCancelacion,omitempty"`
	UnidadesAsignadas          []SolicitudUnidadORM    `gorm:"foreignKey:SolicitudID;constraint:OnDelete:CASCADE" json:"unidadesAsignadas"`
	CreatedAt                  time.Time               `json:"createdAt"`
	UpdatedAt                  time.Time               `json:"updatedAt"`
}

func (SolicitudEmergenciaORM) TableName() string { return "solicitudes_emergencia" }

// SolicitudUnidadORM is the pivot table linking SolicitudEmergencia and UnidadSangre.
type SolicitudUnidadORM struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SolicitudID  string    `gorm:"size:36;not null;index" json:"solicitudId"`
	UnidadID     string    `gorm:"size:36;not null;index" json:"unidadId"`
	CodigoBarras string    `gorm:"size:64;not null" json:"codigoBarras"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (SolicitudUnidadORM) TableName() string { return "solicitud_unidades" }

// DonanteORM maps to the PostgreSQL table donantes.
type DonanteORM struct {
	ID                      string     `gorm:"primaryKey;size:36" json:"donanteId"`
	DNICifrado              string     `gorm:"type:text;not null" json:"-"`
	DNIHash                 string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	NombreCompleto          string     `gorm:"size:160;not null" json:"nombreCompleto"`
	CelularCifrado          string     `gorm:"type:text;not null" json:"-"`
	Correo                  string     `gorm:"size:160;not null" json:"correo"`
	TipoSangre              string     `gorm:"size:4;not null;index" json:"tipoSangre"`
	FactorRh                string     `gorm:"size:2;not null;index" json:"factorRh"`
	BancoPreferidoID        string     `gorm:"size:32;not null" json:"bancoPreferidoId"`
	ValidadoReniec          bool       `gorm:"default:true;not null" json:"validadoReniec"`
	HabilitadoConvocatorias bool       `gorm:"default:true;not null" json:"habilitadoConvocatorias"`
	Activo                  bool       `gorm:"default:true;not null;index" json:"activo"`
	FechaUltimaDonacion     *time.Time `json:"fechaUltimaDonacion,omitempty"`
	FechaRegistro           time.Time  `gorm:"not null" json:"fechaRegistro"`
	UpdatedAt               time.Time  `json:"updatedAt"`
}

func (DonanteORM) TableName() string { return "donantes" }

// AuditLogORM maps to the PostgreSQL table audit_logs.
type AuditLogORM struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Entidad        string    `gorm:"size:48;not null;index" json:"entidad"`
	EntidadID      string    `gorm:"size:64;not null;index" json:"entidadId"`
	Accion         string    `gorm:"size:64;not null" json:"accion"`
	EstadoAnterior string    `gorm:"size:48" json:"estadoAnterior"`
	EstadoNuevo    string    `gorm:"size:48" json:"estadoNuevo"`
	UsuarioID      string    `gorm:"size:64" json:"usuarioId"`
	Detalle        string    `gorm:"type:text" json:"detalle"`
	Timestamp      time.Time `gorm:"not null;index" json:"timestamp"`
}

func (AuditLogORM) TableName() string { return "audit_logs" }

// NotificacionLogORM maps to the PostgreSQL table notificacion_logs.
type NotificacionLogORM struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TipoEvento   string    `gorm:"size:64;not null;index" json:"evento"`
	Canal        string    `gorm:"size:32;not null" json:"canal"`
	Destinatario string    `gorm:"size:160;not null" json:"destinatario"`
	Mensaje      string    `gorm:"type:text;not null" json:"mensaje"`
	CreatedAt    time.Time `gorm:"not null;index" json:"timestamp"`
}

func (NotificacionLogORM) TableName() string { return "notificacion_logs" }
