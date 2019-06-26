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

	modelInput := model.APIDefinitionInput{
		Name:          "Foo",
		ApplicationID: "id",
		TargetURL:     "https://test-url.com",
		Spec:          &model.APISpecInput{},
		Version:       &model.VersionInput{},
	}

	apiDefinitionModel := mock.MatchedBy(func(api *model.APIDefinition) bool {
		return api.Name == modelInput.Name && api.ApplicationID == api.ApplicationID
	})

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
				repo.On("Create", apiDefinitionModel).Return(nil).Once()
				return repo
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("Create", apiDefinitionModel).Return(testErr).Once()
				return repo
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo)

			// when
			result, err := svc.Create(ctx, testCase.Input)

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
		Name:          "Foo",
		ApplicationID: "id",
		TargetURL:     "https://test-url.com",
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
		Name               string
		RepositoryFn       func() *automock.APIRepository
		Input              model.APIDefinitionInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", inputApiDefinitionModel).Return(nil).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", "foo").Return(apiDefinitionModel, nil).Once()
				repo.On("Update", inputApiDefinitionModel).Return(testErr).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", "foo").Return(nil, testErr).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
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
		Name               string
		RepositoryFn       func() *automock.APIRepository
		Input              model.APIDefinitionInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", id).Return(apiDefinitionModel, nil).Once()
				repo.On("Delete", apiDefinitionModel).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", id).Return(apiDefinitionModel, nil).Once()
				repo.On("Delete", apiDefinitionModel).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := api.NewService(repo)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}
