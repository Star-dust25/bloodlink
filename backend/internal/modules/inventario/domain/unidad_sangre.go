package domain

import (
	"errors"
	"fmt"
	"math"
	"time"
)

const (
	EstadoUnidadDisponible   = "Disponible"
	EstadoUnidadReservada    = "Reservada"
	EstadoUnidadTransfundida = "Transfundida"
	EstadoUnidadCaducada     = "Caducada"

	AlertaVerde    = "VERDE"
	AlertaAmarillo = "AMARILLO"
	AlertaRojo     = "ROJO"

	NivelEstable = "ESTABLE"
	NivelCritico = "CRITICO"

	DiasVigenciaSangreTotal = 42
	DiasUmbralAlertaAmarilla = 7
	DiasUmbralAlertaRoja     = 3
	CapacidadBaseRedPorTipo  = 50
	UmbralPorcentajeCritico  = 10.0
)

var (
	ErrUnidadYaCaducada        = errors.New("la unidad de sangre ya superó los 42 días de vigencia y está vencida (CUN-003 Flujo 2A)")
	ErrTransicionEstadoInvalida = errors.New("transición de estado no permitida por la máquina de estados de UnidadSangre")
	ErrTipoSangreInvalido      = errors.New("tipo de sangre o factor Rh inválido")
	ErrUnidadNoEncontrada      = errors.New("unidad de sangre no encontrada")
	ErrBancoNoEncontrado       = errors.New("banco de sangre no encontrado")
)

type BancoSangre struct {
	ID              string  `json:"id"`
	Nombre          string  `json:"nombre"`
	Direccion       string  `json:"direccion"`
	Latitud         float64 `json:"latitud"`
	Longitud        float64 `json:"longitud"`
	Coordenadas     string  `json:"coordenadas"`
	Jurisdiccion    string  `json:"jurisdiccion"`
	TipoInstitucion string  `json:"tipoInstitucion"`
	EsLegacySoap    bool    `json:"esLegacySoap"`
}

func (b *BancoSangre) ConsultarStockLocal(unidades []UnidadSangre, tipoSangre, factorRh string) int {
	count := 0
	for _, u := range unidades {
		if u.BancoID == b.ID && u.Estado == EstadoUnidadDisponible {
			if (tipoSangre == "" || u.TipoSangre == tipoSangre) && (factorRh == "" || u.FactorRh == factorRh) {
				count++
			}
		}
	}
	return count
}

func (b *BancoSangre) EmitirAlertaRed(solicitudID string) string {
	return fmt.Sprintf("ALERTA_RED_EMITIDA[%s]: Solicitud %s requiere soporte inter-hospitalario", b.ID, solicitudID)
}

type UnidadSangre struct {
	ID              string    `json:"unidadId"`
	CodigoBarras    string    `json:"codigoBarras"`
	BancoID         string    `json:"bancoId"`
	TipoSangre      string    `json:"tipoSangre"`
	FactorRh        string    `json:"factorRh"`
	VolumenMl       int       `json:"volumenMl"`
	Origen          string    `json:"origen"`
	FechaExtraccion time.Time `json:"fechaExtraccion"`
	FechaCaducidad  time.Time `json:"fechaCaducidad"`
	Estado          string    `json:"estado"`
	AlertaCaducidad string    `json:"alertaCaducidad,omitempty"`
}

func ValidarGrupoSanguineo(tipoSangre, factorRh string) bool {
	validTipos := map[string]bool{"O": true, "A": true, "B": true, "AB": true}
	validRh := map[string]bool{"+": true, "-": true}
	return validTipos[tipoSangre] && validRh[factorRh]
}

// NuevaUnidadSangre aplica la regla CUN-003 (vigencia 42 días y bloqueo 422 si está vencida al ingreso).
func NuevaUnidadSangre(id, codigoBarras, bancoID, tipoSangre, factorRh, origen string, volumenMl int, fechaExtraccion, ahora time.Time) (*UnidadSangre, error) {
	if !ValidarGrupoSanguineo(tipoSangre, factorRh) {
		return nil, ErrTipoSangreInvalido
	}
	if volumenMl <= 0 {
		volumenMl = 450
	}

	extDate := fechaExtraccion.Truncate(24 * time.Hour)
	hoy := ahora.Truncate(24 * time.Hour)
	fechaCaducidad := extDate.AddDate(0, 0, DiasVigenciaSangreTotal)

	if !fechaCaducidad.After(hoy) {
		return nil, ErrUnidadYaCaducada
	}

	u := &UnidadSangre{
		ID:              id,
		CodigoBarras:    codigoBarras,
		BancoID:         bancoID,
		TipoSangre:      tipoSangre,
		FactorRh:        factorRh,
		VolumenMl:       volumenMl,
		Origen:          origen,
		FechaExtraccion: extDate,
		FechaCaducidad:  fechaCaducidad,
		Estado:          EstadoUnidadDisponible,
	}
	u.AlertaCaducidad = u.VerificarCaducidad(hoy)
	return u, nil
}

// VerificarCaducidad retorna el semáforo de alerta según los días restantes (Sección 7.7.5b / RF-08).
func (u *UnidadSangre) VerificarCaducidad(fechaActual time.Time) string {
	hoy := fechaActual.Truncate(24 * time.Hour)
	cad := u.FechaCaducidad.Truncate(24 * time.Hour)
	diasRestantes := int(math.Ceil(cad.Sub(hoy).Hours() / 24.0))

	if diasRestantes <= DiasUmbralAlertaRoja {
		return AlertaRojo
	}
	if diasRestantes <= DiasUmbralAlertaAmarilla {
		return AlertaAmarillo
	}
	return AlertaVerde
}

func (u *UnidadSangre) DiasParaCaducar(fechaActual time.Time) int {
	hoy := fechaActual.Truncate(24 * time.Hour)
	cad := u.FechaCaducidad.Truncate(24 * time.Hour)
	return int(math.Ceil(cad.Sub(hoy).Hours() / 24.0))
}

// ActualizarEstado valida la máquina de estados estricta de UnidadSangre (Sección 7.7.5b).
func (u *UnidadSangre) ActualizarEstado(nuevoEstado string) error {
	if u.Estado == nuevoEstado {
		return nil
	}

	switch u.Estado {
	case EstadoUnidadDisponible:
		if nuevoEstado == EstadoUnidadReservada || nuevoEstado == EstadoUnidadCaducada {
			u.Estado = nuevoEstado
			return nil
		}
	case EstadoUnidadReservada:
		if nuevoEstado == EstadoUnidadTransfundida || nuevoEstado == EstadoUnidadDisponible {
			u.Estado = nuevoEstado
			return nil
		}
	case EstadoUnidadTransfundida, EstadoUnidadCaducada:
		// Estados terminales
		return fmt.Errorf("%w: '%s' es un estado terminal y no puede cambiar a '%s'", ErrTransicionEstadoInvalida, u.Estado, nuevoEstado)
	}

	return fmt.Errorf("%w: no se puede pasar de '%s' a '%s'", ErrTransicionEstadoInvalida, u.Estado, nuevoEstado)
}

func CalcularNivelAbastecimiento(unidadesDisponibles int) string {
	porcentaje := (float64(unidadesDisponibles) / float64(CapacidadBaseRedPorTipo)) * 100.0
	if porcentaje < UmbralPorcentajeCritico {
		return NivelCritico
	}
	return NivelEstable
}
