package operation_test

import (
	"context"
	"testing"

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

	opInput := fixOperationInput(ordOpType, model.OperationStatusScheduled)
	opModel := fixOperationModel(ordOpType, model.OperationStatusScheduled)
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
			err := svc.Create(ctx, &testCase.Input)

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

	opInput := fixOperationInput(ordOpType, model.OperationStatusScheduled)
	opModel := fixOperationModel(ordOpType, model.OperationStatusScheduled)

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

func TestService_MarkAsCompleted(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opModel := fixOperationModel(ordOpType, model.OperationStatusScheduled)
	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		Input        string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(opModel, nil).Once()
				repo.On("Update", ctx, mock.AnythingOfType("*model.Operation")).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*model.Operation)
					assert.Equal(t, model.OperationStatusCompleted, arg.Status)
				})
				return repo
			},
			Input: operationID,
		},
		{
			Name: "Error - Getting operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(nil, testErr).Once()
				return repo
			},
			Input:       operationID,
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
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.MarkAsCompleted(ctx, testCase.Input)

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

	opModel := fixOperationModel(ordOpType, model.OperationStatusScheduled)
	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		Input        string
		InputErrMsg  string
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
					opErr := operation.NewOperationError(operationErrMsg)
					expectedMsg, err := opErr.ToJSONRawMessage()
					require.NoError(t, err)
					assert.Equal(t, expectedMsg, arg.Error)
				})
				return repo
			},
			Input:       operationID,
			InputErrMsg: operationErrMsg,
		},
		{
			Name: "Error - Getting operation",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("Get", ctx, operationID).Return(nil, testErr).Once()
				return repo
			},
			Input:       operationID,
			InputErrMsg: operationErrMsg,
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
			InputErrMsg: operationErrMsg,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.MarkAsFailed(ctx, testCase.Input, testCase.InputErrMsg)

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
