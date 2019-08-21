package eventapi_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

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
	appID := "bar"
	name := "foo"
	desc := "bar"

	eventAPIDefinition := fixModelEventAPIDefinition(id, appID, name, desc)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

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
				repo.On("GetByID", id).Return(eventAPIDefinition, nil).Once()
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
				repo.On("GetByID", id).Return(nil, testErr).Once()
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

			svc := eventapi.NewService(repo, nil,  nil)

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

	eventAPIDefinitions := []*model.EventAPIDefinition{
		fixModelEventAPIDefinition(id, applicationID, name, desc),
		fixModelEventAPIDefinition(id, applicationID, name, desc),
		fixModelEventAPIDefinition(id, applicationID, name, desc),
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
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.EventAPIRepository
		InputPageSize      *int
		InputCursor        *string
		ExpectedResult     *model.EventAPIDefinitionPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("ListByApplicationID", applicationID, &first, &after).Return(eventAPIDefinitionPage, nil).Once()
				return repo
			},
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     eventAPIDefinitionPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when EventAPI listing failed",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
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
}

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	applicationID := "appid"
	name := "Foo"

	modelInput := model.EventAPIDefinitionInput{
		Name:    name,
		Spec:    &model.EventAPISpecInput{},
		Version: &model.VersionInput{},
	}

	modelEventAPIDefinition := &model.EventAPIDefinition{
		ID:            id,
		ApplicationID: applicationID,
		Name:          name,
		Spec:          &model.EventAPISpec{},
		Version:       &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.EventAPIRepository
		UIDServiceFn func() *automock.UIDService
		Input        model.EventAPIDefinitionInput
		ExpectedErr  error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", modelEventAPIDefinition).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("Create", modelEventAPIDefinition).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id).Once()
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
			uidSvc := testCase.UIDServiceFn()
			svc := eventapi.NewService(repo,nil,  uidSvc)

			// when
			result, err := svc.Create(ctx, applicationID, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelInput := model.EventAPIDefinitionInput{
		Name: "Foo",
		Spec: &model.EventAPISpecInput{
			FetchRequest: &model.FetchRequestInput{
				Auth: &model.AuthInput{},
			},
		},
		Version: &model.VersionInput{},
	}

	inputEventAPIDefinitionModel := mock.MatchedBy(func(api *model.EventAPIDefinition) bool {
		return api.Name == modelInput.Name
	})

	frID := "fr_id"
	eventAPIDefinitionModel := &model.EventAPIDefinition{
		Name:          "Bar",
		ApplicationID: "id",
		Spec: &model.EventAPISpec{
			FetchRequestID: &frID,
		},
		Version: &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

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
				repo.On("GetByID", "foo").Return(eventAPIDefinitionModel, nil).Once()
				repo.On("Update", inputEventAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", "foo").Return(eventAPIDefinitionModel, nil).Once()
				repo.On("Update", inputEventAPIDefinitionModel).Return(testErr).Once()
				return repo
			},
			InputID:     "foo",
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
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

			svc := eventapi.NewService(repo, nil, nil)

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
	eventAPIDefinitionModel := &model.EventAPIDefinition{
		Name:          "Bar",
		ApplicationID: "id",
		Spec: &model.EventAPISpec{
			FetchRequestID: &frID,
		},
		Version: &model.Version{},
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

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
				repo.On("GetByID", id).Return(eventAPIDefinitionModel, nil).Once()
				repo.On("Delete", eventAPIDefinitionModel).Return(nil).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: nil,
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("GetByID", id).Return(eventAPIDefinitionModel, nil).Once()
				repo.On("Delete", eventAPIDefinitionModel).Return(testErr).Once()
				return repo
			},
			InputID:     id,
			ExpectedErr: testErr,
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
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
}

func TestService_RefetchAPISpec(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	apiID := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

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
				repo.On("GetByID", apiID).Return(modelAPIDefinition, nil).Once()
				return repo
			},
			ExpectedAPISpec: modelAPISpec,
			ExpectedErr:     nil,
		},
		{
			Name: "Get from repository error",
			RepositoryFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
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

			svc := eventapi.NewService(repo, nil, nil)

			// when
			result, err := svc.RefetchAPISpec(ctx, apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			repo.AssertExpectations(t)
		})
	}
}
