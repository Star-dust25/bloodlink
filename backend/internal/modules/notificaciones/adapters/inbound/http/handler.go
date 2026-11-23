package http

import (
	"net/http"
	"strconv"

	"github.com/bloodlink/backend/internal/middleware"
	"github.com/bloodlink/backend/internal/modules/notificaciones/ports"
	"github.com/gin-gonic/gin"
)

type NotificacionHandler struct {
	service ports.INotificacionService
}

func NewNotificacionHandler(service ports.INotificacionService) *NotificacionHandler {
	return &NotificacionHandler{service: service}
}

func (h *NotificacionHandler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc) {
	not := rg.Group("/notificaciones", authMw)
	{
		not.GET("", middleware.RequireRoles("COORDINADOR", "PERSONAL_BANCO", "MEDICO"), h.ListarNotificaciones)
		not.POST("/verificar-caducidad", middleware.RequireRoles("PERSONAL_BANCO", "COORDINADOR"), h.VerificarCaducidad)
	}
}

func (h *NotificacionHandler) ListarNotificaciones(c *gin.Context) {
	canal := c.Query("canal")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	list, err := h.service.ListarNotificaciones(c.Request.Context(), canal, limit)
	if err != nil {
		middleware.RespondError(c, http.StatusInternalServerError, "NOTIFICACION_LIST_ERROR", "Error listando notificaciones", err.Error())
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *NotificacionHandler) VerificarCaducidad(c *gin.Context) {
	resp, err := h.service.EjecutarJobVerificacionCaducidad(c.Request.Context())
	if err != nil {
		middleware.RespondError(c, http.StatusInternalServerError, "EXPIRATION_JOB_ERROR", "Error ejecutando verificación de caducidad", err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}
