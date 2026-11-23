package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bloodlink/backend/internal/middleware"
	"github.com/bloodlink/backend/internal/modules/reportes/domain"
	"github.com/bloodlink/backend/internal/modules/reportes/ports"
	"github.com/gin-gonic/gin"
)

type ReporteHandler struct {
	service ports.IReporteService
}

func NewReporteHandler(service ports.IReporteService) *ReporteHandler {
	return &ReporteHandler{service: service}
}

func (h *ReporteHandler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc) {
	rep := rg.Group("/reportes", authMw)
	{
		rep.GET("/dashboard", middleware.RequireRoles("COORDINADOR", "MEDICO", "PERSONAL_BANCO"), h.ObtenerDashboard)
		rep.GET("/trazabilidad/:unidadId", middleware.RequireRoles("COORDINADOR", "PERSONAL_BANCO"), h.DescargarTrazabilidad)
	}
}

func (h *ReporteHandler) ObtenerDashboard(c *gin.Context) {
	dash, err := h.service.ObtenerDashboard(c.Request.Context())
	if err != nil {
		middleware.RespondError(c, http.StatusInternalServerError, "DASHBOARD_ERROR", "Error generando métricas del dashboard", err.Error())
		return
	}
	c.JSON(http.StatusOK, dash)
}

func (h *ReporteHandler) DescargarTrazabilidad(c *gin.Context) {
	unidadID := c.Param("unidadId")
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "json")))

	switch format {
	case "pdf":
		data, filename, err := h.service.GenerarTrazabilidadPDF(c.Request.Context(), unidadID)
		if err != nil {
			if errors.Is(err, domain.ErrTrazabilidadUnidadNoEncontrada) {
				middleware.RespondError(c, http.StatusNotFound, "UNIDAD_NOT_FOUND", err.Error(), "")
				return
			}
			middleware.RespondError(c, http.StatusInternalServerError, "PDF_EXPORT_ERROR", "Error generando PDF de trazabilidad", err.Error())
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Data(http.StatusOK, "application/pdf", data)

	case "csv":
		data, filename, err := h.service.GenerarTrazabilidadCSV(c.Request.Context(), unidadID)
		if err != nil {
			if errors.Is(err, domain.ErrTrazabilidadUnidadNoEncontrada) {
				middleware.RespondError(c, http.StatusNotFound, "UNIDAD_NOT_FOUND", err.Error(), "")
				return
			}
			middleware.RespondError(c, http.StatusInternalServerError, "CSV_EXPORT_ERROR", "Error generando CSV de trazabilidad", err.Error())
			return
		}
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.Data(http.StatusOK, "text/csv; charset=utf-8", data)

	default:
		rep, err := h.service.ObtenerTrazabilidadJSON(c.Request.Context(), unidadID)
		if err != nil {
			if errors.Is(err, domain.ErrTrazabilidadUnidadNoEncontrada) {
				middleware.RespondError(c, http.StatusNotFound, "UNIDAD_NOT_FOUND", err.Error(), "")
				return
			}
			middleware.RespondError(c, http.StatusInternalServerError, "TRACE_ERROR", "Error consultando trazabilidad", err.Error())
			return
		}
		c.JSON(http.StatusOK, rep)
	}
}
