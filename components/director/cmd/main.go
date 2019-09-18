package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"

	graphql2 "github.com/99designs/gqlgen/graphql"
	"github.com/davecgh/go-spew/spew"

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
	TenantMappingEndpoint string `envconfig:"default=/tenant-mapping"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	configureLogger()

	connString := fmt.Sprintf(connStringf, cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)
	transact, closeFunc, err := persistence.Configure(log.StandardLogger(), connString)
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
	gqlCfg.Directives.HasScopes = func(ctx context.Context, obj interface{}, next graphql2.Resolver, scopesDefinition string) (res interface{}, err error) {
		spew.Dump("ctx", ctx)
		spew.Dump("obj", obj)
		spew.Dump("scopesDefintion", scopesDefinition)
		return next(ctx)
	}
	executableSchema := graphql.NewExecutableSchema(gqlCfg)

	router := mux.NewRouter()

	log.Infof("Registering GraphQL endpoint on %s...", cfg.APIEndpoint)
	router.Use(tenant.RequireAndPassContext)
	router.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	router.HandleFunc(cfg.APIEndpoint, handler.GraphQL(executableSchema))

	http.Handle("/", router)

	log.Infof("Registering Tenant Mapping endpoint on %s...", cfg.TenantMappingEndpoint)
	tenantMappingHandler := tenantmapping.NewHandler()
	http.Handle(cfg.TenantMappingEndpoint, tenantMappingHandler)

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
