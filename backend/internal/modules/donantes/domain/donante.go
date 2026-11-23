package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrPoliticaPrivacidadRequerida = errors.New("debe aceptar la política de privacidad y tratamiento de datos personales (Ley N° 29733)")
	ErrDonanteYaRegistrado         = errors.New("el ciudadano con ese DNI ya se encuentra registrado como donante voluntario")
	ErrDonanteNoEncontrado         = errors.New("donante no encontrado")
	ErrDatosDonanteInvalidos       = errors.New("datos de registro de donante inválidos")
)

type Persona struct {
	ID                 string `json:"id"`
	NombreCompleto     string `json:"nombreCompleto"`
	DocumentoIdentidad string `json:"dni"`
}

type Donante struct {
	Persona
	DNICifrado              string     `json:"-"`
	DNIHash                 string     `json:"-"`
	Celular                 string     `json:"celular"`
	CelularCifrado          string     `json:"-"`
	Correo                  string     `json:"correo"`
	TipoSangre              string     `json:"tipoSangre"`
	FactorRh                string     `json:"factorRh"`
	BancoPreferidoID        string     `json:"bancoPreferidoId"`
	ValidadoReniec          bool       `json:"validadoReniec"`
	HabilitadoConvocatorias bool       `json:"habilitadoConvocatorias"`
	Activo                  bool       `json:"activo"`
	FechaUltimaDonacion     *time.Time `json:"fechaUltimaDonacion,omitempty"`
	FechaRegistro           time.Time  `json:"fechaRegistro"`
}

func NuevoDonante(
	id, dni, nombreCompleto, celular, correo, tipoSangre, factorRh, bancoPreferidoID string,
	aceptaPrivacidad bool,
	ahora time.Time,
) (*Donante, error) {
	if !aceptaPrivacidad {
		return nil, ErrPoliticaPrivacidadRequerida
	}
	if len(strings.TrimSpace(dni)) != 8 || strings.TrimSpace(celular) == "" || strings.TrimSpace(correo) == "" {
		return nil, ErrDatosDonanteInvalidos
	}
	if (tipoSangre != "O" && tipoSangre != "A" && tipoSangre != "B" && tipoSangre != "AB") || (factorRh != "+" && factorRh != "-") {
		return nil, ErrDatosDonanteInvalidos
	}
	if bancoPreferidoID == "" {
		bancoPreferidoID = "HOSP-MINSA-01"
	}

	return &Donante{
		Persona: Persona{
			ID:                 id,
			NombreCompleto:     nombreCompleto,
			DocumentoIdentidad: dni,
		},
		Celular:                 celular,
		Correo:                  correo,
		TipoSangre:              tipoSangre,
		FactorRh:                factorRh,
		BancoPreferidoID:        bancoPreferidoID,
		ValidadoReniec:          true,
		HabilitadoConvocatorias: true,
		Activo:                  true,
		FechaRegistro:           ahora,
	}, nil
}

func EnmascararDNI(dni string) string {
	dni = strings.TrimSpace(dni)
	if len(dni) <= 4 {
		return "****"
	}
	return "****" + dni[len(dni)-4:]
}

func EnmascararCelular(celular string) string {
	celular = strings.TrimSpace(celular)
	if len(celular) <= 4 {
		return "****"
	}
	if strings.HasPrefix(celular, "+51") && len(celular) >= 8 {
		return celular[:4] + "****" + celular[len(celular)-4:]
	}
	return "****" + celular[len(celular)-4:]
}
