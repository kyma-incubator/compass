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
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	databaseConnectionRetries = 20
	defaultSyncPeriod         = 3 * time.Minute
)

func newProvisioningService(
	gardenerProject string,
	provisioner provisioning.Provisioner,
	dbsFactory dbsession.Factory,
	releaseRepo release.ReadRepository,
	compassSecrets v1.SecretInterface,
	gardenerSecrets v1.SecretInterface,
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

func newDirectorClient(config config, compassSecrets v1.SecretInterface) (director.DirectorClient, error) {

	gqlClient := graphql.NewGraphQLClient(config.DirectorURL, true, config.SkipDirectorCertVerification)
	oauthClient := oauth.NewOauthClient(newHTTPClient(config.SkipDirectorCertVerification), compassSecrets, config.OauthCredentialsSecretName)

	return director.NewDirectorClient(gqlClient, oauthClient), nil
}

func newShootController(
	cfg config,
	gardenerNamespace string,
	gardenerClusterCfg *restclient.Config,
	gardenerClientSet *gardener_apis.CoreV1beta1Client,
	dbsFactory dbsession.Factory,
	installationSvc installation.Service,
	directorClient director.DirectorClient,
	runtimeConfigurator runtime.Configurator) (*gardener.ShootController, error) {

	gardenerSecrets, err := newGardenerSecretsInterface(gardenerClusterCfg, gardenerNamespace)
	if err != nil {
		return nil, err
	}

	shootClient := gardenerClientSet.Shoots(gardenerNamespace)

	syncPeriod := defaultSyncPeriod

	mgr, err := ctrl.NewManager(gardenerClusterCfg, ctrl.Options{SyncPeriod: &syncPeriod, Namespace: gardenerNamespace})
	if err != nil {
		return nil, fmt.Errorf("unable to create shoot controller manager: %w", err)
	}

	return gardener.NewShootController(gardenerNamespace, mgr, shootClient, gardenerSecrets, installationSvc, dbsFactory, cfg.Installation.Timeout, directorClient, runtimeConfigurator)
}

func newClusterSecretsInterface(namespace string) (v1.SecretInterface, error) {
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

func newGardenerSecretsInterface(gardenerClusterCfg *restclient.Config, gardenerNamespace string) (v1.SecretInterface, error) {

	gardenerClusterClient, err := kubernetes.NewForConfig(gardenerClusterCfg)
	if err != nil {
		return nil, err
	}

	return gardenerClusterClient.CoreV1().Secrets(gardenerNamespace), nil
}

func newHTTPClient(skipCertVeryfication bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipCertVeryfication},
		},
		Timeout: 30 * time.Second,
	}
}
