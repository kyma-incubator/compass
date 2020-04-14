package queue

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/failure"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/stages"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
)

func CreateInstallationQueue(
	installationTimeout time.Duration,
	factory dbsession.Factory,
	installationClient installation.Service,
	configurator runtime.Configurator) OperationQueue {
	configureAgentStep := stages.NewConnectAgentStep(configurator, model.FinishedStage, 10*time.Minute)
	waitForInstallStep := stages.NewWaitForInstallationStep(installationClient, configureAgentStep.Stage(), installationTimeout)
	installStep := stages.NewInstallKymaStep(installationClient, waitForInstallStep.Stage(), 10*time.Minute)

	installSteps := map[model.OperationStage]operations.Step{
		model.ConnectRuntimeAgent:    configureAgentStep,
		model.WaitingForInstallation: waitForInstallStep,
		model.StartingInstallation:   installStep,
	}

	installationExecutor := operations.NewStepsExecutor(factory.NewReadWriteSession(), model.Provision, installSteps, failure.NewNoopFailureHandler())

	return NewQueue(installationExecutor)
}

func CreateUpgradeQueue(
	upgradeTimeout time.Duration,
	factory dbsession.Factory,
	installationClient installation.Service) OperationQueue {

	updatingUpgradeStep := stages.NewUpdateUpgradeStateStep(factory.NewWriteSession(), model.FinishedStage, 5*time.Minute)
	waitForInstallStep := stages.NewWaitForInstallationStep(installationClient, updatingUpgradeStep.Stage(), upgradeTimeout)
	upgradeStep := stages.NewUpgradeKymaStep(installationClient, waitForInstallStep.Stage(), 10*time.Minute)

	upgradeSteps := map[model.OperationStage]operations.Step{
		model.UpdatingUpgradeState:   updatingUpgradeStep,
		model.WaitingForInstallation: waitForInstallStep,
		model.StartingUpgrade:        upgradeStep,
	}

	upgradeExecutor := operations.NewStepsExecutor(factory.NewReadWriteSession(),
		model.Upgrade,
		upgradeSteps,
		failure.NewUpgradeFailureHandler(factory.NewWriteSession()),
	)

	return NewQueue(upgradeExecutor)
}
