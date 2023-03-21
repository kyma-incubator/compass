package api

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/handlers"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal/paths"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/middlewares"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

type Services struct {
	HealthService         handlers.HealthService
	TenantMappingsService handlers.TenantMappingsService
}

func NewServer(ctx context.Context, cfg config.Config, services Services) (*http.Server, error) {
	log := logger.Default()

	router := gin.New()
	routerGroup := router.Group(cfg.APIRootPath)
	routerGroup.Use(gin.Recovery())
	routerGroup.Use(middlewares.Logging)
	authMiddleware, err := middlewares.NewAuthMiddleware(ctx, cfg.TenantInfo)
	if err != nil {
		return nil, errors.Newf("failed to create auth middleware: %w", err)
	}
	routerGroup.Use(authMiddleware.Auth)

	healthHandler := handlers.HealthHandler{Service: services.HealthService}
	routerGroup.GET(paths.HealthPath, healthHandler.Health)

	readyHandler := handlers.ReadyHandler{}
	routerGroup.GET(paths.ReadyPath, readyHandler.Ready)

	tenantMappingsHandler := handlers.TenantMappingsHandler{
		Service: services.TenantMappingsService,
	}
	routerGroup.PATCH(paths.TenantMappingsPath, tenantMappingsHandler.Patch)

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
	}, nil
}
