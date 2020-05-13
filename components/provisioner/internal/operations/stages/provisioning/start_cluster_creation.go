package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
)

type StartClusterCreationStep struct {
	provisioner GardenerProvisioner
	nextStep    model.OperationStage
	timeLimit   time.Duration
}

type TimeWindow struct {
	Begin string
	End   string
}

type GardenerProvisioner interface {
	CreateAndApplyShoot(cluster model.Cluster, operationId string) error
}

func NewStartClusterCreationStep(provisioner GardenerProvisioner, nextStep model.OperationStage, timeLimit time.Duration) *StartClusterCreationStep {
	return &StartClusterCreationStep{
		provisioner: provisioner,
		nextStep:    nextStep,
		timeLimit:   timeLimit,
	}
}

func (s *StartClusterCreationStep) Name() model.OperationStage {
	return model.StartingInstallation
}

func (s *StartClusterCreationStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (tw TimeWindow) isEmpty() bool {
	return tw.Begin == "" || tw.End == ""
}

func (s *StartClusterCreationStep) Run(cluster model.Cluster, operation model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {

	err := s.provisioner.CreateAndApplyShoot(cluster, operation.ID)
	if err != nil {
		return operations.StageResult{}, err
	}

	return operations.StageResult{Stage: s.nextStep, Delay: 30 * time.Second}, nil
}
