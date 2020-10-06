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

package server

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/panic_recovery"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"net/http/pprof"
	"strconv"
	"sync/atomic"
	"time"
)

type Server struct {
	*http.Server

	healthy         int32
	routesProvider  []func(router *mux.Router)
	shutdownTimeout time.Duration
}

func New(c *Config, service log.UUIDService, routesProvider ...func(router *mux.Router)) *Server {
	s := &Server{
		shutdownTimeout: c.ShutdownTimeout,
		routesProvider:  routesProvider,
	}

	router := mux.NewRouter()
	router.Handle(c.RootAPI+"/metrics", promhttp.Handler())
	router.Handle(c.RootAPI+"/healthz", s.healthHandler())

	router.HandleFunc(c.RootAPI+"/debug/pprof/", pprof.Index)
	router.HandleFunc(c.RootAPI+"/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc(c.RootAPI+"/debug/pprof/profile", pprof.Profile)
	router.HandleFunc(c.RootAPI+"/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc(c.RootAPI+"/debug/pprof/trace", pprof.Trace)

	router.Use(log.RequestLogger(service))
	router.Use(panic_recovery.NewRecoveryMiddleware())

	for _, applyRoutes := range routesProvider {
		applyRoutes(router)
	}

	s.Server = &http.Server{
		Addr:    ":" + strconv.Itoa(c.Port),
		Handler: router,

		ReadTimeout:  c.RequestTimeout,
		WriteTimeout: c.RequestTimeout,
		IdleTimeout:  c.RequestTimeout,
	}

	return s
}

func (s *Server) Start(parentCtx context.Context) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	idleConnsClosed := make(chan struct{})
	go func() {
		<-ctx.Done()
		s.stop(parentCtx)
		close(idleConnsClosed)
	}()

	log.C(ctx).Infof("Starting and listening on %s://%s", "http", s.Addr)

	atomic.StoreInt32(&s.healthy, 1)

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.C(ctx).Fatalf("Could not listen on %s://%s: %v\n", "http", s.Addr, err)
	}

	<-idleConnsClosed
}

func (s *Server) stop(parentCtx context.Context) {
	atomic.StoreInt32(&s.healthy, 0)

	ctx, cancel := context.WithTimeout(parentCtx, s.shutdownTimeout)
	defer cancel()

	go func(ctx context.Context) {
		<-ctx.Done()

		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			log.C(ctx).Panic("Timeout while stopping the server, killing instance!")
		}
	}(ctx)

	s.SetKeepAlivesEnabled(false)

	if err := s.Shutdown(ctx); err != nil {
		log.C(ctx).Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}

func (s *Server) healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statusCode := http.StatusServiceUnavailable
		state := "failed"
		if atomic.LoadInt32(&s.healthy) == 1 {
			statusCode = http.StatusOK
			state = "success"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if _, err := w.Write([]byte(fmt.Sprintf(`{"status": "%s"}`, state))); err != nil {
			log.C(r.Context()).Error("Error sending data", err)
		}
	})
}
