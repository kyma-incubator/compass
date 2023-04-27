package server

import (
	"context"
	"fmt"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	*http.Server

	healthy         int32
	shutdownTimeout time.Duration
}

func NewServer(ctx context.Context, handler http.Handler, serverAddress, name string, serverTimeout, shutdownTimeout time.Duration) (func(), func()) {
	logger := log.C(ctx)

	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, serverTimeout)
	if err != nil {
		wrappedError := errors.Wrap(err, "error while configuring handler with timeout")
		log.D().Fatal(wrappedError)
	}

	srv := &http.Server{
		Addr:              serverAddress,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: serverTimeout,
	}

	runFn := func() {
		logger.Infof("Running %s server on %s...", name, serverAddress)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		logger.Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func NewServerWithHandler(c *Config, handler http.Handler) *Server {
	s := &Server{
		shutdownTimeout: c.ShutdownTimeout,
	}

	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, c.Timeout)
	if err != nil {
		wrappedError := errors.Wrap(err, "Could not create timeout handler")
		log.D().Fatal(wrappedError)
	}

	s.Server = &http.Server{
		Addr:              ":" + strconv.Itoa(c.Port),
		Handler:           handlerWithTimeout,
		ReadTimeout:       c.ReadTimeout,
		ReadHeaderTimeout: c.ReadHeaderTimeout,
		WriteTimeout:      c.WriteTimeout,
		IdleTimeout:       c.IdleTimeout,
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
