package application

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bloodlink/backend/internal/modules/reportes/domain"
)

type mockReporteRepo struct{}

func (m *mockReporteRepo) ConstruirMetricasDashboard(_ context.Context) (*domain.DashboardMetricas, error) {
	return &domain.DashboardMetricas{
		TotalBancosRed:           4,
		TotalUnidadesDisponibles: 30,
		TiempoPromedioMatchingMs: 42.5,
	}, nil
}

func (m *mockReporteRepo) ObtenerTrazabilidadUnidad(_ context.Context, unidadID string) (*domain.TrazabilidadUnidadReporte, error) {
	if unidadID == "NO-EXISTE" {
		return nil, domain.ErrTrazabilidadUnidadNoEncontrada
	}
	return &domain.TrazabilidadUnidadReporte{
		UnidadID:        unidadID,
		CodigoBarras:    "BL-MINSA01-20260915",
		BancoID:         "HOSP-MINSA-01",
		BancoNombre:     "Hospital Nacional Santa Rosa - Piura",
		TipoSangre:      "O",
		FactorRh:        "+",
		VolumenMl:       450,
		Origen:          "DONACION_VOLUNTARIA",
		FechaExtraccion: "2026-05-01",
		FechaCaducidad:  "2026-06-12",
		EstadoActual:    "Reservada",
		Historial: []domain.EventoAuditItem{
			{
				ID:             1,
				Entidad:        "UnidadSangre",
				EntidadID:      unidadID,
				Accion:         "REGISTRAR_INGRESO_UNIDAD",
				EstadoAnterior: "NINGUNO",
				EstadoNuevo:    "Disponible",
				UsuarioID:      "13",
				Detalle:        "Ingreso inicial",
				Timestamp:      time.Now().Add(-2 * time.Hour),
			},
			{
				ID:             2,
				Entidad:        "UnidadSangre",
				EntidadID:      unidadID,
				Accion:         "RESERVA_AUTOMATICA_MATCHING",
				EstadoAnterior: "Disponible",
				EstadoNuevo:    "Reservada",
				UsuarioID:      "12",
				Detalle:        "Reserva para emergencia SOL-2026-0801",
				Timestamp:      time.Now().Add(-1 * time.Hour),
			},
		},
	}, nil
}

func TestSrvReportes_Dashboard_PDF_CSV_JSON(t *testing.T) {
	srv := NewSrvReportes(&mockReporteRepo{})
	ctx := context.Background()

	dash, err := srv.ObtenerDashboard(ctx)
	if err != nil || dash.TotalBancosRed != 4 {
		t.Fatalf("ObtenerDashboard error: %v", err)
	}

	jsonRep, err := srv.ObtenerTrazabilidadJSON(ctx, "UND-2026-0915")
	if err != nil || len(jsonRep.Historial) != 2 {
		t.Fatalf("ObtenerTrazabilidadJSON error: %v", err)
	}

	pdfBytes, pdfName, err := srv.GenerarTrazabilidadPDF(ctx, "UND-2026-0915")
	if err != nil || !bytes.HasPrefix(pdfBytes, []byte("%PDF")) || pdfName != "trazabilidad_UND-2026-0915.pdf" {
		t.Fatalf("GenerarTrazabilidadPDF inválido: err=%v, len=%d, name=%s", err, len(pdfBytes), pdfName)
	}

	csvBytes, csvName, err := srv.GenerarTrazabilidadCSV(ctx, "UND-2026-0915")
	if err != nil || !strings.Contains(string(csvBytes), "REGISTRAR_INGRESO_UNIDAD") || csvName != "trazabilidad_UND-2026-0915.csv" {
		t.Fatalf("GenerarTrazabilidadCSV inválido: err=%v, name=%s", err, csvName)
	}
}
