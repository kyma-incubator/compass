package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

const compassURL = "https://github.com/kyma-incubator/compass"

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`

	ClientTimeout time.Duration `envconfig:"default=105s"`
	ServerTimeout time.Duration `envconfig:"default=110s"`

	Log log.Config

	HandlerEndpoint string `envconfig:"APP_HANDLER_ENDPOINT"`
	TenantPathParam string `envconfig:"APP_TENANT_PATH_PARAM"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	if cfg.HandlerEndpoint == "" || cfg.TenantPathParam == "" {
		exitOnError(errors.New("missing handler endpoint or tenant path parameter"), "Error while loading app handler config")
	}

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Failed to configure Logger")
	logger := log.C(ctx)

	mainRouter := mux.NewRouter()

	logger.Infof("Registering Tenant Onboarding endpoint on %s...", cfg.HandlerEndpoint)
	mainRouter.HandleFunc(cfg.HandlerEndpoint, getOnboardingHandlerFunc(cfg.TenantPathParam)).Methods(http.MethodPut)

	logger.Infof("Registering Tenant Decommissioning endpoint on %s...", cfg.HandlerEndpoint)
	mainRouter.HandleFunc(cfg.HandlerEndpoint, getDecommissioningHandlerFunc(cfg.TenantPathParam)).Methods(http.MethodDelete)

	logger.Infof("Registering readiness endpoint...")
	mainRouter.HandleFunc("/readyz", newReadinessHandler())

	logger.Infof("Registering liveness endpoint...")
	mainRouter.HandleFunc("/healthz", newReadinessHandler())

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
	logger := log.C(ctx)

	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, timeout)
	exitOnError(err, "Error while configuring tenant mapping handler")

	srv := &http.Server{
		Addr:              address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: timeout,
	}

	runFn := func() {
		logger.Infof("Running %s server on %s...", name, address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		logger.Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func getOnboardingHandlerFunc(tenantPathParam string) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger := log.C(request.Context())

		logHandlerRequest("onboarding", tenantPathParam, request)
		if err := logBody(request, writer); err != nil {
			logger.Error(errors.Wrapf(err, "while logging request body"))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte(compassURL)); err != nil {
			logger.Error(errors.Wrapf(err, "while writing response body"))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func getDecommissioningHandlerFunc(tenantPathParam string) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger := log.C(request.Context())

		logHandlerRequest("decommissioning", tenantPathParam, request)
		if err := logBody(request, writer); err != nil {
			logger.Error(errors.Wrapf(err, "while logging request body"))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(writer).Encode(map[string]interface{}{})
		if err != nil {
			logger.Error(errors.Wrapf(err, "while writing to response body"))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func logHandlerRequest(operation, tenantPathParam string, request *http.Request) {
	tenantID := mux.Vars(request)[tenantPathParam]
	log.C(request.Context()).Infof("Performing %s for tenant with id %q", operation, tenantID)
}

func logBody(r *http.Request, w http.ResponseWriter) error {
	logger := log.C(r.Context())

	buf, bodyErr := ioutil.ReadAll(r.Body)
	if bodyErr != nil {
		logger.Info("Body Error: ", bodyErr.Error())
		http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
		return nil
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	logger.Infof("Body: %q", rdr1)
	r.Body = rdr2

	return nil
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}
