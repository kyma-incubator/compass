package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/ias"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/storage/postgres"
)

func Start(cfg config.Config) {
	log := logger.Default()

	globalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	log.Info().Msg("Initialized global context")

	postgresConnection, err := postgres.NewConnection(globalCtx, cfg.Postgres)
	if err != nil {
		log.Fatal().Msgf("Failed to create postgres connection: %s", err)
	}
	log.Info().Msg("Opened postgres connection")

	healthService := service.HealthService{
		Storage: postgresConnection,
	}

	iasClient, err := ias.NewClient(cfg.IASConfig)
	if err != nil {
		log.Fatal().Msgf("Failed to create IAS HTTPS client: %s", err)
	}

	tenantMappingsService := service.TenantMappingsService{
		Storage:    postgresConnection,
		IASService: ias.NewService(cfg.IASConfig, iasClient),
	}

	server := api.NewServer(cfg, api.Services{
		HealthService:         healthService,
		TenantMappingsService: tenantMappingsService,
	})
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("Failed to listen and serve: %s", err)
		}
	}()
	log.Info().Msg("Started server")

	<-globalCtx.Done()

	stop()
	log.Info().Msg("Shutting down gracefully")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(timeoutCtx); err != nil {
		log.Fatal().Msgf("Server forced to shutdown: %s", err)
	}
	log.Info().Msgf("Stopped server")

	postgresConnection.Close()
	log.Info().Msgf("Closed postgres connection")
}
