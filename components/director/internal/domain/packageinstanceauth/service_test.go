package packageinstanceauth_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
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

func TestService_SetAuth(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.Background(), testTenant)

	modelInstanceAuthFn := func() *model.PackageInstanceAuth {
		return fixModelPackageInstanceAuth(testID, testPackageID, testTenant, nil, fixModelStatusPending())
	}

	modelSetInput := fixModelSetInput()
	modelUpdatedInstanceAuth := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, nil, fixModelStatusPending())
	modelUpdatedInstanceAuth.Auth = modelSetInput.Auth.ToAuth()
	modelUpdatedInstanceAuth.Status = &model.PackageInstanceAuthStatus{
		Condition: model.PackageInstanceAuthStatusConditionSucceeded,
		Timestamp: testTime,
		Message:   modelSetInput.Status.Message,
		Reason:    modelSetInput.Status.Reason,
	}

	modelSetInputWithoutStatus := model.PackageInstanceAuthSetInput{
		Auth:   fixModelAuthInput(),
		Status: nil,
	}
	modelUpdatedInstanceAuthWithDefaultStatus := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, nil, fixModelStatusPending())
	modelUpdatedInstanceAuthWithDefaultStatus.Auth = modelSetInputWithoutStatus.Auth.ToAuth()
	err := modelUpdatedInstanceAuthWithDefaultStatus.SetDefaultStatus(model.PackageInstanceAuthStatusConditionSucceeded, testTime)
	require.NoError(t, err)

	testCases := []struct {
		Name               string
		InstanceAuthRepoFn func() *automock.Repository
		Input              model.PackageInstanceAuthSetInput
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
			Name: "Error when Package Instance Auth retrieval failed",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(modelInstanceAuthFn(), testError).Once()
				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: testError,
		},
		{
			Name: "Error when Package Instance Auth update failed",
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
			Name: "Error when Package Instance Auth status is nil",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(
					&model.PackageInstanceAuth{
						Status: nil,
					}, nil).Once()
				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: errors.New("auth can be set only on Package Instance Auths in PENDING state"),
		},
		{
			Name: "Error when Package Instance Auth status condition different from PENDING",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(
					&model.PackageInstanceAuth{
						Status: &model.PackageInstanceAuthStatus{
							Condition: model.PackageInstanceAuthStatusConditionSucceeded,
						},
					}, nil).Once()
				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: errors.New("auth can be set only on Package Instance Auths in PENDING state"),
		},
		{
			Name: "Error when retrieved Package Instance Auth is nil",
			InstanceAuthRepoFn: func() *automock.Repository {
				instanceAuthRepo := &automock.Repository{}
				instanceAuthRepo.On("GetByID", contextThatHasTenant(testTenant), testTenant, testID).Return(nil, nil).Once()

				return instanceAuthRepo
			},
			Input:         *modelSetInput,
			ExpectedError: errors.Errorf("Package Instance Auth with ID %s not found", testID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			instanceAuthRepo := testCase.InstanceAuthRepoFn()

			svc := packageinstanceauth.NewService(instanceAuthRepo, nil)
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
		svc := packageinstanceauth.NewService(nil, nil)

		// WHEN
		err := svc.SetAuth(context.Background(), testID, model.PackageInstanceAuthSetInput{})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.Background(), testTenant)

	modelAuth := fixModelAuth()
	modelExpectedInstanceAuth := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, modelAuth, fixModelStatusSucceeded())
	modelExpectedInstanceAuthPending := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, nil, fixModelStatusPending())

	modelRequestInput := fixModelRequestInput()

	testCases := []struct {
		Name               string
		InstanceAuthRepoFn func() *automock.Repository
		UIDSvcFn           func() *automock.UIDService
		Input              model.PackageInstanceAuthRequestInput
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
			Name: "Error when creating Package Instance Auth",
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
			Input:          model.PackageInstanceAuthRequestInput{},
			InputAuth:      modelAuth,
			InputSchema:    str.Ptr("{\"type\": \"string\"}"),
			ExpectedOutput: "",
			ExpectedError:  errors.New("json schema for input parameters was defined for the package but no input parameters were provided"),
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
			Input: model.PackageInstanceAuthRequestInput{
				InputParams: str.Ptr("{"),
			},
			InputAuth:      modelAuth,
			InputSchema:    str.Ptr("{\"type\": \"string\"}"),
			ExpectedOutput: "",
			ExpectedError:  errors.New(`while validating value { against JSON Schema: {"type": "string"}: unexpected EOF`),
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

			svc := packageinstanceauth.NewService(instanceAuthRepo, uidSvc)
			svc.SetTimestampGen(func() time.Time { return testTime })

			// WHEN
			result, err := svc.Create(ctx, testPackageID, testCase.Input, testCase.InputAuth, testCase.InputSchema)

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
		svc := packageinstanceauth.NewService(nil, nil)

		// WHEN
		_, err := svc.Create(context.Background(), testPackageID, model.PackageInstanceAuthRequestInput{}, nil, nil)

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
