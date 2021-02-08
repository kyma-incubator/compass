package mp_bundle_test

import (
	"context"
	"testing"

	mp_bundle "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_API(t *testing.T) {
	{
		// given
		id := "bar"
		bndlID := "1"
		modelAPI := fixModelAPIDefinition(id, bndlID, "name", "bar", "test")
		modelSpec := &model.Spec{
			ID:         id,
			ObjectType: model.APISpecReference,
			ObjectID:   id,
		}
		gqlAPI := fixGQLAPIDefinition(id, bndlID, "name", "bar", "test")
		app := fixGQLBundle("foo", "foo", "foo")
		testErr := errors.New("Test error")
		txGen := txtest.NewTransactionContextGenerator(testErr)

		testCases := []struct {
			Name            string
			TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
			ServiceFn       func() *automock.APIService
			SpecServiceFn   func() *automock.SpecService
			ConverterFn     func() *automock.APIConverter
			InputID         string
			Bundle          *graphql.Bundle
			ExpectedAPI     *graphql.APIDefinition
			ExpectedErr     error
		}{
			{
				Name:            "Success",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPI.ID).Return(modelSpec, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					conv.On("ToGraphQL", modelAPI, modelSpec).Return(gqlAPI, nil).Once()
					return conv
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: gqlAPI,
				ExpectedErr: nil,
			},
			{
				Name:            "Returns error when bundle retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				ConverterFn: func() *automock.APIConverter {
					return &automock.APIConverter{}
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns null when api for bundle not found",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Bundle, "")).Once()
					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				ConverterFn: func() *automock.APIConverter {
					return &automock.APIConverter{}
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: nil,
			},
			{
				Name:            "Returns error when Spec retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPI.ID).Return(nil, testErr).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					return &automock.APIConverter{}
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns error when converting to GraphQL failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPI.ID).Return(modelSpec, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					conv.On("ToGraphQL", modelAPI, modelSpec).Return(nil, testErr).Once()
					return conv
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns error when commit begin error",
				TransactionerFn: txGen.ThatFailsOnBegin,
				ServiceFn: func() *automock.APIService {
					return &automock.APIService{}
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				ConverterFn: func() *automock.APIConverter {
					return &automock.APIConverter{}
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns error when commit failed",
				TransactionerFn: txGen.ThatFailsOnCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPI.ID).Return(modelSpec, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					conv.On("ToGraphQL", modelAPI, modelSpec).Return(gqlAPI, nil).Once()
					return conv
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				persist, transact := testCase.TransactionerFn()
				svc := testCase.ServiceFn()
				converter := testCase.ConverterFn()
				specSvc := testCase.SpecServiceFn()

				resolver := mp_bundle.NewResolver(transact, nil, nil, svc, nil, nil, nil, nil, converter, nil, nil, specSvc)

				// when
				result, err := resolver.APIDefinition(context.TODO(), testCase.Bundle, testCase.InputID)

				// then
				assert.Equal(t, testCase.ExpectedAPI, result)
				if testCase.ExpectedErr != nil {
					require.Error(t, err)
					assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				} else {
					require.Nil(t, err)
				}

				svc.AssertExpectations(t)
				persist.AssertExpectations(t)
				transact.AssertExpectations(t)
				converter.AssertExpectations(t)
				specSvc.AssertExpectations(t)
			})
		}
	}
}

func TestResolver_Apis(t *testing.T) {
	// given
	testErr := errors.New("test error")

	bundleID := "1"
	group := "group"
	app := fixGQLBundle(bundleID, "foo", "foo")
	modelAPIDefinitions := []*model.APIDefinition{
		fixModelAPIDefinition("foo", bundleID, "Foo", "Lorem Ipsum", group),
		fixModelAPIDefinition("bar", bundleID, "Bar", "Lorem Ipsum", group),
	}

	modelSpecs := []*model.Spec{
		{
			ID:         "test-spec-1",
			ObjectType: model.APISpecReference,
			ObjectID:   "foo",
		},
		{
			ID:         "test-spec-2",
			ObjectType: model.APISpecReference,
			ObjectID:   "bar",
		},
	}

	gqlAPIDefinitions := []*graphql.APIDefinition{
		fixGQLAPIDefinition("foo", bundleID, "Foo", "Lorem Ipsum", group),
		fixGQLAPIDefinition("bar", bundleID, "Bar", "Lorem Ipsum", group),
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.APIService
		ConverterFn     func() *automock.APIConverter
		SpecServiceFn   func() *automock.SpecService
		ExpectedResult  *graphql.APIDefinitionPage
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(fixAPIDefinitionPage(modelAPIDefinitions), nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinitions[0].ID).Return(modelSpecs[0], nil).Once()
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinitions[1].ID).Return(modelSpecs[1], nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleToGraphQL", modelAPIDefinitions, modelSpecs).Return(gqlAPIDefinitions, nil).Once()
				return conv
			},
			ExpectedResult: fixGQLAPIDefinitionPage(gqlAPIDefinitions),
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				return &automock.APIService{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when APIS listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(nil, testErr).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(fixAPIDefinitionPage(modelAPIDefinitions), nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinitions[0].ID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				return &automock.APIConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(fixAPIDefinitionPage(modelAPIDefinitions), nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinitions[0].ID).Return(modelSpecs[0], nil).Once()
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinitions[1].ID).Return(modelSpecs[1], nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleToGraphQL", modelAPIDefinitions, modelSpecs).Return(nil, testErr).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(fixAPIDefinitionPage(modelAPIDefinitions), nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinitions[0].ID).Return(modelSpecs[0], nil).Once()
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.APISpecReference, modelAPIDefinitions[1].ID).Return(modelSpecs[1], nil).Once()
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleToGraphQL", modelAPIDefinitions, modelSpecs).Return(gqlAPIDefinitions, nil).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			specService := testCase.SpecServiceFn()

			resolver := mp_bundle.NewResolver(transact, nil, nil, svc, nil, nil, nil, nil, converter, nil, nil, specService)
			// when
			result, err := resolver.APIDefinitions(context.TODO(), app, &group, &first, &gqlAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			specService.AssertExpectations(t)
		})
	}
}

func TestResolver_Event(t *testing.T) {
	{
		// given
		id := "bar"
		bndlID := "1"
		modelEvent := fixModelEventAPIDefinition(id, bndlID, "name", "bar", "test")
		modelSpec := &model.Spec{
			ID:         id,
			ObjectType: model.EventSpecReference,
			ObjectID:   id,
		}
		gqlEvent := fixGQLEventDefinition(id, bndlID, "name", "bar", "test")
		app := fixGQLBundle("foo", "foo", "foo")
		testErr := errors.New("Test error")
		txGen := txtest.NewTransactionContextGenerator(testErr)

		testCases := []struct {
			Name            string
			TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
			ServiceFn       func() *automock.EventService
			SpecServiceFn   func() *automock.SpecService
			ConverterFn     func() *automock.EventConverter
			InputID         string
			Bundle          *graphql.Bundle
			ExpectedEvent   *graphql.EventDefinition
			ExpectedErr     error
		}{
			{
				Name:            "Success",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelEvent, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEvent.ID).Return(modelSpec, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.EventConverter {
					conv := &automock.EventConverter{}
					conv.On("ToGraphQL", modelEvent, modelSpec).Return(gqlEvent, nil).Once()
					return conv
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: gqlEvent,
				ExpectedErr:   nil,
			},
			{
				Name:            "Returns error when bundle retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				ConverterFn: func() *automock.EventConverter {
					return &automock.EventConverter{}
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
			{
				Name:            "Returns null when api for bundle not found",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Bundle, "")).Once()
					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				ConverterFn: func() *automock.EventConverter {
					return &automock.EventConverter{}
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   nil,
			},
			{
				Name:            "Returns error when Spec retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelEvent, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEvent.ID).Return(nil, testErr).Once()
					return svc
				},
				ConverterFn: func() *automock.EventConverter {
					return &automock.EventConverter{}
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
			{
				Name:            "Returns error when converting to GraphQL failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelEvent, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEvent.ID).Return(modelSpec, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.EventConverter {
					conv := &automock.EventConverter{}
					conv.On("ToGraphQL", modelEvent, modelSpec).Return(nil, testErr).Once()
					return conv
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
			{
				Name:            "Returns error when commit begin error",
				TransactionerFn: txGen.ThatFailsOnBegin,
				ServiceFn: func() *automock.EventService {
					return &automock.EventService{}
				},
				SpecServiceFn: func() *automock.SpecService {
					return &automock.SpecService{}
				},
				ConverterFn: func() *automock.EventConverter {
					return &automock.EventConverter{}
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
			{
				Name:            "Returns error when commit failed",
				TransactionerFn: txGen.ThatFailsOnCommit,
				ServiceFn: func() *automock.EventService {
					svc := &automock.EventService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelEvent, nil).Once()

					return svc
				},
				SpecServiceFn: func() *automock.SpecService {
					svc := &automock.SpecService{}
					svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEvent.ID).Return(modelSpec, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.EventConverter {
					conv := &automock.EventConverter{}
					conv.On("ToGraphQL", modelEvent, modelSpec).Return(gqlEvent, nil).Once()
					return conv
				},
				InputID:       "foo",
				Bundle:        app,
				ExpectedEvent: nil,
				ExpectedErr:   testErr,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				persist, transact := testCase.TransactionerFn()
				svc := testCase.ServiceFn()
				converter := testCase.ConverterFn()
				specSvc := testCase.SpecServiceFn()

				resolver := mp_bundle.NewResolver(transact, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil, specSvc)

				// when
				result, err := resolver.EventDefinition(context.TODO(), testCase.Bundle, testCase.InputID)

				// then
				assert.Equal(t, testCase.ExpectedEvent, result)
				if testCase.ExpectedErr != nil {
					require.Error(t, err)
					assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				} else {
					require.Nil(t, err)
				}

				svc.AssertExpectations(t)
				persist.AssertExpectations(t)
				transact.AssertExpectations(t)
				converter.AssertExpectations(t)
				specSvc.AssertExpectations(t)
			})
		}
	}
}

func TestResolver_Events(t *testing.T) {
	// given
	testErr := errors.New("test error")

	bundleID := "1"
	group := "group"
	app := fixGQLBundle(bundleID, "foo", "foo")
	modelEventDefinitions := []*model.EventDefinition{
		fixModelEventAPIDefinition("foo", bundleID, "Foo", "Lorem Ipsum", group),
		fixModelEventAPIDefinition("bar", bundleID, "Bar", "Lorem Ipsum", group),
	}

	modelSpecs := []*model.Spec{
		{
			ID:         "test-spec-1",
			ObjectType: model.EventSpecReference,
			ObjectID:   "foo",
		},
		{
			ID:         "test-spec-2",
			ObjectType: model.EventSpecReference,
			ObjectID:   "bar",
		},
	}

	gqlEventDefinitions := []*graphql.EventDefinition{
		fixGQLEventDefinition("foo", bundleID, "Foo", "Lorem Ipsum", group),
		fixGQLEventDefinition("bar", bundleID, "Bar", "Lorem Ipsum", group),
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.EventService
		ConverterFn     func() *automock.EventConverter
		SpecServiceFn   func() *automock.SpecService
		ExpectedResult  *graphql.EventDefinitionPage
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(fixEventAPIDefinitionPage(modelEventDefinitions), nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinitions[0].ID).Return(modelSpecs[0], nil).Once()
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinitions[1].ID).Return(modelSpecs[1], nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("MultipleToGraphQL", modelEventDefinitions, modelSpecs).Return(gqlEventDefinitions, nil).Once()
				return conv
			},
			ExpectedResult: fixGQLEventDefinitionPage(gqlEventDefinitions),
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventService {
				return &automock.EventService{}
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ConverterFn: func() *automock.EventConverter {
				return &automock.EventConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when EventS listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(nil, testErr).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				return &automock.SpecService{}
			},
			ConverterFn: func() *automock.EventConverter {
				return &automock.EventConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Spec retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(fixEventAPIDefinitionPage(modelEventDefinitions), nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinitions[0].ID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				return &automock.EventConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(fixEventAPIDefinitionPage(modelEventDefinitions), nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinitions[0].ID).Return(modelSpecs[0], nil).Once()
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinitions[1].ID).Return(modelSpecs[1], nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("MultipleToGraphQL", modelEventDefinitions, modelSpecs).Return(nil, testErr).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListForBundle", txtest.CtxWithDBMatcher(), bundleID, first, after).Return(fixEventAPIDefinitionPage(modelEventDefinitions), nil).Once()
				return svc
			},
			SpecServiceFn: func() *automock.SpecService {
				svc := &automock.SpecService{}
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinitions[0].ID).Return(modelSpecs[0], nil).Once()
				svc.On("GetByReferenceObjectID", txtest.CtxWithDBMatcher(), model.EventSpecReference, modelEventDefinitions[1].ID).Return(modelSpecs[1], nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("MultipleToGraphQL", modelEventDefinitions, modelSpecs).Return(gqlEventDefinitions, nil).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			specService := testCase.SpecServiceFn()

			resolver := mp_bundle.NewResolver(transact, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil, specService)
			// when
			result, err := resolver.EventDefinitions(context.TODO(), app, &group, &first, &gqlAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			specService.AssertExpectations(t)
		})
	}
}

func TestResolver_Document(t *testing.T) {
	// given
	id := "bar"

	modelDoc := fixModelDocument("foo", id)
	gqlDoc := fixGQLDocument(id)
	bndl := fixGQLBundle("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.DocumentService
		ConverterFn     func() *automock.DocumentConverter
		InputID         string
		Bundle          *graphql.Bundle
		ExpectedDoc     *graphql.Document
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelDoc, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("ToGraphQL", modelDoc).Return(gqlDoc).Once()
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: gqlDoc,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns null when document for bundle not found",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Bundle, "")).Once()

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: nil,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelDoc, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bndl,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_bundle.NewResolver(transact, nil, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil)

			// when
			result, err := resolver.Document(context.TODO(), testCase.Bundle, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedDoc, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Documents(t *testing.T) {
	// given
	bndlID := "fooid"
	contextParam := txtest.CtxWithDBMatcher()

	modelDocuments := []*model.Document{
		fixModelDocument(bndlID, "foo"),
		fixModelDocument(bndlID, "bar"),
	}
	gqlDocuments := []*graphql.Document{
		fixGQLDocument("foo"),
		fixGQLDocument("bar"),
	}
	bndl := fixGQLBundle(bndlID, "foo", "foo")

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.DocumentService
		ConverterFn     func() *automock.DocumentConverter
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ExpectedResult  *graphql.DocumentPage
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			PersistenceFn:   txtest.PersistenceContextThatExpectsCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("ListForBundle", contextParam, bndlID, first, after).Return(fixModelDocumentPage(modelDocuments), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("MultipleToGraphQL", modelDocuments).Return(gqlDocuments).Once()
				return conv
			},
			ExpectedResult: fixGQLDocumentPage(gqlDocuments),
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when document listing failed",
			PersistenceFn:   txtest.PersistenceContextThatDoesntExpectCommit,
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("ListForBundle", contextParam, bndlID, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)

			resolver := mp_bundle.NewResolver(transact, nil, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil)

			// when
			result, err := resolver.Documents(context.TODO(), bndl, &first, &gqlAfter)

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

func TestResolver_AddBundle(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	appId := "1"
	desc := "bar"
	name := "baz"

	modelBundle := fixBundleModel(name, desc)
	gqlBundle := fixGQLBundle(id, name, desc)
	gqlBundleInput := fixGQLBundleCreateInput(name, desc)
	modelBundleInput := fixModelBundleCreateInput(name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleService
		ConverterFn     func() *automock.BundleConverter
		ExpectedBundle  *graphql.Bundle
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			ExpectedBundle: gqlBundle,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when adding Bundle failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelBundleInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("CreateInputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(nil, testErr).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_bundle.NewResolver(transact, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil, nil)

			// when
			result, err := resolver.AddBundle(context.TODO(), appId, gqlBundleInput)

			// then
			assert.Equal(t, testCase.ExpectedBundle, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateBundle(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "id"
	name := "foo"
	desc := "bar"
	gqlBundleUpdateInput := fixGQLBundleUpdateInput(name, desc)
	modelBundleUpdateInput := fixModelBundleUpdateInput(name, desc)
	gqlBundle := fixGQLBundle(id, name, desc)
	modelBundle := fixBundleModel(name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleService
		ConverterFn     func() *automock.BundleConverter
		InputBundle     graphql.BundleUpdateInput
		ExpectedBundle  *graphql.Bundle
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: gqlBundle,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting from GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&model.BundleUpdateInput{}, testErr).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("UpdateInputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelBundle).Return(nil, testErr).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_bundle.NewResolver(transact, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil, nil)

			// when
			result, err := resolver.UpdateBundle(context.TODO(), id, gqlBundleUpdateInput)

			// then
			assert.Equal(t, testCase.ExpectedBundle, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteBundle(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "id"
	name := "foo"
	desc := "desc"
	modelBundle := fixBundleModel(name, desc)
	gqlBundle := fixGQLBundle(id, name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleService
		ConverterFn     func() *automock.BundleConverter
		ExpectedBundle  *graphql.Bundle
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			ExpectedBundle: gqlBundle,
			ExpectedErr:    nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("ToGraphQL", modelBundle).Return(gqlBundle, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("ToGraphQL", modelBundle).Return(nil, testErr).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_bundle.NewResolver(transact, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil, nil)

			// when
			result, err := resolver.DeleteBundle(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedBundle, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
			transact.AssertExpectations(t)
			persist.AssertExpectations(t)
		})
	}
}

func TestResolver_InstanceAuth(t *testing.T) {
	// given
	id := "foo"
	modelBundleInstanceAuth := fixModelBundleInstanceAuth(id)
	gqlBundleInstanceAuth := fixGQLBundleInstanceAuth(id)
	bndl := fixGQLBundle("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		TransactionerFn            func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn                  func() *automock.BundleInstanceAuthService
		ConverterFn                func() *automock.BundleInstanceAuthConverter
		InputID                    string
		Bundle                     *graphql.Bundle
		ExpectedBundleInstanceAuth *graphql.BundleInstanceAuth
		ExpectedErr                error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelBundleInstanceAuth, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				conv.On("ToGraphQL", modelBundleInstanceAuth).Return(gqlBundleInstanceAuth, nil).Once()
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: gqlBundleInstanceAuth,
			ExpectedErr:                nil,
		},
		{
			Name:            "Returns error when Bundle retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                testErr,
		},
		{
			Name:            "Returns nil when bundle instance auth for bundle not found",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Bundle, "")).Once()

				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                nil,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}

				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelBundleInstanceAuth, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			InputID:                    "foo",
			Bundle:                     bndl,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_bundle.NewResolver(transact, nil, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil)

			// when
			result, err := resolver.InstanceAuth(context.TODO(), testCase.Bundle, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedBundleInstanceAuth, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}

	t.Run("Returns error when Bundle is nil", func(t *testing.T) {
		resolver := mp_bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		//when
		_, err := resolver.InstanceAuth(context.TODO(), nil, "")
		//then
		require.Error(t, err)
		assert.EqualError(t, err, apperrors.NewInternalError("Bundle cannot be empty").Error())
	})
}

func TestResolver_InstanceAuths(t *testing.T) {
	// given
	testErr := errors.New("test error")

	bndl := fixGQLBundle(bundleID, "foo", "bar")
	modelBundleInstanceAuths := []*model.BundleInstanceAuth{
		fixModelBundleInstanceAuth("foo"),
		fixModelBundleInstanceAuth("bar"),
	}

	gqlBundleInstanceAuths := []*graphql.BundleInstanceAuth{
		fixGQLBundleInstanceAuth("foo"),
		fixGQLBundleInstanceAuth("bar"),
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleInstanceAuthService
		ConverterFn     func() *automock.BundleInstanceAuthConverter
		ExpectedResult  []*graphql.BundleInstanceAuth
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), bundleID).Return(modelBundleInstanceAuths, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				conv.On("MultipleToGraphQL", modelBundleInstanceAuths).Return(gqlBundleInstanceAuths, nil).Once()
				return conv
			},
			ExpectedResult: gqlBundleInstanceAuths,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Bundle Instance Auths listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), bundleID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), bundleID).Return(modelBundleInstanceAuths, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleInstanceAuthConverter {
				conv := &automock.BundleInstanceAuthConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_bundle.NewResolver(transact, nil, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil)
			// when
			result, err := resolver.InstanceAuths(context.TODO(), bndl)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}

	t.Run("Returns error when Bundle is nil", func(t *testing.T) {
		resolver := mp_bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		//when
		_, err := resolver.InstanceAuths(context.TODO(), nil)
		//then
		require.Error(t, err)
		assert.EqualError(t, err, apperrors.NewInternalError("Bundle cannot be empty").Error())
	})
}
