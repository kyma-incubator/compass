package main

import (
	"github.com/pkg/errors"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:3000"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	gqlCfg := gqlschema.Config{
		Resolvers: &gqlschema.Resolver{},
	}
	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)

	router := mux.NewRouter()
	router.HandleFunc("/", handler.Playground("Dataloader", "/graphql"))
	router.HandleFunc("/graphql", handler.GraphQL(executableSchema))

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
