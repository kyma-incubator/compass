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
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/panic_recovery"
)

type Server struct {
	*http.Server

	healthy         int32
	shutdownTimeout time.Duration
}

func New(c *Config, middlewares []mux.MiddlewareFunc, routesProvider ...func(router *mux.Router)) *Server {
	s := &Server{
		shutdownTimeout: c.ShutdownTimeout,
	}

	router := mux.NewRouter()

	const (
		healthzEndpoint = "/healthz"
		readyzEndpoint  = "/readyz"
	)

	router.Handle(healthzEndpoint, s.livenessHandler())
	router.Handle(readyzEndpoint, s.readinessHandler())

	router.Use(correlation.AttachCorrelationIDToContext())
	router.Use(log.RequestLogger(healthzEndpoint, readyzEndpoint))
	router.Use(panic_recovery.NewRecoveryMiddleware())

	for _, m := range middlewares {
		router.Use(m)
	}

	for _, applyRoutes := range routesProvider {
		applyRoutes(router)
	}

	handlerWithTimeout, err := handler.WithTimeout(router, c.Timeout)
	if err != nil {
		log.D().Fatalf("Could not create timeout handler: %v\n", err)
	}

	s.Server = &http.Server{
		Addr:    ":" + strconv.Itoa(c.Port),
		Handler: handlerWithTimeout,

		ReadTimeout:  c.Timeout,
		WriteTimeout: c.Timeout,
		IdleTimeout:  c.Timeout,
	}

	return s
}

func NewServerWithRouter(c *Config, router *mux.Router) *Server {
	s := &Server{
		shutdownTimeout: c.ShutdownTimeout,
	}
	handlerWithTimeout, err := handler.WithTimeout(router, c.Timeout)
	if err != nil {
		log.D().Fatalf("Could not create timeout handler: %v\n", err)
	}
	s.Server = &http.Server{
		Addr:         ":" + strconv.Itoa(c.Port),
		Handler:      handlerWithTimeout,
		ReadTimeout:  c.Timeout,
		WriteTimeout: c.Timeout,
		IdleTimeout:  c.Timeout,
	}
	return s
}

func (s *Server) Start(parentCtx context.Context, wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	idleConnsClosed := make(chan struct{})
	go func() {
		defer wg.Done()
		<-ctx.Done()
		s.stop()
		close(idleConnsClosed)
	}()

	log.C(ctx).Infof("Starting and listening on %s://%s", "http", s.Addr)

	atomic.StoreInt32(&s.healthy, 1)

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.C(ctx).Fatalf("Could not listen on %s://%s: %v\n", "http", s.Addr, err)
	}

	<-idleConnsClosed
}

func (s *Server) stop() {
	atomic.StoreInt32(&s.healthy, 0)

	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	s.SetKeepAlivesEnabled(false)

	if err := s.Shutdown(ctx); err != nil {
		log.C(ctx).Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}

func (s *Server) livenessHandler() http.Handler {
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

func (s *Server) readinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
