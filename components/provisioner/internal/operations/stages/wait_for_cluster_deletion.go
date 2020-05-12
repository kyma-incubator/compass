package stages

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
)

type WaitForClusterDeletionStep struct {
	installationClient installation.Service
	nextStep           model.OperationStage
	timeLimit          time.Duration
}

func NewWaitForClusterDeletionStep(installationClient installation.Service, nextStep model.OperationStage, timeLimit time.Duration) *DeleteClusterStep {
	return &StartClusterDeletionStep{
		installationClient: installationClient,
		nextStep:           nextStep,
		timeLimit:          timeLimit,
	}
}

func (s *WaitForClusterDeletionStep) Name() model.OperationStage {
	return model.StartingInstallation
}

func (s *WaitForClusterDeletionStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *WaitForClusterDeletionStep) Run(cluster model.Cluster, _ model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	return operations.StageResult{}, nil
}
