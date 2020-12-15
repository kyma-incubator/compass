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
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const compassURL = "https://github.com/kyma-incubator/compass"

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`

	ServerTimeout   time.Duration `envconfig:"default=110s"`
	ShutdownTimeout time.Duration `envconfig:"default=10s"`

	Log log.Config

	RootAPI string `envconfig:"APP_ROOT_API,default=/tenants"`

	HandlerEndpoint string `envconfig:"APP_HANDLER_ENDPOINT,default=/v1/callback/{tenantId}"`
	TenantPathParam string `envconfig:"APP_TENANT_PATH_PARAM,default=tenantId"`
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
	subrouter := mainRouter.PathPrefix(cfg.RootAPI).Subrouter()

	logger.Infof("Registering Tenant Onboarding endpoint on %s...", cfg.HandlerEndpoint)
	subrouter.HandleFunc(cfg.HandlerEndpoint, getOnboardingHandlerFunc(cfg.TenantPathParam)).Methods(http.MethodPut)

	logger.Infof("Registering Tenant Decommissioning endpoint on %s...", cfg.HandlerEndpoint)
	subrouter.HandleFunc(cfg.HandlerEndpoint, getDecommissioningHandlerFunc(cfg.TenantPathParam)).Methods(http.MethodDelete)

	logger.Infof("Registering readiness endpoint...")
	subrouter.HandleFunc("/readyz", newReadinessHandler())

	logger.Infof("Registering liveness endpoint...")
	subrouter.HandleFunc("/healthz", newReadinessHandler())

	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, subrouter, "main")

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
