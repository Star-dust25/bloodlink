package postgres

import (
	"context"
	"errors"

	"github.com/bloodlink/backend/internal/models"
	"github.com/bloodlink/backend/internal/modules/auth/domain"
	"gorm.io/gorm"
)

type UsuarioRepository struct {
	db *gorm.DB
}

func NewUsuarioRepository(db *gorm.DB) *UsuarioRepository {
	return &UsuarioRepository{db: db}
}

func (r *UsuarioRepository) FindByEmail(ctx context.Context, email string) (*domain.UsuarioPersonal, error) {
	var orm models.UsuarioPersonalORM
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&orm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUsuarioNoEncontrado
		}
		return nil, err
	}
	return toDomain(&orm), nil
}

func (r *UsuarioRepository) FindByID(ctx context.Context, id uint) (*domain.UsuarioPersonal, error) {
	var orm models.UsuarioPersonalORM
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&orm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUsuarioNoEncontrado
		}
		return nil, err
	}
	return toDomain(&orm), nil
}

func (r *UsuarioRepository) UpdateSecurityState(ctx context.Context, user *domain.UsuarioPersonal) error {
	return r.db.WithContext(ctx).Model(&models.UsuarioPersonalORM{}).
		Where("id = ?", user.ID).
		Updates(map[string]any{
			"intentos_fallidos": user.IntentosFallidos,
			"bloqueado_hasta":   user.BloqueadoHasta,
		}).Error
}

func toDomain(orm *models.UsuarioPersonalORM) *domain.UsuarioPersonal {
	return &domain.UsuarioPersonal{
		Persona: domain.Persona{
			ID:                 orm.ID,
			NombreCompleto:     orm.NombreCompleto,
			DocumentoIdentidad: orm.DocumentoIdentidad,
		},
		Email:            orm.Email,
		PasswordHash:     orm.PasswordHash,
		Rol:              orm.Rol,
		HospitalID:       orm.HospitalID,
		IntentosFallidos: orm.IntentosFallidos,
		BloqueadoHasta:   orm.BloqueadoHasta,
	}
}
