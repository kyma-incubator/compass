package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	timeservices "github.com/kyma-incubator/compass/components/gateway/internal/time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-incubator/compass/components/gateway/internal/uuid"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog"

	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:3000"`

	DirectorOrigin  string `envconfig:"default=http://127.0.0.1:3001"`
	ConnectorOrigin string `envconfig:"default=http://127.0.0.1:3002"`
	AuditlogEnabled bool   `envconfig:"default=false"`
}

type AuditogService interface {
	Log(request, response string, claims proxy.Claims) error
}

type HTTPTransport interface {
	RoundTrip(req *http.Request) (resp *http.Response, err error)
}

const (
	auditlogMsgChannelSize = 100
	auditlogTimeout        = time.Second * 5
)

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	router := mux.NewRouter()

	done := make(chan bool)
	var auditlogSvc AuditogService
	if cfg.AuditlogEnabled {
		auditlogSvc, err = initAuditLogs(done)
		exitOnError(err, "Error while initializing auditlog service")
	} else {
		auditlogSvc = &auditlog.NoOpService{}
	}

	defaultTr := http.Transport{}
	err = proxyRequestsForComponent(router, "/connector", cfg.ConnectorOrigin, &defaultTr)
	exitOnError(err, "Error while initializing proxy for Connector")

	tr := proxy.NewTransport(auditlogSvc, http.DefaultTransport)
	err = proxyRequestsForComponent(router, "/director", cfg.DirectorOrigin, tr)
	exitOnError(err, "Error while initializing proxy for Director")

	router.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			log.Println(errors.Wrapf(err, "while writing to response body").Error())
		}
	})

	http.Handle("/", router)

	log.Printf("Listening on %s", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, nil); err != nil {
		done <- true
		panic(err)
	}
}

func proxyRequestsForComponent(router *mux.Router, path string, targetOrigin string, transport HTTPTransport, middleware ...mux.MiddlewareFunc) error {
	log.Printf("Proxying requests on path `%s` to `%s`\n", path, targetOrigin)

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
	log.Println("Auditlog enabled")
	cfg := auditlog.Config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		return nil, errors.Wrap(err, "Error while loading auditlog cfg")
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
			httpClient = auditlog.NewBasicAuthClient(basicCfg)
		}
	case auditlog.OAuth:
		{
			var oauthCfg auditlog.OAuthConfig
			err := envconfig.InitWithPrefix(&oauthCfg, "APP")
			if err != nil {
				return nil, errors.Wrap(err, "while loading auditlog OAuth configuration")
			}

			cfg := fillJWTCredentials(oauthCfg)
			httpClient = cfg.Client(context.Background())
			msgFactory = auditlog.NewMessageFactory(oauthCfg.UserVar, oauthCfg.TenantVar, uuidSvc, timeSvc)
		}
	default:
		return nil, errors.New(fmt.Sprintf("Invalid Auditlog Auth mode: %s", cfg.AuthMode))
	}

	auditlogClient, err := auditlog.NewClient(cfg, httpClient)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating auditlog client from cfg")
	}

	auditlogSvc := auditlog.NewService(auditlogClient, msgFactory)
	msgChannel := make(chan auditlog.Message, cfg.AutitlogMsgChannelSize)
	initWorker(auditlogSvc, done, msgChannel)

	log.Printf("Auditlog configured successfully, auth mode:%s\n", cfg.AuthMode)
	return auditlog.NewSink(msgChannel, cfg.AuditlogMsgChannelTimeout), nil
}

func fillJWTCredentials(cfg auditlog.OAuthConfig) clientcredentials.Config {
	return clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.OAuthURL,
		AuthStyle:    oauth2.AuthStyleAutoDetect,
	}
}

func initWorker(auditlogSvc auditlog.AuditlogService, done chan bool, msgChannel chan auditlog.Message) {
	worker := auditlog.NewWorker(auditlogSvc, msgChannel, done)
	go func() {
		worker.Start()
	}()
}
