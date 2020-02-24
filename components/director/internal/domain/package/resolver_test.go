package mp_package_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/internal/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_API(t *testing.T) {
	{
		// given
		id := "bar"
		appId := str.Ptr("1")
		modelAPI := fixModelAPIDefinition(id, appId, "name", "bar", "test")
		gqlAPI := fixGQLAPIDefinition(id, appId, "name", "bar", "test")
		app := fixGQLPackage("foo", "foo", "foo")
		testErr := errors.New("Test error")
		txGen := txtest.NewTransactionContextGenerator(testErr)

		testCases := []struct {
			Name            string
			TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
			ServiceFn       func() *automock.APIService
			ConverterFn     func() *automock.APIConverter
			InputID         string
			Package         *graphql.Package
			ExpectedAPI     *graphql.APIDefinition
			ExpectedErr     error
		}{
			{
				Name:            "Success",
				TransactionerFn: txGen.ThatSucceeds,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
					return conv
				},
				InputID:     "foo",
				Package:     app,
				ExpectedAPI: gqlAPI,
				ExpectedErr: nil,
			},
			{
				Name:            "Returns error when application retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					return conv
				},
				InputID:     "foo",
				Package:     app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns null when application retrieval failed",
				TransactionerFn: txGen.ThatDoesntExpectCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError("")).Once()

					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					return conv
				},
				InputID:     "foo",
				Package:     app,
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
				Package:     app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
			{
				Name:            "Returns error when commit failed",
				TransactionerFn: txGen.ThatFailsOnCommit,
				ServiceFn: func() *automock.APIService {
					svc := &automock.APIService{}
					svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()
					return svc
				},
				ConverterFn: func() *automock.APIConverter {
					conv := &automock.APIConverter{}
					return conv
				},
				InputID:     "foo",
				Package:     app,
				ExpectedAPI: nil,
				ExpectedErr: testErr,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				persist, transact := testCase.TransactionerFn()
				svc := testCase.ServiceFn()
				converter := testCase.ConverterFn()

				resolver := mp_package.NewResolver(transact, nil, svc, nil, nil, nil, converter, nil, nil)

				// when
				result, err := resolver.APIDefinition(context.TODO(), testCase.Package, testCase.InputID)

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

	packageID := "1"
	group := "group"
	app := fixGQLPackage(packageID, "foo", "foo")
	modelAPIDefinitions := []*model.APIDefinition{

		fixModelAPIDefinition("foo", &packageID, "Foo", "Lorem Ipsum", group),
		fixModelAPIDefinition("bar", &packageID, "Bar", "Lorem Ipsum", group),
	}

	gqlAPIDefinitions := []*graphql.APIDefinition{
		fixGQLAPIDefinition("foo", &packageID, "Foo", "Lorem Ipsum", group),
		fixGQLAPIDefinition("bar", &packageID, "Bar", "Lorem Ipsum", group),
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
				svc.On("ListForPackage", txtest.CtxWithDBMatcher(), packageID, first, after).Return(fixAPIDefinitionPage(modelAPIDefinitions), nil).Once()
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
				svc.On("ListForPackage", txtest.CtxWithDBMatcher(), packageID, first, after).Return(nil, testErr).Once()
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
				svc.On("ListForPackage", txtest.CtxWithDBMatcher(), packageID, first, after).Return(fixAPIDefinitionPage(modelAPIDefinitions), nil).Once()
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

			resolver := mp_package.NewResolver(transact, nil, svc, nil, nil, nil, converter, nil, nil)
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
	gqlAPI := fixGQLEventDefinition(id, str.Ptr("placeholder"), str.Ptr("placeholder"), "placeholder", "placeholder", "placeholder")
	pkg := fixGQLPackage("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.EventService
		ConverterFn     func() *automock.EventConverter
		InputID         string
		Package         *graphql.Package
		ExpectedAPI     *graphql.EventDefinition
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("ToGraphQL", modelAPI).Return(gqlAPI).Once()
				return conv
			},
			InputID:     "foo",
			Package:     pkg,
			ExpectedAPI: gqlAPI,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				return conv
			},
			InputID:     "foo",
			Package:     pkg,
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns null when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError("")).Once()

				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				return conv
			},
			InputID:     "foo",
			Package:     pkg,
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
			Package:     pkg,
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.EventService {
				svc := &automock.EventService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelAPI, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				return conv
			},
			InputID:     "foo",
			Package:     pkg,
			ExpectedAPI: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_package.NewResolver(transact, nil, nil, svc, nil, nil, nil, converter, nil)

			// when
			result, err := resolver.EventDefinition(context.TODO(), testCase.Package, testCase.InputID)

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

	packageID := "1"
	applicationID := "1"
	group := "group"
	pkg := fixGQLPackage(packageID, "foo", "foo")
	modelEventAPIDefinitions := []*model.EventDefinition{

		fixModelEventAPIDefinition("foo", &packageID, &applicationID, "Foo", "Lorem Ipsum", group),
		fixModelEventAPIDefinition("bar", &packageID, &applicationID, "Bar", "Lorem Ipsum", group),
	}

	gqlEventAPIDefinitions := []*graphql.EventDefinition{
		fixGQLEventDefinition("foo", &packageID, &applicationID, "Foo", "Lorem Ipsum", group),
		fixGQLEventDefinition("bar", &packageID, &applicationID, "Bar", "Lorem Ipsum", group),
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
				svc.On("ListForPackage", contextParam, packageID, first, after).Return(fixEventAPIDefinitionPage(modelEventAPIDefinitions), nil).Once()
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
				svc.On("ListForPackage", contextParam, packageID, first, after).Return(nil, testErr).Once()
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

			resolver := mp_package.NewResolver(transact, nil, nil, svc, nil, nil, nil, converter, nil)
			// when
			result, err := resolver.EventDefinitions(context.TODO(), pkg, &group, testCase.InputFirst, testCase.InputAfter)

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

	modelDoc := fixModelDocument("foo", "bar", id)
	gqlDoc := fixGQLDocument(id)
	pkg := fixGQLPackage("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.DocumentService
		ConverterFn     func() *automock.DocumentConverter
		InputID         string
		Package         *graphql.Package
		ExpectedDoc     *graphql.Document
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelDoc, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("ToGraphQL", modelDoc).Return(gqlDoc).Once()
				return conv
			},
			InputID:     "foo",
			Package:     pkg,
			ExpectedDoc: gqlDoc,
			ExpectedErr: nil,
		},
		{
			Name:            "Returns error when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Package:     pkg,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns null when application retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError("")).Once()

				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Package:     pkg,
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
			Package:     pkg,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.DocumentService {
				svc := &automock.DocumentService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelDoc, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				return conv
			},
			InputID:     "foo",
			Package:     pkg,
			ExpectedDoc: nil,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_package.NewResolver(transact, nil, nil, nil, svc, nil, nil, nil, converter)

			// when
			result, err := resolver.Document(context.TODO(), testCase.Package, testCase.InputID)

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
	pkgID := "fooid"
	appID := "barid"
	contextParam := txtest.CtxWithDBMatcher()

	modelDocuments := []*model.Document{
		fixModelDocument(pkgID, appID, "foo"),
		fixModelDocument(pkgID, appID, "bar"),
	}
	gqlDocuments := []*graphql.Document{
		fixGQLDocument("foo"),
		fixGQLDocument("bar"),
	}
	pkg := fixGQLPackage(pkgID, "foo", "foo")

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
				svc.On("ListForPackage", contextParam, pkgID, first, after).Return(fixModelDocumentPage(modelDocuments), nil).Once()
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
				svc.On("ListForPackage", contextParam, pkgID, first, after).Return(nil, testErr).Once()
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

			resolver := mp_package.NewResolver(transact, nil, nil, nil, svc, nil, nil, nil, converter)

			// when
			result, err := resolver.Documents(context.TODO(), pkg, &first, &gqlAfter)

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

func TestResolver_AddPackage(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	appId := "1"
	desc := "bar"
	name := "baz"

	modelPackage := fixPackageModel(t, name, desc)
	gqlPackage := fixGQLPackage(id, name, desc)
	gqlPackageInput := fixGQLPackageCreateInput(name, desc)
	modelPackageInput := fixModelPackageCreateInput(t, name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.PackageService
		ConverterFn     func() *automock.PackageConverter
		ExpectedPackage *graphql.Package
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(&modelPackageInput, nil).Once()
				conv.On("ToGraphQL", modelPackage).Return(gqlPackage, nil).Once()
				return conv
			},
			ExpectedPackage: gqlPackage,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when starting transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when converting input from GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(&model.PackageCreateInput{}, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when adding Package failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(&modelPackageInput, nil).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(&modelPackageInput, nil).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(&modelPackageInput, nil).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), appId, modelPackageInput).Return(id, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("CreateInputFromGraphQL", gqlPackageInput).Return(&modelPackageInput, nil).Once()
				conv.On("ToGraphQL", modelPackage).Return(nil, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_package.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.AddPackage(context.TODO(), appId, gqlPackageInput)

			// then
			assert.Equal(t, testCase.ExpectedPackage, result)
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
	gqlPackageUpdateInput := fixGQLPackageUpdateInput(name, desc)
	modelPackageUpdateInput := fixModelPackageUpdateInput(t, name, desc)
	gqlPackage := fixGQLPackage(id, name, desc)
	modelPackage := fixPackageModel(t, name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.PackageService
		ConverterFn     func() *automock.PackageConverter
		InputPackage    graphql.PackageUpdateInput
		ExpectedPackage *graphql.Package
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(&modelPackageUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelPackage).Return(gqlPackage, nil).Once()
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: gqlPackage,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when starting transaction failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when converting from GraphQL failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(&model.PackageUpdateInput{}, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package update failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(&modelPackageUpdateInput, nil).Once()
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(&modelPackageUpdateInput, nil).Once()
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when commit transaction failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(&modelPackageUpdateInput, nil).Once()
				return conv
			},
			InputPackage:    gqlPackageUpdateInput,
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when converting to GraphQL failed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), id, modelPackageUpdateInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("UpdateInputFromGraphQL", gqlPackageUpdateInput).Return(&modelPackageUpdateInput, nil).Once()
				conv.On("ToGraphQL", modelPackage).Return(nil, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_package.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.UpdatePackage(context.TODO(), id, gqlPackageUpdateInput)

			// then
			assert.Equal(t, testCase.ExpectedPackage, result)
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

func TestResolver_DeletePackage(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "id"
	name := "foo"
	desc := "desc"
	modelPackage := fixPackageModel(t, name, desc)
	gqlPackage := fixGQLPackage(id, name, desc)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.PackageService
		ConverterFn     func() *automock.PackageConverter
		ExpectedPackage *graphql.Package
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("ToGraphQL", modelPackage).Return(gqlPackage, nil).Once()
				return conv
			},
			ExpectedPackage: gqlPackage,
			ExpectedErr:     nil,
		},
		{
			Name:            "Return error when starting transaction fails",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when Package deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Return error when commit transaction fails",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageService {
				svc := &automock.PackageService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), id).Return(modelPackage, nil).Once()
				svc.On("Delete", txtest.CtxWithDBMatcher(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageConverter {
				conv := &automock.PackageConverter{}
				conv.On("ToGraphQL", modelPackage).Return(nil, testErr).Once()
				return conv
			},
			ExpectedPackage: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// given
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_package.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.DeletePackage(context.TODO(), id)

			// then
			assert.Equal(t, testCase.ExpectedPackage, result)
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
	modelPackageInstanceAuth := fixModelPackageInstanceAuth(id)
	gqlPackageInstanceAuth := fixGQLPackageInstanceAuth(id)
	pkg := fixGQLPackage("foo", "foo", "foo")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                        string
		TransactionerFn             func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn                   func() *automock.PackageInstanceAuthService
		ConverterFn                 func() *automock.PackageInstanceAuthConverter
		InputID                     string
		Package                     *graphql.Package
		ExpectedPackageInstanceAuth *graphql.PackageInstanceAuth
		ExpectedErr                 error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageInstanceAuthService {
				svc := &automock.PackageInstanceAuthService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelPackageInstanceAuth, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.PackageInstanceAuthConverter {
				conv := &automock.PackageInstanceAuthConverter{}
				conv.On("ToGraphQL", modelPackageInstanceAuth).Return(gqlPackageInstanceAuth, nil).Once()
				return conv
			},
			InputID:                     "foo",
			Package:                     pkg,
			ExpectedPackageInstanceAuth: gqlPackageInstanceAuth,
			ExpectedErr:                 nil,
		},
		{
			Name:            "Returns error when Package retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageInstanceAuthService {
				svc := &automock.PackageInstanceAuthService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.PackageInstanceAuthConverter {
				conv := &automock.PackageInstanceAuthConverter{}
				return conv
			},
			InputID:                     "foo",
			Package:                     pkg,
			ExpectedPackageInstanceAuth: nil,
			ExpectedErr:                 testErr,
		},
		{
			Name:            "Returns nil when Package retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageInstanceAuthService {
				svc := &automock.PackageInstanceAuthService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(nil, apperrors.NewNotFoundError("")).Once()

				return svc
			},
			ConverterFn: func() *automock.PackageInstanceAuthConverter {
				conv := &automock.PackageInstanceAuthConverter{}
				return conv
			},
			InputID:                     "foo",
			Package:                     pkg,
			ExpectedPackageInstanceAuth: nil,
			ExpectedErr:                 nil,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.PackageInstanceAuthService {
				svc := &automock.PackageInstanceAuthService{}

				return svc
			},
			ConverterFn: func() *automock.PackageInstanceAuthConverter {
				conv := &automock.PackageInstanceAuthConverter{}
				return conv
			},
			InputID:                     "foo",
			Package:                     pkg,
			ExpectedPackageInstanceAuth: nil,
			ExpectedErr:                 testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.PackageInstanceAuthService {
				svc := &automock.PackageInstanceAuthService{}
				svc.On("GetForPackage", txtest.CtxWithDBMatcher(), "foo", "foo").Return(modelPackageInstanceAuth, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageInstanceAuthConverter {
				conv := &automock.PackageInstanceAuthConverter{}
				return conv
			},
			InputID:                     "foo",
			Package:                     pkg,
			ExpectedPackageInstanceAuth: nil,
			ExpectedErr:                 testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := mp_package.NewResolver(transact, nil, nil, svc, converter)

			// when
			result, err := resolver.InstanceAuth(context.TODO(), testCase.Package, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedPackageInstanceAuth, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}

	t.Run("Returns error when Package is nil", func(t *testing.T) {
		resolver := mp_package.NewResolver(nil, nil, nil, nil, nil)
		//when
		_, err := resolver.InstanceAuth(context.TODO(), nil, "")
		//then
		require.Error(t, err)
		assert.EqualError(t, err, "Package cannot be empty")
	})
}

func TestResolver_InstanceAuths(t *testing.T) {
	// given
	testErr := errors.New("test error")

	pkg := fixGQLPackage(packageID, "foo", "bar")
	modelPackageInstanceAuths := []*model.PackageInstanceAuth{
		fixModelPackageInstanceAuth("foo"),
		fixModelPackageInstanceAuth("bar"),
	}

	gqlPackageInstanceAuths := []*graphql.PackageInstanceAuth{
		fixGQLPackageInstanceAuth("foo"),
		fixGQLPackageInstanceAuth("bar"),
	}

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.PackageInstanceAuthService
		ConverterFn     func() *automock.PackageInstanceAuthConverter
		ExpectedResult  []*graphql.PackageInstanceAuth
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.PackageInstanceAuthService {
				svc := &automock.PackageInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), packageID).Return(modelPackageInstanceAuths, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageInstanceAuthConverter {
				conv := &automock.PackageInstanceAuthConverter{}
				conv.On("MultipleToGraphQL", modelPackageInstanceAuths).Return(gqlPackageInstanceAuths, nil).Once()
				return conv
			},
			ExpectedResult: gqlPackageInstanceAuths,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.PackageInstanceAuthService {
				svc := &automock.PackageInstanceAuthService{}
				return svc
			},
			ConverterFn: func() *automock.PackageInstanceAuthConverter {
				conv := &automock.PackageInstanceAuthConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Package Instance Auths listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.PackageInstanceAuthService {
				svc := &automock.PackageInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), packageID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageInstanceAuthConverter {
				conv := &automock.PackageInstanceAuthConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.PackageInstanceAuthService {
				svc := &automock.PackageInstanceAuthService{}
				svc.On("List", txtest.CtxWithDBMatcher(), packageID).Return(modelPackageInstanceAuths, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.PackageInstanceAuthConverter {
				conv := &automock.PackageInstanceAuthConverter{}
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

			resolver := mp_package.NewResolver(transact, nil, nil, svc, converter)
			// when
			result, err := resolver.InstanceAuths(context.TODO(), pkg)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			persist.AssertExpectations(t)
			transact.AssertExpectations(t)
			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}

	t.Run("Returns error when Package is nil", func(t *testing.T) {
		resolver := mp_package.NewResolver(nil, nil, nil, nil, nil)
		//when
		_, err := resolver.InstanceAuths(context.TODO(), nil)
		//then
		require.Error(t, err)
		assert.EqualError(t, err, "Package cannot be empty")
	})
}
