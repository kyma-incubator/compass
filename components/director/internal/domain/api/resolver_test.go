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
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
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
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.APIService
		AppServiceFn    func() *automock.ApplicationService
		ConverterFn     func() *automock.APIConverter
		ExpectedAPI     *graphql.APIDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelAPI, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), appId).Return(true, nil)
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
			Name:            "Returns error when application not exist",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), appId).Return(false, nil)
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
			Name:            "Returns error when application existence check failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), appId).Return(false, testErr)
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
			Name:            "Returns error when API creation failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, *modelAPIInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), appId).Return(true, nil)
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
			Name:            "Returns error when API retrieval failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, *modelAPIInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), appId).Return(true, nil)
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

			resolver := api.NewResolver(tx, svc, appSvc, nil, nil, converter, nil, nil, nil)

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

			resolver := api.NewResolver(nil, svc, nil, nil, nil, converter, nil, nil, nil)

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
		PersistenceFn         func() *persistenceautomock.PersistenceTx
		TransactionerFn       func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn             func() *automock.APIService
		ConverterFn           func() *automock.APIConverter
		InputWebhookID        string
		InputAPI              graphql.APIDefinitionInput
		ExpectedAPIDefinition *graphql.APIDefinition
		ExpectedErr           error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelAPIDefinition, nil).Once()
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
			Name:            "Returns error when API update failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput).Return(testErr).Once()
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
			Name:            "Returns error when API retrieval failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
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

			resolver := api.NewResolver(tx, svc, nil, nil, nil, converter, nil, nil, nil)

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

func TestResolver_Auth(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	rtmID := "foo"
	apiID := "bar"

	parentAPI := fixGQLAPIDefinition(apiID, "baz", "Test API", "API used by tests")

	modelRtmAuth := fixModelRuntimeAuth(rtmID, fixModelAuth())
	gqlRtmAuth := fixGQLRuntimeAuth(rtmID, fixGQLAuth())

	testErr := errors.New("this is a test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		RtmSvcFn        func() *automock.RuntimeService
		RtmAuthSvcFn    func() *automock.RuntimeAuthService
		RtmAuthConvFn   func() *automock.RuntimeAuthConverter
		ExpectedOutput  *graphql.RuntimeAuth
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			RtmSvcFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Get", txtest.CtxWithDBMatcher(), rtmID).Return(nil, nil).Once()
				return rtmSvc
			},
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("GetOrDefault", txtest.CtxWithDBMatcher(), apiID, rtmID).Return(modelRtmAuth, nil).Once()
				return rtmAuthSvc
			},
			RtmAuthConvFn: func() *automock.RuntimeAuthConverter {
				rtmAuthConv := &automock.RuntimeAuthConverter{}
				rtmAuthConv.On("ToGraphQL", modelRtmAuth).Return(gqlRtmAuth).Once()
				return rtmAuthConv
			},
			ExpectedOutput: gqlRtmAuth,
			ExpectedError:  nil,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			RtmSvcFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				return rtmSvc
			},
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				return rtmAuthSvc
			},
			RtmAuthConvFn: func() *automock.RuntimeAuthConverter {
				rtmAuthConv := &automock.RuntimeAuthConverter{}
				return rtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when getting Runtime",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RtmSvcFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Get", txtest.CtxWithDBMatcher(), rtmID).Return(nil, testErr).Once()
				return rtmSvc
			},
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				return rtmAuthSvc
			},
			RtmAuthConvFn: func() *automock.RuntimeAuthConverter {
				rtmAuthConv := &automock.RuntimeAuthConverter{}
				return rtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when getting Runtime Auth",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RtmSvcFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Get", txtest.CtxWithDBMatcher(), rtmID).Return(nil, nil).Once()
				return rtmSvc
			},
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("GetOrDefault", txtest.CtxWithDBMatcher(), apiID, rtmID).Return(nil, testErr).Once()
				return rtmAuthSvc
			},
			RtmAuthConvFn: func() *automock.RuntimeAuthConverter {
				rtmAuthConv := &automock.RuntimeAuthConverter{}
				return rtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			RtmSvcFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Get", txtest.CtxWithDBMatcher(), rtmID).Return(nil, nil).Once()
				return rtmSvc
			},
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("GetOrDefault", txtest.CtxWithDBMatcher(), apiID, rtmID).Return(modelRtmAuth, nil).Once()
				return rtmAuthSvc
			},
			RtmAuthConvFn: func() *automock.RuntimeAuthConverter {
				rtmAuthConv := &automock.RuntimeAuthConverter{}
				return rtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			rtmSvc := testCase.RtmSvcFn()
			rtmAuthSvc := testCase.RtmAuthSvcFn()
			rtmAuthConv := testCase.RtmAuthConvFn()
			persist, transact := testCase.TransactionerFn()

			resolver := api.NewResolver(transact, nil, nil, rtmSvc, rtmAuthSvc, nil, nil, nil, rtmAuthConv)

			// WHEN
			ra, err := resolver.Auth(ctx, parentAPI, rtmID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, ra)

			rtmSvc.AssertExpectations(t)
			rtmAuthSvc.AssertExpectations(t)
			rtmAuthConv.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}
}

func TestResolver_Auths(t *testing.T) {
	// GIVEN
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	apiID := "bar"

	parentAPI := fixGQLAPIDefinition(apiID, "baz", "Test API", "API used by tests")

	modelRtmAuths := []model.RuntimeAuth{
		*fixModelRuntimeAuth("r1", fixModelAuth()),
		*fixModelRuntimeAuth("r2", fixModelAuth()),
		*fixModelRuntimeAuth("r3", fixModelAuth()),
	}
	gqlRtmAuths := []*graphql.RuntimeAuth{
		fixGQLRuntimeAuth("r1", fixGQLAuth()),
		fixGQLRuntimeAuth("r2", fixGQLAuth()),
		fixGQLRuntimeAuth("r3", fixGQLAuth()),
	}

	testErr := errors.New("this is a test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		RtmAuthSvcFn    func() *automock.RuntimeAuthService
		RtmAuthConvFn   func() *automock.RuntimeAuthConverter
		ExpectedOutput  []*graphql.RuntimeAuth
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("ListForAllRuntimes", txtest.CtxWithDBMatcher(), apiID).Return(modelRtmAuths, nil).Once()
				return rtmAuthSvc
			},
			RtmAuthConvFn: func() *automock.RuntimeAuthConverter {
				rtmAuthConv := &automock.RuntimeAuthConverter{}
				rtmAuthConv.On("ToGraphQL", &modelRtmAuths[0]).Return(gqlRtmAuths[0]).Once()
				rtmAuthConv.On("ToGraphQL", &modelRtmAuths[1]).Return(gqlRtmAuths[1]).Once()
				rtmAuthConv.On("ToGraphQL", &modelRtmAuths[2]).Return(gqlRtmAuths[2]).Once()
				return rtmAuthConv
			},
			ExpectedOutput: gqlRtmAuths,
			ExpectedError:  nil,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				return rtmAuthSvc
			},
			RtmAuthConvFn: func() *automock.RuntimeAuthConverter {
				rtmAuthConv := &automock.RuntimeAuthConverter{}
				return rtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when listing for all runtimes",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("ListForAllRuntimes", txtest.CtxWithDBMatcher(), apiID).Return(nil, testErr).Once()
				return rtmAuthSvc
			},
			RtmAuthConvFn: func() *automock.RuntimeAuthConverter {
				rtmAuthConv := &automock.RuntimeAuthConverter{}
				return rtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("ListForAllRuntimes", txtest.CtxWithDBMatcher(), apiID).Return(modelRtmAuths, nil).Once()
				return rtmAuthSvc
			},
			RtmAuthConvFn: func() *automock.RuntimeAuthConverter {
				rtmAuthConv := &automock.RuntimeAuthConverter{}
				return rtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			rtmAuthSvc := testCase.RtmAuthSvcFn()
			rtmAuthConv := testCase.RtmAuthConvFn()
			persist, transact := testCase.TransactionerFn()

			resolver := api.NewResolver(transact, nil, nil, nil, rtmAuthSvc, nil, nil, nil, rtmAuthConv)

			// WHEN
			ra, err := resolver.Auths(ctx, parentAPI)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, ra)

			rtmAuthSvc.AssertExpectations(t)
			rtmAuthConv.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
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

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	modelAuthInput := fixModelAuthInput(headers)
	modelRuntimeAuth := fixModelRuntimeAuth(runtimeID, modelAuthInput.ToAuth())
	gqlAuthInput := fixGQLAuthInput(headers)
	graphqlRuntimeAuth := fixGQLRuntimeAuth(runtimeID, gqlAuth)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		RtmAuthSvcFn        func() *automock.RuntimeAuthService
		AuthConvFn          func() *automock.AuthConverter
		ExpectedRuntimeAuth *graphql.RuntimeAuth
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("Set", txtest.CtxWithDBMatcher(), apiID, runtimeID, *modelAuthInput).Return(nil).Once()
				rtmAuthSvc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelRuntimeAuth, nil).Once()
				return rtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				conv.On("ToGraphQL", modelRuntimeAuth.Value).Return(gqlAuth).Once()
				return conv
			},
			ExpectedRuntimeAuth: graphqlRuntimeAuth,
			ExpectedErr:         nil,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				return rtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Error when setting up auth failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("Set", txtest.CtxWithDBMatcher(), apiID, runtimeID, *modelAuthInput).Return(testErr).Once()
				return rtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Error when getting runtime auth",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("Set", txtest.CtxWithDBMatcher(), apiID, runtimeID, *modelAuthInput).Return(nil).Once()
				rtmAuthSvc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(nil, testErr).Once()
				return rtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Error when input converted to nil",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				return rtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(nil).Once()
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         errors.New("object cannot be empty"),
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				rtmAuthSvc := &automock.RuntimeAuthService{}
				rtmAuthSvc.On("Set", txtest.CtxWithDBMatcher(), apiID, runtimeID, *modelAuthInput).Return(nil).Once()
				rtmAuthSvc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelRuntimeAuth, nil).Once()
				return rtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			rtmAuthSvc := testCase.RtmAuthSvcFn()
			conv := testCase.AuthConvFn()
			persist, transact := testCase.TransactionerFn()

			resolver := api.NewResolver(transact, nil, nil, nil, rtmAuthSvc, nil, conv, nil, nil)

			// when
			result, err := resolver.SetAPIAuth(ctx, apiID, runtimeID, *gqlAuthInput)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedRuntimeAuth, result)

			rtmAuthSvc.AssertExpectations(t)
			conv.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteAPIAuth(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	apiID := "apiID"
	runtimeID := "runtimeID"

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	headers := map[string][]string{"header": {"hval1", "hval2"}}
	httpHeaders := graphql.HttpHeaders(headers)
	gqlAuth := &graphql.Auth{
		AdditionalHeaders: &httpHeaders,
	}

	modelAuthInput := fixModelAuthInput(headers)
	modelRuntimeAuth := fixModelRuntimeAuth(runtimeID, modelAuthInput.ToAuth())
	graphqlRuntimeAuth := fixGQLRuntimeAuth(runtimeID, gqlAuth)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                string
		TransactionerFn     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		RtmAuthSvcFn        func() *automock.RuntimeAuthService
		AuthConvFn          func() *automock.AuthConverter
		ExpectedRuntimeAuth *graphql.RuntimeAuth
		ExpectedErr         error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				svc := &automock.RuntimeAuthService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelRuntimeAuth, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(nil).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", modelRuntimeAuth.Value).Return(gqlAuth).Once()
				return conv
			},
			ExpectedRuntimeAuth: graphqlRuntimeAuth,
			ExpectedErr:         nil,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				svc := &automock.RuntimeAuthService{}
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Error when getting runtime auth",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				svc := &automock.RuntimeAuthService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(nil, testErr).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Error when deleting auth failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				svc := &automock.RuntimeAuthService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelRuntimeAuth, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(testErr).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			RtmAuthSvcFn: func() *automock.RuntimeAuthService {
				svc := &automock.RuntimeAuthService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelRuntimeAuth, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(nil).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedRuntimeAuth: nil,
			ExpectedErr:         testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			rtmAuthSvc := testCase.RtmAuthSvcFn()
			authConv := testCase.AuthConvFn()
			persist, transact := testCase.TransactionerFn()
			resolver := api.NewResolver(transact, nil, nil, nil, rtmAuthSvc, nil, authConv, nil, nil)

			// when
			result, err := resolver.DeleteAPIAuth(ctx, apiID, runtimeID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedRuntimeAuth, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			rtmAuthSvc.AssertExpectations(t)
			authConv.AssertExpectations(t)
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
			resolver := api.NewResolver(nil, svc, nil, nil, nil, conv, nil, nil, nil)

			// when
			result, err := resolver.RefetchAPISpec(context.TODO(), apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			conv.AssertExpectations(t)
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
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), id).Return(frModel, nil).Once()
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
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), id).Return(nil, nil).Once()
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
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), id).Return(nil, nil).Once()
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
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
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

			resolver := api.NewResolver(transact, svc, nil, nil, nil, nil, nil, converter, nil)

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
