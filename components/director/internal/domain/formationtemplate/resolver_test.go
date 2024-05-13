package formationtemplate_test

import (
	"reflect"
	"testing"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolver_FormationTemplate(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		TxFn                       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationTemplateConverter func() *automock.FormationTemplateConverter
		FormationTemplateService   func() *automock.FormationTemplateService
		ExpectedOutput             *graphql.FormationTemplate
		ExpectedError              error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&formationTemplateModel, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("ToGraphQL", &formationTemplateModel).Return(&graphQLFormationTemplate, nil)

				return converter
			},
			ExpectedOutput: &graphQLFormationTemplate,
			ExpectedError:  nil,
		},
		{
			Name: "Error when getting from service fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
		{
			Name: "Returns error when formation template not found",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(nil, formationTemplateNotFoundErr)

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              formationTemplateNotFoundErr,
		},
		{
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&formationTemplateModel, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("ToGraphQL", &formationTemplateModel).Return(&graphQLFormationTemplate, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name: "Error when converting ot graphql fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&formationTemplateModel, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("ToGraphQL", &formationTemplateModel).Return(nil, testErr)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                       "Returns error when failing on the beginning of a transaction",
			TxFn:                       txGen.ThatFailsOnBegin,
			FormationTemplateService:   UnusedFormationTemplateService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationTemplateSvc := testCase.FormationTemplateService()
			formationTemplateConverter := testCase.FormationTemplateConverter()

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc, nil, nil, nil)

			// WHEN
			result, err := resolver.FormationTemplate(ctx, testFormationTemplateID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationTemplateSvc, formationTemplateConverter)
		})
	}
}

func TestResolver_FormationTemplatesByName(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)
	testFormationTemplateName := formationTemplateName

	first := 2
	after := "test"
	gqlAfter := graphql.PageCursor(after)

	testCases := []struct {
		Name                       string
		First                      *int
		TxFn                       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationTemplateConverter func() *automock.FormationTemplateConverter
		FormationTemplateService   func() *automock.FormationTemplateService
		ExpectedOutput             *graphql.FormationTemplatePage
		ExpectedError              error
	}{
		{
			Name:  "Success",
			TxFn:  txGen.ThatSucceeds,
			First: &first,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), nilLabelFilters, &testFormationTemplateName, first, after).Return(&formationTemplateModelPage, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("MultipleToGraphQL", formationTemplateModelPage.Data).Return(graphQLFormationTemplatePage.Data, nil)

				return converter
			},
			ExpectedOutput: &graphQLFormationTemplatePage,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when listing from service fails",
			First: &first,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), nilLabelFilters, &testFormationTemplateName, first, after).Return(nil, testErr)

				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:  "Error when converting to graphql fails",
			First: &first,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), nilLabelFilters, &testFormationTemplateName, first, after).Return(&formationTemplateModelPage, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("MultipleToGraphQL", formationTemplateModelPage.Data).Return(nil, testErr)

				return converter
			},
			ExpectedError: testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			First: &first,
			TxFn:  txGen.ThatFailsOnCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), nilLabelFilters, &testFormationTemplateName, first, after).Return(&formationTemplateModelPage, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("MultipleToGraphQL", formationTemplateModelPage.Data).Return(graphQLFormationTemplatePage.Data, nil)

				return converter
			},
			ExpectedError: testErr,
		},
		{
			Name:          "Returns error missing first parameter",
			First:         nil,
			TxFn:          txGen.ThatDoesntExpectCommit,
			ExpectedError: apperrors.NewInvalidDataError("missing required parameter: 'first'"),
		},
		{
			Name:          "Returns error when failing on the beginning of a transaction",
			First:         &first,
			TxFn:          txGen.ThatFailsOnBegin,
			ExpectedError: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()

			formationTemplateSvc := UnusedFormationTemplateService()
			if testCase.FormationTemplateService != nil {
				formationTemplateSvc = testCase.FormationTemplateService()
			}

			formationTemplateConverter := UnusedFormationTemplateConverter()
			if testCase.FormationTemplateConverter != nil {
				formationTemplateConverter = testCase.FormationTemplateConverter()
			}

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc, nil, nil, nil)

			// WHEN
			result, err := resolver.FormationTemplatesByName(ctx, &testFormationTemplateName, testCase.First, &gqlAfter)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationTemplateSvc, formationTemplateConverter)
		})
	}
}

func TestResolver_FormationTemplates(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	after := "test"
	gqlAfter := graphql.PageCursor(after)

	testCases := []struct {
		Name                       string
		First                      *int
		TxFn                       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationTemplateConverter func() *automock.FormationTemplateConverter
		FormationTemplateService   func() *automock.FormationTemplateService
		ExpectedOutput             *graphql.FormationTemplatePage
		ExpectedError              error
	}{
		{
			Name:  "Success",
			TxFn:  txGen.ThatSucceeds,
			First: &first,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), emptyLabelFilters, nilStr, first, after).Return(&formationTemplateModelPage, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("MultipleToGraphQL", formationTemplateModelPage.Data).Return(graphQLFormationTemplatePage.Data, nil)

				return converter
			},
			ExpectedOutput: &graphQLFormationTemplatePage,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when listing from service fails",
			First: &first,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), emptyLabelFilters, nilStr, first, after).Return(nil, testErr)

				return svc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when converting to graphql fails",
			First: &first,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), emptyLabelFilters, nilStr, first, after).Return(&formationTemplateModelPage, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("MultipleToGraphQL", formationTemplateModelPage.Data).Return(nil, testErr)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			First: &first,
			TxFn:  txGen.ThatFailsOnCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), emptyLabelFilters, nilStr, first, after).Return(&formationTemplateModelPage, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("MultipleToGraphQL", formationTemplateModelPage.Data).Return(graphQLFormationTemplatePage.Data, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:           "Returns error missing first parameter",
			First:          nil,
			TxFn:           txGen.ThatDoesntExpectCommit,
			ExpectedOutput: nil,
			ExpectedError:  apperrors.NewInvalidDataError("missing required parameter: 'first'"),
		},
		{
			Name:           "Returns error when failing on the beginning of a transaction",
			First:          &first,
			TxFn:           txGen.ThatFailsOnBegin,
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()

			formationTemplateSvc := UnusedFormationTemplateService()
			if testCase.FormationTemplateService != nil {
				formationTemplateSvc = testCase.FormationTemplateService()
			}

			formationTemplateConverter := UnusedFormationTemplateConverter()
			if testCase.FormationTemplateConverter != nil {
				formationTemplateConverter = testCase.FormationTemplateConverter()
			}

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc, nil, nil, nil)

			// WHEN
			result, err := resolver.FormationTemplates(ctx, nil, testCase.First, &gqlAfter)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationTemplateSvc, formationTemplateConverter)
		})
	}
}

func TestResolver_UpdateFormationTemplate(t *testing.T) {
	modelWithoutWebhooks := formationTemplateModel
	modelWithoutWebhooks.Webhooks = nil

	graphQLFormationTemplateWithoutWebhooks := graphQLFormationTemplate
	graphQLFormationTemplateWithoutWebhooks.Webhooks = nil

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		Input                      graphql.FormationTemplateUpdateInput
		TxFn                       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationTemplateConverter func() *automock.FormationTemplateConverter
		FormationTemplateService   func() *automock.FormationTemplateService
		ExpectedOutput             *graphql.FormationTemplate
		ExpectedError              error
	}{
		{
			Name:  "Success",
			TxFn:  txGen.ThatSucceeds,
			Input: formationTemplateUpdateInputGraphQL,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testFormationTemplateID, &formationTemplateUpdateInputModel).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromUpdateInputGraphQL", &formationTemplateUpdateInputGraphQL).Return(&formationTemplateUpdateInputModel, nil)
				converter.On("ToGraphQL", &modelWithoutWebhooks).Return(&graphQLFormationTemplateWithoutWebhooks, nil)

				return converter
			},
			ExpectedOutput: &graphQLFormationTemplateWithoutWebhooks,
			ExpectedError:  nil,
		},
		{
			Name:                     "Error when converting from graphql fails",
			Input:                    formationTemplateUpdateInputGraphQL,
			TxFn:                     txGen.ThatDoesntExpectCommit,
			FormationTemplateService: UnusedFormationTemplateService,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromUpdateInputGraphQL", &formationTemplateUpdateInputGraphQL).Return(nil, testErr)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when updating call in service fails",
			Input: formationTemplateUpdateInputGraphQL,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testFormationTemplateID, &formationTemplateUpdateInputModel).Return(testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromUpdateInputGraphQL", &formationTemplateUpdateInputGraphQL).Return(&formationTemplateUpdateInputModel, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when get call in service fails",
			Input: formationTemplateUpdateInputGraphQL,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testFormationTemplateID, &formationTemplateUpdateInputModel).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromUpdateInputGraphQL", &formationTemplateUpdateInputGraphQL).Return(&formationTemplateUpdateInputModel, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when converting to graphql fails",
			Input: formationTemplateUpdateInputGraphQL,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testFormationTemplateID, &formationTemplateUpdateInputModel).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromUpdateInputGraphQL", &formationTemplateUpdateInputGraphQL).Return(&formationTemplateUpdateInputModel, nil)
				converter.On("ToGraphQL", &modelWithoutWebhooks).Return(nil, testErr)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			TxFn:  txGen.ThatFailsOnCommit,
			Input: formationTemplateUpdateInputGraphQL,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testFormationTemplateID, &formationTemplateUpdateInputModel).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromUpdateInputGraphQL", &formationTemplateUpdateInputGraphQL).Return(&formationTemplateUpdateInputModel, nil)
				converter.On("ToGraphQL", &modelWithoutWebhooks).Return(&graphQLFormationTemplateWithoutWebhooks, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                       "Returns error when input validation fails",
			Input:                      graphql.FormationTemplateUpdateInput{},
			TxFn:                       txGen.ThatDoesntExpectCommit,
			FormationTemplateService:   UnusedFormationTemplateService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              errors.New("cannot be blank"),
		},
		{
			Name:                       "Returns error when failing on the beginning of a transaction",
			Input:                      formationTemplateUpdateInputGraphQL,
			TxFn:                       txGen.ThatFailsOnBegin,
			FormationTemplateService:   UnusedFormationTemplateService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationTemplateSvc := testCase.FormationTemplateService()
			formationTemplateConverter := testCase.FormationTemplateConverter()

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc, nil, nil, nil)

			// WHEN
			result, err := resolver.UpdateFormationTemplate(ctx, testFormationTemplateID, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationTemplateSvc, formationTemplateConverter)
		})
	}
}

func TestResolver_DeleteFormationTemplate(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		TxFn                       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationTemplateConverter func() *automock.FormationTemplateConverter
		FormationTemplateService   func() *automock.FormationTemplateService
		ExpectedOutput             *graphql.FormationTemplate
		ExpectedError              error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&formationTemplateModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("ToGraphQL", &formationTemplateModel).Return(&graphQLFormationTemplate, nil)

				return converter
			},
			ExpectedOutput: &graphQLFormationTemplate,
			ExpectedError:  nil,
		},
		{
			Name: "Error when get call in service fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
		{
			Name: "Error when delete call in service fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&formationTemplateModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(testErr)

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
		{
			Name: "Error when converting to graphql fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&formationTemplateModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("ToGraphQL", &formationTemplateModel).Return(nil, testErr)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&formationTemplateModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("ToGraphQL", &formationTemplateModel).Return(&graphQLFormationTemplate, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                       "Returns error when failing on the beginning of a transaction",
			TxFn:                       txGen.ThatFailsOnBegin,
			FormationTemplateService:   UnusedFormationTemplateService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationTemplateSvc := testCase.FormationTemplateService()
			formationTemplateConverter := testCase.FormationTemplateConverter()

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc, nil, nil, nil)

			// WHEN
			result, err := resolver.DeleteFormationTemplate(ctx, testFormationTemplateID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationTemplateSvc, formationTemplateConverter)
		})
	}
}

func TestResolver_CreateFormationTemplate(t *testing.T) {
	// removing the webhooks below because they do not affect the flow in any way but their validation is difficult to set up properly
	gqlInputWithoutWebhooks := formationTemplateRegisterInputGraphQL
	gqlInputWithoutWebhooks.Webhooks = nil

	modelInputWithoutWebhooks := formationTemplateRegisterInputModel
	modelInputWithoutWebhooks.Webhooks = nil

	modelWithoutWebhooks := formationTemplateModel
	modelWithoutWebhooks.Webhooks = nil

	graphQLFormationTemplateWithoutWebhooks := graphQLFormationTemplate
	graphQLFormationTemplateWithoutWebhooks.Webhooks = nil

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		Input                      graphql.FormationTemplateRegisterInput
		TxFn                       func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationTemplateConverter func() *automock.FormationTemplateConverter
		FormationTemplateService   func() *automock.FormationTemplateService
		ExpectedOutput             *graphql.FormationTemplate
		ExpectedError              error
	}{
		{
			Name:  "Success",
			TxFn:  txGen.ThatSucceeds,
			Input: gqlInputWithoutWebhooks,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), &modelInputWithoutWebhooks).Return(testFormationTemplateID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromRegisterInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)
				converter.On("ToGraphQL", &modelWithoutWebhooks).Return(&graphQLFormationTemplateWithoutWebhooks, nil)

				return converter
			},
			ExpectedOutput: &graphQLFormationTemplateWithoutWebhooks,
			ExpectedError:  nil,
		},
		{
			Name:                     "Error when converting from graphql fails",
			Input:                    gqlInputWithoutWebhooks,
			TxFn:                     txGen.ThatDoesntExpectCommit,
			FormationTemplateService: UnusedFormationTemplateService,
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromRegisterInputGraphQL", &gqlInputWithoutWebhooks).Return(nil, testErr)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when creating call to service fails",
			Input: gqlInputWithoutWebhooks,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), &modelInputWithoutWebhooks).Return("", testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromRegisterInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when get call to service fails",
			Input: gqlInputWithoutWebhooks,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), &modelInputWithoutWebhooks).Return(testFormationTemplateID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromRegisterInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when converting to graphql fails",
			Input: gqlInputWithoutWebhooks,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), &modelInputWithoutWebhooks).Return(testFormationTemplateID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromRegisterInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)
				converter.On("ToGraphQL", &modelWithoutWebhooks).Return(nil, testErr)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			TxFn:  txGen.ThatFailsOnCommit,
			Input: gqlInputWithoutWebhooks,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), &modelInputWithoutWebhooks).Return(testFormationTemplateID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testFormationTemplateID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromRegisterInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)
				converter.On("ToGraphQL", &modelWithoutWebhooks).Return(&graphQLFormationTemplateWithoutWebhooks, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                       "Returns error when input validation fails",
			Input:                      graphql.FormationTemplateRegisterInput{},
			TxFn:                       txGen.ThatDoesntExpectCommit,
			FormationTemplateService:   UnusedFormationTemplateService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              errors.New("cannot be blank"),
		},
		{
			Name:                       "Returns error when failing on the beginning of a transaction",
			Input:                      gqlInputWithoutWebhooks,
			TxFn:                       txGen.ThatFailsOnBegin,
			FormationTemplateService:   UnusedFormationTemplateService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationTemplateSvc := testCase.FormationTemplateService()
			formationTemplateConverter := testCase.FormationTemplateConverter()

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc, nil, nil, nil)

			// WHEN
			result, err := resolver.CreateFormationTemplate(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationTemplateSvc, formationTemplateConverter)
		})
	}
}

func TestResolver_Webhooks(t *testing.T) {
	modelWebhooks := []*model.Webhook{fixFormationTemplateModelWebhook()}
	gqlWebhooks := []*graphql.Webhook{fixFormationTemplateGQLWebhook()}

	txGen := txtest.NewTransactionContextGenerator(testErr)
	testCases := []struct {
		Name                     string
		TxFn                     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		Input                    *graphql.FormationTemplate
		WebhookConverter         func() *automock.WebhookConverter
		FormationTemplateService func() *automock.FormationTemplateService
		ExpectedOutput           []*graphql.Webhook
		ExpectedError            error
	}{
		{
			Name:  "Success",
			TxFn:  txGen.ThatSucceeds,
			Input: &graphQLFormationTemplate,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("ListWebhooksForFormationTemplate", txtest.CtxWithDBMatcher(), graphQLFormationTemplate.ID).Return(modelWebhooks, nil)
				return svc
			},
			WebhookConverter: func() *automock.WebhookConverter {
				converter := &automock.WebhookConverter{}
				converter.On("MultipleToGraphQL", modelWebhooks).Return(gqlWebhooks, nil)
				return converter
			},
			ExpectedOutput: gqlWebhooks,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when listing webhooks fails",
			TxFn:  txGen.ThatDoesntExpectCommit,
			Input: &graphQLFormationTemplate,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("ListWebhooksForFormationTemplate", txtest.CtxWithDBMatcher(), graphQLFormationTemplate.ID).Return(nil, testErr)
				return svc
			},
			WebhookConverter: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when converting webhooks fails",
			TxFn:  txGen.ThatDoesntExpectCommit,
			Input: &graphQLFormationTemplate,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("ListWebhooksForFormationTemplate", txtest.CtxWithDBMatcher(), graphQLFormationTemplate.ID).Return(modelWebhooks, nil)
				return svc
			},
			WebhookConverter: func() *automock.WebhookConverter {
				converter := &automock.WebhookConverter{}
				converter.On("MultipleToGraphQL", modelWebhooks).Return(nil, testErr)
				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			TxFn:  txGen.ThatFailsOnCommit,
			Input: &graphQLFormationTemplate,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("ListWebhooksForFormationTemplate", txtest.CtxWithDBMatcher(), graphQLFormationTemplate.ID).Return(modelWebhooks, nil)
				return svc
			},
			WebhookConverter: func() *automock.WebhookConverter {
				converter := &automock.WebhookConverter{}
				converter.On("MultipleToGraphQL", modelWebhooks).Return(gqlWebhooks, nil)
				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when failing on the beginning of a transaction",
			TxFn:  txGen.ThatFailsOnBegin,
			Input: &graphQLFormationTemplate,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("ListWebhooksForFormationTemplate", txtest.CtxWithDBMatcher(), graphQLFormationTemplate.ID).Return(modelWebhooks, nil)
				return svc
			},
			WebhookConverter: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when input formation template is nil",
			TxFn:  txGen.ThatFailsOnBegin,
			Input: nil,
			FormationTemplateService: func() *automock.FormationTemplateService {
				return &automock.FormationTemplateService{}
			},
			WebhookConverter: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedOutput: nil,
			ExpectedError:  apperrors.NewInternalError("Formation Template cannot be empty"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationTemplateSvc := testCase.FormationTemplateService()
			whConv := testCase.WebhookConverter()

			resolver := formationtemplate.NewResolver(transact, nil, formationTemplateSvc, whConv, nil, nil)

			// WHEN
			result, err := resolver.Webhooks(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, whConv)
		})
	}
}

func TestResolver_FormationConstraints(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	formationConstraintIDs := []string{constraintID1, constraintID2}

	formationConstraintsModel := [][]*model.FormationConstraint{{formationConstraint1}, {formationConstraint2}}

	formationConstraintsGql := [][]*graphql.FormationConstraint{{formationConstraintGql1}, {formationConstraintGql2}}

	testCases := []struct {
		Name                         string
		TxFn                         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		Input                        []dataloader.ParamFormationConstraint
		FormationConstraintSvc       func() *automock.FormationConstraintService
		FormationConstraintConverter func() *automock.FormationConstraintConverter
		ExpectedConstraints          [][]*graphql.FormationConstraint
		ExpectedErrors               []error
	}{
		{
			Name:  "Success",
			TxFn:  txGen.ThatSucceeds,
			Input: []dataloader.ParamFormationConstraint{{ID: formationConstraintIDs[0], Ctx: ctx}, {ID: formationConstraintIDs[1], Ctx: ctx}},
			FormationConstraintSvc: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("ListByFormationTemplateIDs", txtest.CtxWithDBMatcher(), formationConstraintIDs).Return(formationConstraintsModel, nil).Once()
				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("MultipleToGraphQL", formationConstraintsModel[0]).Return(formationConstraintsGql[0]).Once()
				converter.On("MultipleToGraphQL", formationConstraintsModel[1]).Return(formationConstraintsGql[1]).Once()
				return converter
			},
			ExpectedConstraints: formationConstraintsGql,
			ExpectedErrors:      nil,
		},
		{
			Name:  "Returns error if commit fails",
			TxFn:  txGen.ThatFailsOnCommit,
			Input: []dataloader.ParamFormationConstraint{{ID: formationConstraintIDs[0], Ctx: ctx}, {ID: formationConstraintIDs[1], Ctx: ctx}},
			FormationConstraintSvc: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("ListByFormationTemplateIDs", txtest.CtxWithDBMatcher(), formationConstraintIDs).Return(formationConstraintsModel, nil).Once()
				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("MultipleToGraphQL", formationConstraintsModel[0]).Return(formationConstraintsGql[0]).Once()
				converter.On("MultipleToGraphQL", formationConstraintsModel[1]).Return(formationConstraintsGql[1]).Once()
				return converter
			},
			ExpectedConstraints: nil,
			ExpectedErrors:      []error{testErr},
		},
		{
			Name:  "Returns error when listing the formation templates by ids fail",
			TxFn:  txGen.ThatDoesntExpectCommit,
			Input: []dataloader.ParamFormationConstraint{{ID: formationConstraintIDs[0], Ctx: ctx}, {ID: formationConstraintIDs[1], Ctx: ctx}},
			FormationConstraintSvc: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("ListByFormationTemplateIDs", txtest.CtxWithDBMatcher(), formationConstraintIDs).Return(nil, testErr).Once()
				return svc
			},
			ExpectedConstraints: nil,
			ExpectedErrors:      []error{testErr},
		},
		{
			Name:                "Returns error when can't start transaction",
			TxFn:                txGen.ThatFailsOnBegin,
			Input:               []dataloader.ParamFormationConstraint{{ID: formationConstraintIDs[0], Ctx: ctx}, {ID: formationConstraintIDs[1], Ctx: ctx}},
			ExpectedConstraints: nil,
			ExpectedErrors:      []error{testErr},
		},
		{
			Name:                "Returns error when input does not contain formation templates",
			TxFn:                txGen.ThatDoesntStartTransaction,
			Input:               []dataloader.ParamFormationConstraint{},
			ExpectedConstraints: nil,
			ExpectedErrors:      []error{apperrors.NewInternalError("No Formation Templates found")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			svc := UnusedFormationConstraintService()
			if testCase.FormationConstraintSvc != nil {
				svc = testCase.FormationConstraintSvc()
			}
			converter := UnusedFormationConstraintConverter()
			if testCase.FormationConstraintConverter != nil {
				converter = testCase.FormationConstraintConverter()
			}

			resolver := formationtemplate.NewResolver(transact, nil, nil, nil, svc, converter)

			res, errs := resolver.FormationConstraintsDataLoader(testCase.Input)
			if testCase.ExpectedErrors != nil {
				assert.Error(t, errs[0])
				assert.Nil(t, res)
			} else {
				require.Nil(t, errs)
				reflect.DeepEqual(res, testCase.ExpectedConstraints)
			}

			mock.AssertExpectationsForObjects(t, persist, svc, converter)
		})
	}
}

func TestResolver_SetFormationTemplateLabel(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                     string
		Input                    graphql.LabelInput
		TxFn                     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationTemplateService func() *automock.FormationTemplateService
		ExpectedOutput           *graphql.Label
		ExpectedError            error
	}{
		{
			Name:  "Success",
			Input: lblInput,
			TxFn:  txGen.ThatSucceeds,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("SetLabel", txtest.CtxWithDBMatcher(), formationTemplateLabelInput).Return(nil)
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(modelLabel, nil)
				return svc
			},
			ExpectedOutput: gqlLabel,
		},
		{
			Name:  "Error when setting label fails",
			Input: lblInput,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("SetLabel", txtest.CtxWithDBMatcher(), formationTemplateLabelInput).Return(testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:  "Error when getting label fails",
			Input: lblInput,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("SetLabel", txtest.CtxWithDBMatcher(), formationTemplateLabelInput).Return(nil)
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(nil, testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:          "Error when beginning of a transaction fails",
			Input:         lblInput,
			TxFn:          txGen.ThatFailsOnBegin,
			ExpectedError: testErr,
		},
		{
			Name:  "Error when committing transaction fails",
			Input: lblInput,
			TxFn:  txGen.ThatFailsOnCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("SetLabel", txtest.CtxWithDBMatcher(), formationTemplateLabelInput).Return(nil)
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(modelLabel, nil)
				return svc
			},
			ExpectedError: testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()

			formationTemplateSvc := UnusedFormationTemplateService()
			if testCase.FormationTemplateService != nil {
				formationTemplateSvc = testCase.FormationTemplateService()
			}

			resolver := formationtemplate.NewResolver(transact, nil, formationTemplateSvc, nil, nil, nil)

			// WHEN
			result, err := resolver.SetFormationTemplateLabel(ctx, testFormationTemplateID, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationTemplateSvc)
		})
	}
}

func TestResolver_DeleteFormationTemplateLabel(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                     string
		TxFn                     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationTemplateService func() *automock.FormationTemplateService
		ExpectedOutput           *graphql.Label
		ExpectedError            error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(modelLabel, nil)
				svc.On("DeleteLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(nil)
				return svc
			},
			ExpectedOutput: gqlLabel,
		},
		{
			Name: "Error when getting label fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(nil, testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name: "Error when deleting label fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(modelLabel, nil)
				svc.On("DeleteLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:          "Error when beginning of a transaction fails",
			TxFn:          txGen.ThatFailsOnBegin,
			ExpectedError: testErr,
		},
		{
			Name: "Error when committing transaction fails",
			TxFn: txGen.ThatFailsOnCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("GetLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(modelLabel, nil)
				svc.On("DeleteLabel", txtest.CtxWithDBMatcher(), testFormationTemplateID, lblInput.Key).Return(nil)
				return svc
			},
			ExpectedError: testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()

			formationTemplateSvc := UnusedFormationTemplateService()
			if testCase.FormationTemplateService != nil {
				formationTemplateSvc = testCase.FormationTemplateService()
			}

			resolver := formationtemplate.NewResolver(transact, nil, formationTemplateSvc, nil, nil, nil)

			// WHEN
			result, err := resolver.DeleteFormationTemplateLabel(ctx, testFormationTemplateID, testLabelKey)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationTemplateSvc)
		})
	}
}

func TestResolver_Labels(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                     string
		TxFn                     func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		Input                    *graphql.FormationTemplate
		FormationTemplateService func() *automock.FormationTemplateService
		ExpectedOutput           graphql.Labels
		ExpectedError            error
	}{
		{
			Name:  "Success",
			TxFn:  txGen.ThatSucceeds,
			Input: &graphQLFormationTemplate,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("ListLabels", txtest.CtxWithDBMatcher(), graphQLFormationTemplate.ID).Return(modelLabels, nil)
				return svc
			},
			ExpectedOutput: gqlLabels,
		},
		{
			Name:  "Success when listing labels returns not found error",
			TxFn:  txGen.ThatSucceeds,
			Input: &graphQLFormationTemplate,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("ListLabels", txtest.CtxWithDBMatcher(), graphQLFormationTemplate.ID).Return(nil, formationTemplateNotFoundErr)
				return svc
			},
			ExpectedOutput: nil,
		},
		{
			Name:  "Returns error when listing labels fails",
			TxFn:  txGen.ThatDoesntExpectCommit,
			Input: &graphQLFormationTemplate,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("ListLabels", txtest.CtxWithDBMatcher(), graphQLFormationTemplate.ID).Return(nil, testErr)
				return svc
			},
			ExpectedError: testErr,
		},
		{
			Name:          "Returns error when input formation template is nil",
			TxFn:          txGen.ThatFailsOnBegin,
			Input:         nil,
			ExpectedError: apperrors.NewInternalError("Formation Template cannot be empty"),
		},
		{
			Name:          "Error when beginning of a transaction fails",
			TxFn:          txGen.ThatFailsOnBegin,
			Input:         &graphQLFormationTemplate,
			ExpectedError: testErr,
		},
		{
			Name:  "Error when committing transaction fails",
			TxFn:  txGen.ThatFailsOnCommit,
			Input: &graphQLFormationTemplate,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("ListLabels", txtest.CtxWithDBMatcher(), graphQLFormationTemplate.ID).Return(modelLabels, nil)
				return svc
			},
			ExpectedError: testErr,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationTemplateSvc := UnusedFormationTemplateService()
			if testCase.FormationTemplateService != nil {
				formationTemplateSvc = testCase.FormationTemplateService()
			}

			resolver := formationtemplate.NewResolver(transact, nil, formationTemplateSvc, nil, nil, nil)

			// WHEN
			result, err := resolver.Labels(ctx, testCase.Input, &testLabelKey)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationTemplateSvc)
		})
	}
}
