package http

import (
	"errors"
	"net/http"

	"github.com/bloodlink/backend/internal/middleware"
	"github.com/bloodlink/backend/internal/modules/donantes/domain"
	"github.com/bloodlink/backend/internal/modules/donantes/ports"
	integDomain "github.com/bloodlink/backend/internal/modules/integracion/domain"
	"github.com/gin-gonic/gin"
)

type DonanteHandler struct {
	service ports.IDonanteService
}

func NewDonanteHandler(service ports.IDonanteService) *DonanteHandler {
	return &DonanteHandler{service: service}
}

func (h *DonanteHandler) RegisterRoutes(rg *gin.RouterGroup, authMw gin.HandlerFunc) {
	don := rg.Group("/donantes")
	{
		// Endpoints públicos (Portal de Donantes CUN-002 / CP-03)
		don.POST("/validar-dni", h.ValidarDNIReniec)
		don.POST("", h.RegistrarDonante)

		// Endpoints protegidos por RBAC
		don.GET("", authMw, middleware.RequireRoles("COORDINADOR", "PERSONAL_BANCO"), h.ListarDonantes)
		don.GET("/:donanteId", authMw, middleware.RequireRoles("COORDINADOR", "PERSONAL_BANCO"), h.ObtenerPerfil)
		don.PUT("/:donanteId", authMw, middleware.RequireRoles("COORDINADOR", "PERSONAL_BANCO"), h.ActualizarContacto)
		don.DELETE("/:donanteId", authMw, middleware.RequireRoles("COORDINADOR", "PERSONAL_BANCO"), h.DarDeBaja)
	}
}

func (h *DonanteHandler) ValidarDNIReniec(c *gin.Context) {
	var req ports.ValidarDNIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Debe proporcionar un DNI válido", err.Error())
		return
	}

	ciudadano, err := h.service.ValidarDNIReniec(c.Request.Context(), req.DNI)
	if err != nil {
		if errors.Is(err, integDomain.ErrDNIFormatoInvalido) {
			middleware.RespondError(c, http.StatusBadRequest, "INVALID_DNI_FORMAT", err.Error(), "")
			return
		}
		if errors.Is(err, integDomain.ErrDNINoEncontrado) {
			middleware.RespondError(c, http.StatusNotFound, "RENIEC_DNI_NOT_FOUND", err.Error(), "Verifique el número de documento en RENIEC")
			return
		}
		if errors.Is(err, integDomain.ErrDNIMenorDeEdad) {
			middleware.RespondError(c, http.StatusUnprocessableEntity, "RENIEC_MINOR_AGE", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusBadGateway, "SOAP_RENIEC_ERROR", "Error consultando servicio SOAP RENIEC", err.Error())
		return
	}

	c.JSON(http.StatusOK, ciudadano)
}

func (h *DonanteHandler) RegistrarDonante(c *gin.Context) {
	var req ports.RegistrarDonanteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Datos de registro de donante incompletos", err.Error())
		return
	}

	resp, err := h.service.RegistrarDonante(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrPoliticaPrivacidadRequerida) || errors.Is(err, domain.ErrDatosDonanteInvalidos) || errors.Is(err, integDomain.ErrDNIFormatoInvalido) {
			middleware.RespondError(c, http.StatusBadRequest, "DONOR_VALIDATION_ERROR", err.Error(), "")
			return
		}
		if errors.Is(err, integDomain.ErrDNINoEncontrado) {
			middleware.RespondError(c, http.StatusNotFound, "RENIEC_DNI_NOT_FOUND", err.Error(), "")
			return
		}
		if errors.Is(err, integDomain.ErrDNIMenorDeEdad) {
			middleware.RespondError(c, http.StatusUnprocessableEntity, "RENIEC_MINOR_AGE", err.Error(), "")
			return
		}
		if errors.Is(err, domain.ErrDonanteYaRegistrado) {
			middleware.RespondError(c, http.StatusConflict, "DONOR_ALREADY_EXISTS", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "DONOR_REGISTER_ERROR", "Error registrando donante", err.Error())
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *DonanteHandler) ListarDonantes(c *gin.Context) {
	list, err := h.service.ListarDonantes(c.Request.Context(), c.Query("tipoSangre"), c.Query("factorRh"))
	if err != nil {
		middleware.RespondError(c, http.StatusInternalServerError, "DONOR_LIST_ERROR", "Error listando donantes", err.Error())
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *DonanteHandler) ObtenerPerfil(c *gin.Context) {
	id := c.Param("donanteId")
	d, err := h.service.ObtenerPerfilDonante(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrDonanteNoEncontrado) {
			middleware.RespondError(c, http.StatusNotFound, "DONOR_NOT_FOUND", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "DONOR_GET_ERROR", "Error consultando donante", err.Error())
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *DonanteHandler) ActualizarContacto(c *gin.Context) {
	id := c.Param("donanteId")
	var req ports.ActualizarContactoDonanteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Payload inválido", err.Error())
		return
	}

	d, err := h.service.ActualizarContacto(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, domain.ErrDonanteNoEncontrado) {
			middleware.RespondError(c, http.StatusNotFound, "DONOR_NOT_FOUND", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "DONOR_UPDATE_ERROR", "Error actualizando donante", err.Error())
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *DonanteHandler) DarDeBaja(c *gin.Context) {
	id := c.Param("donanteId")
	if err := h.service.DarDeBaja(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrDonanteNoEncontrado) {
			middleware.RespondError(c, http.StatusNotFound, "DONOR_NOT_FOUND", err.Error(), "")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "DONOR_DELETE_ERROR", "Error al dar de baja", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"mensaje": "Donante dado de baja exitosamente", "donanteId": id})
}
