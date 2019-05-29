package main

import (
	"github.com/kyma-incubator/compass/components/gateway/internal/director"
	"net/http"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/gateway/internal/gqlschema"
)

func main() {
	directorCli, err := director.NewClient()
	if err != nil {
		panic(err)
	}

	cfg := gqlschema.Config{
		Resolvers: gqlschema.NewResolver(directorCli),
	}
	executableSchema := gqlschema.NewExecutableSchema(cfg)

	router := mux.NewRouter()
	router.HandleFunc("/", handler.Playground("DataLoader", "/graphql"))
	router.HandleFunc("/graphql", handler.GraphQL(executableSchema))

	http.Handle("/", router)

	if err := http.ListenAndServe("127.0.0.1:3000", nil); err != nil {
		panic(err)
	}
}
