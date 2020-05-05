package server

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/pprof"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	serverReadTimeout  = 5 * time.Second
	serverWriteTimeout = 10 * time.Second
	serverIdleTimeout  = 15 * time.Second
	serverStopTimeout  = 30 * time.Second

	sessionCacheCapacity = 128
)

// List of default ciphers to use.
// https://golang.org/pkg/crypto/tls/#pkg-constants
var defaultCiphers = []uint16{
	tls.TLS_FALLBACK_SCSV, // TLS_FALLBACK_SCSV should always be first, see https://tools.ietf.org/html/rfc7507#section-6
	tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
}

// List of default curves to use.
var defaultCurves = []tls.CurveID{
	tls.X25519,
	tls.CurveP256,
}

type Config struct {
	ListenAddr  string `kong:"help='Address and port for the server to listen on.',env='METRIS_LISTENADDR',required=true"`
	TLSCertFile string `kong:"help='Path to TLS certificate file.',type='path',env='METRIS_TLSCERTFILE'"`
	TLSKeyFile  string `kong:"help='Path to TLS key file.',type='path',env='METRIS_TLSKEYFILE'"`
}

// Server represents an HTTP server
type Server struct {
	Server  *http.Server
	Healthy int32
	useTLS  bool
	logger  *zap.SugaredLogger
}

func NewServer(c Config, logger *zap.SugaredLogger) (*Server, error) {
	s := &Server{
		logger: logger.With("component", "metris"),
	}

	tlsConfig := &tls.Config{}

	if c.TLSCertFile != "" && c.TLSKeyFile != "" {
		logger.Debug("TLS Configuration found")

		cert, err := tls.LoadX509KeyPair(c.TLSCertFile, c.TLSKeyFile)
		if err != nil {
			return nil, err
		}

		tlsConfig = &tls.Config{
			Certificates:             []tls.Certificate{cert},
			MinVersion:               tls.VersionTLS12,
			MaxVersion:               tls.VersionTLS13,
			CurvePreferences:         defaultCurves,
			PreferServerCipherSuites: true,
			CipherSuites:             defaultCiphers,
			ClientSessionCache:       tls.NewLRUClientSessionCache(sessionCacheCapacity),
		}
		s.useTLS = true
	}

	router := http.NewServeMux()
	router.Handle("/metrics", promhttp.Handler())
	router.Handle("/healthz", s.HealthHandler())

	// adding go profiling tools
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)

	s.Server = &http.Server{
		Addr:      c.ListenAddr,
		TLSConfig: tlsConfig,
		Handler:   router,

		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
		IdleTimeout:  serverIdleTimeout,
	}

	return s, nil
}

func (s *Server) Start(parentctx context.Context, parentwg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(parentctx)
	defer cancel()

	parentwg.Add(1)
	defer parentwg.Done()

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	proto := "http"
	if s.useTLS {
		proto = "https"
	}

	s.logger.Infof("Starting and listening on %s://%s", proto, s.Server.Addr)

	atomic.StoreInt32(&s.Healthy, 1)

	if s.useTLS {
		if err := s.Server.ListenAndServeTLS("", ""); err != http.ErrServerClosed {
			s.logger.Fatalf("Could not listen on %s://%s: %v\n", proto, s.Server.Addr, err)
		}
	} else {
		if err := s.Server.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Fatalf("Could not listen on %s://%s: %v\n", proto, s.Server.Addr, err)
		}
	}
}

func (s *Server) Stop() {
	atomic.StoreInt32(&s.Healthy, 0)

	ctx, cancel := context.WithTimeout(context.Background(), serverStopTimeout)
	defer cancel()

	go func(ctx context.Context) {
		<-ctx.Done()

		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			s.logger.Panic("Timeout while stopping the server, killing instance!")
		}
	}(ctx)

	s.Server.SetKeepAlivesEnabled(false)

	if err := s.Server.Shutdown(ctx); err != nil {
		s.logger.Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}

func (s *Server) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		statuscode := http.StatusServiceUnavailable
		if atomic.LoadInt32(&s.Healthy) == 1 {
			statuscode = http.StatusOK
		}
		w.WriteHeader(statuscode)
	})
}
