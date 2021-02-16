package tests

import (
	"crypto/rsa"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-incubator/compass/tests/pkg/testkit-connector/connector"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/kyma-incubator/compass/tests/pkg/testkit-connector"
	"github.com/sirupsen/logrus"
)

const (
	apiAccessTimeout  = 60 * time.Second
	apiAccessInterval = 2 * time.Second
)

var (
	config           testkit_connector.TestConfig
	internalClient   *connector.InternalClient
	hydratorClient   *connector.HydratorClient
	connectorClient  *connector.TokenSecuredClient
	configmapCleaner *testkit_connector.ConfigmapCleaner

	clientKey *rsa.PrivateKey
)

func TestMain(m *testing.M) {
	logrus.Info("Starting Connector Test")

	cfg, err := testkit_connector.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		os.Exit(1)
	}

	config = cfg
	clientKey, err = testkit_connector.GenerateKey()
	if err != nil {
		logrus.Errorf("Failed to generate private key: %s", err.Error())
		os.Exit(1)
	}
	internalClient = connector.NewInternalClient(config.InternalConnectorURL)
	hydratorClient = connector.NewHydratorClient(config.HydratorURL)
	connectorClient = connector.NewConnectorClient(config.ConnectorURL)

	configmapInterface, err := newConfigMapInterface()
	if err != nil {
		logrus.Errorf("Failed to create config map interface: %s", err.Error())
		os.Exit(1)
	}
	configmapCleaner = testkit_connector.NewConfigMapCleaner(configmapInterface, config.RevocationConfigMapName)

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

	configmapInterface := coreClientSet.CoreV1().ConfigMaps(config.RevocationConfigMapNamespace)
	return configmapInterface, nil
}
