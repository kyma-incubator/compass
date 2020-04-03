package stages

import (
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operation"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/sirupsen/logrus"
	"time"
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

// TODO: I think it would be better to return some step result
func (s *UpgradeKymaStep) Run(cluster model.Cluster, logger logrus.FieldLogger) (operation.StageResult, error) {

	if cluster.Kubeconfig == nil {
		// TODO: recoverable or not?
		return operation.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	// TODO: cleanup kubeconfig stuff
	k8sConfig, err := ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		// TODO: recoverable or not?
		return operation.StageResult{}, fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
	}

	installationState, err := s.installationClient.CheckInstallationState(k8sConfig) // TODO: modify signature of this method
	if err != nil {
		installErr := installationSDK.InstallationError{}
		if errors.As(err, &installErr) {
			logger.Warnf("Upgrade already in progress, proceeding to next step...")
			return operation.StageResult{Step: s.Name(), Delay: 0}, nil
		}

		return operation.StageResult{}, fmt.Errorf("error: failed to check installation CR state: %s", err.Error())
	}

	// TODO Start upgrade
	if installationState.State == installationSDK.NoInstallationState {

	}

	return operation.StageResult{Step: s.nextStep, Delay: 0}, nil

	//
	//if installationState.State == installationSDK.NoInstallationState {
	//	// TODO: it needs to run apply instead of create
	//	err := s.installationClient.TriggerInstallation(
	//		[]byte(*cluster.Kubeconfig),
	//		cluster.KymaConfig.Release,
	//		cluster.KymaConfig.GlobalConfiguration,
	//		cluster.KymaConfig.Components)
	//	if err != nil {
	//		// TODO: if it runs apply then recoverable error (else not)
	//		return operation, 2 * time.Second, fmt.Errorf("error: failed to start installation: %s", err.Error())
	//	}
	//
	//	// TODO: update operation that installation started
	//}
	//
	//if installationState.State == "Installed" {
	//	operation.Message = "Installation completed"
	//	logger.Infof("Installation completed: %s", installationState.Description)
	//	return operation, 0, nil
	//}
	//
	//operation.Message = fmt.Sprintf("Installation in progress: %s", installationState.Description)
	//
	//logger.Infof("Installation in progress: %s", installationState.Description)
	//return operation, 10 * time.Second, nil
}
