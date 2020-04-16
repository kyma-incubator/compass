package queue

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/failure"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/stages"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
)

type ProvisioningTimeouts struct {
	Installation       time.Duration `envconfig:"default=60m"`
	Upgrade            time.Duration `envconfig:"default=60m"`
	AgentConfiguration time.Duration `envconfig:"default=15m"`
	AgentConnection    time.Duration `envconfig:"default=15m"`
}

func CreateInstallationQueue(
	timeouts ProvisioningTimeouts,
	factory dbsession.Factory,
	installationClient installation.Service,
	configurator runtime.Configurator,
	ccClientConstructor stages.CompassConnectionClientConstructor,
	directorClient director.DirectorClient) OperationQueue {

	waitForAgentToConnectStep := stages.NewWaitForAgentToConnectStep(ccClientConstructor, model.FinishedStage, timeouts.AgentConnection, directorClient)
	configureAgentStep := stages.NewConnectAgentStep(configurator, waitForAgentToConnectStep.Name(), timeouts.AgentConfiguration)
	waitForInstallStep := stages.NewWaitForInstallationStep(installationClient, configureAgentStep.Name(), timeouts.Installation)
	installStep := stages.NewInstallKymaStep(installationClient, waitForInstallStep.Name(), 20*time.Minute)

	installSteps := map[model.OperationStage]operations.Step{
		model.WaitForAgentToConnect:  waitForAgentToConnectStep,
		model.ConnectRuntimeAgent:    configureAgentStep,
		model.WaitingForInstallation: waitForInstallStep,
		model.StartingInstallation:   installStep,
	}

	installationExecutor := operations.NewExecutor(factory.NewReadWriteSession(), model.Provision, installSteps, failure.NewNoopFailureHandler())

	return NewQueue(installationExecutor)
}

func CreateUpgradeQueue(
	timeouts ProvisioningTimeouts,
	factory dbsession.Factory,
	installationClient installation.Service) OperationQueue {

	updatingUpgradeStep := stages.NewUpdateUpgradeStateStep(factory.NewWriteSession(), model.FinishedStage, 5*time.Minute)
	waitForInstallStep := stages.NewWaitForInstallationStep(installationClient, updatingUpgradeStep.Name(), timeouts.Installation)
	upgradeStep := stages.NewUpgradeKymaStep(installationClient, waitForInstallStep.Name(), 10*time.Minute)

	upgradeSteps := map[model.OperationStage]operations.Step{
		model.UpdatingUpgradeState:   updatingUpgradeStep,
		model.WaitingForInstallation: waitForInstallStep,
		model.StartingUpgrade:        upgradeStep,
	}

	upgradeExecutor := operations.NewExecutor(factory.NewReadWriteSession(),
		model.Upgrade,
		upgradeSteps,
		failure.NewUpgradeFailureHandler(factory.NewWriteSession()),
	)

	return NewQueue(upgradeExecutor)
}
