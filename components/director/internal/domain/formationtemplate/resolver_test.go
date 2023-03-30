package formationtemplate_test

import (
	"context"
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
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolver_FormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

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
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)

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
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
		{
			Name: "Returns nil when formation template not found",
			TxFn: txGen.ThatSucceeds,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, apperrors.NewNotFoundError(resource.FormationTemplate, testID))

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              nil,
		},
		{
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)

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
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)

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
			result, err := resolver.FormationTemplate(ctx, testID)

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
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

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
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(&formationTemplateModelPage, nil)

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
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
		{
			Name:  "Error when converting to graphql fails",
			First: &first,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(&formationTemplateModelPage, nil)

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
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(&formationTemplateModelPage, nil)

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
			Name:                       "Returns error missing first parameter",
			First:                      nil,
			TxFn:                       txGen.ThatDoesntExpectCommit,
			FormationTemplateService:   UnusedFormationTemplateService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              apperrors.NewInvalidDataError("missing required parameter 'first'"),
		},
		{
			Name:                       "Returns error when failing on the beginning of a transaction",
			First:                      &first,
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
			result, err := resolver.FormationTemplates(ctx, testCase.First, &gqlAfter)

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
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	// removing the webhooks below because they do not affect the flow in any way but their validation is difficult to set up properly
	gqlInputWithoutWebhooks := formationTemplateGraphQLInput
	gqlInputWithoutWebhooks.Webhooks = nil

	modelInputWithoutWebhooks := formationTemplateModelInput
	modelInputWithoutWebhooks.Webhooks = nil

	modelWithoutWebhooks := formationTemplateModel
	modelWithoutWebhooks.Webhooks = nil

	graphQLFormationTemplateWithoutWebhooks := graphQLFormationTemplate
	graphQLFormationTemplateWithoutWebhooks.Webhooks = nil

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		Input                      graphql.FormationTemplateInput
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
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, &modelInputWithoutWebhooks).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)
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
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(nil, testErr)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when updating call in service fails",
			Input: gqlInputWithoutWebhooks,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, &modelInputWithoutWebhooks).Return(testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when get call in service fails",
			Input: gqlInputWithoutWebhooks,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, &modelInputWithoutWebhooks).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)

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
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, &modelInputWithoutWebhooks).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)
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
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, &modelInputWithoutWebhooks).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)
				converter.On("ToGraphQL", &modelWithoutWebhooks).Return(&graphQLFormationTemplateWithoutWebhooks, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                       "Returns error when input validation fails",
			Input:                      graphql.FormationTemplateInput{},
			TxFn:                       txGen.ThatDoesntExpectCommit,
			FormationTemplateService:   UnusedFormationTemplateService,
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              errors.New("cannot be blank"),
		},
		{
			Name:                       "Returns error when failing on the beginning of a transaction",
			Input:                      formationTemplateGraphQLInput,
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
			result, err := resolver.UpdateFormationTemplate(ctx, testID, testCase.Input)

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
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

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
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil)

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
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr)

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
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(testErr)

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
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil)

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
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil)

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
			result, err := resolver.DeleteFormationTemplate(ctx, testID)

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
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	// removing the webhooks below because they do not affect the flow in any way but their validation is difficult to set up properly
	gqlInputWithoutWebhooks := formationTemplateGraphQLInput
	gqlInputWithoutWebhooks.Webhooks = nil

	modelInputWithoutWebhooks := formationTemplateModelInput
	modelInputWithoutWebhooks.Webhooks = nil

	modelWithoutWebhooks := formationTemplateModel
	modelWithoutWebhooks.Webhooks = nil

	graphQLFormationTemplateWithoutWebhooks := graphQLFormationTemplate
	graphQLFormationTemplateWithoutWebhooks.Webhooks = nil

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                       string
		Input                      graphql.FormationTemplateInput
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
				svc.On("Create", txtest.CtxWithDBMatcher(), &modelInputWithoutWebhooks).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)
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
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(nil, testErr)

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
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)

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
				svc.On("Create", txtest.CtxWithDBMatcher(), &modelInputWithoutWebhooks).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)

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
				svc.On("Create", txtest.CtxWithDBMatcher(), &modelInputWithoutWebhooks).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)
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
				svc.On("Create", txtest.CtxWithDBMatcher(), &modelInputWithoutWebhooks).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&modelWithoutWebhooks, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &gqlInputWithoutWebhooks).Return(&modelInputWithoutWebhooks, nil)
				converter.On("ToGraphQL", &modelWithoutWebhooks).Return(&graphQLFormationTemplateWithoutWebhooks, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                       "Returns error when input validation fails",
			Input:                      graphql.FormationTemplateInput{},
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
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

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
	ctx := context.TODO()
	txGen := txtest.NewTransactionContextGenerator(testErr)

	formationConstraintIDs := []string{"123", "345"}

	firstFormationConstraintModel := &model.FormationConstraint{
		ID:              formationConstraintIDs[0],
		Name:            "test",
		ConstraintType:  "test",
		TargetOperation: "test",
		Operator:        "test",
		ResourceType:    "test",
		ResourceSubtype: "test",
		InputTemplate:   "test",
		ConstraintScope: "test",
	}
	secondFormationConstraintModel := &model.FormationConstraint{
		ID:              formationConstraintIDs[1],
		Name:            "test",
		ConstraintType:  "test",
		TargetOperation: "test",
		Operator:        "test",
		ResourceType:    "test",
		ResourceSubtype: "test",
		InputTemplate:   "test",
		ConstraintScope: "test",
	}
	formationConstraintsModel := [][]*model.FormationConstraint{{firstFormationConstraintModel}, {secondFormationConstraintModel}}

	firstFormationConstraintGql := &graphql.FormationConstraint{
		ID:              formationConstraintIDs[0],
		Name:            "test",
		ConstraintType:  "test",
		TargetOperation: "test",
		Operator:        "test",
		ResourceType:    "test",
		ResourceSubtype: "test",
		InputTemplate:   "test",
		ConstraintScope: "test",
	}
	secondFormationConstraintGql := &graphql.FormationConstraint{
		ID:              formationConstraintIDs[1],
		Name:            "test",
		ConstraintType:  "test",
		TargetOperation: "test",
		Operator:        "test",
		ResourceType:    "test",
		ResourceSubtype: "test",
		InputTemplate:   "test",
		ConstraintScope: "test",
	}
	formationConstraintsGql := [][]*graphql.FormationConstraint{{firstFormationConstraintGql}, {secondFormationConstraintGql}}

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
				svc.On("ListByFormationTemplateIDs", txtest.CtxWithDBMatcher(), formationConstraintIDs).Return(formationConstraintsModel, nil)
				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("MultipleToGraphQL", formationConstraintsModel[0]).Return(formationConstraintsGql[0])
				converter.On("MultipleToGraphQL", formationConstraintsModel[1]).Return(formationConstraintsGql[1])
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
				svc.On("ListByFormationTemplateIDs", txtest.CtxWithDBMatcher(), formationConstraintIDs).Return(formationConstraintsModel, nil)
				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("MultipleToGraphQL", formationConstraintsModel[0]).Return(formationConstraintsGql[0])
				converter.On("MultipleToGraphQL", formationConstraintsModel[1]).Return(formationConstraintsGql[1])
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
				svc.On("ListByFormationTemplateIDs", txtest.CtxWithDBMatcher(), formationConstraintIDs).Return(nil, testErr)
				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedConstraints:          nil,
			ExpectedErrors:               []error{testErr},
		},
		{
			Name:                         "Returns error when can't start transaction",
			TxFn:                         txGen.ThatFailsOnBegin,
			Input:                        []dataloader.ParamFormationConstraint{{ID: formationConstraintIDs[0], Ctx: ctx}, {ID: formationConstraintIDs[1], Ctx: ctx}},
			FormationConstraintSvc:       UnusedFormationConstraintService,
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedConstraints:          nil,
			ExpectedErrors:               []error{testErr},
		},
		{
			Name:                         "Returns error when input does not contain formation templates",
			TxFn:                         txGen.ThatDoesntStartTransaction,
			Input:                        []dataloader.ParamFormationConstraint{},
			FormationConstraintSvc:       UnusedFormationConstraintService,
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedConstraints:          nil,
			ExpectedErrors:               []error{apperrors.NewInternalError("No Formation Templates found")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			svc := testCase.FormationConstraintSvc()
			converter := testCase.FormationConstraintConverter()

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
