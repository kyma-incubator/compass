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

type WaitForInstallationStep struct {
	installationClient installation.Service
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

func NewWaitForInstallationStep(installationClient installation.Service, nextStep model.OperationStage, timeLimit time.Duration) *WaitForInstallationStep {
	return &WaitForInstallationStep{
		installationClient: installationClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
	}
}

func (s *WaitForInstallationStep) Name() model.OperationStage {
	return model.WaitingForInstallation
}

func (s *WaitForInstallationStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *WaitForInstallationStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	if cluster.Kubeconfig == nil {
		return operations.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	k8sConfig, err := k8s.ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		return operations.StageResult{}, fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
	}

	installationState, err := s.installationClient.CheckInstallationState(k8sConfig)
	if err != nil {
		installErr := installationSDK.InstallationError{}
		if errors.As(err, &installErr) {
			logger.Warnf("installation error occurred: %s", installErr.Error())
			return operations.StageResult{Stage: s.Name(), Delay: 30 * time.Second}, nil
		}

		return operations.StageResult{}, fmt.Errorf("error: failed to check installation state: %s", err.Error())
	}

	if installationState.State == "Installed" {
		logger.Infof("Installation completed: %s", installationState.Description)
		return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
	}

	if installationState.State == installationSDK.NoInstallationState {
		return operations.StageResult{}, fmt.Errorf("installation not yet started")
	}

	logger.Infof("Installation in progress: %s", installationState.Description)
	return operations.StageResult{Stage: s.Name(), Delay: 30 * time.Second}, nil
}
