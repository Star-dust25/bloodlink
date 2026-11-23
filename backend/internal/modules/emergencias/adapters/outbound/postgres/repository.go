package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bloodlink/backend/internal/models"
	"github.com/bloodlink/backend/internal/modules/emergencias/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type EmergenciaRepository struct {
	db *gorm.DB
}

func NewEmergenciaRepository(db *gorm.DB) *EmergenciaRepository {
	return &EmergenciaRepository{db: db}
}

func (r *EmergenciaRepository) ObtenerBanco(ctx context.Context, bancoID string) (*domain.CandidatoBanco, error) {
	var b models.BancoSangreORM
	if err := r.db.WithContext(ctx).Where("id = ?", bancoID).First(&b).Error; err != nil {
		return nil, err
	}
	return &domain.CandidatoBanco{
		BancoID:      b.ID,
		Nombre:       b.Nombre,
		Coordenadas:  b.Coordenadas,
		Latitud:      b.Latitud,
		Longitud:     b.Longitud,
		Jurisdiccion: b.Jurisdiccion,
		EsLegacySoap: b.EsLegacySoap,
	}, nil
}

func (r *EmergenciaRepository) ListarCandidatosExternos(ctx context.Context, hospitalOrigenID, tipoSangre, factorRh string) ([]domain.CandidatoBanco, error) {
	var bancos []models.BancoSangreORM
	if err := r.db.WithContext(ctx).Where("id <> ?", hospitalOrigenID).Find(&bancos).Error; err != nil {
		return nil, err
	}

	now := time.Now().Truncate(24 * time.Hour)
	candidatos := make([]domain.CandidatoBanco, 0, len(bancos))

	for _, b := range bancos {
		var count int64
		r.db.WithContext(ctx).Model(&models.UnidadSangreORM{}).
			Where("banco_id = ? AND tipo_sangre = ? AND factor_rh = ? AND estado = ? AND fecha_caducidad > ?",
				b.ID, tipoSangre, factorRh, "Disponible", now).
			Count(&count)

		if count > 0 {
			candidatos = append(candidatos, domain.CandidatoBanco{
				BancoID:             b.ID,
				Nombre:              b.Nombre,
				Coordenadas:         b.Coordenadas,
				Latitud:             b.Latitud,
				Longitud:            b.Longitud,
				Jurisdiccion:        b.Jurisdiccion,
				EsLegacySoap:        b.EsLegacySoap,
				UnidadesDisponibles: int(count),
			})
		}
	}

	return candidatos, nil
}

// IntentarReservaAtomica utiliza SELECT ... FOR UPDATE SKIP LOCKED para reservar unidades sin colisión bajo alta concurrencia (RNF-01 / Riesgo R4).
func (r *EmergenciaRepository) IntentarReservaAtomica(ctx context.Context, bancoID, tipoSangre, factorRh string, cantidad int, usuarioID string) ([]domain.UnidadAsignada, error) {
	var unidades []models.UnidadSangreORM
	now := time.Now().Truncate(24 * time.Hour)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
			return domain.ErrStockNoDisponibleEnRed
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
				Accion:         "RESERVA_AUTOMATICA_MATCHING",
				EstadoAnterior: "Disponible",
				EstadoNuevo:    "Reservada",
				UsuarioID:      usuarioID,
				Detalle:        fmt.Sprintf("Bloqueo atómico FOR UPDATE SKIP LOCKED en %s", bancoID),
				Timestamp:      time.Now(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	res := make([]domain.UnidadAsignada, 0, len(unidades))
	for _, u := range unidades {
		res = append(res, domain.UnidadAsignada{
			UnidadID:     u.ID,
			CodigoBarras: u.CodigoBarras,
		})
	}
	return res, nil
}

func (r *EmergenciaRepository) GuardarSolicitud(ctx context.Context, sol *domain.SolicitudEmergencia, usuarioID, detalle string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		pivots := make([]models.SolicitudUnidadORM, 0, len(sol.UnidadesAsignadas))
		for _, u := range sol.UnidadesAsignadas {
			pivots = append(pivots, models.SolicitudUnidadORM{
				SolicitudID:  sol.ID,
				UnidadID:     u.UnidadID,
				CodigoBarras: u.CodigoBarras,
			})
		}

		orm := models.SolicitudEmergenciaORM{
			ID:                        sol.ID,
			FechaHora:                 sol.FechaHora,
			HospitalOrigenID:          sol.HospitalOrigenID,
			CoordenadasOrigen:         sol.CoordenadasOrigen,
			MedicoID:                  sol.MedicoID,
			TipoSangreReq:             sol.TipoSangreReq,
			FactorRhReq:               sol.FactorRhReq,
			CantidadReq:               sol.CantidadReq,
			Urgencia:                  sol.Urgencia,
			EstadoSolicitud:           sol.EstadoSolicitud,
			OrigenAsignacion:          sol.OrigenAsignacion,
			BancoAsignadoID:           sol.BancoAsignadoID,
			DistanciaKm:               sol.DistanciaKm,
			TiempoEstimadoMin:         sol.TiempoEstimadoMin,
			TiempoRespuestaMatchingMs: sol.TiempoRespuestaMatchingMs,
			FallbackJurisdiccionUsado: sol.FallbackJurisdiccionUsado,
			MotivoCancelacion:         sol.MotivoCancelacion,
			UnidadesAsignadas:         pivots,
		}
		if err := tx.Create(&orm).Error; err != nil {
			return err
		}

		return tx.Create(&models.AuditLogORM{
			Entidad:        "SolicitudEmergencia",
			EntidadID:      sol.ID,
			Accion:         "CREAR_SOLICITUD_Y_MATCHING",
			EstadoAnterior: "Registrada",
			EstadoNuevo:    sol.EstadoSolicitud,
			UsuarioID:      usuarioID,
			Detalle:        detalle,
			Timestamp:      time.Now(),
		}).Error
	})
}

func (r *EmergenciaRepository) ObtenerSolicitudPorID(ctx context.Context, solicitudID string) (*domain.SolicitudEmergencia, error) {
	var orm models.SolicitudEmergenciaORM
	if err := r.db.WithContext(ctx).Preload("UnidadesAsignadas").Where("id = ?", solicitudID).First(&orm).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSolicitudNoEncontrada
		}
		return nil, err
	}
	return r.toDomainSolicitud(ctx, &orm), nil
}

func (r *EmergenciaRepository) ListarSolicitudes(ctx context.Context, estado string) ([]domain.SolicitudEmergencia, error) {
	var orms []models.SolicitudEmergenciaORM
	q := r.db.WithContext(ctx).Preload("UnidadesAsignadas")
	if estado != "" {
		q = q.Where("estado_solicitud = ?", estado)
	}
	if err := q.Order("fecha_hora DESC").Find(&orms).Error; err != nil {
		return nil, err
	}

	res := make([]domain.SolicitudEmergencia, 0, len(orms))
	for i := range orms {
		res = append(res, *r.toDomainSolicitud(ctx, &orms[i]))
	}
	return res, nil
}

func (r *EmergenciaRepository) CancelarSolicitudYLiberarUnidades(ctx context.Context, sol *domain.SolicitudEmergencia, usuarioID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.SolicitudEmergenciaORM{}).
			Where("id = ?", sol.ID).
			Updates(map[string]any{
				"estado_solicitud":   sol.EstadoSolicitud,
				"motivo_cancelacion": sol.MotivoCancelacion,
				"updated_at":         time.Now(),
			}).Error; err != nil {
			return err
		}

		if len(sol.UnidadesAsignadas) > 0 {
			ids := make([]string, 0, len(sol.UnidadesAsignadas))
			for _, u := range sol.UnidadesAsignadas {
				ids = append(ids, u.UnidadID)
			}

			if err := tx.Model(&models.UnidadSangreORM{}).
				Where("id IN ?", ids).
				Updates(map[string]any{
					"estado":     "Disponible",
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}

			for _, u := range sol.UnidadesAsignadas {
				tx.Create(&models.AuditLogORM{
					Entidad:        "UnidadSangre",
					EntidadID:      u.UnidadID,
					Accion:         "LIBERAR_RESERVA_POR_CANCELACION",
					EstadoAnterior: "Reservada",
					EstadoNuevo:    "Disponible",
					UsuarioID:      usuarioID,
					Detalle:        fmt.Sprintf("Liberada por cancelación de emergencia %s: %s", sol.ID, sol.MotivoCancelacion),
					Timestamp:      time.Now(),
				})
			}
		}

		return tx.Create(&models.AuditLogORM{
			Entidad:        "SolicitudEmergencia",
			EntidadID:      sol.ID,
			Accion:         "CANCELAR_SOLICITUD",
			EstadoAnterior: "Reservada",
			EstadoNuevo:    "Cancelada",
			UsuarioID:      usuarioID,
			Detalle:        sol.MotivoCancelacion,
			Timestamp:      time.Now(),
		}).Error
	})
}

func (r *EmergenciaRepository) ConfirmarTransfusionYUnidades(ctx context.Context, sol *domain.SolicitudEmergencia, usuarioID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.SolicitudEmergenciaORM{}).
			Where("id = ?", sol.ID).
			Updates(map[string]any{
				"estado_solicitud": sol.EstadoSolicitud,
				"updated_at":       time.Now(),
			}).Error; err != nil {
			return err
		}

		if len(sol.UnidadesAsignadas) > 0 {
			ids := make([]string, 0, len(sol.UnidadesAsignadas))
			for _, u := range sol.UnidadesAsignadas {
				ids = append(ids, u.UnidadID)
			}

			if err := tx.Model(&models.UnidadSangreORM{}).
				Where("id IN ?", ids).
				Updates(map[string]any{
					"estado":     "Transfundida",
					"updated_at": time.Now(),
				}).Error; err != nil {
				return err
			}

			for _, u := range sol.UnidadesAsignadas {
				tx.Create(&models.AuditLogORM{
					Entidad:        "UnidadSangre",
					EntidadID:      u.UnidadID,
					Accion:         "CONFIRMAR_TRANSFUSION",
					EstadoAnterior: "Reservada",
					EstadoNuevo:    "Transfundida",
					UsuarioID:      usuarioID,
					Detalle:        fmt.Sprintf("Transfusión confirmada en emergencia %s", sol.ID),
					Timestamp:      time.Now(),
				})
			}
		}

		return tx.Create(&models.AuditLogORM{
			Entidad:        "SolicitudEmergencia",
			EntidadID:      sol.ID,
			Accion:         "CONFIRMAR_TRANSFUSION",
			EstadoAnterior: "Reservada",
			EstadoNuevo:    "Atendida",
			UsuarioID:      usuarioID,
			Detalle:        "Transfusión completada y unidades marcadas como Transfundida",
			Timestamp:      time.Now(),
		}).Error
	})
}

func (r *EmergenciaRepository) ContarStockDisponibleGlobalPorTipo(ctx context.Context, tipoSangre, factorRh string) (int, error) {
	var count int64
	now := time.Now().Truncate(24 * time.Hour)
	err := r.db.WithContext(ctx).Model(&models.UnidadSangreORM{}).
		Where("tipo_sangre = ? AND factor_rh = ? AND estado = ? AND fecha_caducidad > ?",
			tipoSangre, factorRh, "Disponible", now).
		Count(&count).Error
	return int(count), err
}

func (r *EmergenciaRepository) toDomainSolicitud(ctx context.Context, orm *models.SolicitudEmergenciaORM) *domain.SolicitudEmergencia {
	unidades := make([]domain.UnidadAsignada, 0, len(orm.UnidadesAsignadas))
	for _, u := range orm.UnidadesAsignadas {
		unidades = append(unidades, domain.UnidadAsignada{
			UnidadID:     u.UnidadID,
			CodigoBarras: u.CodigoBarras,
		})
	}
	bancoNombre := orm.BancoAsignadoID
	if orm.BancoAsignadoID != "" {
		var b models.BancoSangreORM
		if err := r.db.WithContext(ctx).Select("nombre").Where("id = ?", orm.BancoAsignadoID).First(&b).Error; err == nil {
			bancoNombre = b.Nombre
		}
	}
	return &domain.SolicitudEmergencia{
		ID:                        orm.ID,
		FechaHora:                 orm.FechaHora,
		HospitalOrigenID:          orm.HospitalOrigenID,
		CoordenadasOrigen:         orm.CoordenadasOrigen,
		MedicoID:                  orm.MedicoID,
		TipoSangreReq:             orm.TipoSangreReq,
		FactorRhReq:               orm.FactorRhReq,
		CantidadReq:               orm.CantidadReq,
		Urgencia:                  orm.Urgencia,
		EstadoSolicitud:           orm.EstadoSolicitud,
		OrigenAsignacion:          orm.OrigenAsignacion,
		BancoAsignadoID:           orm.BancoAsignadoID,
		BancoAsignadoNombre:       bancoNombre,
		DistanciaKm:               orm.DistanciaKm,
		TiempoEstimadoMin:         orm.TiempoEstimadoMin,
		TiempoRespuestaMatchingMs: orm.TiempoRespuestaMatchingMs,
		FallbackJurisdiccionUsado: orm.FallbackJurisdiccionUsado,
		MotivoCancelacion:         orm.MotivoCancelacion,
		UnidadesAsignadas:         unidades,
	}
}
