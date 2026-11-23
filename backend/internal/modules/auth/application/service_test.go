package application

import (
	"context"
	"errors"
	"testing"

	"github.com/bloodlink/backend/internal/modules/auth/domain"
	"github.com/bloodlink/backend/internal/modules/auth/ports"
	"golang.org/x/crypto/bcrypt"
)

type mockUsuarioRepo struct {
	user *domain.UsuarioPersonal
}

func (m *mockUsuarioRepo) FindByEmail(_ context.Context, email string) (*domain.UsuarioPersonal, error) {
	if m.user != nil && m.user.Email == email {
		return m.user, nil
	}
	return nil, domain.ErrUsuarioNoEncontrado
}

func (m *mockUsuarioRepo) FindByID(_ context.Context, id uint) (*domain.UsuarioPersonal, error) {
	if m.user != nil && m.user.ID == id {
		return m.user, nil
	}
	return nil, domain.ErrUsuarioNoEncontrado
}

func (m *mockUsuarioRepo) UpdateSecurityState(_ context.Context, user *domain.UsuarioPersonal) error {
	m.user = user
	return nil
}

func TestAuthService_AutenticarYBloqueo(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("Medico123!"), bcrypt.MinCost)
	repo := &mockUsuarioRepo{
		user: &domain.UsuarioPersonal{
			Persona:      domain.Persona{ID: 12, NombreCompleto: "Dr. Carlos Mendoza"},
			Email:        "medico@bloodlink.gob.pe",
			PasswordHash: string(hash),
			Rol:          domain.RolMedico,
			HospitalID:   "HOSP-MINSA-01",
		},
	}

	svc := NewAuthService(repo, "test_secret_jwt_2026", 0) // tests default expiration fallback
	ctx := context.Background()

	// 0. Usuario inexistente
	if _, err := svc.Autenticar(ctx, ports.LoginRequest{Email: "noexiste@b.pe", Password: "123"}); !errors.Is(err, domain.ErrCredencialesInvalidas) {
		t.Fatalf("esperaba ErrCredencialesInvalidas para email inexistente, obtuvo %v", err)
	}

	// 1. 1 fallo y luego login exitoso -> limpia intentos fallidos
	_, _ = svc.Autenticar(ctx, ports.LoginRequest{Email: "medico@bloodlink.gob.pe", Password: "wrong"})
	resp, err := svc.Autenticar(ctx, ports.LoginRequest{
		Email:    "medico@bloodlink.gob.pe",
		Password: "Medico123!",
	})
	if err != nil || resp.AccessToken == "" || resp.Usuario.Rol != "MEDICO" {
		t.Fatalf("Autenticar exitoso falló: %v", err)
	}

	// 2. ObtenerPerfil exitoso y no encontrado
	perfil, err := svc.ObtenerPerfil(ctx, 12)
	if err != nil || perfil.Email != "medico@bloodlink.gob.pe" {
		t.Fatalf("ObtenerPerfil falló: %v", err)
	}
	if _, err := svc.ObtenerPerfil(ctx, 999); !errors.Is(err, domain.ErrUsuarioNoEncontrado) {
		t.Fatalf("esperaba ErrUsuarioNoEncontrado, obtuvo %v", err)
	}

	// 3. 3 intentos fallidos bloquean la cuenta, y el 4to intento con cuenta ya bloqueada retorna ErrCuentaBloqueada
	_, _ = svc.Autenticar(ctx, ports.LoginRequest{Email: "medico@bloodlink.gob.pe", Password: "wrong"})
	_, _ = svc.Autenticar(ctx, ports.LoginRequest{Email: "medico@bloodlink.gob.pe", Password: "wrong"})
	_, err3 := svc.Autenticar(ctx, ports.LoginRequest{Email: "medico@bloodlink.gob.pe", Password: "wrong"})
	if !errors.Is(err3, domain.ErrCuentaBloqueada) {
		t.Fatalf("esperaba ErrCuentaBloqueada al 3er fallo, obtuvo %v", err3)
	}
	_, err4 := svc.Autenticar(ctx, ports.LoginRequest{Email: "medico@bloodlink.gob.pe", Password: "Medico123!"})
	if !errors.Is(err4, domain.ErrCuentaBloqueada) {
		t.Fatalf("esperaba ErrCuentaBloqueada estando bloqueado, obtuvo %v", err4)
	}
}
