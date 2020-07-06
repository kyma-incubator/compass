package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"time"

	"github.com/kyma-project/control-plane/components/provisioner/internal/operations/queue"

	"github.com/kyma-project/control-plane/components/provisioner/internal/provisioning/persistence/dbsession"

	"github.com/kyma-project/control-plane/components/provisioner/internal/director"
	"github.com/kyma-project/control-plane/components/provisioner/internal/gardener"
	"github.com/kyma-project/control-plane/components/provisioner/internal/graphql"
	"github.com/kyma-project/control-plane/components/provisioner/internal/installation/release"
	"github.com/kyma-project/control-plane/components/provisioner/internal/oauth"
	"github.com/kyma-project/control-plane/components/provisioner/internal/provisioning"
	"github.com/kyma-project/control-plane/components/provisioner/internal/uuid"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	databaseConnectionRetries = 20
	defaultSyncPeriod         = 10 * time.Minute
)

func newProvisioningService(
	gardenerProject string,
	provisioner provisioning.Provisioner,
	dbsFactory dbsession.Factory,
	releaseRepo release.Provider,
	directorService director.DirectorClient,
	provisioningQueue queue.OperationQueue,
	deprovisioningQueue queue.OperationQueue,
	upgradeQueue queue.OperationQueue,
	shootUpgradeQueue queue.OperationQueue) provisioning.Service {
	uuidGenerator := uuid.NewUUIDGenerator()

	inputConverter := provisioning.NewInputConverter(uuidGenerator, releaseRepo, gardenerProject)
	graphQLConverter := provisioning.NewGraphQLConverter()

	return provisioning.NewProvisioningService(inputConverter, graphQLConverter, directorService, dbsFactory, provisioner, uuidGenerator, provisioningQueue, deprovisioningQueue, upgradeQueue, shootUpgradeQueue)
}

func newDirectorClient(config config) (director.DirectorClient, error) {
	secretsRepo, err := newSecretsInterface(config.CredentialsNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create secrets interface")
	}

	gqlClient := graphql.NewGraphQLClient(config.DirectorURL, true, config.SkipDirectorCertVerification)
	oauthClient := oauth.NewOauthClient(newHTTPClient(config.SkipDirectorCertVerification), secretsRepo, config.OauthCredentialsSecretName)

	return director.NewDirectorClient(gqlClient, oauthClient), nil
}

func newShootController(gardenerNamespace string, gardenerClusterCfg *restclient.Config, dbsFactory dbsession.Factory, auditLogTenantConfigPath string) (*gardener.ShootController, error) {

	syncPeriod := defaultSyncPeriod

	mgr, err := ctrl.NewManager(gardenerClusterCfg, ctrl.Options{SyncPeriod: &syncPeriod, Namespace: gardenerNamespace})
	if err != nil {
		return nil, fmt.Errorf("unable to create shoot controller manager: %w", err)
	}

	return gardener.NewShootController(mgr, dbsFactory, auditLogTenantConfigPath)
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
		return nil, fmt.Errorf("failed to create Gardener cluster config: %s", err.Error())
	}

	return gardenerClusterConfig, nil
}

func newHTTPClient(skipCertVerification bool) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipCertVerification},
		},
		Timeout: 30 * time.Second,
	}
}
