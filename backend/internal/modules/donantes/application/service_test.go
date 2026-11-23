package application

import (
	"context"
	"errors"
	"testing"

	"github.com/bloodlink/backend/internal/modules/donantes/adapters/outbound/security"
	"github.com/bloodlink/backend/internal/modules/donantes/domain"
	"github.com/bloodlink/backend/internal/modules/donantes/ports"
	integDomain "github.com/bloodlink/backend/internal/modules/integracion/domain"
)

type mockDonanteRepo struct {
	byID   map[string]*domain.Donante
	byHash map[string]*domain.Donante
	seq    int64
}

func newMockDonanteRepo() *mockDonanteRepo {
	return &mockDonanteRepo{
		byID:   make(map[string]*domain.Donante),
		byHash: make(map[string]*domain.Donante),
		seq:    4410,
	}
}

func (m *mockDonanteRepo) ExistePorDNIHash(_ context.Context, dniHash string) (bool, error) {
	_, ok := m.byHash[dniHash]
	return ok, nil
}

func (m *mockDonanteRepo) Guardar(_ context.Context, d *domain.Donante) error {
	copyD := *d
	m.byID[d.ID] = &copyD
	m.byHash[d.DNIHash] = &copyD
	return nil
}

func (m *mockDonanteRepo) ListarActivos(_ context.Context, _, _ string) ([]domain.Donante, error) {
	var res []domain.Donante
	for _, d := range m.byID {
		if d.Activo {
			res = append(res, *d)
		}
	}
	return res, nil
}

func (m *mockDonanteRepo) ObtenerPorID(_ context.Context, id string) (*domain.Donante, error) {
	d, ok := m.byID[id]
	if !ok || !d.Activo {
		return nil, domain.ErrDonanteNoEncontrado
	}
	copyD := *d
	return &copyD, nil
}

func (m *mockDonanteRepo) Actualizar(_ context.Context, d *domain.Donante) error {
	copyD := *d
	m.byID[d.ID] = &copyD
	return nil
}

func (m *mockDonanteRepo) DarDeBajaLogica(_ context.Context, id string) error {
	d, ok := m.byID[id]
	if !ok {
		return domain.ErrDonanteNoEncontrado
	}
	d.Activo = false
	return nil
}

func (m *mockDonanteRepo) NextSequenceNumber(_ context.Context) (int64, error) {
	m.seq++
	return m.seq, nil
}

type mockReniec struct{}

func (m *mockReniec) ValidarIdentidad(_ context.Context, dni string) (*integDomain.CiudadanoReniec, error) {
	if dni == "76543210" {
		return &integDomain.CiudadanoReniec{
			DNI:             "76543210",
			DNIValido:       true,
			NombreCompleto:  "JUAN CARLOS PEREZ GOMEZ",
			FechaNacimiento: "1995-04-12",
			MayorEdad:       true,
		}, nil
	}
	return nil, integDomain.ErrDNINoEncontrado
}

func TestSrvRegistrarDonante_CP03_Completo(t *testing.T) {
	repo := newMockDonanteRepo()
	crypto, _ := security.NewAESGCMCryptoService("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	srv := NewSrvRegistrarDonante(repo, &mockReniec{}, crypto)
	ctx := context.Background()

	// 0. Rechazo si no acepta política de privacidad
	if _, err := srv.RegistrarDonante(ctx, ports.RegistrarDonanteRequest{DNI: "76543210", AceptaPoliticaPrivacidad: false}); !errors.Is(err, domain.ErrPoliticaPrivacidadRequerida) {
		t.Fatalf("esperaba ErrPoliticaPrivacidadRequerida, obtuvo %v", err)
	}

	// 1. Validar DNI vía SOAP RENIEC
	ciud, err := srv.ValidarDNIReniec(ctx, "76543210")
	if err != nil || ciud.NombreCompleto != "JUAN CARLOS PEREZ GOMEZ" {
		t.Fatalf("ValidarDNIReniec falló: %v", err)
	}

	// 2. Registrar donante con cifrado AES-256-GCM
	reg, err := srv.RegistrarDonante(ctx, ports.RegistrarDonanteRequest{
		DNI:                      "76543210",
		Celular:                  "+51987654321",
		Correo:                   "jperez@correo.pe",
		TipoSangre:               "O",
		FactorRh:                 "+",
		BancoPreferidoID:         "HOSP-MINSA-01",
		AceptaPoliticaPrivacidad: true,
	})
	if err != nil {
		t.Fatalf("RegistrarDonante error: %v", err)
	}

	// 3. Bloqueo de registro duplicado por DNIHash
	_, errDup := srv.RegistrarDonante(ctx, ports.RegistrarDonanteRequest{
		DNI:                      "76543210",
		Celular:                  "+51987654321",
		Correo:                   "jperez@correo.pe",
		TipoSangre:               "O",
		FactorRh:                 "+",
		AceptaPoliticaPrivacidad: true,
	})
	if !errors.Is(errDup, domain.ErrDonanteYaRegistrado) {
		t.Fatalf("esperaba ErrDonanteYaRegistrado, obtuvo %v", errDup)
	}

	// 4. Listar donantes enmascarados
	list, err := srv.ListarDonantes(ctx, "O", "+")
	if err != nil || len(list) != 1 || list[0].DNI != "****3210" {
		t.Fatalf("ListarDonantes debe enmascarar DNI como ****3210, obtuvo %+v", list)
	}

	// 5. Actualizar todos los campos de contacto y Dar de Baja
	habilitado := false
	act, err := srv.ActualizarContacto(ctx, reg.DonanteID, ports.ActualizarContactoDonanteRequest{
		Celular:                 "+51911223344",
		Correo:                  "nuevo@correo.pe",
		BancoPreferidoID:        "HOSP-MINSA-02",
		HabilitadoConvocatorias: &habilitado,
	})
	if err != nil || act.BancoPreferidoID != "HOSP-MINSA-02" || act.HabilitadoConvocatorias {
		t.Fatalf("ActualizarContacto error: %v, act=%+v", err, act)
	}
	if err := srv.DarDeBaja(ctx, reg.DonanteID); err != nil {
		t.Fatalf("DarDeBaja error: %v", err)
	}
}
