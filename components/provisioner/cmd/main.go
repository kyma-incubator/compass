package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/provisioner/internal/api"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address               string `envconfig:"default=127.0.0.1:3050"`
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`
}

func (c *config) String() string {
	return fmt.Sprintf("Address: %s, APIEndpoint: %s", c.Address, c.APIEndpoint)
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	log.Println("Starting Provisioner")
	log.Printf("Config: %s", cfg.String())

	repository := make(map[string]api.RuntimeOperation)

	resolver := api.NewMockResolver(repository)

	gqlCfg := gqlschema.Config{
		Resolvers: resolver,
	}
	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)

	log.Printf("Registering endpoint on %s...", cfg.APIEndpoint)

	router := mux.NewRouter()
	router.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	router.HandleFunc(cfg.APIEndpoint, handler.GraphQL(executableSchema))

	http.Handle("/", router)

	log.Printf("API listening on %s...", cfg.Address)
	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		panic(err)
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
