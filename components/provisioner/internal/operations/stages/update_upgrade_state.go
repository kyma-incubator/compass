package stages

import (
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations"
	"github.com/sirupsen/logrus"
)

type UpdateUpgradeStateStep struct {
	session   dbsession.WriteSession
	nextStep  model.OperationStage
	timeLimit time.Duration
}

func NewUpdateUpgradeStateStep(dbSession dbsession.WriteSession, nextStep model.OperationStage, timeLimit time.Duration) *UpdateUpgradeStateStep {
	return &UpdateUpgradeStateStep{
		session:   dbSession,
		nextStep:  nextStep,
		timeLimit: timeLimit,
	}
}

func (s *UpdateUpgradeStateStep) Name() model.OperationStage {
	return model.UpdatingUpgradeState
}

func (s *UpdateUpgradeStateStep) TimeLimit() time.Duration {
	return s.timeLimit
}

func (s *UpdateUpgradeStateStep) Run(_ model.Cluster, operation model.Operation, logger logrus.FieldLogger) (operations.StageResult, error) {
	dberr := s.session.UpdateUpgradeState(operation.ID, model.UpgradeSucceeded)

	if dberr != nil {
		return operations.StageResult{}, dberr
	}

	logger.Warn("Upgrade state updated. Proceeding to next step...")
	return operations.StageResult{Stage: s.nextStep, Delay: 0}, nil
}
