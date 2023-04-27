package tenantbusinesstype_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantbusinesstype"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantbusinesstype/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

	tbtModel := fixModelTenantBusinessType(tbtID, tbtCode, tbtName)
	tbtModelInput := fixModelTenantBusinessTypeInput(tbtCode, tbtName)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.TenantBusinessTypeRepository
		UIDServiceFn func() *automock.UIDService
		Input        *model.TenantBusinessTypeInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.TenantBusinessTypeRepository {
				repo := &automock.TenantBusinessTypeRepository{}
				repo.On("Create", ctx, tbtModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(tbtID)
				return svc
			},
			Input:       tbtModelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when creating tenant business type failed",
			RepositoryFn: func() *automock.TenantBusinessTypeRepository {
				repo := &automock.TenantBusinessTypeRepository{}
				repo.On("Create", ctx, tbtModel).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(tbtID)
				return svc
			},
			Input:       tbtModelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			repo := testCase.RepositoryFn()
			uidSvc := testCase.UIDServiceFn()

			svc := tenantbusinesstype.NewService(repo, uidSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			mock.AssertExpectationsForObjects(t, repo)
		})
	}
}

func TestService_GetByID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	ctx := context.TODO()

	tbtModel := fixModelTenantBusinessType(tbtID, tbtCode, tbtName)

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.TenantBusinessTypeRepository
		Input          *model.TenantBusinessTypeInput
		InputID        string
		ExpectedResult *model.TenantBusinessType
		ExpectedErr    error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.TenantBusinessTypeRepository {
				repo := &automock.TenantBusinessTypeRepository{}
				repo.On("GetByID", ctx, tbtID).Return(tbtModel, nil).Once()
				return repo
			},
			InputID:        tbtID,
			ExpectedResult: tbtModel,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when tenant business type retrieval failed",
			RepositoryFn: func() *automock.TenantBusinessTypeRepository {
				repo := &automock.TenantBusinessTypeRepository{}
				repo.On("GetByID", ctx, tbtID).Return(nil, testErr).Once()
				return repo
			},
			InputID:        tbtID,
			ExpectedResult: tbtModel,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := tenantbusinesstype.NewService(repo, nil)

			// WHEN
			result, err := svc.GetByID(ctx, testCase.InputID)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_ListAll(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	tbts := []*model.TenantBusinessType{
		fixModelTenantBusinessType(tbtID, tbtCode, tbtName),
		fixModelTenantBusinessType("tbt-id-2", "test-code-2", "tbt-name-2"),
		fixModelTenantBusinessType("tbt-id-3", "test-code-3", "tbt-name-3"),
	}

	ctx := context.TODO()

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.TenantBusinessTypeRepository
		ExpectedResult []*model.TenantBusinessType
		ExpectedErr    error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.TenantBusinessTypeRepository {
				repo := &automock.TenantBusinessTypeRepository{}
				repo.On("ListAll", ctx).Return(tbts, nil).Once()
				return repo
			},
			ExpectedResult: tbts,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when listing all tenant business types failed",
			RepositoryFn: func() *automock.TenantBusinessTypeRepository {
				repo := &automock.TenantBusinessTypeRepository{}
				repo.On("ListAll", ctx).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := tenantbusinesstype.NewService(repo, nil)

			// WHEN
			result, err := svc.ListAll(ctx)

			// THEN
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, result)
			}

			repo.AssertExpectations(t)
		})
	}
}
