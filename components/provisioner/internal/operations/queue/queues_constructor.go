package queue

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/gardener"

	gardener_apis "github.com/gardener/gardener/pkg/client/core/clientset/versioned/typed/core/v1beta1"
	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/failure"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/stages/deprovisioning"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/stages/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/stages/upgrade"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ProvisioningTimeouts struct {
	Installation       time.Duration `envconfig:"default=60m"`
	Upgrade            time.Duration `envconfig:"default=60m"`
	AgentConfiguration time.Duration `envconfig:"default=15m"`
	AgentConnection    time.Duration `envconfig:"default=15m"`
}

func CreateProvisioningQueue(
	timeouts ProvisioningTimeouts,
	factory dbsession.Factory,
	installationClient installation.Service,
	configurator runtime.Configurator,
	ccClientConstructor provisioning.CompassConnectionClientConstructor,
	directorClient director.DirectorClient,
	shootClient gardener_apis.ShootInterface,
	secretsClient v1core.SecretInterface) OperationQueue {

	waitForAgentToConnectStep := provisioning.NewWaitForAgentToConnectStep(ccClientConstructor, model.FinishedStage, timeouts.AgentConnection, directorClient)
	configureAgentStep := provisioning.NewConnectAgentStep(configurator, waitForAgentToConnectStep.Name(), timeouts.AgentConfiguration)
	waitForInstallStep := provisioning.NewWaitForInstallationStep(installationClient, configureAgentStep.Name(), timeouts.Installation)
	installStep := provisioning.NewInstallKymaStep(installationClient, waitForInstallStep.Name(), 20*time.Minute)
	waitForClusterCreationStep := provisioning.NewWaitForClusterCreationStep(shootClient, factory.NewReadWriteSession(), gardener.NewKubeconfigProvider(secretsClient), installStep.Name(), 40*time.Minute)
	waitForClusterDomainStep := provisioning.NewWaitForClusterDomainStep(shootClient, directorClient, waitForClusterCreationStep.Name(), 10*time.Minute)

	provisionSteps := map[model.OperationStage]operations.Step{
		model.WaitForAgentToConnect:     waitForAgentToConnectStep,
		model.ConnectRuntimeAgent:       configureAgentStep,
		model.WaitingForInstallation:    waitForInstallStep,
		model.StartingInstallation:      installStep,
		model.WaitingForClusterDomain:   waitForClusterDomainStep,
		model.WaitingForClusterCreation: waitForClusterCreationStep,
	}

	provisioningExecutor := operations.NewExecutor(
		factory.NewReadWriteSession(),
		model.Provision,
		provisionSteps,
		failure.NewNoopFailureHandler(),
		directorClient,
	)

	return NewQueue(provisioningExecutor)
}

func CreateUpgradeQueue(
	timeouts ProvisioningTimeouts,
	factory dbsession.Factory,
	directorClient director.DirectorClient,
	installationClient installation.Service) OperationQueue {

	updatingUpgradeStep := upgrade.NewUpdateUpgradeStateStep(factory.NewWriteSession(), model.FinishedStage, 5*time.Minute)
	waitForInstallStep := provisioning.NewWaitForInstallationStep(installationClient, updatingUpgradeStep.Name(), timeouts.Installation)
	upgradeStep := upgrade.NewUpgradeKymaStep(installationClient, waitForInstallStep.Name(), 10*time.Minute)

	upgradeSteps := map[model.OperationStage]operations.Step{
		model.UpdatingUpgradeState:   updatingUpgradeStep,
		model.WaitingForInstallation: waitForInstallStep,
		model.StartingUpgrade:        upgradeStep,
	}

	upgradeExecutor := operations.NewExecutor(factory.NewReadWriteSession(),
		model.Upgrade,
		upgradeSteps,
		failure.NewUpgradeFailureHandler(factory.NewWriteSession()),
		directorClient,
	)

	return NewQueue(upgradeExecutor)
}

func CreateDeprovisioningQueue(
	factory dbsession.Factory,
	installationClient installation.Service,
	directorClient director.DirectorClient,
	shootClient gardener_apis.ShootInterface,
	deleteDelay time.Duration) OperationQueue {

	// TODO: consider adding timeouts to the configuration
	waitForClusterDeletion := deprovisioning.NewWaitForClusterDeletionStep(shootClient, factory, directorClient, model.FinishedStage, 5*time.Minute)
	deleteCluster := deprovisioning.NewDeleteClusterStep(shootClient, waitForClusterDeletion.Name(), 5*time.Minute)
	triggerKymaUninstall := deprovisioning.NewTriggerKymaUninstallStep(installationClient, deleteCluster.Name(), 5*time.Minute, deleteDelay)

	deprovisioningSteps := map[model.OperationStage]operations.Step{
		model.DeleteCluster:          deleteCluster,
		model.WaitForClusterDeletion: waitForClusterDeletion,
		model.TriggerKymaUninstall:   triggerKymaUninstall,
	}

	deprovisioningExecutor := operations.NewExecutor(
		factory.NewReadWriteSession(),
		model.Deprovision,
		deprovisioningSteps,
		failure.NewNoopFailureHandler(),
		directorClient,
	)

	return NewQueue(deprovisioningExecutor)
}
