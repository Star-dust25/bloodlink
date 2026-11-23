package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bloodlink/backend/internal/middleware"
	"github.com/bloodlink/backend/internal/modules/emergencias/domain"
	"github.com/bloodlink/backend/internal/modules/emergencias/ports"
	"github.com/gin-gonic/gin"
)

type EmergenciaHandler struct {
	service ports.IMatching
}

func NewEmergenciaHandler(service ports.IMatching) *EmergenciaHandler {
	return &EmergenciaHandler{service: service}
}

func (h *EmergenciaHandler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc) {
	em := rg.Group("/emergencias", authMw)
	{
		em.POST("", middleware.RequireRoles("MEDICO"), h.CrearSolicitudYMatching)
		em.GET("", middleware.RequireRoles("MEDICO", "PERSONAL_BANCO", "COORDINADOR"), h.ListarSolicitudes)
		em.GET("/:solicitudId", middleware.RequireRoles("MEDICO", "PERSONAL_BANCO", "COORDINADOR"), h.ObtenerSolicitud)
		em.PATCH("/:solicitudId/cancelar", middleware.RequireRoles("MEDICO"), h.CancelarSolicitud)
		em.PUT("/:solicitudId/confirmar", middleware.RequireRoles("MEDICO", "PERSONAL_BANCO"), h.ConfirmarTransfusion)
	}
}

func (h *EmergenciaHandler) CrearSolicitudYMatching(c *gin.Context) {
	var req ports.CrearEmergenciaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Datos de solicitud de emergencia inválidos", err.Error())
		return
	}

	userID := c.GetUint("userId")
	if req.MedicoID == 0 && userID > 0 {
		req.MedicoID = userID
	}
	usuarioIDStr := fmt.Sprintf("%d", userID)

	resp, conflict, err := h.service.CrearSolicitudYMatching(c.Request.Context(), req, usuarioIDStr)
	if err != nil {
		if errors.Is(err, domain.ErrStockNoDisponibleEnRed) && conflict != nil {
			c.JSON(http.StatusConflict, conflict)
			return
		}
		if errors.Is(err, domain.ErrParametrosEmergenciaInvalidos) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_EMERGENCY_PARAMS", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "MATCHING_ERROR", "Error al procesar solicitud de emergencia", err.Error())
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *EmergenciaHandler) ListarSolicitudes(c *gin.Context) {
	estado := c.Query("estado")
	list, err := h.service.ListarSolicitudes(c.Request.Context(), estado)
	if err != nil {
		middleware.RespondError(c, http.StatusInternalServerError, "EMERGENCIA_LIST_ERROR", "Error al listar emergencias", err.Error())
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *EmergenciaHandler) ObtenerSolicitud(c *gin.Context) {
	solicitudID := c.Param("solicitudId")
	sol, err := h.service.ObtenerSolicitud(c.Request.Context(), solicitudID)
	if err != nil {
		if errors.Is(err, domain.ErrSolicitudNoEncontrada) {
			middleware.RespondError(c, http.StatusNotFound, "SOLICITUD_NOT_FOUND", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "EMERGENCIA_GET_ERROR", "Error al consultar solicitud", err.Error())
		return
	}
	c.JSON(http.StatusOK, sol)
}

func (h *EmergenciaHandler) CancelarSolicitud(c *gin.Context) {
	solicitudID := c.Param("solicitudId")
	var req ports.CancelarEmergenciaRequest
	_ = c.ShouldBindJSON(&req)

	usuarioID := fmt.Sprintf("%d", c.GetUint("userId"))
	sol, err := h.service.CancelarSolicitud(c.Request.Context(), solicitudID, req.Motivo, usuarioID)
	if err != nil {
		if errors.Is(err, domain.ErrSolicitudNoEncontrada) {
			middleware.RespondError(c, http.StatusNotFound, "SOLICITUD_NOT_FOUND", err.Error(), "")
			return
		}
		if errors.Is(err, domain.ErrTransicionSolicitudInvalida) {
			middleware.RespondError(c, http.StatusUnprocessableEntity, "INVALID_STATE_TRANSITION", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "CANCEL_ERROR", "Error al cancelar solicitud", err.Error())
		return
	}
	c.JSON(http.StatusOK, sol)
}

func (h *EmergenciaHandler) ConfirmarTransfusion(c *gin.Context) {
	solicitudID := c.Param("solicitudId")
	usuarioID := fmt.Sprintf("%d", c.GetUint("userId"))
	sol, err := h.service.ConfirmarTransfusion(c.Request.Context(), solicitudID, usuarioID)
	if err != nil {
		if errors.Is(err, domain.ErrSolicitudNoEncontrada) {
			middleware.RespondError(c, http.StatusNotFound, "SOLICITUD_NOT_FOUND", err.Error(), "")
			return
		}
		if errors.Is(err, domain.ErrTransicionSolicitudInvalida) {
			middleware.RespondError(c, http.StatusUnprocessableEntity, "INVALID_STATE_TRANSITION", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "CONFIRM_ERROR", "Error al confirmar transfusión", err.Error())
		return
	}
	c.JSON(http.StatusOK, sol)
}
