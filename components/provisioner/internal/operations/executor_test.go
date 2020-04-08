package operations

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/operations/failure"
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

		mockStage := NewMockStep(model.WaitingForInstallation, model.FinishedStage, 10*time.Second, 10*time.Second)

		installationStages := map[model.OperationStage]Stage{
			model.WaitingForInstallation: mockStage,
		}

		executor := NewStepsExecutor(dbSession, model.Provision, installationStages, failure.NewNoopFailureHandler())

		// when
		result := executor.Execute(operationId)

		// then
		assert.Equal(t, false, result.Requeue)
		assert.True(t, mockStage.called)
	})

	t.Run("should requeue operation if error occured", func(t *testing.T) {
		// given
		dbSession := &mocks.ReadWriteSession{}
		dbSession.On("GetOperation", operationId).Return(operation, nil)
		dbSession.On("GetCluster", clusterId).Return(cluster, nil)

		mockStage := NewErrorStep(model.ShootProvisioning, fmt.Errorf("error"), time.Second*10)

		installationStages := map[model.OperationStage]Stage{
			model.WaitingForInstallation: mockStage,
		}

		executor := NewStepsExecutor(dbSession, model.Provision, installationStages, failure.NewNoopFailureHandler())

		// when
		result := executor.Execute(operationId)

		// then
		assert.Equal(t, true, result.Requeue)
		assert.True(t, mockStage.called)
	})

	t.Run("should not requeue operation and run failure handler if NonRecoverable error occurred", func(t *testing.T) {
		// given
		dbSession := &mocks.ReadWriteSession{}
		dbSession.On("GetOperation", operationId).Return(operation, nil)
		dbSession.On("GetCluster", clusterId).Return(cluster, nil)
		dbSession.On("UpdateOperationState", operationId, "error", model.Failed, mock.AnythingOfType("time.Time")).
			Return(nil)

		mockStage := NewErrorStep(model.ShootProvisioning, NewNonRecoverableError(fmt.Errorf("error")), 10*time.Second)

		installationStages := map[model.OperationStage]Stage{
			model.WaitingForInstallation: mockStage,
		}

		failureHandler := MockFailureHandler{}
		executor := NewStepsExecutor(dbSession, model.Provision, installationStages, &failureHandler)

		// when
		result := executor.Execute(operationId)

		// then
		assert.Equal(t, false, result.Requeue)
		assert.True(t, mockStage.called)
		assert.True(t, failureHandler.called)
	})

	t.Run("should not requeue operation and run failure handler if timeout reached", func(t *testing.T) {
		// given
		dbSession := &mocks.ReadWriteSession{}
		dbSession.On("GetOperation", operationId).Return(operation, nil)
		dbSession.On("GetCluster", clusterId).Return(cluster, nil)
		dbSession.On("TransitionOperation", operationId, "Operation in progress", model.ConnectRuntimeAgent, mock.AnythingOfType("time.Time")).
			Return(nil)
		dbSession.On("UpdateOperationState", operationId, "error: timeout while processing operation", model.Failed, mock.AnythingOfType("time.Time")).
			Return(nil)

		mockStage := NewMockStep(model.WaitingForInstallation, model.ConnectRuntimeAgent, 0, 0*time.Second)

		installationStages := map[model.OperationStage]Stage{
			model.WaitingForInstallation: mockStage,
		}

		failureHandler := MockFailureHandler{}
		executor := NewStepsExecutor(dbSession, model.Provision, installationStages, &failureHandler)

		// when
		result := executor.Execute(operationId)

		// then
		assert.Equal(t, false, result.Requeue)
		assert.False(t, mockStage.called)
		assert.True(t, failureHandler.called)
	})
}

type mockStep struct {
	name      model.OperationStage
	next      model.OperationStage
	delay     time.Duration
	timeLimit time.Duration
	err       error

	called bool
}

func NewMockStep(name, next model.OperationStage, delay time.Duration, timeLimit time.Duration) *mockStep {
	return &mockStep{
		name:      name,
		next:      next,
		delay:     delay,
		timeLimit: timeLimit,
	}
}

func NewErrorStep(name model.OperationStage, err error, timeLimit time.Duration) *mockStep {
	return &mockStep{
		name:      name,
		err:       err,
		timeLimit: timeLimit,
	}
}

func (m mockStep) Name() model.OperationStage {
	return m.name
}

func (m *mockStep) Run(cluster model.Cluster, logger logrus.FieldLogger) (StageResult, error) {

	m.called = true

	if m.err != nil {
		return StageResult{}, m.err
	}

	return StageResult{
		Stage: m.next,
		Delay: m.delay,
	}, nil
}

func (m mockStep) TimeLimit() time.Duration {
	return m.timeLimit
}

type MockFailureHandler struct {
	called bool
}

func (m *MockFailureHandler) HandleFailure(operation model.Operation, cluster model.Cluster) error {
	m.called = true
	return nil
}
