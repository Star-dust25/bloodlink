package http

import (
	"errors"
	"net/http"

	"github.com/bloodlink/backend/internal/middleware"
	"github.com/bloodlink/backend/internal/modules/auth/domain"
	"github.com/bloodlink/backend/internal/modules/auth/ports"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service ports.IAuthService
}

func NewAuthHandler(service ports.IAuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.GET("/me", authMiddleware, h.Me)
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req ports.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Payload de login inválido", err.Error())
		return
	}

	resp, err := h.service.Autenticar(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrCuentaBloqueada) {
			middleware.RespondError(c, http.StatusLocked, "ACCOUNT_LOCKED", err.Error(), "Intente nuevamente después de 15 minutos")
			return
		}
		if errors.Is(err, domain.ErrCredencialesInvalidas) {
			middleware.RespondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error(), "Verifique su correo o contraseña")
			return
		}
		middleware.RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error al procesar autenticación", err.Error())
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("userId")
	if !exists {
		middleware.RespondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "Sesión no válida", "")
		return
	}
	userID, _ := userIDVal.(uint)
	perfil, err := h.service.ObtenerPerfil(c.Request.Context(), userID)
	if err != nil {
		middleware.RespondError(c, http.StatusNotFound, "USER_NOT_FOUND", err.Error(), "")
		return
	}
	c.JSON(http.StatusOK, perfil)
}
