package deprovisioning

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util/k8s"
	"github.com/sirupsen/logrus"
)

type CleanupClusterStep struct {
	resourceClient installation.ResourceClient
	nextStep       model.OperationStage
	timeLimit      time.Duration
}

func NewCleanupClusterStep(nextStep model.OperationStage, timeLimit time.Duration) *CleanupClusterStep {
	return &CleanupClusterStep{
		nextStep:  nextStep,
		timeLimit: timeLimit,
	}
}

func (s *CleanupClusterStep) Name() model.OperationStage {
	return model.CleanupCluster
}

func (s *CleanupClusterStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *CleanupClusterStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {
	logger.Infof("Starting cleanup cluster step for %s ...", cluster.ID)
	if cluster.Kubeconfig == nil {
		logger.Errorf("got nil kubeconfig while trying to cleanup cluster")
		// Kubeconfig can be nil if Gardener failed to create cluster. We must go to the next step to finalize deprovisioning
		return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
	}

	k8sConfig, err := k8s.ParseToK8sConfig([]byte(*cluster.Kubeconfig))
	if err != nil {
		err := fmt.Errorf("error: failed to create kubernetes config from raw: %s", err.Error())
		return operations.StageResult{}, operations.NewNonRecoverableError(err)
	}

	cli, err := installation.NewServiceCatalogClient(k8sConfig)
	if err != nil {
		logger.Errorf("error creating Service Catalog Client for cluster id %q: %s", cluster.ID, err.Error())
		return operations.StageResult{}, err
	}

	err = cli.PerformCleanup()
	if err != nil {
		logger.Errorf("while performing resource cleanup for cluster id: %q: %s", cluster.ID, err.Error())
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}
