package mp_bundle_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	mp_bundle "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_API(t *testing.T) {
	{
		// given
		id := "bar"
		bundleID := "1"
		modelAPI := fixModelAPIDefinition(id, bundleID, "name", "bar", "test")
		gqlAPI := fixGQLAPIDefinition(id, bundleID, "name", "bar", "test")
		app := fixGQLBundle("foo", "foo", "foo")
		testErr := errors.New("Test error")
		txGen := txtest.NewTransactionContextGenerator(testErr)

		testCases := []struct {
			Name            string
			TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
			ServiceFn       func() *automock.APIService
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
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
					return conv
				},
				InputID:     "foo",
				Bundle:      app,
				ExpectedAPI: gqlAPI,
				ExpectedErr: nil,
			},
			{
				Name:            "Returns error when application retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					return conv
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
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					return conv
				},
				InputID:     "foo",
				Bundle:      app,
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
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
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

				resolver := mp_bundle.NewResolver(transact, nil, nil, svc, nil, nil, nil, nil, converter, nil, nil)

				// when
				result, err := resolver.APIDefinition(context.TODO(), testCase.Bundle, testCase.InputID)

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
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleToGraphQL", modelAPIDefinitions).Return(gqlAPIDefinitions).Once()
				return conv
			},
			ExpectedResult: fixGQLAPIDefinitionPage(gqlAPIDefinitions),
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.APIService {
				svc := &automock.APIService{}
				return svc
			},
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				return conv
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
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
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
			ConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
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

			resolver := mp_bundle.NewResolver(transact, nil, nil, svc, nil, nil, nil, nil, converter, nil, nil)
			// when
			result, err := resolver.APIDefinitions(context.TODO(), app, &group, &first, &gqlAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_EventAPI(t *testing.T) {
	// given
	id := "bar"

	modelAPI := fixMinModelEventAPIDefinition(id, "placeholder")
	gqlAPI := fixGQLEventDefinition(id, "placeholder", "placeholder", "placeholder", "placeholder")
	bundle := fixGQLBundle("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.EventService
		ConverterFn     func() *automock.EventConverter
		InputID         string
		Bundle          *graphql.Bundle
		ExpectedAPI     *graphql.EventDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
				return conv
			},
			InputID:     "foo",
			Bundle:      bundle,
			ExpectedAPI: gqlAPI,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bundle,
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns null when event for bundle not found",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError(resource.Bundle, "")).Once()

				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bundle,
			ExpectedAPI: nil,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}

				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bundle,
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("GetForBundle", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				return conv
			},
			InputID:     "foo",
			Bundle:      bundle,
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_bundle.NewResolver(transact, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil)

			// when
			result, err := resolver.EventDefinition(context.TODO(), testCase.Bundle, testCase.InputID)

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

func TestResolver_EventAPIs(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	bundleID := "1"
	group := "group"
	bundle := fixGQLBundle(bundleID, "foo", "foo")
	modelEventAPIDefinitions := []*model.EventDefinition{

		fixModelEventAPIDefinition("foo", bundleID, "Foo", "Lorem Ipsum", group),
		fixModelEventAPIDefinition("bar", bundleID, "Bar", "Lorem Ipsum", group),
	}

	gqlEventAPIDefinitions := []*graphql.EventDefinition{
		fixGQLEventDefinition("foo", bundleID, "Foo", "Lorem Ipsum", group),
		fixGQLEventDefinition("bar", bundleID, "Bar", "Lorem Ipsum", group),
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)
	contextParam := txtest.CtxWithDBMatcher()
	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.EventService
		ConverterFn     func() *automock.EventConverter
		InputFirst      *int
		InputAfter      *graphql.PageCursor
		ExpectedResult  *graphql.EventDefinitionPage
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListForBundle", contextParam, bundleID, first, after).Return(fixEventAPIDefinitionPage(modelEventAPIDefinitions), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("MultipleToGraphQL", modelEventAPIDefinitions).Return(gqlEventAPIDefinitions).Once()
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
			ExpectedResult: fixGQLEventDefinitionPage(gqlEventAPIDefinitions),
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when APIS listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("ListForBundle", contextParam, bundleID, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				return conv
			},
			InputFirst:     &first,
			InputAfter:     &gqlAfter,
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

			resolver := mp_bundle.NewResolver(transact, nil, nil, nil, svc, nil, nil, nil, nil, converter, nil)
			// when
			result, err := resolver.EventDefinitions(context.TODO(), bundle, &group, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_Document(t *testing.T) {
	// given
	id := "bar"

	modelDoc := fixModelDocument("foo", id)
	gqlDoc := fixGQLDocument(id)
	bundle := fixGQLBundle("foo", "foo", "foo")
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
			Bundle:      bundle,
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
			Bundle:      bundle,
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
			Bundle:      bundle,
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
			Bundle:      bundle,
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
			Bundle:      bundle,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_bundle.NewResolver(transact, nil, nil, nil, nil, svc, nil, nil, nil, nil, converter)

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
	bundleID := "fooid"
	contextParam := txtest.CtxWithDBMatcher()

	modelDocuments := []*model.Document{
		fixModelDocument(bundleID, "foo"),
		fixModelDocument(bundleID, "bar"),
	}
	gqlDocuments := []*graphql.Document{
		fixGQLDocument("foo"),
		fixGQLDocument("bar"),
	}
	bundle := fixGQLBundle(bundleID, "foo", "foo")

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
				svc.On("ListForBundle", contextParam, bundleID, first, after).Return(fixModelDocumentPage(modelDocuments), nil).Once()
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
				svc.On("ListForBundle", contextParam, bundleID, first, after).Return(nil, testErr).Once()
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

			resolver := mp_bundle.NewResolver(transact, nil, nil, nil, nil, svc, nil, nil, nil, nil, converter)

			// when
			result, err := resolver.Documents(context.TODO(), bundle, &first, &gqlAfter)

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

	modelBundle := fixBundleModel(t, name, desc)
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
				conv.On("InputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
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
				conv.On("InputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
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
				conv.On("InputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
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
				conv.On("InputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelBundleInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("InputFromGraphQL", gqlBundleInput).Return(modelBundleInput, nil).Once()
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

			resolver := mp_bundle.NewResolver(transact, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil)

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

func TestResolver_UpdateAPI(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "id"
	name := "foo"
	desc := "bar"
	gqlBundleUpdateInput := fixGQLBundleUpdateInput(name, desc)
	modelBundleUpdateInput := fixModelBundleUpdateInput(t, name, desc)
	gqlBundle := fixGQLBundle(id, name, desc)
	modelBundle := fixBundleModel(t, name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.BundleService
		ConverterFn     func() *automock.BundleConverter
		InputBundle     graphql.BundleInput
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
				conv.On("InputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
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
				conv.On("InputFromGraphQL", gqlBundleUpdateInput).Return(&model.BundleInput{}, testErr).Once()
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
				conv.On("InputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
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
				conv.On("InputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
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
				conv.On("InputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
				return conv
			},
			InputBundle:    gqlBundleUpdateInput,
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.BundleService {
				svc := &automock.BundleService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelBundleUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelBundle, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.BundleConverter {
				conv := &automock.BundleConverter{}
				conv.On("InputFromGraphQL", gqlBundleUpdateInput).Return(&modelBundleUpdateInput, nil).Once()
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

			resolver := mp_bundle.NewResolver(transact, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil)

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
	modelBundle := fixBundleModel(t, name, desc)
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
				return conv
			},
			ExpectedBundle: nil,
			ExpectedErr:    testErr,
		},
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

			resolver := mp_bundle.NewResolver(transact, svc, nil, nil, nil, nil, converter, nil, nil, nil, nil)

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
	bundle := fixGQLBundle("foo", "foo", "foo")
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
			Bundle:                     bundle,
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
			Bundle:                     bundle,
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
			Bundle:                     bundle,
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
			Bundle:                     bundle,
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
			Bundle:                     bundle,
			ExpectedBundleInstanceAuth: nil,
			ExpectedErr:                testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_bundle.NewResolver(transact, nil, svc, nil, nil, nil, nil, converter, nil, nil, nil)

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
		resolver := mp_bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
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

	bundle := fixGQLBundle(bundleID, "foo", "bar")
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

			resolver := mp_bundle.NewResolver(transact, nil, svc, nil, nil, nil, nil, converter, nil, nil, nil)
			// when
			result, err := resolver.InstanceAuths(context.TODO(), bundle)

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
		resolver := mp_bundle.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		//when
		_, err := resolver.InstanceAuths(context.TODO(), nil)
		//then
		require.Error(t, err)
		assert.EqualError(t, err, apperrors.NewInternalError("Bundle cannot be empty").Error())
	})
}
