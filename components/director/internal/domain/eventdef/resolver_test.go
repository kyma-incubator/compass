package eventdef_test

import (
	"context"
	"testing"
	"time"

	event "github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_AddEventToBundle(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"

	modelEvent, spec := fixFullEventDefinitionModel("test")
	gqlEvent := fixFullGQLEventDefinition("test")
	gqlEventInput := fixGQLEventDefinitionInput("name", "foo", "bar")
	modelEventInput, specInput := fixModelEventDefinitionInput("name", "foo", "bar")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.EventDefService
		BndlServiceFn   func() *automock.BundleService
		SpecServiceFn   func() *automock.SpecService
		ConverterFn     func() *automock.EventDefConverter
		ExpectedEvent   *graphql.EventDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelEventInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEvent, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, gqlEvent.ID).Return(&spec, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventInput).Return(modelEventInput, specInput, nil).Once()
				conv.On("ToGraphQL", &modelEvent, &spec).Return(gqlEvent, nil).Once()
				return conv
			},
			ExpectedEvent: gqlEvent,
			ExpectedErr:   nil,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefService {
				return &automock.EventDefService{}
			},
			BndlServiceFn: func() *automock.BundleService {
				return &automock.BundleService{}
			},
			ConverterFn: func() *automock.EventDefConverter {
				return &automock.EventDefConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when bundle not exist",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				return &automock.EventDefService{}
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventInput).Return(modelEventInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedEvent: nil,
			ExpectedErr:   errors.New("cannot add Event Definition to not existing Bundle"),
		},
		{
			Name:            "Returns error when bundle existence check failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(false, testErr)
				return appSvc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventInput).Return(modelEventInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when Event creation failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelEventInput, specInput).Return("", testErr).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventInput).Return(modelEventInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when Event retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelEventInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventInput).Return(modelEventInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelEventInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEvent, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventInput).Return(modelEventInput, specInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, gqlEvent.ID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelEventInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEvent, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, gqlEvent.ID).Return(&spec, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventInput).Return(modelEventInput, specInput, nil).Once()
				conv.On("ToGraphQL", &modelEvent, &spec).Return(gqlEvent, testErr).Once()
				return conv
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("CreateInBundle", txtest.CtxWithDBMatcher(), bundleID, *modelEventInput, specInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEvent, nil).Once()
				return svc
			},
			BndlServiceFn: func() *automock.BundleService {
				appSvc := &automock.BundleService{}
				appSvc.On("Exist", txtest.CtxWithDBMatcher(), bundleID).Return(true, nil)
				return appSvc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventInput).Return(modelEventInput, specInput, nil).Once()
				conv.On("ToGraphQL", &modelEvent, &spec).Return(gqlEvent, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, gqlEvent.ID).Return(&spec, nil).Once()
				return svc
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
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

			resolver := event.NewResolver(transact, svc, nil, bndlSvc, converter, nil, specSvc, nil)

			// when
			result, err := resolver.AddEventDefinitionToBundle(context.TODO(), bundleID, *gqlEventInput)

			// then
			assert.Equal(t, testCase.ExpectedEvent, result)
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

func TestResolver_DeleteEvent(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	modelEventDefinition, spec := fixFullEventDefinitionModel("test")
	gqlEventDefinition := fixFullGQLEventDefinition("test")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.EventDefService
		ConverterFn     func() *automock.EventDefConverter
		SpecServiceFn   func() *automock.SpecService
		ExpectedEvent   *graphql.EventDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEventDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("ToGraphQL", &modelEventDefinition, &spec).Return(gqlEventDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			ExpectedEvent: gqlEventDefinition,
			ExpectedErr:   nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefService {
				return &automock.EventDefService{}
			},
			ConverterFn: func() *automock.EventDefConverter {
				return &automock.EventDefConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when Event retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				return &automock.EventDefConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEventDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				return &automock.EventDefConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinition.ID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when Event conversion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEventDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("ToGraphQL", &modelEventDefinition, &spec).Return(nil, testErr).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Returns error when Event deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEventDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("ToGraphQL", &modelEventDefinition, &spec).Return(gqlEventDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEventDefinition, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("ToGraphQL", &modelEventDefinition, &spec).Return(gqlEventDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinition.ID).Return(&spec, nil).Once()
				return svc
			},
			ExpectedEvent: nil,
			ExpectedErr:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			specService := testCase.SpecServiceFn()
			converter := testCase.ConverterFn()

			resolver := event.NewResolver(transact, svc, nil, nil, converter, nil, specService, nil)

			// when
			result, err := resolver.DeleteEventDefinition(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedEvent, result)
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

func TestResolver_UpdateEvent(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "bar"
	gqlEventDefinitionInput := fixGQLEventDefinitionInput(id, "foo", "bar")
	modelEventDefinitionInput, modelSpecInput := fixModelEventDefinitionInput(id, "foo", "bar")
	gqlEventDefinition := fixFullGQLEventDefinition("test")
	modelEventDefinition, modelSpec := fixFullEventDefinitionModel("test")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn               func() *automock.EventDefService
		ConverterFn             func() *automock.EventDefConverter
		SpecServiceFn           func() *automock.SpecService
		InputWebhookID          string
		InputEvent              graphql.EventDefinitionInput
		ExpectedEventDefinition *graphql.EventDefinition
		ExpectedErr             error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelEventDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEventDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventDefinitionInput).Return(modelEventDefinitionInput, modelSpecInput, nil).Once()
				conv.On("ToGraphQL", &modelEventDefinition, &modelSpec).Return(gqlEventDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
			},
			InputWebhookID:          id,
			InputEvent:              *gqlEventDefinitionInput,
			ExpectedEventDefinition: gqlEventDefinition,
			ExpectedErr:             nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventDefService {
				return &automock.EventDefService{}
			},
			ConverterFn: func() *automock.EventDefConverter {
				return &automock.EventDefConverter{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			InputWebhookID:          id,
			InputEvent:              *gqlEventDefinitionInput,
			ExpectedEventDefinition: nil,
			ExpectedErr:             testErr,
		},
		{
			Name:            "Returns error when converting input to GraphQL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				return &automock.EventDefService{}
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventDefinitionInput).Return(nil, nil, testErr).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			InputWebhookID:          id,
			InputEvent:              *gqlEventDefinitionInput,
			ExpectedEventDefinition: nil,
			ExpectedErr:             testErr,
		},
		{
			Name:            "Returns error when Event update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelEventDefinitionInput, modelSpecInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventDefinitionInput).Return(modelEventDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			InputWebhookID:          id,
			InputEvent:              *gqlEventDefinitionInput,
			ExpectedEventDefinition: nil,
			ExpectedErr:             testErr,
		},
		{
			Name:            "Returns error when Event retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelEventDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventDefinitionInput).Return(modelEventDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			InputWebhookID:          id,
			InputEvent:              *gqlEventDefinitionInput,
			ExpectedEventDefinition: nil,
			ExpectedErr:             testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelEventDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEventDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventDefinitionInput).Return(modelEventDefinitionInput, modelSpecInput, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinition.ID).Return(nil, testErr).Once()
				return svc
			},
			InputWebhookID:          id,
			InputEvent:              *gqlEventDefinitionInput,
			ExpectedEventDefinition: nil,
			ExpectedErr:             testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelEventDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEventDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventDefinitionInput).Return(modelEventDefinitionInput, modelSpecInput, nil).Once()
				conv.On("ToGraphQL", &modelEventDefinition, &modelSpec).Return(nil, testErr).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
			},
			InputWebhookID:          id,
			InputEvent:              *gqlEventDefinitionInput,
			ExpectedEventDefinition: nil,
			ExpectedErr:             testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, *modelEventDefinitionInput, modelSpecInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(&modelEventDefinition, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventDefConverter {
				conv := &automock.EventDefConverter{}
				conv.On("InputFromGraphQL", gqlEventDefinitionInput).Return(modelEventDefinitionInput, modelSpecInput, nil).Once()
				conv.On("ToGraphQL", &modelEventDefinition, &modelSpec).Return(gqlEventDefinition, nil).Once()
				return conv
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinition.ID).Return(&modelSpec, nil).Once()
				return svc
			},
			InputWebhookID:          id,
			InputEvent:              *gqlEventDefinitionInput,
			ExpectedEventDefinition: nil,
			ExpectedErr:             testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			specService := testCase.SpecServiceFn()

			resolver := event.NewResolver(transact, svc, nil, nil, converter, nil, specService, nil)

			// when
			result, err := resolver.UpdateEventDefinition(context.TODO(), id, *gqlEventDefinitionInput)

			// then
			assert.Equal(t, testCase.ExpectedEventDefinition, result)
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

func TestResolver_RefetchEventSpec(t *testing.T) {
	// given
	testErr := errors.New("test error")

	eventID := "eventID"
	specID := "specID"

	dataBytes := "data"
	modelSpec := &model.Spec{
		ID:   specID,
		Data: &dataBytes,
	}

	clob := graphql.CLOB(dataBytes)
	gqlEventSpec := &graphql.EventSpec{
		Data: &clob,
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn         func() *automock.SpecService
		ConvFn            func() *automock.SpecConverter
		ExpectedEventSpec *graphql.EventSpec
		ExpectedErr       error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID).Return(modelSpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLEventSpec", modelSpec).Return(gqlEventSpec, nil).Once()
				return conv
			},
			ExpectedEventSpec: gqlEventSpec,
			ExpectedErr:       nil,
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
			ExpectedEventSpec: nil,
			ExpectedErr:       testErr,
		},
		{
			Name:            "Returns error when getting spec failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventID).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedEventSpec: nil,
			ExpectedErr:       testErr,
		},
		{
			Name:            "Returns error when spec not found",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventID).Return(nil, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedEventSpec: nil,
			ExpectedErr:       errors.Errorf("spec for Event with id %q not found", eventID),
		},
		{
			Name:            "Returns error when refetching event spec failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID).Return(nil, testErr).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
			ExpectedEventSpec: nil,
			ExpectedErr:       testErr,
		},
		{
			Name:            "Returns error converting to GraphQL fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID).Return(modelSpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLEventSpec", modelSpec).Return(nil, testErr).Once()
				return conv
			},
			ExpectedEventSpec: nil,
			ExpectedErr:       testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, eventID).Return(modelSpec, nil).Once()
				svc.On("RefetchSpec", txtest.CtxWithDBMatcher(), specID).Return(modelSpec, nil).Once()
				return svc
			},
			ConvFn: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLEventSpec", modelSpec).Return(gqlEventSpec, nil).Once()
				return conv
			},
			ExpectedEventSpec: nil,
			ExpectedErr:       testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			svc := testCase.ServiceFn()
			conv := testCase.ConvFn()
			persist, transact := testCase.TransactionerFn()
			resolver := event.NewResolver(transact, nil, nil, nil, nil, nil, svc, conv)

			// when
			result, err := resolver.RefetchEventDefinitionSpec(context.TODO(), eventID)

			// then
			assert.Equal(t, testCase.ExpectedEventSpec, result)
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
		ServiceFn       func() *automock.EventDefService
		ConverterFn     func() *automock.FetchRequestConverter
		ExpectedResult  *graphql.FetchRequest
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
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
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
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
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
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
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
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
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
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
			ServiceFn: func() *automock.EventDefService {
				svc := &automock.EventDefService{}
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

			resolver := event.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil)

			// when
			result, err := resolver.FetchRequest(context.TODO(), &graphql.EventSpec{DefinitionID: id})

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
