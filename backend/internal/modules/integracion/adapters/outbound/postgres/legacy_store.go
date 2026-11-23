package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bloodlink/backend/internal/models"
	"github.com/bloodlink/backend/internal/modules/integracion/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LegacyStockPostgresStore struct {
	db *gorm.DB
}

func NewLegacyStockPostgresStore(db *gorm.DB) *LegacyStockPostgresStore {
	return &LegacyStockPostgresStore{db: db}
}

func (s *LegacyStockPostgresStore) ReservarUnidadesLegacyEnBD(ctx context.Context, bancoID, tipoSangre, factorRh string, cantidad int) (*domain.ReservaLegacyResultado, error) {
	var unidades []models.UnidadSangreORM
	now := time.Now().Truncate(24 * time.Hour)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{
			Strength: "UPDATE",
			Options:  "SKIP LOCKED",
		}).
			Where("banco_id = ? AND tipo_sangre = ? AND factor_rh = ? AND estado = ? AND fecha_caducidad > ?",
				bancoID, tipoSangre, factorRh, "Disponible", now).
			Order("fecha_caducidad ASC").
			Limit(cantidad).
			Find(&unidades).Error; err != nil {
			return err
		}

		if len(unidades) < cantidad {
			return domain.ErrLegacyStockFaltante
		}

		ids := make([]string, 0, len(unidades))
		for _, u := range unidades {
			ids = append(ids, u.ID)
		}

		if err := tx.Model(&models.UnidadSangreORM{}).
			Where("id IN ?", ids).
			Updates(map[string]any{
				"estado":     "Reservada",
				"updated_at": time.Now(),
			}).Error; err != nil {
			return err
		}

		for _, u := range unidades {
			tx.Create(&models.AuditLogORM{
				Entidad:        "UnidadSangre",
				EntidadID:      u.ID,
				Accion:         "RESERVA_SOAP_LEGACY",
				EstadoAnterior: "Disponible",
				EstadoNuevo:    "Reservada",
				UsuarioID:      "SOAP_GATEWAY_ESSALUD",
				Detalle:        fmt.Sprintf("Reserva vía ServicioBloodLinkSOAP en %s", bancoID),
				Timestamp:      time.Now(),
			})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	unitIDs := make([]string, 0, len(unidades))
	barcodes := make([]string, 0, len(unidades))
	for _, u := range unidades {
		unitIDs = append(unitIDs, u.ID)
		barcodes = append(barcodes, u.CodigoBarras)
	}

	shortUUID := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:8])
	return &domain.ReservaLegacyResultado{
		Exito:             true,
		BancoID:           bancoID,
		CodigoTransaccion: fmt.Sprintf("RES-%d-%s", time.Now().Year(), shortUUID),
		UnidadesIDs:       unitIDs,
		CodigosBarras:     barcodes,
		Mensaje:           "Stock reservado exitosamente mediante adaptador SOAP 1.1",
	}, nil
}
