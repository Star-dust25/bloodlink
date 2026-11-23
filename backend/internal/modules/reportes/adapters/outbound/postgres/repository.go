package postgres

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/bloodlink/backend/internal/models"
	"github.com/bloodlink/backend/internal/modules/reportes/domain"
	"gorm.io/gorm"
)

type ReporteRepository struct {
	db *gorm.DB
}

func NewReporteRepository(db *gorm.DB) *ReporteRepository {
	return &ReporteRepository{db: db}
}

func (r *ReporteRepository) ConstruirMetricasDashboard(ctx context.Context) (*domain.DashboardMetricas, error) {
	now := time.Now().Truncate(24 * time.Hour)
	limite7d := now.AddDate(0, 0, 7)
	limite3d := now.AddDate(0, 0, 3)

	var totalBancos int64
	r.db.WithContext(ctx).Model(&models.BancoSangreORM{}).Count(&totalBancos)

	var totalDisponibles int64
	r.db.WithContext(ctx).Model(&models.UnidadSangreORM{}).
		Where("estado = ? AND fecha_caducidad > ?", "Disponible", now).
		Count(&totalDisponibles)

	var porCaducar7d int64
	r.db.WithContext(ctx).Model(&models.UnidadSangreORM{}).
		Where("estado = ? AND fecha_caducidad <= ?", "Disponible", limite7d).
		Count(&porCaducar7d)

	var alertaRoja3d int64
	r.db.WithContext(ctx).Model(&models.UnidadSangreORM{}).
		Where("estado = ? AND fecha_caducidad <= ?", "Disponible", limite3d).
		Count(&alertaRoja3d)

	var totalEmergencias, emergenciasActivas, emergenciasAtendidas int64
	r.db.WithContext(ctx).Model(&models.SolicitudEmergenciaORM{}).Count(&totalEmergencias)
	r.db.WithContext(ctx).Model(&models.SolicitudEmergenciaORM{}).
		Where("estado_solicitud IN ?", []string{"Registrada", "Buscando Stock", "Reservada", "En Espera de Donantes"}).
		Count(&emergenciasActivas)
	r.db.WithContext(ctx).Model(&models.SolicitudEmergenciaORM{}).
		Where("estado_solicitud = ?", "Atendida").
		Count(&emergenciasAtendidas)

	var avgLatency float64
	r.db.WithContext(ctx).Model(&models.SolicitudEmergenciaORM{}).
		Select("COALESCE(AVG(tiempo_respuesta_matching_ms), 0)").
		Scan(&avgLatency)

	var donantesHabilitados int64
	r.db.WithContext(ctx).Model(&models.DonanteORM{}).
		Where("activo = ? AND habilitado_convocatorias = ?", true, true).
		Count(&donantesHabilitados)

	// Stock por tipo de sangre y Rh
	grupos := []struct {
		Tipo string
		Rh   string
	}{
		{"O", "+"}, {"O", "-"},
		{"A", "+"}, {"A", "-"},
		{"B", "+"}, {"B", "-"},
		{"AB", "+"}, {"AB", "-"},
	}

	stockPorTipo := make([]domain.StockPorGrupoKPI, 0, len(grupos))
	for _, g := range grupos {
		var disp, resv int64
		r.db.WithContext(ctx).Model(&models.UnidadSangreORM{}).
			Where("tipo_sangre = ? AND factor_rh = ? AND estado = ? AND fecha_caducidad > ?", g.Tipo, g.Rh, "Disponible", now).
			Count(&disp)
		r.db.WithContext(ctx).Model(&models.UnidadSangreORM{}).
			Where("tipo_sangre = ? AND factor_rh = ? AND estado = ?", g.Tipo, g.Rh, "Reservada").
			Count(&resv)

		stockPorTipo = append(stockPorTipo, domain.StockPorGrupoKPI{
			GrupoSanguineo:      g.Tipo + g.Rh,
			TipoSangre:          g.Tipo,
			FactorRh:            g.Rh,
			UnidadesDisponibles: int(disp),
			UnidadesReservadas:  int(resv),
		})
	}

	estados := []string{"Reservada", "Atendida", "En Espera de Donantes", "Cancelada"}
	emergenciasPorEstado := make([]domain.EmergenciaEstadoKPI, 0, len(estados))
	for _, est := range estados {
		var cnt int64
		r.db.WithContext(ctx).Model(&models.SolicitudEmergenciaORM{}).
			Where("estado_solicitud = ?", est).
			Count(&cnt)
		emergenciasPorEstado = append(emergenciasPorEstado, domain.EmergenciaEstadoKPI{
			Estado:   est,
			Cantidad: int(cnt),
		})
	}

	var notifORMs []models.NotificacionLogORM
	r.db.WithContext(ctx).Order("created_at DESC").Limit(15).Find(&notifORMs)
	notifs := make([]domain.NotificacionResumenItem, 0, len(notifORMs))
	for _, n := range notifORMs {
		notifs = append(notifs, domain.NotificacionResumenItem{
			ID:           n.ID,
			Evento:       n.TipoEvento,
			Canal:        n.Canal,
			Destinatario: n.Destinatario,
			Mensaje:      n.Mensaje,
			Timestamp:    n.CreatedAt,
		})
	}

	return &domain.DashboardMetricas{
		TotalBancosRed:             int(totalBancos),
		TotalUnidadesDisponibles:   int(totalDisponibles),
		UnidadesProximasACaducar7d: int(porCaducar7d),
		UnidadesAlertaRoja3d:       int(alertaRoja3d),
		TotalEmergencias:           int(totalEmergencias),
		EmergenciasActivas:         int(emergenciasActivas),
		EmergenciasAtendidas:       int(emergenciasAtendidas),
		TiempoPromedioMatchingMs:   math.Round(avgLatency*100) / 100,
		DonantesHabilitados:        int(donantesHabilitados),
		StockPorTipoSangre:         stockPorTipo,
		EmergenciasPorEstado:       emergenciasPorEstado,
		UltimasNotificaciones:      notifs,
		GeneradoEn:                 time.Now().Format(time.RFC3339),
	}, nil
}

func (r *ReporteRepository) ObtenerTrazabilidadUnidad(ctx context.Context, unidadID string) (*domain.TrazabilidadUnidadReporte, error) {
	var u models.UnidadSangreORM
	if err := r.db.WithContext(ctx).Where("id = ?", unidadID).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrTrazabilidadUnidadNoEncontrada
		}
		return nil, err
	}

	bancoNombre := u.BancoID
	var b models.BancoSangreORM
	if err := r.db.WithContext(ctx).Where("id = ?", u.BancoID).First(&b).Error; err == nil {
		bancoNombre = b.Nombre
	}

	var logs []models.AuditLogORM
	if err := r.db.WithContext(ctx).
		Where("entidad_id = ?", unidadID).
		Order("timestamp ASC, id ASC").
		Find(&logs).Error; err != nil {
		return nil, err
	}

	historial := make([]domain.EventoAuditItem, 0, len(logs))
	for _, l := range logs {
		historial = append(historial, domain.EventoAuditItem{
			ID:             l.ID,
			Entidad:        l.Entidad,
			EntidadID:      l.EntidadID,
			Accion:         l.Accion,
			EstadoAnterior: l.EstadoAnterior,
			EstadoNuevo:    l.EstadoNuevo,
			UsuarioID:      l.UsuarioID,
			Detalle:        l.Detalle,
			Timestamp:      l.Timestamp,
		})
	}

	return &domain.TrazabilidadUnidadReporte{
		UnidadID:        u.ID,
		CodigoBarras:    u.CodigoBarras,
		BancoID:         u.BancoID,
		BancoNombre:     bancoNombre,
		TipoSangre:      u.TipoSangre,
		FactorRh:        u.FactorRh,
		VolumenMl:       u.VolumenMl,
		Origen:          u.Origen,
		FechaExtraccion: u.FechaExtraccion.Format("2006-01-02"),
		FechaCaducidad:  u.FechaCaducidad.Format("2006-01-02"),
		EstadoActual:    u.Estado,
		Historial:       historial,
		GeneradoEn:      time.Now().Format(time.RFC3339),
	}, nil
}
