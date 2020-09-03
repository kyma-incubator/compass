package mp_bundle_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	mp_bundle "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	timestamp := time.Now()
	id := "foo"
	applicationID := "appid"
	name := "foo"
	desc := "bar"
	spec := "test"

	modelInput := model.BundleCreateInput{
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &model.AuthInput{},
		Documents: []*model.DocumentInput{
			{Title: "foo", Description: "test", FetchRequest: &model.FetchRequestInput{URL: "doc.foo.bar"}},
			{Title: "bar", Description: "test"},
		},
		APIDefinitions: []*model.APIDefinitionInput{
			{
				Name: "foo",
				Spec: &model.APISpecInput{FetchRequest: &model.FetchRequestInput{URL: "api.foo.bar"}},
			}, {Name: "bar"},
		},
		EventDefinitions: []*model.EventDefinitionInput{
			{
				Name: "foo",
				Spec: &model.EventSpecInput{FetchRequest: &model.FetchRequestInput{URL: "eventapi.foo.bar"}},
			}, {Name: "bar"},
		},
	}

	modelBundle := &model.Bundle{
		ID:                             id,
		TenantID:                       tenantID,
		ApplicationID:                  applicationID,
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &model.Auth{},
	}
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	modelFr := fixFetchRequest("api.foo.bar", model.APIFetchRequestReference, timestamp)

	testCases := []struct {
		Name                  string
		RepositoryFn          func() *automock.BundleRepository
		APIRepoFn             func() *automock.APIRepository
		EventAPIRepoFn        func() *automock.EventAPIRepository
		DocumentRepoFn        func() *automock.DocumentRepository
		FetchRequestRepoFn    func() *automock.FetchRequestRepository
		UIDServiceFn          func() *automock.UIDService
		FetchRequestServiceFn func() *automock.FetchRequestService
		Input                 model.BundleCreateInput
		ExpectedErr           error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Create", ctx, modelBundle).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "bar"}).Return(nil).Once()
				repo.On("Update", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.EventSpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "bar"}).Return(nil).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, mock.Anything).Return(nil).Times(2)
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()
				repo.On("Create", ctx, fixFetchRequest("eventapi.foo.bar", model.EventAPIFetchRequestReference, timestamp)).Return(nil).Once()
				repo.On("Create", ctx, fixFetchRequest("doc.foo.bar", model.DocumentFetchRequestReference, timestamp)).Return(nil).Once()
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error - Bundle creation",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Create", ctx, modelBundle).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - API creation",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Create", ctx, modelBundle).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(testErr).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Event creation",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Create", ctx, modelBundle).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "bar"}).Return(nil).Once()
				repo.On("Update", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.EventSpec{}}).Return(testErr).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - API Update",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Create", ctx, modelBundle).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				repo.On("Update", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(testErr).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Document creation",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Create", ctx, modelBundle).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "bar"}).Return(nil).Once()
				repo.On("Update", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()

				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.EventSpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "bar"}).Return(nil).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, mock.Anything).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()
				repo.On("Create", ctx, fixFetchRequest("eventapi.foo.bar", model.EventAPIFetchRequestReference, timestamp)).Return(nil).Once()
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Success when fetching API Spec failed",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Create", ctx, modelBundle).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "bar"}).Return(nil).Once()
				repo.On("Update", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.EventSpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "bar"}).Return(nil).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, mock.Anything).Return(nil).Times(2)
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()
				repo.On("Create", ctx, fixFetchRequest("eventapi.foo.bar", model.EventAPIFetchRequestReference, timestamp)).Return(nil).Once()
				repo.On("Create", ctx, fixFetchRequest("doc.foo.bar", model.DocumentFetchRequestReference, timestamp)).Return(nil).Once()
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Success - fetched api schema",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Create", ctx, modelBundle).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "bar"}).Return(nil).Once()
				repo.On("Update", ctx, &model.APIDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.APISpec{Data: &spec}}).Return(nil).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "foo", Spec: &model.EventSpec{}}).Return(nil).Once()
				repo.On("Create", ctx, &model.EventDefinition{ID: "foo", BundleID: "foo", Tenant: tenantID, Name: "bar"}).Return(nil).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("Create", ctx, mock.Anything).Return(nil).Times(2)
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, modelFr).Return(nil).Once()
				repo.On("Create", ctx, fixFetchRequest("eventapi.foo.bar", model.EventAPIFetchRequestReference, timestamp)).Return(nil).Once()
				repo.On("Create", ctx, fixFetchRequest("doc.foo.bar", model.DocumentFetchRequestReference, timestamp)).Return(nil).Once()
				return repo
			},
			FetchRequestServiceFn: func() *automock.FetchRequestService {
				svc := &automock.FetchRequestService{}
				svc.On("HandleAPISpec", ctx, modelFr).Return(&spec)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			uidService := testCase.UIDServiceFn()

			apiRepo := testCase.APIRepoFn()
			eventRepo := testCase.EventAPIRepoFn()
			documentRepo := testCase.DocumentRepoFn()
			frRepo := testCase.FetchRequestRepoFn()
			frSvc := testCase.FetchRequestServiceFn()
			svc := mp_bundle.NewService(repo, apiRepo, eventRepo, documentRepo, frRepo, uidService, frSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

			// when
			result, err := svc.Create(ctx, applicationID, testCase.Input)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.IsType(t, "string", result)
			}

			mock.AssertExpectationsForObjects(t, repo, apiRepo, eventRepo, documentRepo, frRepo, uidService)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := mp_bundle.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", model.BundleCreateInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	name := "bar"
	desc := "baz"

	modelInput := model.BundleUpdateInput{
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &model.AuthInput{},
	}

	inputBundleModel := mock.MatchedBy(func(pkg *model.Bundle) bool {
		return pkg.Name == modelInput.Name
	})

	bundleModel := &model.Bundle{
		ID:                             id,
		TenantID:                       tenantID,
		ApplicationID:                  "id",
		Name:                           name,
		Description:                    &desc,
		InstanceAuthRequestInputSchema: fixBasicSchema(),
		DefaultInstanceAuth:            &model.Auth{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.BundleRepository
		Input        model.BundleUpdateInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(bundleModel, nil).Once()
				repo.On("Update", ctx, inputBundleModel).Return(nil).Once()
				return repo
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(bundleModel, nil).Once()
				repo.On("Update", ctx, inputBundleModel).Return(testErr).Once()
				return repo
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("GetByID", ctx, tenantID, "foo").Return(nil, testErr).Once()
				return repo
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := mp_bundle.NewService(repo, nil, nil, nil, nil, nil, nil)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := mp_bundle.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		err := svc.Update(context.TODO(), "", model.BundleUpdateInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.BundleRepository
		Input        model.BundleCreateInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(testErr).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := mp_bundle.NewService(repo, nil, nil, nil, nil, nil, nil)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := mp_bundle.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Exist(t *testing.T) {
	// GIVEN
	testErr := errors.New("Test error")
	ctx := tenant.SaveToContext(context.TODO(), tenantID, externalTenantID)
	id := "foo"

	testCases := []struct {
		Name           string
		RepoFn         func() *automock.BundleRepository
		ExpectedError  error
		ExpectedOutput bool
	}{
		{
			Name: "Success",
			RepoFn: func() *automock.BundleRepository {
				pkgRepo := &automock.BundleRepository{}
				pkgRepo.On("Exists", ctx, tenantID, id).Return(true, nil).Once()
				return pkgRepo
			},
			ExpectedOutput: true,
		},
		{
			Name: "Error when getting Bundle",
			RepoFn: func() *automock.BundleRepository {
				pkgRepo := &automock.BundleRepository{}
				pkgRepo.On("Exists", ctx, tenantID, id).Return(false, testErr).Once()
				return pkgRepo
			},
			ExpectedError:  testErr,
			ExpectedOutput: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			pkgRepo := testCase.RepoFn()
			svc := mp_bundle.NewService(pkgRepo, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := svc.Exist(ctx, id)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			pkgRepo.AssertExpectations(t)
		})
	}

	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := mp_bundle.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.Exist(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	name := "foo"
	desc := "bar"

	pkg := fixBundleModel(t, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.BundleRepository
		Input              model.BundleCreateInput
		InputID            string
		ExpectedBundle     *model.Bundle
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(pkg, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedBundle:     pkg,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Bundle retrieval failed",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedBundle:     pkg,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := mp_bundle.NewService(repo, nil, nil, nil, nil, nil, nil)

			// when
			pkg, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedBundle, pkg)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := mp_bundle.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetForApplication(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	appID := "bar"
	name := "foo"
	desc := "bar"

	pkg := fixBundleModel(t, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.BundleRepository
		Input              model.BundleCreateInput
		InputID            string
		ApplicationID      string
		ExpectedBundle     *model.Bundle
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("GetForApplication", ctx, tenantID, id, appID).Return(pkg, nil).Once()
				return repo
			},
			InputID:            id,
			ApplicationID:      appID,
			ExpectedBundle:     pkg,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Bundle retrieval failed",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("GetForApplication", ctx, tenantID, id, appID).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ApplicationID:      appID,
			ExpectedBundle:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := mp_bundle.NewService(repo, nil, nil, nil, nil, nil, nil)

			// when
			document, err := svc.GetForApplication(ctx, testCase.InputID, testCase.ApplicationID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedBundle, document)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := mp_bundle.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.GetForApplication(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_GetByInstanceAuthID(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	appID := "bar"
	name := "foo"
	desc := "bar"

	pkg := fixBundleModel(t, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.BundleRepository
		Input              model.BundleCreateInput
		InstanceAuthID     string
		ExpectedBundle     *model.Bundle
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("GetByInstanceAuthID", ctx, tenantID, appID).Return(pkg, nil).Once()
				return repo
			},
			InstanceAuthID:     appID,
			ExpectedBundle:     pkg,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when Bundle retrieval failed",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("GetByInstanceAuthID", ctx, tenantID, appID).Return(nil, testErr).Once()
				return repo
			},
			InstanceAuthID:     appID,
			ExpectedBundle:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := mp_bundle.NewService(repo, nil, nil, nil, nil, nil, nil)

			// when
			document, err := svc.GetByInstanceAuthID(ctx, testCase.InstanceAuthID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedBundle, document)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := mp_bundle.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.GetForApplication(context.TODO(), "", "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}

func TestService_ListByApplicationID(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "bar"
	name := "foo"
	desc := "bar"

	bundles := []*model.Bundle{
		fixBundleModel(t, name, desc),
		fixBundleModel(t, name, desc),
		fixBundleModel(t, name, desc),
	}
	bundlePage := &model.BundlePage{
		Data:       bundles,
		TotalCount: len(bundles),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID, externalTenantID)

	testCases := []struct {
		Name               string
		PageSize           int
		RepositoryFn       func() *automock.BundleRepository
		ExpectedResult     *model.BundlePage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, applicationID, 2, after).Return(bundlePage, nil).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     bundlePage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				return repo
			},
			PageSize:           0,
			ExpectedResult:     bundlePage,
			ExpectedErrMessage: "page size must be between 1 and 100",
		},
		{
			Name: "Return error when page size is bigger than 100",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				return repo
			},
			PageSize:           101,
			ExpectedResult:     bundlePage,
			ExpectedErrMessage: "page size must be between 1 and 100",
		},
		{
			Name: "Returns error when Bundle listing failed",
			RepositoryFn: func() *automock.BundleRepository {
				repo := &automock.BundleRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, applicationID, 2, after).Return(nil, testErr).Once()
				return repo
			},
			PageSize:           2,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := mp_bundle.NewService(repo, nil, nil, nil, nil, nil, nil)

			// when
			docs, err := svc.ListByApplicationID(ctx, applicationID, testCase.PageSize, after)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, docs)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := mp_bundle.NewService(nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := svc.ListByApplicationID(context.TODO(), "", 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot read tenant from context")
	})
}
