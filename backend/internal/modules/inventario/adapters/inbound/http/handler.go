package http

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bloodlink/backend/internal/middleware"
	"github.com/bloodlink/backend/internal/modules/inventario/domain"
	"github.com/bloodlink/backend/internal/modules/inventario/ports"
	"github.com/gin-gonic/gin"
)

type InventarioHandler struct {
	service ports.IGestorInventario
}

func NewInventarioHandler(service ports.IGestorInventario) *InventarioHandler {
	return &InventarioHandler{service: service}
}

func (h *InventarioHandler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc) {
	inv := rg.Group("/inventario", authMw)
	{
		inv.GET("/stock", middleware.RequireRoles("MEDICO", "PERSONAL_BANCO", "COORDINADOR"), h.ConsultarStockRed)
		inv.GET("/unidades", middleware.RequireRoles("MEDICO", "PERSONAL_BANCO", "COORDINADOR"), h.ListarUnidades)
		inv.GET("/unidades/:unidadId", middleware.RequireRoles("MEDICO", "PERSONAL_BANCO", "COORDINADOR"), h.ObtenerUnidad)
		inv.POST("/unidades", middleware.RequireRoles("PERSONAL_BANCO"), h.RegistrarUnidad)
		inv.PATCH("/unidades/:unidadId/estado", middleware.RequireRoles("PERSONAL_BANCO"), h.ActualizarEstado)
	}
}

func (h *InventarioHandler) ConsultarStockRed(c *gin.Context) {
	filtro := ports.FiltroStock{
		TipoSangre: c.Query("tipoSangre"),
		FactorRh:   c.Query("factorRh"),
		Estado:     c.Query("estado"),
		HospitalID: c.Query("hospitalId"),
	}

	resp, err := h.service.ConsultarStockRed(c.Request.Context(), filtro)
	if err != nil {
		middleware.RespondError(c, http.StatusInternalServerError, "INVENTARIO_ERROR", "Error al consultar stock de red", err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *InventarioHandler) ListarUnidades(c *gin.Context) {
	hospitalID := c.Query("hospitalId")
	if hospitalID == "" {
		hospitalID = c.Query("bancoId")
	}
	filtro := ports.FiltroStock{
		TipoSangre: c.Query("tipoSangre"),
		FactorRh:   c.Query("factorRh"),
		Estado:     c.Query("estado"),
		HospitalID: hospitalID,
	}

	unidades, err := h.service.ListarUnidades(c.Request.Context(), filtro)
	if err != nil {
		middleware.RespondError(c, http.StatusInternalServerError, "INVENTARIO_ERROR", "Error al listar unidades", err.Error())
		return
	}
	c.JSON(http.StatusOK, unidades)
}

func (h *InventarioHandler) ObtenerUnidad(c *gin.Context) {
	unidadID := c.Param("unidadId")
	u, err := h.service.ObtenerUnidad(c.Request.Context(), unidadID)
	if err != nil {
		if errors.Is(err, domain.ErrUnidadNoEncontrada) {
			middleware.RespondError(c, http.StatusNotFound, "UNIDAD_NOT_FOUND", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "INVENTARIO_ERROR", "Error al obtener unidad", err.Error())
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *InventarioHandler) RegistrarUnidad(c *gin.Context) {
	var req ports.RegistrarUnidadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Datos de ingreso de unidad inválidos", err.Error())
		return
	}

	usuarioID := fmt.Sprintf("%v", c.GetUint("userId"))
	resp, err := h.service.RegistrarUnidad(c.Request.Context(), req, usuarioID)
	if err != nil {
		if errors.Is(err, domain.ErrUnidadYaCaducada) {
			middleware.RespondError(c, http.StatusUnprocessableEntity, "UNIDAD_CADUCADA_CUN003", err.Error(), "No se permite registrar bolsas con más de 42 días desde su extracción")
			return
		}
		if errors.Is(err, domain.ErrTipoSangreInvalido) || errors.Is(err, domain.ErrBancoNoEncontrado) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_BLOOD_UNIT_DATA", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "INVENTARIO_ERROR", "Error registrando unidad", err.Error())
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *InventarioHandler) ActualizarEstado(c *gin.Context) {
	unidadID := c.Param("unidadId")
	var req ports.ActualizarEstadoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Payload inválido", err.Error())
		return
	}

	usuarioID := fmt.Sprintf("%v", c.GetUint("userId"))
	resp, err := h.service.ActualizarEstado(c.Request.Context(), unidadID, req, usuarioID)
	if err != nil {
		if errors.Is(err, domain.ErrUnidadNoEncontrada) {
			middleware.RespondError(c, http.StatusNotFound, "UNIDAD_NOT_FOUND", err.Error(), "")
			return
		}
		if errors.Is(err, domain.ErrTransicionEstadoInvalida) {
			middleware.RespondError(c, http.StatusUnprocessableEntity, "INVALID_STATE_TRANSITION", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "INVENTARIO_ERROR", "Error actualizando estado", err.Error())
		return
	}

	c.JSON(http.StatusOK, resp)
}
