package steps

import (
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/sirupsen/logrus"
	"time"
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

func (s *WaitForInstallationStep) Run(operation model.Operation, cluster model.Cluster, logger logrus.FieldLogger) (installation.StepResult, error) {

	if cluster.Kubeconfig == nil {
		return installation.StepResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	// TODO: cleanup kubeconfig stuff
	k8sConfig, err := ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		return installation.StepResult{}, fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
	}

	installationState, err := s.installationClient.CheckInstallationState(k8sConfig) // TODO: modify signature of this method
	if err != nil {
		installErr := installationSDK.InstallationError{}
		if errors.As(err, &installErr) {
			logger.Warnf("installation error occurred: %s", installErr.Error())
			return installation.StepResult{Step: s.Name(), Delay: 30 * time.Second}, nil
		}

		return installation.StepResult{}, fmt.Errorf("error: failed to check installation state: %s", err.Error())
	}

	if installationState.State == "Installed" {
		logger.Infof("Installation completed: %s", installationState.Description)
		return installation.StepResult{Step: s.nextStep, Delay: 0}, nil
	}

	if installationState.State == installationSDK.NoInstallationState {
		// TODO: not recoverable?
		return installation.StepResult{}, fmt.Errorf("installation not yet started")
	}

	logger.Infof("Installation in progress: %s", installationState.Description)
	return installation.StepResult{Step: s.Name(), Delay: 30 * time.Second}, nil
}
