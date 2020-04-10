package scenarioassignment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEngine_EnsureScenarioAssigned(t *testing.T) {
	selectorKey := "KEY"
	selectorValue := "VALUE"
	selectorScenario := "SELECTOR_SCENARIO"
	in := fixAutomaticScenarioAssigment(selectorScenario, selectorKey, selectorValue)
	testErr := errors.New("test err")

	t.Run("Success", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    scenarios,
			ObjectID: "runtime_id",
		}
		expectedScenarioLabel := scenarioLabel
		expectedScenarioLabel.Value = append(scenarios, selectorScenario)

		ctx := context.TODO()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return([]model.Label{scenarioLabel}, nil).Once()
		labelRepo.On("Upsert", ctx, mock.MatchedBy(matchExpectedScenarios(t, &expectedScenarioLabel))).
			Return(nil).Once()

		eng := scenarioassignment.NewEngine(labelRepo)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo)
		labelRepo.AssertExpectations(t)
	})

	t.Run("Failed when Label Upsert failed ", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    scenarios,
			ObjectID: "runtime_id",
		}

		expectedScenarioLabel := scenarioLabel
		expectedScenarioLabel.Value = append(scenarios, selectorScenario)

		ctx := context.TODO()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return([]model.Label{scenarioLabel}, nil).Once()
		labelRepo.On("Upsert", ctx, mock.MatchedBy(matchExpectedScenarios(t, &expectedScenarioLabel))).
			Return(testErr)

		eng := scenarioassignment.NewEngine(labelRepo)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		labelRepo.AssertExpectations(t)
	})

	t.Run("Failed when GetRuntimeScenariosWhereLabelsMatchSelector returns error", func(t *testing.T) {
		ctx := context.TODO()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return([]model.Label{}, testErr).Once()

		eng := scenarioassignment.NewEngine(labelRepo)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		labelRepo.AssertExpectations(t)
	})
}

func matchExpectedScenarios(t *testing.T, expected *model.Label) func(label *model.Label) bool {
	return func(actual *model.Label) bool {
		actualArray, ok := actual.Value.([]interface{})
		require.True(t, ok)

		expectedArray, ok := expected.Value.([]interface{})
		require.True(t, ok)
		return assert.ElementsMatch(t, expectedArray, actualArray)
	}
}

func fixAutomaticScenarioAssigment(selectorScenario, selectorKey, selectorValue string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName: selectorScenario,
		Tenant:       tenantID,
		Selector: model.LabelSelector{
			Key:   selectorKey,
			Value: selectorValue,
		},
	}
}
