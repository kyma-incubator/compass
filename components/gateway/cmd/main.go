package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/gateway/internal/metrics"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
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

	done := make(chan bool)
	var auditlogSink proxy.AuditlogService
	var auditlogSvc proxy.AuditlogService
	if cfg.AuditlogEnabled {
		log.Println("Auditlog is enabled")
		auditlogSink, auditlogSvc, err = initAuditLogs(done, metricsCollector)
		exitOnError(err, "Error while initializing auditlog service")
	} else {
		log.Println("Auditlog is disabled")
		auditlogSink = &auditlog.NoOpService{}
		auditlogSvc = &auditlog.NoOpService{}
	}

	correlationTr := httputil.NewCorrelationIDTransport(http.DefaultTransport)
	tr := proxy.NewTransport(auditlogSink, auditlogSvc, correlationTr)

	err = proxyRequestsForComponent(router, "/connector", cfg.ConnectorOrigin, tr)
	exitOnError(err, "Error while initializing proxy for Connector")

	err = proxyRequestsForComponent(router, "/director", cfg.DirectorOrigin, tr)
	exitOnError(err, "Error while initializing proxy for Director")

	router.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			log.Println(errors.Wrapf(err, "while writing to response body").Error())
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

	runMetricsSrv, shutdownMetricsSrv := createServer(cfg.MetricsAddress, metricsHandler, "metrics", cfg.ServerTimeout)
	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMetricsSrv()
	}()
	go runMetricsSrv()

	log.Printf("Listening on %s", cfg.Address)
	if err := server.ListenAndServe(); err != nil {
		done <- true
		panic(err)
	}
}

func proxyRequestsForComponent(router *mux.Router, path string, targetOrigin string, transport http.RoundTripper, middleware ...mux.MiddlewareFunc) error {
	log.Printf("Proxying requests on path `%s` to `%s`", path, targetOrigin)

	componentProxy, err := proxy.New(targetOrigin, path, transport)
	if err != nil {
		return errors.Wrapf(err, "while initializing proxy for component")
	}

	connector := router.PathPrefix(path).Subrouter()
	connector.PathPrefix("").HandlerFunc(componentProxy.ServeHTTP)
	connector.Use(middleware...)

	return nil
}

func createServer(address string, handler http.Handler, name string, timeout time.Duration) (func(), func()) {
	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, timeout)
	exitOnError(err, "Error while configuring tenant mapping handler")

	srv := &http.Server{
		Addr:              address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: timeout,
	}

	runFn := func() {
		logrus.Infof("Running %s server on %s...", name, address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logrus.Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		logrus.Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(context.Background()); err != nil {
			logrus.Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

// initAuditLogs creates proxy.AuditlogService instances, the first one is an asynchronous sink,
// while the second one is a synchronous service with pre-logging functionality.
func initAuditLogs(done chan bool, collector *metrics.AuditlogCollector) (proxy.AuditlogService, proxy.AuditlogService, error) {
	cfg := auditlog.Config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		return nil, nil, errors.Wrap(err, "while loading auditlog cfg")
	}

	uuidSvc := uuid.NewService()
	timeSvc := &timeservices.TimeService{}

	var httpClient auditlog.HttpClient
	var msgFactory auditlog.AuditlogMessageFactory

	switch cfg.AuthMode {
	case auditlog.Basic:
		{
			var basicCfg auditlog.BasicAuthConfig
			err := envconfig.InitWithPrefix(&basicCfg, "APP")
			if err != nil {
				return nil, nil, errors.Wrap(err, "while loading auditlog basic auth configuration")
			}

			msgFactory = auditlog.NewMessageFactory("proxy", basicCfg.Tenant, uuidSvc, timeSvc)
			tr := httputil.NewCorrelationIDTransport(http.DefaultTransport)
			baseHttpClient := &http.Client{
				Transport: tr,
				Timeout:   cfg.ClientTimeout,
			}

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
				Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
				Timeout:   cfg.ClientTimeout,
			}
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, baseClient)
			client := ccCfg.Client(ctx)

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
	initWorkers(workers, auditlogSvc, done, msgChannel, collector)

	log.Printf("Auditlog configured successfully, auth mode:%s", cfg.AuthMode)
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

func initWorkers(workers chan bool, auditlogSvc proxy.AuditlogService, done chan bool, msgChannel chan proxy.AuditlogMessage, collector *metrics.AuditlogCollector) {
	go func() {
		for {
			select {
			case <-done:
				log.Println("Worker starter goroutine finished")
				return
			case workers <- true:
			}
			worker := auditlog.NewWorker(auditlogSvc, msgChannel, done, collector)
			go func() {
				log.Println("Starting worker for auditlog message processing")
				worker.Start()
				<-workers
			}()
		}
	}()
}
