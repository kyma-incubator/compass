package formationconstraint_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/automock"
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
)

func TestResolver_FormationConstraints(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	formationConstraints := []*model.FormationConstraint{
		{Name: "test"},
		{Name: "test2"},
	}

	formationConstraintsGql := []*graphql.FormationConstraint{
		{Name: "test"},
		{Name: "test2"},
	}

	testCases := []struct {
		Name                         string
		TxFn                         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationConstraintConverter func() *automock.FormationConstraintConverter
		FormationConstraintService   func() *automock.FormationConstraintService
		ExpectedOutput               []*graphql.FormationConstraint
		ExpectedError                error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(formationConstraints, nil)

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("MultipleToGraphQL", formationConstraints).Return(formationConstraintsGql)

				return converter
			},
			ExpectedOutput: formationConstraintsGql,
			ExpectedError:  nil,
		},
		{
			Name: "Error when listing from service fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(nil, testErr)

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
		{
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("List", txtest.CtxWithDBMatcher()).Return(formationConstraints, nil)

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
		{
			Name:                         "Returns error when failing on the beginning of a transaction",
			TxFn:                         txGen.ThatFailsOnBegin,
			FormationConstraintService:   UnusedFormationConstraintService,
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationConstraintSvc := testCase.FormationConstraintService()
			formationConstraintConverter := testCase.FormationConstraintConverter()

			resolver := formationconstraint.NewResolver(transact, formationConstraintConverter, formationConstraintSvc)

			// WHEN
			result, err := resolver.FormationConstraints(ctx)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationConstraintSvc, formationConstraintConverter)
		})
	}
}

func TestResolver_FormationConstraintsByFormationType(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	formationConstraints := []*model.FormationConstraint{
		{Name: "test"},
		{Name: "test2"},
	}

	formationConstraintsGql := []*graphql.FormationConstraint{
		{Name: "test"},
		{Name: "test2"},
	}

	testCases := []struct {
		Name                         string
		TxFn                         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationConstraintConverter func() *automock.FormationConstraintConverter
		FormationConstraintService   func() *automock.FormationConstraintService
		ExpectedOutput               []*graphql.FormationConstraint
		ExpectedError                error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("ListByFormationTemplateID", txtest.CtxWithDBMatcher(), formationTemplateID).Return(formationConstraints, nil)

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("MultipleToGraphQL", formationConstraints).Return(formationConstraintsGql)

				return converter
			},
			ExpectedOutput: formationConstraintsGql,
			ExpectedError:  nil,
		},
		{
			Name: "Error when listing from service fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("ListByFormationTemplateID", txtest.CtxWithDBMatcher(), formationTemplateID).Return(nil, testErr)

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
		{
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("ListByFormationTemplateID", txtest.CtxWithDBMatcher(), formationTemplateID).Return(formationConstraints, nil)

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
		{
			Name:                         "Returns error when failing on the beginning of a transaction",
			TxFn:                         txGen.ThatFailsOnBegin,
			FormationConstraintService:   UnusedFormationConstraintService,
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationConstraintSvc := testCase.FormationConstraintService()
			formationConstraintConverter := testCase.FormationConstraintConverter()

			resolver := formationconstraint.NewResolver(transact, formationConstraintConverter, formationConstraintSvc)

			// WHEN
			result, err := resolver.FormationConstraintsByFormationType(ctx, formationTemplateID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationConstraintSvc, formationConstraintConverter)
		})
	}
}

func TestResolver_FormationConstraint(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                         string
		TxFn                         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationConstraintConverter func() *automock.FormationConstraintConverter
		FormationConstraintService   func() *automock.FormationConstraintService
		ExpectedOutput               *graphql.FormationConstraint
		ExpectedError                error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil)

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("ToGraphQL", formationConstraintModel).Return(gqlFormationConstraint, nil)

				return converter
			},
			ExpectedOutput: gqlFormationConstraint,
			ExpectedError:  nil,
		},
		{
			Name: "Error when getting from service fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr)

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
		{
			Name: "Returns nil when formation constraint not found",
			TxFn: txGen.ThatSucceeds,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, apperrors.NewNotFoundError(resource.FormationTemplate, testID))

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                nil,
		},
		{
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil)

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
		{
			Name:                         "Returns error when failing on the beginning of a transaction",
			TxFn:                         txGen.ThatFailsOnBegin,
			FormationConstraintService:   UnusedFormationConstraintService,
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationConstraintSvc := testCase.FormationConstraintService()
			formationConstraintConverter := testCase.FormationConstraintConverter()

			resolver := formationconstraint.NewResolver(transact, formationConstraintConverter, formationConstraintSvc)

			// WHEN
			result, err := resolver.FormationConstraint(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationConstraintSvc, formationConstraintConverter)
		})
	}
}

func TestResolver_CreateFormationConstraint(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                         string
		Input                        graphql.FormationConstraintInput
		TxFn                         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationConstraintConverter func() *automock.FormationConstraintConverter
		FormationConstraintService   func() *automock.FormationConstraintService
		ExpectedOutput               *graphql.FormationConstraint
		ExpectedError                error
	}{
		{
			Name:  "Success",
			TxFn:  txGen.ThatSucceeds,
			Input: formationConstraintInput,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), formationConstraintModelInput).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil)

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInput).Return(formationConstraintModelInput, nil)
				converter.On("ToGraphQL", formationConstraintModel).Return(gqlFormationConstraint, nil)

				return converter
			},
			ExpectedOutput: gqlFormationConstraint,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when creating call to service fails",
			Input: formationConstraintInput,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), formationConstraintModelInput).Return("", testErr)

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInput).Return(formationConstraintModelInput, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when get call to service fails",
			Input: formationConstraintInput,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), formationConstraintModelInput).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr)

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInput).Return(formationConstraintModelInput, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			TxFn:  txGen.ThatFailsOnCommit,
			Input: formationConstraintInput,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Create", txtest.CtxWithDBMatcher(), formationConstraintModelInput).Return(testID, nil)
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil)

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInput).Return(formationConstraintModelInput, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:                         "Returns error when input validation fails",
			Input:                        graphql.FormationConstraintInput{},
			TxFn:                         txGen.ThatDoesntExpectCommit,
			FormationConstraintService:   UnusedFormationConstraintService,
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                errors.New("cannot be blank"),
		},
		{
			Name:                         "Returns error when failing on the beginning of a transaction",
			Input:                        formationConstraintInput,
			TxFn:                         txGen.ThatFailsOnBegin,
			FormationConstraintService:   UnusedFormationConstraintService,
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationConstraintSvc := testCase.FormationConstraintService()
			formationConstraintConverter := testCase.FormationConstraintConverter()

			resolver := formationconstraint.NewResolver(transact, formationConstraintConverter, formationConstraintSvc)

			// WHEN
			result, err := resolver.CreateFormationConstraint(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationConstraintSvc, formationConstraintConverter)
		})
	}
}

func TestResolver_DeleteFormationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                         string
		TxFn                         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationConstraintConverter func() *automock.FormationConstraintConverter
		FormationConstraintService   func() *automock.FormationConstraintService
		ExpectedOutput               *graphql.FormationConstraint
		ExpectedError                error
	}{
		{
			Name: "Success",
			TxFn: txGen.ThatSucceeds,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil)

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("ToGraphQL", formationConstraintModel).Return(gqlFormationConstraint, nil)

				return converter
			},
			ExpectedOutput: gqlFormationConstraint,
			ExpectedError:  nil,
		},
		{
			Name: "Error when get call in service fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr)

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
		{
			Name: "Error when delete call in service fails",
			TxFn: txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(testErr)

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
		{
			Name: "Returns error when failing on the committing of a transaction",
			TxFn: txGen.ThatFailsOnCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), testID).Return(nil)

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
		{
			Name:                         "Returns error when failing on the beginning of a transaction",
			TxFn:                         txGen.ThatFailsOnBegin,
			FormationConstraintService:   UnusedFormationConstraintService,
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			formationConstraintSvc := testCase.FormationConstraintService()
			formationConstraintConverter := testCase.FormationConstraintConverter()

			resolver := formationconstraint.NewResolver(transact, formationConstraintConverter, formationConstraintSvc)

			// WHEN
			result, err := resolver.DeleteFormationConstraint(ctx, testID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationConstraintSvc, formationConstraintConverter)
		})
	}
}

func TestResolver_UpdateFormationConstraint(t *testing.T) {
	// GIVEN
	ctx := context.TODO()

	testErr := errors.New("test error")

	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                         string
		Input                        graphql.FormationConstraintUpdateInput
		TxFn                         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		FormationConstraintConverter func() *automock.FormationConstraintConverter
		FormationConstraintService   func() *automock.FormationConstraintService
		ExpectedOutput               *graphql.FormationConstraint
		ExpectedError                error
	}{
		{
			Name:  "Success when all updatable fields are provided",
			Input: formationConstraintUpdateInput,
			TxFn:  txGen.ThatSucceeds,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, formationConstraintModelInputUpdatedAllFields).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModelUpdatedAllFields, nil).Once()

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInputUpdatedAllFields).Return(formationConstraintModelInputUpdatedAllFields)
				converter.On("ToGraphQL", formationConstraintModelUpdatedAllFields).Return(gqlFormationConstraintUpdated)

				return converter
			},
			ExpectedOutput: gqlFormationConstraintUpdated,
			ExpectedError:  nil,
		},
		{
			Name:  "Success when only input template and description are updated",
			Input: formationConstraintUpdateInputWithTemplateAndDescription,
			TxFn:  txGen.ThatSucceeds,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, formationConstraintModelInputUpdatedTemplateAndDescription).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModelUpdatedTemplateAndDescription, nil).Once()

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInputUpdatedTemplateAndDescription).Return(formationConstraintModelInputUpdatedTemplateAndDescription)
				converter.On("ToGraphQL", formationConstraintModelUpdatedTemplateAndDescription).Return(gqlFormationConstraintUpdatedTemplateAndDescription)

				return converter
			},
			ExpectedOutput: gqlFormationConstraintUpdatedTemplateAndDescription,
			ExpectedError:  nil,
		},
		{
			Name:  "Success when only input template and priority are updated",
			Input: formationConstraintUpdateInputWithTemplateAndPriority,
			TxFn:  txGen.ThatSucceeds,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, formationConstraintModelInputUpdatedTemplateAndPriority).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModelUpdatedTemplateAndPriority, nil).Once()

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInputUpdatedTemplateAndPriority).Return(formationConstraintModelInputUpdatedTemplateAndPriority)
				converter.On("ToGraphQL", formationConstraintModelUpdatedTemplateAndPriority).Return(gqlFormationConstraintUpdatedTemplateAndPriority)

				return converter
			},
			ExpectedOutput: gqlFormationConstraintUpdatedTemplateAndPriority,
			ExpectedError:  nil,
		},
		{
			Name:  "Error when getting constraint",
			Input: formationConstraintUpdateInput,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr).Once()

				return svc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when updating constraint",
			Input: formationConstraintUpdateInput,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, formationConstraintModelInput).Return(testErr).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil).Once()

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInputUpdatedAllFields).Return(formationConstraintModelInput, nil)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Error when getting updated constraint",
			Input: formationConstraintUpdateInput,
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, formationConstraintModelInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(nil, testErr).Once()

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInputUpdatedAllFields).Return(formationConstraintModelInput)

				return converter
			},
			ExpectedError: testErr,
		},
		{
			Name:  "Returns error when failing on the committing of a transaction",
			Input: formationConstraintUpdateInput,
			TxFn:  txGen.ThatFailsOnCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Update", txtest.CtxWithDBMatcher(), testID, formationConstraintModelInput).Return(nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil).Once()
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModelUpdatedAllFields, nil).Once()

				return svc
			},
			FormationConstraintConverter: func() *automock.FormationConstraintConverter {
				converter := &automock.FormationConstraintConverter{}
				converter.On("FromInputGraphQL", &formationConstraintInputUpdatedAllFields).Return(formationConstraintModelInput)

				return converter
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:  "Returns error when input validation fails",
			Input: graphql.FormationConstraintUpdateInput{},
			TxFn:  txGen.ThatDoesntExpectCommit,
			FormationConstraintService: func() *automock.FormationConstraintService {
				svc := &automock.FormationConstraintService{}
				svc.On("Get", txtest.CtxWithDBMatcher(), testID).Return(formationConstraintModel, nil).Once()

				return svc
			},
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                errors.New("cannot be blank"),
		},
		{
			Name:                         "Returns error when failing on the beginning of a transaction",
			TxFn:                         txGen.ThatFailsOnBegin,
			Input:                        formationConstraintUpdateInput,
			FormationConstraintService:   UnusedFormationConstraintService,
			FormationConstraintConverter: UnusedFormationConstraintConverter,
			ExpectedOutput:               nil,
			ExpectedError:                testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()

			formationConstraintSvc := UnusedFormationConstraintService()
			if testCase.FormationConstraintService != nil {
				formationConstraintSvc = testCase.FormationConstraintService()
			}
			formationConstraintConverter := UnusedFormationConstraintConverter()
			if testCase.FormationConstraintConverter != nil {
				formationConstraintConverter = testCase.FormationConstraintConverter()
			}

			resolver := formationconstraint.NewResolver(transact, formationConstraintConverter, formationConstraintSvc)

			// WHEN
			result, err := resolver.UpdateFormationConstraint(ctx, testID, testCase.Input)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, persist, formationConstraintSvc, formationConstraintConverter)
		})
	}
}
