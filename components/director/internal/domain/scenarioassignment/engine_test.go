package scenarioassignment_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
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

		eng := scenarioassignment.NewEngine(labelRepo, nil)

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

		eng := scenarioassignment.NewEngine(labelRepo, nil)

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

		eng := scenarioassignment.NewEngine(labelRepo, nil)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		labelRepo.AssertExpectations(t)
	})
}

func TestEngine_RemoveAssignedScenario(t *testing.T) {
	selectorKey := "KEY"
	selectorValue := "VALUE"
	selectorScenario := "SELECTOR_SCENARIO"
	in := fixAutomaticScenarioAssigment(selectorScenario, selectorKey, selectorValue)
	testErr := errors.New("test err")

	t.Run("Success", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    append(scenarios, selectorScenario),
			ObjectID: "runtime_id",
		}
		expectedScenarioLabel := scenarioLabel
		expectedScenarioLabel.Value = scenarios

		ctx := context.TODO()

		labels := []model.Label{scenarioLabel}
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return(labels, nil).Once()
		labelRepo.On("Upsert", ctx, mock.MatchedBy(matchExpectedScenarios(t, &expectedScenarioLabel))).
			Return(nil).Once()

		eng := scenarioassignment.NewEngine(labelRepo, nil)

		//WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo)
		labelRepo.AssertExpectations(t)
	})

	t.Run("Failed when Label Upsert failed ", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    append(scenarios, selectorScenario),
			ObjectID: "runtime_id",
		}

		expectedScenarioLabel := scenarioLabel
		expectedScenarioLabel.Value = scenarios

		ctx := context.TODO()

		labels := []model.Label{scenarioLabel, {Key: selectorKey}}
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return(labels, nil).Once()
		labelRepo.On("Upsert", ctx, mock.MatchedBy(matchExpectedScenarios(t, &expectedScenarioLabel))).
			Return(testErr)

		eng := scenarioassignment.NewEngine(labelRepo, nil)

		//WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		labelRepo.AssertExpectations(t)
	})

	t.Run("Failed when GetRuntimeScenariosWhereRuntimesLabelsMatchSelector returns error", func(t *testing.T) {
		ctx := context.TODO()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return([]model.Label{}, testErr).Once()

		eng := scenarioassignment.NewEngine(labelRepo, nil)

		//WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		labelRepo.AssertExpectations(t)
	})
}

func TestEngine_RemoveAssignedScenarios(t *testing.T) {
	selectorKey, selectorValue := "KEY", "VALUE"
	expctedScenario1 := []interface{}{"SCENARIO2"}
	in := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: "SCENARIO1",
			Tenant:       tenantID,
			Selector: model.LabelSelector{
				Key:   selectorKey,
				Value: selectorValue,
			}}}

	labels := []model.Label{{
		Value:    []interface{}{"SCENARIO1", "SCENARIO2"},
		Key:      model.ScenariosKey,
		ObjectID: "651038e0-e4b6-4036-a32f-f6e9846003f4",
	}}

	t.Run("Success", func(t *testing.T) {
		//GIVEN
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return(labels, nil).Once()
		labelRepo.On("Upsert", ctx, mock.MatchedBy(matchExpectedScenarios(t, &model.Label{Value: expctedScenario1}))).Return(nil).Once()
		eng := scenarioassignment.NewEngine(labelRepo, nil)
		//WHEN
		err := eng.RemoveAssignedScenarios(ctx, in)

		//THEN
		require.NoError(t, err)
		labelRepo.AssertExpectations(t)
	})

	t.Run("Error, while removing scenario", func(t *testing.T) {
		//GIVEN
		testErr := errors.New("test error")
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return(labels, testErr).Once()
		eng := scenarioassignment.NewEngine(labelRepo, nil)
		//WHEN
		err := eng.RemoveAssignedScenarios(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		labelRepo.AssertExpectations(t)
	})
}

func matchExpectedScenarios(t *testing.T, expected *model.Label) func(label *model.Label) bool {
	return func(actual *model.Label) bool {
		actualArray, ok := actual.Value.([]string)
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

func TestEngine_GetScenariosForSelectorLabels_Success(t *testing.T) {
	// given
	key := "foo"
	value := "bar"

	selectorLabels := map[string]string{
		key: value,
	}

	selector := model.LabelSelector{
		Key:   key,
		Value: value,
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioName,
			Tenant:       tenantID,
			Selector: model.LabelSelector{
				Key:   key,
				Value: value,
			},
		},
	}

	expectedScenarios := []string{scenarioName}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListForSelector", fixCtxWithTenant(), selector, tenantID).Return(assignments, nil)
	defer mock.AssertExpectationsForObjects(t, mockRepo)

	engineSvc := scenarioassignment.NewEngine(nil, mockRepo)

	// when
	actualScenarios, err := engineSvc.GetScenariosForSelectorLabels(fixCtxWithTenant(), selectorLabels)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedScenarios, actualScenarios)
}

func TestEngine_GetScenariosForSelectorLabels_ShouldFailOnGettingForSelector(t *testing.T) {
	// given
	testErr := errors.New("test error")
	key := "foo"
	value := "bar"

	selectorLabels := map[string]string{
		key: value,
	}

	selector := model.LabelSelector{
		Key:   key,
		Value: value,
	}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListForSelector", fixCtxWithTenant(), selector, tenantID).Return(nil, testErr)
	defer mock.AssertExpectationsForObjects(t, mockRepo)

	engineSvc := scenarioassignment.NewEngine(nil, mockRepo)

	// when
	_, err := engineSvc.GetScenariosForSelectorLabels(fixCtxWithTenant(), selectorLabels)

	// then
	require.Error(t, err)
	assert.EqualError(t, fmt.Errorf("while getting Automatic Scenario Assignments for selector [key: %s, val: %s]: %s", key, value, testErr.Error()), err.Error())
}

func TestEngine_GetScenariosForSelectorLabels_ShouldFailOnLoadingTenant(t *testing.T) {
	// given
	svc := scenarioassignment.NewEngine(nil, nil)
	// when
	_, err := svc.GetScenariosForSelectorLabels(context.TODO(), nil)
	// then
	assert.EqualError(t, err, "cannot read tenant from context")
}

func TestEngine_MergeScenariosFromInputLabelsAndAssignments_Success(t *testing.T) {
	// given
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	selector := model.LabelSelector{
		Key:   labelKey,
		Value: labelValue,
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioName,
			Tenant:       tenantID,
			Selector: model.LabelSelector{
				Key:   labelKey,
				Value: labelValue,
			},
		},
	}

	expectedScenarios := []interface{}{scenarioName}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListForSelector", fixCtxWithTenant(), selector, tenantID).Return(assignments, nil)
	engineSvc := scenarioassignment.NewEngine(nil, mockRepo)

	// when
	actualScenarios, err := engineSvc.MergeScenariosFromInputLabelsAndAssignments(fixCtxWithTenant(), inputLabels)

	// then

	require.NoError(t, err)
	assert.ElementsMatch(t, expectedScenarios, actualScenarios)

	mockRepo.AssertExpectations(t)
}

func TestEngine_MergeScenariosFromInputLabelsAndAssignments_SuccessIfScenariosLabelIsInInput(t *testing.T) {
	// given
	labelKey := "key"
	labelValue := "val"

	scenario := "SCENARIO"
	inputLabels := map[string]interface{}{
		labelKey:           labelValue,
		model.ScenariosKey: []interface{}{scenario},
	}

	selector := model.LabelSelector{
		Key:   labelKey,
		Value: labelValue,
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioName,
			Tenant:       tenantID,
			Selector: model.LabelSelector{
				Key:   labelKey,
				Value: labelValue,
			},
		},
	}

	expectedScenarios := []interface{}{scenarioName, scenario}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListForSelector", fixCtxWithTenant(), selector, tenantID).Return(assignments, nil)
	engineSvc := scenarioassignment.NewEngine(nil, mockRepo)

	// when
	actualScenarios, err := engineSvc.MergeScenariosFromInputLabelsAndAssignments(fixCtxWithTenant(), inputLabels)

	// then
	require.NoError(t, err)
	assert.ElementsMatch(t, expectedScenarios, actualScenarios)

	mockRepo.AssertExpectations(t)
}

func TestEngine_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfListForSelectorFailed(t *testing.T) {
	// given
	testErr := errors.New("testErr")
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	selector := model.LabelSelector{
		Key:   labelKey,
		Value: labelValue,
	}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListForSelector", fixCtxWithTenant(), selector, tenantID).Return(nil, testErr)
	engineSvc := scenarioassignment.NewEngine(nil, mockRepo)

	// when
	_, err := engineSvc.MergeScenariosFromInputLabelsAndAssignments(fixCtxWithTenant(), inputLabels)

	// then
	require.Error(t, err)

	mockRepo.AssertExpectations(t)
}

func TestEngine_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfScenariosFromInputWereNotInterfaceSlice(t *testing.T) {
	// given
	labelKey := "key"
	labelValue := "val"

	scenario := "SCENARIO"
	inputLabels := map[string]interface{}{
		labelKey:           labelValue,
		model.ScenariosKey: []string{scenario},
	}

	selector := model.LabelSelector{
		Key:   labelKey,
		Value: labelValue,
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: scenarioName,
			Tenant:       tenantID,
			Selector: model.LabelSelector{
				Key:   labelKey,
				Value: labelValue,
			},
		},
	}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListForSelector", fixCtxWithTenant(), selector, tenantID).Return(assignments, nil)
	engineSvc := scenarioassignment.NewEngine(nil, mockRepo)

	// when
	_, err := engineSvc.MergeScenariosFromInputLabelsAndAssignments(fixCtxWithTenant(), inputLabels)

	// then
	require.Error(t, err)

	mockRepo.AssertExpectations(t)
}

func TestEngine_MergeScenarios_Success(t *testing.T) {
	// given
	oldScenariosLabel := []interface{}{"DEFAULT", "CUSTOM"}
	previousScenariosFromAssignments := []interface{}{"DEFAULT"}
	newScenariosFromAssignments := []interface{}{"CUSTOM"}

	expectedScenarios := []interface{}{"CUSTOM"}

	engineSvc := scenarioassignment.NewEngine(nil, nil)

	// when
	actualScenarios := engineSvc.MergeScenarios(oldScenariosLabel, previousScenariosFromAssignments, newScenariosFromAssignments)

	// then
	assert.Equal(t, expectedScenarios, actualScenarios)
}
