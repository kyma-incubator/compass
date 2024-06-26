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
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/outbound"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/processor"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/service/ucl"
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
	defer func() {
		postgresConnection.Close()
		log.Info().Msgf("Closed postgres connection")
	}()

	healthService := service.HealthService{
		Storage: postgresConnection,
	}

	outboundClientCert, err := outbound.LoadClientCert(cfg.IASConfig.CockpitSecretPath)
	if err != nil {
		log.Fatal().Msgf("Failed to load outbound client cert: %s", err)
	}
	outboundClientConfig := outbound.ClientConfig{
		Certificate: outboundClientCert,
		Timeout:     cfg.IASConfig.RequestTimeout,
	}
	outboundClient := outbound.NewClient(outboundClientConfig)

	tenantMappingsService := service.TenantMappingsService{
		Storage:    postgresConnection,
		IASService: ias.NewService(outboundClient),
	}

	asyncProcessor := processor.AsyncProcessor{
		TenantMappingsService: tenantMappingsService,
		UCLService:            ucl.NewService(outboundClient),
	}

	server, err := api.NewServer(globalCtx, cfg, api.Services{
		HealthService:         healthService,
		TenantMappingsService: tenantMappingsService,
		AsyncProcessor:        asyncProcessor,
	})
	if err != nil {
		log.Fatal().Msgf("Failed to create server: %s", err)
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Msgf("Failed to listen and serve: %s", err)
		}
	}()
	log.Info().Msg("Started server")
	defer func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(timeoutCtx); err != nil {
			log.Fatal().Msgf("Server forced to shutdown: %s", err)
		}
		log.Info().Msgf("Stopped server")
	}()

	<-globalCtx.Done()

	stop()
	log.Info().Msg("Shutting down gracefully")
}
