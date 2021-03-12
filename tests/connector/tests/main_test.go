package tests

import (
	"crypto/rsa"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/sirupsen/logrus"
)

const (
	apiAccessTimeout  = 60 * time.Second
	apiAccessInterval = 2 * time.Second
)

var (
	cfg              config.ConnectorTestConfig
	internalClient   *clients.InternalClient
	hydratorClient   *clients.HydratorClient
	connectorClient  *clients.TokenSecuredClient
	configmapCleaner *k8s.ConfigmapCleaner

	clientKey *rsa.PrivateKey
)

func TestMain(m *testing.M) {
	logrus.Info("Starting Connector Test")
	cfg := config.ConnectorTestConfig{}
	config.ReadConfig(&cfg)

	key, err := certs.GenerateKey()
	if err != nil {
		logrus.Errorf("Failed to generate private key: %s", err.Error())
		os.Exit(1)
	}

	clientKey = key
	internalClient = clients.NewInternalClient(cfg.InternalConnectorURL)
	hydratorClient = clients.NewHydratorClient(cfg.HydratorURL)
	connectorClient = clients.NewTokenSecuredClient(cfg.ConnectorURL)

	configmapInterface, err := newConfigMapInterface()
	if err != nil {
		logrus.Errorf("Failed to create config map interface: %s", err.Error())
		os.Exit(1)
	}
	configmapCleaner = k8s.NewConfigMapCleaner(configmapInterface, cfg.RevocationConfigMapName)

	exitCode := m.Run()

	logrus.Info("Tests finished. Exit code: ", exitCode)

	os.Exit(exitCode)
}

func newConfigMapInterface() (v1.ConfigMapInterface, error) {
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

	coreClientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create k8s core client, %s", err.Error())
	}

	configmapInterface := coreClientSet.CoreV1().ConfigMaps(cfg.RevocationConfigMapNamespace)
	return configmapInterface, nil
}
