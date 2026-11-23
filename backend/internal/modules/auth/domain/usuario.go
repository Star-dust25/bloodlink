package domain

import (
	"errors"
	"time"
)

const (
	RolMedico        = "MEDICO"
	RolPersonalBanco = "PERSONAL_BANCO"
	RolCoordinador   = "COORDINADOR"

	MaxIntentosFallidos = 3
	MinutosBloqueoCuenta = 15
)

var (
	ErrCredencialesInvalidas = errors.New("credenciales inválidas")
	ErrCuentaBloqueada       = errors.New("cuenta bloqueada temporalmente por múltiples intentos fallidos")
	ErrUsuarioNoEncontrado   = errors.New("usuario no encontrado")
)

type Persona struct {
	ID                 uint   `json:"id"`
	NombreCompleto     string `json:"nombreCompleto"`
	DocumentoIdentidad string `json:"documentoIdentidad"`
}

type UsuarioPersonal struct {
	Persona
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	Rol              string     `json:"rol"`
	HospitalID       string     `json:"hospitalId"`
	IntentosFallidos int        `json:"intentosFallidos"`
	BloqueadoHasta   *time.Time `json:"bloqueadoHasta,omitempty"`
}

func (u *UsuarioPersonal) EstaBloqueado(ahora time.Time) bool {
	return u.BloqueadoHasta != nil && u.BloqueadoHasta.After(ahora)
}

func (u *UsuarioPersonal) RegistrarFalloAutenticacion(ahora time.Time) {
	u.IntentosFallidos++
	if u.IntentosFallidos >= MaxIntentosFallidos {
		hasta := ahora.Add(MinutosBloqueoCuenta * time.Minute)
		u.BloqueadoHasta = &hasta
	}
}

func (u *UsuarioPersonal) LimpiarIntentosFallidos() {
	u.IntentosFallidos = 0
	u.BloqueadoHasta = nil
}
