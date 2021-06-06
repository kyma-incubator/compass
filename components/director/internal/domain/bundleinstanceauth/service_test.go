package bundleinstanceauth_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth/automock"
	labelpkg "github.com/kyma-incubator/compass/components/director/internal/domain/label"
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

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil, nil, nil, nil)

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
		svc := bundleinstanceauth.NewService(nil, nil, nil, nil, nil)

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

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil, nil, nil, nil)

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
		svc := bundleinstanceauth.NewService(nil, nil, nil, nil, nil)

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

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil, nil, nil, nil)

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
		svc := bundleinstanceauth.NewService(nil, nil, nil, nil, nil)

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
				instanceAuthRepo.On("Update", contextThatHasTenant(testTenant), modelUpdatedInstanceAuth).Return(nil).Once()
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
				instanceAuthRepo.On("Update", contextThatHasTenant(testTenant), modelUpdatedInstanceAuthWithDefaultStatus).Return(nil).Once()
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
				instanceAuthRepo.On("Update", contextThatHasTenant(testTenant), modelUpdatedInstanceAuth).Return(testError).Once()
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

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil, nil, nil, nil)
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
		svc := bundleinstanceauth.NewService(nil, nil, nil, nil, nil)

		// WHEN
		err := svc.SetAuth(context.TODO(), testID, model.BundleInstanceAuthSetInput{})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Create(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
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

	appRuntimeWithCommmonScenarios := appAndRuntimeScenarios{
		appScenarios:     []string{"scenario-1", "scenario-2", "scenario-3"},
		runtimeScenarios: []string{"scenario-2", "scenario-3", "scenario-4"},
	}

	appRuntimeWithNoCommmonScenarios := appAndRuntimeScenarios{
		appScenarios:     []string{"scenario-1", "scenario-2"},
		runtimeScenarios: []string{"scenario-3", "scenario-4"},
	}

	testCases := []struct {
		Name               string
		InstanceAuthRepoFn func() *automock.Repository
		UIDSvcFn           func() *automock.UIDService
		BundleSvcFn        func() *automock.BundleService
		ScenarioSvcFn      func() *automock.ScenarioService
		LabelSvcFn         func() *automock.LabelService
		Input              model.BundleInstanceAuthRequestInput
		InputAuth          *model.Auth
		InputSchema        *string
		ExpectedOutput     string
		ExpectedError      error
	}{
		{
			Name: "Success when there is matching scenarios between application and runtime",
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
			BundleSvcFn:    newBundleSvcThatGetBundleById,
			ScenarioSvcFn:  newScenarioSvcFn(&appRuntimeWithCommmonScenarios),
			LabelSvcFn:     newLabelSvcFnThatSucceeds(&appRuntimeWithCommmonScenarios),
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    nil,
			ExpectedOutput: testID,
			ExpectedError:  nil,
		},
		{
			Name: "Success - When no matching scenarios between application and runtime Then no bundle instance auth scenario association is performed",
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
			BundleSvcFn:    newBundleSvcThatGetBundleById,
			ScenarioSvcFn:  newScenarioSvcFn(&appRuntimeWithNoCommmonScenarios),
			LabelSvcFn:     unusedLabelService,
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
			BundleSvcFn:    newBundleSvcThatGetBundleById,
			ScenarioSvcFn:  newScenarioSvcFn(&appRuntimeWithCommmonScenarios),
			LabelSvcFn:     newLabelSvcFnThatSucceeds(&appRuntimeWithCommmonScenarios),
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
			BundleSvcFn:    newBundleSvcThatGetBundleById,
			ScenarioSvcFn:  newScenarioSvcFn(&appRuntimeWithCommmonScenarios),
			LabelSvcFn:     newLabelSvcFnThatSucceeds(&appRuntimeWithCommmonScenarios),
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
			BundleSvcFn:    unusedBundleService,
			ScenarioSvcFn:  unusedScenarioService,
			LabelSvcFn:     unusedLabelService,
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
			BundleSvcFn:    unusedBundleService,
			ScenarioSvcFn:  unusedScenarioService,
			LabelSvcFn:     unusedLabelService,
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
			BundleSvcFn:    unusedBundleService,
			ScenarioSvcFn:  unusedScenarioService,
			LabelSvcFn:     unusedLabelService,
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
			BundleSvcFn:   unusedBundleService,
			ScenarioSvcFn: unusedScenarioService,
			LabelSvcFn:    unusedLabelService,
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
			BundleSvcFn:    unusedBundleService,
			ScenarioSvcFn:  unusedScenarioService,
			LabelSvcFn:     unusedLabelService,
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    str.Ptr("{\"type\": \"string\"}"),
			ExpectedOutput: "",
			ExpectedError:  errors.New(`while validating value {"bar": "baz"} against JSON Schema: {"type": "string"}: (root): Invalid type. Expected: string, given: object`),
		},
		{
			Name: "Error while fetching application id for bundle",
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
			BundleSvcFn: func() *automock.BundleService {
				svc := automock.BundleService{}
				svc.On("Get", contextThatHasTenant(testTenant), testBundleID).Return(nil, testErr).Once()
				return &svc
			},
			ScenarioSvcFn:  unusedScenarioService,
			LabelSvcFn:     unusedLabelService,
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    nil,
			ExpectedOutput: "",
			ExpectedError:  errors.New("while fetching application id"),
		},
		{
			Name: "Error when fetching scenario names for application",
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
			BundleSvcFn: newBundleSvcThatGetBundleById,
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := automock.ScenarioService{}
				svc.On("GetScenarioNamesForApplication", contextThatHasTenant(testTenant), testApplicationID).Return(nil, testErr)
				return &svc
			},
			LabelSvcFn:     unusedLabelService,
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    nil,
			ExpectedOutput: "",
			ExpectedError:  errors.Errorf("while fetching scenario names for application: %s", testApplicationID),
		},
		{
			Name: "Error when fetching scenario names for runtime",
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
			BundleSvcFn: newBundleSvcThatGetBundleById,
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := automock.ScenarioService{}
				svc.On("GetScenarioNamesForApplication", contextThatHasTenant(testTenant), testApplicationID).Return(appRuntimeWithCommmonScenarios.appScenarios, nil)
				svc.On("GetScenarioNamesForRuntime", contextThatHasTenant(testTenant), testRuntimeID).Return(nil, testErr)
				return &svc
			},
			LabelSvcFn:     unusedLabelService,
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    nil,
			ExpectedOutput: "",
			ExpectedError:  errors.Errorf("while fetching scenario names for runtime: %s", testRuntimeID),
		},
		{
			Name: "Error when creating bundle instance auth scenario labels",
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
			BundleSvcFn:    newBundleSvcThatGetBundleById,
			ScenarioSvcFn:  newScenarioSvcFn(&appRuntimeWithCommmonScenarios),
			LabelSvcFn:     newLabelSvcFnThatFail(&appRuntimeWithCommmonScenarios, testErr),
			Input:          *modelRequestInput,
			InputAuth:      modelAuth,
			InputSchema:    nil,
			ExpectedOutput: "",
			ExpectedError:  errors.New("while creating bundle instance auth scenario label"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			instanceAuthRepo := testCase.InstanceAuthRepoFn()
			bundleSvc := testCase.BundleSvcFn()
			labelSvc := testCase.LabelSvcFn()
			scenarioSvc := testCase.ScenarioSvcFn()

			uidSvc := testCase.UIDSvcFn()

			svc := bundleinstanceauth.NewService(instanceAuthRepo, uidSvc, bundleSvc, scenarioSvc, labelSvc)
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

			mock.AssertExpectationsForObjects(t, instanceAuthRepo, uidSvc, bundleSvc, scenarioSvc, labelSvc)
		})
	}

	t.Run("Error when consumer is not in the context", func(t *testing.T) {
		// GIVEN
		svc := bundleinstanceauth.NewService(nil, nil, nil, nil, nil)
		ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

		// WHEN
		_, err := svc.Create(ctx, testBundleID, *modelRequestInput, modelAuth, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read consumer from context")
	})

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := bundleinstanceauth.NewService(nil, nil, nil, nil, nil)

		// WHEN
		_, err := svc.Create(context.TODO(), testBundleID, model.BundleInstanceAuthRequestInput{}, nil, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})

	t.Run("Success - when consumer time is not Runtime then no bundle instance auth scenario association is performed", func(t *testing.T) {
		// GIVEN
		instanceAuthRepo := &automock.Repository{}
		instanceAuthRepo.On("Create", contextThatHasTenant(testTenant), mock.Anything).Return(nil).Once()

		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(testID).Once()

		bundleSvc := unusedBundleService()
		scenarioSvc := unusedScenarioService()
		labelSvc := unusedLabelService()

		svc := bundleinstanceauth.NewService(instanceAuthRepo, uidSvc, bundleSvc, scenarioSvc, labelSvc)
		ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

		consumerEntity := consumer.Consumer{
			ConsumerID:   "",
			ConsumerType: consumer.Application,
		}
		ctx = consumer.SaveToContext(ctx, consumerEntity)

		// WHEN
		_, err := svc.Create(ctx, testBundleID, *modelRequestInput, nil, nil)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, instanceAuthRepo, uidSvc, bundleSvc, scenarioSvc, labelSvc)
	})

}

func TestService_ListByApplicationID(t *testing.T) {
	// given
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

			svc := bundleinstanceauth.NewService(repo, nil, nil, nil, nil)

			// when
			pia, err := svc.List(ctx, testBundleID)

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
		svc := bundleinstanceauth.NewService(nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.List(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByRuntimeID(t *testing.T) {
	// given
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

			svc := bundleinstanceauth.NewService(repo, nil, nil, nil, nil)

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
		svc := bundleinstanceauth.NewService(nil, nil, nil, nil, nil)

		// WHEN
		_, err := svc.ListByRuntimeID(context.TODO(), "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
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
				repo.On("Update", ctx, bundleInstanceAuth).Return(nil).Once()
				return repo
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when bundle instance auth update failed",
			RepositoryFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("Update", ctx, bundleInstanceAuth).Return(testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := bundleinstanceauth.NewService(repo, nil, nil, nil, nil)

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
				instanceAuthRepo.On("Update", contextThatHasTenant(tnt), mock.MatchedBy(func(in *model.BundleInstanceAuth) bool {
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
				instanceAuthRepo.On("Update", contextThatHasTenant(tnt), mock.MatchedBy(func(in *model.BundleInstanceAuth) bool {
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

			svc := bundleinstanceauth.NewService(instanceAuthRepo, nil, nil, nil, nil)
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
		svc := bundleinstanceauth.NewService(nil, nil, nil, nil, nil)
		_, err := svc.RequestDeletion(ctx, nil, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), expectedError.Error())
	})
}

func TestService_GetForAppAndAnyMatchingScenarios(t *testing.T) {
	methodExecution := func(ctx context.Context, repo *automock.Repository, appId string, scenarios []string) ([]*model.BundleInstanceAuth, error) {
		svc := bundleinstanceauth.NewService(repo, nil, nil, nil, nil)
		return svc.GetForAppAndAnyMatchingScenarios(ctx, appId, scenarios)
	}
	testService_GetForObjectAndAnyMatchingScenarios(t, "GetForAppAndAnyMatchingScenarios", methodExecution)
}

func TestService_GetForRuntimeAndAnyMatchingScenarios(t *testing.T) {
	methodExecution := func(ctx context.Context, repo *automock.Repository, appId string, scenarios []string) ([]*model.BundleInstanceAuth, error) {
		svc := bundleinstanceauth.NewService(repo, nil, nil, nil, nil)
		return svc.GetForRuntimeAndAnyMatchingScenarios(ctx, appId, scenarios)
	}
	testService_GetForObjectAndAnyMatchingScenarios(t, "GetForRuntimeAndAnyMatchingScenarios", methodExecution)
}

func TestService_AssociateBundleInstanceAuthForNewApplicationScenarios(t *testing.T) {
	// Given
	//testErr := errors.New("Test error")
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	applicationID := "foo"

	scenarioToRemove := "existing-sc-1"
	scenarioToKeep := "existing-sc-2"
	scenarioToAdd1 := "scenario-1"
	scenarioToAdd2 := "scenario-2"

	existingRuntimeScenarioLabels := []model.Label{
		{ObjectID: "runtime-1", Value: []string{scenarioToKeep, scenarioToAdd1}},
		{ObjectID: "runtime-2", Value: []string{scenarioToKeep, scenarioToAdd2}},
	}

	testCases := []struct {
		Name                   string
		ScenarioSvcFn          func() *automock.ScenarioService
		LabelSvcFn             func() *automock.LabelService
		ExistingScenariosParam []string
		InputScenariosParam    []string
		ExpectedError          error
	}{
		{
			Name: "Success when no old scenarios exist",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				return &automock.ScenarioService{}
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep},
			ExpectedError:          nil,
		},
		{
			Name: "Success when not adding any new scenario",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				return &automock.ScenarioService{}
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep},
			ExpectedError:          nil,
		},
		{
			Name: "Success when both remove some old scenarios and add new ones",
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertScenarios", contextThatHasTenant(testTenant), testTenant, mock.Anything, []string{scenarioToAdd1}, mock.Anything).
					Return(nil).Once()
				svc.On("UpsertScenarios", contextThatHasTenant(testTenant), testTenant, mock.Anything, []string{scenarioToAdd2}, mock.Anything).
					Return(nil).Once()
				return svc
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetRuntimeScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(existingRuntimeScenarioLabels, nil).Once()

				rmt1Labels := []model.Label{{ObjectID: "runtime-2", Value: []string{scenarioToKeep, scenarioToAdd2}}}
				rmt2Labels := []model.Label{{ObjectID: "runtime-2", Value: []string{scenarioToKeep, scenarioToAdd2}}}

				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), applicationID, "runtime-1").Return(rmt1Labels, nil).Once()
				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), applicationID, "runtime-2").Return(rmt2Labels, nil).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToRemove, scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1, scenarioToAdd2},
			ExpectedError:          nil,
		},
		{
			Name: "Success when there are no scenario to keep",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				return &automock.ScenarioService{}
			},
			ExistingScenariosParam: []string{scenarioToRemove},
			InputScenariosParam:    []string{scenarioToAdd1},
			ExpectedError:          nil,
		},
		{
			Name: "Success when there are NO common runtimes that application is connected between existing scenarios and any of the new ones",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetRuntimeScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).Return([]model.Label{}, nil).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          nil,
		},
		{
			Name: "Success when there are NO scenarios to remove",
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}

				svc.On("UpsertScenarios", contextThatHasTenant(testTenant), testTenant, mock.Anything, []string{scenarioToAdd1}, mock.Anything).
					Return(nil).Once()
				return svc
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetRuntimeScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(existingRuntimeScenarioLabels, nil).Once()

				rmt1Labels := []model.Label{{ObjectID: "runtime-1", Value: []string{scenarioToKeep, scenarioToAdd1}}}
				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), applicationID, "runtime-1").Return(rmt1Labels, nil).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          nil,
		},
		{
			Name: "Success when there are only scenarios to remove",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				return &automock.ScenarioService{}
			},
			ExistingScenariosParam: []string{scenarioToRemove},
			InputScenariosParam:    []string{},
			ExpectedError:          nil,
		},
		{
			Name: "Error when UpsertScenarios fails",
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}

				svc.On("UpsertScenarios", contextThatHasTenant(testTenant), testTenant, mock.Anything, []string{scenarioToAdd1}, mock.Anything).
					Return(testError).Once()
				return svc
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetRuntimeScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(existingRuntimeScenarioLabels, nil).Once()

				rmt1Labels := []model.Label{{ObjectID: "runtime-1", Value: []string{scenarioToKeep, scenarioToAdd1}}}
				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), applicationID, "runtime-1").Return(rmt1Labels, nil).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          testError,
		},
		{
			Name: "Error when getting runtime scenario labels for scenarios that should be kept",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetRuntimeScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(nil, testError).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToRemove, scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          testError,
		},
		{
			Name: "Error when getting bundle instance auth scenario labels by appId and runtimeId",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetRuntimeScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(existingRuntimeScenarioLabels, nil).Once()

				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), applicationID, "runtime-1").Return(nil, testError).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToRemove, scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelSvc := testCase.LabelSvcFn()
			scenarioSvc := testCase.ScenarioSvcFn()

			svc := bundleinstanceauth.NewService(nil, nil, nil, scenarioSvc, labelSvc)
			svc.SetTimestampGen(func() time.Time { return testTime })

			// WHEN
			err := svc.AssociateBundleInstanceAuthForNewApplicationScenarios(ctx, testCase.ExistingScenariosParam, testCase.InputScenariosParam, applicationID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, scenarioSvc, labelSvc)
		})
	}
}

func TestService_AssociateBundleInstanceAuthForNewRuntimeScenarios(t *testing.T) {
	// Given
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	runtimeId := "foo"

	scenarioToRemove := "existing-sc-1"
	scenarioToKeep := "existing-sc-2"
	scenarioToAdd1 := "scenario-1"
	scenarioToAdd2 := "scenario-2"

	existingApplicationScenarioLabels := []model.Label{
		{ObjectID: "app-1", Value: []string{scenarioToKeep, scenarioToAdd1}},
		{ObjectID: "app-2", Value: []string{scenarioToKeep, scenarioToAdd2}},
	}

	testCases := []struct {
		Name                   string
		ScenarioSvcFn          func() *automock.ScenarioService
		LabelSvcFn             func() *automock.LabelService
		ExistingScenariosParam []string
		InputScenariosParam    []string
		ExpectedError          error
	}{
		{
			Name: "Success when no old scenarios exist",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				return &automock.ScenarioService{}
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep},
			ExpectedError:          nil,
		},
		{
			Name: "Success when not adding any new scenario",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				return &automock.ScenarioService{}
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep},
			ExpectedError:          nil,
		},
		{
			Name: "Success when both remove some old scenarios and add new ones",
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("UpsertScenarios", contextThatHasTenant(testTenant), testTenant, mock.Anything, []string{scenarioToAdd1}, mock.Anything).
					Return(nil).Once()
				svc.On("UpsertScenarios", contextThatHasTenant(testTenant), testTenant, mock.Anything, []string{scenarioToAdd2}, mock.Anything).
					Return(nil).Once()
				return svc
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetApplicationScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(existingApplicationScenarioLabels, nil).Once()

				app1Labels := []model.Label{{ObjectID: "app-1", Value: []string{scenarioToKeep, scenarioToAdd2}}}
				app2Labels := []model.Label{{ObjectID: "app-2", Value: []string{scenarioToKeep, scenarioToAdd2}}}

				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), "app-1", runtimeId).Return(app1Labels, nil).Once()
				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), "app-2", runtimeId).Return(app2Labels, nil).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToRemove, scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1, scenarioToAdd2},
			ExpectedError:          nil,
		},
		{
			Name: "Success when there are no scenario to keep",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				return &automock.ScenarioService{}
			},
			ExistingScenariosParam: []string{scenarioToRemove},
			InputScenariosParam:    []string{scenarioToAdd1},
			ExpectedError:          nil,
		},
		{
			Name: "Success when there are NO common applications that runtime is connected between existing scenarios and any of the new ones",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetApplicationScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).Return([]model.Label{}, nil).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          nil,
		},
		{
			Name: "Success when there are NO scenarios to remove",
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}

				svc.On("UpsertScenarios", contextThatHasTenant(testTenant), testTenant, mock.Anything, []string{scenarioToAdd1}, mock.Anything).
					Return(nil).Once()
				return svc
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetApplicationScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(existingApplicationScenarioLabels, nil).Once()

				app1Labels := []model.Label{{ObjectID: "app-1", Value: []string{scenarioToKeep, scenarioToAdd1}}}
				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), "app-1", runtimeId).Return(app1Labels, nil).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          nil,
		},
		{
			Name: "Success when there are only scenarios to remove and no bundle_instance_auth associated to it",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				return &automock.ScenarioService{}
			},
			ExistingScenariosParam: []string{scenarioToRemove},
			InputScenariosParam:    []string{},
			ExpectedError:          nil,
		},
		{
			Name: "Error when UpsertScenarios fails",
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}

				svc.On("UpsertScenarios", contextThatHasTenant(testTenant), testTenant, mock.Anything, []string{scenarioToAdd1}, mock.Anything).
					Return(testError).Once()
				return svc
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetApplicationScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(existingApplicationScenarioLabels, nil).Once()

				app1Labels := []model.Label{{ObjectID: "app-1", Value: []string{scenarioToKeep, scenarioToAdd1}}}
				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), "app-1", runtimeId).Return(app1Labels, nil).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          testError,
		},
		{
			Name: "Error when getting application scenario labels for scenarios that should be kept",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetApplicationScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(nil, testError).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToRemove, scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          testError,
		},
		{
			Name: "Error when getting bundle instance auth scenario labels by appId and runtimeId",
			LabelSvcFn: func() *automock.LabelService {
				return &automock.LabelService{}
			},
			ScenarioSvcFn: func() *automock.ScenarioService {
				svc := &automock.ScenarioService{}
				svc.On("GetApplicationScenarioLabelsForAnyMatchingScenario", contextThatHasTenant(testTenant), []string{scenarioToKeep}).
					Return(existingApplicationScenarioLabels, nil).Once()

				svc.On("GetBundleInstanceAuthsScenarioLabels", contextThatHasTenant(testTenant), "app-1", runtimeId).Return(nil, testError).Once()
				return svc
			},
			ExistingScenariosParam: []string{scenarioToRemove, scenarioToKeep},
			InputScenariosParam:    []string{scenarioToKeep, scenarioToAdd1},
			ExpectedError:          testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			labelSvc := testCase.LabelSvcFn()
			scenarioSvc := testCase.ScenarioSvcFn()

			svc := bundleinstanceauth.NewService(nil, nil, nil, scenarioSvc, labelSvc)
			svc.SetTimestampGen(func() time.Time { return testTime })

			// WHEN
			err := svc.AssociateBundleInstanceAuthForNewRuntimeScenarios(ctx, testCase.ExistingScenariosParam, testCase.InputScenariosParam, runtimeId)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, scenarioSvc, labelSvc)
		})
	}
}

func testService_GetForObjectAndAnyMatchingScenarios(t *testing.T, repoMethodName string, call func(ctx context.Context, repo *automock.Repository, appId string, scenarios []string) ([]*model.BundleInstanceAuth, error)) {
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	objId := "foo"

	t.Run("Success", func(t *testing.T) {
		scenarios := []string{"scenario-1"}
		authId := "bar"
		auth := &model.BundleInstanceAuth{ID: authId}

		repo := &automock.Repository{}
		repo.On(repoMethodName, ctx, testTenant, objId, scenarios).Return([]*model.BundleInstanceAuth{auth}, nil)

		result, err := call(ctx, repo, objId, scenarios)

		require.NoError(t, err)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, authId, result[0].ID)
		repo.AssertExpectations(t)
	})

	t.Run("Success when empty scenarios slice is provided", func(t *testing.T) {
		result, err := call(ctx, nil, objId, []string{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Success when repository fails", func(t *testing.T) {
		scenarios := []string{"scenario-1"}
		repo := &automock.Repository{}
		repo.On(repoMethodName, ctx, testTenant, objId, scenarios).Return(nil, testError)

		_, err := call(ctx, repo, objId, scenarios)
		require.Error(t, err)
		repo.AssertExpectations(t)
	})
}

type appAndRuntimeScenarios struct {
	appScenarios     []string
	runtimeScenarios []string
}

func (sc *appAndRuntimeScenarios) getCommonScenarios() []string {
	return str.IntersectSlice(sc.appScenarios, sc.runtimeScenarios)
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

func unusedScenarioService() *automock.ScenarioService {
	return &automock.ScenarioService{}
}

func unusedBundleService() *automock.BundleService {
	return &automock.BundleService{}
}

func unusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}

func newScenarioSvcFn(appRuntScenarios *appAndRuntimeScenarios) func() *automock.ScenarioService {
	return func() *automock.ScenarioService {
		svc := automock.ScenarioService{}
		svc.On("GetScenarioNamesForApplication", contextThatHasTenant(testTenant), testApplicationID).Return(appRuntScenarios.appScenarios, nil)
		svc.On("GetScenarioNamesForRuntime", contextThatHasTenant(testTenant), testRuntimeID).Return(appRuntScenarios.runtimeScenarios, nil)
		return &svc
	}
}

func newLabelSvcFnThatSucceeds(appRuntScenarios *appAndRuntimeScenarios) func() *automock.LabelService {
	return func() *automock.LabelService {
		svc := automock.LabelService{}
		svc.On("UpsertLabel", contextThatHasTenant(testTenant), testTenant, matchLabelInputScenarios(appRuntScenarios)).Return(nil)
		return &svc
	}
}

func newLabelSvcFnThatFail(appRuntScenarios *appAndRuntimeScenarios, err error) func() *automock.LabelService {
	return func() *automock.LabelService {
		svc := automock.LabelService{}
		svc.On("UpsertLabel", contextThatHasTenant(testTenant), testTenant, matchLabelInputScenarios(appRuntScenarios)).Return(err)
		return &svc
	}
}

func matchLabelInputScenarios(appRuntScenarios *appAndRuntimeScenarios) interface{} {
	return mock.MatchedBy(func(lbl *model.LabelInput) bool {
		scenarios, err := labelpkg.GetScenariosFromValueAsStringSlice(lbl.Value)
		return err == nil && reflect.DeepEqual(scenarios, appRuntScenarios.getCommonScenarios())
	})
}

func newBundleSvcThatGetBundleById() *automock.BundleService {
	svc := automock.BundleService{}
	bndl := &model.Bundle{ApplicationID: testApplicationID}
	svc.On("Get", contextThatHasTenant(testTenant), testBundleID).Return(bndl, nil).Once()
	return &svc
}
