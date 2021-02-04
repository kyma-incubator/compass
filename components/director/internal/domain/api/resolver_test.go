package api_test

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_AddAPIToBundle(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"

	modelAPI, spec := fixFullAPIDefinitionModel("test")
	gqlAPI := fixFullGQLAPIDefinition("test")
	gqlAPIInput := fixGQLAPIDefinitionInput("name", "foo", "bar")
	modelAPIInput, specInput := fixModelAPIDefinitionInput("name", "foo", "bar")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		BndlServiceFn   func() *automock.BundleService
		SpecServiceFn   func() *automock.SpecService
		ConverterFn     func() *automock.APIConverter
		ExpectedAPI     *graphql.APIDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, gqlAPI.ID).Return(&spec, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				conv.On("ToGraphQL", &modelAPI, &spec).Return(gqlAPI, nil).Once()
				return conv
			},
			ExpectedAPI: gqlAPI,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			BndlServiceFn: func() *automock.BundleService {
				return &automock.BundleService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when bundle not exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedAPI: nil,
			ExpectedErr: errors.New("cannot add API to not existing bundle"),
		},
		{
			Name:            "Returns error when bundle existence check failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(false, testErr)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when API creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelAPIInput, specInput).Return("", testErr).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when API retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, gqlAPI.ID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, gqlAPI.ID).Return(&spec, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				conv.On("ToGraphQL", &modelAPI, &spec).Return(gqlAPI, testErr).Once()
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
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelAPIInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPI, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIInput).Return(modelAPIInput, specInput, nil).Once()
				conv.On("ToGraphQL", &modelAPI, &spec).Return(gqlAPI, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, gqlAPI.ID).Return(&spec, nil).Once()
				return svc
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
			bndlSvc := testCase.BndlServiceFn()
			specSvc := testCase.SpecServiceFn()

			resolver := api.NewResolver(transact, svc, nil, nil, bndlSvc, converter, nil, specSvc, nil)

			// when
			result, err := resolver.AddAPIDefinitionToBundle(context.TODO(), bundleID, *gqlAPIInput)

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
			bndlSvc.AssertExpectations(t)
			specSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelAPIDefinition, spec := fixFullAPIDefinitionModel("test")
	gqlAPIDefinition := fixFullGQLAPIDefinition("test")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		ConverterFn     func() *automock.APIConverter
		SpecServiceFn   func() *automock.SpecService
		ExpectedAPI     *graphql.APIDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", &modelAPIDefinition, &spec).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			ExpectedAPI: gqlAPIDefinition,
			ExpectedErr: nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
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
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when API conversion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", &modelAPIDefinition, &spec).Return(nil, testErr).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when API deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", &modelAPIDefinition, &spec).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("ToGraphQL", &modelAPIDefinition, &spec).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&spec, nil).Once()
				return svc
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
			specService := testCase.SpecServiceFn()
			converter := testCase.ConverterFn()

			resolver := api.NewResolver(transact, svc, nil, nil, nil, converter, nil, specService, nil)

			// when
			result, err := resolver.DeleteAPIDefinition(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedAPI, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			svc.AssertExpectations(t)
			specService.AssertExpectations(t)
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
	modelAPIDefinitionInput, modelSpecInput := fixModelAPIDefinitionInput(id, "foo", "bar")
	gqlAPIDefinition := fixFullGQLAPIDefinition("test")
	modelAPIDefinition, modelSpec := fixFullAPIDefinitionModel("test")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                  string
		TransactionerFn       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn             func() *automock.APIService
		ConverterFn           func() *automock.APIConverter
		SpecServiceFn         func() *automock.SpecService
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
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				conv.On("ToGraphQL", &modelAPIDefinition, &modelSpec).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
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
				return &automock.APIService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when converting input to GraphQL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(nil, nil, testErr).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
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
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
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
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(nil, testErr).Once()
				return svc
			},
			InputWebhookID:        id,
			InputAPI:              *gqlAPIDefinitionInput,
			ExpectedAPIDefinition: nil,
			ExpectedErr:           testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				conv.On("ToGraphQL", &modelAPIDefinition, &modelSpec).Return(nil, testErr).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
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
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelAPIDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelAPIDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput).Return(modelAPIDefinitionInput, modelSpecInput, nil).Once()
				conv.On("ToGraphQL", &modelAPIDefinition, &modelSpec).Return(gqlAPIDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
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
			specService := testCase.SpecServiceFn()

			resolver := api.NewResolver(transact, svc, nil, nil, nil, converter, nil, specService, nil)

			// when
			result, err := resolver.UpdateAPIDefinition(context.TODO(), id, *gqlAPIDefinitionInput)

			// then
			assert.Equal(t, testCase.ExpectedAPIDefinition, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			specService.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_RefetchAPISpec(t *testing.T) {
	// given
	testErr := errors.New("test error")

	apiID := "apiID"
	specID := "specID"

	dataBytes := "data"
	modelSpec := &model.Spec{
		ID:   specID,
		Data: &dataBytes,
	}

	clob := graphql.CLOB(dataBytes)
	gqlAPISpec := &graphql.APISpec{
		Data: &clob,
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.SpecService
		ConvFn          func() *automock.SpecConverter
		ExpectedAPISpec *graphql.APISpec
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID).Return(modelSpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLAPISpec", modelSpec).Return(gqlAPISpec, nil).Once()
				return conv
			},
			ExpectedAPISpec: gqlAPISpec,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when getting spec failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when spec not found",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(nil, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     errors.Errorf("spec for API with id %q not found", apiID),
		},
		{
			Name:            "Returns error when refetching api spec failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error converting to GraphQL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID).Return(modelSpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLAPISpec", modelSpec).Return(nil, testErr).Once()
				return conv
			},
			ExpectedAPISpec: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, apiID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID).Return(modelSpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLAPISpec", modelSpec).Return(gqlAPISpec, nil).Once()
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
			resolver := api.NewResolver(transact, nil, nil, nil, nil, nil, nil, svc, conv)

			// when
			result, err := resolver.RefetchAPISpec(context.TODO(), apiID)

			// then
			assert.Equal(t, testCase.ExpectedAPISpec, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

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
				conv.On("ToGraphQL", frModel).Return(frGQL, nil).Once()
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

			resolver := api.NewResolver(transact, svc, nil, nil, nil, nil, converter, nil, nil)

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
