package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNuevoDonante_Y_Enmascaramiento(t *testing.T) {
	now := time.Now()

	if _, err := NuevoDonante("DON-1", "76543210", "JUAN PEREZ", "+51987654321", "j@p.pe", "O", "+", "HOSP-MINSA-01", false, now); !errors.Is(err, ErrPoliticaPrivacidadRequerida) {
		t.Fatalf("esperaba ErrPoliticaPrivacidadRequerida, obtuvo %v", err)
	}

	if _, err := NuevoDonante("DON-1", "123", "JUAN PEREZ", "+51987654321", "j@p.pe", "O", "+", "HOSP-MINSA-01", true, now); !errors.Is(err, ErrDatosDonanteInvalidos) {
		t.Fatalf("esperaba ErrDatosDonanteInvalidos por DNI corto, obtuvo %v", err)
	}

	if _, err := NuevoDonante("DON-1", "76543210", "JUAN PEREZ", "+51987654321", "j@p.pe", "Z", "+", "HOSP-MINSA-01", true, now); !errors.Is(err, ErrDatosDonanteInvalidos) {
		t.Fatalf("esperaba ErrDatosDonanteInvalidos por tipo sangre inválido, obtuvo %v", err)
	}

	d, err := NuevoDonante("DON-2026-4410", "76543210", "JUAN CARLOS PEREZ GOMEZ", "+51987654321", "jperez@correo.pe", "O", "+", "", true, now)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if d.BancoPreferidoID != "HOSP-MINSA-01" || !d.ValidadoReniec || !d.HabilitadoConvocatorias {
		t.Errorf("valores por defecto incorrectos: %+v", d)
	}

	if got := EnmascararDNI("76543210"); got != "****3210" {
		t.Errorf("EnmascararDNI = %s, esperaba ****3210", got)
	}
	if got := EnmascararDNI("123"); got != "****" {
		t.Errorf("EnmascararDNI corto = %s, esperaba ****", got)
	}
	if got := EnmascararCelular("+51987654321"); got != "+519****4321" {
		t.Errorf("EnmascararCelular = %s, esperaba +519****4321", got)
	}
	if got := EnmascararCelular("987654321"); got != "****4321" {
		t.Errorf("EnmascararCelular local = %s, esperaba ****4321", got)
	}
	if got := EnmascararCelular("123"); got != "****" {
		t.Errorf("EnmascararCelular corto = %s, esperaba ****", got)
	}
}
