package deprovisioning

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util/k8s"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
)

type TriggerKymaUninstallStep struct {
	installationClient installation.Service
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

func NewTriggerKymaUninstallStep(installationClient installation.Service, nextStep model.OperationStage, timeLimit time.Duration) *TriggerKymaUninstallStep {
	return &TriggerKymaUninstallStep{
		installationClient: installationClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
	}
}

func (s *TriggerKymaUninstallStep) Name() model.OperationStage {
	return model.TriggerKymaUninstall
}

func (s *TriggerKymaUninstallStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *TriggerKymaUninstallStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	if cluster.Kubeconfig == nil {
		err := fmt.Errorf("error: kubeconfig is nil")

		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	k8sConfig, err := k8s.ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		err := fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	err = s.installationClient.TriggerUninstall(k8sConfig)
	if err != nil {
		logger.Errorf("error triggering uninstalling: %s", err.Error())
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 5 * time.Second}, nil
}
