package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/internalschema"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"

	"github.com/pkg/errors"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/connector/internal/api"
)

type config struct {
	ExternalAddress       string `envconfig:"default=127.0.0.1:3000"`
	InternalAddress       string `envconfig:"default=127.0.0.1:3001"`
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`

	Token struct {
		Length                int
		RuntimeExpiration     time.Duration
		ApplicationExpiration time.Duration
	}
}

func (c *config) String() string {
	return fmt.Sprintf("ExternalAddress: %s, InternalAddress: %s, APIEndpoint: %s, "+
		"TokenLength: %d, TokenRuntimeExpiration: %s, TokenApplicationExpiration: %s",
		c.ExternalAddress, c.InternalAddress, c.APIEndpoint,
		c.Token.Length, c.Token.RuntimeExpiration.String(), c.Token.ApplicationExpiration.String())
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	log.Println("Starting Connector Service")
	log.Printf("Config: %s", cfg.String())

	tokenCache := tokens.NewTokenCache(cfg.Token.ApplicationExpiration, cfg.Token.RuntimeExpiration)
	tokenService := tokens.NewTokenService(tokenCache, tokens.NewTokenGenerator(cfg.Token.Length))

	authenticator := authentication.NewAuthenticator(tokenService)

	tokenResolver := api.NewTokenResolver(tokenService)
	certificateResolver := api.NewCertificateResolver(authenticator, tokenService)

	internalServer := prepareInternalServer(cfg, tokenResolver)
	externalServer := prepareExternalServer(cfg, certificateResolver)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		log.Printf("Internal API listening on %s...", cfg.InternalAddress)
		if err := internalServer.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	go func() {
		log.Printf("Extranal API listening on %s...", cfg.ExternalAddress)
		if err := externalServer.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}

func prepareInternalServer(cfg config, tokenResolver api.TokenResolver) *http.Server {
	internalResolver := api.InternalResolver{TokenResolver: tokenResolver}

	gqlInternalCfg := internalschema.Config{
		Resolvers: &internalResolver,
	}

	internalExecutableSchema := internalschema.NewExecutableSchema(gqlInternalCfg)

	internalRouter := mux.NewRouter()
	internalRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	internalRouter.HandleFunc(cfg.APIEndpoint, handler.GraphQL(internalExecutableSchema))

	return &http.Server{
		Addr:    cfg.InternalAddress,
		Handler: internalRouter,
	}
}

func prepareExternalServer(cfg config, certResolver api.CertificateResolver) *http.Server {
	externalResolver := api.ExternalResolver{CertificateResolver: certResolver}

	gqlInternalCfg := externalschema.Config{
		Resolvers: &externalResolver,
	}

	externalExecutableSchema := externalschema.NewExecutableSchema(gqlInternalCfg)

	externalRouter := mux.NewRouter()
	externalRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	externalRouter.HandleFunc(cfg.APIEndpoint, handler.GraphQL(externalExecutableSchema))

	// TODO: Get values from config
	certHeaderParser := authentication.NewHeaderParser("", "", "", "", "")
	authContextMiddleware := authentication.NewAuthenticationContextMiddleware(certHeaderParser)

	externalRouter.Use(authContextMiddleware.PropagateAuthentication)

	return &http.Server{
		Addr:    cfg.ExternalAddress,
		Handler: externalRouter,
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
