package stages

import (
	"errors"
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/director"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util/k8s"
	installationSDK "github.com/kyma-incubator/hydroform/install/installation"
	"github.com/sirupsen/logrus"
)

type InstallKymaStep struct {
	installationClient installation.Service
	directorClient     director.DirectorClient
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

func NewInstallKymaStep(
	installationClient installation.Service,
	nextStep model.OperationStage,
	timeLimit time.Duration,
	directorClient director.DirectorClient) *InstallKymaStep {

	return &InstallKymaStep{
		installationClient: installationClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
		directorClient:     directorClient,
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

	installationState, err := s.installationClient.CheckInstallationState(k8sConfig)
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

	err = s.installationClient.TriggerInstallation(
		k8sConfig,
		cluster.KymaConfig.Release,
		cluster.KymaConfig.GlobalConfiguration,
		cluster.KymaConfig.Components)
	if err != nil {
		return operations.StageResult{}, fmt.Errorf("error: failed to start installation: %s", err.Error())
	}

	statusCondition := gqlschema.RuntimeStatusConditionInstalling
	runtimeInput := &gqlschema.RuntimeInput{
		// TODO: Add name, description and labels. Will the directorClient.GetRuntime(runtimeId) call be necessary?
		StatusCondition: &statusCondition,
	}
	err = s.directorClient.UpdateRuntime(cluster.ID, runtimeInput, cluster.Tenant)
	if err != nil {
		logger.Errorf("Failed to update Director with Runtime status INSTALLING: %s", err.Error())
	}

	logger.Warnf("Installation started, proceeding to next step...")
	return operations.StageResult{Stage: s.nextStep, Delay: 30 * time.Second}, nil
}
