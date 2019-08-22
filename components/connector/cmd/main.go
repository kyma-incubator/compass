package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"

	"github.com/pkg/errors"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"
)

type config struct {
	Address               string `envconfig:"default=127.0.0.1:3000"`
	InternalAddress       string `envconfig:"default=127.0.0.1:3001"` // TODO: figure out how to split schema to two different APIs
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`

	Token struct {
		Length                int
		RuntimeExpiration     time.Duration
		ApplicationExpiration time.Duration
	}
}

func (c *config) String() string {
	return fmt.Sprintf("Address: %s, InternalAddress: %s, APIEndpoint: %s, "+
		"TokenLength: %d, TokenRuntimeExpiration: %s, TokenApplicationExpiration: %s",
		c.Address, c.InternalAddress, c.APIEndpoint,
		c.Token.Length, c.Token.RuntimeExpiration.String(), c.Token.ApplicationExpiration.String())
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	log.Println("Starting Connector Service")
	log.Printf("Config: %s", cfg.String())

	// TODO: Get values from config
	certHeaderParser := authentication.NewHeaderParser("", "", "", "", "")
	authContextMiddleware := authentication.NewAuthenticationContextMiddleware(certHeaderParser)

	tokenCache := tokens.NewTokenCache(cfg.Token.ApplicationExpiration, cfg.Token.RuntimeExpiration)
	tokenService := tokens.NewTokenService(tokenCache, tokens.NewTokenGenerator(cfg.Token.Length))

	authenticator := authentication.NewAuthenticator(tokenService)

	tokenResolver := api.NewTokenResolver(tokenService)
	certificateResolver := api.NewCertificateResolver(authenticator, tokenService)
	resolver := api.Resolver{TokenResolver: tokenResolver, CertificateResolver: certificateResolver}

	gqlCfg := gqlschema.Config{
		Resolvers: &resolver,
	}
	executableSchema := gqlschema.NewExecutableSchema(gqlCfg)

	log.Printf("Registering endpoint on %s...", cfg.APIEndpoint)
	router := mux.NewRouter()
	router.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	router.HandleFunc(cfg.APIEndpoint, handler.GraphQL(executableSchema))

	router.Use(authContextMiddleware.PropagateAuthentication)

	http.Handle("/", router)

	log.Printf("Listening on %s...", cfg.Address)
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
