package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/gateway/internal/metrics"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog"
	timeservices "github.com/kyma-incubator/compass/components/gateway/internal/time"
	"github.com/kyma-incubator/compass/components/gateway/internal/uuid"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:3000"`

	ServerTimeout time.Duration `envconfig:"default=114s"`

	Log log.Config

	DirectorOrigin  string `envconfig:"default=http://127.0.0.1:3001"`
	ConnectorOrigin string `envconfig:"default=http://127.0.0.1:3002"`
	MetricsAddress  string `envconfig:"default=127.0.0.1:3003"`
	AuditlogEnabled bool   `envconfig:"default=false"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	router := mux.NewRouter()
	router.Use(correlation.AttachCorrelationIDToContext())
	metricsCollector := metrics.NewAuditlogMetricCollector()
	prometheus.MustRegister(metricsCollector)

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Failed to configure Logger")
	logger := log.C(ctx)

	done := make(chan bool)
	var auditlogSink proxy.AuditlogService
	var auditlogSvc proxy.AuditlogService
	if cfg.AuditlogEnabled {
		logger.Println("Auditlog is enabled")
		auditlogSink, auditlogSvc, err = initAuditLogs(ctx, done, metricsCollector)
		exitOnError(err, "Error while initializing auditlog service")
	} else {
		logger.Println("Auditlog is disabled")
		auditlogSink = &auditlog.NoOpService{}
		auditlogSvc = &auditlog.NoOpService{}
	}

	correlationTr := httputil.NewCorrelationIDTransport(http.DefaultTransport)
	tr := proxy.NewTransport(auditlogSink, auditlogSvc, correlationTr)

	err = proxyRequestsForComponent(ctx, router, "/connector", cfg.ConnectorOrigin, tr)
	exitOnError(err, "Error while initializing proxy for Connector")

	err = proxyRequestsForComponent(ctx, router, "/director", cfg.DirectorOrigin, tr)
	exitOnError(err, "Error while initializing proxy for Director")

	router.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			logger.WithError(err).Error("An error has occurred while writing to response body")
		}
	})

	handlerWithTimeout, err := handler.WithTimeout(router, cfg.ServerTimeout)
	exitOnError(err, "Failed configuring timeout on handler")

	http.Handle("/", handlerWithTimeout)

	server := &http.Server{
		Addr:              cfg.Address,
		ReadHeaderTimeout: cfg.ServerTimeout,
	}

	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())

	runMetricsSrv, shutdownMetricsSrv := createServer(ctx, cfg.MetricsAddress, metricsHandler, "metrics", cfg.ServerTimeout)
	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMetricsSrv()
	}()
	go runMetricsSrv()

	logger.Infof("Listening on %s", cfg.Address)
	if err := server.ListenAndServe(); err != nil {
		done <- true
		panic(err)
	}
}

func proxyRequestsForComponent(ctx context.Context, router *mux.Router, path string, targetOrigin string, transport http.RoundTripper, middleware ...mux.MiddlewareFunc) error {
	log.C(ctx).Infof("Proxying requests on path `%s` to `%s`", path, targetOrigin)
	componentProxy, err := proxy.New(targetOrigin, path, transport)
	if err != nil {
		return errors.Wrapf(err, "while initializing proxy for component")
	}

	connector := router.PathPrefix(path).Subrouter()
	connector.PathPrefix("").HandlerFunc(componentProxy.ServeHTTP)
	connector.Use(middleware...)

	return nil
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
			logger.WithError(err).Errorf("%s HTTP server ListenAndServe", name)
		}
	}

	shutdownFn := func() {
		logger.Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.WithError(err).Errorf("%s HTTP server Shutdown", name)
		}
	}

	return runFn, shutdownFn
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

// initAuditLogs creates proxy.AuditlogService instances, the first one is an asynchronous sink,
// while the second one is a synchronous service with pre-logging functionality.
func initAuditLogs(ctx context.Context, done chan bool, collector *metrics.AuditlogCollector) (proxy.AuditlogService, proxy.AuditlogService, error) {
	cfg := auditlog.Config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		return nil, nil, errors.Wrap(err, "while loading auditlog cfg")
	}

	uuidSvc := uuid.NewService()
	timeSvc := &timeservices.TimeService{}

	var httpClient auditlog.HttpClient
	var msgFactory auditlog.AuditlogMessageFactory

	tr := httputil.NewCorrelationIDTransport(http.DefaultTransport)

	switch cfg.AuthMode {
	case auditlog.Basic:
		{
			var basicCfg auditlog.BasicAuthConfig
			err := envconfig.InitWithPrefix(&basicCfg, "APP")
			if err != nil {
				return nil, nil, errors.Wrap(err, "while loading auditlog basic auth configuration")
			}
			baseHttpClient := &http.Client{
				Transport: tr,
				Timeout:   cfg.ClientTimeout,
			}

			collector.InstrumentAuditlogHTTPClient(baseHttpClient)
			msgFactory = auditlog.NewMessageFactory("proxy", basicCfg.Tenant, uuidSvc, timeSvc)

			httpClient = auditlog.NewBasicAuthClient(basicCfg, baseHttpClient)
		}
	case auditlog.OAuth:
		{
			var oauthCfg auditlog.OAuthConfig
			err := envconfig.InitWithPrefix(&oauthCfg, "APP")
			if err != nil {
				return nil, nil, errors.Wrap(err, "while loading auditlog OAuth configuration")
			}

			ccCfg := fillJWTCredentials(oauthCfg)
			baseClient := &http.Client{
				Transport: tr,
				Timeout:   cfg.ClientTimeout,
			}
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, baseClient)
			client := ccCfg.Client(ctx)

			collector.InstrumentAuditlogHTTPClient(client)
			httpClient = client

			msgFactory = auditlog.NewMessageFactory(oauthCfg.User, oauthCfg.Tenant, uuidSvc, timeSvc)
		}
	default:
		return nil, nil, fmt.Errorf("invalid auditlog auth mode: %s", cfg.AuthMode)
	}

	auditlogClient, err := auditlog.NewClient(cfg, httpClient)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Error while creating auditlog client from cfg")
	}

	auditlogSvc := auditlog.NewService(auditlogClient, msgFactory)
	msgChannel := make(chan proxy.AuditlogMessage, cfg.MsgChannelSize)
	workers := make(chan bool, cfg.WriteWorkers)
	initWorkers(ctx, workers, auditlogSvc, done, msgChannel, collector)

	log.C(ctx).Infof("Auditlog configured successfully, auth mode: %s", cfg.AuthMode)
	return auditlog.NewSink(msgChannel, cfg.MsgChannelTimeout, collector), auditlogSvc, nil
}

func fillJWTCredentials(cfg auditlog.OAuthConfig) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.OAuthURL,
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}
}

func initWorkers(ctx context.Context, workers chan bool, auditlogSvc proxy.AuditlogService, done chan bool, msgChannel chan proxy.AuditlogMessage, collector *metrics.AuditlogCollector) {
	logger := log.C(ctx)

	go func(log *logrus.Entry) {
		for {
			select {
			case <-done:
				log.Info("Worker starter goroutine finished")
				return
			case workers <- true:
			}
			worker := auditlog.NewWorker(auditlogSvc, msgChannel, done, collector)
			go func() {
				log.Info("Starting worker for auditlog message processing")
				worker.Start()
				<-workers
			}()
		}
	}(logger)
}
