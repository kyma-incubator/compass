package formationtemplate_test

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResolver_ApplicationTemplate(t *testing.T) {
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
				converter.On("ToGraphQL", &formationTemplateModel).Return(&formationTemplateGraphQLModel, nil)

				return converter
			},
			ExpectedOutput: &formationTemplateGraphQLModel,
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
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
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

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc)

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

func TestResolver_ApplicationTemplates(t *testing.T) {
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
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(formationTemplateModelPage, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("MultipleToGraphQL", formationTemplateModelPage.Data).Return(formationTemplateGraphQLModelPage.Data)

				return converter
			},
			ExpectedOutput: &formationTemplateGraphQLModelPage,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when listing from service fails",
			First: &first,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(model.FormationTemplatePage{}, testErr)

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			First: &first,
			TxFn:  txGen.ThatFailsOnCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(formationTemplateModelPage, nil)

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
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

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc)

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
			Input: inputFormationTemplateGraphQLModel,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, inputFormationTemplateModel).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &inputFormationTemplateGraphQLModel).Return(&inputFormationTemplateModel, nil)
				converter.On("ToGraphQL", &formationTemplateModel).Return(&formationTemplateGraphQLModel, nil)

				return converter
			},
			ExpectedOutput: &formationTemplateGraphQLModel,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when updating call in service fails",
			Input: inputFormationTemplateGraphQLModel,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, inputFormationTemplateModel).Return(testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &inputFormationTemplateGraphQLModel).Return(&inputFormationTemplateModel, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when get call in service fails",
			Input: inputFormationTemplateGraphQLModel,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, inputFormationTemplateModel).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &inputFormationTemplateGraphQLModel).Return(&inputFormationTemplateModel, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			TxFn:  txGen.ThatFailsOnCommit,
			Input: inputFormationTemplateGraphQLModel,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, inputFormationTemplateModel).Return(nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &inputFormationTemplateGraphQLModel).Return(&inputFormationTemplateModel, nil)

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
			Input:                      inputFormationTemplateGraphQLModel,
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

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc)

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
				converter.On("ToGraphQL", &formationTemplateModel).Return(&formationTemplateGraphQLModel, nil)

				return converter
			},
			ExpectedOutput: &formationTemplateGraphQLModel,
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
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil)

				return svc
			},
			FormationTemplateConverter: UnusedFormationTemplateConverter,
			ExpectedOutput:             nil,
			ExpectedError:              testErr,
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

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc)

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
			Input: inputFormationTemplateGraphQLModel,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), inputFormationTemplateModel).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &inputFormationTemplateGraphQLModel).Return(&inputFormationTemplateModel, nil)
				converter.On("ToGraphQL", &formationTemplateModel).Return(&formationTemplateGraphQLModel, nil)

				return converter
			},
			ExpectedOutput: &formationTemplateGraphQLModel,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when creating call to service fails",
			Input: inputFormationTemplateGraphQLModel,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), inputFormationTemplateModel).Return("", testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &inputFormationTemplateGraphQLModel).Return(&inputFormationTemplateModel, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when get call to service fails",
			Input: inputFormationTemplateGraphQLModel,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), inputFormationTemplateModel).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &inputFormationTemplateGraphQLModel).Return(&inputFormationTemplateModel, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			TxFn:  txGen.ThatFailsOnCommit,
			Input: inputFormationTemplateGraphQLModel,
			FormationTemplateService: func() *automock.FormationTemplateService {
				svc := &automock.FormationTemplateService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), inputFormationTemplateModel).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(&formationTemplateModel, nil)

				return svc
			},
			FormationTemplateConverter: func() *automock.FormationTemplateConverter {
				converter := &automock.FormationTemplateConverter{}
				converter.On("FromInputGraphQL", &inputFormationTemplateGraphQLModel).Return(&inputFormationTemplateModel, nil)

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
			Input:                      inputFormationTemplateGraphQLModel,
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

			resolver := formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc)

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
