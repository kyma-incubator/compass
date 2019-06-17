package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/domain"

	"github.com/pkg/errors"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address               string `envconfig:"default=127.0.0.1:3000"`
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	configureLogger()

	gqlCfg := graphql.Config{
		Resolvers: domain.NewRootResolver(),
	}
	executableSchema := graphql.NewExecutableSchema(gqlCfg)

	log.Info("Registering endpoint on %s...", cfg.APIEndpoint)
	router := mux.NewRouter()
	router.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	router.HandleFunc(cfg.APIEndpoint, handler.GraphQL(executableSchema))

	http.Handle("/", router)

	log.Info("Listening on %s...", cfg.Address)
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

func configureLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}