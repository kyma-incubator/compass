package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/kyma-incubator/compass/components/pairing-adapter/internal/adapter"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	conf := adapter.Configuration{}
	err := envconfig.Init(&conf)
	exitOnError(err, "while reading Pairing Adapter Configuration")

	cc := clientcredentials.Config{
		TokenURL:     conf.OAuth.URL,
		ClientID:     conf.OAuth.ClientID,
		ClientSecret: conf.OAuth.ClientSecret,
	}

	client := cc.Client(context.Background())

	cli := adapter.NewClient(client, conf.Mapping)

	h := adapter.NewHandler(cli)

	http.Handle("/adapter", h)
	http.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	err = http.ListenAndServe(fmt.Sprintf(":%s", conf.Port), nil)
	exitOnError(err, "on starting HTTP server")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
