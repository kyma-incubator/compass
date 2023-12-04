package main

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/api/handlers"
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/api/paths"
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/server"
	httputildirector "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg, err := config.New()
	exitOnError(err, "Error while loading app config")

	ctx, err = log.Configure(ctx, cfg.Log)
	exitOnError(err, "Failed to configure Logger")

	err = cfg.MapInstanceConfigs()
	exitOnError(err, "Failed to load configuration")

	handler := initAPIHandlers(ctx, cfg)
	mainSrv := server.NewServerWithHandler(cfg.Server, handler)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go mainSrv.Start(ctx, wg)

	wg.Wait()
}

func initAPIHandlers(ctx context.Context, cfg config.Config) http.Handler {
	logger := log.C(ctx)

	certCache, err := credloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	exitOnError(err, "failed to initialize certificate loader")

	mtlsHTTPClient := httputildirector.PrepareMTLSClientWithSSLValidation(cfg.HTTPClient.Timeout, certCache, cfg.HTTPClient.SkipSSLValidation, cfg.ExternalClientCertSecretName)

	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(paths.LivezEndpoint, paths.ReadyzEndpoint))

	tmRouter := mainRouter.PathPrefix(cfg.Server.RootAPIPath).Subrouter()
	healthCheckRouter := mainRouter.PathPrefix(cfg.Server.RootAPIPath).Subrouter()

	logger.Infof("Registering tenant mapping endpoints...")
	tmHandler := handlers.NewHandler(cfg, mtlsHTTPClient)
	tmRouter.Handle(cfg.Server.TenantMappingAPIEndpoint, tmHandler).Methods(http.MethodPatch)

	logger.Info("Registering liveness endpoint...")
	healthCheckRouter.Handle(paths.LivezEndpoint, readinessHandler())

	logger.Info("Registering readiness endpoint...")
	healthCheckRouter.Handle(paths.ReadyzEndpoint, readinessHandler())

	logger.Info("Registering health endpoint...")
	healthCheckRouter.Handle(paths.HealthzEndpoint, readinessHandler())

	return mainRouter
}

func readinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}
