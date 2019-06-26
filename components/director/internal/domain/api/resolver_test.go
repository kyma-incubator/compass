package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/stretchr/testify/assert"
)

func TestResolver_AddAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelApi := fixModelAPIDefinition(id, "foo", "bar")
	gqlApi := fixGQLAPIDefinition(id, "foo", "bar")
	gqlApiInput := fixGQLAPIDefinitionInput(id, "foo", "bar")
	modelApiInput := fixModelAPIDefinitionInput(id, "foo", "bar")

	testCases := []struct {
		Name        string
		ServiceFn   func() *automock.APIService
		ConverterFn func() *automock.APIConverter
		ExpectedApi *graphql.APIDefinition
		ExpectedErr error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", context.TODO(), *modelApiInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelApi, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlApiInput).Return(modelApiInput).Once()
				conv.On("ToGraphQL", modelApi).Return(gqlApi).Once()
				return conv
			},
			ExpectedApi: gqlApi,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when API creation failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", context.TODO(), *modelApiInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlApiInput).Return(modelApiInput).Once()
				return conv
			},
			ExpectedApi: nil,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when API retrieval failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", context.TODO(), *modelApiInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlApiInput).Return(modelApiInput).Once()
				return conv
			},
			ExpectedApi: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(svc, converter)

			// when
			result, err := resolver.AddAPI(context.TODO(), "1", *gqlApiInput)

			// then
			assert.Equal(t, testCase.ExpectedApi, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelApiDefinition := fixModelAPIDefinition(id, "foo", "bar")
	gqlApiDefinition := fixGQLAPIDefinition(id, "foo", "bar")

	testCases := []struct {
		Name        string
		ServiceFn   func() *automock.APIService
		ConverterFn func() *automock.APIConverter
		ExpectedApi *graphql.APIDefinition
		ExpectedErr error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", context.TODO(), id).Return(modelApiDefinition, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", modelApiDefinition).Return(gqlApiDefinition).Once()
				return conv
			},
			ExpectedApi: gqlApiDefinition,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when API retrieval failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			ExpectedApi: nil,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when API deletion failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", context.TODO(), id).Return(modelApiDefinition, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", modelApiDefinition).Return(gqlApiDefinition).Once()
				return conv
			},
			ExpectedApi: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(svc, converter)

			// when
			result, err := resolver.DeleteAPI(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedApi, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	gqlApiDefinitionInput := fixGQLAPIDefinitionInput(id, "foo", "bar")
	modelApiDefinitionInput := fixModelAPIDefinitionInput(id, "foo", "bar")
	gqlApiDefinition := fixGQLAPIDefinition(id, "foo", "bar")
	modelApiDefinition := fixModelAPIDefinition(id, "foo", "bar")

	testCases := []struct {
		Name                  string
		ServiceFn             func() *automock.APIService
		ConverterFn           func() *automock.APIConverter
		InputWebhookID        string
		InputApi              graphql.APIDefinitionInput
		ExpectedApiDefinition *graphql.APIDefinition
		ExpectedErr           error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", context.TODO(), id, *modelApiDefinitionInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelApiDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlApiDefinitionInput).Return(modelApiDefinitionInput).Once()
				conv.On("ToGraphQL", modelApiDefinition).Return(gqlApiDefinition).Once()
				return conv
			},
			InputWebhookID:        id,
			InputApi:              *gqlApiDefinitionInput,
			ExpectedApiDefinition: gqlApiDefinition,
			ExpectedErr:           nil,
		},
		{
			Name: "Returns error when API update failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", context.TODO(), id, *modelApiDefinitionInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlApiDefinitionInput).Return(modelApiDefinitionInput).Once()
				return conv
			},
			InputWebhookID:        id,
			InputApi:              *gqlApiDefinitionInput,
			ExpectedApiDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name: "Returns error when API retrieval failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", context.TODO(), id, *modelApiDefinitionInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlApiDefinitionInput).Return(modelApiDefinitionInput).Once()
				return conv
			},
			InputWebhookID:        id,
			InputApi:              *gqlApiDefinitionInput,
			ExpectedApiDefinition: nil,
			ExpectedErr:           testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(svc, converter)

			// when
			result, err := resolver.UpdateAPI(context.TODO(), id, *gqlApiDefinitionInput)

			// then
			assert.Equal(t, testCase.ExpectedApiDefinition, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
