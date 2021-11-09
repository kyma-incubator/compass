package main

import (
	"context"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	directorHandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/kyma-incubator/compass/components/ns-adapter/internal/adapter"
	"github.com/kyma-incubator/compass/components/ns-adapter/internal/handler"
	"github.com/kyma-incubator/compass/components/ns-adapter/internal/httputil"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"net/http"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	conf := adapter.Configuration{}
	err := envconfig.Init(&conf)
	exitOnError(err, "while reading Pairing Adapter configuration")

	h := handler.NewHandler()
	//TODO fix timeout error, What is the proper timeout value?

	handlerWithTimeout, err := directorHandler.WithTimeout(h, conf.ServerTimeout)
	exitOnError(err, "Failed configuring timeout on handler")

	router := mux.NewRouter()

	router.Use(correlation.AttachCorrelationIDToContext())
	router.NewRoute().
		Methods(http.MethodPut).
		Path("/api/v1/notifications").
		Handler(handlerWithTimeout)
	router.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	validation.ErrRequired = validation.ErrRequired.SetMessage("the value is required")

	server := &http.Server{
		Addr:              fmt.Sprintf(":%s", conf.Port),
		Handler:           router,
		ReadHeaderTimeout: conf.ServerTimeout,
	}
	ctx, err = log.Configure(ctx, conf.Log)
	exitOnError(err, "while configuring logger")

	log.C(ctx).Infof("API listening on %s", conf.Address)
	exitOnError(server.ListenAndServe(), "on starting HTTP server")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}
