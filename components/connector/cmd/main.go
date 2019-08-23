package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/secrets"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/internalschema"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type config struct {
	ExternalAddress       string `envconfig:"default=127.0.0.1:3000"`
	InternalAddress       string `envconfig:"default=127.0.0.1:3001"`
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`

	CSRSubject struct {
		Country            string `envconfig:"default=C"`
		Organization       string `envconfig:"default=O"`
		OrganizationalUnit string `envconfig:"default=OU"`
		Locality           string `envconfig:"default=L"`
		Province           string `envconfig:"default=ST"`
	}
	CertificateValidityTime     time.Duration        `envconfig:"default=90d"`
	CASecretName                types.NamespacedName `envconfig:"default=namespace/casecretname"`
	RootCACertificateSecretName types.NamespacedName `envconfig:"default="`

	Token struct {
		Length                int           `envconfig:"default=64"`
		RuntimeExpiration     time.Duration `envconfig:"default=60m"`
		ApplicationExpiration time.Duration `envconfig:"default=5m"`
	}
}

func (c *config) String() string {
	return fmt.Sprintf("ExternalAddress: %s, InternalAddress: %s, APIEndpoint: %s, "+
		"CSRSubjectCountry: %s, CSRSubjectOrganization: %s, CSRSubjectOrganizationalUnit: %s"+
		"CSRSubjectLocality: %s, CSRSubjectProvince: %s"+
		"CertificateValidityTime: %s, CASecretName: %s, RootCACertificateSecretName: %s"+
		"TokenLength: %d, TokenRuntimeExpiration: %s, TokenApplicationExpiration: %s",
		c.ExternalAddress, c.InternalAddress, c.APIEndpoint,
		c.CSRSubject.Country, c.CSRSubject.Organization, c.CSRSubject.OrganizationalUnit,
		c.CSRSubject.Locality, c.CSRSubject.Province,
		c.CertificateValidityTime, c.CASecretName, c.RootCACertificateSecretName,
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

	coreClientSet, appErr := newCoreClientSet()
	exitOnError(appErr, "Failed to initialize Kubernetes client.")
	secretsRepository := newSecretsRepository(coreClientSet)
	certificateUtility := certificates.NewCertificateUtility(cfg.CertificateValidityTime)

	certificateService := certificates.NewCertificateService(secretsRepository, certificateUtility, cfg.CASecretName, cfg.RootCACertificateSecretName)

	certificateResolver := api.NewCertificateResolver(authenticator, tokenService, certificateService)

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

	certHeaderParser := authentication.NewHeaderParser(
		cfg.CSRSubject.Country,
		cfg.CSRSubject.Locality,
		cfg.CSRSubject.Province,
		cfg.CSRSubject.Organization,
		cfg.CSRSubject.OrganizationalUnit,
	)
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

func newCoreClientSet() (*kubernetes.Clientset, apperrors.AppError) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, apperrors.Internal("failed to read k8s in-cluster configuration, %s", err)
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, apperrors.Internal("failed to create k8s core client, %s", err)
	}

	return coreClientset, nil
}

func newSecretsRepository(coreClientSet *kubernetes.Clientset) secrets.Repository {
	core := coreClientSet.CoreV1()

	return secrets.NewRepository(func(namespace string) secrets.Manager {
		return core.Secrets(namespace)
	})
}
