package eventapi_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

var contextParam = mock.MatchedBy(func(ctx context.Context) bool {
	persistenceOp, err := persistence.FromCtx(ctx)
	return err == nil && persistenceOp != nil
})

func TestResolver_AddEventAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	appId := "1"

	modelAPI := fixModelEventAPIDefinition(id, appId, "name", "bar")
	gqlAPI := fixGQLEventAPIDefinition(id, appId, "name", "bar")
	gqlAPIInput := fixGQLEventAPIDefinitionInput("name", "foo", "bar")
	modelAPIInput := fixModelEventAPIDefinitionInput("name", "foo", "bar")

	testCases := []struct {
		Name         string
		ServiceFn    func() *automock.EventAPIService
		AppServiceFn func() *automock.ApplicationService
		ConverterFn  func() *automock.EventAPIConverter
		ExpectedAPI  *graphql.EventAPIDefinition
		ExpectedErr  error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("Create", context.TODO(), appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelAPI, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
				return conv
			},
			ExpectedAPI: gqlAPI,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when application not exist",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: errors.New("Cannot add EventAPI to not existing Application"),
		},
		{
			Name: "Returns error when application existence check failed",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(false, testErr)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when EventAPI creation failed",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("Create", context.TODO(), appId, *modelAPIInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput).Once()
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when EventAPI retrieval failed",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("Create", context.TODO(), appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), appId).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
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

			resolver := eventapi.NewResolver(nil, svc, appSvc, converter, nil)

			// when
			result, err := resolver.AddEventAPI(context.TODO(), appId, *gqlAPIInput)

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

func TestResolver_DeleteEventAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelAPIDefinition := fixModelEventAPIDefinition(id, "1", "foo", "bar")
	gqlAPIDefinition := fixGQLEventAPIDefinition(id, "1", "foo", "bar")

	testCases := []struct {
		Name        string
		ServiceFn   func() *automock.EventAPIService
		ConverterFn func() *automock.EventAPIConverter
		ExpectedAPI *graphql.EventAPIDefinition
		ExpectedErr error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("Get", context.TODO(), id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("ToGraphQL", modelAPIDefinition).Return(gqlAPIDefinition).Once()
				return conv
			},
			ExpectedAPI: gqlAPIDefinition,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when EventAPI retrieval failed",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when API deletion failed",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("Get", context.TODO(), id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
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

			resolver := eventapi.NewResolver(nil, svc, nil, converter, nil)

			// when
			result, err := resolver.DeleteEventAPI(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedAPI, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateEventAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	gqlAPIDefinitionInput := fixGQLEventAPIDefinitionInput(id, "foo", "bar")
	modelAPIDefinitionInput := fixModelEventAPIDefinitionInput(id, "foo", "bar")
	gqlAPIDefinition := fixGQLEventAPIDefinition(id, "1", "foo", "bar")
	modelAPIDefinition := fixModelEventAPIDefinition(id, "1", "foo", "bar")

	testCases := []struct {
		Name                  string
		ServiceFn             func() *automock.EventAPIService
		ConverterFn           func() *automock.EventAPIConverter
		InputWebhookID        string
		InputAPI              graphql.EventAPIDefinitionInput
		ExpectedAPIDefinition *graphql.EventAPIDefinition
		ExpectedErr           error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("Update", context.TODO(), id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
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
			Name: "Returns error when EventAPI update failed",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("Update", context.TODO(), id, *modelAPIDefinitionInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput).Once()
				return conv
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name: "Returns error when EventAPI retrieval failed",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("Update", context.TODO(), id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
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

			resolver := eventapi.NewResolver(nil, svc, nil, converter, nil)

			// when
			result, err := resolver.UpdateEventAPI(context.TODO(), id, *gqlAPIDefinitionInput)

			// then
			assert.Equal(t, testCase.ExpectedAPIDefinition, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_RefetchAPISpec(t *testing.T) {
	// given
	testErr := errors.New("test error")

	apiID := "apiID"

	dataBytes := "data"
	modelEventAPISpec := &model.EventAPISpec{
		Data: &dataBytes,
	}

	modelEventAPIDefinition := &model.EventAPIDefinition{
		Spec: modelEventAPISpec,
	}

	clob := graphql.CLOB(dataBytes)
	gqlEventAPISpec := &graphql.EventAPISpec{
		Data: &clob,
	}

	gqlEventAPIDefinition := &graphql.EventAPIDefinition{
		Spec: gqlEventAPISpec,
	}

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.EventAPIService
		ConvFn          func() *automock.EventAPIConverter
		ExpectedAPISpec *graphql.EventAPISpec
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("RefetchAPISpec", context.TODO(), apiID).Return(modelEventAPISpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("ToGraphQL", modelEventAPIDefinition).Return(gqlEventAPIDefinition).Once()
				return conv
			},
			ExpectedAPISpec: gqlEventAPISpec,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when refetching EventAPI spec failed",
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
				svc.On("RefetchAPISpec", context.TODO(), apiID).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.EventAPIConverter {
				conv := &automock.EventAPIConverter{}
				conv.On("ToGraphQL", modelEventAPIDefinition).Return(gqlEventAPIDefinition).Once()
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
			resolver := eventapi.NewResolver(nil, svc, nil, conv, nil)

			// when
			result, err := resolver.RefetchEventAPISpec(context.TODO(), apiID)

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
		ServiceFn       func() *automock.EventAPIService
		ConverterFn     func() *automock.FetchRequestConverter
		ExpectedResult  *graphql.FetchRequest
		ExpectedErr     error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
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
			Name: "Doesn't exist",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
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
			Name: "Error",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommited", persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.EventAPIService {
				svc := &automock.EventAPIService{}
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

			resolver := eventapi.NewResolver(transact, svc, nil, nil, converter)

			// when
			result, err := resolver.FetchRequest(context.TODO(), &graphql.EventAPISpec{DefinitionID: id})

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
