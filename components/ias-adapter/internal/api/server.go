package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/handlers"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/paths"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/middlewares"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type Services struct {
	HealthService         handlers.HealthService
	TenantMappingsService handlers.TenantMappingsService
}

func NewServer(cfg config.Config, services Services) *http.Server {
	log := logger.Default()

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middlewares.Logging)
	healthHandler := handlers.HealthsHandler{
		Service: services.HealthService,
	}
	router.GET(paths.HealthPath, healthHandler.Health)
	tenantMappingsHandler := handlers.TenantMappingsHandler{
		Service: services.TenantMappingsService,
	}
	router.PATCH(paths.TenantMappingsPath, tenantMappingsHandler.Patch)

	routes := router.Routes()
	for _, route := range routes {
		log.Info().Msgf("%s %s", route.Method, route.Path)
	}

	return &http.Server{
		Addr:              cfg.Address,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		Handler:           router,
	}
}
