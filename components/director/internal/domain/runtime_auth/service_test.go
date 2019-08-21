package runtime_auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_auth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/strings"

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

	modelRtmAuth := fixModelRuntimeAuth(strings.Ptr("foo"), rtmID, apiID, fixModelAuth())

	testErr := errors.New("test error")

	testCases := []struct {
		Name           string
		rtmAuthRepoFn  func() *automock.Repository
		ExpectedOutput *model.RuntimeAuth
		ExpectedError  error
	}{
		{
			Name: "Success",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("Get", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(modelRtmAuth, nil).Once()
				return rtmAuthRepo
			},
			ExpectedOutput: modelRtmAuth,
			ExpectedError:  nil,
		},
		{
			Name: "Error when getting runtime auth",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("Get", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(nil, testErr).Once()
				return rtmAuthRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			rtmAuthRepo := testCase.rtmAuthRepoFn()

			svc := runtime_auth.NewService(rtmAuthRepo, nil)

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

			rtmAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := runtime_auth.NewService(nil, nil)

		// WHEN
		_, err := svc.Get(context.TODO(), "", "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
	})
}

func TestService_GetOrDefault(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	apiID := "foo"
	rtmID := "bar"

	modelRtmAuth := fixModelRuntimeAuth(strings.Ptr("foo"), rtmID, apiID, fixModelAuth())

	testErr := errors.New("test error")

	testCases := []struct {
		Name           string
		rtmAuthRepoFn  func() *automock.Repository
		ExpectedOutput *model.RuntimeAuth
		ExpectedError  error
	}{
		{
			Name: "Success",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("GetOrDefault", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(modelRtmAuth, nil).Once()
				return rtmAuthRepo
			},
			ExpectedOutput: modelRtmAuth,
			ExpectedError:  nil,
		},
		{
			Name: "Error when getting runtime auth",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("GetOrDefault", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(nil, testErr).Once()
				return rtmAuthRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			rtmAuthRepo := testCase.rtmAuthRepoFn()

			svc := runtime_auth.NewService(rtmAuthRepo, nil)

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

			rtmAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := runtime_auth.NewService(nil, nil)

		// WHEN
		_, err := svc.GetOrDefault(context.TODO(), "", "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
	})
}

func TestService_ListForAllRuntimes(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	apiID := "foo"

	modelRtmAuths := []model.RuntimeAuth{
		*fixModelRuntimeAuth(strings.Ptr("foo"), "1", apiID, fixModelAuth()),
		*fixModelRuntimeAuth(strings.Ptr("bar"), "2", apiID, fixModelAuth()),
		*fixModelRuntimeAuth(strings.Ptr("baz"), "3", apiID, fixModelAuth()),
	}

	testErr := errors.New("test error")

	testCases := []struct {
		Name           string
		rtmAuthRepoFn  func() *automock.Repository
		ExpectedOutput []model.RuntimeAuth
		ExpectedError  error
	}{
		{
			Name: "Success",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("ListForAllRuntimes", contextThatHasTenant(tnt), tnt, apiID).Return(modelRtmAuths, nil).Once()
				return rtmAuthRepo
			},
			ExpectedOutput: modelRtmAuths,
			ExpectedError:  nil,
		},
		{
			Name: "Error when listing runtime auths",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("ListForAllRuntimes", contextThatHasTenant(tnt), tnt, apiID).Return(nil, testErr).Once()
				return rtmAuthRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			rtmAuthRepo := testCase.rtmAuthRepoFn()

			svc := runtime_auth.NewService(rtmAuthRepo, nil)

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

			rtmAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := runtime_auth.NewService(nil, nil)

		// WHEN
		_, err := svc.ListForAllRuntimes(context.TODO(), apiID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
	})
}

func TestService_Set(t *testing.T) {
	// GIVEN
	tnt := testTenant
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	apiID := "foo"
	rtmID := "bar"
	rtmAuthID := "baz"

	modelAuthInput := fixModelAuthInput()
	modelRtmAuth := fixModelRuntimeAuth(&rtmAuthID, rtmID, apiID, fixModelAuth())

	testErr := errors.New("test error")

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(rtmAuthID).Once()
		return uidSvc
	}

	testCases := []struct {
		Name          string
		rtmAuthRepoFn func() *automock.Repository
		ExpectedError error
	}{
		{
			Name: "Success",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("Upsert", contextThatHasTenant(tnt), *modelRtmAuth).Return(nil).Once()
				return rtmAuthRepo
			},
			ExpectedError: nil,
		},
		{
			Name: "Error when getting runtime auth",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("Upsert", contextThatHasTenant(tnt), *modelRtmAuth).Return(testErr).Once()
				return rtmAuthRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			rtmAuthRepo := testCase.rtmAuthRepoFn()
			uidSvc := uidSvcFn()

			svc := runtime_auth.NewService(rtmAuthRepo, uidSvc)

			// WHEN
			err := svc.Set(ctx, apiID, rtmID, modelAuthInput)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			rtmAuthRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := runtime_auth.NewService(nil, nil)

		// WHEN
		err := svc.Set(context.TODO(), "", "", model.AuthInput{})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
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
		Name          string
		rtmAuthRepoFn func() *automock.Repository
		ExpectedError error
	}{
		{
			Name: "Success",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("Delete", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(nil).Once()
				return rtmAuthRepo
			},
			ExpectedError: nil,
		},
		{
			Name: "Error when getting runtime auth",
			rtmAuthRepoFn: func() *automock.Repository {
				rtmAuthRepo := &automock.Repository{}
				rtmAuthRepo.On("Delete", contextThatHasTenant(tnt), tnt, apiID, rtmID).Return(testErr).Once()
				return rtmAuthRepo
			},
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			rtmAuthRepo := testCase.rtmAuthRepoFn()

			svc := runtime_auth.NewService(rtmAuthRepo, nil)

			// WHEN
			err := svc.Delete(ctx, apiID, rtmID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			rtmAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := runtime_auth.NewService(nil, nil)

		// WHEN
		err := svc.Delete(context.TODO(), "", "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
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
