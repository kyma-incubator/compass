package main

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/api"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/database"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/converters"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/hyperscaler"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/pkg/errors"

	"path/filepath"

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

func newProvisioningService(config config, persistenceService persistence.Service, compassSecrets v1.SecretInterface,
		gardenerSecrets v1.SecretInterface, releaseRepo release.ReadRepository) provisioning.Service {

	hydroformClient := client.NewHydroformClient()
	hydroformService := hydroform.NewHydroformService(hydroformClient, compassSecrets)
	uuidGenerator := uuid.NewUUIDGenerator()
	installationService := installation.NewInstallationService(config.Installation.Timeout, installationSDK.NewKymaInstaller, config.Installation.ErrorsCountFailureThreshold)

	compassAccountPool := hyperscaler.NewAccountPool(compassSecrets)
	var gardenerAccountPool hyperscaler.AccountPool = nil
	if gardenerSecrets != nil {
		gardenerAccountPool = hyperscaler.NewAccountPool(gardenerSecrets)
	}

	//TODO: integrate tenantName lookup
	tenantName := config.DefaultTenant
	accountProvider := hyperscaler.NewAccountProvider(compassAccountPool, gardenerAccountPool, tenantName)

	inputConverter := converters.NewInputConverter(uuidGenerator, releaseRepo, accountProvider)
	graphQLConverter := converters.NewGraphQLConverter()

	return provisioning.NewProvisioningService(persistenceService, inputConverter, graphQLConverter, hydroformService, installationService)
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

	//TODO: is the Gardener credentials namespace the same as the Compass credentials namespace?
	return coreClientset.CoreV1().Secrets(config.CredentialsNamespace), nil
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

	return api.NewResolver(newProvisioningService(config, persistenceService, compassSecrets, gardenerSecrets, releaseRepo)), nil
}

func initRepositories(config config, connectionString string) (persistence.Service, release.Repository, error) {
	connection, err := database.InitializeDatabase(connectionString, config.Database.SchemaFilePath, databaseConnectionRetries)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to initialize persistence")
	}
	dbSessionFactory := dbsession.NewFactory(connection)

	persistenceService := persistence.NewService(dbSessionFactory, uuid.NewUUIDGenerator())

	releaseRepo := release.NewReleaseRepository(connection, uuid.NewUUIDGenerator())

	return persistenceService, releaseRepo, nil
}
