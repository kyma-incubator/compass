package api_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	appID := "bar"
	name := "foo"
	desc := "bar"

	apiDefinition := fixModelAPIDefinition(id, appID, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.APIRepository
		Input              model.APIDefinitionInput
		InputID            string
		ExpectedDocument   *model.APIDefinition
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", id).Return(apiDefinition, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   apiDefinition,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition retrieval failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedDocument:   apiDefinition,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo, nil,nil)

			// when
			document, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedDocument, document)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	applicationID := "bar"
	name := "foo"
	desc := "bar"

	apiDefinitions := []*model.APIDefinition{
		fixModelAPIDefinition(id, applicationID, name, desc),
		fixModelAPIDefinition(id, applicationID, name, desc),
		fixModelAPIDefinition(id, applicationID, name, desc),
	}
	apiDefinitionPage := &model.APIDefinitionPage{
		Data:       apiDefinitions,
		TotalCount: len(apiDefinitions),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	first := 2
	after := "test"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.APIRepository
		InputPageSize      *int
		InputCursor        *string
		ExpectedResult     *model.APIDefinitionPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListByApplicationID", applicationID, &first, &after).Return(apiDefinitionPage, nil).Once()
				return repo
			},
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     apiDefinitionPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when APIDefinition listing failed",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("ListByApplicationID", applicationID, &first, &after).Return(nil, testErr).Once()
				return repo
			},
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo, nil,nil)

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
}

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	applicationID := "appid"
	name := "Foo"
	targetUrl := "https://test-url.com"

	modelInput := model.APIDefinitionInput{
		Name:      name,
		TargetURL: targetUrl,
		Spec:      &model.APISpecInput{},
		Version:   &model.VersionInput{},
	}

	modelAPIDefinition := &model.APIDefinition{
		ID:            id,
		ApplicationID: applicationID,
		Name:          name,
		TargetURL:     targetUrl,
		Spec:          &model.APISpec{},
		Version:       &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.APIRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.APIDefinitionInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", modelAPIDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("foo").Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", modelAPIDefinition).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("foo").Once()
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
			idSvc := testCase.UIDServiceFn()
			svc := api.NewService(repo, nil, idSvc)

			// when
			result, err := svc.Create(ctx, applicationID, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelInput := model.APIDefinitionInput{
		Name:      "Foo",
		TargetURL: "https://test-url.com",
		Spec: &model.APISpecInput{
			FetchRequest: &model.FetchRequestInput{
				Auth: &model.AuthInput{},
			},
		},
		DefaultAuth: &model.AuthInput{},
		Version:     &model.VersionInput{},
	}

	inputAPIDefinitionModel := mock.MatchedBy(func(api *model.APIDefinition) bool {
		return api.Name == modelInput.Name
	})

	frID := "fr_id"
	apiDefinitionModel := &model.APIDefinition{
		Name:          "Bar",
		ApplicationID: "id",
		TargetURL:     "https://test-url-updated.com",
		Spec: &model.APISpec{
			FetchRequestID: &frID,
		},
		DefaultAuth: &model.Auth{},
		Version:     &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.APIRepository
		Input        model.APIDefinitionInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", inputAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", inputAPIDefinitionModel).Return(testErr).Once()
				return repo
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", "foo").Return(nil, testErr).Once()
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

			svc := api.NewService(repo,nil, nil)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	frID := "fr_id"
	apiDefinitionModel := &model.APIDefinition{
		Name:          "Bar",
		ApplicationID: "id",
		TargetURL:     "https://test-url-updated.com",
		Spec: &model.APISpec{
			FetchRequestID: &frID,
		},
		Version: &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.APIRepository
		Input        model.APIDefinitionInput
		InputID      string
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", id).Return(apiDefinitionModel, nil).Once()
				repo.On("Delete", apiDefinitionModel).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", id).Return(apiDefinitionModel, nil).Once()
				repo.On("Delete", apiDefinitionModel).Return(testErr).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
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

			svc := api.NewService(repo, nil,nil)

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
}

func TestService_SetAPIAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	apiID := "foo"
	runtimeID := "bar"

	headers := map[string][]string{"header": {"hval1", "hval2"}}
	modelAuthInput := fixModelAuthInput(headers)

	modelRuntimeAuth := fixModelRuntimeAuth(runtimeID, modelAuthInput.ToAuth())

	modelAPIDefinition := &model.APIDefinition{
		ID:    apiID,
		Auths: []*model.RuntimeAuth{modelRuntimeAuth},
	}

	modelAPIDefinitionWithEmptyAuths := &model.APIDefinition{
		ID:    apiID,
		Auths: []*model.RuntimeAuth{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name                string
		RepositoryFn        func() *automock.APIRepository
		Input               model.AuthInput
		ExpectedRuntimeAuth *model.RuntimeAuth
		ExpectedErr         error
	}{
		{
			Name: "Success on replacing existing auth",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(modelAPIDefinition, nil).Once()
				repo.On("Update", modelAPIDefinition).Return(nil).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: modelRuntimeAuth,
			ExpectedErr:         nil,
		},
		{
			Name: "Success on appending new auth",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(modelAPIDefinitionWithEmptyAuths, nil).Once()
				repo.On("Update", modelAPIDefinitionWithEmptyAuths).Return(nil).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: modelRuntimeAuth,
			ExpectedErr:         nil,
		},
		{
			Name: "Set api auth failed on get",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(nil, testErr).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "Set api auth failed on update",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(modelAPIDefinition, nil).Once()
				repo.On("Update", modelAPIDefinition).Return(testErr).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo, nil,nil)

			// when
			result, err := svc.SetAPIAuth(ctx, apiID, runtimeID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeAuth, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteAPIAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")
	apiID := "foo"
	runtimeID := "bar"
	invalidRuntimeID := "invalid"
	headers := map[string][]string{"header": {"hval1", "hval2"}}
	modelAuthInput := fixModelAuthInput(headers)

	fixModelAPIDefinition := func(runtimeID string) *model.APIDefinition {
		return &model.APIDefinition{
			ID:    apiID,
			Auths: []*model.RuntimeAuth{fixModelRuntimeAuth(runtimeID, modelAuthInput.ToAuth())},
		}
	}

	fixModelApiDefinitionWithEmptyAuths := &model.APIDefinition{
		ID:    apiID,
		Auths: []*model.RuntimeAuth{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name                string
		RepositoryFn        func() *automock.APIRepository
		Input               model.AuthInput
		ExpectedRuntimeAuth *model.RuntimeAuth
		ExpectedErr         error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(fixModelAPIDefinition(runtimeID), nil).Once()
				repo.On("Update", fixModelApiDefinitionWithEmptyAuths).Return(nil).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: fixModelRuntimeAuth(runtimeID, modelAuthInput.ToAuth()),
			ExpectedErr:         nil,
		},
		{
			Name: "Delete api auth failed on get",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(nil, testErr).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "Delete api auth failed on update",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(fixModelAPIDefinition(runtimeID), nil).Once()
				repo.On("Update", fixModelApiDefinitionWithEmptyAuths).Return(testErr).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
		{
			Name: "No auth found",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(fixModelAPIDefinition(invalidRuntimeID), nil).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo, nil,nil)

			// when
			result, err := svc.DeleteAPIAuth(ctx, apiID, runtimeID)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeAuth, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
		})
	}
}

func TestService_RefetchAPISpec(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	apiID := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	dataBytes := "data"
	modelAPISpec := &model.APISpec{
		Data: &dataBytes,
	}

	modelAPIDefinition := &model.APIDefinition{
		Spec: modelAPISpec,
	}

	testCases := []struct {
		Name            string
		RepositoryFn    func() *automock.APIRepository
		ExpectedAPISpec *model.APISpec
		ExpectedErr     error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(modelAPIDefinition, nil).Once()
				return repo
			},
			ExpectedAPISpec: modelAPISpec,
			ExpectedErr:     nil,
		},
		{
			Name: "Get from repository error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", apiID).Return(nil, testErr).Once()
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

			svc := api.NewService(repo, nil,nil)

			// when
			result, err := svc.RefetchAPISpec(ctx, apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
		})
	}
}
