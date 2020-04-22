package scenarioassignment_test

import (
	"context"
	"fmt"
	"testing"

	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResolverCreateAutomaticScenarioAssignment(t *testing.T) {
	givenInput := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: scenarioName,
		Selector: &graphql.LabelSelectorInput{
			Key:   "key",
			Value: "value",
		},
	}
	expectedOutput := graphql.AutomaticScenarioAssignment{
		ScenarioName: scenarioName,
		Selector: &graphql.Label{
			Key:   "key",
			Value: "value",
		},
	}

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		tx, transact := txGen.ThatSucceeds()
		mockConverter := &automock.Converter{}
		mockConverter.On("FromInputGraphQL", givenInput).Return(fixModel()).Once()
		mockConverter.On("ToGraphQL", fixModel()).Return(expectedOutput).Once()
		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		mockSvc.On("Create", mock.Anything, fixModel()).Return(fixModel(), nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.CreateAutomaticScenarioAssignment(context.TODO(), givenInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedOutput, actual)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		sut := scenarioassignment.NewResolver(transact, nil, nil)

		// WHEN
		_, err := sut.CreateAutomaticScenarioAssignment(context.TODO(), graphql.AutomaticScenarioAssignmentSetInput{})

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on creating assignment by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()
		mockConverter := &automock.Converter{}
		mockConverter.On("FromInputGraphQL", mock.Anything).Return(fixModel()).Once()
		mockSvc := &automock.Service{}
		mockSvc.On("Create", mock.Anything, fixModel()).Return(model.AutomaticScenarioAssignment{}, fixError()).Once()
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		_, err := sut.CreateAutomaticScenarioAssignment(context.TODO(), givenInput)

		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while creating Assignment: %s", errMsg))
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromInputGraphQL", givenInput).Return(fixModel()).Once()
		mockSvc := &automock.Service{}
		mockSvc.On("Create", mock.Anything, fixModel()).Return(fixModel(), nil).Once()
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		_, err := sut.CreateAutomaticScenarioAssignment(context.TODO(), givenInput)

		// THEN
		assert.EqualError(t, err, "while committing transaction: some persistence error")
	})
}

func TestResolver_GetAutomaticScenarioAssignmentByScenario(t *testing.T) {
	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))
	expectedOutput := fixGQL()

	t.Run("happy path", func(t *testing.T) {
		tx, transact := txGen.ThatSucceeds()
		mockConverter := &automock.Converter{}
		mockConverter.On("ToGraphQL", fixModel()).Return(expectedOutput).Once()
		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(fixModel(), nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedOutput, actual)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		sut := scenarioassignment.NewResolver(transact, nil, nil)

		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on receiving assignment by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()
		mockSvc := &automock.Service{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(model.AutomaticScenarioAssignment{}, fixError()).Once()
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
		sut := scenarioassignment.NewResolver(transact, mockSvc, nil)

		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)

		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while getting Assignment: %s", errMsg))
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()
		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(fixModel(), nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil)

		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)

		// THEN
		assert.EqualError(t, err, "while committing transaction: some persistence error")
	})
}

func TestResolver_AutomaticScenarioAssignmentsForSelector(t *testing.T) {
	givenInput := graphql.LabelSelectorInput{
		Key:   "key",
		Value: "value",
	}

	expectedModels := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioName,
			Selector: model.LabelSelector{
				Key:   "key",
				Value: "value",
			},
		},
		{
			ScenarioName: "scenario-B",
			Selector: model.LabelSelector{
				Key:   "key",
				Value: "value",
			},
		},
	}

	expectedOutput := []*graphql.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioName,
			Selector: &graphql.Label{
				Key:   "key",
				Value: "value",
			},
		},
		{
			ScenarioName: "scenario-B",
			Selector: &graphql.Label{
				Key:   "key",
				Value: "value",
			},
		},
	}

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		tx, transact := txGen.ThatSucceeds()

		mockConverter := &automock.Converter{}
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector()).Once()
		mockConverter.On("MultipleToGraphQL", expectedModels).Return(expectedOutput).Once()

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		mockSvc.On("ListForSelector", mock.Anything, fixLabelSelector()).Return(expectedModels, nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedOutput, actual)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		sut := scenarioassignment.NewResolver(transact, nil, nil)

		// WHEN
		_, err := sut.AutomaticScenarioAssignmentsForSelector(context.TODO(), graphql.LabelSelectorInput{})

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on getting assignments by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.Converter{}
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector()).Once()

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		mockSvc.On("ListForSelector", mock.Anything, fixLabelSelector()).Return(nil, fixError()).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.Nil(t, actual)
		require.EqualError(t, err, "while getting the assignments: some error")

	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()

		mockConverter := &automock.Converter{}
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector()).Once()
		mockConverter.On("MultipleToGraphQL", expectedModels).Return(expectedOutput).Once()

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		mockSvc.On("ListForSelector", mock.Anything, fixLabelSelector()).Return(expectedModels, nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.EqualError(t, err, "while committing transaction: some persistence error")
		require.Nil(t, actual)
	})
}

func TestResolver_AutomaticScenarioAssignments(t *testing.T) {
	testErr := errors.New("test error")

	mod1 := fixModelWithScenarioName("foo")
	mod2 := fixModelWithScenarioName("bar")
	modItems := []*model.AutomaticScenarioAssignment{
		&mod1, &mod2,
	}
	modelPage := fixModelPageWithItems(modItems)

	gql1 := fixGQLWithScenarioName("foo")
	gql2 := fixGQLWithScenarioName("bar")
	gqlItems := []*graphql.AutomaticScenarioAssignment{
		&gql1, &gql2,
	}
	gqlPage := fixGQLPageWithItems(gqlItems)

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.Service
		ConverterFn     func() *automock.Converter
		ExpectedResult  *graphql.AutomaticScenarioAssignmentPage
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(&modelPage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				conv.On("MultipleToGraphQL", modItems).Return(gqlItems).Once()
				return conv
			},
			ExpectedResult: &gqlPage,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when Assignments listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.Service {
				svc := &automock.Service{}
				svc.On("List", txtest.CtxWithDBMatcher(), first, after).Return(&modelPage, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.Converter {
				conv := &automock.Converter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := scenarioassignment.NewResolver(transact, svc, converter)

			// WHEN
			result, err := resolver.AutomaticScenarioAssignments(context.TODO(), &first, &gqlAfter)

			// THEN
			assert.Equal(t, testCase.ExpectedResult, result)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testErr.Error())
			}

			mock.AssertExpectationsForObjects(t, persist, transact, svc, converter)
		})
	}
}

func TestResolver_DeleteAutomaticScenarioAssignmentsForSelector(t *testing.T) {
	givenInput := graphql.LabelSelectorInput{
		Key:   "key",
		Value: "value",
	}

	scenarioNameA := "scenario-A"
	scenarioNameB := "scenario-B"
	expectedModels := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioNameA,
			Selector: model.LabelSelector{
				Key:   "key",
				Value: "value",
			},
		},
		{
			ScenarioName: scenarioNameB,
			Selector: model.LabelSelector{
				Key:   "key",
				Value: "value",
			},
		},
	}

	expectedOutput := []*graphql.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioNameA,
			Selector: &graphql.Label{
				Key:   "key",
				Value: "value",
			},
		},
		{
			ScenarioName: scenarioNameB,
			Selector: &graphql.Label{
				Key:   "key",
				Value: "value",
			},
		},
	}

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatSucceeds()

		mockConverter := &automock.Converter{}
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector()).Once()
		mockConverter.On("MultipleToGraphQL", expectedModels).Return(expectedOutput).Once()

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		mockSvc.On("ListForSelector", txtest.CtxWithDBMatcher(), fixLabelSelector()).Return(expectedModels, nil).Once()
		mockSvc.On("DeleteManyForSameSelector", txtest.CtxWithDBMatcher(), expectedModels).Return(nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedOutput, actual)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		sut := scenarioassignment.NewResolver(transact, nil, nil)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(context.TODO(), graphql.LabelSelectorInput{})

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on getting assignments by service", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.Converter{}
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector()).Once()

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		mockSvc.On("ListForSelector", txtest.CtxWithDBMatcher(), fixLabelSelector()).Return(nil, fixError()).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.Nil(t, actual)
		require.EqualError(t, err, fmt.Sprintf("while getting the Assignments for selector [{key value}]: %s", errMsg))
	})

	t.Run("error on deleting assignments by service", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.Converter{}
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector()).Once()

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		mockSvc.On("ListForSelector", txtest.CtxWithDBMatcher(), fixLabelSelector()).Return(expectedModels, nil).Once()
		mockSvc.On("DeleteManyForSameSelector", txtest.CtxWithDBMatcher(), expectedModels).Return(fixError()).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.Nil(t, actual)
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignments for selector [{key value}]: %s", errMsg))
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatFailsOnCommit()

		mockConverter := &automock.Converter{}
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector()).Once()

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
		mockSvc.On("ListForSelector", txtest.CtxWithDBMatcher(), fixLabelSelector()).Return(expectedModels, nil).Once()
		mockSvc.On("DeleteManyForSameSelector", txtest.CtxWithDBMatcher(), expectedModels).Return(nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentsForSelector(fixCtxWithTenant(), givenInput)

		// THEN
		require.EqualError(t, err, "while committing transaction: some persistence error")
		require.Nil(t, actual)
	})
}

func TestResolver_DeleteAutomaticScenarioAssignmentForScenario(t *testing.T) {
	expectedModel := fixModel()
	expectedOutput := fixGQL()

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatSucceeds()

		mockConverter := &automock.Converter{}
		mockConverter.On("ToGraphQL", expectedModel).Return(expectedOutput).Once()

		mockSvc := &automock.Service{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(expectedModel, nil).Once()
		mockSvc.On("Delete", txtest.CtxWithDBMatcher(), expectedModel).Return(nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, mockConverter)

		// WHEN
		actual, err := sut.DeleteAutomaticScenarioAssignmentForScenario(fixCtxWithTenant(), scenarioName)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedOutput, actual)
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc, mockConverter)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatFailsOnBegin()
		sut := scenarioassignment.NewResolver(transact, nil, nil)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentForScenario(context.TODO(), scenarioName)

		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
		mock.AssertExpectationsForObjects(t, tx, transact)
	})

	t.Run("error on getting assignment by service", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockSvc := &automock.Service{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(model.AutomaticScenarioAssignment{}, fixError()).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentForScenario(fixCtxWithTenant(), scenarioName)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while getting the Assignment for scenario [name=%s]: %s", scenarioName, errMsg))
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
	})

	t.Run("error on deleting assignments by service", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockSvc := &automock.Service{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(expectedModel, nil).Once()
		mockSvc.On("Delete", txtest.CtxWithDBMatcher(), expectedModel).Return(fixError()).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentForScenario(fixCtxWithTenant(), scenarioName)

		// THEN
		require.EqualError(t, err, fmt.Sprintf("while deleting the Assignment for scenario [name=%s]: %s", scenarioName, errMsg))
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		// GIVEN
		tx, transact := txGen.ThatFailsOnCommit()

		mockSvc := &automock.Service{}
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(expectedModel, nil).Once()
		mockSvc.On("Delete", txtest.CtxWithDBMatcher(), expectedModel).Return(nil).Once()

		sut := scenarioassignment.NewResolver(transact, mockSvc, nil)

		// WHEN
		_, err := sut.DeleteAutomaticScenarioAssignmentForScenario(fixCtxWithTenant(), scenarioName)

		// THEN
		require.EqualError(t, err, "while committing transaction: some persistence error")
		mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
	})
}
