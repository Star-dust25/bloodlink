package domain

import (
	"testing"
	"time"
)

func TestUsuarioPersonal_BloqueoPorIntentosFallidos(t *testing.T) {
	now := time.Now()
	u := &UsuarioPersonal{
		Persona: Persona{ID: 12, NombreCompleto: "Dr. Carlos Mendoza"},
		Email:   "medico@bloodlink.gob.pe",
		Rol:     RolMedico,
	}

	if u.EstaBloqueado(now) {
		t.Fatal("usuario nuevo no debe estar bloqueado")
	}

	u.RegistrarFalloAutenticacion(now)
	u.RegistrarFalloAutenticacion(now)
	if u.EstaBloqueado(now) {
		t.Fatal("usuario con 2 fallos aún no debe bloquearse")
	}

	// 3er fallo consecutivo -> bloqueo por 15 minutos
	u.RegistrarFalloAutenticacion(now)
	if !u.EstaBloqueado(now) {
		t.Fatal("usuario con 3 fallos consecutivos debe estar bloqueado")
	}

	u.LimpiarIntentosFallidos()
	if u.EstaBloqueado(now) || u.IntentosFallidos != 0 {
		t.Fatal("LimpiarIntentosFallidos debe desbloquear la cuenta")
	}
}
