/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	auth "github.com/kyma-incubator/compass/components/tenant-fetcher/internal/authenticator"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/tenant"
	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/uuid"

	"github.com/gorilla/mux"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const compassURL = "https://github.com/kyma-incubator/compass"
const envPrefix = "APP"

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`

	Database persistence.DatabaseConfig

	ServerTimeout   time.Duration `envconfig:"default=110s"`
	ShutdownTimeout time.Duration `envconfig:"default=10s"`

	Log log.Config

	RootAPI string `envconfig:"APP_ROOT_API,default=/tenants"`

	HandlerEndpoint string `envconfig:"APP_HANDLER_ENDPOINT,default=/v1/callback/{tenantId}"`
	TenantPathParam string `envconfig:"APP_TENANT_PATH_PARAM,default=tenantId"`

	JwksEndpoint              string `envconfig:"APP_JWKS_ENDPOINT"`
	CISIdentityZone           string `envconfig:"APP_CIS_IDENTITY_ZONE"`
	SubscriptionCallbackScope string `envconfig:"APP_SUBSCRIPTION_CALLBACK_SCOPE"`
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

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	authenticatorsConfig, err := authenticator.InitFromEnv(envPrefix)
	exitOnError(err, "Failed to retrieve authenticators config")
	log.C(ctx).Infof("%+v", authenticatorsConfig)

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	uidSvc := uuid.NewService()
	converter := tenant.NewConverter()
	repo := tenant.NewRepository(converter)
	service := tenant.NewService(repo, transact, uidSvc)

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Failed to configure Logger")
	logger := log.C(ctx)

	middleware := auth.New(cfg.JwksEndpoint, cfg.CISIdentityZone, cfg.SubscriptionCallbackScope, extractTrustedIssuersScopePrefixes(authenticatorsConfig))

	mainRouter := mux.NewRouter()
	subrouter := mainRouter.PathPrefix(cfg.RootAPI).Subrouter()
	subrouter.Use(middleware.Handler())

	logger.Infof("Registering Tenant Onboarding endpoint on %s...", cfg.HandlerEndpoint)
	subrouter.HandleFunc(cfg.HandlerEndpoint, getOnboardingHandlerFunc(service, cfg.TenantPathParam)).Methods(http.MethodPut)

	logger.Infof("Registering Tenant Decommissioning endpoint on %s...", cfg.HandlerEndpoint)
	subrouter.HandleFunc(cfg.HandlerEndpoint, getDecommissioningHandlerFunc(service, cfg.TenantPathParam)).Methods(http.MethodDelete)

	healthCheckSubrouter := mainRouter.PathPrefix(cfg.RootAPI).Subrouter()
	logger.Infof("Registering readiness endpoint...")
	healthCheckSubrouter.HandleFunc("/readyz", newReadinessHandler())

	logger.Infof("Registering liveness endpoint...")
	healthCheckSubrouter.HandleFunc("/healthz", newReadinessHandler())

	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, mainRouter, "main")

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

func createServer(ctx context.Context, cfg config, handler http.Handler, name string) (func(), func()) {
	logger := log.C(ctx)

	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, cfg.ServerTimeout)
	exitOnError(err, "Error while configuring tenant mapping handler")

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: cfg.ServerTimeout,
	}

	runFn := func() {
		logger.Infof("Running %s server on %s...", name, cfg.Address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		logger.Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func getOnboardingHandlerFunc(svc tenant.TenantService, tenantPathParam string) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger := log.C(request.Context())

		logHandlerRequest("onboarding", tenantPathParam, request)
		body, err := extractBody(request, writer)
		if err != nil {
			logger.Error(errors.Wrapf(err, "while extracting request body"))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		var tenant model.TenantModel
		if err := json.Unmarshal(body, &tenant); err != nil {
			logger.Error(errors.Wrapf(err, "while unmarshalling body"))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := svc.Create(request.Context(), tenant); err != nil {
			logger.Error(errors.Wrapf(err, "while creating tenant"))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "text/plain")
		writer.WriteHeader(http.StatusOK)
		if _, err := writer.Write([]byte(compassURL)); err != nil {
			logger.Error(errors.Wrapf(err, "while writing response body"))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func getDecommissioningHandlerFunc(svc tenant.TenantService, tenantPathParam string) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger := log.C(request.Context())

		logHandlerRequest("decommissioning", tenantPathParam, request)

		body, err := extractBody(request, writer)

		var tenant model.TenantModel
		if err := json.Unmarshal(body, &tenant); err != nil {
			logger.Error(errors.Wrapf(err, "while unmarshalling body"))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := svc.DeleteByTenant(request.Context(), tenant.GlobalAccountGUID); err != nil {
			logger.Error(errors.Wrapf(err, "while deleting tenant"))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(writer).Encode(map[string]interface{}{})
		if err != nil {
			logger.Error(errors.Wrapf(err, "while writing to response body"))
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func logHandlerRequest(operation, tenantPathParam string, request *http.Request) {
	tenantID := mux.Vars(request)[tenantPathParam]
	log.C(request.Context()).Infof("Performing %s for tenant with id %q", operation, tenantID)
}

func extractBody(r *http.Request, w http.ResponseWriter) ([]byte, error) {
	logger := log.C(r.Context())

	buf, bodyErr := ioutil.ReadAll(r.Body)
	if bodyErr != nil {
		logger.Info("Body Error: ", bodyErr.Error())
		http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
		return nil, bodyErr
	}

	return buf, nil
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func extractTrustedIssuersScopePrefixes(config []authenticator.Config) []string {
	var prefixes []string

	for _, authenticator := range config {
		if len(authenticator.TrustedIssuers) == 0 {
			continue
		}

		for _, trustedIssuers := range authenticator.TrustedIssuers {
			prefixes = append(prefixes, trustedIssuers.ScopePrefix)
		}
	}

	return prefixes
}
