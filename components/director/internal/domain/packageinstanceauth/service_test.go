package packageinstanceauth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testTenant = "tenant"
)

func TestService_Get(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	id := "foo"

	modelInstanceAuth := fixModelPackageInstanceAuth(id)

	testErr := errors.New("test error")

	testCases := []struct {
		Name               string
		instanceAuthRepoFn func() *automock.Repository
		ExpectedOutput     *model.PackageInstanceAuth
		ExpectedError      error
	}{
		{
			Name: "Success",
			instanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(tnt), tnt, id).Return(modelInstanceAuth, nil).Once()
				return instanceAuthRepo
			},
			ExpectedOutput: modelInstanceAuth,
			ExpectedError:  nil,
		},
		{
			Name: "Error",
			instanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(tnt), tnt, id).Return(nil, testErr).Once()
				return instanceAuthRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			instanceAuthRepo := testCase.instanceAuthRepoFn()

			svc := packageinstanceauth.NewService(instanceAuthRepo, nil)

			// WHEN
			result, err := svc.Get(ctx, id)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			instanceAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := packageinstanceauth.NewService(nil, nil)

		// WHEN
		_, err := svc.Get(context.TODO(), id)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	id := "foo"

	testErr := errors.New("test error")

	testCases := []struct {
		Name               string
		instanceAuthRepoFn func() *automock.Repository
		ExpectedError      error
	}{
		{
			Name: "Success",
			instanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Delete", contextThatHasTenant(tnt), tnt, id).Return(nil).Once()
				return instanceAuthRepo
			},
			ExpectedError: nil,
		},
		{
			Name: "Error",
			instanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Delete", contextThatHasTenant(tnt), tnt, id).Return(testErr).Once()
				return instanceAuthRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			instanceAuthRepo := testCase.instanceAuthRepoFn()

			svc := packageinstanceauth.NewService(instanceAuthRepo, nil)

			// WHEN
			err := svc.Delete(ctx, id)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			instanceAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := packageinstanceauth.NewService(nil, nil)

		// WHEN
		err := svc.Delete(context.TODO(), id)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func contextThatHasTenant(expectedTenant string) interface{} {
	return mock.MatchedBy(func(actual context.Context) bool {
		actualTenant, err := tenant.LoadFromContext(actual)
		if err != nil {
			return false
		}
		return actualTenant == expectedTenant
	})
}

func fixModelPackageInstanceAuth(id string) *model.PackageInstanceAuth {
	return &model.PackageInstanceAuth{
		ID: id,
	}
}
