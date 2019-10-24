package api_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

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

func TestResolver_API(t *testing.T) {
	// given
	id := "bar"
	appId := "1"
	modelAPI := fixAPIDefinitionModel(id, appId, "name", "bar")
	gqlAPI := fixGQLAPIDefinition(id, appId, "name", "bar")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		ConverterFn     func() *automock.APIConverter
		InputID         string
		ExpectedAPI     *graphql.APIDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), "foo").Return(modelAPI, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
				return conv
			},
			InputID:     "foo",
			ExpectedAPI: gqlAPI,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			InputID:     "foo",
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns null when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), "foo").Return(nil, apperrors.NewNotFoundError("")).Once()

				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			InputID:     "foo",
			ExpectedAPI: nil,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}

				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			InputID:     "foo",
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), "foo").Return(modelAPI, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			InputID:     "foo",
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.API(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedAPI, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_AddAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	appId := "1"

	modelAPI := fixAPIDefinitionModel(id, appId, "name", "bar")
	gqlAPI := fixGQLAPIDefinition(id, appId, "name", "bar")
	gqlAPIInput := fixGQLAPIDefinitionInput("name", "foo", "bar")
	modelAPIInput := fixModelAPIDefinitionInput("name", "foo", "bar")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		AppServiceFn    func() *automock.ApplicationService
		ConverterFn     func() *automock.APIConverter
		ExpectedAPI     *graphql.APIDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
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
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when application not exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
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
			TransactionerFn: txGen.ThatDoesntExpectCommit,
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
			TransactionerFn: txGen.ThatDoesntExpectCommit,
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
			TransactionerFn: txGen.ThatDoesntExpectCommit,
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
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
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
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			appSvc := testCase.AppServiceFn()

			resolver := api.NewResolver(transact, svc, appSvc, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.AddAPI(context.TODO(), appId, *gqlAPIInput)

			// then
			assert.Equal(t, testCase.ExpectedAPI, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
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
	modelAPIDefinition := fixAPIDefinitionModel(id, "1", "foo", "bar")
	gqlAPIDefinition := fixGQLAPIDefinition(id, "1", "foo", "bar")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		ConverterFn     func() *automock.APIConverter
		ExpectedAPI     *graphql.APIDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
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
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
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
			Name:            "Returns error when API retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
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
			Name:            "Returns error when API deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
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
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelAPIDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.DeleteAPI(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedAPI, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transact.AssertExpectations(t)
			persist.AssertExpectations(t)
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
	modelAPIDefinition := fixAPIDefinitionModel(id, "1", "foo", "bar")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                  string
		TransactionerFn       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn             func() *automock.APIService
		ConverterFn           func() *automock.APIConverter
		InputWebhookID        string
		InputAPI              graphql.APIDefinitionInput
		ExpectedAPIDefinition *graphql.APIDefinition
		ExpectedErr           error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
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
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when API update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
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
			TransactionerFn: txGen.ThatDoesntExpectCommit,
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
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelAPIDefinition, nil).Once()
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
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.UpdateAPI(context.TODO(), id, *gqlAPIDefinitionInput)

			// then
			assert.Equal(t, testCase.ExpectedAPIDefinition, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
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

	modelAPIRtmAuth := fixModelAPIRtmAuth(rtmID, fixModelAuth())
	gqlAPIRtmAuth := fixGQLAPIRtmAuth(rtmID, fixGQLAuth())

	testErr := errors.New("this is a test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		RtmSvcFn         func() *automock.RuntimeService
		APIRtmAuthSvcFn  func() *automock.APIRuntimeAuthService
		APIRtmAuthConvFn func() *automock.APIRuntimeAuthConverter
		ExpectedOutput   *graphql.APIRuntimeAuth
		ExpectedError    error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			RtmSvcFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Get", txtest.CtxWithDBMatcher(), rtmID).Return(nil, nil).Once()
				return rtmSvc
			},
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("GetOrDefault", txtest.CtxWithDBMatcher(), apiID, rtmID).Return(modelAPIRtmAuth, nil).Once()
				return apiRtmAuthSvc
			},
			APIRtmAuthConvFn: func() *automock.APIRuntimeAuthConverter {
				apiRtmAuthConv := &automock.APIRuntimeAuthConverter{}
				apiRtmAuthConv.On("ToGraphQL", modelAPIRtmAuth).Return(gqlAPIRtmAuth).Once()
				return apiRtmAuthConv
			},
			ExpectedOutput: gqlAPIRtmAuth,
			ExpectedError:  nil,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			RtmSvcFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				return rtmSvc
			},
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				return apiRtmAuthSvc
			},
			APIRtmAuthConvFn: func() *automock.APIRuntimeAuthConverter {
				apiRtmAuthConv := &automock.APIRuntimeAuthConverter{}
				return apiRtmAuthConv
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
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				return apiRtmAuthSvc
			},
			APIRtmAuthConvFn: func() *automock.APIRuntimeAuthConverter {
				apiRtmAuthConv := &automock.APIRuntimeAuthConverter{}
				return apiRtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when getting API Runtime Auth",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			RtmSvcFn: func() *automock.RuntimeService {
				rtmSvc := &automock.RuntimeService{}
				rtmSvc.On("Get", txtest.CtxWithDBMatcher(), rtmID).Return(nil, nil).Once()
				return rtmSvc
			},
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("GetOrDefault", txtest.CtxWithDBMatcher(), apiID, rtmID).Return(nil, testErr).Once()
				return apiRtmAuthSvc
			},
			APIRtmAuthConvFn: func() *automock.APIRuntimeAuthConverter {
				apiRtmAuthConv := &automock.APIRuntimeAuthConverter{}
				return apiRtmAuthConv
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
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("GetOrDefault", txtest.CtxWithDBMatcher(), apiID, rtmID).Return(modelAPIRtmAuth, nil).Once()
				return apiRtmAuthSvc
			},
			APIRtmAuthConvFn: func() *automock.APIRuntimeAuthConverter {
				apiRtmAuthConv := &automock.APIRuntimeAuthConverter{}
				return apiRtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			rtmSvc := testCase.RtmSvcFn()
			apiRtmAuthSvc := testCase.APIRtmAuthSvcFn()
			apiRtmAuthConv := testCase.APIRtmAuthConvFn()
			persist, transact := testCase.TransactionerFn()

			resolver := api.NewResolver(transact, nil, nil, rtmSvc, apiRtmAuthSvc, nil, nil, nil, apiRtmAuthConv)

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
			apiRtmAuthSvc.AssertExpectations(t)
			apiRtmAuthConv.AssertExpectations(t)
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

	modelAPIRtmAuths := []model.APIRuntimeAuth{
		*fixModelAPIRtmAuth("r1", fixModelAuth()),
		*fixModelAPIRtmAuth("r2", fixModelAuth()),
		*fixModelAPIRtmAuth("r3", fixModelAuth()),
	}
	gqlAPIRtmAuths := []*graphql.APIRuntimeAuth{
		fixGQLAPIRtmAuth("r1", fixGQLAuth()),
		fixGQLAPIRtmAuth("r2", fixGQLAuth()),
		fixGQLAPIRtmAuth("r3", fixGQLAuth()),
	}

	testErr := errors.New("this is a test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIRtmAuthSvcFn  func() *automock.APIRuntimeAuthService
		APIRtmAuthConvFn func() *automock.APIRuntimeAuthConverter
		ExpectedOutput   []*graphql.APIRuntimeAuth
		ExpectedError    error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("ListForAllRuntimes", txtest.CtxWithDBMatcher(), apiID).Return(modelAPIRtmAuths, nil).Once()
				return apiRtmAuthSvc
			},
			APIRtmAuthConvFn: func() *automock.APIRuntimeAuthConverter {
				apiRtmAuthConv := &automock.APIRuntimeAuthConverter{}
				apiRtmAuthConv.On("ToGraphQL", &modelAPIRtmAuths[0]).Return(gqlAPIRtmAuths[0]).Once()
				apiRtmAuthConv.On("ToGraphQL", &modelAPIRtmAuths[1]).Return(gqlAPIRtmAuths[1]).Once()
				apiRtmAuthConv.On("ToGraphQL", &modelAPIRtmAuths[2]).Return(gqlAPIRtmAuths[2]).Once()
				return apiRtmAuthConv
			},
			ExpectedOutput: gqlAPIRtmAuths,
			ExpectedError:  nil,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				return apiRtmAuthSvc
			},
			APIRtmAuthConvFn: func() *automock.APIRuntimeAuthConverter {
				apiRtmAuthConv := &automock.APIRuntimeAuthConverter{}
				return apiRtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when listing for all runtimes",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("ListForAllRuntimes", txtest.CtxWithDBMatcher(), apiID).Return(nil, testErr).Once()
				return apiRtmAuthSvc
			},
			APIRtmAuthConvFn: func() *automock.APIRuntimeAuthConverter {
				apiRtmAuthConv := &automock.APIRuntimeAuthConverter{}
				return apiRtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("ListForAllRuntimes", txtest.CtxWithDBMatcher(), apiID).Return(modelAPIRtmAuths, nil).Once()
				return apiRtmAuthSvc
			},
			APIRtmAuthConvFn: func() *automock.APIRuntimeAuthConverter {
				apiRtmAuthConv := &automock.APIRuntimeAuthConverter{}
				return apiRtmAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			apiRtmAuthSvc := testCase.APIRtmAuthSvcFn()
			apiRtmAuthConv := testCase.APIRtmAuthConvFn()
			persist, transact := testCase.TransactionerFn()

			resolver := api.NewResolver(transact, nil, nil, nil, apiRtmAuthSvc, nil, nil, nil, apiRtmAuthConv)

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

			apiRtmAuthSvc.AssertExpectations(t)
			apiRtmAuthConv.AssertExpectations(t)
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
	modelAPIRtmAuth := fixModelAPIRtmAuth(runtimeID, modelAuthInput.ToAuth())
	gqlAuthInput := fixGQLAuthInput(headers)
	graphqlAPIRtmAuth := fixGQLAPIRtmAuth(runtimeID, gqlAuth)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIRtmAuthSvcFn    func() *automock.APIRuntimeAuthService
		AuthConvFn         func() *automock.AuthConverter
		ExpectedAPIRtmAuth *graphql.APIRuntimeAuth
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("Set", txtest.CtxWithDBMatcher(), apiID, runtimeID, *modelAuthInput).Return(nil).Once()
				apiRtmAuthSvc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelAPIRtmAuth, nil).Once()
				return apiRtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				conv.On("ToGraphQL", modelAPIRtmAuth.Value).Return(gqlAuth).Once()
				return conv
			},
			ExpectedAPIRtmAuth: graphqlAPIRtmAuth,
			ExpectedErr:        nil,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				return apiRtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedAPIRtmAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error when setting up auth failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("Set", txtest.CtxWithDBMatcher(), apiID, runtimeID, *modelAuthInput).Return(testErr).Once()
				return apiRtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				return conv
			},
			ExpectedAPIRtmAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error when getting api runtime auth",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("Set", txtest.CtxWithDBMatcher(), apiID, runtimeID, *modelAuthInput).Return(nil).Once()
				apiRtmAuthSvc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(nil, testErr).Once()
				return apiRtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				return conv
			},
			ExpectedAPIRtmAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error when input converted to nil",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				return apiRtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(nil).Once()
				return conv
			},
			ExpectedAPIRtmAuth: nil,
			ExpectedErr:        errors.New("object cannot be empty"),
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				apiRtmAuthSvc := &automock.APIRuntimeAuthService{}
				apiRtmAuthSvc.On("Set", txtest.CtxWithDBMatcher(), apiID, runtimeID, *modelAuthInput).Return(nil).Once()
				apiRtmAuthSvc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelAPIRtmAuth, nil).Once()
				return apiRtmAuthSvc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAuthInput).Return(modelAuthInput).Once()
				return conv
			},
			ExpectedAPIRtmAuth: nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			apiRtmAuthSvc := testCase.APIRtmAuthSvcFn()
			conv := testCase.AuthConvFn()
			persist, transact := testCase.TransactionerFn()

			resolver := api.NewResolver(transact, nil, nil, nil, apiRtmAuthSvc, nil, conv, nil, nil)

			// when
			result, err := resolver.SetAPIAuth(ctx, apiID, runtimeID, *gqlAuthInput)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedAPIRtmAuth, result)

			apiRtmAuthSvc.AssertExpectations(t)
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
	modelAPIRtmAuth := fixModelAPIRtmAuth(runtimeID, modelAuthInput.ToAuth())
	graphqlAPIRtmAuth := fixGQLAPIRtmAuth(runtimeID, gqlAuth)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name               string
		TransactionerFn    func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		APIRtmAuthSvcFn    func() *automock.APIRuntimeAuthService
		AuthConvFn         func() *automock.AuthConverter
		ExpectedAPIRtmAuth *graphql.APIRuntimeAuth
		ExpectedErr        error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				svc := &automock.APIRuntimeAuthService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelAPIRtmAuth, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(nil).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", modelAPIRtmAuth.Value).Return(gqlAuth).Once()
				return conv
			},
			ExpectedAPIRtmAuth: graphqlAPIRtmAuth,
			ExpectedErr:        nil,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				svc := &automock.APIRuntimeAuthService{}
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedAPIRtmAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error when getting api runtime auth",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				svc := &automock.APIRuntimeAuthService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(nil, testErr).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedAPIRtmAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error when deleting auth failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				svc := &automock.APIRuntimeAuthService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelAPIRtmAuth, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(testErr).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedAPIRtmAuth: nil,
			ExpectedErr:        testErr,
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			APIRtmAuthSvcFn: func() *automock.APIRuntimeAuthService {
				svc := &automock.APIRuntimeAuthService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(modelAPIRtmAuth, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), apiID, runtimeID).Return(nil).Once()
				return svc
			},
			AuthConvFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				return conv
			},
			ExpectedAPIRtmAuth: nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			apiRtmAuthSvc := testCase.APIRtmAuthSvcFn()
			authConv := testCase.AuthConvFn()
			persist, transact := testCase.TransactionerFn()
			resolver := api.NewResolver(transact, nil, nil, nil, apiRtmAuthSvc, nil, authConv, nil, nil)

			// when
			result, err := resolver.DeleteAPIAuth(ctx, apiID, runtimeID)

			// then
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedAPIRtmAuth, result)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			apiRtmAuthSvc.AssertExpectations(t)
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

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		ConvFn          func() *automock.APIConverter
		ExpectedAPISpec *graphql.APISpec
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("RefetchAPISpec", txtest.CtxWithDBMatcher(), apiID).Return(modelAPISpec, nil).Once()
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
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			ConvFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when refetching api spec failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("RefetchAPISpec", txtest.CtxWithDBMatcher(), apiID).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("RefetchAPISpec", txtest.CtxWithDBMatcher(), apiID).Return(modelAPISpec, nil).Once()
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
			persist, transact := testCase.TransactionerFn()
			resolver := api.NewResolver(transact, svc, nil, nil, nil, conv, nil, nil, nil)

			// when
			result, err := resolver.RefetchAPISpec(context.TODO(), apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
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

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		ConverterFn     func() *automock.FetchRequestConverter
		ExpectedResult  *graphql.FetchRequest
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
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
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			ConverterFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Doesn't exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
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
			TransactionerFn: txGen.ThatDoesntExpectCommit,
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
			TransactionerFn: txGen.ThatDoesntExpectCommit,
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
		{
			Name:            "Returns error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("GetFetchRequest", txtest.CtxWithDBMatcher(), id).Return(frModel, nil).Once()
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
			persist, transact := testCase.TransactionerFn()
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
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
		})
	}
}
