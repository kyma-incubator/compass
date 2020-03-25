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
		mockConverter.On("FromInputGraphQL", givenInput).Return(fixModel(), nil)
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

	t.Run("error on converting input to model", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()
		mockConverter := &automock.Converter{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockConverter)
		mockConverter.On("FromInputGraphQL", mock.Anything).Return(model.AutomaticScenarioAssignment{}, errors.New("conversion error"))
		sut := scenarioassignment.NewResolver(transact, mockConverter, nil)
		// WHEN
		_, err := sut.SetAutomaticScenarioAssignment(context.TODO(), givenInput)
		// THEN
		assert.EqualError(t, err, "while converting to model: conversion error")
	})

	t.Run("error on creating assignment by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()
		mockConverter := &automock.Converter{}
		mockConverter.On("FromInputGraphQL", mock.Anything).Return(fixModel(), nil)
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
		mockConverter.On("FromInputGraphQL", givenInput).Return(fixModel(), nil)
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
		mockSvc.On("GetByScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(fixModel(), nil)

		sut := scenarioassignment.NewResolver(transact, mockConverter, mockSvc)
		// WHEN
		actual, err := sut.GetAutomaticScenarioAssignmentByScenarioName(context.TODO(), scenarioName)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &expectedOutput, actual)
	})

	t.Run("error on starting transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnBegin()
		defer mock.AssertExpectationsForObjects(t, tx, transact)
		sut := scenarioassignment.NewResolver(transact, nil, nil)
		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentByScenarioName(context.TODO(), scenarioName)
		// THEN
		assert.EqualError(t, err, "while beginning transaction: some persistence error")
	})

	t.Run("error on receiving assignment by service", func(t *testing.T) {
		tx, transact := txGen.ThatDoesntExpectCommit()
		mockSvc := &automock.Service{}
		mockSvc.On("GetByScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(model.AutomaticScenarioAssignment{}, fixError())
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
		sut := scenarioassignment.NewResolver(transact, nil, mockSvc)
		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentByScenarioName(context.TODO(), scenarioName)
		// THEN
		assert.EqualError(t, err, fmt.Sprintf("while getting Assignment: %s", errMsg))
	})

	t.Run("error on committing transaction", func(t *testing.T) {
		tx, transact := txGen.ThatFailsOnCommit()
		mockSvc := &automock.Service{}
		defer mock.AssertExpectationsForObjects(t, tx, transact, mockSvc)
		mockSvc.On("GetByScenarioName", txtest.CtxWithDBMatcher(), scenarioName).Return(fixModel(), nil)

		sut := scenarioassignment.NewResolver(transact, nil, mockSvc)
		// WHEN
		_, err := sut.GetAutomaticScenarioAssignmentByScenarioName(context.TODO(), scenarioName)
		// THEN
		assert.EqualError(t, err, "while committing transaction: some persistence error")
	})
}
