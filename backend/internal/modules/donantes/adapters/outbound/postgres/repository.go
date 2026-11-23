package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/bloodlink/backend/internal/models"
	"github.com/bloodlink/backend/internal/modules/donantes/domain"
	"gorm.io/gorm"
)

type DonanteRepository struct {
	db *gorm.DB
}

func NewDonanteRepository(db *gorm.DB) *DonanteRepository {
	return &DonanteRepository{db: db}
}

func (r *DonanteRepository) ExistePorDNIHash(ctx context.Context, dniHash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.DonanteORM{}).Where("dni_hash = ?", dniHash).Count(&count).Error
	return count > 0, err
}

func (r *DonanteRepository) Guardar(ctx context.Context, d *domain.Donante) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		orm := models.DonanteORM{
			ID:                      d.ID,
			DNICifrado:              d.DNICifrado,
			DNIHash:                 d.DNIHash,
			NombreCompleto:          d.NombreCompleto,
			CelularCifrado:          d.CelularCifrado,
			Correo:                  d.Correo,
			TipoSangre:              d.TipoSangre,
			FactorRh:                d.FactorRh,
			BancoPreferidoID:        d.BancoPreferidoID,
			ValidadoReniec:          d.ValidadoReniec,
			HabilitadoConvocatorias: d.HabilitadoConvocatorias,
			Activo:                  d.Activo,
			FechaUltimaDonacion:     d.FechaUltimaDonacion,
			FechaRegistro:           d.FechaRegistro,
		}
		if err := tx.Create(&orm).Error; err != nil {
			return err
		}

		return tx.Create(&models.AuditLogORM{
			Entidad:        "Donante",
			EntidadID:      d.ID,
			Accion:         "REGISTRAR_DONANTE_RENIEC_AES256",
			EstadoAnterior: "NINGUNO",
			EstadoNuevo:    "Habilitado",
			UsuarioID:      "PORTAL_PUBLICO",
			Detalle:        "Registro validado con SOAP RENIEC y datos sensibles cifrados con AES-256-GCM",
			Timestamp:      time.Now(),
		}).Error
	})
}

func (r *DonanteRepository) ListarActivos(ctx context.Context, tipoSangre, factorRh string) ([]domain.Donante, error) {
	var orms []models.DonanteORM
	q := r.db.WithContext(ctx).Where("activo = ?", true)
	if tipoSangre != "" {
		q = q.Where("tipo_sangre = ?", tipoSangre)
	}
	if factorRh != "" {
		q = q.Where("factor_rh = ?", factorRh)
	}
	if err := q.Order("fecha_registro DESC").Find(&orms).Error; err != nil {
		return nil, err
	}

	res := make([]domain.Donante, 0, len(orms))
	for _, o := range orms {
		res = append(res, *toDomainDonante(&o))
	}
	return res, nil
}

func (r *DonanteRepository) ObtenerPorID(ctx context.Context, id string) (*domain.Donante, error) {
	var orm models.DonanteORM
	if err := r.db.WithContext(ctx).Where("id = ? AND activo = ?", id, true).First(&orm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrDonanteNoEncontrado
		}
		return nil, err
	}
	return toDomainDonante(&orm), nil
}

func (r *DonanteRepository) Actualizar(ctx context.Context, d *domain.Donante) error {
	return r.db.WithContext(ctx).Model(&models.DonanteORM{}).
		Where("id = ?", d.ID).
		Updates(map[string]any{
			"celular_cifrado":          d.CelularCifrado,
			"correo":                   d.Correo,
			"banco_preferido_id":       d.BancoPreferidoID,
			"habilitado_convocatorias": d.HabilitadoConvocatorias,
			"updated_at":               time.Now(),
		}).Error
}

func (r *DonanteRepository) DarDeBajaLogica(ctx context.Context, id string) error {
	res := r.db.WithContext(ctx).Model(&models.DonanteORM{}).
		Where("id = ? AND activo = ?", id, true).
		Updates(map[string]any{
			"activo":                   false,
			"habilitado_convocatorias": false,
			"updated_at":               time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrDonanteNoEncontrado
	}
	return nil
}

func toDomainDonante(o *models.DonanteORM) *domain.Donante {
	return &domain.Donante{
		Persona: domain.Persona{
			ID:             o.ID,
			NombreCompleto: o.NombreCompleto,
		},
		DNICifrado:              o.DNICifrado,
		DNIHash:                 o.DNIHash,
		CelularCifrado:          o.CelularCifrado,
		Correo:                  o.Correo,
		TipoSangre:              o.TipoSangre,
		FactorRh:                o.FactorRh,
		BancoPreferidoID:        o.BancoPreferidoID,
		ValidadoReniec:          o.ValidadoReniec,
		HabilitadoConvocatorias: o.HabilitadoConvocatorias,
		Activo:                  o.Activo,
		FechaUltimaDonacion:     o.FechaUltimaDonacion,
		FechaRegistro:           o.FechaRegistro,
	}
}
