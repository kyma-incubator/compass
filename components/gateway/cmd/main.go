package main

import (
	"log"
	"net/http"

	"github.com/kyma-incubator/compass/components/gateway/internal/tenant"
	"github.com/kyma-incubator/compass/components/gateway/pkg/proxy"
	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:3001"`

	DirectorOrigin  string `envconfig:"default=http://127.0.0.1:3000"`
	ConnectorOrigin string `envconfig:"default=http://127.0.0.1:3000"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	router := mux.NewRouter()

	err = proxyRequestsForComponent(router, "/connector", cfg.ConnectorOrigin)
	exitOnError(err, "Error while initializing proxy for Connector")

	err = proxyRequestsForComponent(router, "/director", cfg.DirectorOrigin, tenant.RequireTenantHeader("GET"))
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
		panic(err)
	}
}

func proxyRequestsForComponent(router *mux.Router, path string, targetOrigin string, middleware ...mux.MiddlewareFunc) error {
	log.Printf("Proxying requests on path `%s` to `%s`\n", path, targetOrigin)

	componentProxy, err := proxy.New(targetOrigin, path)
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
