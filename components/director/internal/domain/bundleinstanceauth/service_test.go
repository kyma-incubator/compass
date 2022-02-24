package bundleinstanceauth_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Get(t *testing.T) {
	// GIVEN
	tnt := testTenant
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	id := "foo"

	modelInstanceAuth := fixSimpleModelBundleInstanceAuth(id)

	testErr := errors.New("test error")

	testCases := []struct {
		Name               string
		instanceAuthRepoFn func() *automock.Repository
		ExpectedOutput     *model.BundleInstanceAuth
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

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil)

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
		svc := bundleinstanceauth.NewService(nil, nil)

		// WHEN
		_, err := svc.Get(context.TODO(), id)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetForBundle(t *testing.T) {
	// GIVEN
	tnt := testTenant
	externalTnt := testExternalTenant

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	id := "foo"
	bundleID := "bar"

	modelInstanceAuth := fixSimpleModelBundleInstanceAuth(id)

	testErr := errors.New("test error")

	testCases := []struct {
		Name               string
		instanceAuthRepoFn func() *automock.Repository
		ExpectedOutput     *model.BundleInstanceAuth
		ExpectedError      error
	}{
		{
			Name: "Success",
			instanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetForBundle", contextThatHasTenant(tnt), tnt, id, bundleID).Return(modelInstanceAuth, nil).Once()
				return instanceAuthRepo
			},
			ExpectedOutput: modelInstanceAuth,
			ExpectedError:  nil,
		},
		{
			Name: "Error",
			instanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetForBundle", contextThatHasTenant(tnt), tnt, id, bundleID).Return(nil, testErr).Once()
				return instanceAuthRepo
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			instanceAuthRepo := testCase.instanceAuthRepoFn()

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil)

			// WHEN
			result, err := svc.GetForBundle(ctx, id, bundleID)

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
		svc := bundleinstanceauth.NewService(nil, nil)

		// WHEN
		_, err := svc.GetForBundle(context.TODO(), id, bundleID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// GIVEN
	tnt := testTenant
	externalTnt := testExternalTenant

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

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

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil)

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
		svc := bundleinstanceauth.NewService(nil, nil)

		// WHEN
		err := svc.Delete(context.TODO(), id)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_SetAuth(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	modelInstanceAuthFn := func() *model.BundleInstanceAuth {
		return fixModelBundleInstanceAuth(testID, testBundleID, testTenant, nil, fixModelStatusPending(), nil)
	}

	modelSetInput := fixModelSetInput()
	modelUpdatedInstanceAuth := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, nil, fixModelStatusPending(), nil)
	modelUpdatedInstanceAuth.Auth = modelSetInput.Auth.ToAuth()
	modelUpdatedInstanceAuth.Status = &model.BundleInstanceAuthStatus{
		Condition: model.BundleInstanceAuthStatusConditionSucceeded,
		Timestamp: testTime,
		Message:   modelSetInput.Status.Message,
		Reason:    modelSetInput.Status.Reason,
	}

	modelSetInputWithoutStatus := model.BundleInstanceAuthSetInput{
		Auth:   fixModelAuthInput(),
		Status: nil,
	}
	modelUpdatedInstanceAuthWithDefaultStatus := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, nil, fixModelStatusPending(), nil)
	modelUpdatedInstanceAuthWithDefaultStatus.Auth = modelSetInputWithoutStatus.Auth.ToAuth()
	err := modelUpdatedInstanceAuthWithDefaultStatus.SetDefaultStatus(model.BundleInstanceAuthStatusConditionSucceeded, testTime)
	require.NoError(t, err)

	testCases := []struct {
		Name               string
		InstanceAuthRepoFn func() *automock.Repository
		Input              model.BundleInstanceAuthSetInput
		ExpectedError      error
	}{
		{
			Name: "Success",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(modelInstanceAuthFn(), nil).Once()
				instanceAuthRepo.On("Update", contextThatHasTenant(testTenant), testTenant, modelUpdatedInstanceAuth).Return(nil).Once()
				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: nil,
		},
		{
			Name: "Success when new status not provided",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(modelInstanceAuthFn(), nil).Once()
				instanceAuthRepo.On("Update", contextThatHasTenant(testTenant), testTenant, modelUpdatedInstanceAuthWithDefaultStatus).Return(nil).Once()
				return instanceAuthRepo
			},
			Input:         modelSetInputWithoutStatus,
			ExpectedError: nil,
		},
		{
			Name: "Error when Bundle Instance Auth retrieval failed",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(modelInstanceAuthFn(), testError).Once()
				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: testError,
		},
		{
			Name: "Error when Bundle Instance Auth update failed",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(modelInstanceAuthFn(), nil).Once()
				instanceAuthRepo.On("Update", contextThatHasTenant(testTenant), testTenant, modelUpdatedInstanceAuth).Return(testError).Once()
				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: testError,
		},
		{
			Name: "Error when Bundle Instance Auth status is nil",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(
					&model.BundleInstanceAuth{
						Status: nil,
					}, nil).Once()
				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: errors.New("auth can be set only on BundleInstanceAuths in PENDING state"),
		},
		{
			Name: "Error when Bundle Instance Auth status condition different from PENDING",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(
					&model.BundleInstanceAuth{
						Status: &model.BundleInstanceAuthStatus{
							Condition: model.BundleInstanceAuthStatusConditionSucceeded,
						},
					}, nil).Once()
				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: errors.New("auth can be set only on BundleInstanceAuths in PENDING state"),
		},
		{
			Name: "Error when retrieved Bundle Instance Auth is nil",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(nil, nil).Once()

				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: errors.Errorf("BundleInstanceAuth with id %s not found", testID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			instanceAuthRepo := testCase.InstanceAuthRepoFn()

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil)
			svc.SetTimestampGen(func() time.Time { return testTime })

			// WHEN
			err := svc.SetAuth(ctx, testID, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, instanceAuthRepo)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := bundleinstanceauth.NewService(nil, nil)

		// WHEN
		err := svc.SetAuth(context.TODO(), testID, model.BundleInstanceAuthSetInput{})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	consumerEntity := consumer.Consumer{
		ConsumerID:   testRuntimeID,
		ConsumerType: consumer.Runtime,
	}
	ctx = consumer.SaveToContext(ctx, consumerEntity)

	modelAuth := fixModelAuth()
	modelExpectedInstanceAuth := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, modelAuth, fixModelStatusSucceeded(), &testRuntimeID)
	modelExpectedInstanceAuthPending := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, nil, fixModelStatusPending(), &testRuntimeID)

	modelRequestInput := fixModelRequestInput()

	testCases := []struct {
		Name               string
		InstanceAuthRepoFn func() *automock.Repository
		UIDSvcFn           func() *automock.UIDService
		Input              model.BundleInstanceAuthRequestInput
		InputAuth          *model.Auth
		InputSchema        *string
		ExpectedOutput     string
		ExpectedError      error
	}{
		{
			Name: "Success",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Create", contextThatHasTenant(testTenant), modelExpectedInstanceAuth).Return(nil).Once()
				return instanceAuthRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				svc := automock.UIDService{}
				svc.On("Generate").Return(testID).Once()
				return &svc
			},
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    nil,
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name: "Success when input auth is nil",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Create", contextThatHasTenant(testTenant), modelExpectedInstanceAuthPending).Return(nil).Once()
				return instanceAuthRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				svc := automock.UIDService{}
				svc.On("Generate").Return(testID).Once()
				return &svc
			},
			Input:          *modelRequestInput,
			InputAuth:      nil,
			InputSchema:    nil,
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name: "Success when schema provided",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Create", contextThatHasTenant(testTenant), modelExpectedInstanceAuth).Return(nil).Once()
				return instanceAuthRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				svc := automock.UIDService{}
				svc.On("Generate").Return(testID).Once()
				return &svc
			},
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    str.Ptr("{\"type\": \"object\"}"),
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name: "Error when creating Bundle Instance Auth",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Create", contextThatHasTenant(testTenant), modelExpectedInstanceAuth).Return(testError).Once()
				return instanceAuthRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				svc := automock.UIDService{}
				svc.On("Generate").Return(testID).Once()
				return &svc
			},
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    nil,
			ExpectedOutput: "",
			ExpectedError:  testError,
		},
		{
			Name: "Error when schema defined but no input params provided",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				return instanceAuthRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				svc := automock.UIDService{}
				return &svc
			},
			Input:          model.BundleInstanceAuthRequestInput{},
			InputAuth:      modelAuth,
			InputSchema:    str.Ptr("{\"type\": \"string\"}"),
			ExpectedOutput: "",
			ExpectedError:  errors.New("json schema for input parameters was defined for the bundle but no input parameters were provided"),
		},
		{
			Name: "Error when invalid schema provided",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				return instanceAuthRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				svc := automock.UIDService{}
				return &svc
			},
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    str.Ptr("error"),
			ExpectedOutput: "",
			ExpectedError:  errors.New("while creating JSON Schema validator for schema error: invalid character 'e' looking for beginning of value"),
		},
		{
			Name: "Error when invalid input params",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				return instanceAuthRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				svc := automock.UIDService{}
				return &svc
			},
			Input: model.BundleInstanceAuthRequestInput{
				InputParams: str.Ptr("{"),
			},
			InputAuth:      modelAuth,
			InputSchema:    str.Ptr("{\"type\": \"string\"}"),
			ExpectedOutput: "",
			ExpectedError:  errors.New(`while validating value { against JSON Schema: {"type": "string"}: unexpected EOF`),
		},
		{
			Name: "Error when input doesn't match schema",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				return instanceAuthRepo
			},
			UIDSvcFn: func() *automock.UIDService {
				svc := automock.UIDService{}
				return &svc
			},
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    str.Ptr("{\"type\": \"string\"}"),
			ExpectedOutput: "",
			ExpectedError:  errors.New(`while validating value {"bar": "baz"} against JSON Schema: {"type": "string"}: (root): Invalid type. Expected: string, given: object`),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			instanceAuthRepo := testCase.InstanceAuthRepoFn()
			uidSvc := testCase.UIDSvcFn()

			svc := bundleinstanceauth.NewService(instanceAuthRepo, uidSvc)
			svc.SetTimestampGen(func() time.Time { return testTime })

			// WHEN
			result, err := svc.Create(ctx, testBundleID, testCase.Input, testCase.InputAuth, testCase.InputSchema)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, instanceAuthRepo, uidSvc)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := bundleinstanceauth.NewService(nil, nil)

		// WHEN
		_, err := svc.Create(context.TODO(), testBundleID, model.BundleInstanceAuthRequestInput{}, nil, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})

	t.Run("Error when consumer is not in the context", func(t *testing.T) {
		// GIVEN
		svc := bundleinstanceauth.NewService(nil, nil)
		ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

		// WHEN
		_, err := svc.Create(ctx, testBundleID, *modelRequestInput, modelAuth, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read consumer from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	tnt := testTenant
	externalTnt := testExternalTenant

	bundleInstanceAuths := []*model.BundleInstanceAuth{
		fixSimpleModelBundleInstanceAuth(testBundleID),
		fixSimpleModelBundleInstanceAuth(testBundleID),
		fixSimpleModelBundleInstanceAuth(testBundleID),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.Repository
		ExpectedResult     []*model.BundleInstanceAuth
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListByBundleID", ctx, tnt, testBundleID).Return(bundleInstanceAuths, nil).Once()
				return repo
			},
			ExpectedResult:     bundleInstanceAuths,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Bundle Instance Auth listing failed",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListByBundleID", ctx, tnt, testBundleID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := bundleinstanceauth.NewService(repo, nil)

			// WHEN
			pia, err := svc.List(ctx, testBundleID)

			// THEN
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
		svc := bundleinstanceauth.NewService(nil, nil)
		// WHEN
		_, err := svc.List(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByRuntimeID(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	tnt := testTenant
	externalTnt := testExternalTenant

	bundleInstanceAuths := []*model.BundleInstanceAuth{
		fixSimpleModelBundleInstanceAuth(testBundleID),
		fixSimpleModelBundleInstanceAuth(testBundleID),
		fixSimpleModelBundleInstanceAuth(testBundleID),
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.Repository
		ExpectedResult     []*model.BundleInstanceAuth
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListByRuntimeID", ctx, tnt, testRuntimeID).Return(bundleInstanceAuths, nil).Once()
				return repo
			},
			ExpectedResult:     bundleInstanceAuths,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Bundle Instance Auth listing by runtime ID failed",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListByRuntimeID", ctx, tnt, testRuntimeID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := bundleinstanceauth.NewService(repo, nil)

			// WHEN
			bundleInstanceAuth, err := svc.ListByRuntimeID(ctx, testRuntimeID)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, bundleInstanceAuth)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := bundleinstanceauth.NewService(nil, nil)

		// WHEN
		_, err := svc.ListByRuntimeID(context.TODO(), "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")

	tnt := testTenant
	externalTnt := testExternalTenant

	bundleInstanceAuth := fixSimpleModelBundleInstanceAuth(testBundleID)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.Repository
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Update", ctx, testTenant, bundleInstanceAuth).Return(nil).Once()
				return repo
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when bundle instance auth update failed",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Update", ctx, testTenant, bundleInstanceAuth).Return(testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := bundleinstanceauth.NewService(repo, nil)

			// WHEN
			err := svc.Update(ctx, bundleInstanceAuth)

			// THEN
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_RequestDeletion(t *testing.T) {
	// GIVEN
	tnt := testTenant
	externalTnt := testExternalTenant

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	id := "foo"
	timestampNow := time.Now()
	bndlInstanceAuth := fixSimpleModelBundleInstanceAuth(id)
	//testErr := errors.New("test error")

	testCases := []struct {
		Name                      string
		BundleDefaultInstanceAuth *model.Auth
		InstanceAuthRepoFn        func() *automock.Repository

		ExpectedResult bool
		ExpectedError  error
	}{
		{
			Name: "Success - No Bundle Default Instance Auth",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("Update", contextThatHasTenant(tnt), tnt, mock.MatchedBy(func(in *model.BundleInstanceAuth) bool {
					return in.ID == id && in.Status.Condition == model.BundleInstanceAuthStatusConditionUnused
				})).Return(nil).Once()
				return instanceAuthRepo
			},
			ExpectedResult: false,
			ExpectedError:  nil,
		},
		{
			Name:                      "Success - Bundle Default Instance Auth",
			BundleDefaultInstanceAuth: fixModelAuth(),
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
				instanceAuthRepo.On("Update", contextThatHasTenant(tnt), tnt, mock.MatchedBy(func(in *model.BundleInstanceAuth) bool {
					return in.ID == id && in.Status.Condition == model.BundleInstanceAuthStatusConditionUnused
				})).Return(testError).Once()
				return instanceAuthRepo
			},
			ExpectedError: testError,
		},
		{
			Name:                      "Error - Delete",
			BundleDefaultInstanceAuth: fixModelAuth(),
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

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil)
			svc.SetTimestampGen(func() time.Time {
				return timestampNow
			})

			// WHEN
			res, err := svc.RequestDeletion(ctx, bndlInstanceAuth, testCase.BundleDefaultInstanceAuth)

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
		expectedError := errors.New("BundleInstanceAuth is required to request its deletion")

		// WHEN
		svc := bundleinstanceauth.NewService(nil, nil)
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
