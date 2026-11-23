package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID         uint   `json:"userId"`
	Rol            string `json:"rol"`
	HospitalID     string `json:"hospitalId"`
	NombreCompleto string `json:"nombreCompleto"`
	Email          string `json:"email"`
	jwt.RegisteredClaims
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Codigo    string `json:"codigo"`
	Detalle   string `json:"detalle,omitempty"`
	Timestamp string `json:"timestamp"`
}

func RespondError(c *gin.Context, status int, codigo, msg, detalle string) {
	c.AbortWithStatusJSON(status, ErrorResponse{
		Error:     msg,
		Codigo:    codigo,
		Detalle:   detalle,
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowAll := len(allowedOrigins) == 0 || (len(allowedOrigins) == 1 && allowedOrigins[0] == "*")
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[strings.TrimSpace(o)] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowAll || originSet[origin] || origin == "" {
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
			} else {
				c.Header("Access-Control-Allow-Origin", "*")
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, SOAPAction")
		c.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		slog.Info("http_request",
			"method", method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			RespondError(c, http.StatusUnauthorized, "AUTH_TOKEN_MISSING", "Token de autenticación requerido", "Incluya header Authorization: Bearer <token>")
			return
		}

		tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims := &JWTClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algoritmo de firma inesperado: %v", t.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})
		if err != nil || !token.Valid {
			RespondError(c, http.StatusUnauthorized, "AUTH_TOKEN_INVALID", "Token JWT inválido o expirado", "")
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("rol", claims.Rol)
		c.Set("hospitalId", claims.HospitalID)
		c.Set("nombreCompleto", claims.NombreCompleto)
		c.Set("email", claims.Email)
		c.Set("claims", claims)
		c.Next()
	}
}

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	roleMap := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		roleMap[r] = true
	}

	return func(c *gin.Context) {
		rolVal, exists := c.Get("rol")
		if !exists {
			RespondError(c, http.StatusUnauthorized, "AUTH_ROLE_MISSING", "Sesión no autenticada", "")
			return
		}
		rol, _ := rolVal.(string)
		if !roleMap[rol] {
			RespondError(c, http.StatusForbidden, "RBAC_FORBIDDEN", "Acceso denegado por política RBAC", fmt.Sprintf("El rol '%s' no tiene permisos para este recurso", rol))
			return
		}
		c.Next()
	}
}
