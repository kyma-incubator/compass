package operation_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	opInput := fixOperationInput(ordOpType, scheduledOpStatus)
	opModel := fixOperationModel(ordOpType, scheduledOpStatus)
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

	opInput := fixOperationInput(ordOpType, scheduledOpStatus)
	opModel := fixOperationModel(ordOpType, scheduledOpStatus)

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

func TestService_DeleteOlderThan(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	ctx := context.TODO()

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.OperationRepository
		OpType       string
		OpStatus     string
		Days         int
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("DeleteOlderThan", ctx, ordOpType, scheduledOpStatus, mock.AnythingOfType("Time")).Return(nil).Once()
				return repo
			},
			OpType:   ordOpType,
			OpStatus: scheduledOpStatus,
			Days:     1,
		},
		{
			Name: "Error while deleting operations",
			RepositoryFn: func() *automock.OperationRepository {
				repo := &automock.OperationRepository{}
				repo.On("DeleteOlderThan", ctx, ordOpType, scheduledOpStatus, mock.AnythingOfType("Time")).Return(testErr).Once()
				return repo
			},
			OpType:      ordOpType,
			OpStatus:    scheduledOpStatus,
			Days:        1,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()

			svc := operation.NewService(repo, nil)

			// WHEN
			err := svc.DeleteOlderThan(ctx, testCase.OpType, testCase.OpStatus, testCase.Days)

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
