package operation_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opInput := fixOperationInput(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	opModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.OperationInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Create", ctx, opModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(operationID)
				return svc
			},
			Input: *opInput,
		},
		{
			Name: "Error - Operation creation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Create", ctx, opModel).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(operationID)
				return svc
			},
			Input:       *opInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			uidService := testCase.UIDServiceFn()

			svc := operation.NewService(repo, uidService)

			// WHEN
			_, err := svc.Create(ctx, &testCase.Input)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo, uidService)
		})
	}
}

func TestService_CreateMultiple(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opInput := fixOperationInput(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	opModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)

	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		UIDServiceFn func() *automock.UIDService
		Input        []*model.OperationInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Create", ctx, opModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(operationID)
				return svc
			},
			Input: []*model.OperationInput{opInput},
		},
		{
			Name: "Success - nil operation input",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Create", ctx, opModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(operationID)
				return svc
			},
			Input: []*model.OperationInput{opInput, nil},
		},
		{
			Name: "Nil operation inputs",
			RepositoryFn: func() *automock.OperationRepository {
				return &automock.OperationRepository{}
			},
			UIDServiceFn: func() *automock.UIDService {
				return &automock.UIDService{}
			},
			Input: nil,
		},
		{
			Name: "Error - Operation creation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Create", ctx, opModel).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(operationID)
				return svc
			},
			Input:       []*model.OperationInput{opInput},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			uidService := testCase.UIDServiceFn()

			svc := operation.NewService(repo, uidService)

			// WHEN
			err := svc.CreateMultiple(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo, uidService)
		})
	}
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		Input        *model.Operation
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Update", ctx, opModel).Return(nil).Once()
				return repo
			},
			Input: opModel,
		},
		{
			Name: "Error - Operation update",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Update", ctx, opModel).Return(testErr).Once()
				return repo
			},
			Input:       opModel,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.Update(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_RescheduleOperation(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opModelWithLowPriority := fixOperationModelWithPriority(testOpType, model.OperationStatusCompleted, lowOperationPriority, model.OperationErrorSeverityNone)
	opModelWithHighPriority := fixOperationModelWithPriority(testOpType, model.OperationStatusScheduled, highOperationPriority, model.OperationErrorSeverityNone)
	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		Input        string
		Priority     int
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModelWithLowPriority, nil).Once()
				repo.On("Update", ctx, mock.AnythingOfType("*model.Operation")).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*model.Operation)
					assert.Equal(t, model.OperationStatusScheduled, arg.Status)
					assert.Equal(t, highOperationPriority, arg.Priority)
					assert.Equal(t, testOpType, arg.OpType)
					assert.Equal(t, json.RawMessage(errorMsg), arg.Error)
				})
				return repo
			},
			Input:    operationID,
			Priority: highOperationPriority,
		},
		{
			Name: "Error while getting operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModelWithHighPriority, testErr).Once()
				return repo
			},
			Input:       operationID,
			Priority:    lowOperationPriority,
			ExpectedErr: testErr,
		},
		{
			Name: "Error while updating operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModelWithLowPriority, nil).Once()
				repo.On("Update", ctx, mock.AnythingOfType("*model.Operation")).Return(testErr).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*model.Operation)
					assert.Equal(t, model.OperationStatusScheduled, arg.Status)
					assert.Equal(t, highOperationPriority, arg.Priority)
					assert.Equal(t, testOpType, arg.OpType)
					assert.Equal(t, json.RawMessage(errorMsg), arg.Error)
				})
				return repo
			},
			Input:       operationID,
			Priority:    highOperationPriority,
			ExpectedErr: testErr,
		},
		{
			Name: "Error while trying to reschedule operation that is in IN_PROGRESS state",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				opModelInProgress := fixOperationModelWithPriority(testOpType, model.OperationStatusInProgress, lowOperationPriority, model.OperationErrorSeverityNone)
				repo.On("Get", ctx, operationID).Return(opModelInProgress, nil).Once()
				return repo
			},
			Input:       operationID,
			Priority:    highOperationPriority,
			ExpectedErr: apperrors.NewOperationInProgressError(operationID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.RescheduleOperation(ctx, testCase.Input, testCase.Priority)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_Get(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	ctx := context.TODO()

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.OperationRepository
		Input          string
		ExpectedErr    error
		ExpectedOutput *model.Operation
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModel, nil).Once()
				return repo
			},
			Input:          operationID,
			ExpectedOutput: opModel,
		},
		{
			Name: "Error while getting operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(nil, testErr).Once()
				return repo
			},
			Input:       operationID,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			op, err := svc.Get(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedOutput, op)
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_GetByDataAndType(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	ctx := context.TODO()

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.OperationRepository
		Input          string
		ExpectedErr    error
		ExpectedOutput *model.Operation
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("GetByDataAndType", ctx, fixOperationDataAsString(applicationID, applicationTemplateID), model.OperationTypeOrdAggregation).Return(opModel, nil).Once()
				return repo
			},
			Input:          fixOperationDataAsString(applicationID, applicationTemplateID),
			ExpectedOutput: opModel,
		},
		{
			Name: "Error while getting operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("GetByDataAndType", ctx, fixOperationDataAsString(applicationID, applicationTemplateID), model.OperationTypeOrdAggregation).Return(nil, testErr).Once()
				return repo
			},
			Input:       fixOperationDataAsString(applicationID, applicationTemplateID),
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			op, err := svc.GetByDataAndType(ctx, testCase.Input, model.OperationTypeOrdAggregation)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedOutput, op)
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_ListPriorityQueue(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	operationModels := []*model.Operation{opModel}
	ctx := context.TODO()

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.OperationRepository
		OpType         model.OperationType
		QueueLimit     int
		ExpectedErr    error
		ExpectedOutput []*model.Operation
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("PriorityQueueListByType", ctx, queueLimit, model.OperationTypeOrdAggregation).Return(operationModels, nil).Once()
				return repo
			},
			QueueLimit:     queueLimit,
			OpType:         model.OperationTypeOrdAggregation,
			ExpectedOutput: operationModels,
		},
		{
			Name: "Error while listing priority queue",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("PriorityQueueListByType", ctx, queueLimit, model.OperationTypeOrdAggregation).Return(nil, testErr).Once()
				return repo
			},
			QueueLimit:  queueLimit,
			OpType:      model.OperationTypeOrdAggregation,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			op, err := svc.ListPriorityQueue(ctx, testCase.QueueLimit, testCase.OpType)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedOutput, op)
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_LockOperation(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.OperationRepository
		Input          string
		ExpectedErr    error
		ExpectedOutput bool
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("LockOperation", ctx, operationID).Return(true, nil).Once()
				return repo
			},
			Input:          operationID,
			ExpectedOutput: true,
		},
		{
			Name: "Error while locking operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("LockOperation", ctx, operationID).Return(false, testErr).Once()
				return repo
			},
			Input:       operationID,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			isLocked, err := svc.LockOperation(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedOutput, isLocked)
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_RescheduleOperations(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		Input        time.Duration
		Type         model.OperationType
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("RescheduleOperations", ctx, model.OperationTypeOrdAggregation, time.Minute, []string{model.OperationStatusCompleted.ToString(), model.OperationStatusFailed.ToString()}).Return(nil).Once()
				return repo
			},
			Type:  model.OperationTypeOrdAggregation,
			Input: time.Minute,
		},
		{
			Name: "Error while rescheduling operations",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("RescheduleOperations", ctx, model.OperationTypeOrdAggregation, time.Minute, []string{model.OperationStatusCompleted.ToString(), model.OperationStatusFailed.ToString()}).Return(testErr).Once()
				return repo
			},
			Type:        model.OperationTypeOrdAggregation,
			Input:       time.Minute,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.RescheduleOperations(ctx, testCase.Type, testCase.Input, []string{model.OperationStatusCompleted.ToString(), model.OperationStatusFailed.ToString()})

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_DeleteOperations(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		Input        time.Duration
		Type         model.OperationType
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("DeleteOperations", ctx, model.OperationTypeOrdAggregation, time.Minute).Return(nil).Once()
				return repo
			},
			Type:  model.OperationTypeOrdAggregation,
			Input: time.Minute,
		},
		{
			Name: "Error while rescheduling operations",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("DeleteOperations", ctx, model.OperationTypeOrdAggregation, time.Minute).Return(testErr).Once()
				return repo
			},
			Type:        model.OperationTypeOrdAggregation,
			Input:       time.Minute,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.DeleteOperations(ctx, testCase.Type, testCase.Input)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_RescheduleHangedOperations(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		Type         model.OperationType
		Input        time.Duration
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("RescheduleHangedOperations", ctx, model.OperationTypeOrdAggregation, time.Minute).Return(nil).Once()
				return repo
			},
			Type:  model.OperationTypeOrdAggregation,
			Input: time.Minute,
		},
		{
			Name: "Error while rescheduling operations",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("RescheduleHangedOperations", ctx, model.OperationTypeOrdAggregation, time.Minute).Return(testErr).Once()
				return repo
			},
			Type:        model.OperationTypeOrdAggregation,
			Input:       time.Minute,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.RescheduleHangedOperations(ctx, testCase.Type, testCase.Input)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_MarkAsCompleted(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		Input        string
		ErrorMsg     error
		ExpectedErr  error
	}{
		{
			Name: "Success when there is no error message",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModel, nil).Once()
				repo.On("Update", ctx, mock.AnythingOfType("*model.Operation")).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*model.Operation)
					assert.Equal(t, model.OperationStatusCompleted, arg.Status)
					assert.Equal(t, json.RawMessage("{}"), arg.Error)
				})
				return repo
			},
			Input:    operationID,
			ErrorMsg: nil,
		},
		{
			Name: "Success when there is an error message",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModel, nil).Once()
				repo.On("Update", ctx, mock.AnythingOfType("*model.Operation")).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*model.Operation)
					assert.Equal(t, model.OperationStatusCompleted, arg.Status)
					assert.Equal(t, json.RawMessage(`{"error":{"message":{"validation_errors":null,"runtime_error":{"message":"err"}}}}`), arg.Error)
				})
				return repo
			},
			Input:    operationID,
			ErrorMsg: &ord.ProcessingError{RuntimeError: &ord.RuntimeError{Message: "err"}},
		},
		{
			Name: "Error - Getting operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(nil, testErr).Once()
				return repo
			},
			Input:       operationID,
			ErrorMsg:    nil,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Updating operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModel, nil).Once()
				repo.On("Update", ctx, mock.Anything).Return(testErr).Once()
				return repo
			},
			Input:       operationID,
			ErrorMsg:    nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.MarkAsCompleted(ctx, testCase.Input, testCase.ErrorMsg)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_MarkAsFailed(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		Input        string
		InputErr     error
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModel, nil).Once()
				repo.On("Update", ctx, mock.AnythingOfType("*model.Operation")).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*model.Operation)
					assert.Equal(t, model.OperationStatusFailed, arg.Status)
					opErr := operation.NewOperationError(errors.New(operationErrMsg))
					expectedMsg, err := opErr.ToJSONRawMessage()
					require.NoError(t, err)
					assert.Equal(t, expectedMsg, arg.Error)
				})
				return repo
			},
			Input:    operationID,
			InputErr: errors.New(operationErrMsg),
		},
		{
			Name: "Error - Getting operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(nil, testErr).Once()
				return repo
			},
			Input:       operationID,
			InputErr:    errors.New(operationErrMsg),
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Updating operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModel, nil).Once()
				repo.On("Update", ctx, mock.Anything).Return(testErr).Once()
				return repo
			},
			Input:       operationID,
			InputErr:    errors.New(operationErrMsg),
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.MarkAsFailed(ctx, testCase.Input, testCase.InputErr)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_SetErrorSeverity(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

	testCases := []struct {
		Name          string
		RepositoryFn  func() *automock.OperationRepository
		ErrorSeverity model.OperationErrorSeverity
		ExpectedErr   error
	}{
		{
			Name: "Success when set Error severity to Info",
			RepositoryFn: func() *automock.OperationRepository {
				opModelWithErrorSeverityNone := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityNone)
				opModelWithErrorSeverityInfo := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityInfo)

				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModelWithErrorSeverityNone, nil).Once()
				repo.On("Update", ctx, opModelWithErrorSeverityInfo).Return(nil).Once()
				return repo
			},
			ErrorSeverity: model.OperationErrorSeverityInfo,
		},
		{
			Name: "Success when set Error severity to Warning",
			RepositoryFn: func() *automock.OperationRepository {
				opModelWithErrorSeverityNone := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityNone)
				opModelWithErrorSeverityWarning := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityWarning)

				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModelWithErrorSeverityNone, nil).Once()
				repo.On("Update", ctx, opModelWithErrorSeverityWarning).Return(nil).Once()
				return repo
			},
			ErrorSeverity: model.OperationErrorSeverityWarning,
		},
		{
			Name: "Success when set Error severity to Error",
			RepositoryFn: func() *automock.OperationRepository {
				opModelWithErrorSeverityNone := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityNone)
				opModelWithErrorSeverityError := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityError)

				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModelWithErrorSeverityNone, nil).Once()
				repo.On("Update", ctx, opModelWithErrorSeverityError).Return(nil).Once()
				return repo
			},
			ErrorSeverity: model.OperationErrorSeverityError,
		},
		{
			Name: "Success when set Error severity to None",
			RepositoryFn: func() *automock.OperationRepository {
				opModelWithErrorSeverityError := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityError)
				opModelWithErrorSeverityNone := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityNone)

				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModelWithErrorSeverityError, nil).Once()
				repo.On("Update", ctx, opModelWithErrorSeverityNone).Return(nil).Once()
				return repo
			},
			ErrorSeverity: model.OperationErrorSeverityNone,
		},
		{
			Name: "Error while getting operation",
			RepositoryFn: func() *automock.OperationRepository {
				opModelWithErrorSeverityNone := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityNone)

				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModelWithErrorSeverityNone, testErr).Once()
				return repo
			},
			ErrorSeverity: model.OperationErrorSeverityInfo,
			ExpectedErr:   testErr,
		},
		{
			Name: "Error while updating operation",
			RepositoryFn: func() *automock.OperationRepository {
				opModelWithErrorSeverityNone := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityNone)
				opModelWithErrorSeverityInfo := fixOperationModelWithErrorSeverity(model.OperationErrorSeverityInfo)

				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModelWithErrorSeverityNone, nil).Once()
				repo.On("Update", ctx, opModelWithErrorSeverityInfo).Return(testErr).Once()

				return repo
			},
			ErrorSeverity: model.OperationErrorSeverityInfo,
			ExpectedErr:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.SetErrorSeverity(ctx, operationID, testCase.ErrorSeverity)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_ListAllByType(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opModel := fixOperationModel(testOpType, model.OperationStatusScheduled, model.OperationErrorSeverityNone)
	operationModels := []*model.Operation{opModel}
	ctx := context.TODO()

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.OperationRepository
		OpType         model.OperationType
		ExpectedErr    error
		ExpectedOutput []*model.Operation
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("ListAllByType", ctx, model.OperationTypeOrdAggregation).Return(operationModels, nil).Once()
				return repo
			},
			OpType:         model.OperationTypeOrdAggregation,
			ExpectedOutput: operationModels,
		},
		{
			Name: "Error while listing by type",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("ListAllByType", ctx, model.OperationTypeOrdAggregation).Return(nil, testErr).Once()
				return repo
			},
			OpType:      model.OperationTypeOrdAggregation,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			op, err := svc.ListAllByType(ctx, testCase.OpType)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedOutput, op)
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}
