package scenarioassignment_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	labelpkg "github.com/kyma-incubator/compass/components/director/internal/domain/label"
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
	otherScenario := "OTHER"
	basicScenario := "SCENARIO"
	scenarios := []interface{}{otherScenario, basicScenario}
	stringScenarios := []string{otherScenario, basicScenario}

	rtmIDWithScenario := "rtm1_scenario"
	rtmIDWithoutScenario := "rtm1_no_scenario"

	expectedScenarios := map[string][]string{
		rtmIDWithScenario:    stringScenarios,
		rtmIDWithoutScenario: {},
	}
	runtimesIDs := []string{rtmIDWithoutScenario, rtmIDWithScenario}
	scenarioLabel := model.Label{
		Key:        model.ScenariosKey,
		Value:      scenarios,
		ObjectID:   rtmIDWithScenario,
		ObjectType: model.RuntimeLabelableObject,
	}

	t.Run("Success", func(t *testing.T) {
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimesIDsByStringLabel", ctx, tenantID, selectorKey, selectorValue).
			Return(runtimesIDs, nil)

		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, runtimesIDs).
			Return([]model.Label{scenarioLabel}, nil)

		upsertSvc := &automock.LabelUpsertService{}

		matchExpectedScenariosAddFn := mock.MatchedBy(matchAddNewScenarioFn([][]string{stringScenarios, {}}, in.ScenarioName))
		upsertSvc.On("UpsertScenarios", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(expectedScenarios)), []string{in.ScenarioName}, matchExpectedScenariosAddFn).Return(nil).Once()

		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		bundleInstanceAuthSvc.On("AssociateBundleInstanceAuthForNewRuntimeScenarios", ctx, stringScenarios, []string{in.ScenarioName}, scenarioLabel.ObjectID).Return(nil).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, bundleInstanceAuthSvc)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, bundleInstanceAuthSvc)
	})

	t.Run("Failed when upsert runtime scenario labels returns error ", func(t *testing.T) {
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimesIDsByStringLabel", ctx, tenantID, selectorKey, selectorValue).
			Return(runtimesIDs, nil).Once()
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, runtimesIDs).
			Return([]model.Label{scenarioLabel}, nil)

		upsertSvc := &automock.LabelUpsertService{}
		matchExpectedScenariosAddFn := mock.MatchedBy(matchAddNewScenarioFn([][]string{stringScenarios, {}}, in.ScenarioName))
		upsertSvc.On("UpsertScenarios", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(expectedScenarios)), []string{in.ScenarioName}, matchExpectedScenariosAddFn).Return(testErr).Once()

		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		bundleInstanceAuthSvc.On("AssociateBundleInstanceAuthForNewRuntimeScenarios", ctx, stringScenarios, []string{in.ScenarioName}, scenarioLabel.ObjectID).Return(nil).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, bundleInstanceAuthSvc)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, bundleInstanceAuthSvc)
	})

	t.Run("Failed when associating existing bundle instance auths with scenarios returns error ", func(t *testing.T) {
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimesIDsByStringLabel", ctx, tenantID, selectorKey, selectorValue).
			Return(runtimesIDs, nil).Once()
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, runtimesIDs).
			Return([]model.Label{scenarioLabel}, nil)

		upsertSvc := &automock.LabelUpsertService{}
		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		bundleInstanceAuthSvc.On("AssociateBundleInstanceAuthForNewRuntimeScenarios", ctx, stringScenarios, []string{in.ScenarioName}, scenarioLabel.ObjectID).Return(testErr).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, bundleInstanceAuthSvc)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, bundleInstanceAuthSvc)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes returns error", func(t *testing.T) {
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimesIDsByStringLabel", ctx, tenantID, selectorKey, selectorValue).
			Return(runtimesIDs, nil).Once()
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, runtimesIDs).Return(nil, testErr)

		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, bundleInstanceAuthSvc)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, bundleInstanceAuthSvc)

	})

	t.Run("Failed when GetRuntimesIDsByStringLabel returns error", func(t *testing.T) {
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimesIDsByStringLabel", ctx, tenantID, selectorKey, selectorValue).
			Return(runtimesIDs, testErr).Once()

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, nil)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo)
	})

	t.Run("Success, no runtimes found", func(t *testing.T) {
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimesIDsByStringLabel", ctx, tenantID, selectorKey, selectorValue).
			Return([]string{}, nil).Once()

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, nil)

		//WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo)
	})
}

func TestEngine_RemoveAssignedScenario(t *testing.T) {
	selectorKey := "KEY"
	selectorValue := "VALUE"
	selectorScenario := "SELECTOR_SCENARIO"
	rtmID := "8c4de4d8-dcfa-47a9-95c9-3c8b1f5b907c"
	in := fixAutomaticScenarioAssigment(selectorScenario, selectorKey, selectorValue)
	testErr := errors.New("test err")

	t.Run("Success", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    append(scenarios, selectorScenario),
			ObjectID: rtmID,
		}

		existingScenariosSlice := []string{"OTHER", "SCENARIO", selectorScenario}
		existingScenarios := map[string][]string{
			rtmID: existingScenariosSlice,
		}

		ctx := context.TODO()

		labels := []model.Label{scenarioLabel}
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return(labels, nil).Once()

		upsertSvc := &automock.LabelUpsertService{}
		upsertSvc.On("UpsertScenarios", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(existingScenarios)), []string{in.ScenarioName}, mock.MatchedBy(matchRemoveScenarioFn([][]string{existingScenariosSlice}, in.ScenarioName))).
			Return(nil).Once()

		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		bundleInstanceAuthSvc.On("IsAnyExistForRuntimeAndScenario", ctx, []string{in.ScenarioName}, scenarioLabel.ObjectID).
			Return(false, nil).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, bundleInstanceAuthSvc)

		//WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, bundleInstanceAuthSvc)
	})

	t.Run("Failed when upsert scenario label returns error ", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    append(scenarios, selectorScenario),
			ObjectID: rtmID,
		}

		existingScenariosSlice := []string{"OTHER", "SCENARIO", selectorScenario}
		existingScenarios := map[string][]string{
			rtmID: existingScenariosSlice,
		}

		ctx := context.TODO()

		labels := []model.Label{scenarioLabel}
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return(labels, nil).Once()

		upsertSvc := &automock.LabelUpsertService{}
		upsertSvc.On("UpsertScenarios", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(existingScenarios)), []string{in.ScenarioName}, mock.MatchedBy(matchRemoveScenarioFn([][]string{existingScenariosSlice}, in.ScenarioName))).
			Return(testErr).Once()

		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		bundleInstanceAuthSvc.On("IsAnyExistForRuntimeAndScenario", ctx, []string{in.ScenarioName}, scenarioLabel.ObjectID).
			Return(false, nil).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, bundleInstanceAuthSvc)

		//WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, bundleInstanceAuthSvc)
	})

	t.Run("Failed when check for existing bundle instance auths for runtime and scenario returns error", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    append(scenarios, selectorScenario),
			ObjectID: rtmID,
		}

		ctx := context.TODO()

		runtimeIdForLabelThatCauseError := "foo"
		labels := []model.Label{scenarioLabel, {ObjectID: runtimeIdForLabelThatCauseError}}

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return(labels, nil).Once()

		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		bundleInstanceAuthSvc.On("IsAnyExistForRuntimeAndScenario", ctx, []string{in.ScenarioName}, scenarioLabel.ObjectID).Return(false, nil).Once()
		bundleInstanceAuthSvc.On("IsAnyExistForRuntimeAndScenario", ctx, []string{in.ScenarioName}, runtimeIdForLabelThatCauseError).Return(false, testErr).Once()

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, bundleInstanceAuthSvc)

		//WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, bundleInstanceAuthSvc)
	})

	t.Run("Failed when there are any existing bundle instance auths for this scenario", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    append(scenarios, selectorScenario),
			ObjectID: rtmID,
		}

		ctx := context.TODO()

		runtimeIdThatHasExistingBundleInstanceAuths := "foo"
		labels := []model.Label{scenarioLabel, {ObjectID: runtimeIdThatHasExistingBundleInstanceAuths}}

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return(labels, nil).Once()

		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		bundleInstanceAuthSvc.On("IsAnyExistForRuntimeAndScenario", ctx, []string{in.ScenarioName}, scenarioLabel.ObjectID).Return(false, nil).Once()
		bundleInstanceAuthSvc.On("IsAnyExistForRuntimeAndScenario", ctx, []string{in.ScenarioName}, runtimeIdThatHasExistingBundleInstanceAuths).Return(true, nil).Once()

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, bundleInstanceAuthSvc)

		//WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Bundle Instance Auths should be deleted first")
		mock.AssertExpectationsForObjects(t, labelRepo, bundleInstanceAuthSvc)
	})

	t.Run("Failed when GetRuntimeScenariosWhereRuntimesLabelsMatchSelector returns error", func(t *testing.T) {
		ctx := context.TODO()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return([]model.Label{}, testErr).Once()

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, nil)

		//WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo)

	})
}

func TestEngine_RemoveAssignedScenarios(t *testing.T) {
	selectorKey, selectorValue := "KEY", "VALUE"
	selectorScenario := "SCENARIO1"

	in := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName: "SCENARIO1",
			Tenant:       tenantID,
			Selector: model.LabelSelector{
				Key:   selectorKey,
				Value: selectorValue,
			}}}
	rtmID := "651038e0-e4b6-4036-a32f-f6e9846003f4"
	labels := []model.Label{{
		Value:    []interface{}{"SCENARIO1", "SCENARIO2"},
		Key:      model.ScenariosKey,
		ObjectID: rtmID,
	}}

	t.Run("Success", func(t *testing.T) {
		//GIVEN
		existingScenariosSlice := []string{"SCENARIO2", selectorScenario}
		existingScenarios := map[string][]string{
			rtmID: existingScenariosSlice,
		}

		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetRuntimeScenariosWhereLabelsMatchSelector", ctx, tenantID, selectorKey, selectorValue).
			Return(labels, nil).Once()

		upsertSvc := &automock.LabelUpsertService{}
		upsertSvc.On("UpsertScenarios", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(existingScenarios)), []string{selectorScenario}, mock.MatchedBy(matchRemoveScenarioFn([][]string{existingScenariosSlice}, selectorScenario))).
			Return(nil).Once()

		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		bundleInstanceAuthSvc.On("IsAnyExistForRuntimeAndScenario", ctx, []string{selectorScenario}, rtmID).
			Return(false, nil).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, bundleInstanceAuthSvc)

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

		bundleInstanceAuthSvc := &automock.BundleInstanceAuthService{}
		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, bundleInstanceAuthSvc)
		//WHEN
		err := eng.RemoveAssignedScenarios(ctx, in)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		labelRepo.AssertExpectations(t)
	})
}

func matchRemoveScenarioFn(scenariosPerRuntime [][]string, toRemove string) func(removeFn func(scenarios []string, toRemove string) []string) bool {
	return func(removeFn func(scenarios []string, toRemove string) []string) bool {
		for _, scenarios := range scenariosPerRuntime {
			result := removeFn(scenarios, toRemove)
			existForCurrentRuntimeScenarios := false

			for _, scenario := range result {
				if scenario == toRemove {
					existForCurrentRuntimeScenarios = true
					break
				}
			}

			if existForCurrentRuntimeScenarios {
				return false
			}
		}

		return true
	}
}

func matchAddNewScenarioFn(scenariosPerRuntime [][]string, toAdd string) func(addFn func(scenarios []string, toAdd string) []string) bool {
	return func(addFn func(scenarios []string, toRemove string) []string) bool {
		for _, scenarios := range scenariosPerRuntime {
			result := addFn(scenarios, toAdd)
			existForCurrentRuntimeScenarios := false

			for _, scenario := range result {
				if scenario == toAdd {
					existForCurrentRuntimeScenarios = true
					break
				}
			}

			if !existForCurrentRuntimeScenarios {
				return false
			}
		}

		return true
	}
}

func matchExpectedScenarios(expectedByRuntimeId map[string][]string) func(labels []model.Label) bool {
	return func(actual []model.Label) bool {
		labelsByRuntimeID := make(map[string]model.Label, 0)
		for _, label := range actual {
			labelsByRuntimeID[label.ObjectID] = label
		}

		if len(expectedByRuntimeId) != len(labelsByRuntimeID) {
			return false
		}

		for runtimeId, expectedScenarios := range expectedByRuntimeId {
			actualLabel, ok := labelsByRuntimeID[runtimeId]
			if !ok {
				return false
			}

			actualScenarios, err := labelpkg.GetScenariosAsStringSlice(actualLabel)
			if err != nil {
				return false
			}

			if !assert.ElementsMatch(dummyTest{}, expectedScenarios, actualScenarios) {
				return false
			}
		}

		return true
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

	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, nil)

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

	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, nil)

	// when
	_, err := engineSvc.GetScenariosForSelectorLabels(fixCtxWithTenant(), selectorLabels)

	// then
	require.Error(t, err)
	assert.EqualError(t, fmt.Errorf("while getting Automatic Scenario Assignments for selector [key: %s, val: %s]: %s", key, value, testErr.Error()), err.Error())
}

func TestEngine_GetScenariosForSelectorLabels_ShouldFailOnLoadingTenant(t *testing.T) {
	// given
	svc := scenarioassignment.NewEngine(nil, nil, nil, nil)
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
	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, nil)

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
	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, nil)

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
	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, nil)

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
	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, nil)

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

	engineSvc := scenarioassignment.NewEngine(nil, nil, nil, nil)

	// when
	actualScenarios := engineSvc.MergeScenarios(oldScenariosLabel, previousScenariosFromAssignments, newScenariosFromAssignments)

	// then
	assert.Equal(t, expectedScenarios, actualScenarios)
}

type dummyTest struct{}

func (t dummyTest) Errorf(string, ...interface{}) {}
