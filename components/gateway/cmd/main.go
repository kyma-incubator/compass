package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/gateway/internal/metrics"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
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
	"github.com/vrischmann/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

//
type config struct {
	Address string `envconfig:"default=127.0.0.1:3000"`

	MetricsServerTimeout      time.Duration `envconfig:"APP_METRICS_SERVER_TIMEOUT,default=114s"`
	NsAdapterTimeout          time.Duration `envconfig:"APP_NS_ADAPTER_TIMEOUT,default=3600s"`
	DefaultHandlerTimeout     time.Duration `envconfig:"APP_DEFAULT_HANDLERS_TIMEOUT,default=114s"`
	ReadRequestHeadersTimeout time.Duration `envconfig:"APP_READ_REQUEST_HEADERS_TIMEOUT,default=114s"`

	Log log.Config

	DirectorOrigin  string `envconfig:"default=http://127.0.0.1:3001"`
	ConnectorOrigin string `envconfig:"default=http://127.0.0.1:3002"`
	NsadapterOrigin string `envconfig:"default=http://127.0.0.1:3005"`
	MetricsAddress  string `envconfig:"default=127.0.0.1:3003"`
	AuditlogEnabled bool   `envconfig:"default=false"`

	AuditLogMessageBodySizeLimit int `envconfig:"APP_AUDIT_LOG_MSG_BODY_SIZE_LIMIT"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	const healthzEndpoint = "/healthz"
	router := mux.NewRouter()
	router.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(healthzEndpoint))
	metricsCollector := metrics.NewAuditlogMetricCollector()
	prometheus.MustRegister(metricsCollector)

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Failed to configure Logger")
	logger := log.C(ctx)

	var auditlogSink proxy.AuditlogService
	var auditlogSvc proxy.PreAuditlogService
	if cfg.AuditlogEnabled {
		logger.Infoln("Auditlog is enabled")
		auditlogSink, auditlogSvc, err = initAuditLogs(ctx, metricsCollector)
		exitOnError(err, "Error while initializing auditlog service")
	} else {
		logger.Infoln("Auditlog is disabled")
		auditlogSink = &auditlog.NoOpService{}
		auditlogSvc = &auditlog.NoOpService{}
	}

	correlationTr := httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport)))
	tr := proxy.NewTransport(auditlogSink, auditlogSvc, correlationTr)
	adapterCfg := proxy.AdapterConfig{
		MsgBodySizeLimit: cfg.AuditLogMessageBodySizeLimit,
	}
	adapterTr := proxy.NewAdapterTransport(auditlogSink, auditlogSvc, correlationTr, adapterCfg)

	err = proxyRequestsForComponent(ctx, router, "/connector", cfg.ConnectorOrigin, tr, cfg.DefaultHandlerTimeout)
	exitOnError(err, "Error while initializing proxy for Connector")

	err = proxyRequestsForComponent(ctx, router, "/director", cfg.DirectorOrigin, tr, cfg.DefaultHandlerTimeout)
	exitOnError(err, "Error while initializing proxy for Director")

	err = proxyRequestsForComponent(ctx, router, "/nsadapter", cfg.NsadapterOrigin, adapterTr, cfg.NsAdapterTimeout)
	exitOnError(err, "Error while initializing proxy for NSAdapter")

	router.HandleFunc(healthzEndpoint, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			logger.WithError(err).Error("An error has occurred while writing to response body")
		}
	})

	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())

	metricsServer := createServer(cfg.MetricsAddress, metricsHandler, cfg.MetricsServerTimeout)
	mainServer := createServer(cfg.Address, router, cfg.ReadRequestHeadersTimeout)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go startServer(ctx, metricsServer, "metrics", wg)
	go startServer(ctx, mainServer, "main", wg)

	wg.Wait()
}

func proxyRequestsForComponent(ctx context.Context, router *mux.Router, path string, targetOrigin string, transport http.RoundTripper, timeout time.Duration, middleware ...mux.MiddlewareFunc) error {
	log.C(ctx).Infof("Proxying requests on path `%s` to `%s`", path, targetOrigin)

	componentProxy, err := proxy.New(targetOrigin, path, transport)
	if err != nil {
		return errors.Wrapf(err, "while initializing proxy for component")
	}

	handlerWithTimeout, err := timeouthandler.WithTimeout(componentProxy, timeout)
	if err != nil {
		return errors.Wrapf(err, "while initializing timeout handler for component")
	}

	connector := router.PathPrefix(path).Subrouter()
	connector.PathPrefix("").HandlerFunc(handlerWithTimeout.ServeHTTP)
	connector.Use(middleware...)

	return nil
}

func createServer(address string, handler http.Handler, readHeadersTimeout time.Duration) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: readHeadersTimeout,
	}
}

func startServer(parentCtx context.Context, server *http.Server, name string, wg *sync.WaitGroup) {
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	go func() {
		defer wg.Done()
		<-ctx.Done()
		stopServer(server)
	}()

	log.C(ctx).Infof("Running %s server on %s...", name, server.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.C(ctx).Fatalf("Could not listen on %s://%s: %v\n", "http", server.Addr, err)
	}
}

func stopServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		<-ctx.Done()

		if ctx.Err() == context.Canceled {
			return
		} else if ctx.Err() == context.DeadlineExceeded {
			log.C(ctx).Panic("Timeout while stopping the server, killing instance!")
		}
	}(ctx)

	server.SetKeepAlivesEnabled(false)

	if err := server.Shutdown(ctx); err != nil {
		log.C(ctx).Fatalf("Could not gracefully shutdown the server: %v\n", err)
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

// initAuditLogs creates proxy.AuditlogService instances, the first one is an asynchronous sink,
// while the second one is a synchronous service with pre-logging functionality.
func initAuditLogs(ctx context.Context, collector *metrics.AuditlogCollector) (proxy.AuditlogService, proxy.PreAuditlogService, error) {
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
			baseHttpClient := &http.Client{
				Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
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
				Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
				Timeout:   cfg.ClientTimeout,
			}
			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, baseClient)
			client := ccCfg.Client(ctx)

			collector.InstrumentAuditlogHTTPClient(client)
			httpClient = client

			msgFactory = auditlog.NewMessageFactory(oauthCfg.User, oauthCfg.Tenant, uuidSvc, timeSvc)
		}
	case auditlog.OAuthMtls:
		{
			var mtlsConfig auditlog.OAuthMtlsConfig
			if err = envconfig.InitWithPrefix(&mtlsConfig, "APP"); err != nil {
				return nil, nil, errors.Wrap(err, "while loading auditlog oauth-mTLS configuration")
			}

			cert, err := mtlsConfig.ParseCertificate()
			if err != nil {
				return nil, nil, errors.Wrap(err, "while loading x509 cert for mTLS")
			}

			transport := httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(&http.Transport{
				TLSClientConfig: &tls.Config{
					Certificates:       []tls.Certificate{*cert},
					InsecureSkipVerify: mtlsConfig.SkipSSLValidation,
				},
			}))

			mtlsClient := &http.Client{
				Transport: transport,
				Timeout:   cfg.ClientTimeout,
			}

			ctx := context.WithValue(context.Background(), oauth2.HTTPClient, mtlsClient)
			ccCfg := fillOAuthMtlsCredentials(mtlsConfig)
			client := ccCfg.Client(ctx)
			collector.InstrumentAuditlogHTTPClient(client)

			httpClient = client
			msgFactory = auditlog.NewMessageFactory(mtlsConfig.User, mtlsConfig.Tenant, uuidSvc, timeSvc)
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
	initWorkers(ctx, workers, auditlogSvc, msgChannel, collector)

	log.C(ctx).Infof("Auditlog configured successfully, auth mode: %s", cfg.AuthMode)
	return auditlog.NewSink(msgChannel, cfg.MsgChannelTimeout, collector), auditlogSvc, nil
}

func fillJWTCredentials(cfg auditlog.OAuthConfig) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.OAuthURL + cfg.TokenPath,
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}
}

func fillOAuthMtlsCredentials(cfg auditlog.OAuthMtlsConfig) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: "",
		TokenURL:     cfg.OAuthURL + cfg.TokenPath,
		AuthStyle:    oauth2.AuthStyleInParams,
	}
}

func initWorkers(ctx context.Context, workers chan bool, auditlogSvc proxy.AuditlogService, msgChannel chan proxy.AuditlogMessage, collector *metrics.AuditlogCollector) {
	logger := log.C(ctx)

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Infoln("Worker starter goroutine finished")
				return
			case workers <- true:
			}
			worker := auditlog.NewWorker(auditlogSvc, msgChannel, collector)
			go func() {
				logger.Infoln("Starting worker for auditlog message processing")
				worker.Start(ctx)
				<-workers
			}()
		}
	}()
}
