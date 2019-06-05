package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/pkg/errors"

	"github.com/gorilla/mux"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:3001"`

	DirectorOrigin string `envconfig:"default=http://127.0.0.1:3000"`
}

func main() {

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	directorUrl, err := url.Parse(cfg.DirectorOrigin)
	exitOnError(err, "Error while parsing Director URL")

	log.Printf("Proxying requests to Director: %s\n", cfg.DirectorOrigin)
	directorProxy := httputil.NewSingleHostReverseProxy(directorUrl)

	router := mux.NewRouter()
	router.HandleFunc("/", directorProxy.ServeHTTP) // GraphQL Playground

	// TODO: To proxy Subscriptions, recompile against Go 1.12: https://github.com/golang/go/issues/26937
	router.HandleFunc("/graphql", directorProxy.ServeHTTP) // GraphQL API Endpoint

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

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
