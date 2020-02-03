package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/gardener"
	"github.com/kyma-incubator/compass/components/provisioner/internal/graphql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"
	"github.com/kyma-incubator/compass/components/provisioner/internal/oauth"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/hyperscaler"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	"github.com/pkg/errors"

	gardener_apis "github.com/gardener/gardener/pkg/client/garden/clientset/versioned/typed/garden/v1beta1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	databaseConnectionRetries = 20
)

func newProvisioningService(
	gardenerProject string,
	provisioner provisioning.Provisioner,
	dbsFactory dbsession.Factory,
	releaseRepo release.ReadRepository,
	gardenerSecrets v1.SecretInterface,
	compassSecrets v1.SecretInterface,
	directorService director.DirectorClient) provisioning.Service {
	uuidGenerator := uuid.NewUUIDGenerator()

	compassAccountPool := hyperscaler.NewAccountPool(compassSecrets)
	var gardenerAccountPool hyperscaler.AccountPool = nil
	if gardenerSecrets != nil {
		gardenerAccountPool = hyperscaler.NewAccountPool(gardenerSecrets)
	}

	accountProvider := hyperscaler.NewAccountProvider(compassAccountPool, gardenerAccountPool)

	inputConverter := provisioning.NewInputConverter(uuidGenerator, releaseRepo, gardenerProject, accountProvider)

	graphQLConverter := provisioning.NewGraphQLConverter()

	return provisioning.NewProvisioningService(inputConverter, graphQLConverter, directorService, dbsFactory, provisioner, uuidGenerator)
}

func newDirectorClient(config config) (director.DirectorClient, error) {
	secretsRepo, err := newSecretsInterface(config.CredentialsNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create secrets interface")
	}

	gqlClient := graphql.NewGraphQLClient(config.DirectorURL, true, config.SkipDirectorCertVerification)
	oauthClient := oauth.NewOauthClient(newHTTPClient(config.SkipDirectorCertVerification), compassSecrets, config.OauthCredentialsSecretName)
	oauthClient := oauth.NewOauthClient(newHTTPClient(config.SkipDirectorCertVerification), secretsRepo, config.OauthCredentialsSecretName)

	return director.NewDirectorClient(gqlClient, oauthClient), nil
}

func newShootController(
	cfg config,
	gardenerNamespace string,
	gardenerClusterCfg *restclient.Config,
	gardenerClientSet *gardener_apis.GardenV1beta1Client,
	dbsFactory dbsession.Factory,
	installationSvc installation.Service,
	direcotrClietnt director.DirectorClient,
	runtimeConfigurator runtime.Configurator) (*gardener.ShootController, error) {
	gardenerClusterClient, err := kubernetes.NewForConfig(gardenerClusterCfg)
	if err != nil {
		return nil, err
	}

	secretsInterface := gardenerClusterClient.CoreV1().Secrets(gardenerNamespace)

	shootClient := gardenerClientSet.Shoots(gardenerNamespace)

	return gardener.NewShootController(gardenerNamespace, gardenerClusterCfg, shootClient, secretsInterface, installationSvc, dbsFactory, cfg.Installation.Timeout, direcotrClietnt, runtimeConfigurator)
}

func newSecretsInterface(namespace string) (v1.SecretInterface, error) {
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

	return coreClientset.CoreV1().Secrets(namespace), nil
}

func newGardenerClusterConfig(cfg config) (*restclient.Config, error) {
	rawKubeconfig, err := ioutil.ReadFile(cfg.Gardener.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Gardener Kubeconfig from path %s: %s", cfg.Gardener.KubeconfigPath, err.Error())
	}

	gardenerClusterConfig, err := gardener.Config(rawKubeconfig)
	if err != nil {
		return nil, fmt.Errorf("")
	}

	return gardenerClusterConfig, nil
}

/*
old code

func newGardenerSecretsInterface(config config) (v1.SecretInterface, error) {

	kubeconfig := config.GardenerKubeconfigPath
	if len(kubeconfig) == 0 {
		logrus.Warn("GardenerKubeconfigPath environment value missing")
		return nil, nil
	}
	k8sConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, errors.Errorf("failed to build kubeconfig from GardenerKubeconfigPath: %s, error: %s", kubeconfig, err.Error())
	}

	coreClientset, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client for Gardener, %s", err.Error())
	}

	//TODO: what is the Gardener credentials namespace?
	return coreClientset.CoreV1().Secrets("default"), nil
}

func newResolver(config config, persistenceService persistence.Service, releaseRepo release.Repository) (*api.Resolver, error) {
	compassSecrets, err := newSecretsInterface(config.CredentialsNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Compass secrets interface")
	}
	gardenerSecrets, err := newGardenerSecretsInterface(config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Gardener secrets interface")
	}

	return api.NewResolver(
		newProvisioningService(config, persistenceService, compassSecrets, gardenerSecrets, releaseRepo),
		api.NewValidator(persistenceService),
	), nil
}


*/

func newHTTPClient(skipCertVeryfication bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipCertVeryfication},
		},
		Timeout: 30 * time.Second,
	}
}
