package main

import (
	"fmt"
	"os"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
)

const (
	examplesDirectory = "../../examples"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to load config", err.Error())
		os.Exit(2)
	}

	err = api.Generate(cfg) //api.AddPlugin(scopesdecorator.NewPlugin("schema.graphql")),
	//api.AddPlugin(descriptionsdecorator.NewPlugin("schema.graphql", examplesDirectory))

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(3)
	}
}
