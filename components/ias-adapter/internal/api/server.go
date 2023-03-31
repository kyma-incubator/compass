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
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/jwk"
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
	routerGroup.GET(paths.HealthPath, handlers.HealthHandler{Service: services.HealthService}.Health)
	routerGroup.GET(paths.ReadyPath, handlers.ReadyHandler{}.Ready)

	tenantMappingRouter := routerGroup.Group("")

	jwkCache, err := jwk.NewJWKCache(ctx, cfg.JWKCache)
	if err != nil {
		return nil, errors.Newf("failed to create jwk cache: %w", err)
	}
	jwtMiddleware := middlewares.NewJWTMiddleware(jwkCache)
	tenantMappingRouter.Use(jwtMiddleware.JWT)
	authMiddleware, err := middlewares.NewAuthMiddleware(ctx, cfg.TenantInfo)
	if err != nil {
		return nil, errors.Newf("failed to create auth middleware: %w", err)
	}
	routerGroup.Use(authMiddleware.Auth)
	tenantMappingsHandler := handlers.TenantMappingsHandler{
		Service: services.TenantMappingsService,
	}
	tenantMappingRouter.PATCH(paths.TenantMappingsPath, tenantMappingsHandler.Patch)

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
