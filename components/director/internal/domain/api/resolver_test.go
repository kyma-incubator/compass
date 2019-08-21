package api_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestResolver_AddAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	appId := "1"

	modelAPI := fixModelAPIDefinition(id, appId, "name", "bar")
	gqlAPI := fixGQLAPIDefinition(id, appId, "name", "bar")
	gqlAPIInput := fixGQLAPIDefinitionInput("name", "foo", "bar")
	modelAPIInput := fixModelAPIDefinitionInput("name", "foo", "bar")

	testCases := []struct {
		Name         string
		ServiceFn    func() *automock.APIService
		AppServiceFn func() *automock.ApplicationService
		ConverterFn  func() *automock.APIConverter
		ExpectedAPI  *graphql.APIDefinition
		ExpectedErr  error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", context.TODO(), appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelAPI, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
				return conv
			},
			ExpectedAPI: gqlAPI,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when application not exist",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: errors.New("Cannot add API to not existing Application"),
		},
		{
			Name: "Returns error when application existence check failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(false, testErr)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when API creation failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", context.TODO(), appId, *modelAPIInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when API retrieval failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", context.TODO(), appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			appSvc := testCase.AppServiceFn()

			resolver := api.NewResolver(nil, svc, appSvc, converter, nil)

			// when
			result, err := resolver.AddAPI(context.TODO(), appId, *gqlAPIInput)

			// then
			assert.Equal(t, testCase.ExpectedAPI, result)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			svc.AssertExpectations(t)
			appSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelAPIDefinition := fixModelAPIDefinition(id, "1", "foo", "bar")
	gqlAPIDefinition := fixGQLAPIDefinition(id, "1", "foo", "bar")

	testCases := []struct {
		Name        string
		ServiceFn   func() *automock.APIService
		ConverterFn func() *automock.APIConverter
		ExpectedAPI *graphql.APIDefinition
		ExpectedErr error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", context.TODO(), id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			ExpectedAPI: gqlAPIDefinition,
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
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when API deletion failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", context.TODO(), id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(nil, svc, nil, converter, nil)

			// when
			result, err := resolver.DeleteAPI(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedAPI, result)
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
	gqlAPIDefinitionInput := fixGQLAPIDefinitionInput(id, "foo", "bar")
	modelAPIDefinitionInput := fixModelAPIDefinitionInput(id, "foo", "bar")
	gqlAPIDefinition := fixGQLAPIDefinition(id, "1", "foo", "bar")
	modelAPIDefinition := fixModelAPIDefinition(id, "1", "foo", "bar")

	testCases := []struct {
		Name                  string
		ServiceFn             func() *automock.APIService
		ConverterFn           func() *automock.APIConverter
		InputWebhookID        string
		InputAPI              graphql.APIDefinitionInput
		ExpectedAPIDefinition *graphql.APIDefinition
		ExpectedErr           error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", context.TODO(), id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput).Once()
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: gqlAPIDefinition,
			ExpectedErr:           nil,
		},
		{
			Name: "Returns error when API update failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", context.TODO(), id, *modelAPIDefinitionInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput).Once()
				return conv
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name: "Returns error when API retrieval failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", context.TODO(), id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput).Once()
				return conv
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(nil, svc, nil, converter, nil)

			// when
			result, err := resolver.UpdateAPI(context.TODO(), id, *gqlAPIDefinitionInput)

			// then
			assert.Equal(t, testCase.ExpectedAPIDefinition, result)
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
	httpHeaders := graphql.HttpHeaders(headers)
	gqlAuth := &graphql.Auth{
		AdditionalHeaders: &httpHeaders,
	}

	modelAuthInput := fixModelAuthInput(headers)
	modelRuntimeAuth := fixModelRuntimeAuth(runtimeID, modelAuthInput.ToAuth())
	gqlAuthInput := fixGQLAuthInput(headers)
	graphqlRuntimeAuth := fixGQLRuntimeAuth(runtimeID, gqlAuth)

	testCases := []struct {
		Name                string
		ServiceFn           func() *automock.APIService
		AuthConvFn          func() *automock.AuthConverter
		ExpectedRuntimeAuth *graphql.RuntimeAuth
		ExpectedErr         error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("SetAPIAuth", context.TODO(), apiID, runtimeID, *modelAuthInput).Return(modelRuntimeAuth, nil).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				conv.On("ToGraphQL", modelRuntimeAuth.Auth).Return(gqlAuth).Once()
				return conv
			},
			ExpectedRuntimeAuth: graphqlRuntimeAuth,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when setting up auth failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("SetAPIAuth", context.TODO(), apiID, runtimeID, *modelAuthInput).Return(nil, testErr).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				conv.On("ToGraphQL", modelRuntimeAuth.Auth).Return(gqlAuth).Once()
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			svc := testCase.ServiceFn()
			conv := testCase.AuthConvFn()
			resolver := api.NewResolver(nil, svc, nil, nil, conv)

			// when
			result, err := resolver.SetAPIAuth(context.TODO(), apiID, runtimeID, *gqlAuthInput)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeAuth, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteAPIAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	apiID := "apiID"
	runtimeID := "runtimeID"

	headers := map[string][]string{"header": {"hval1", "hval2"}}
	httpHeaders := graphql.HttpHeaders(headers)
	gqlAuth := &graphql.Auth{
		AdditionalHeaders: &httpHeaders,
	}

	modelAuthInput := fixModelAuthInput(headers)
	modelRuntimeAuth := fixModelRuntimeAuth(runtimeID, modelAuthInput.ToAuth())
	graphqlRuntimeAuth := fixGQLRuntimeAuth(runtimeID, gqlAuth)

	testCases := []struct {
		Name                string
		ServiceFn           func() *automock.APIService
		AuthConvFn          func() *automock.AuthConverter
		ExpectedRuntimeAuth *graphql.RuntimeAuth
		ExpectedErr         error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("DeleteAPIAuth", context.TODO(), apiID, runtimeID).Return(modelRuntimeAuth, nil).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", modelRuntimeAuth.Auth).Return(gqlAuth).Once()
				return conv
			},
			ExpectedRuntimeAuth: graphqlRuntimeAuth,
			ExpectedErr:         nil,
		},
		{
			Name: "Returns error when deleting auth failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("DeleteAPIAuth", context.TODO(), apiID, runtimeID).Return(nil, testErr).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", modelRuntimeAuth.Auth).Return(gqlAuth).Once()
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			svc := testCase.ServiceFn()
			conv := testCase.AuthConvFn()
			resolver := api.NewResolver(nil, svc, nil, nil, conv)

			// when
			result, err := resolver.DeleteAPIAuth(context.TODO(), apiID, runtimeID)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeAuth, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
		})
	}
}

func TestResolver_RefetchAPISpec(t *testing.T) {
	// given
	testErr := errors.New("test error")

	apiID := "apiID"

	dataBytes := "data"
	modelAPISpec := &model.APISpec{
		Data: &dataBytes,
	}

	modelAPIDefinition := &model.APIDefinition{
		Spec: modelAPISpec,
	}

	clob := graphql.CLOB(dataBytes)
	gqlAPISpec := &graphql.APISpec{
		Data: &clob,
	}

	gqlAPIDefinition := &graphql.APIDefinition{
		Spec: gqlAPISpec,
	}

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.APIService
		ConvFn          func() *automock.APIConverter
		ExpectedAPISpec *graphql.APISpec
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("RefetchAPISpec", context.TODO(), apiID).Return(modelAPISpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			ExpectedAPISpec: gqlAPISpec,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when refetching api spec failed",
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("RefetchAPISpec", context.TODO(), apiID).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			svc := testCase.ServiceFn()
			conv := testCase.ConvFn()
			resolver := api.NewResolver(nil, svc, nil, conv, nil)

			// when
			result, err := resolver.RefetchAPISpec(context.TODO(), apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
		})
	}
}
