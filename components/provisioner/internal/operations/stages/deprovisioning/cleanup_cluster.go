package deprovisioning

import (
	"fmt"
	"time"

	"github.com/kyma-project/control-plane/components/provisioner/internal/installation"
	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	"github.com/kyma-project/control-plane/components/provisioner/internal/operations"
	"github.com/kyma-project/control-plane/components/provisioner/internal/util/k8s"
	"github.com/sirupsen/logrus"
)

type CleanupClusterStep struct {
	installationService installation.Service
	nextStep            model.OperationStage
	timeLimit           time.Duration
}

func NewCleanupClusterStep(installationService installation.Service, nextStep model.OperationStage, timeLimit time.Duration) *CleanupClusterStep {
	return &CleanupClusterStep{
		installationService: installationService,
		nextStep:            nextStep,
		timeLimit:           timeLimit,
	}
}

func (s *CleanupClusterStep) Name() model.OperationStage {
	return model.CleanupCluster
}

func (s *CleanupClusterStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *CleanupClusterStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {
	logger.Debugf("Starting cleanup cluster step for %s ...", cluster.ID)
	if cluster.Kubeconfig == nil {
		// Kubeconfig can be nil if Gardener failed to create cluster. We must go to the next step to finalize deprovisioning
		return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
	}

	k8sConfig, err := k8s.ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		err := fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	err = s.installationService.PerformCleanup(k8sConfig)
	if err != nil {
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}
