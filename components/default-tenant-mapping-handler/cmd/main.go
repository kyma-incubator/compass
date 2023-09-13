package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/default-tenant-mapping-handler/internal/claims"
	"github.com/kyma-incubator/compass/components/default-tenant-mapping-handler/internal/config"
	"github.com/kyma-incubator/compass/components/default-tenant-mapping-handler/internal/handler"
	"github.com/kyma-incubator/compass/components/default-tenant-mapping-handler/internal/healthz"
	"github.com/kyma-incubator/compass/components/default-tenant-mapping-handler/internal/tenant"
	authmiddleware "github.com/kyma-incubator/compass/components/director/pkg/auth-middleware"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/header"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	panicrecovery "github.com/kyma-incubator/compass/components/director/pkg/panic_recovery"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const (
	envPrefix       = "APP"
	healthzEndpoint = "/healthz"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config.Config{}
	err := envconfig.InitWithPrefix(&cfg, envPrefix)
	exitOnError(err, "Error while loading app config")

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Failed to configure Logger")

	fetchJWKSClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	tokenValidationMiddleware := authmiddleware.New(fetchJWKSClient, cfg.JWKSEndpoint, cfg.AllowJWTSigningNone, "", &claims.Validator{})

	tenantValidationMiddleware, err := tenant.NewMiddleware(ctx, cfg.TenantInfo)
	exitOnError(err, "Error while preparing tenant validation middleware")

	mainRouter := mux.NewRouter()
	mainRouter.Use(panicrecovery.NewPanicRecoveryMiddleware(), correlation.AttachCorrelationIDToContext(), log.RequestLogger(healthzEndpoint), header.AttachHeadersToContext())

	defaultTMHandler := mainRouter.PathPrefix(cfg.APIRootPath).Subrouter()
	defaultTMHandler.Use(tokenValidationMiddleware.Handler())
	defaultTMHandler.Use(tenantValidationMiddleware.Handler())

	h := handler.NewHandler()

	defaultTMHandler.HandleFunc(cfg.APITenantMappingsEndpoint, h.HandlerFunc).Methods(http.MethodPatch)
	mainRouter.HandleFunc(healthzEndpoint, healthz.NewHTTPHandler())

	runMainSrv, shutdownMainSrv := createServer(ctx, cfg.Address, mainRouter, "main", cfg.ServerTimeout)

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	runMainSrv()
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func createServer(ctx context.Context, address string, handler http.Handler, name string, timeout time.Duration) (func(), func()) {
	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, timeout)
	exitOnError(err, "Error while configuring tenant mapping handler")

	srv := &http.Server{
		Addr:              address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: timeout,
	}

	runFn := func() {
		log.C(ctx).Infof("Running %s server on %s...", name, address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.C(ctx).WithError(err).Errorf("An error has occurred with %s HTTP server when ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		log.C(ctx).Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(context.Background()); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while shutting down HTTP server %s: %v", name, err)
		}
	}

	return runFn, shutdownFn
}
