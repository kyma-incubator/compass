package main

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/api"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/client"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/configuration"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/database"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
	installation2 "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/pkg/errors"

	"path/filepath"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func newPersistenceService(connectionString, schemaPath string) (persistence.Service, error) {
	connection, err := database.InitializeDatabase(connectionString, schemaPath)
	if err != nil {
		return nil, err
	}

	dbSessionFactory := dbsession.NewFactory(connection)
	uuidGenerator := uuid.NewUUIDGenerator()

	return persistence.NewService(dbSessionFactory, uuidGenerator), nil
}

func newProvisioningService(config config, persistenceService persistence.Service, secrets v1.SecretInterface, artifactsProvider installation.ArtifactsProvider) provisioning.Service {
	hydroformClient := client.NewHydroformClient()
	hydroformService := hydroform.NewHydroformService(secrets, hydroformClient)
	uuidGenerator := uuid.NewUUIDGenerator()
	factory := configuration.NewConfigBuilderFactory(secrets)
	installationService := installation.NewInstallationService(config.Installation.Timeout, artifactsProvider, installation2.NewKymaInstaller, config.Installation.ErrorsCountFailureThreshold)

	return provisioning.NewProvisioningService(persistenceService, uuidGenerator, hydroformService, factory,installationService)
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

func newResolver(connectionString string, schemaFilePath string, namespace string) (*api.Resolver, error) {
	persistenceService, err := newPersistenceService(connectionString, schemaFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize persistence")
	}

	secretInterface, err := newSecretsInterface(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create secrets interface")
	}

	return api.NewResolver(newProvisioningService(persistenceService, secretInterface)), nil
}
