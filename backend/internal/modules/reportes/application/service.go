package application

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/bloodlink/backend/internal/modules/reportes/domain"
	"github.com/bloodlink/backend/internal/modules/reportes/ports"
	"github.com/go-pdf/fpdf"
)

type SrvReportes struct {
	repo ports.IReporteRepository
}

func NewSrvReportes(repo ports.IReporteRepository) *SrvReportes {
	return &SrvReportes{repo: repo}
}

func (s *SrvReportes) ObtenerDashboard(ctx context.Context) (*domain.DashboardMetricas, error) {
	return s.repo.ConstruirMetricasDashboard(ctx)
}

func (s *SrvReportes) ObtenerTrazabilidadJSON(ctx context.Context, unidadID string) (*domain.TrazabilidadUnidadReporte, error) {
	return s.repo.ObtenerTrazabilidadUnidad(ctx, unidadID)
}

func (s *SrvReportes) GenerarTrazabilidadPDF(ctx context.Context, unidadID string) ([]byte, string, error) {
	rep, err := s.repo.ObtenerTrazabilidadUnidad(ctx, unidadID)
	if err != nil {
		return nil, "", err
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Encabezado Institucional MINSA - PRONAHEBAS
	pdf.SetFont("Helvetica", "B", 15)
	pdf.CellFormat(190, 9, "MINSA - RED NACIONAL DE BANCOS DE SANGRE BLOODLINK", "0", 1, "C", false, 0, "")
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(190, 7, fmt.Sprintf("Certificado de Trazabilidad Clinica de Unidad: %s", rep.UnidadID), "0", 1, "C", false, 0, "")
	pdf.Ln(4)

	// Ficha Técnica de la Unidad
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetFillColor(235, 240, 248)
	pdf.CellFormat(190, 7, "1. DATOS DE LA UNIDAD HEMODERIVADA", "1", 1, "L", true, 0, "")

	pdf.SetFont("Helvetica", "", 9)
	pdf.CellFormat(95, 6, fmt.Sprintf("Unidad ID: %s", rep.UnidadID), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Codigo de Barras: %s", rep.CodigoBarras), "1", 1, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Banco Custodio: %s (%s)", rep.BancoID, rep.BancoNombre), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Grupo y Factor Rh: %s%s (%d mL)", rep.TipoSangre, rep.FactorRh, rep.VolumenMl), "1", 1, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Fecha Extraccion: %s", rep.FechaExtraccion), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Fecha Caducidad (42d): %s", rep.FechaCaducidad), "1", 1, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Estado Actual: %s", rep.EstadoActual), "1", 0, "L", false, 0, "")
	pdf.CellFormat(95, 6, fmt.Sprintf("Origen: %s", rep.Origen), "1", 1, "L", false, 0, "")
	pdf.Ln(6)

	// Tabla de Eventos de Auditoría (AuditLog)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.CellFormat(190, 7, "2. HISTORIAL INMUTABLE DE AUDITORIA (CADENA DE CUSTODIA RF-13)", "1", 1, "L", true, 0, "")

	pdf.SetFont("Helvetica", "B", 8)
	pdf.CellFormat(34, 6, "Fecha / Hora", "1", 0, "C", true, 0, "")
	pdf.CellFormat(44, 6, "Accion", "1", 0, "C", true, 0, "")
	pdf.CellFormat(36, 6, "Transicion", "1", 0, "C", true, 0, "")
	pdf.CellFormat(24, 6, "Usuario", "1", 0, "C", true, 0, "")
	pdf.CellFormat(52, 6, "Detalle", "1", 1, "C", true, 0, "")

	pdf.SetFont("Helvetica", "", 7)
	for _, ev := range rep.Historial {
		ts := ev.Timestamp.Format("2006-01-02 15:04:05")
		trans := fmt.Sprintf("%s -> %s", ev.EstadoAnterior, ev.EstadoNuevo)
		detalle := ev.Detalle
		if len(detalle) > 38 {
			detalle = detalle[:35] + "..."
		}
		pdf.CellFormat(34, 6, ts, "1", 0, "L", false, 0, "")
		pdf.CellFormat(44, 6, ev.Accion, "1", 0, "L", false, 0, "")
		pdf.CellFormat(36, 6, trans, "1", 0, "C", false, 0, "")
		pdf.CellFormat(24, 6, ev.UsuarioID, "1", 0, "C", false, 0, "")
		pdf.CellFormat(52, 6, detalle, "1", 1, "L", false, 0, "")
	}

	pdf.Ln(8)
	pdf.SetFont("Helvetica", "I", 8)
	pdf.CellFormat(190, 5, fmt.Sprintf("Emitido electronicamente por BloodLink Platform el %s", time.Now().Format(time.RFC3339)), "0", 1, "R", false, 0, "")

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, "", err
	}

	filename := fmt.Sprintf("trazabilidad_%s.pdf", rep.UnidadID)
	return buf.Bytes(), filename, nil
}

func (s *SrvReportes) GenerarTrazabilidadCSV(ctx context.Context, unidadID string) ([]byte, string, error) {
	rep, err := s.repo.ObtenerTrazabilidadUnidad(ctx, unidadID)
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	// UTF-8 BOM para compatibilidad con Microsoft Excel
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{
		"UnidadID", "CodigoBarras", "BancoID", "GrupoSanguineo",
		"EventoID", "Timestamp", "Accion", "EstadoAnterior", "EstadoNuevo", "UsuarioID", "Detalle",
	})

	grupo := rep.TipoSangre + rep.FactorRh
	for _, ev := range rep.Historial {
		_ = writer.Write([]string{
			rep.UnidadID,
			rep.CodigoBarras,
			rep.BancoID,
			grupo,
			fmt.Sprintf("%d", ev.ID),
			ev.Timestamp.Format(time.RFC3339),
			ev.Accion,
			ev.EstadoAnterior,
			ev.EstadoNuevo,
			ev.UsuarioID,
			ev.Detalle,
		})
	}
	writer.Flush()

	filename := fmt.Sprintf("trazabilidad_%s.csv", rep.UnidadID)
	return buf.Bytes(), filename, writer.Error()
}
