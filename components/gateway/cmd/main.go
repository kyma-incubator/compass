package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog"
	timeservices "github.com/kyma-incubator/compass/components/gateway/internal/time"
	"github.com/kyma-incubator/compass/components/gateway/internal/uuid"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:3000"`

	ServerTimeout time.Duration `envconfig:"default=114s"`

	DirectorOrigin  string `envconfig:"default=http://127.0.0.1:3001"`
	ConnectorOrigin string `envconfig:"default=http://127.0.0.1:3002"`
	AuditlogEnabled bool   `envconfig:"default=false"`
}

type AuditogService interface {
	Log(ctx context.Context, request, response string, claims proxy.Claims) error
}

type HTTPTransport interface {
	RoundTrip(req *http.Request) (resp *http.Response, err error)
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	router := mux.NewRouter()
	router.Use(correlation.AttachCorrelationIDToContext())

	done := make(chan bool)
	var auditlogSvc AuditogService
	if cfg.AuditlogEnabled {
		log.Println("Auditlog is enabled")
		auditlogSvc, err = initAuditLogs(done)
		exitOnError(err, "Error while initializing auditlog service")
	} else {
		log.Println("Auditlog is disabled")
		auditlogSvc = &auditlog.NoOpService{}
	}

	correlationTr := httputil.NewCorrelationIDTransport(http.DefaultTransport)
	tr := proxy.NewTransport(auditlogSvc, correlationTr)

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

	log.Printf("Listening on %s", cfg.Address)
	if err := server.ListenAndServe(); err != nil {
		done <- true
		panic(err)
	}
}

func proxyRequestsForComponent(router *mux.Router, path string, targetOrigin string, transport HTTPTransport, middleware ...mux.MiddlewareFunc) error {
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

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func initAuditLogs(done chan bool) (AuditogService, error) {
	cfg := auditlog.Config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		return nil, errors.Wrap(err, "while loading auditlog cfg")
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
				return nil, errors.Wrap(err, "while loading auditlog basic auth configuration")
			}

			msgFactory = auditlog.NewMessageFactory("proxy", basicCfg.Tenant, uuidSvc, timeSvc)
			tr := httputil.NewCorrelationIDTransport(http.DefaultTransport)
			baseHttpClient := http.Client{
				Timeout:   cfg.ClientTimeout,
				Transport: tr,
			}

			httpClient = auditlog.NewBasicAuthClient(basicCfg, baseHttpClient)
		}
	case auditlog.OAuth:
		{
			var oauthCfg auditlog.OAuthConfig
			err := envconfig.InitWithPrefix(&oauthCfg, "APP")
			if err != nil {
				return nil, errors.Wrap(err, "while loading auditlog OAuth configuration")
			}

			ccCfg := fillJWTCredentials(oauthCfg)
			client := ccCfg.Client(context.Background())

			client.Transport = httputil.NewCorrelationIDTransport(client.Transport)
			client.Timeout = cfg.ClientTimeout
			httpClient = client

			msgFactory = auditlog.NewMessageFactory(oauthCfg.User, oauthCfg.Tenant, uuidSvc, timeSvc)
		}
	default:
		return nil, fmt.Errorf("invalid auditlog auth mode: %s", cfg.AuthMode)
	}

	auditlogClient, err := auditlog.NewClient(cfg, httpClient)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating auditlog client from cfg")
	}

	auditlogSvc := auditlog.NewService(auditlogClient, msgFactory)
	msgChannel := make(chan auditlog.Message, cfg.MsgChannelSize)
	workers := make(chan bool, cfg.WriteWorkers)
	initWorkers(workers, auditlogSvc, done, msgChannel)

	log.Printf("Auditlog configured successfully, auth mode:%s", cfg.AuthMode)
	return auditlog.NewSink(msgChannel, cfg.MsgChannelTimeout), nil
}

func fillJWTCredentials(cfg auditlog.OAuthConfig) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.OAuthURL,
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}
}

func initWorkers(workers chan bool, auditlogSvc auditlog.AuditlogService, done chan bool, msgChannel chan auditlog.Message) {
	go func() {
		for {
			select {
			case <-done:
				log.Println("Worker starter goroutine finished")
				return
			case workers <- true:
			}
			worker := auditlog.NewWorker(auditlogSvc, msgChannel, done)
			go func() {
				log.Println("Starting worker for auditlog message processing")
				worker.Start()
				<-workers
			}()
		}
	}()
}
