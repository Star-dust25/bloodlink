package postgres

import (
	"context"
	"math"
	"time"

	"github.com/bloodlink/backend/internal/models"
	"github.com/bloodlink/backend/internal/modules/notificaciones/domain"
	"gorm.io/gorm"
)

type NotificacionRepository struct {
	db *gorm.DB
}

func NewNotificacionRepository(db *gorm.DB) *NotificacionRepository {
	return &NotificacionRepository{db: db}
}

func (r *NotificacionRepository) GuardarNotificacion(ctx context.Context, n *domain.Notificacion) error {
	orm := models.NotificacionLogORM{
		TipoEvento:   n.Evento,
		Canal:        n.Canal,
		Destinatario: n.Destinatario,
		Mensaje:      n.Mensaje,
		CreatedAt:    n.Timestamp,
	}
	if orm.CreatedAt.IsZero() {
		orm.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(&orm).Error
}

func (r *NotificacionRepository) ListarNotificaciones(ctx context.Context, canal string, limit int) ([]domain.Notificacion, error) {
	var orms []models.NotificacionLogORM
	q := r.db.WithContext(ctx).Model(&models.NotificacionLogORM{})
	if canal != "" {
		q = q.Where("canal = ?", canal)
	}
	if err := q.Order("created_at DESC").Limit(limit).Find(&orms).Error; err != nil {
		return nil, err
	}

	res := make([]domain.Notificacion, 0, len(orms))
	for _, o := range orms {
		res = append(res, domain.Notificacion{
			ID:           o.ID,
			Evento:       o.TipoEvento,
			Canal:        o.Canal,
			Destinatario: o.Destinatario,
			Mensaje:      o.Mensaje,
			Timestamp:    o.CreatedAt,
		})
	}
	return res, nil
}

func (r *NotificacionRepository) BuscarDonantesCompatibles(ctx context.Context, tipoSangre, factorRh string) ([]domain.DonanteConvocable, error) {
	var orms []models.DonanteORM
	q := r.db.WithContext(ctx).Where("activo = ? AND habilitado_convocatorias = ?", true, true)
	if tipoSangre != "" && factorRh != "" {
		// Incluye donantes del mismo tipo o donantes universales O-
		q = q.Where("(tipo_sangre = ? AND factor_rh = ?) OR (tipo_sangre = 'O' AND factor_rh = '-')", tipoSangre, factorRh)
	}
	if err := q.Find(&orms).Error; err != nil {
		return nil, err
	}

	res := make([]domain.DonanteConvocable, 0, len(orms))
	for _, o := range orms {
		res = append(res, domain.DonanteConvocable{
			DonanteID:        o.ID,
			NombreCompleto:   o.NombreCompleto,
			Correo:           o.Correo,
			TipoSangre:       o.TipoSangre,
			FactorRh:         o.FactorRh,
			BancoPreferidoID: o.BancoPreferidoID,
		})
	}
	return res, nil
}

func (r *NotificacionRepository) BuscarUnidadesProximasACaducar7Dias(ctx context.Context) ([]domain.UnidadPorCaducar, error) {
	now := time.Now().Truncate(24 * time.Hour)
	limite7Dias := now.AddDate(0, 0, 7)

	var orms []models.UnidadSangreORM
	if err := r.db.WithContext(ctx).
		Where("estado = ? AND fecha_caducidad <= ?", "Disponible", limite7Dias).
		Order("fecha_caducidad ASC").
		Find(&orms).Error; err != nil {
		return nil, err
	}

	res := make([]domain.UnidadPorCaducar, 0, len(orms))
	for _, o := range orms {
		dias := int(math.Ceil(o.FechaCaducidad.Truncate(24*time.Hour).Sub(now).Hours() / 24.0))
		alerta := "AMARILLO"
		if dias <= 3 {
			alerta = "ROJO"
		}
		res = append(res, domain.UnidadPorCaducar{
			UnidadID:        o.ID,
			CodigoBarras:    o.CodigoBarras,
			BancoID:         o.BancoID,
			TipoSangre:      o.TipoSangre,
			FactorRh:        o.FactorRh,
			FechaCaducidad:  o.FechaCaducidad.Format("2006-01-02"),
			DiasRestantes:   dias,
			AlertaCaducidad: alerta,
		})
	}
	return res, nil
}
