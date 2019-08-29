package api_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

var contextParam = txtest.CtxWithDBMatcher()

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
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn    func() *automock.APIService
		AppServiceFn func() *automock.ApplicationService
		ConverterFn  func() *automock.APIConverter
		ExpectedAPI  *graphql.APIDefinition
		ExpectedErr  error
	}{
		{
			Name: "Success",
			PersistenceFn: txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", contextParam, appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(modelAPI, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(true, nil)
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
			PersistenceFn: txtest.PersistenceContextThatDontExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(false, nil)
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
			PersistenceFn: txtest.PersistenceContextThatDontExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(false, testErr)
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
			PersistenceFn: txtest.PersistenceContextThatDontExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", contextParam, appId, *modelAPIInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(true, nil)
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
			PersistenceFn: txtest.PersistenceContextThatDontExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", contextParam, appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", contextParam, appId).Return(true, nil)
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
			persistTx := testCase.PersistenceFn()
			tx := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			appSvc := testCase.AppServiceFn()

			resolver := api.NewResolver(tx, svc, appSvc, converter, nil, nil)

			// when
			result, err := resolver.AddAPI(context.TODO(), appId, *gqlAPIInput)

			// then
			assert.Equal(t, testCase.ExpectedAPI, result)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persistTx.AssertExpectations(t)
			tx.AssertExpectations(t)
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

			resolver := api.NewResolver(nil, svc, nil, converter, nil, nil)

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
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn             func() *automock.APIService
		ConverterFn           func() *automock.APIConverter
		InputWebhookID        string
		InputAPI              graphql.APIDefinitionInput
		ExpectedAPIDefinition *graphql.APIDefinition
		ExpectedErr           error
	}{
		{
			Name: "Success",
			PersistenceFn: txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", contextParam, id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", contextParam, id).Return(modelAPIDefinition, nil).Once()
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
			PersistenceFn: txtest.PersistenceContextThatDontExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", contextParam, id, *modelAPIDefinitionInput).Return(testErr).Once()
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
			PersistenceFn: txtest.PersistenceContextThatDontExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", contextParam, id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", contextParam, id).Return(nil, testErr).Once()
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
			persistTx := testCase.PersistenceFn()
			tx := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(tx, svc, nil, converter, nil, nil)

			// when
			result, err := resolver.UpdateAPI(context.TODO(), id, *gqlAPIDefinitionInput)

			// then
			assert.Equal(t, testCase.ExpectedAPIDefinition, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persistTx.AssertExpectations(t)
			tx.AssertExpectations(t)
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
			resolver := api.NewResolver(nil, svc, nil, nil, conv, nil)

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
			resolver := api.NewResolver(nil, svc, nil, nil, conv, nil)

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
			resolver := api.NewResolver(nil, svc, nil, conv, nil, nil)

			// when
			result, err := resolver.RefetchAPISpec(context.TODO(), apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
		})
	}
}

func TestResolver_FetchRequest(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	url := "foo.bar"

	timestamp := time.Now()
	frModel := fixModelFetchRequest("foo", url, timestamp)
	frGQL := fixGQLFetchRequest(url, timestamp)
	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.APIService
		ConverterFn     func() *automock.FetchRequestConverter
		ExpectedResult  *graphql.FetchRequest
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("GetFetchRequest", contextParam, id).Return(frModel, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", frModel).Return(frGQL).Once()
				return conv
			},
			ExpectedResult: frGQL,
			ExpectedErr:    nil,
		},
		{
			Name:            "Doesn't exist",
			PersistenceFn:   txtest.PersistenceContextThatDontExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("GetFetchRequest", contextParam, id).Return(nil, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    nil,
		},
		{
			Name:            "Parent Object is nil",
			PersistenceFn:   txtest.PersistenceContextThatDontExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("GetFetchRequest", contextParam, id).Return(nil, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    nil,
		},
		{
			Name:            "Error",
			PersistenceFn:   txtest.PersistenceContextThatDontExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceed,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("GetFetchRequest", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(transact, svc, nil, nil, nil, converter)

			// when
			result, err := resolver.FetchRequest(context.TODO(), &graphql.APISpec{DefinitionID: id})

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			persistTx.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}
}
