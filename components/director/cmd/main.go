package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"

	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/executor"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/signal"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

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
	APIEndpoint                   string `envconfig:"default=/graphql"`
	TenantMappingEndpoint         string `envconfig:"default=/tenant-mapping"`
	PlaygroundAPIEndpoint         string `envconfig:"default=/graphql"`
	ScopesConfigurationFile       string
	ScopesConfigurationFileReload time.Duration `envconfig:"default=1m"`

	JWKSEndpoint        string        `envconfig:"default=file://hack/default-jwks.json"`
	JWKSSyncPeriod      time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone bool          `envconfig:"default=true"`

	StaticUsersSrc string `envconfig:"default=/data/static-users.yaml"`

	OneTimeToken onetimetoken.Config
	OAuth20      oauth20.Config
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	configureLogger()

	connString := fmt.Sprintf(connStringf, cfg.Database.Host, cfg.Database.Port, cfg.Database.User,
		cfg.Database.Password, cfg.Database.Name, cfg.Database.SSLMode)
	transact, closeFunc, err := persistence.Configure(log.StandardLogger(), connString)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	stopCh := signal.SetupChannel()
	scopeCfgProvider := createAndRunScopeConfigProvider(stopCh, cfg)

	gqlCfg := graphql.Config{
		Resolvers: domain.NewRootResolver(transact, scopeCfgProvider, cfg.OneTimeToken, cfg.OAuth20),
		Directives: graphql.DirectiveRoot{
			HasScopes: scope.NewDirective(scopeCfgProvider).VerifyScopes,
		},
	}

	executableSchema := graphql.NewExecutableSchema(gqlCfg)

	mainRouter := mux.NewRouter()

	log.Infof("Registering GraphQL endpoint on %s...", cfg.APIEndpoint)
	authMiddleware := authenticator.New(cfg.JWKSEndpoint, cfg.AllowJWTSigningNone)

	if cfg.JWKSSyncPeriod != 0 {
		log.Infof("JWKS synchronization enabled. Sync period: %v", cfg.JWKSSyncPeriod)
		periodicExecutor := executor.NewPeriodic(cfg.JWKSSyncPeriod, func(stopCh <-chan struct{}) {
			err := authMiddleware.SynchronizeJWKS()
			if err != nil {
				log.Error(errors.Wrap(err, "while synchronizing JWKS"))
			}
		})
		go periodicExecutor.Run(stopCh)
	}

	mainRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))

	gqlAPIRouter := mainRouter.PathPrefix(cfg.APIEndpoint).Subrouter()
	gqlAPIRouter.Use(authMiddleware.Handler())
	gqlAPIRouter.HandleFunc("", handler.GraphQL(executableSchema))

	log.Infof("Registering Tenant Mapping endpoint on %s...", cfg.TenantMappingEndpoint)
	tenantMappingHandlerFunc, err := getTenantMappingHanderFunc(transact, cfg.StaticUsersSrc, scopeCfgProvider)
	exitOnError(err, "Error while configuring tenant mapping handler")

	mainRouter.HandleFunc(cfg.TenantMappingEndpoint, tenantMappingHandlerFunc)

	mainRouter.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
		_, err := writer.Write([]byte("ok"))
		if err != nil {
			log.Errorf(errors.Wrapf(err, "while writing to response body").Error())
		}
	})

	examplesServer := http.FileServer(http.Dir("./examples/"))
	mainRouter.PathPrefix("/examples/").Handler(http.StripPrefix("/examples/", examplesServer))

	srv := &http.Server{Addr: cfg.Address, Handler: mainRouter}
	log.Infof("Listening on %s...", cfg.Address)
	go func() {
		<-stopCh
		// Interrupt signal received - shut down the server
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Errorf("HTTP server Shutdown: %v", err)
		}
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("HTTP server ListenAndServe: %v", err)
	}
}

func createAndRunScopeConfigProvider(stopCh <-chan struct{}, cfg config) *scope.Provider {
	provider := scope.NewProvider(cfg.ScopesConfigurationFile)
	err := provider.Load()
	exitOnError(err, "Error on loading scopes config file")
	executor.NewPeriodic(cfg.ScopesConfigurationFileReload, func(stopCh <-chan struct{}) {
		if err := provider.Load(); err != nil {
			exitOnError(err, "Error from Reloader watch")
		}
		log.Infof("Successfully reloaded scopes configuration")

	}).Run(stopCh)

	return provider
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

func getTenantMappingHanderFunc(transact persistence.Transactioner, staticUsersSrc string, scopeProvider *scope.Provider) (func(writer http.ResponseWriter, request *http.Request), error) {
	uidSvc := uid.NewService()
	authConverter := auth.NewConverter()
	systemAuthConverter := systemauth.NewConverter(authConverter)
	systemAuthRepo := systemauth.NewRepository(systemAuthConverter)
	systemAuthSvc := systemauth.NewService(systemAuthRepo, uidSvc)
	staticUsersRepo, err := tenantmapping.NewStaticUserRepository(staticUsersSrc)
	if err != nil {
		return nil, errors.Wrap(err, "while creating StaticUser repository instance")
	}

	mapperForUser := tenantmapping.NewMapperForUser(staticUsersRepo)
	mapperForSystemAuth := tenantmapping.NewMapperForSystemAuth(systemAuthSvc, scopeProvider)

	reqDataParser := tenantmapping.NewReqDataParser()

	return tenantmapping.NewHandler(reqDataParser, transact, mapperForUser, mapperForSystemAuth).ServeHTTP, nil
}
