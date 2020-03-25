package scenarioassignment_test

import (
	"context"
	"fmt"
	"testing"

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

func TestResolverSetAutomaticScenarioAssignment(t *testing.T) {
	givenInput := graphql.AutomaticScenarioAssignmentSetInput{
		ScenarioName: "scenario-A",
		Selector: &graphql.LabelSelectorInput{
			Key:   "key",
			Value: "value",
		},
	}
	expectedOutput := graphql.AutomaticScenarioAssignment{
		ScenarioName: "scenario-A",
		Selector: &graphql.Label{
			Key:   "key",
			Value: "value",
		},
	}

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		tx, transact := txGen.ThatSucceeds()
		mockConverter := &automock.Converter{}
		mockConverter.On("FromInputGraphQL", givenInput).Return(fixModel())
		mockConverter.On("ToGraphQL", fixModel()).Return(expectedOutput)
		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockConverter, mockSvc)
		mockSvc.On("Create", mock.Anything, fixModel()).Return(fixModel(), nil)

		sut := scenarioassignment.NewResolver(transact, mockConverter, mockSvc)
		// WHEN
		actual, err := sut.SetAutomaticScenarioAssignment(context.TODO(), givenInput)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedOutput, actual)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		sut := scenarioassignment.NewResolver(transact, nil, nil)
		// WHEN
		_, err := sut.SetAutomaticScenarioAssignment(context.TODO(), graphql.AutomaticScenarioAssignmentSetInput{})
		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on creating assignment by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()
		mockConverter := &automock.Converter{}
		mockConverter.On("FromInputGraphQL", mock.Anything).Return(fixModel())
		mockSvc := &automock.Service{}
		mockSvc.On("Create", mock.Anything, fixModel()).Return(model.AutomaticScenarioAssignment{}, fixError())
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockConverter, mockSvc)
		sut := scenarioassignment.NewResolver(transact, mockConverter, mockSvc)
		// WHEN
		_, err := sut.SetAutomaticScenarioAssignment(context.TODO(), givenInput)
		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while creating Assignment: %s", errMsg))
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromInputGraphQL", givenInput).Return(fixModel())
		mockSvc := &automock.Service{}
		mockSvc.On("Create", mock.Anything, fixModel()).Return(fixModel(), nil)
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockConverter, mockSvc)
		sut := scenarioassignment.NewResolver(transact, mockConverter, mockSvc)
		// WHEN
		_, err := sut.SetAutomaticScenarioAssignment(context.TODO(), givenInput)
		// THEN
		assert.EqualError(t, err, "while committing transaction: some persistence error")
	})
}

func TestResolver_GetAutomaticScenarioAssignmentByScenario(t *testing.T) {
	expectedOutput := graphql.AutomaticScenarioAssignment{
		ScenarioName: "scenario-A",
		Selector: &graphql.Label{
			Key:   "key",
			Value: "value",
		},
	}

	txGen := txtest.NewTransactionContextGenerator(errors.New("some persistence error"))

	t.Run("happy path", func(t *testing.T) {
		tx, transact := txGen.ThatSucceeds()
		mockConverter := &automock.Converter{}
		mockConverter.On("ToGraphQL", fixModel()).Return(expectedOutput)
		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockConverter, mockSvc)
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(fixModel(), nil)

		sut := scenarioassignment.NewResolver(transact, mockConverter, mockSvc)
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
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(model.AutomaticScenarioAssignment{}, fixError())
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
		sut := scenarioassignment.NewResolver(transact, nil, mockSvc)
		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)
		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while getting Assignment: %s", errMsg))
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()
		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
		mockSvc.On("GetForScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(fixModel(), nil)

		sut := scenarioassignment.NewResolver(transact, nil, mockSvc)
		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentForScenarioName(context.TODO(), scenarioName)
		// THEN
		assert.EqualError(t, err, "while committing transaction: some persistence error")
	})
}

func TestResolver_AutomaticScenarioAssignmentForSelector(t *testing.T) {

	givenInput := graphql.LabelSelectorInput{
		Key:   "key",
		Value: "value",
	}

	expectedModels := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: "scenario-A",
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
			ScenarioName: "scenario-A",
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
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector())
		mockConverter.On("ToGraphQL", *expectedModels[0]).Return(*expectedOutput[0])
		mockConverter.On("ToGraphQL", *expectedModels[1]).Return(*expectedOutput[1])

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockConverter, mockSvc)
		mockSvc.On("GetForSelector", mock.Anything, fixLabelSelector(), DefaultTenant).Return(expectedModels, nil)

		sut := scenarioassignment.NewResolver(transact, mockConverter, mockSvc)
		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentForSelector(fixCtxWithTenant(), givenInput)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedOutput, actual)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		sut := scenarioassignment.NewResolver(transact, nil, nil)
		// WHEN
		_, err := sut.SetAutomaticScenarioAssignment(context.TODO(), graphql.AutomaticScenarioAssignmentSetInput{})
		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on missing tenant in context", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		sut := scenarioassignment.NewResolver(transact, nil, nil)
		// WHEN
		_, err := sut.SetAutomaticScenarioAssignment(context.TODO(), graphql.AutomaticScenarioAssignmentSetInput{})
		// THEN
		assert.EqualError(t, err, "cannot read tenant from context")
	})

	t.Run("error on getting assignments by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()

		mockConverter := &automock.Converter{}
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector())

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockConverter, mockSvc)
		mockSvc.On("GetForSelector", mock.Anything, fixLabelSelector(), DefaultTenant).Return(nil, fixError())

		sut := scenarioassignment.NewResolver(transact, mockConverter, mockSvc)
		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentForSelector(fixCtxWithTenant(), givenInput)
		// THEN
		require.Nil(t, actual)
		require.EqualError(t, err, "while getting the assignments: some error")

	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()

		mockConverter := &automock.Converter{}
		mockConverter.On("LabelSelectorFromInput", givenInput).Return(fixLabelSelector())
		mockConverter.On("ToGraphQL", *expectedModels[0]).Return(*expectedOutput[0])
		mockConverter.On("ToGraphQL", *expectedModels[1]).Return(*expectedOutput[1])

		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockConverter, mockSvc)
		mockSvc.On("GetForSelector", mock.Anything, fixLabelSelector(), DefaultTenant).Return(expectedModels, nil)

		sut := scenarioassignment.NewResolver(transact, mockConverter, mockSvc)
		// WHEN
		actual, err := sut.AutomaticScenarioAssignmentForSelector(fixCtxWithTenant(), givenInput)
		// THEN
		require.EqualError(t, err, "while committing transaction: some persistence error")
		require.Nil(t, actual)
	})
}
