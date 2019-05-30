package main

import (
	"log"
	"net/http"
	"os"

	"github.com/kyma-incubator/compass/components/gateway/internal/director"

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

	addr := os.Getenv("GATEWAY_ADDRESS")
	log.Printf("Gateway server addr %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}
