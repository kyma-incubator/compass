package packageinstanceauth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Get(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	id := "foo"

	modelInstanceAuth := fixSimpleModelPackageInstanceAuth(id)

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

func TestService_GetForPackage(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	id := "foo"
	packageID := "bar"

	modelInstanceAuth := fixSimpleModelPackageInstanceAuth(id)

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
				instanceAuthRepo.On("GetForPackage", contextThatHasTenant(tnt), tnt, id, packageID).Return(modelInstanceAuth, nil).Once()
				return instanceAuthRepo
			},
			ExpectedOutput: modelInstanceAuth,
			ExpectedError:  nil,
		},
		{
			Name: "Error",
			instanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetForPackage", contextThatHasTenant(tnt), tnt, id, packageID).Return(nil, testErr).Once()
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
			result, err := svc.GetForPackage(ctx, id, packageID)

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
		_, err := svc.GetForPackage(context.TODO(), id, packageID)

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

func TestService_ListByApplicationID(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	tnt := testTenant

	packageInstanceAuths := []*model.PackageInstanceAuth{
		fixSimpleModelPackageInstanceAuth(id),
		fixSimpleModelPackageInstanceAuth(id),
		fixSimpleModelPackageInstanceAuth(id),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.Repository
		ExpectedResult     []*model.PackageInstanceAuth
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListByPackageID", ctx, tnt, id).Return(packageInstanceAuths, nil).Once()
				return repo
			},
			ExpectedResult:     packageInstanceAuths,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Package Instance Auth listing failed",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListByPackageID", ctx, tnt, id).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := packageinstanceauth.NewService(repo, nil)

			// when
			pia, err := svc.List(ctx, id)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, pia)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := packageinstanceauth.NewService(nil, nil)
		// WHEN
		_, err := svc.List(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_RequestDeletion(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	id := "foo"
	timestampNow := time.Now()
	pkgInstanceAuth := fixSimpleModelPackageInstanceAuth(id)
	//testErr := errors.New("test error")

	testCases := []struct {
		Name                       string
		PackageDefaultInstanceAuth *model.Auth
		InstanceAuthRepoFn         func() *automock.Repository

		ExpectedResult bool
		ExpectedError  error
	}{
		{
			Name: "Success - No Package Default Instance Auth",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Update", contextThatHasTenant(tnt), mock.MatchedBy(func(in *model.PackageInstanceAuth) bool {
					return in.ID == id && in.Status.Condition == model.PackageInstanceAuthStatusConditionUnused
				})).Return(nil).Once()
				return instanceAuthRepo
			},
			ExpectedResult: false,
			ExpectedError:  nil,
		},
		{
			Name:                       "Success - Package Default Instance Auth",
			PackageDefaultInstanceAuth: fixModelAuth(),
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Delete", contextThatHasTenant(tnt), tnt, id).Return(nil).Once()
				return instanceAuthRepo
			},
			ExpectedResult: true,
			ExpectedError:  nil,
		},
		{
			Name: "Error - Update",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Update", contextThatHasTenant(tnt), mock.MatchedBy(func(in *model.PackageInstanceAuth) bool {
					return in.ID == id && in.Status.Condition == model.PackageInstanceAuthStatusConditionUnused
				})).Return(testError).Once()
				return instanceAuthRepo
			},
			ExpectedError: testError,
		},
		{
			Name:                       "Error - Delete",
			PackageDefaultInstanceAuth: fixModelAuth(),
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Delete", contextThatHasTenant(tnt), tnt, id).Return(testError).Once()
				return instanceAuthRepo
			},
			ExpectedError: testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			instanceAuthRepo := testCase.InstanceAuthRepoFn()

			svc := packageinstanceauth.NewService(instanceAuthRepo, nil)
			svc.SetTimestampGen(func() time.Time {
				return timestampNow
			})

			// WHEN
			res, err := svc.RequestDeletion(ctx, pkgInstanceAuth, testCase.PackageDefaultInstanceAuth)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, res)
			}

			instanceAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error - nil", func(t *testing.T) {
		// GIVEN
		expectedError := errors.New("instance auth is required to request its deletion")

		// WHEN
		svc := packageinstanceauth.NewService(nil, nil)
		_, err := svc.RequestDeletion(ctx, nil, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
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
