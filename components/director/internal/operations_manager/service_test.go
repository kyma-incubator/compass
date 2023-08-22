package operationsmanager_test

import (
	"context"
	"testing"

	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/internal/operations_manager/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateORDOperations(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	testErr := errors.New("Test error")

	testCases := []struct {
		Name        string
		OpCreatorFn func() *automock.OperationCreator
		ExpectedErr error
	}{
		{
			Name: "Success",
			OpCreatorFn: func() *automock.OperationCreator {
				creator := &automock.OperationCreator{}
				creator.On("Create", ctx).Return(nil).Once()
				return creator
			},
		},
		{
			Name: "Error while creating ord operations",
			OpCreatorFn: func() *automock.OperationCreator {
				creator := &automock.OperationCreator{}
				creator.On("Create", ctx).Return(testErr).Once()
				return creator
			},
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			opCreator := testCase.OpCreatorFn()
			svc := operationsmanager.NewOperationService(nil, nil, opCreator)

			// WHEN
			err := svc.CreateORDOperations(ctx)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Nil(t, err)
			}

			mock.AssertExpectationsForObjects(t, opCreator)
		})
	}
}
