package deprovisioning

import (
	"fmt"
	"time"

	"github.com/kyma-project/control-plane/components/provisioner/internal/util/k8s"

	"github.com/kyma-project/control-plane/components/provisioner/internal/installation"
	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	"github.com/kyma-project/control-plane/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
)

type TriggerKymaUninstallStep struct {
	installationClient installation.Service
	nextStep           model.OperationStage
	timeLimit          time.Duration
	delay              time.Duration
}

func NewTriggerKymaUninstallStep(installationClient installation.Service, nextStep model.OperationStage, timeLimit time.Duration, delay time.Duration) *TriggerKymaUninstallStep {
	return &TriggerKymaUninstallStep{
		installationClient: installationClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
		delay:              delay,
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
		// Kubeconfig can be nil if Gardener failed to create cluster. We must go to the next step to finalize deprovisioning
		return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
	}

	k8sConfig, err := k8s.ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		err := fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	err = s.installationClient.TriggerUninstall(k8sConfig)
	if err != nil {
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: s.delay}, nil
}
