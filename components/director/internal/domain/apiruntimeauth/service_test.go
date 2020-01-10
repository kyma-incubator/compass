package apiruntimeauth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apiruntimeauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apiruntimeauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Get(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	apiID := "foo"
	rtmID := "bar"

	modelAPIRtmAuth := fixModelAPIRuntimeAuth(str.Ptr("foo"), rtmID, apiID, fixModelAuth())

	testErr := errors.New("test error")

	testCases := []struct {
		Name             string
		apiRtmAuthRepoFn func() *automock.Repository
		ExpectedOutput   *model.APIRuntimeAuth
		ExpectedError    error
	}{
		{
			Name: "Success",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("Get", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(modelAPIRtmAuth, nil).Once()
				return apiRtmAuthRepo
			},
			ExpectedOutput: modelAPIRtmAuth,
			ExpectedError:  nil,
		},
		{
			Name: "Error when getting api runtime auth",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("Get", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(nil, testErr).Once()
				return apiRtmAuthRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			apiRtmAuthRepo := testCase.apiRtmAuthRepoFn()

			svc := apiruntimeauth.NewService(apiRtmAuthRepo, nil)

			// WHEN
			result, err := svc.Get(ctx, apiID, rtmID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			apiRtmAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := apiruntimeauth.NewService(nil, nil)

		// WHEN
		_, err := svc.Get(context.TODO(), "", "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetOrDefault(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	apiID := "foo"
	rtmID := "bar"

	modelAPIRtmAuth := fixModelAPIRuntimeAuth(str.Ptr("foo"), rtmID, apiID, fixModelAuth())

	testErr := errors.New("test error")

	testCases := []struct {
		Name             string
		apiRtmAuthRepoFn func() *automock.Repository
		ExpectedOutput   *model.APIRuntimeAuth
		ExpectedError    error
	}{
		{
			Name: "Success",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("GetOrDefault", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(modelAPIRtmAuth, nil).Once()
				return apiRtmAuthRepo
			},
			ExpectedOutput: modelAPIRtmAuth,
			ExpectedError:  nil,
		},
		{
			Name: "Error when getting api runtime auth",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("GetOrDefault", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(nil, testErr).Once()
				return apiRtmAuthRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			apiRtmAuthRepo := testCase.apiRtmAuthRepoFn()

			svc := apiruntimeauth.NewService(apiRtmAuthRepo, nil)

			// WHEN
			result, err := svc.GetOrDefault(ctx, apiID, rtmID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			apiRtmAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := apiruntimeauth.NewService(nil, nil)

		// WHEN
		_, err := svc.GetOrDefault(context.TODO(), "", "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListForAllRuntimes(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	apiID := "foo"

	modelAPIRtmAuths := []model.APIRuntimeAuth{
		*fixModelAPIRuntimeAuth(str.Ptr("foo"), "1", apiID, fixModelAuth()),
		*fixModelAPIRuntimeAuth(str.Ptr("bar"), "2", apiID, fixModelAuth()),
		*fixModelAPIRuntimeAuth(str.Ptr("baz"), "3", apiID, fixModelAuth()),
	}

	testErr := errors.New("test error")

	testCases := []struct {
		Name             string
		apiRtmAuthRepoFn func() *automock.Repository
		ExpectedOutput   []model.APIRuntimeAuth
		ExpectedError    error
	}{
		{
			Name: "Success",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("ListForAllRuntimes", contextThatHasTenant(tnt), tnt, apiID).Return(modelAPIRtmAuths, nil).Once()
				return apiRtmAuthRepo
			},
			ExpectedOutput: modelAPIRtmAuths,
			ExpectedError:  nil,
		},
		{
			Name: "Error when listing api runtime auths",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("ListForAllRuntimes", contextThatHasTenant(tnt), tnt, apiID).Return(nil, testErr).Once()
				return apiRtmAuthRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			apiRtmAuthRepo := testCase.apiRtmAuthRepoFn()

			svc := apiruntimeauth.NewService(apiRtmAuthRepo, nil)

			// WHEN
			result, err := svc.ListForAllRuntimes(ctx, apiID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			apiRtmAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := apiruntimeauth.NewService(nil, nil)

		// WHEN
		_, err := svc.ListForAllRuntimes(context.TODO(), apiID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Set(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	apiID := "foo"
	rtmID := "bar"
	apiRtmAuthID := "baz"

	modelAuthInput := fixModelAuthInput()
	modelAPIRtmAuth := fixModelAPIRuntimeAuth(&apiRtmAuthID, rtmID, apiID, fixModelAuth())

	testErr := errors.New("test error")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(apiRtmAuthID).Once()
		return uidSvc
	}

	testCases := []struct {
		Name             string
		apiRtmAuthRepoFn func() *automock.Repository
		ExpectedError    error
	}{
		{
			Name: "Success",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("Upsert", contextThatHasTenant(tnt), *modelAPIRtmAuth).Return(nil).Once()
				return apiRtmAuthRepo
			},
			ExpectedError: nil,
		},
		{
			Name: "Error when getting api runtime auth",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("Upsert", contextThatHasTenant(tnt), *modelAPIRtmAuth).Return(testErr).Once()
				return apiRtmAuthRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			apiRtmAuthRepo := testCase.apiRtmAuthRepoFn()
			uidSvc := uidSvcFn()

			svc := apiruntimeauth.NewService(apiRtmAuthRepo, uidSvc)

			// WHEN
			err := svc.Set(ctx, apiID, rtmID, modelAuthInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			apiRtmAuthRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := apiruntimeauth.NewService(nil, nil)

		// WHEN
		err := svc.Set(context.TODO(), "", "", model.AuthInput{})

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

	apiID := "foo"
	rtmID := "bar"

	testErr := errors.New("test error")

	testCases := []struct {
		Name             string
		apiRtmAuthRepoFn func() *automock.Repository
		ExpectedError    error
	}{
		{
			Name: "Success",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("Delete", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(nil).Once()
				return apiRtmAuthRepo
			},
			ExpectedError: nil,
		},
		{
			Name: "Error when getting api runtime auth",
			apiRtmAuthRepoFn: func() *automock.Repository {
				apiRtmAuthRepo := &automock.Repository{}
				apiRtmAuthRepo.On("Delete", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(testErr).Once()
				return apiRtmAuthRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			apiRtmAuthRepo := testCase.apiRtmAuthRepoFn()

			svc := apiruntimeauth.NewService(apiRtmAuthRepo, nil)

			// WHEN
			err := svc.Delete(ctx, apiID, rtmID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			apiRtmAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := apiruntimeauth.NewService(nil, nil)

		// WHEN
		err := svc.Delete(context.TODO(), "", "")

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
