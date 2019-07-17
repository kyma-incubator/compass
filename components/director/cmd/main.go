package main

import (
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/domain"

	"github.com/pkg/errors"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/vrischmann/envconfig"
)

const connStringf string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

type config struct {
	Address  string `envconfig:"default=127.0.0.1:3000"`
	Database struct {
		User     string `envconfig:"default=postgres,APP_DB_USER"`
		Password string `envconfig:"default=pgsql@12345,APP_DB_PASSWORD"`
		Host     string `envconfig:"default=localhost,APP_DB_HOST"`
		Port     string `envconfig:"default=5432,APP_DB_PORT"`
		Name     string `envconfig:"default=postgres,APP_DB_NAME"`
		SSLMode  string `envconfig:"default=disable,APP_DB_SSL"`
	}
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	configureLogger()

	connString := fmt.Sprintf(connStringf, cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)
	transact, closeFunc, err := persistence.Configure(connString)
	if err != nil {
		panic(err)
	}

	defer func() {
		err := closeFunc()
		if err != nil {
			panic(err)
		}
	}()

	gqlCfg := graphql.Config{
		Resolvers: domain.NewRootResolver(transact),
	}
	executableSchema := graphql.NewExecutableSchema(gqlCfg)

	log.Infof("Registering endpoint on %s...", cfg.APIEndpoint)
	router := mux.NewRouter()

	router.Use(tenant.RequireAndPassContext)
	router.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	router.HandleFunc(cfg.APIEndpoint, handler.GraphQL(executableSchema))

	http.Handle("/", router)

	log.Infof("Listening on %s...", cfg.Address)
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
