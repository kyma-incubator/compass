package eventapi_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/stretchr/testify/assert"
)

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")
	id := "foo"
	eventAPIDefinition := fixMinModelEventAPIDefinition(id, "placeholder")

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.EventAPIRepository
		Input              model.EventAPIDefinitionInput
		InputID            string
		ExpectedDocument   *model.EventAPIDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(eventAPIDefinition, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   eventAPIDefinition,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when EventAPI retrieval failed",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   eventAPIDefinition,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := eventapi.NewService(repo, nil, nil)

			// when
			eventAPIDefinition, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedDocument, eventAPIDefinition)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := eventapi.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.Get(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
	})
}

func TestService_List(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	applicationID := "bar"

	eventAPIDefinitions := []*model.EventAPIDefinition{
		fixMinModelEventAPIDefinition(id, "placeholder"),
		fixMinModelEventAPIDefinition(id, "placeholder"),
		fixMinModelEventAPIDefinition(id, "placeholder"),
	}
	eventAPIDefinitionPage := &model.EventAPIDefinitionPage{
		Data:       eventAPIDefinitions,
		TotalCount: len(eventAPIDefinitions),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	first := 2
	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.EventAPIRepository
		InputPageSize      int
		InputCursor        string
		ExpectedResult     *model.EventAPIDefinitionPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, applicationID, first, after).Return(eventAPIDefinitionPage, nil).Once()
				return repo
			},
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     eventAPIDefinitionPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Return error when page size is less than 1",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			InputPageSize:      0,
			InputCursor:        after,
			ExpectedResult:     eventAPIDefinitionPage,
			ExpectedErrMessage: "page size must be between 1 and 100",
		},
		{
			Name: "Return error when page size is bigger than 100",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			InputPageSize:      101,
			InputCursor:        after,
			ExpectedResult:     eventAPIDefinitionPage,
			ExpectedErrMessage: "page size must be between 1 and 100",
		},
		{
			Name: "Returns error when EventAPI listing failed",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("ListByApplicationID", ctx, tenantID, applicationID, first, after).Return(nil, testErr).Once()
				return repo
			},
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := eventapi.NewService(repo, nil, nil)

			// when
			docs, err := svc.List(ctx, applicationID, testCase.InputPageSize, testCase.InputCursor)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, docs)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := eventapi.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.List(context.TODO(), "", 5, "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
	})
}

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	applicationID := "appid"
	name := "Foo"

	timestamp := time.Now()
	frID := "fr-id"
	frURL := "foo.bar"

	modelInput := model.EventAPIDefinitionInput{
		Name: name,
		Spec: &model.EventAPISpecInput{
			FetchRequest: &model.FetchRequestInput{
				URL: frURL,
			},
		},
		Version: &model.VersionInput{},
	}

	modelEventAPIDefinition := &model.EventAPIDefinition{
		ID:            id,
		Tenant:        tenantID,
		ApplicationID: applicationID,
		Name:          name,
		Spec:          &model.EventAPISpec{},
		Version:       &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.EventAPIRepository
		FetchRequestRepoFn func() *automock.FetchRequestRepository
		UIDServiceFn       func() *automock.UIDService
		Input              model.EventAPIDefinitionInput
		ExpectedErr        error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, modelEventAPIDefinition).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error - EventAPI Creation",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", ctx, modelEventAPIDefinition).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Error - Fetch Request Creation",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("Create", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			uidSvc := testCase.UIDServiceFn()

			svc := eventapi.NewService(repo, fetchRequestRepo, uidSvc)
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

			repo.AssertExpectations(t)
			fetchRequestRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := eventapi.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.Create(context.TODO(), "", model.EventAPIDefinitionInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	timestamp := time.Now()
	frID := "fr-id"
	frURL := "foo.bar"

	modelInput := model.EventAPIDefinitionInput{
		Name: "Foo",
		Spec: &model.EventAPISpecInput{
			FetchRequest: &model.FetchRequestInput{
				URL: frURL,
			},
		},
		Version: &model.VersionInput{},
	}

	inputEventAPIDefinitionModel := mock.MatchedBy(func(api *model.EventAPIDefinition) bool {
		return api.Name == modelInput.Name
	})

	eventAPIDefinitionModel := &model.EventAPIDefinition{
		Name:          "Bar",
		Tenant:        tenantID,
		ApplicationID: "id",
		Spec:          &model.EventAPISpec{},
		Version:       &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.EventAPIRepository
		FetchRequestRepoFn func() *automock.FetchRequestRepository
		UIDServiceFn       func() *automock.UIDService
		Input              model.EventAPIDefinitionInput
		InputID            string
		ExpectedErr        error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(eventAPIDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenantID, model.EventAPIFetchRequestReference, id).Return(nil).Once()
				repo.On("Create", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(eventAPIDefinitionModel, nil).Once()
				repo.On("Update", ctx, inputEventAPIDefinitionModel).Return(testErr).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("DeleteByReferenceObjectID", ctx, tenantID, model.EventAPIFetchRequestReference, id).Return(nil).Once()
				repo.On("Create", ctx, fixModelFetchRequest(frID, frURL, timestamp)).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(frID).Once()
				return svc
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, id).Return(nil, testErr).Once()
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
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			uidSvc := testCase.UIDServiceFn()

			svc := eventapi.NewService(repo, fetchRequestRepo, uidSvc)
			svc.SetTimestampGen(func() time.Time { return timestamp })

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
			fetchRequestRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := eventapi.NewService(nil, nil, nil)
		// WHEN
		err := svc.Update(context.TODO(), "", model.EventAPIDefinitionInput{})
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.EventAPIRepository
		Input        model.EventAPIDefinitionInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Delete", ctx, tenantID, id).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
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

			svc := eventapi.NewService(repo, nil, nil)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := eventapi.NewService(nil, nil, nil)
		// WHEN
		err := svc.Delete(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
	})
}

func TestService_RefetchAPISpec(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	apiID := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	dataBytes := "data"
	modelAPISpec := &model.EventAPISpec{
		Data: &dataBytes,
	}

	modelAPIDefinition := &model.EventAPIDefinition{
		Spec: modelAPISpec,
	}

	testCases := []struct {
		Name            string
		RepositoryFn    func() *automock.EventAPIRepository
		ExpectedAPISpec *model.EventAPISpec
		ExpectedErr     error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, apiID).Return(modelAPIDefinition, nil).Once()
				return repo
			},
			ExpectedAPISpec: modelAPISpec,
			ExpectedErr:     nil,
		},
		{
			Name: "Get from repository error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", ctx, tenantID, apiID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := eventapi.NewService(repo, nil, nil)

			// when
			result, err := svc.RefetchAPISpec(ctx, apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
		})
	}
	t.Run("Error when tenant not in context", func(t *testing.T) {
		svc := eventapi.NewService(nil, nil, nil)
		// WHEN
		_, err := svc.RefetchAPISpec(context.TODO(), "")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot read tenant from context")
	})
}

func TestService_GetFetchRequest(t *testing.T) {
	// given
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tenantID)

	testErr := errors.New("Test error")

	id := "foo"
	refID := "doc-id"
	frURL := "foo.bar"
	timestamp := time.Now()

	fetchRequestModel := fixModelFetchRequest(id, frURL, timestamp)

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.EventAPIRepository
		FetchRequestRepoFn   func() *automock.FetchRequestRepository
		ExpectedFetchRequest *model.FetchRequest
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Exists", ctx, tenantID, refID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.EventAPIFetchRequestReference, refID).Return(fetchRequestModel, nil).Once()
				return repo
			},
			ExpectedFetchRequest: fetchRequestModel,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Success - Not Found",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Exists", ctx, tenantID, refID).Return(true, nil).Once()
				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.EventAPIFetchRequestReference, refID).Return(nil, apperrors.NewNotFoundError("")).Once()
				return repo
			},
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   "",
		},
		{
			Name: "Error - Get FetchRequest",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Exists", ctx, tenantID, refID).Return(true, nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				repo.On("GetByReferenceObjectID", ctx, tenantID, model.EventAPIFetchRequestReference, refID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedFetchRequest: nil,
			ExpectedErrMessage:   testErr.Error(),
		},
		{
			Name: "Error - EventAPI doesn't exist",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Exists", ctx, tenantID, refID).Return(false, nil).Once()

				return repo
			},
			FetchRequestRepoFn: func() *automock.FetchRequestRepository {
				repo := &automock.FetchRequestRepository{}
				return repo
			},
			ExpectedErrMessage:   "EventAPI Definition with ID doc-id doesn't exist",
			ExpectedFetchRequest: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			fetchRequestRepo := testCase.FetchRequestRepoFn()
			svc := eventapi.NewService(repo, fetchRequestRepo, nil)

			// when
			l, err := svc.GetFetchRequest(ctx, refID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedFetchRequest)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			fetchRequestRepo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		svc := eventapi.NewService(nil, nil, nil)
		// when
		_, err := svc.GetFetchRequest(context.TODO(), "dd")
		assert.Equal(t, tenant.NoTenantError, err)
	})
}
