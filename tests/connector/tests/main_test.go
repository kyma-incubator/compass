package tests

import (
	"context"
	"crypto/rsa"
	"os"
	"path/filepath"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/config"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	cfg                          config.ConnectorTestConfig
	directorClient               *clients.CertSecuredGraphQLClient
	directorAppsForRuntimeClient *clients.CertSecuredGraphQLClient
	hydratorClient               *clients.HydratorClient
	connectorClient              *clients.TokenSecuredClient
	configmapCleaner             *k8s.ConfigmapCleaner
	ctx                          context.Context
	clientKey                    *rsa.PrivateKey
	appsForRuntimeTenantID       string
)

func TestMain(m *testing.M) {
	log.D().Info("Starting Connector Test")
	cfg = config.ConnectorTestConfig{}
	config.ReadConfig(&cfg)
	ctx = context.Background()

	cc, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := util.WaitForCache(cc); err != nil {
		log.D().Fatal(err)
	}

	appsForRuntimeTenantID = cfg.AppsForRuntimeTenant

	key, err := certs.GenerateKey()
	if err != nil {
		log.D().Errorf("Failed to generate private key: %s", err.Error())
		os.Exit(1)
	}
	clientKey = key
	directorClient, err = clients.NewCertSecuredGraphQLClient(ctx, cfg.DirectorExternalCertSecuredURL, cfg.Tenant, cc.Get()[cfg.ExternalClientCertSecretName].PrivateKey, cc.Get()[cfg.ExternalClientCertSecretName].Certificate, cfg.SkipSSLValidation)
	if err != nil {
		log.D().Errorf("Failed to create director client: %s", err.Error())
		os.Exit(1)
	}
	directorAppsForRuntimeClient, err = clients.NewCertSecuredGraphQLClient(ctx, cfg.DirectorExternalCertSecuredURL, cfg.AppsForRuntimeTenant, cc.Get()[cfg.ExternalClientCertSecretName].PrivateKey, cc.Get()[cfg.ExternalClientCertSecretName].Certificate, cfg.SkipSSLValidation)
	if err != nil {
		log.D().Errorf("Failed to create director client: %s", err.Error())
		os.Exit(1)
	}

	hydratorClient = clients.NewHydratorClient(cfg.HydratorURL)
	connectorClient = clients.NewTokenSecuredClient(cfg.ConnectorURL)

	configmapInterface, err := newConfigMapInterface()
	if err != nil {
		log.D().Errorf("Failed to create config map interface: %s", err.Error())
		os.Exit(1)
	}
	configmapCleaner = k8s.NewConfigMapCleaner(configmapInterface, cfg.RevocationConfigMapName)

	exitCode := m.Run()

	log.D().Info("Tests finished. Exit code: ", exitCode)

	os.Exit(exitCode)
}

func newConfigMapInterface() (v1.ConfigMapInterface, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		log.D().Warnf("Failed to read in cluster config: %s", err.Error())
		log.D().Info("Trying to initialize with local config")
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
