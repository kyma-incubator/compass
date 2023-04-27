package main

import (
	"context"
	"github.com/gorilla/mux"
	authpkg "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/api/handlers"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/api/paths"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/external_caller"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/server"
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

	handler := initAPIHandlers(ctx, &cfg)
	mainSrv := server.NewServerWithHandler(cfg.Server, handler)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go mainSrv.Start(ctx, wg)

	wg.Wait()
}

func initAPIHandlers(ctx context.Context, cfg *config.Config) http.Handler {
	logger := log.C(ctx)

	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(paths.LivezEndpoint, paths.ReadyzEndpoint))

	tmRouter := mainRouter.PathPrefix(cfg.Server.RootAPIPath).Subrouter()
	healthCheckRouter := mainRouter.PathPrefix(cfg.Server.RootAPIPath).Subrouter()

	securedHTTPClient := authpkg.PrepareHTTPClientWithSSLValidation(cfg.HTTPClient.Timeout, cfg.HTTPClient.SkipSSLValidation)
	caller := external_caller.NewCaller(securedHTTPClient, cfg.OAuthProvider)

	logger.Infof("Registering tenant mapping endpoints...")
	tmHandler := handlers.NewHandler(cfg, caller)
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
