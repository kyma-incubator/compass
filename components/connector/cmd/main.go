package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/connector/internal/oathkeeper"

	"github.com/kyma-incubator/compass/components/connector/pkg/gqlschema"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/connector/internal/namespacedname"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/secrets"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type config struct {
	Address               string `envconfig:"default=127.0.0.1:3000"`
	APIEndpoint           string `envconfig:"default=/graphql"`
	PlaygroundAPIEndpoint string `envconfig:"default=/graphql"`

	HydratorAddress string `envconfig:"default=127.0.0.1:8080"`

	CSRSubject struct {
		Country            string `envconfig:"default=PL"`
		Organization       string `envconfig:"default=Org"`
		OrganizationalUnit string `envconfig:"default=OrgUnit"`
		Locality           string `envconfig:"default=Locality"`
		Province           string `envconfig:"default=State"`
	}
	CertificateValidityTime     time.Duration `envconfig:"default=2160h"`
	CASecretName                string        `envconfig:"default=namespace/name"`
	RootCACertificateSecretName string        `envconfig:"optional"`

	Token struct {
		Length                int           `envconfig:"default=64"`
		RuntimeExpiration     time.Duration `envconfig:"default=60m"`
		ApplicationExpiration time.Duration `envconfig:"default=5m"`
		CSRExpiration         time.Duration `envconfig:"default=5m"`
	}

	DirectorURL string `envconfig:"default=127.0.0.1:3003"`
}

func (c *config) String() string {
	return fmt.Sprintf("Address: %s, APIEndpoint: %s, HydratorAddress: %s, "+
		"CSRSubjectCountry: %s, CSRSubjectOrganization: %s, CSRSubjectOrganizationalUnit: %s, "+
		"CSRSubjectLocality: %s, CSRSubjectProvince: %s, "+
		"CertificateValidityTime: %s, CASecretName: %s, RootCACertificateSecretName: %s, "+
		"TokenLength: %d, TokenRuntimeExpiration: %s, TokenApplicationExpiration: %s, TokenCSRExpiration: %s, "+
		"DirectorURL: %s",
		c.Address, c.APIEndpoint, c.HydratorAddress,
		c.CSRSubject.Country, c.CSRSubject.Organization, c.CSRSubject.OrganizationalUnit,
		c.CSRSubject.Locality, c.CSRSubject.Province,
		c.CertificateValidityTime, c.CASecretName, c.RootCACertificateSecretName,
		c.Token.Length, c.Token.RuntimeExpiration.String(), c.Token.ApplicationExpiration.String(), c.Token.CSRExpiration.String(),
		c.DirectorURL)
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	log.Println("Starting Connector Service")
	log.Printf("Config: %s", cfg.String())

	tokenCache := tokens.NewTokenCache(cfg.Token.ApplicationExpiration, cfg.Token.RuntimeExpiration, cfg.Token.CSRExpiration)
	tokenService := tokens.NewTokenService(tokenCache, tokens.NewTokenGenerator(cfg.Token.Length))

	authenticator := authentication.NewAuthenticator()

	tokenResolver := api.NewTokenResolver(tokenService)

	coreClientSet, appErr := newCoreClientSet()
	exitOnError(appErr, "Failed to initialize Kubernetes client.")
	secretsRepository := newSecretsRepository(coreClientSet)
	certificateUtility := certificates.NewCertificateUtility(cfg.CertificateValidityTime)
	certificateService := certificates.NewCertificateService(
		secretsRepository,
		certificateUtility,
		namespacedname.Parse(cfg.CASecretName),
		namespacedname.Parse(cfg.RootCACertificateSecretName),
	)
	csrSubjectConsts := certificates.CSRSubjectConsts{
		Country:            cfg.CSRSubject.Country,
		Organization:       cfg.CSRSubject.Organization,
		OrganizationalUnit: cfg.CSRSubject.OrganizationalUnit,
		Locality:           cfg.CSRSubject.Locality,
		Province:           cfg.CSRSubject.Province,
	}

	certificateResolver := api.NewCertificateResolver(
		authenticator,
		tokenService,
		certificateService,
		csrSubjectConsts,
		cfg.DirectorURL)

	server := prepareGraphQLServer(cfg, tokenResolver, certificateResolver)
	hydratorServer := prepareHydratorServer(cfg, tokenService, csrSubjectConsts)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		log.Printf("API listening on %s...", cfg.Address)
		if err := server.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	go func() {
		log.Printf("Hydrator API listening on %s...", cfg.HydratorAddress)
		if err := hydratorServer.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}

func prepareGraphQLServer(cfg config, tokenResolver api.TokenResolver, certResolver api.CertificateResolver) *http.Server {
	externalResolver := api.Resolver{CertificateResolver: certResolver, TokenResolver: tokenResolver}

	gqlInternalCfg := gqlschema.Config{
		Resolvers: &externalResolver,
	}

	externalExecutableSchema := gqlschema.NewExecutableSchema(gqlInternalCfg)

	externalRouter := mux.NewRouter()
	externalRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	externalRouter.HandleFunc(cfg.APIEndpoint, handler.GraphQL(externalExecutableSchema))

	authContextMiddleware := authentication.NewAuthenticationContextMiddleware()

	externalRouter.Use(authContextMiddleware.PropagateAuthentication)

	return &http.Server{
		Addr:    cfg.Address,
		Handler: externalRouter,
	}
}

func prepareHydratorServer(cfg config, tokenService tokens.Service, subjectConsts certificates.CSRSubjectConsts) *http.Server {
	certHeaderParser := oathkeeper.NewHeaderParser(subjectConsts)

	validationHydrator := oathkeeper.NewValidationHydrator(tokenService, certHeaderParser)

	router := mux.NewRouter()
	v1Router := router.PathPrefix("/v1").Subrouter()
	v1Router.HandleFunc("/tokens/resolve", validationHydrator.ResolveConnectorTokenHeader)

	return &http.Server{
		Addr:    cfg.HydratorAddress,
		Handler: router,
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func newCoreClientSet() (*kubernetes.Clientset, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		logrus.Warnf("Failed to read in cluster config: %s", err.Error())
		logrus.Info("Trying to initialize with local config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			return nil, errors.Errorf("failed to read k8s in-cluster configuration, %s", err.Error())
		}
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client, %s", err.Error())
	}

	return coreClientset, nil
}

func newSecretsRepository(coreClientSet *kubernetes.Clientset) secrets.Repository {
	core := coreClientSet.CoreV1()

	return secrets.NewRepository(func(namespace string) secrets.Manager {
		return core.Secrets(namespace)
	})
}
