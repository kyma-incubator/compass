package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"

	"github.com/kyma-incubator/compass/components/pairing-adapter/internal/adapter"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	conf := adapter.Configuration{}
	err := envconfig.Init(&conf)
	exitOnError(err, "while reading Pairing Adapter Configuration")

	authStyle, err := getAuthStyle(conf.OAuth.AuthStyle)
	exitOnError(err, "while getting Auth Style")

	cc := clientcredentials.Config{
		TokenURL:     conf.OAuth.URL,
		ClientID:     conf.OAuth.ClientID,
		ClientSecret: conf.OAuth.ClientSecret,
		AuthStyle:    authStyle,
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

func getAuthStyle(style adapter.AuthStyle) (oauth2.AuthStyle, error) {
	switch style {
	case adapter.AuthStyleInParams:
		return oauth2.AuthStyleInParams, nil
	case adapter.AuthStyleInHeader:
		return oauth2.AuthStyleInHeader, nil
	case adapter.AuthStyleAutoDetect:
		return oauth2.AuthStyleAutoDetect, nil
	default:
		return -1, errors.New("unknown Auth style")
	}
}
