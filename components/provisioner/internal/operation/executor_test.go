package operation

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

const (
	operationId = "operation-id"
	clusterId   = "cluster-id"
)

func TestStagesExecutor_Execute(t *testing.T) {

	// TODO: add more tests

	tNow := time.Now()

	operation := model.Operation{
		ID:             operationId,
		Type:           model.Provision,
		StartTimestamp: tNow,
		EndTimestamp:   nil,
		State:          model.InProgress,
		ClusterID:      clusterId,
		Stage:          model.WaitingForInstallation,
		LastTransition: &tNow,
	}

	cluster := model.Cluster{ID: clusterId}

	t.Run("should not requeue operation when stage if Finished", func(t *testing.T) {
		// given
		dbSession := &mocks.ReadWriteSession{}
		dbSession.On("GetOperation", operationId).Return(operation, nil)
		dbSession.On("GetCluster", clusterId).Return(cluster, nil)
		dbSession.On("TransitionOperation", operationId, "Provisioning steps finished", model.FinishedStage, mock.AnythingOfType("time.Time")).
			Return(nil)
		dbSession.On("UpdateOperationState", operationId, "Operation succeeded", model.Succeeded, mock.AnythingOfType("time.Time")).
			Return(nil)

		mockStage := NewMockStep(model.WaitingForInstallation, model.FinishedStage, 10*time.Second)

		installationStages := map[model.OperationStage]Stage{
			model.WaitingForInstallation: mockStage,
		}

		executor := NewStepsExecutor(dbSession, model.Provision, installationStages)

		// when
		result := executor.Execute(operationId)

		// then
		assert.Equal(t, false, result.Requeue)
		assert.True(t, mockStage.called)
	})

}

type mockStep struct {
	name  model.OperationStage
	next  model.OperationStage
	delay time.Duration

	called bool
}

func NewMockStep(name, next model.OperationStage, delay time.Duration) *mockStep {
	return &mockStep{
		name:  name,
		next:  next,
		delay: delay,
	}
}

func (m mockStep) Name() model.OperationStage {
	return m.name
}

func (m *mockStep) Run(cluster model.Cluster, logger logrus.FieldLogger) (StageResult, error) {

	m.called = true

	return StageResult{
		Stage: m.next,
		Delay: m.delay,
	}, nil
}

func (m mockStep) TimeLimit() time.Duration {
	return 10 * time.Second
}
