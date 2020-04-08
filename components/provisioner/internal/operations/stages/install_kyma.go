package stages

import (
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util/k8s"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/sirupsen/logrus"
	"time"
)

type InstallKymaStep struct {
	installationClient installation.Service
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

func NewInstallKymaStep(installationClient installation.Service, nextStep model.OperationStage, timeLimit time.Duration) *InstallKymaStep {
	return &InstallKymaStep{
		installationClient: installationClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
	}
}

func (s *InstallKymaStep) Name() model.OperationStage {
	return model.StartingInstallation
}

func (s *InstallKymaStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *InstallKymaStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	if cluster.Kubeconfig == nil {
		return operations.StageResult{}, fmt.Errorf("error: kubeconfig is nil")
	}

	k8sConfig, err := k8s.ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		return operations.StageResult{}, fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
	}

	// TODO: check if installation is started
	installationState, err := s.installationClient.CheckInstallationState(k8sConfig) // TODO: modify signature of this method
	if err != nil {
		installErr := installationSDK.InstallationError{}
		if errors.As(err, &installErr) {
			logger.Warnf("Installation already in progress, proceeding to next step...")
			return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
		}

		return operations.StageResult{}, fmt.Errorf("error: failed to check installation state: %s", err.Error())
	}

	if installationState.State != installationSDK.NoInstallationState {
		logger.Warnf("Installation already in progress, proceeding to next step...")
		return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
	}

	// TODO: it needs to run apply instead of create
	err = s.installationClient.TriggerInstallation(
		k8sConfig,
		cluster.KymaConfig.Release,
		cluster.KymaConfig.GlobalConfiguration,
		cluster.KymaConfig.Components)
	if err != nil {
		// TODO: if it runs apply then recoverable error (else not)
		return operations.StageResult{}, fmt.Errorf("error: failed to start installation: %s", err.Error())
	}

	logger.Warnf("Installation started, proceeding to next step...")
	return operations.StageResult{Stage: s.nextStep, Delay: 30 * time.Second}, nil
}
