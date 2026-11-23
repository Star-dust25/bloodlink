package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bloodlink/backend/internal/config"
	"github.com/bloodlink/backend/internal/database"
	"github.com/bloodlink/backend/internal/events"
	"github.com/bloodlink/backend/internal/middleware"
	authHttp "github.com/bloodlink/backend/internal/modules/auth/adapters/inbound/http"
	authPg "github.com/bloodlink/backend/internal/modules/auth/adapters/outbound/postgres"
	authApp "github.com/bloodlink/backend/internal/modules/auth/application"
	donHttp "github.com/bloodlink/backend/internal/modules/donantes/adapters/inbound/http"
	donPg "github.com/bloodlink/backend/internal/modules/donantes/adapters/outbound/postgres"
	donSec "github.com/bloodlink/backend/internal/modules/donantes/adapters/outbound/security"
	donApp "github.com/bloodlink/backend/internal/modules/donantes/application"
	emHttp "github.com/bloodlink/backend/internal/modules/emergencias/adapters/inbound/http"
	emPg "github.com/bloodlink/backend/internal/modules/emergencias/adapters/outbound/postgres"
	emApp "github.com/bloodlink/backend/internal/modules/emergencias/application"
	integSoapIn "github.com/bloodlink/backend/internal/modules/integracion/adapters/inbound/soap"
	integMaps "github.com/bloodlink/backend/internal/modules/integracion/adapters/outbound/maps"
	integPg "github.com/bloodlink/backend/internal/modules/integracion/adapters/outbound/postgres"
	integSoapOut "github.com/bloodlink/backend/internal/modules/integracion/adapters/outbound/soap"
	invHttp "github.com/bloodlink/backend/internal/modules/inventario/adapters/inbound/http"
	invPg "github.com/bloodlink/backend/internal/modules/inventario/adapters/outbound/postgres"
	invApp "github.com/bloodlink/backend/internal/modules/inventario/application"
	notHttp "github.com/bloodlink/backend/internal/modules/notificaciones/adapters/inbound/http"
	notPg "github.com/bloodlink/backend/internal/modules/notificaciones/adapters/outbound/postgres"
	notApp "github.com/bloodlink/backend/internal/modules/notificaciones/application"
	repHttp "github.com/bloodlink/backend/internal/modules/reportes/adapters/inbound/http"
	repPg "github.com/bloodlink/backend/internal/modules/reportes/adapters/outbound/postgres"
	repApp "github.com/bloodlink/backend/internal/modules/reportes/application"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	// 1. Conexión a PostgreSQL 15+ con Pool de Conexiones y Seeder Idempotente
	db, err := database.InitPostgres(cfg)
	if err != nil {
		slog.Error("fallo crítico al inicializar PostgreSQL", "error", err)
		os.Exit(1)
	}

	// 2. Bus de Eventos Asíncrono (IEventBus)
	eventBus := events.NewInMemoryEventBus(1024, 4)

	// 3. Adaptadores de Salida (Outbound Adapters)
	geoCache := integMaps.NewInMemoryGeoCache()
	mapasAdapter := integMaps.NewHaversineMapAdapter(geoCache)
	soapClientAdapter := integSoapOut.NewSOAPClientAdapter(cfg.SOAPEndpointURL)
	legacyStockStore := integPg.NewLegacyStockPostgresStore(db)
	cryptoService, err := donSec.NewAESGCMCryptoService(cfg.AESKeyHex)
	if err != nil {
		slog.Error("error inicializando servicio criptográfico AES-256-GCM", "error", err)
		os.Exit(1)
	}

	usuarioRepo := authPg.NewUsuarioRepository(db)
	inventarioRepo := invPg.NewInventarioRepository(db)
	emergenciaRepo := emPg.NewEmergenciaRepository(db)
	donanteRepo := donPg.NewDonanteRepository(db)
	notificacionRepo := notPg.NewNotificacionRepository(db)
	reporteRepo := repPg.NewReporteRepository(db)

	// 4. Servicios de Aplicación / Casos de Uso (Application Layer)
	authService := authApp.NewAuthService(usuarioRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	inventarioService := invApp.NewSrvGestionarInventario(inventarioRepo, eventBus)
	emergenciaService := emApp.NewSrvAtenderEmergencia(emergenciaRepo, mapasAdapter, soapClientAdapter, eventBus)
	donanteService := donApp.NewSrvRegistrarDonante(donanteRepo, soapClientAdapter, cryptoService)
	notificacionService := notApp.NewSrvNotificaciones(notificacionRepo, eventBus)
	notificacionService.RegistrarSuscripcionesEventos()
	reporteService := repApp.NewSrvReportes(reporteRepo)

	// Disparar verificación inicial de unidades próximas a caducar (RF-08)
	_, _ = notificacionService.EjecutarJobVerificacionCaducidad(context.Background())

	// 5. Adaptadores de Entrada HTTP REST y SOAP 1.1 (Inbound Adapters)
	router := gin.New()
	router.Use(gin.Recovery(), middleware.RequestLogger(), middleware.CORS(cfg.CORSAllowedOrigins))

	healthHandler := func(c *gin.Context) {
		sqlDB, errDB := db.DB()
		dbOK := errDB == nil && sqlDB.Ping() == nil
		busOK := eventBus.IsHealthy()

		status := http.StatusOK
		estadoStr := "UP"
		if !dbOK || !busOK {
			status = http.StatusServiceUnavailable
			estadoStr = "DEGRADED"
		}

		c.JSON(status, gin.H{
			"status":    estadoStr,
			"service":   "bloodlink-backend-modular-monolith",
			"postgres":  dbOK,
			"eventBus":  busOK,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
	router.GET("/health", healthHandler)
	router.HEAD("/health", healthHandler)

	apiV1 := router.Group("/api/v1")
	authMw := middleware.RequireAuth(cfg.JWTSecret)

	authHandler := authHttp.NewAuthHandler(authService)
	authHandler.RegisterRoutes(apiV1, authMw)

	soapHandler := integSoapIn.NewSOAPServerHandler(legacyStockStore, cfg.SOAPEndpointURL)
	soapHandler.RegisterRoutes(router, apiV1)

	invHandler := invHttp.NewInventarioHandler(inventarioService)
	invHandler.RegisterRoutes(apiV1, authMw)

	emHandler := emHttp.NewEmergenciaHandler(emergenciaService)
	emHandler.RegisterRoutes(apiV1, authMw)

	donHandler := donHttp.NewDonanteHandler(donanteService)
	donHandler.RegisterRoutes(apiV1, authMw)

	notHandler := notHttp.NewNotificacionHandler(notificacionService)
	notHandler.RegisterRoutes(apiV1, authMw)

	repHandler := repHttp.NewReporteHandler(reporteService)
	repHandler.RegisterRoutes(apiV1, authMw)

	// 6. Servidor HTTP con Graceful Shutdown (SIGINT / SIGTERM)
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("BloodLink Backend iniciado",
			"port", cfg.Port,
			"rest_api", "http://localhost:"+cfg.Port+"/api/v1",
			"soap_wsdl", "http://localhost:"+cfg.Port+"/soap/v1/interoperabilidad?wsdl",
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("error fatal en servidor HTTP", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("señal de apagado recibida, iniciando graceful shutdown...", "signal", sig.String())

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		slog.Error("error cerrando servidor HTTP", "error", err)
	}

	eventBus.Close()
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
	slog.Info("BloodLink Backend detenido correctamente")
}
