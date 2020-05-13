package queue

import (
	"time"

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
	shootClient gardener_apis.ShootInterface) OperationQueue {

	waitForAgentToConnectStep := provisioning.NewWaitForAgentToConnectStep(ccClientConstructor, model.FinishedStage, timeouts.AgentConnection, directorClient)
	configureAgentStep := provisioning.NewConnectAgentStep(configurator, waitForAgentToConnectStep.Name(), timeouts.AgentConfiguration)
	waitForInstallStep := provisioning.NewWaitForInstallationStep(installationClient, configureAgentStep.Name(), timeouts.Installation)
	installStep := provisioning.NewInstallKymaStep(installationClient, waitForInstallStep.Name(), 20*time.Minute)
	waitForClusterInitializationStep := provisioning.NewWaitForClusterInitializationStep(shootClient, factory, installStep.Name(), 20*time.Minute)
	waitForClusterCreationStep := provisioning.NewWaitForClusterCreationStep(shootClient, factory, directorClient, waitForClusterInitializationStep.Name(), 10*time.Minute)

	installSteps := map[model.OperationStage]operations.Step{
		model.WaitForAgentToConnect:           waitForAgentToConnectStep,
		model.ConnectRuntimeAgent:             configureAgentStep,
		model.WaitingForInstallation:          waitForInstallStep,
		model.StartingInstallation:            installStep,
		model.WaitingForClusterCreation:       waitForClusterCreationStep,
		model.WaitingForClusterInitialization: waitForClusterInitializationStep,
	}

	provisioningExecutor := operations.NewExecutor(
		factory.NewReadWriteSession(),
		model.Provision,
		installSteps,
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
	shootClient gardener_apis.ShootInterface) OperationQueue {

	// TODO: consider adding timeouts to the configuration
	waitForClusterDeletion := deprovisioning.NewDeprovisionClusterStep(installationClient, shootClient, factory, directorClient, model.FinishedStage, 10*time.Second)
	deprovisionCluster := deprovisioning.NewTriggerKymaUninstallStep(installationClient, shootClient, waitForClusterDeletion.Name(), 10*time.Second)

	deprovisioningSteps := map[model.OperationStage]operations.Step{
		model.DeprovisionCluster:     deprovisionCluster,
		model.WaitForClusterDeletion: waitForClusterDeletion,
	}

	deprovisioningExecutor := operations.NewExecutor(
		factory.NewReadWriteSession(),
		model.Provision,
		deprovisioningSteps,
		failure.NewNoopFailureHandler(),
		directorClient,
	)

	return NewQueue(deprovisioningExecutor)
}
