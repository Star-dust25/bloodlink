package ports

import (
	"context"

	"github.com/bloodlink/backend/internal/modules/donantes/domain"
	integDomain "github.com/bloodlink/backend/internal/modules/integracion/domain"
)

type ValidarDNIRequest struct {
	DNI string `json:"dni" binding:"required"`
}

type RegistrarDonanteRequest struct {
	DNI                      string `json:"dni" binding:"required"`
	Celular                  string `json:"celular" binding:"required"`
	Correo                   string `json:"correo" binding:"required,email"`
	TipoSangre               string `json:"tipoSangre" binding:"required"`
	FactorRh                 string `json:"factorRh" binding:"required"`
	BancoPreferidoID         string `json:"bancoPreferidoId"`
	AceptaPoliticaPrivacidad bool   `json:"aceptaPoliticaPrivacidad"`
}

type ActualizarContactoDonanteRequest struct {
	Celular                 string `json:"celular"`
	Correo                  string `json:"correo"`
	BancoPreferidoID        string `json:"bancoPreferidoId"`
	HabilitadoConvocatorias *bool  `json:"habilitadoConvocatorias"`
}

type DonanteResponseDTO struct {
	DonanteID               string `json:"donanteId"`
	DNI                     string `json:"dni"`
	NombreCompleto          string `json:"nombreCompleto"`
	Celular                 string `json:"celular,omitempty"`
	Correo                  string `json:"correo,omitempty"`
	TipoSangre              string `json:"tipoSangre"`
	FactorRh                string `json:"factorRh"`
	BancoPreferidoID        string `json:"bancoPreferidoId"`
	ValidadoReniec          bool   `json:"validadoReniec"`
	HabilitadoConvocatorias bool   `json:"habilitadoConvocatorias"`
	FechaRegistro           string `json:"fechaRegistro"`
}

// ICryptoService defines the outbound port for AES-256-GCM symmetric encryption and SHA-256 hashing (RNF-04).
type ICryptoService interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertextBase64 string) (string, error)
	HashSHA256(plaintext string) string
}

// IDonanteService defines the inbound port for MS-Donantes.
type IDonanteService interface {
	ValidarDNIReniec(ctx context.Context, dni string) (*integDomain.CiudadanoReniec, error)
	RegistrarDonante(ctx context.Context, req RegistrarDonanteRequest) (*DonanteResponseDTO, error)
	ListarDonantes(ctx context.Context, tipoSangre, factorRh string) ([]DonanteResponseDTO, error)
	ObtenerPerfilDonante(ctx context.Context, donanteID string) (*DonanteResponseDTO, error)
	ActualizarContacto(ctx context.Context, donanteID string, req ActualizarContactoDonanteRequest) (*DonanteResponseDTO, error)
	DarDeBaja(ctx context.Context, donanteID string) error
}

// IDonanteRepository defines the outbound persistence port for Donante.
type IDonanteRepository interface {
	ExistePorDNIHash(ctx context.Context, dniHash string) (bool, error)
	Guardar(ctx context.Context, d *domain.Donante) error
	ListarActivos(ctx context.Context, tipoSangre, factorRh string) ([]domain.Donante, error)
	ObtenerPorID(ctx context.Context, id string) (*domain.Donante, error)
	Actualizar(ctx context.Context, d *domain.Donante) error
	DarDeBajaLogica(ctx context.Context, id string) error
}
