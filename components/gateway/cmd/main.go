package main

import (
	"log"
	"net/http"

	"github.com/kyma-incubator/compass/components/gateway/internal/time"
	"github.com/kyma-incubator/compass/components/gateway/internal/uuid"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog"

	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:3001"`

	DirectorOrigin  string `envconfig:"default=http://127.0.0.1:3000"`
	ConnectorOrigin string `envconfig:"default=http://127.0.0.1:3000"`
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
		auditlogSvc = &auditlog.DummyAuditlog{}
	}

	tr := proxy.NewTransport(auditlogSvc, http.DefaultTransport)

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
	auditlogCfg := auditlog.AuditlogConfig{}
	err := envconfig.InitWithPrefix(&auditlogCfg, "APP")
	exitOnError(err, "Error while loading auditlog cfg")

	uuidSvc := uuid.NewService()
	timeSvc := &time.TimeService{}

	auditlogClient, err := auditlog.NewClient(auditlogCfg, uuidSvc, timeSvc)
	exitOnError(err, "Error while creating auditlog client from cfg")

	auditlogMsgChannel := make(chan auditlog.AuditlogMessage)

	logger := auditlog.NewService(auditlogClient)
	worker := auditlog.NewWorker(logger, auditlogMsgChannel, done)
	go func() {
		worker.Start()
	}()

	return auditlog.NewAuditlogSink(auditlogMsgChannel)
}
