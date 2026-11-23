package ports

import (
	"context"

	"github.com/bloodlink/backend/internal/modules/auth/domain"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UsuarioDTO struct {
	ID             uint   `json:"id"`
	NombreCompleto string `json:"nombreCompleto"`
	Email          string `json:"email"`
	Rol            string `json:"rol"`
	HospitalID     string `json:"hospitalId"`
}

type LoginResponse struct {
	AccessToken string     `json:"accessToken"`
	TokenType   string     `json:"tokenType"`
	ExpiresIn   int        `json:"expiresIn"`
	Usuario     UsuarioDTO `json:"usuario"`
}

// IAuthService defines the inbound port for authentication and RBAC profile lookup.
type IAuthService interface {
	Autenticar(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	ObtenerPerfil(ctx context.Context, userID uint) (*UsuarioDTO, error)
}

// IUsuarioRepository defines the outbound persistence port for UsuarioPersonal.
type IUsuarioRepository interface {
	FindByEmail(ctx context.Context, email string) (*domain.UsuarioPersonal, error)
	FindByID(ctx context.Context, id uint) (*domain.UsuarioPersonal, error)
	UpdateSecurityState(ctx context.Context, user *domain.UsuarioPersonal) error
}
