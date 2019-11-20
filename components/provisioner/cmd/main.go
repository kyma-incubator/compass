package main

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const connStringFormat string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

type config struct {
	Address               string `envconfig:"default=127.0.0.1:3050"`
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`
	CredentialsNamespace  string `envconfig:"default=compass-system"`

	Database struct {
		User     string `envconfig:"default=postgres"`
		Password string `envconfig:"default=password"`
		Host     string `envconfig:"default=localhost"`
		Port     string `envconfig:"default=5432"`
		Name     string `envconfig:"default=provisioner"`
		SSLMode  string `envconfig:"default=disable"`

		SchemaFilePath string `envconfig:"default=assets/database/provisioner.sql"`
	}
}

func (c *config) String() string {
	return fmt.Sprintf("Address: %s, APIEndpoint: %s, CredentialsNamespace: %s "+
		"DatabaseUser: %s, DatabaseHost: %s, DatabasePort: %s, "+
		"DatabaseName: %s, DatabaseSSLMode: %s, DatabaseSchemaFilePath: %s",
		c.Address, c.APIEndpoint, c.CredentialsNamespace,
		c.Database.User, c.Database.Host, c.Database.Port,
		c.Database.Name, c.Database.SSLMode, c.Database.SchemaFilePath)
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Failed to load application config")

	log.Println("Starting Provisioner")
	log.Printf("Config: %s", cfg.String())

	connString := fmt.Sprintf(connStringFormat, cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)

	resolver, err := newResolver(connString, cfg.Database.SchemaFilePath, cfg.CredentialsNamespace)
	exitOnError(err, "Failed to initialize GraphQL resolver ")

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
