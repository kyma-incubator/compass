package main

import (
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connector/internal/gqlschema"
)

func main() {
	cfg := gqlschema.Config{}
	executableSchema := gqlschema.NewExecutableSchema(cfg)

	router := mux.NewRouter()
	router.HandleFunc("/", handler.Playground("Dataloader", "/graphql"))
	router.HandleFunc("/graphql", handler.GraphQL(executableSchema))

	http.Handle("/", router)

	if err := http.ListenAndServe(":3000", nil); err != nil {
		panic(err)
	}
}
