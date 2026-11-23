package ports

import (
	"context"

	"github.com/bloodlink/backend/internal/modules/reportes/domain"
)

// IReporteService defines the inbound port for MS-Reportes (RF-08, RF-10, RF-13).
type IReporteService interface {
	ObtenerDashboard(ctx context.Context) (*domain.DashboardMetricas, error)
	ObtenerTrazabilidadJSON(ctx context.Context, unidadID string) (*domain.TrazabilidadUnidadReporte, error)
	GenerarTrazabilidadPDF(ctx context.Context, unidadID string) ([]byte, string, error)
	GenerarTrazabilidadCSV(ctx context.Context, unidadID string) ([]byte, string, error)
}

// IReporteRepository defines the outbound read-model port for consolidated metrics and unit traceability.
type IReporteRepository interface {
	ConstruirMetricasDashboard(ctx context.Context) (*domain.DashboardMetricas, error)
	ObtenerTrazabilidadUnidad(ctx context.Context, unidadID string) (*domain.TrazabilidadUnidadReporte, error)
}
