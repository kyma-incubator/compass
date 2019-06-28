package api_test

import (
	"context"
	"errors"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/mock"
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
	appId := "1"

	modelApi := fixModelAPIDefinition(id, appId, "name", "bar")
	gqlApi := fixGQLAPIDefinition(id, appId, "name", "bar")
	gqlApiInput := fixGQLAPIDefinitionInput("name", "foo", "bar")
	modelApiInput := fixModelAPIDefinitionInput("name", "foo", "bar")

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
				svc.On("Create", context.TODO(),mock.Anything, appId, *modelApiInput).Return(id, nil).Once()
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
				svc.On("Create", context.TODO(),mock.Anything, appId, *modelApiInput).Return("", testErr).Once()
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
				svc.On("Create", context.TODO(),mock.Anything, appId, *modelApiInput).Return(id, nil).Once()
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

			resolver := api.NewResolver(svc, converter,nil)

			// when
			result, err := resolver.AddAPI(context.TODO(), appId, *gqlApiInput)

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
	modelApiDefinition := fixModelAPIDefinition(id, "1", "foo", "bar")
	gqlApiDefinition := fixGQLAPIDefinition(id, "1", "foo", "bar")

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

			resolver := api.NewResolver(svc, converter,nil)

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
	gqlApiDefinition := fixGQLAPIDefinition(id, "1", "foo", "bar")
	modelApiDefinition := fixModelAPIDefinition(id, "1", "foo", "bar")

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

			resolver := api.NewResolver(svc, converter,nil)

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

func TestResolver_SetAPIAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	apiID := "apiID"
	runtimeID := "runtimeID"

	headers := map[string][]string{"header": {"hval1", "hval2"}}
	modelAuthInput := &model.AuthInput{
		AdditionalHeaders:     headers,
	}

	modelRuntimeAuth := &model.RuntimeAuth{
		RuntimeID: runtimeID,
		Auth:      modelAuthInput.ToAuth(),
	}

	httpHeaders := graphql.HttpHeaders(headers)
	gqlAuthInput := &graphql.AuthInput{
		AdditionalHeaders:     &httpHeaders,
	}

	gqlAuth := &graphql.Auth{
		AdditionalHeaders: &httpHeaders,
	}
	graphqlRuntimeAuth := graphql.RuntimeAuth{
		RuntimeID: runtimeID,
		Auth:      gqlAuth,
	}

	testCases := []struct {
		Name        string
		ServiceFn   func() *automock.APIService
		AuthConvFn   func() *automock.AuthConverter
		ExpectedRuntimeAuth *graphql.RuntimeAuth
		ExpectedErr error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("SetAPIAuth", context.TODO(), apiID,runtimeID, *modelAuthInput).Return(modelRuntimeAuth, nil).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL",gqlAuthInput).Return(modelAuthInput).Once()
				conv.On("ToGraphQL",modelRuntimeAuth.Auth).Return(gqlAuth).Once()
				return conv
			},
			ExpectedRuntimeAuth: &graphqlRuntimeAuth,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when setting up auth failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("SetAPIAuth", context.TODO(), apiID,runtimeID, *modelAuthInput).Return(nil, testErr).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL",gqlAuthInput).Return(modelAuthInput).Once()
				conv.On("ToGraphQL",modelRuntimeAuth.Auth).Return(gqlAuth).Once()
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			conv := testCase.AuthConvFn()
			resolver := api.NewResolver(svc, nil,conv)

			// when
			result, err := resolver.SetAPIAuth(context.TODO(), apiID,runtimeID, *gqlAuthInput)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeAuth, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
		})
	}
}
