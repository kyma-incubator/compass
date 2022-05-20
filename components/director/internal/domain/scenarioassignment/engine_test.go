package scenarioassignment_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
			Name:           "Returns error when checking if asa matches runtime",
			LabelServiceFn: unusedLabelService,
			LabelRepoFn:    unusedLabelRepo,
			ScenarioAssignmentRepoFn: func() *automock.Repository {
				repo := &automock.Repository{}
				repo.On("ListAll", ctx, tenantID).Return(testScenarios, nil)
				return repo
			},
			RuntimeRepoFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, targetTenantID, runtimeID).Return(false, testErr)
				return repo
			},
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
