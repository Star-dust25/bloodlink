package application

import (
	"context"
	"time"

	"github.com/bloodlink/backend/internal/middleware"
	"github.com/bloodlink/backend/internal/modules/auth/domain"
	"github.com/bloodlink/backend/internal/modules/auth/ports"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo               ports.IUsuarioRepository
	jwtSecret          string
	jwtExpirationHours int
}

func NewAuthService(repo ports.IUsuarioRepository, jwtSecret string, jwtExpirationHours int) *AuthService {
	if jwtExpirationHours <= 0 {
		jwtExpirationHours = 8
	}
	return &AuthService{
		repo:               repo,
		jwtSecret:          jwtSecret,
		jwtExpirationHours: jwtExpirationHours,
	}
}

func (s *AuthService) Autenticar(ctx context.Context, req ports.LoginRequest) (*ports.LoginResponse, error) {
	now := time.Now()
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil || user == nil {
		return nil, domain.ErrCredencialesInvalidas
	}

	if user.EstaBloqueado(now) {
		return nil, domain.ErrCuentaBloqueada
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		user.RegistrarFalloAutenticacion(now)
		_ = s.repo.UpdateSecurityState(ctx, user)
		if user.EstaBloqueado(now) {
			return nil, domain.ErrCuentaBloqueada
		}
		return nil, domain.ErrCredencialesInvalidas
	}

	if user.IntentosFallidos > 0 || user.BloqueadoHasta != nil {
		user.LimpiarIntentosFallidos()
		_ = s.repo.UpdateSecurityState(ctx, user)
	}

	expDuration := time.Duration(s.jwtExpirationHours) * time.Hour
	claims := middleware.JWTClaims{
		UserID:         user.ID,
		Rol:            user.Rol,
		HospitalID:     user.HospitalID,
		NombreCompleto: user.NombreCompleto,
		Email:          user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "bloodlink-minsa-iam",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, err
	}

	return &ports.LoginResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int(expDuration.Seconds()),
		Usuario: ports.UsuarioDTO{
			ID:             user.ID,
			NombreCompleto: user.NombreCompleto,
			Email:          user.Email,
			Rol:            user.Rol,
			HospitalID:     user.HospitalID,
		},
	}, nil
}

func (s *AuthService) ObtenerPerfil(ctx context.Context, userID uint) (*ports.UsuarioDTO, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return nil, domain.ErrUsuarioNoEncontrado
	}
	return &ports.UsuarioDTO{
		ID:             user.ID,
		NombreCompleto: user.NombreCompleto,
		Email:          user.Email,
		Rol:            user.Rol,
		HospitalID:     user.HospitalID,
	}, nil
}
