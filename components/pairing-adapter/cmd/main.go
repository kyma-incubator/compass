package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/pairing-adapter/internal/adapter"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

func main() {
	conf := adapter.Configuration{}
	err := envconfig.Init(&conf)
	exitOnError(err, "while reading Pairing Adapter configuration")

	ctx, err := log.Configure(context.Background(), conf.Log)
	exitOnError(err, "while configuring logger")

	authStyle, err := getAuthStyle(conf.OAuth.AuthStyle)
	exitOnError(err, "while getting Auth Style")

	cc := clientcredentials.Config{
		TokenURL:     conf.OAuth.URL,
		ClientID:     conf.OAuth.ClientID,
		ClientSecret: conf.OAuth.ClientSecret,
		AuthStyle:    authStyle,
	}

	baseClient := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
	}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, baseClient)

	client := cc.Client(ctx)
	client.Timeout = conf.ClientTimeout

	cli := adapter.NewClient(client, conf.Mapping)

	h := adapter.NewHandler(cli)
	handlerWithTimeout, err := handler.WithTimeout(h, conf.ServerTimeout)
	exitOnError(err, "Failed configuring timeout on handler")

	router := mux.NewRouter()

	router.Use(correlation.AttachCorrelationIDToContext())
	router.Handle("/adapter", handlerWithTimeout)
	router.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", conf.Port),
		Handler:           router,
		ReadHeaderTimeout: conf.ServerTimeout,
	}

	exitOnError(server.ListenAndServe(), "on starting HTTP server")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
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
