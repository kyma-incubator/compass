package apitests

import (
	"context"
	"crypto/rsa"
	"os"
	"path/filepath"
	"testing"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/director"

	"github.com/pkg/errors"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit/connector"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/kyma-incubator/compass/tests/connector-tests/test/testkit"
	"github.com/sirupsen/logrus"
)

var (
	config                  testkit.TestConfig
	directorClient          *director.Client
	connectorHydratorClient *connector.HydratorClient
	directorHydratorClient  *director.HydratorClient
	connectorClient         *connector.TokenSecuredClient
	configmapCleaner        *testkit.ConfigmapCleaner

	clientKey *rsa.PrivateKey
)

func TestMain(m *testing.M) {
	logrus.Info("Starting Connector Test")

	cfg, err := testkit.ReadConfig()
	if err != nil {
		logrus.Errorf("Failed to read config: %s", err.Error())
		os.Exit(1)
	}

	config = cfg
	clientKey, err = testkit.GenerateKey()
	if err != nil {
		logrus.Errorf("Failed to generate private key: %s", err.Error())
		os.Exit(1)
	}

	directorClient, err = director.NewClient(context.Background(), config.DirectorURL, config.Tenant)
	if err != nil {
		logrus.Errorf("Failed to create director client: %s", err.Error())
		os.Exit(1)
	}
	connectorHydratorClient = connector.NewHydratorClient(config.ConnectorHydratorURL)
	directorHydratorClient = director.NewHydratorClient(config.DirectorHydratorURL)
	connectorClient = connector.NewConnectorClient(config.ConnectorURL)

	configmapInterface, err := newConfigMapInterface()
	if err != nil {
		logrus.Errorf("Failed to create config map interface: %s", err.Error())
		os.Exit(1)
	}
	configmapCleaner = testkit.NewConfigMapCleaner(configmapInterface, config.RevocationConfigMapName)

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
