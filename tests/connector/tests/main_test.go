package tests

import (
	"context"
	"crypto/rsa"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

const (
	apiAccessTimeout  = 60 * time.Second
	apiAccessInterval = 2 * time.Second
)

var (
	cfg                     config.ConnectorTestConfig
	directorClient          *clients.StaticUserClient
	connectorHydratorClient *clients.HydratorClient
	directorHydratorClient  *clients.HydratorClient
	connectorClient         *clients.TokenSecuredClient
	configmapCleaner        *k8s.ConfigmapCleaner
	ctx                     context.Context
	clientKey               *rsa.PrivateKey
)

func TestMain(m *testing.M) {
	logrus.Info("Starting Connector Test")
	cfg = config.ConnectorTestConfig{}
	config.ReadConfig(&cfg)
	ctx = context.Background()

	key, err := certs.GenerateKey()
	if err != nil {
		logrus.Errorf("Failed to generate private key: %s", err.Error())
		os.Exit(1)
	}
	clientKey = key
	directorClient, err = clients.NewStaticUserClient(ctx, cfg.DirectorURL, cfg.Tenant)
	if err != nil {
		logrus.Errorf("Failed to create director client: %s", err.Error())
		os.Exit(1)
	}
	connectorHydratorClient = clients.NewHydratorClient(cfg.ConnectorHydratorURL)
	directorHydratorClient = clients.NewHydratorClient(cfg.DirectorHydratorURL)
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
