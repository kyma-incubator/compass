package runtime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var labelsWithNormalization = map[string]interface{}{runtime.IsNormalizedLabel: "true"}

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	desc := "Lorem ipsum"
	labels := map[string]interface{}{
		model.ScenariosKey:          "DEFAULT",
		"protected_defaultEventing": "true",
	}
	labelsForDbMock := map[string]interface{}{
		model.ScenariosKey:        []interface{}{"DEFAULT"},
		runtime.IsNormalizedLabel: "true",
	}

	modelInput := model.RuntimeInput{
		Name:        "foo.bar-not",
		Description: &desc,
		Labels:      labels,
	}

	modelInputWithoutLabels := model.RuntimeInput{
		Name:        "foo.bar-not",
		Description: &desc,
	}
	var nilLabels map[string]interface{}

	runtimeModel := mock.MatchedBy(func(rtm *model.Runtime) bool {
		return rtm.Name == modelInput.Name && rtm.Description == modelInput.Description &&
			rtm.Status.Condition == model.RuntimeStatusConditionInitial
	})

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                 string
		RuntimeRepositoryFn  func() *automock.RuntimeRepository
		ScenariosServiceFn   func() *automock.ScenariosService
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		UIDServiceFn         func() *automock.UIDService
		EngineServiceFn      func() *automock.ScenarioAssignmentEngine
		Input                model.RuntimeInput
		ExpectedErr          error
	}{
		{
			Name: "Success",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("EnsureScenariosLabelDefinitionExists", contextThatHasTenant(tnt), tnt).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeLabelableObject, id, labelsForDbMock).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels).Return([]interface{}{"DEFAULT"}, nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Success when labels are empty",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("EnsureScenariosLabelDefinitionExists", contextThatHasTenant(tnt), tnt).Return(nil).Once()
				repo.On("AddDefaultScenarioIfEnabled", mock.Anything, &nilLabels).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeLabelableObject, id, labelsWithNormalization).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, nilLabels).Return([]interface{}{}, nil)
				return svc
			},
			Input:       modelInputWithoutLabels,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when ensuring default label definition failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("EnsureScenariosLabelDefinitionExists", contextThatHasTenant(tnt), tnt).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when runtime creation failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, runtimeModel).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("").Once()
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Return error when merge of scenarios and assignments failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("EnsureScenariosLabelDefinitionExists", contextThatHasTenant(tnt), tnt).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels).Return(nil, testErr)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
		{
			Name: "Returns error when label upserting failed",
			RuntimeRepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Create", ctx, runtimeModel).Return(nil).Once()
				return repo
			},
			ScenariosServiceFn: func() *automock.ScenariosService {
				repo := &automock.ScenariosService{}
				repo.On("EnsureScenariosLabelDefinitionExists", contextThatHasTenant(tnt), tnt).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, "tenant", model.RuntimeLabelableObject, id, labelsForDbMock).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels).Return([]interface{}{"DEFAULT"}, nil)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RuntimeRepositoryFn()
			idSvc := testCase.UIDServiceFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			scenariosSvc := testCase.ScenariosServiceFn()
			engineSvc := testCase.EngineServiceFn()
			svc := runtime.NewService(repo, nil, scenariosSvc, labelSvc, idSvc, engineSvc, ".*_defaultEventing$")

			// when
			result, err := svc.Create(ctx, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			if err == nil {
				require.Nil(t, testCase.ExpectedErr)
			} else {
				require.NotNil(t, testCase.ExpectedErr)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			idSvc.AssertExpectations(t)
			labelSvc.AssertExpectations(t)
			scenariosSvc.AssertExpectations(t)
			engineSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, ".*_defaultEventing$")
		// when
		_, err := svc.Create(context.TODO(), model.RuntimeInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"

	labelsDbMock := map[string]interface{}{
		"label1":                  "val1",
		"scenarios":               []interface{}{"SCENARIO"},
		runtime.IsNormalizedLabel: "true",
	}
	labels := map[string]interface{}{
		"label1": "val1",
	}
	protectedLabels := map[string]interface{}{
		"protected_defaultEventing": "true",
		"label1":                    "val1",
	}
	modelInput := model.RuntimeInput{
		Name:   "bar",
		Labels: labels,
	}

	modelInputWithProtectedLabels := model.RuntimeInput{
		Name:   "bar",
		Labels: protectedLabels,
	}

	inputRuntimeModel := mock.MatchedBy(func(rtm *model.Runtime) bool {
		return rtm.Name == modelInput.Name
	})

	inputProtectedRuntimeModel := mock.MatchedBy(func(rtm *model.Runtime) bool {
		return rtm.Name == modelInput.Name
	})

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.RuntimeRepository
		LabelRepositoryFn    func() *automock.LabelRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		EngineServiceFn      func() *automock.ScenarioAssignmentEngine
		Input                model.RuntimeInput
		InputID              string
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, modelInput.Labels).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels).Return([]interface{}{}, nil)
				return svc
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when updating with protected labels",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputProtectedRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, labels).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels).Return([]interface{}{}, nil)
				return svc
			},
			InputID:            "foo",
			Input:              modelInputWithProtectedLabels,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when there are scenarios to set from assignments",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, labelsDbMock).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels).Return([]interface{}{"SCENARIO"}, nil)
				return svc
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when labels are nil",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, labelsWithNormalization).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labelsWithNormalization).Return([]interface{}{}, nil)
				return svc
			},
			InputID: "foo",
			Input: model.RuntimeInput{
				Name: "bar",
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime update failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeModel).Return(testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(nil, testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label deletion failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error if merge of scenarios and assignments failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels).Return(nil, testErr)
				return svc
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when upserting labels failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, "foo").Return(runtimeModel, nil).Once()
				repo.On("Update", ctx, inputRuntimeModel).Return(nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("DeleteByKeyNegationPattern", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, mock.AnythingOfType("string")).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				repo := &automock.LabelUpsertService{}
				repo.On("UpsertMultipleLabels", ctx, tnt, model.RuntimeLabelableObject, runtimeModel.ID, modelInput.Labels).Return(testErr).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, labels).Return([]interface{}{}, nil)
				return svc
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			engineSvc := testCase.EngineServiceFn()
			svc := runtime.NewService(repo, labelRepo, nil, labelSvc, nil, engineSvc, ".*_defaultEventing$")

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
			labelSvc.AssertExpectations(t)
			engineSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, "")
		// when
		err := svc.Update(context.TODO(), "id", model.RuntimeInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	desc := "Lorem ipsum"

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		Input              model.RuntimeInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Delete", ctx, tnt, runtimeModel.ID).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime deletion failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Delete", ctx, tnt, runtimeModel.ID).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, "")

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, "")
		// when
		err := svc.Delete(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	desc := "Lorem ipsum"
	tnt := "tenant"
	externalTnt := "external-tnt"

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		Input              model.RuntimeInput
		InputID            string
		ExpectedRuntime    *model.Runtime
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(runtimeModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByID", ctx, tnt, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, "")

			// when
			rtm, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedRuntime, rtm)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, "")
		// when
		_, err := svc.Get(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_GetByTokenIssuer(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	desc := "Lorem ipsum"
	tokenIssuer := "https://dex.domain.local"
	filter := []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("runtime_consoleUrl", `"https://console.domain.local"`)}

	runtimeModel := &model.Runtime{
		ID:          "foo",
		Name:        "Foo",
		Description: &desc,
	}

	ctx := context.TODO()

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		Input              model.RuntimeInput
		InputID            string
		ExpectedRuntime    *model.Runtime
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFiltersGlobal", ctx, filter).Return(runtimeModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime retrieval failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFiltersGlobal", ctx, filter).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedRuntime:    runtimeModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, "")

			// when
			rtm, err := svc.GetByTokenIssuer(ctx, tokenIssuer)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedRuntime, rtm)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Exist(t *testing.T) {
	tnt := "tenant"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)
	testError := errors.New("Test error")

	rtmID := "id"

	testCases := []struct {
		Name           string
		RepositoryFn   func() *automock.RuntimeRepository
		InputRuntimeID string
		ExpectedValue  bool
		ExpectedError  error
	}{
		{
			Name: "Runtime exits",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, rtmID).Return(true, nil)
				return repo
			},
			InputRuntimeID: rtmID,
			ExpectedValue:  true,
			ExpectedError:  nil,
		},
		{
			Name: "Runtime not exits",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, rtmID).Return(false, nil)
				return repo
			},
			InputRuntimeID: rtmID,
			ExpectedValue:  false,
			ExpectedError:  nil,
		},
		{
			Name: "Returns error",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, rtmID).Return(false, testError)
				return repo
			},
			InputRuntimeID: rtmID,
			ExpectedValue:  false,
			ExpectedError:  testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			rtmRepo := testCase.RepositoryFn()
			svc := runtime.NewService(rtmRepo, nil, nil, nil, nil, nil, "")

			// WHEN
			value, err := svc.Exist(ctx, testCase.InputRuntimeID)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, testCase.ExpectedValue, value)
			rtmRepo.AssertExpectations(t)
		})
	}
	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, "")
		// when
		_, err := svc.Exist(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_List(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelRuntimes := []*model.Runtime{
		fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem Ipsum"),
		fixModelRuntime(t, "bar", "tenant-bar", "Bar", "Lorem Ipsum"),
	}
	runtimePage := &model.RuntimePage{
		Data:       modelRuntimes,
		TotalCount: len(modelRuntimes),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	first := 2
	after := "test"
	filter := []*labelfilter.LabelFilter{{Key: ""}}

	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		InputLabelFilters  []*labelfilter.LabelFilter
		InputPageSize      int
		InputCursor        string
		ExpectedResult     *model.RuntimePage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("List", ctx, tnt, filter, first, after).Return(runtimePage, nil).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     runtimePage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when runtime listing failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("List", ctx, tnt, filter, first, after).Return(nil, testErr).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      first,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when pageSize is less than 1",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      0,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
		{
			Name: "Returns error when pageSize is bigger than 200",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      201,
			InputCursor:        after,
			ExpectedResult:     nil,
			ExpectedErrMessage: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := runtime.NewService(repo, nil, nil, nil, nil, nil, "")

			// when
			rtm, err := svc.List(ctx, testCase.InputLabelFilters, testCase.InputPageSize, testCase.InputCursor)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, rtm)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, "")
		// when
		_, err := svc.List(context.TODO(), nil, 1, "")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_GetLabel(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Tenant:     tnt,
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputRuntimeID     string
		InputLabel         *model.LabelInput
		ExpectedLabel      *model.Label
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(modelLabel, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedLabel:      modelLabel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when label receiving failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("GetByKey", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil, testErr).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedLabel:      nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when exists function for runtime failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime doesn't exist",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil, nil, nil, "")

			// when
			l, err := svc.GetLabel(ctx, testCase.InputRuntimeID, testCase.InputLabel.Key)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, l, testCase.ExpectedLabel)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, "")
		// when
		_, err := svc.GetLabel(context.TODO(), "id", "key")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_ListLabel(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"
	labelKey := "key"
	labelValue := []string{"value1"}

	label := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c11",
		Tenant:     tnt,
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	protectedModelLabel := &model.Label{
		ID:         "5d23d9d9-3d04-4fa9-95e6-d22e1ae62c12",
		Tenant:     tnt,
		Key:        "protected_defaultEventing",
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	labels := map[string]*model.Label{"protected_defaultEventing": protectedModelLabel, "first": modelLabel, "second": modelLabel}
	expectedLabelWithoutProtected := map[string]*model.Label{"first": modelLabel, "second": modelLabel}
	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		LabelRepositoryFn  func() *automock.LabelRepository
		InputRuntimeID     string
		InputLabel         *model.LabelInput
		ExpectedOutput     map[string]*model.Label
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labels, nil).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedOutput:     expectedLabelWithoutProtected,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when labels receiving failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedOutput:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime exists function failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime does not exists",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()

				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         label,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			svc := runtime.NewService(repo, labelRepo, nil, nil, nil, nil, ".*_defaultEventing$")

			// when
			l, err := svc.ListLabels(ctx, testCase.InputRuntimeID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedOutput, l)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, "")
		// when
		_, err := svc.ListLabels(context.TODO(), "id")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_SetLabel(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"

	labelKey := "key"
	protectedLabelKey := "protected_defaultEventing"

	modelLabelInput := model.LabelInput{
		Key:        labelKey,
		Value:      []string{"value1"},
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	modelProtectedLabelInput := model.LabelInput{
		Key:        protectedLabelKey,
		Value:      []string{"value1"},
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	scenariosLabelValue := []interface{}{"SCENARIO"}
	modelScenariosLabelInput := model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      scenariosLabelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	labelMapWithScenariosLabel := map[string]*model.Label{
		model.ScenariosKey: {
			ID:         "id",
			Tenant:     "tenant",
			Key:        model.ScenariosKey,
			Value:      scenariosLabelValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	labelMapWithInvalidScenariosLabel := map[string]*model.Label{
		model.ScenariosKey: {
			ID:         "id",
			Tenant:     "tenant",
			Key:        model.ScenariosKey,
			Value:      []int{},
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	labelMap := map[string]*model.Label{
		labelKey: {
			ID:         "id",
			Tenant:     "tenant",
			Key:        labelKey,
			Value:      []string{"val"},
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.RuntimeRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		LabelRepositoryFn    func() *automock.LabelRepository
		EngineServiceFn      func() *automock.ScenarioAssignmentEngine
		InputRuntimeID       string
		InputLabel           *model.LabelInput
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, &modelLabelInput).Return(nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, []interface{}{}, []interface{}{}).Return([]interface{}{}, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when label key is scenarios",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, &modelScenariosLabelInput).Return(nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabel, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{model.ScenariosKey: scenariosLabelValue}).Return(scenariosLabelValue, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelScenariosLabelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when label from input is selector",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, &modelScenariosLabelInput).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, &modelLabelInput).Return(nil).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, []interface{}{}, []interface{}{}).Return(scenariosLabelValue, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when checking if runtime exists failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime doesn't exists",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
		{
			Name: "Returns error when getting current labels for runtime failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label key is scenarios and merge scenarios and assignments failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabel, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{model.ScenariosKey: scenariosLabelValue}).Return(nil, testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelScenariosLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label from input is selector and getScenariosForSelectorLabels failed during getting old scenarios label and previous scenarios from assignments",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return(nil, testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label from input is selector and getScenariosForSelectorLabels failed during getting old scenarios label and previous scenarios from assignments",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return(nil, testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label from input is selector and upserting scenario label failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, &modelScenariosLabelInput).Return(testErr).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, []interface{}{}, []interface{}{}).Return(scenariosLabelValue, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label from input is selector and upserting label failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				svc.On("UpsertLabel", ctx, tnt, &modelScenariosLabelInput).Return(nil).Once()
				svc.On("UpsertLabel", ctx, tnt, &modelLabelInput).Return(testErr).Once()
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, []interface{}{}, []interface{}{}).Return(scenariosLabelValue, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when scenarios label value is not []interface{}",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithInvalidScenariosLabel, nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelLabelInput,
			ExpectedErrMessage: "value for scenarios label must be []interface{}",
		},
		{
			Name: "Returns an error when trying to set protected label",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(nil).Once()
				return repo
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, []interface{}{}, []interface{}{}).Return([]interface{}{}, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputLabel:         &modelProtectedLabelInput,
			ExpectedErrMessage: "could not set protected label key protected_defaultEventing",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelSvc := testCase.LabelUpsertServiceFn()
			labelRepo := testCase.LabelRepositoryFn()
			engineSvc := testCase.EngineServiceFn()
			svc := runtime.NewService(repo, labelRepo, nil, labelSvc, nil, engineSvc, ".*_defaultEventing$")

			// when
			err := svc.SetLabel(ctx, testCase.InputLabel)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelSvc.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
			engineSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, ".*_defaultEventing$")
		// when
		err := svc.SetLabel(context.TODO(), &model.LabelInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_DeleteLabel(t *testing.T) {
	// given
	tnt := "tenant"
	externalTnt := "external-tnt"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	testErr := errors.New("Test error")

	runtimeID := "foo"

	labelKey := "key"
	protectedLabelKey := "protected_defaultEventing"
	labelValue := "val"
	labelKey2 := "key2"
	scenario := "SCENARIO"
	secondScenario := "SECOND_SCENARIO"
	scenariosLabelValue := []interface{}{scenario}
	scenariosLabelValueWithMultipleValues := []interface{}{scenario, secondScenario}

	labelMap := map[string]*model.Label{
		labelKey: {
			ID:         "id",
			Tenant:     "tenant",
			Key:        labelKey,
			Value:      []string{"val"},
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	labelMapWithScenariosLabel := map[string]*model.Label{
		model.ScenariosKey: {
			ID:         "id",
			Tenant:     "tenant",
			Key:        model.ScenariosKey,
			Value:      scenariosLabelValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	labelMapWithScenariosLabelWithMultipleValues := map[string]*model.Label{
		model.ScenariosKey: {
			ID:         "id",
			Tenant:     "tenant",
			Key:        model.ScenariosKey,
			Value:      scenariosLabelValueWithMultipleValues,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
		labelKey: {
			ID:         "id",
			Tenant:     "tenant",
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	labelSelectorValue := "selector"

	labelMapWithTwoSelectors := map[string]*model.Label{
		labelKey: {
			ID:         "id",
			Tenant:     "tenant",
			Key:        labelKey,
			Value:      labelSelectorValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
		labelKey2: {
			ID:         "id",
			Tenant:     "tenant",
			Key:        labelKey2,
			Value:      labelSelectorValue,
			ObjectID:   "obj-id",
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	testCases := []struct {
		Name                 string
		RepositoryFn         func() *automock.RuntimeRepository
		LabelRepositoryFn    func() *automock.LabelRepository
		LabelUpsertServiceFn func() *automock.LabelUpsertService
		EngineServiceFn      func() *automock.ScenarioAssignmentEngine
		InputRuntimeID       string
		InputKey             string
		ExpectedErrMessage   string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, []interface{}{}, []interface{}{}).Return([]interface{}{}, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when label key is scenarios",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabelWithMultipleValues, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				modelLabelInput := &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []interface{}{scenario, secondScenario},
					ObjectID:   runtimeID,
					ObjectType: model.RuntimeLabelableObject,
				}
				svc.On("UpsertLabel", ctx, tnt, modelLabelInput).Return(nil).Once()
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{labelKey: labelValue}).Return(scenariosLabelValueWithMultipleValues, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           model.ScenariosKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when label key is selector",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithTwoSelectors, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey: labelSelectorValue, labelKey2: labelSelectorValue}).Return([]string{scenario}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey2: labelSelectorValue}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, scenariosLabelValue, []interface{}{}).Return([]interface{}{}, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Success when label key is selector and the scenarios label has to be restored",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithTwoSelectors, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				modelLabelInput := &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []interface{}{secondScenario},
					ObjectID:   runtimeID,
					ObjectType: model.RuntimeLabelableObject,
				}
				svc.On("UpsertLabel", ctx, tnt, modelLabelInput).Return(nil).Once()
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey: labelSelectorValue, labelKey2: labelSelectorValue}).Return([]string{scenario, secondScenario}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey2: labelSelectorValue}).Return([]string{secondScenario}, nil).Once()
				svc.On("MergeScenarios", nilInterface, []interface{}{scenario, secondScenario}, []interface{}{secondScenario}).Return([]interface{}{secondScenario}, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when checking if runtime exists failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, testErr).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when checking if runtime does not exists",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(false, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: fmt.Sprintf("Runtime with ID %s doesn't exist", runtimeID),
		},
		{
			Name: "Returns error if listing current labels for runtime failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(nil, testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label key is scenarios and merging scenarios and input assignments for old labels failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabel, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{}).Return(nil, testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           model.ScenariosKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label key is not scenarios label and getting scenarios for selector failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithTwoSelectors, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey: labelSelectorValue, labelKey2: labelSelectorValue}).Return(nil, testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when label key is not scenarios label and getting scenarios for selector for new labels failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithTwoSelectors, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey: labelSelectorValue, labelKey2: labelSelectorValue}).Return([]string{scenario}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey2: labelSelectorValue}).Return(nil, testErr).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime scenario label delete failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithTwoSelectors, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey: labelSelectorValue, labelKey2: labelSelectorValue}).Return([]string{scenario}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey2: labelSelectorValue}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, scenariosLabelValue, []interface{}{}).Return([]interface{}{}, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when runtime label delete failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithTwoSelectors, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, labelKey).Return(testErr).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey: labelSelectorValue, labelKey2: labelSelectorValue}).Return([]string{scenario}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{labelKey2: labelSelectorValue}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, scenariosLabelValue, []interface{}{}).Return([]interface{}{}, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           labelKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when upserting scenarios label failed",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMapWithScenariosLabelWithMultipleValues, nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				modelLabelInput := &model.LabelInput{
					Key:        model.ScenariosKey,
					Value:      []interface{}{scenario, secondScenario},
					ObjectID:   runtimeID,
					ObjectType: model.RuntimeLabelableObject,
				}
				svc.On("UpsertLabel", ctx, tnt, modelLabelInput).Return(testErr).Once()
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("MergeScenariosFromInputLabelsAndAssignments", ctx, map[string]interface{}{labelKey: labelValue}).Return(scenariosLabelValueWithMultipleValues, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           model.ScenariosKey,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns an error when trying to delete protected label",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("Exists", ctx, tnt, runtimeID).Return(true, nil).Once()
				return repo
			},
			LabelRepositoryFn: func() *automock.LabelRepository {
				repo := &automock.LabelRepository{}
				repo.On("ListForObject", ctx, tnt, model.RuntimeLabelableObject, runtimeID).Return(labelMap, nil).Once()
				repo.On("Delete", ctx, tnt, model.RuntimeLabelableObject, runtimeID, model.ScenariosKey).Return(nil).Once()
				return repo
			},
			LabelUpsertServiceFn: func() *automock.LabelUpsertService {
				svc := &automock.LabelUpsertService{}
				return svc
			},
			EngineServiceFn: func() *automock.ScenarioAssignmentEngine {
				var nilInterface []interface{}

				svc := &automock.ScenarioAssignmentEngine{}
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("GetScenariosForSelectorLabels", ctx, map[string]string{}).Return([]string{}, nil).Once()
				svc.On("MergeScenarios", nilInterface, []interface{}{}, []interface{}{}).Return([]interface{}{}, nil).Once()
				return svc
			},
			InputRuntimeID:     runtimeID,
			InputKey:           protectedLabelKey,
			ExpectedErrMessage: "could not delete protected label key protected_defaultEventing",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepo := testCase.LabelRepositoryFn()
			labelUpsertSvc := testCase.LabelUpsertServiceFn()
			engineSvc := testCase.EngineServiceFn()
			svc := runtime.NewService(repo, labelRepo, nil, labelUpsertSvc, nil, engineSvc, ".*_defaultEventing$")

			// when
			err := svc.DeleteLabel(ctx, testCase.InputRuntimeID, testCase.InputKey)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepo.AssertExpectations(t)
			labelUpsertSvc.AssertExpectations(t)
			engineSvc.AssertExpectations(t)
		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, ".*_defaultEventing$")
		// when
		err := svc.DeleteLabel(context.TODO(), "id", "key")
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_UpdateTenantID(t *testing.T) {
	// given
	testErr := errors.New("Test error")
	runtimeID := "sample-runtime-id"
	tntID := "tenantID"
	ctx := context.TODO()

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("UpdateTenantID", ctx, runtimeID, tntID).Return(nil).Once()
				return repo
			},

			ExpectedErrMessage: "",
		},
		{
			Name: "Fails on repository error",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("UpdateTenantID", ctx, runtimeID, tntID).Return(testErr).Once()
				return repo
			},

			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepository := &automock.LabelRepository{}
			labelUpsertService := &automock.LabelUpsertService{}
			scenariosService := &automock.ScenariosService{}
			scenarioAssignmentEngine := &automock.ScenarioAssignmentEngine{}
			uidSvc := &automock.UIDService{}
			svc := runtime.NewService(repo, labelRepository, scenariosService, labelUpsertService, uidSvc, scenarioAssignmentEngine, ".*_defaultEventing$")

			// when
			err := svc.UpdateTenantID(ctx, runtimeID, tntID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepository.AssertExpectations(t)
			labelUpsertService.AssertExpectations(t)
			scenariosService.AssertExpectations(t)
			scenarioAssignmentEngine.AssertExpectations(t)
			uidSvc.AssertExpectations(t)

		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, "")
		// when
		err := svc.Update(context.TODO(), "id", model.RuntimeInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func TestService_GetByFiltersGlobal(t *testing.T) {
	// given
	testErr := errors.New("Test error")
	filters := []*labelfilter.LabelFilter{
		&labelfilter.LabelFilter{Key: "test-key", Query: str.Ptr("test-filter")},
	}
	testRuntime := &model.Runtime{
		ID:     "test-id",
		Name:   "test-runtime",
		Tenant: "test-tenant-id",
	}
	ctx := context.TODO()

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.RuntimeRepository
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFiltersGlobal", ctx, filters).Return(testRuntime, nil).Once()
				return repo
			},

			ExpectedErrMessage: "",
		},
		{
			Name: "Fails on repository error",
			RepositoryFn: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("GetByFiltersGlobal", ctx, filters).Return(nil, testErr).Once()
				return repo
			},

			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			labelRepository := &automock.LabelRepository{}
			labelUpsertService := &automock.LabelUpsertService{}
			scenariosService := &automock.ScenariosService{}
			scenarioAssignmentEngine := &automock.ScenarioAssignmentEngine{}
			uidSvc := &automock.UIDService{}
			svc := runtime.NewService(repo, labelRepository, scenariosService, labelUpsertService, uidSvc, scenarioAssignmentEngine, ".*_defaultEventing$")

			// when
			actualRuntime, err := svc.GetByFiltersGlobal(ctx, filters)
			// then
			if testCase.ExpectedErrMessage == "" {
				require.Equal(t, testRuntime, actualRuntime)
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
			labelRepository.AssertExpectations(t)
			labelUpsertService.AssertExpectations(t)
			scenariosService.AssertExpectations(t)
			scenarioAssignmentEngine.AssertExpectations(t)
			uidSvc.AssertExpectations(t)

		})
	}

	t.Run("Returns error on loading tenant", func(t *testing.T) {
		// given
		svc := runtime.NewService(nil, nil, nil, nil, nil, nil, "")
		// when
		err := svc.Update(context.TODO(), "id", model.RuntimeInput{})
		// then
		require.Error(t, err)
		assert.EqualError(t, err, "while loading tenant from context: cannot read tenant from context")
	})
}

func contextThatHasTenant(expectedTenant string) interface{} {
	return mock.MatchedBy(func(actual context.Context) bool {
		actualTenant, err := tenant.LoadFromContext(actual)
		if err != nil {
			return false
		}
		return actualTenant == expectedTenant
	})
}
