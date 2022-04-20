package scenarioassignment_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEngine_EnsureScenarioAssigned(t *testing.T) {
	selectorScenario := "SELECTOR_SCENARIO"
	in := fixAutomaticScenarioAssigment(selectorScenario)
	testErr := errors.New("test err")
	otherScenario := "OTHER"
	basicScenario := "SCENARIO"
	scenarios := []interface{}{otherScenario, basicScenario}
	stringScenarios := []string{otherScenario, basicScenario}

	rtmIDWithScenario := "rtm1_scenario"
	rtmIDWithoutScenario := "rtm1_no_scenario"

	expectedScenarios := map[string][]string{
		rtmIDWithScenario:    append(stringScenarios, selectorScenario),
		rtmIDWithoutScenario: []string{selectorScenario},
	}
	runtimes := []*model.Runtime{{ID: rtmIDWithoutScenario}, {ID: rtmIDWithScenario}}
	runtimesIDs := []string{rtmIDWithoutScenario, rtmIDWithScenario}
	scenarioLabel := model.Label{
		Key:        model.ScenariosKey,
		Value:      scenarios,
		ObjectID:   rtmIDWithScenario,
		ObjectType: model.RuntimeLabelableObject,
	}

	t.Run("Success", func(t *testing.T) {
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, runtimesIDs).
			Return([]model.Label{scenarioLabel}, nil)

		upsertSvc := &automock.LabelUpsertService{}
		upsertSvc.On("UpsertLabel", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(t, expectedScenarios))).Return(nil).Once()
		upsertSvc.On("UpsertLabel", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(t, expectedScenarios))).Return(nil).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, runtimeRepo)
	})

	t.Run("Failed when insert new Label on upsert failed ", func(t *testing.T) {
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, runtimesIDs).
			Return([]model.Label{scenarioLabel}, nil)

		upsertSvc := &automock.LabelUpsertService{}
		upsertSvc.On("UpsertLabel", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(t, expectedScenarios))).Return(nil).Once()
		upsertSvc.On("UpsertLabel", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(t, expectedScenarios))).Return(testErr).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, runtimeRepo)
	})

	t.Run("Failed when Label update on upsert failed ", func(t *testing.T) {
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    scenarios,
			ObjectID: rtmIDWithScenario,
		}

		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, []string{rtmIDWithoutScenario, rtmIDWithScenario}).
			Return([]model.Label{scenarioLabel}, nil)

		upsertSvc := &automock.LabelUpsertService{}
		upsertSvc.On("UpsertLabel", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(t, expectedScenarios))).Return(testErr).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, runtimeRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes returns error", func(t *testing.T) {
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, runtimesIDs).Return(nil, testErr)

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})

	t.Run("Failed when ListAll returns error", func(t *testing.T) {
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})

	t.Run("Success, no runtimes found", func(t *testing.T) {
		ctx := context.TODO()
		labelRepo := &automock.LabelRepository{}

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, nil).Once()

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.EnsureScenarioAssigned(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})
}

func TestEngine_RemoveAssignedScenario(t *testing.T) {
	selectorScenario := "SELECTOR_SCENARIO"
	rtmID := "8c4de4d8-dcfa-47a9-95c9-3c8b1f5b907c"
	in := fixAutomaticScenarioAssigment(selectorScenario)
	runtimes := []*model.Runtime{{ID: rtmID}}
	testErr := errors.New("test err")

	t.Run("Success", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    append(scenarios, selectorScenario),
			ObjectID: rtmID,
		}

		expectedScenarios := map[string][]string{
			rtmID: {"OTHER", "SCENARIO"},
		}

		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, []string{rtmID}).
			Return([]model.Label{scenarioLabel}, nil)

		upsertSvc := &automock.LabelUpsertService{}
		upsertSvc.On("UpsertLabel", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(t, expectedScenarios))).
			Return(nil).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, runtimeRepo)
	})

	t.Run("Success, empty scenarios label deleted", func(t *testing.T) {
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    []interface{}{selectorScenario},
			ObjectID: rtmID,
		}
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, []string{rtmID}).
			Return([]model.Label{scenarioLabel}, nil)

		labelRepo.On("Delete", ctx, tenantID, model.RuntimeLabelableObject, rtmID, model.ScenariosKey).Return(nil)

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})

	t.Run("Failed when Label Upsert failed ", func(t *testing.T) {
		scenarios := []interface{}{"OTHER", "SCENARIO"}
		scenarioLabel := model.Label{
			Key:      model.ScenariosKey,
			Value:    append(scenarios, selectorScenario),
			ObjectID: rtmID,
		}
		expectedScenarios := map[string][]string{rtmID: {"OTHER", "SCENARIO"}}

		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, []string{rtmID}).
			Return([]model.Label{scenarioLabel}, nil)

		upsertSvc := &automock.LabelUpsertService{}
		upsertSvc.On("UpsertLabel", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(t, expectedScenarios))).
			Return(testErr).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, runtimeRepo)
	})

	t.Run("Failed when ListAll returns error", func(t *testing.T) {
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()

		labelRepo := &automock.LabelRepository{}

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})

	t.Run("Failed when GetScenarioLabelsForRuntimes failed", func(t *testing.T) {
		ctx := context.TODO()

		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, []string{rtmID}).
			Return(nil, testErr)

		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.RemoveAssignedScenario(ctx, in)

		// THEN
		require.Error(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})
}

func TestEngine_RemoveAssignedScenarios(t *testing.T) {
	in := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   "SCENARIO1",
			Tenant:         tenantID,
			TargetTenantID: targetTenantID,
		},
	}
	rtmID := "651038e0-e4b6-4036-a32f-f6e9846003f4"
	runtimes := []*model.Runtime{{ID: rtmID}}
	labels := []model.Label{{
		Value:    []interface{}{"SCENARIO1", "SCENARIO2"},
		Key:      model.ScenariosKey,
		ObjectID: rtmID,
	}}

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		expectedScenarios := map[string][]string{
			rtmID: {"SCENARIO2"},
		}

		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, []string{rtmID}).
			Return(labels, nil)

		upsertSvc := &automock.LabelUpsertService{}
		upsertSvc.On("UpsertLabel", ctx, tenantID, mock.MatchedBy(matchExpectedScenarios(t, expectedScenarios))).
			Return(nil).Once()

		eng := scenarioassignment.NewEngine(upsertSvc, labelRepo, nil, runtimeRepo)

		// WHEN
		err := eng.RemoveAssignedScenarios(ctx, in)

		// THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, labelRepo, upsertSvc, runtimeRepo)
	})

	t.Run("Error, while removing scenario - ListAll fail", func(t *testing.T) {
		// GIVEN
		testErr := errors.New("test error")
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(nil, testErr).Once()

		labelRepo := &automock.LabelRepository{}
		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, runtimeRepo)
		// WHEN
		err := eng.RemoveAssignedScenarios(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})

	t.Run("Error, while removing scenario - GetScenarioLabelsForRuntimes fail", func(t *testing.T) {
		// GIVEN
		testErr := errors.New("test error")
		ctx := context.TODO()
		runtimeRepo := &automock.RuntimeRepository{}
		runtimeRepo.On("ListAll", ctx, targetTenantID, []*labelfilter.LabelFilter(nil)).Return(runtimes, nil).Once()

		labelRepo := &automock.LabelRepository{}
		labelRepo.On("GetScenarioLabelsForRuntimes", ctx, tenantID, []string{rtmID}).
			Return(nil, testErr)
		eng := scenarioassignment.NewEngine(nil, labelRepo, nil, runtimeRepo)
		// WHEN
		err := eng.RemoveAssignedScenarios(ctx, in)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		mock.AssertExpectationsForObjects(t, labelRepo, runtimeRepo)
	})
}

func TestEngine_MergeScenariosFromInputLabelsAndAssignments_Success(t *testing.T) {
	// GIVEN
	differentTargetTenant := "differentTargetTenant"
	runtimeID := "runtimeID"
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioName,
			Tenant:         tenantID,
			TargetTenantID: targetTenantID,
		},
		{
			ScenarioName:   scenarioName2,
			Tenant:         tenantID,
			TargetTenantID: differentTargetTenant,
		},
	}

	expectedScenarios := []interface{}{scenarioName}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListAll", fixCtxWithTenant(), tenantID).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", fixCtxWithTenant(), targetTenantID, runtimeID).Return(true, nil).Once()
	runtimeRepo.On("Exists", fixCtxWithTenant(), differentTargetTenant, runtimeID).Return(false, nil).Once()

	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, runtimeRepo)

	// WHEN
	actualScenarios, err := engineSvc.MergeScenariosFromInputLabelsAndAssignments(fixCtxWithTenant(), inputLabels, runtimeID)

	// THEN

	require.NoError(t, err)
	require.ElementsMatch(t, expectedScenarios, actualScenarios)

	mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)
}

func TestEngine_MergeScenariosFromInputLabelsAndAssignments_SuccessIfScenariosLabelIsInInput(t *testing.T) {
	// GIVEN
	runtimeID := "runtimeID"
	labelKey := "key"
	labelValue := "val"

	scenario := "SCENARIO"
	inputLabels := map[string]interface{}{
		labelKey:           labelValue,
		model.ScenariosKey: []interface{}{scenario},
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioName,
			Tenant:         tenantID,
			TargetTenantID: targetTenantID,
		},
	}

	expectedScenarios := []interface{}{scenarioName, scenario}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListAll", fixCtxWithTenant(), tenantID).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", fixCtxWithTenant(), targetTenantID, runtimeID).Return(true, nil).Once()

	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, runtimeRepo)

	// WHEN
	actualScenarios, err := engineSvc.MergeScenariosFromInputLabelsAndAssignments(fixCtxWithTenant(), inputLabels, runtimeID)

	// THEN
	require.NoError(t, err)
	require.ElementsMatch(t, expectedScenarios, actualScenarios)

	mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)
}

func TestEngine_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfListAllFailed(t *testing.T) {
	// GIVEN
	testErr := errors.New("testErr")
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListAll", fixCtxWithTenant(), tenantID).Return(nil, testErr)
	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, nil)

	// WHEN
	_, err := engineSvc.MergeScenariosFromInputLabelsAndAssignments(fixCtxWithTenant(), inputLabels, "runtimeID")

	// THEN
	require.Error(t, err)

	mockRepo.AssertExpectations(t)
}

func TestEngine_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfExistsFailed(t *testing.T) {
	// GIVEN
	runtimeID := "runtimeID"
	testErr := errors.New("testErr")
	labelKey := "key"
	labelValue := "val"

	inputLabels := map[string]interface{}{
		labelKey: labelValue,
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioName,
			Tenant:         tenantID,
			TargetTenantID: targetTenantID,
		},
	}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListAll", fixCtxWithTenant(), tenantID).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", fixCtxWithTenant(), targetTenantID, runtimeID).Return(false, testErr).Once()

	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, runtimeRepo)
	// WHEN
	_, err := engineSvc.MergeScenariosFromInputLabelsAndAssignments(fixCtxWithTenant(), inputLabels, runtimeID)

	// THEN
	require.Error(t, err)

	mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)
}

func TestEngine_MergeScenariosFromInputLabelsAndAssignments_ReturnsErrorIfScenariosFromInputWereNotInterfaceSlice(t *testing.T) {
	// GIVEN
	runtimeID := "runtimeID"
	labelKey := "key"
	labelValue := "val"

	scenario := "SCENARIO"
	inputLabels := map[string]interface{}{
		labelKey:           labelValue,
		model.ScenariosKey: []string{scenario},
	}

	assignments := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioName,
			Tenant:         tenantID,
			TargetTenantID: targetTenantID,
		},
	}

	mockRepo := &automock.Repository{}
	mockRepo.On("ListAll", fixCtxWithTenant(), tenantID).Return(assignments, nil)

	runtimeRepo := &automock.RuntimeRepository{}
	runtimeRepo.On("Exists", fixCtxWithTenant(), targetTenantID, runtimeID).Return(true, nil).Once()

	engineSvc := scenarioassignment.NewEngine(nil, nil, mockRepo, runtimeRepo)

	// WHEN
	_, err := engineSvc.MergeScenariosFromInputLabelsAndAssignments(fixCtxWithTenant(), inputLabels, runtimeID)

	// THEN
	require.Error(t, err)

	mock.AssertExpectationsForObjects(t, mockRepo, runtimeRepo)
}

func Test_engine_GetScenariosFromMatchingASAs(t *testing.T) {
	ctx := fixCtxWithTenant()
	testErr := errors.New(errMsg)
	testScenarios := []*model.AutomaticScenarioAssignment{
		{
			ScenarioName:   scenarioName,
			Tenant:         tenantID,
			TargetTenantID: targetTenantID,
		},
		{
			ScenarioName:   scenarioName2,
			Tenant:         tenantID2,
			TargetTenantID: targetTenantID2,
		},
	}

	testCases := []struct {
		Name                     string
		LabelServiceFn           func() *automock.LabelUpsertService
		LabelRepoFn              func() *automock.LabelRepository
		ScenarioAssignmentRepoFn func() *automock.Repository
		RuntimeRepoFn            func() *automock.RuntimeRepository
		RuntimeID                string
		ExpectedError            error
		ExpectedScenarios        []string
	}{
		{
			Name:           "Success",
			LabelServiceFn: unusedLabelService,
			LabelRepoFn:    unusedLabelRepo,
			ScenarioAssignmentRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListAll", ctx, tenantID).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, targetTenantID, runtimeID).Return(true, nil)
				repo.On("Exists", ctx, targetTenantID2, runtimeID).Return(false, nil)
				return repo
			},
			RuntimeID:         runtimeID,
			ExpectedError:     nil,
			ExpectedScenarios: []string{scenarioName},
		},
		{
			Name:           "Returns error when can't list ASAs",
			LabelServiceFn: unusedLabelService,
			LabelRepoFn:    unusedLabelRepo,
			ScenarioAssignmentRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListAll", ctx, tenantID).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFn:     unusedRuntimeRepo,
			RuntimeID:         runtimeID,
			ExpectedError:     testErr,
			ExpectedScenarios: nil,
		},
		{
			Name:           "Returns error when can't list ASAs",
			LabelServiceFn: unusedLabelService,
			LabelRepoFn:    unusedLabelRepo,
			ScenarioAssignmentRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListAll", ctx, tenantID).Return(nil, testErr)
				return repo
			},
			RuntimeRepoFn:     unusedRuntimeRepo,
			RuntimeID:         runtimeID,
			ExpectedError:     testErr,
			ExpectedScenarios: nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			engine := scenarioassignment.NewEngine(testCase.LabelServiceFn(), testCase.LabelRepoFn(), testCase.ScenarioAssignmentRepoFn(), testCase.RuntimeRepoFn())

			// WHEN
			scenarios, err := engine.GetScenariosFromMatchingASAs(ctx, testCase.RuntimeID)

			// THEN
			if testCase.ExpectedError == nil {
				require.ElementsMatch(t, scenarios, testCase.ExpectedScenarios)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedError.Error())
				require.Nil(t, testCase.ExpectedScenarios)
			}
		})
	}
}
