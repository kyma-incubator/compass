package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-incubator/compass/components/gateway/internal/time"
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

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	router := mux.NewRouter()

	done := make(chan bool)
	var auditlogSvc AuditogService
	if cfg.AuditlogEnabled {
		auditlogSvc = initAuditLogsSvc(done)
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

func initAuditLogsSvc(done chan bool) AuditogService {
	log.Println("Auditlog enabled")
	auditlogCfg := auditlog.Config{}
	err := envconfig.InitWithPrefix(&auditlogCfg, "APP")
	exitOnError(err, "Error while loading auditlog cfg")

	httpClient, tenant := initAuditlogClient(auditlogCfg)
	uuidSvc := uuid.NewService()
	timeSvc := &time.TimeService{}

	auditlogClient, err := auditlog.NewClient(auditlogCfg, httpClient, uuidSvc, timeSvc, tenant)
	exitOnError(err, "Error while creating auditlog client from cfg")

	auditlogMsgChannel := make(chan auditlog.AuditlogMessage)

	logger := auditlog.NewService(auditlogClient)
	worker := auditlog.NewWorker(logger, auditlogMsgChannel, done)
	go func() {
		worker.Start()
	}()

	log.Printf("Auditlog configured successfully, auth mode:%s\n", auditlogCfg.AuthMode)
	return auditlog.NewSink(auditlogMsgChannel)
}

func initAuditlogClient(cfg auditlog.Config) (auditlog.HttpClient, *string) {
	if cfg.AuthMode == auditlog.Basic {
		var basicCfg auditlog.BasicAuthConfig
		err := envconfig.InitWithPrefix(&basicCfg, "APP")
		exitOnError(err, "while loading basic auth config from envs")

		return auditlog.NewBasicAuthClient(basicCfg), &basicCfg.Tenant
	} else if cfg.AuthMode == auditlog.OAuth {
		ctx := context.Background()

		var oauthCfg auditlog.OAuthConfig
		err := envconfig.InitWithPrefix(&oauthCfg, "APP")
		exitOnError(err, "while loading oauth config from envs")

		cfg := clientcredentials.Config{
			ClientID:     oauthCfg.ClientID,
			ClientSecret: oauthCfg.ClientSecret,
			TokenURL:     oauthCfg.OAuthURL,
			AuthStyle:    oauth2.AuthStyleAutoDetect,
		}

		return cfg.Client(ctx), nil
	} else {
		log.Fatal(fmt.Sprintf("Invalid Auditlog Auth mode: %s", cfg.AuthMode))
		return nil, nil
	}
}
