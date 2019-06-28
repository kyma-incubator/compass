package api_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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

	modelApiDefinition := &model.APIDefinition{
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
		Input        model.APIDefinitionInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", modelApiDefinition).Return(nil).Once()
				return repo
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", modelApiDefinition).Return(testErr).Once()
				return repo
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo)

			// when
			result, err := svc.Create(ctx, id, applicationID, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
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

	inputApiDefinitionModel := mock.MatchedBy(func(api *model.APIDefinition) bool {
		return api.Name == modelInput.Name
	})

	apiDefinitionModel := &model.APIDefinition{
		Name:          "Bar",
		ApplicationID: "id",
		TargetURL:     "https://test-url-updated.com",
		Spec: &model.APISpec{
			FetchRequest: &model.FetchRequest{},
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
				repo.On("Update", inputApiDefinitionModel).Return(nil).Once()
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
				repo.On("Update", inputApiDefinitionModel).Return(testErr).Once()
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

			svc := api.NewService(repo)

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

	apiDefinitionModel := &model.APIDefinition{
		Name:          "Bar",
		ApplicationID: "id",
		TargetURL:     "https://test-url-updated.com",
		Spec: &model.APISpec{
			FetchRequest: &model.FetchRequest{
				Mode: model.FetchModePackage,
			},
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

			svc := api.NewService(repo)

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
		ID:          apiID,
		Auths:       []*model.RuntimeAuth{modelRuntimeAuth},
		DefaultAuth: modelAuthInput.ToAuth(),
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
				repo.On("GetByID", apiID).Return(modelAPIDefinition, nil).Once()
				repo.On("Update", modelAPIDefinition).Return(nil).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: modelRuntimeAuth,
			ExpectedErr:         nil,
		},
		{
			Name: "Setting api auth failed",
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

			svc := api.NewService(repo)

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
	apiID := "foo"
	runtimeID := "bar"

	headers := map[string][]string{"header": {"hval1", "hval2"}}
	modelAuthInput := fixModelAuthInput(headers)

	modelRuntimeAuth := fixModelRuntimeAuth(runtimeID, modelAuthInput.ToAuth())
	modelAPIDefinition := &model.APIDefinition{
		ID:          apiID,
		Auths:       []*model.RuntimeAuth{modelRuntimeAuth},
		DefaultAuth: modelAuthInput.ToAuth(),
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
				repo.On("GetByID", apiID).Return(modelAPIDefinition, nil).Once()
				repo.On("Update", modelAPIDefinition).Return(nil).Once()
				return repo
			},
			Input:               *modelAuthInput,
			ExpectedRuntimeAuth: modelRuntimeAuth,
			ExpectedErr:         nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo)

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
	apiID := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	dataBytes := []byte("data")
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
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo)

			// when
			result, err := svc.RefetchAPISpec(ctx, apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
		})
	}
}
