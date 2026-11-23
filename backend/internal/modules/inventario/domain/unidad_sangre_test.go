package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNuevaUnidadSangre_TableDriven(t *testing.T) {
	hoy := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name            string
		tipoSangre      string
		factorRh        string
		volumenMl       int
		fechaExtraccion time.Time
		wantErr         error
		wantAlerta      string
	}{
		{
			name:            "Ingreso válido recién extraída (Alerta VERDE)",
			tipoSangre:      "O",
			factorRh:        "+",
			volumenMl:       450,
			fechaExtraccion: hoy.AddDate(0, 0, -5),
			wantErr:         nil,
			wantAlerta:      AlertaVerde,
		},
		{
			name:            "Ingreso válido con 37 días desde extracción (5 días restantes -> Alerta AMARILLO)",
			tipoSangre:      "A",
			factorRh:        "-",
			volumenMl:       0, // default 450
			fechaExtraccion: hoy.AddDate(0, 0, -37),
			wantErr:         nil,
			wantAlerta:      AlertaAmarillo,
		},
		{
			name:            "Ingreso válido con 40 días desde extracción (2 días restantes -> Alerta ROJO)",
			tipoSangre:      "B",
			factorRh:        "+",
			volumenMl:       450,
			fechaExtraccion: hoy.AddDate(0, 0, -40),
			wantErr:         nil,
			wantAlerta:      AlertaRojo,
		},
		{
			name:            "Bloqueo CUN-003 Flujo 2A: bolsa ya vencida (43 días desde extracción -> 422)",
			tipoSangre:      "O",
			factorRh:        "-",
			volumenMl:       450,
			fechaExtraccion: hoy.AddDate(0, 0, -43),
			wantErr:         ErrUnidadYaCaducada,
		},
		{
			name:            "Bloqueo por grupo sanguíneo inválido",
			tipoSangre:      "X",
			factorRh:        "+",
			volumenMl:       450,
			fechaExtraccion: hoy,
			wantErr:         ErrTipoSangreInvalido,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u, err := NuevaUnidadSangre("UND-2026-0001", "BL-MINSA01-0001", "HOSP-MINSA-01", tc.tipoSangre, tc.factorRh, "DONACION", tc.volumenMl, tc.fechaExtraccion, hoy)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("esperaba error %v, obtuvo %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if u.Estado != EstadoUnidadDisponible {
				t.Errorf("esperaba estado Disponible, obtuvo %s", u.Estado)
			}
			if u.AlertaCaducidad != tc.wantAlerta {
				t.Errorf("esperaba alerta %s, obtuvo %s", tc.wantAlerta, u.AlertaCaducidad)
			}
			if u.VolumenMl != 450 {
				t.Errorf("esperaba volumen 450, obtuvo %d", u.VolumenMl)
			}
		})
	}
}

func TestUnidadSangre_MaquinaDeEstados(t *testing.T) {
	tests := []struct {
		name          string
		estadoInicial string
		nuevoEstado   string
		wantErr       bool
	}{
		{"Disponible -> Reservada", EstadoUnidadDisponible, EstadoUnidadReservada, false},
		{"Disponible -> Caducada", EstadoUnidadDisponible, EstadoUnidadCaducada, false},
		{"Disponible -> Transfundida (inválido sin reserva)", EstadoUnidadDisponible, EstadoUnidadTransfundida, true},
		{"Reservada -> Transfundida", EstadoUnidadReservada, EstadoUnidadTransfundida, false},
		{"Reservada -> Disponible (liberación por cancelación)", EstadoUnidadReservada, EstadoUnidadDisponible, false},
		{"Transfundida -> Disponible (terminal)", EstadoUnidadTransfundida, EstadoUnidadDisponible, true},
		{"Caducada -> Reservada (terminal)", EstadoUnidadCaducada, EstadoUnidadReservada, true},
		{"Mismo estado idempotente", EstadoUnidadDisponible, EstadoUnidadDisponible, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			u := &UnidadSangre{ID: "UND-1", Estado: tc.estadoInicial}
			err := u.ActualizarEstado(tc.nuevoEstado)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ActualizarEstado(%s -> %s) error = %v, wantErr = %v", tc.estadoInicial, tc.nuevoEstado, err, tc.wantErr)
			}
		})
	}
}

func TestBancoSangre_ComportamientosYAbastecimiento(t *testing.T) {
	b := &BancoSangre{ID: "HOSP-MINSA-01", Nombre: "Santa Rosa"}
	unidades := []UnidadSangre{
		{ID: "1", BancoID: "HOSP-MINSA-01", TipoSangre: "O", FactorRh: "+", Estado: EstadoUnidadDisponible},
		{ID: "2", BancoID: "HOSP-MINSA-01", TipoSangre: "O", FactorRh: "+", Estado: EstadoUnidadReservada},
		{ID: "3", BancoID: "HOSP-MINSA-01", TipoSangre: "A", FactorRh: "-", Estado: EstadoUnidadDisponible},
	}

	if got := b.ConsultarStockLocal(unidades, "O", "+"); got != 1 {
		t.Errorf("ConsultarStockLocal O+ = %d, esperaba 1", got)
	}
	if msg := b.EmitirAlertaRed("SOL-2026-0801"); msg == "" {
		t.Error("EmitirAlertaRed no debe retornar vacío")
	}
	if niv := CalcularNivelAbastecimiento(2); niv != NivelCritico {
		t.Errorf("CalcularNivelAbastecimiento(2) = %s, esperaba CRITICO", niv)
	}
	if niv := CalcularNivelAbastecimiento(20); niv != NivelEstable {
		t.Errorf("CalcularNivelAbastecimiento(20) = %s, esperaba ESTABLE", niv)
	}
}
