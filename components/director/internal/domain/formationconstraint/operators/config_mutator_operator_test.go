package operators_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/stretchr/testify/assert"
)

func TestConstraintOperators_ConfigMutator(t *testing.T) {
	cfg := "{\"config2\": {\"description\": \"dummy description\"}}"
	subtype := "someType"
	sourceID := "ID"
	otherSubtype := "otherType"
	state := string(model.ConfigPendingAssignmentState)
	faInInput := fixFormationAssignmentWithState(model.InitialAssignmentState)
	faInInput.SourceType = model.FormationAssignmentTypeApplication
	faInInput.Source = sourceID

	testCases := []struct {
		Name                  string
		InputFa               *model.FormationAssignment
		NewState              *string
		NewConfig             *string
		OnlyForSourceSubtypes []string
		LabelSvcFn            func() *automock.LabelService
		StatusReport          *statusreport.NotificationStatusReport
		ExpectedState         string
		ExpectedConfig        json.RawMessage
		ExpectedErrorMsg      string
	}{
		{
			Name:           "Update State and Config",
			InputFa:        faInInput,
			NewState:       &state,
			NewConfig:      &cfg,
			StatusReport:   fixNotificationStatusReport(),
			ExpectedState:  string(model.ConfigPendingAssignmentState),
			ExpectedConfig: json.RawMessage(cfg),
		},
		{
			Name:          "Update State only",
			InputFa:       faInInput,
			NewState:      &state,
			StatusReport:  fixNotificationStatusReport(),
			ExpectedState: string(model.ConfigPendingAssignmentState),
		},
		{
			Name:           "Update Config only",
			InputFa:        faInInput,
			NewConfig:      &cfg,
			StatusReport:   fixNotificationStatusReport(),
			ExpectedConfig: json.RawMessage(cfg),
		},
		{
			Name:                  "Update State and Config when source subtype is supported",
			InputFa:               faInInput,
			NewState:              &state,
			NewConfig:             &cfg,
			OnlyForSourceSubtypes: []string{subtype},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceID, applicationTypeLabel).Return(&model.Label{Value: subtype}, nil).Once()
				return svc
			},
			StatusReport:   fixNotificationStatusReport(),
			ExpectedState:  string(model.ConfigPendingAssignmentState),
			ExpectedConfig: json.RawMessage(cfg),
		},
		{
			Name:                  "No op when source subtype is not supported",
			InputFa:               faInInput,
			NewState:              &state,
			NewConfig:             &cfg,
			OnlyForSourceSubtypes: []string{otherSubtype},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceID, applicationTypeLabel).Return(&model.Label{Value: subtype}, nil).Once()
				return svc
			},
			StatusReport: fixNotificationStatusReport(),
		},
		{
			Name:                  "Error while getting label",
			InputFa:               faInInput,
			NewState:              &state,
			NewConfig:             &cfg,
			OnlyForSourceSubtypes: []string{otherSubtype},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceID, applicationTypeLabel).Return(nil, testErr).Once()
				return svc
			},
			StatusReport:     fixNotificationStatusReport(),
			ExpectedErrorMsg: "while getting subtype of resource with type:",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN

			labelService := &automock.LabelService{}
			if testCase.LabelSvcFn != nil {
				labelService = testCase.LabelSvcFn()
			}
			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, labelService, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			input := fixConfigMutatorInput(testCase.InputFa, testCase.StatusReport, testCase.NewState, testCase.NewConfig, testCase.OnlyForSourceSubtypes)
			result, err := engine.MutateConfig(ctx, input)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
				assert.Equal(t, false, result)
			} else {
				assert.Equal(t, true, result)
				retrieveNotificationStatusReport, err := operators.RetrieveNotificationStatusReportPointer(ctx, input.NotificationStatusReportMemoryAddress)
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedConfig, retrieveNotificationStatusReport.Configuration)
				assert.Equal(t, testCase.ExpectedState, retrieveNotificationStatusReport.State)
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, labelService)
		})
	}

	t.Run("Error when incorrect input is provided", func(t *testing.T) {
		// GIVEN

		engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		input := &formationconstraintpkg.DestinationCreatorInput{}
		result, err := engine.MutateConfig(ctx, input)

		// THEN
		assert.Equal(t, false, result)
		assert.Equal(t, "Incompatible input for operator: ConfigMutator", err.Error())
	})
}
