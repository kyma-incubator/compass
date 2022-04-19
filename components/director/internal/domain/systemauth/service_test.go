package systemauth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	sysAuthID := "foo"
	objID := "bar"

	modelAuthInput := fixModelAuthInput()
	modelAuth := fixModelAuth()

	uidSvcFn := func() *automock.UIDService {
		uidSvc := &automock.UIDService{}
		uidSvc.On("Generate").Return(sysAuthID)
		return uidSvc
	}

	testCases := []struct {
		Name            string
		sysAuthRepoFn   func() *automock.Repository
		InputObjectType pkgmodel.SystemAuthReferenceObjectType
		InputAuth       *model.AuthInput
		ExpectedOutput  string
		ExpectedError   error
	}{
		{
			Name: "Success creating auth for Runtime",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("Create", contextThatHasTenant(testTenant), *fixModelSystemAuth(sysAuthID, pkgmodel.RuntimeReference, objID, modelAuth)).Return(nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.RuntimeReference,
			InputAuth:       &modelAuthInput,
			ExpectedOutput:  sysAuthID,
			ExpectedError:   nil,
		},
		{
			Name: "Success creating auth for Application",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("Create", contextThatHasTenant(testTenant), *fixModelSystemAuth(sysAuthID, pkgmodel.ApplicationReference, objID, modelAuth)).Return(nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.ApplicationReference,
			InputAuth:       &modelAuthInput,
			ExpectedOutput:  sysAuthID,
			ExpectedError:   nil,
		},
		{
			Name: "Success creating auth for Integration System",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("Create", contextThatHasTenant(testTenant), *fixModelSystemAuth(sysAuthID, pkgmodel.IntegrationSystemReference, objID, modelAuth)).Return(nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.IntegrationSystemReference,
			InputAuth:       &modelAuthInput,
			ExpectedOutput:  sysAuthID,
			ExpectedError:   nil,
		},
		{
			Name: "Success creating auth with nil value",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("Create", contextThatHasTenant(testTenant), *fixModelSystemAuth(sysAuthID, pkgmodel.RuntimeReference, objID, nil)).Return(nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.RuntimeReference,
			InputAuth:       nil,
			ExpectedOutput:  sysAuthID,
			ExpectedError:   nil,
		},
		{
			Name: "Error creating auth for unknown object type",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				return sysAuthRepo
			},
			InputObjectType: "unknown",
			InputAuth:       &modelAuthInput,
			ExpectedOutput:  "",
			ExpectedError:   errors.New("unknown reference object type"),
		},
		{
			Name: "Error creating System Auth",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("Create", contextThatHasTenant(testTenant), *fixModelSystemAuth(sysAuthID, pkgmodel.RuntimeReference, objID, modelAuth)).Return(testErr)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.RuntimeReference,
			InputAuth:       &modelAuthInput,
			ExpectedOutput:  "",
			ExpectedError:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			sysAuthRepo := testCase.sysAuthRepoFn()
			uidSvc := uidSvcFn()
			svc := systemauth.NewService(sysAuthRepo, uidSvc)

			// WHEN
			result, err := svc.Create(ctx, testCase.InputObjectType, objID, testCase.InputAuth)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			sysAuthRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		uidSvc := uidSvcFn()
		defer uidSvc.AssertExpectations(t)
		svc := systemauth.NewService(nil, uidSvc)

		// WHEN
		_, err := svc.Create(context.TODO(), "", "", nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

// Just happy path, as it is the same as Create method
func TestService_CreateWithCustomID(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	sysAuthID := "bla"
	objID := "bar"

	modelAuthInput := fixModelAuthInput()
	modelAuth := fixModelAuth()

	sysAuthRepo := &automock.Repository{}
	sysAuthRepo.On("Create", contextThatHasTenant(testTenant), *fixModelSystemAuth(sysAuthID, pkgmodel.RuntimeReference, objID, modelAuth)).Return(nil)
	defer sysAuthRepo.AssertExpectations(t)

	svc := systemauth.NewService(sysAuthRepo, nil)

	// WHEN
	result, err := svc.CreateWithCustomID(ctx, sysAuthID, pkgmodel.RuntimeReference, objID, &modelAuthInput)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, sysAuthID, result)
}

func TestService_ListForObject(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	objID := "bar"

	modelAuth := fixModelAuth()

	expectedRtmSysAuths := []pkgmodel.SystemAuth{
		{
			ID:        "foo",
			TenantID:  &testTenant,
			RuntimeID: str.Ptr("bar"),
			Value:     modelAuth,
		},
		{
			ID:        "foo2",
			TenantID:  &testTenant,
			RuntimeID: str.Ptr("bar2"),
			Value:     modelAuth,
		},
	}
	expectedAppSysAuths := []pkgmodel.SystemAuth{
		{
			ID:       "foo",
			TenantID: &testTenant,
			AppID:    str.Ptr("bar"),
			Value:    modelAuth,
		},
		{
			ID:       "foo2",
			TenantID: &testTenant,
			AppID:    str.Ptr("bar2"),
			Value:    modelAuth,
		},
	}
	expectedIntSysAuths := []pkgmodel.SystemAuth{
		{
			ID:                  "foo",
			TenantID:            nil,
			IntegrationSystemID: str.Ptr("bar"),
			Value:               modelAuth,
		},
		{
			ID:                  "foo2",
			TenantID:            nil,
			IntegrationSystemID: str.Ptr("bar2"),
			Value:               modelAuth,
		},
	}

	testCases := []struct {
		Name            string
		sysAuthRepoFn   func() *automock.Repository
		InputObjectType pkgmodel.SystemAuthReferenceObjectType
		ExpectedOutput  []pkgmodel.SystemAuth
		ExpectedError   error
	}{
		{
			Name: "Success listing Auths for Runtime",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("ListForObject", contextThatHasTenant(testTenant), testTenant, pkgmodel.RuntimeReference, objID).Return(expectedRtmSysAuths, nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.RuntimeReference,
			ExpectedOutput:  expectedRtmSysAuths,
			ExpectedError:   nil,
		},
		{
			Name: "Success listing Auths for Application",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("ListForObject", contextThatHasTenant(testTenant), testTenant, pkgmodel.ApplicationReference, objID).Return(expectedAppSysAuths, nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.ApplicationReference,
			ExpectedOutput:  expectedAppSysAuths,
			ExpectedError:   nil,
		},
		{
			Name: "Success listing Auths for Integration System",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("ListForObjectGlobal", contextThatHasTenant(testTenant), pkgmodel.IntegrationSystemReference, objID).Return(expectedIntSysAuths, nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.IntegrationSystemReference,
			ExpectedOutput:  expectedIntSysAuths,
			ExpectedError:   nil,
		},
		{
			Name: "Error listing System Auths",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("ListForObject", contextThatHasTenant(testTenant), testTenant, pkgmodel.RuntimeReference, objID).Return(nil, testErr)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.RuntimeReference,
			ExpectedOutput:  nil,
			ExpectedError:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			sysAuthRepo := testCase.sysAuthRepoFn()
			svc := systemauth.NewService(sysAuthRepo, nil)

			// WHEN
			result, err := svc.ListForObject(ctx, testCase.InputObjectType, objID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			sysAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := systemauth.NewService(nil, nil)

		// WHEN
		_, err := svc.ListForObject(context.TODO(), "", "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetByIDForObject(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)

	sysAuthID := "foo"
	modelSysAuth := fixModelSystemAuth(sysAuthID, pkgmodel.RuntimeReference, "bar", nil)

	testCases := []struct {
		Name            string
		sysAuthRepoFn   func() *automock.Repository
		InputObjectType pkgmodel.SystemAuthReferenceObjectType
		ExpectedSysAuth *pkgmodel.SystemAuth
		ExpectedError   error
	}{
		{
			Name: "Success getting auth for Runtime",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("GetByIDForObject", contextThatHasTenant(testTenant), testTenant, sysAuthID, pkgmodel.RuntimeReference).Return(modelSysAuth, nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.RuntimeReference,
			ExpectedError:   nil,
			ExpectedSysAuth: modelSysAuth,
		},
		{
			Name: "Success getting auth for Application",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("GetByIDForObject", contextThatHasTenant(testTenant), testTenant, sysAuthID, pkgmodel.ApplicationReference).Return(modelSysAuth, nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.ApplicationReference,
			ExpectedError:   nil,
			ExpectedSysAuth: modelSysAuth,
		},
		{
			Name: "Success getting auth for Integration System",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("GetByIDForObjectGlobal", contextThatHasTenant(testTenant), sysAuthID, pkgmodel.IntegrationSystemReference).Return(modelSysAuth, nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.IntegrationSystemReference,
			ExpectedError:   nil,
			ExpectedSysAuth: modelSysAuth,
		},
		{
			Name: "Error getting System Auths",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("GetByIDForObject", contextThatHasTenant(testTenant), testTenant, sysAuthID, pkgmodel.RuntimeReference).Return(nil, testErr)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.RuntimeReference,
			ExpectedError:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			sysAuthRepo := testCase.sysAuthRepoFn()
			svc := systemauth.NewService(sysAuthRepo, nil)

			// WHEN
			sysAuth, err := svc.GetByIDForObject(ctx, testCase.InputObjectType, sysAuthID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.ExpectedSysAuth, sysAuth)
			}

			sysAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := systemauth.NewService(nil, nil)

		// WHEN
		err := svc.DeleteByIDForObject(context.TODO(), "", "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_DeleteByIDForObject(t *testing.T) {
	// GIVEN
	ctx := tenant.SaveToContext(context.TODO(), testTenant, testExternalTenant)
	sysAuthID := "foo"

	testCases := []struct {
		Name            string
		sysAuthRepoFn   func() *automock.Repository
		InputObjectType pkgmodel.SystemAuthReferenceObjectType
		ExpectedError   error
	}{
		{
			Name: "Success deleting auth for Runtime",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("DeleteByIDForObject", contextThatHasTenant(testTenant), testTenant, sysAuthID, pkgmodel.RuntimeReference).Return(nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.RuntimeReference,
			ExpectedError:   nil,
		},
		{
			Name: "Success deleting auth for Application",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("DeleteByIDForObject", contextThatHasTenant(testTenant), testTenant, sysAuthID, pkgmodel.ApplicationReference).Return(nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.ApplicationReference,
			ExpectedError:   nil,
		},
		{
			Name: "Success deleting auth for Integration System",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("DeleteByIDForObjectGlobal", contextThatHasTenant(testTenant), sysAuthID, pkgmodel.IntegrationSystemReference).Return(nil)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.IntegrationSystemReference,
			ExpectedError:   nil,
		},
		{
			Name: "Error deleting System Auths",
			sysAuthRepoFn: func() *automock.Repository {
				sysAuthRepo := &automock.Repository{}
				sysAuthRepo.On("DeleteByIDForObject", contextThatHasTenant(testTenant), testTenant, sysAuthID, pkgmodel.RuntimeReference).Return(testErr)
				return sysAuthRepo
			},
			InputObjectType: pkgmodel.RuntimeReference,
			ExpectedError:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			sysAuthRepo := testCase.sysAuthRepoFn()
			svc := systemauth.NewService(sysAuthRepo, nil)

			// WHEN
			err := svc.DeleteByIDForObject(ctx, testCase.InputObjectType, sysAuthID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			sysAuthRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := systemauth.NewService(nil, nil)

		// WHEN
		err := svc.DeleteByIDForObject(context.TODO(), "", "")

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetGlobal(t *testing.T) {
	authID := "authID"

	t.Run("success when systemAuth can be fetched from repo", func(t *testing.T) {
		// GIVEN
		repo := &automock.Repository{}
		defer repo.AssertExpectations(t)
		repo.On("GetByIDGlobal", context.Background(), authID).Return(&pkgmodel.SystemAuth{}, nil)
		svc := systemauth.NewService(repo, nil)
		// WHEN
		item, err := svc.GetGlobal(context.Background(), authID)
		// THEN
		assert.Nil(t, err)
		assert.Equal(t, &pkgmodel.SystemAuth{}, item)
	})

	t.Run("error when systemAuth cannot be fetched from repo", func(t *testing.T) {
		// GIVEN
		repo := &automock.Repository{}
		defer repo.AssertExpectations(t)
		repo.On("GetByIDGlobal", context.Background(), authID).Return(nil, errors.New("could not fetch"))
		svc := systemauth.NewService(repo, nil)
		// WHEN
		item, err := svc.GetGlobal(context.Background(), authID)
		// THEN
		assert.Nil(t, item)
		assert.Error(t, err, fmt.Sprintf("while getting SystemAuth with ID %s could not fetch", authID))
	})
}

func TestService_GetByToken(t *testing.T) {
	token := "YWJj"
	input := map[string]interface{}{
		"OneTimeToken": map[string]interface{}{
			"Token": token,
			"Used":  false,
		}}

	t.Run("success when systemAuth can be fetched from repo", func(t *testing.T) {
		// GIVEN
		repo := &automock.Repository{}
		defer repo.AssertExpectations(t)
		repo.On("GetByJSONValue", context.Background(), input).Return(&pkgmodel.SystemAuth{}, nil)
		svc := systemauth.NewService(repo, nil)
		// WHEN
		item, err := svc.GetByToken(context.Background(), token)
		// THEN
		assert.Nil(t, err)
		assert.Equal(t, &pkgmodel.SystemAuth{}, item)
	})

	t.Run("error when systemAuth cannot be fetched from repo", func(t *testing.T) {
		// GIVEN
		repo := &automock.Repository{}
		defer repo.AssertExpectations(t)
		repo.On("GetByJSONValue", context.Background(), input).Return(nil, errors.New("err"))
		svc := systemauth.NewService(repo, nil)
		// WHEN
		item, err := svc.GetByToken(context.Background(), token)
		// THEN
		assert.Error(t, err)
		assert.Nil(t, item)
	})
}

func TestService_UpdateValue(t *testing.T) {
	authID := "authID"
	modelAuth := fixModelAuth()

	t.Run("error when systemAuth cannot be fetched from repo", func(t *testing.T) {
		// GIVEN
		repo := &automock.Repository{}
		defer repo.AssertExpectations(t)
		repo.On("GetByIDGlobal", context.Background(), authID).Return(nil, errors.New("could not fetch"))
		svc := systemauth.NewService(repo, nil)
		// WHEN
		item, err := svc.UpdateValue(context.Background(), authID, modelAuth)
		// THEN
		assert.Nil(t, item)
		assert.Error(t, err, fmt.Sprintf("while getting SystemAuth with ID %s could not fetch", authID))
	})

	t.Run("Error when systemAuth cannot be updated", func(t *testing.T) {
		// GIVEN
		repo := &automock.Repository{}
		defer repo.AssertExpectations(t)
		repo.On("GetByIDGlobal", context.Background(), authID).Return(&pkgmodel.SystemAuth{}, nil)
		repo.On("Update", context.Background(), mock.Anything).Return(errors.New("could not update"))
		svc := systemauth.NewService(repo, nil)
		// WHEN
		item, err := svc.UpdateValue(context.Background(), authID, modelAuth)
		// THEN
		assert.Nil(t, item)
		assert.Error(t, err, fmt.Sprintf("while getting SystemAuth with ID %s could not update", authID))
	})

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		repo := &automock.Repository{}
		defer repo.AssertExpectations(t)
		sysAuth := &pkgmodel.SystemAuth{
			Value: modelAuth,
		}
		repo.On("GetByIDGlobal", context.Background(), authID).Return(sysAuth, nil)
		repo.On("Update", context.Background(), sysAuth).Return(nil)
		svc := systemauth.NewService(repo, nil)
		// WHEN
		item, err := svc.UpdateValue(context.Background(), authID, modelAuth)

		// THEN
		assert.Nil(t, err)
		assert.Equal(t, sysAuth, item)
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
