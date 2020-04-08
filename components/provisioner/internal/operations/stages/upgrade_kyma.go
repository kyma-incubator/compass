package stages

import (
	"errors"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util/k8s"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/sirupsen/logrus"
)

// TODO: consider extracting Name and TimeLimit to some common struct for all steps

type UpgradeKymaStep struct {
	installationClient installation.Service
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

func NewUpgradeKymaStep(installationClient installation.Service, nextStep model.OperationStage, timeLimit time.Duration) *UpgradeKymaStep {
	return &UpgradeKymaStep{
		installationClient: installationClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
	}
}

func (s *UpgradeKymaStep) Name() model.OperationStage {
	return model.StartingUpgrade
}

func (s *UpgradeKymaStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *UpgradeKymaStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	if cluster.Kubeconfig == nil {
		return operations.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	k8sConfig, err := k8s.ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		return operations.StageResult{}, fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
	}

	installationState, err := s.installationClient.CheckInstallationState(k8sConfig) // TODO: modify signature of this method
	if err != nil {
		installErr := installationSDK.InstallationError{}
		if errors.As(err, &installErr) {
			logger.Warnf("Upgrade already in progress, proceeding to next step...")
			return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
		}

		return operations.StageResult{}, fmt.Errorf("error: failed to check installation CR state: %s", err.Error())
	}

	if installationState.State == installationSDK.NoInstallationState {
		return operations.StageResult{}, operations.NewNonRecoverableError(fmt.Errorf("error: Installation CR not found in the cluster, cannot trigger upgrade"))
	}

	if installationState.State == "Installed" {
		err = s.installationClient.TriggerUpgrade(
			k8sConfig,
			cluster.KymaConfig.Release,
			cluster.KymaConfig.GlobalConfiguration,
			cluster.KymaConfig.Components)
		if err != nil {
			return operations.StageResult{}, fmt.Errorf("error: failed to trigger upgrade: %s", err.Error())
		}
	}

	if installationState.State == "InProgress" {
		// TODO: How should it handle when installation in progress? Should it check the Kyma version?

		logger.Warnf("Upgrade already in progress, proceeding to next step...")
		return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}
