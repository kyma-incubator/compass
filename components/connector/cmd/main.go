package main

import (
	"context"
	"log"
	"path/filepath"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/kyma-incubator/compass/components/connector/config"

	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	cfg := config.Config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app Config")

	log.Println("Starting Connector Service")
	log.Printf("Config: %s", cfg.String())

	k8sClientSet, appErr := newCoreClientSet(cfg.KubernetesClient.PollInteval, cfg.KubernetesClient.PollTimeout, cfg.KubernetesClient.Timeout)
	exitOnError(appErr, "Failed to initialize Kubernetes client.")

	internalComponents, certificateLoader, revocationListLoader := config.InitInternalComponents(cfg, k8sClientSet)
	go certificateLoader.Run()
	go revocationListLoader.Run(context.Background())

	tokenResolver := api.NewTokenResolver(internalComponents.TokenService)
	certificateResolver := api.NewCertificateResolver(
		internalComponents.Authenticator,
		internalComponents.TokenService,
		internalComponents.CertificateService,
		internalComponents.SubjectConsts,
		cfg.DirectorURL,
		cfg.CertificateSecuredConnectorURL,
		internalComponents.RevocationsRepository)

	authContextMiddleware := authentication.NewAuthenticationContextMiddleware()

	externalGqlServer, err := config.PrepareExternalGraphQLServer(cfg, certificateResolver, authContextMiddleware.PropagateAuthentication)
	exitOnError(err, "Failed configuring external graphQL handler")

	internalGqlServer, err := config.PrepareInternalGraphQLServer(cfg, tokenResolver)
	exitOnError(err, "Failed configuring internal graphQL handler")

	hydratorServer, err := config.PrepareHydratorServer(cfg, internalComponents.TokenService, internalComponents.SubjectConsts, internalComponents.RevocationsRepository)
	exitOnError(err, "Failed configuring hydrator handler")

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		log.Printf("External GraphQL API listening on %s...", cfg.ExternalAddress)
		if err := externalGqlServer.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	go func() {
		log.Printf("Internal GraphQL API listening on %s...", cfg.InternalAddress)
		if err := internalGqlServer.ListenAndServe(); err != nil {
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

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func newCoreClientSet(interval, pollingTimeout, timeout time.Duration) (*kubernetes.Clientset, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		logrus.Warnf("Failed to read in cluster Config: %s", err.Error())
		logrus.Info("Trying to initialize with local Config")
		home := homedir.HomeDir()
		k8sConfPath := filepath.Join(home, ".kube", "Config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", k8sConfPath)
		if err != nil {
			return nil, errors.Errorf("failed to read k8s in-cluster configuration, %s", err.Error())
		}
	}

	k8sConfig.Timeout = timeout

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client, %s", err.Error())
	}

	err = wait.PollImmediate(interval, pollingTimeout, func() (bool, error) {
		_, err := coreClientset.ServerVersion()
		if err != nil {
			logrus.Debugf("Failed to access API Server: %s", err.Error())
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	logrus.Info("Successfully initialized kubernetes client")
	return coreClientset, nil
}
