package operators_test

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/mock"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/stretchr/testify/assert"
)

func TestConstraintOperators_ConfigMutator(t *testing.T) {
	cfg := "{\"config\": {\"description\": \"dummy description\", \"credentials\": {\"url\":\"test.test\", \"mode\":\"SYNC\"}}}"
	cfg2 := "{\"config2\": {\"description\": \"dummy description\"}}"
	subtype := "someType"
	sourceID := "ID"
	otherSubtype := "otherType"
	state := string(model.ConfigPendingAssignmentState)
	faWithConfig := fixFormationAssignmentWithConfig(json.RawMessage(cfg))
	faWithConfig.SourceType = model.FormationAssignmentTypeApplication
	faWithConfig.Source = sourceID

	testCases := []struct {
		Name                  string
		InputFa               *model.FormationAssignment
		NewState              *string
		NewConfig             *string
		OnlyForSourceSubtypes []string
		LabelSvcFn            func() *automock.LabelService
		ExpectedState         string
		ExpectedConfig        json.RawMessage
		ExpectedErrorMsg      string
	}{
		{
			Name:           "Update State and Config",
			InputFa:        faWithConfig.Clone(),
			NewState:       &state,
			NewConfig:      &cfg2,
			ExpectedState:  string(model.ConfigPendingAssignmentState),
			ExpectedConfig: json.RawMessage(cfg2),
		},
		{
			Name:           "Update State only",
			InputFa:        faWithConfig.Clone(),
			NewState:       &state,
			ExpectedState:  string(model.ConfigPendingAssignmentState),
			ExpectedConfig: json.RawMessage(cfg),
		},
		{
			Name:           "Update Config only",
			InputFa:        faWithConfig.Clone(),
			NewConfig:      &cfg2,
			ExpectedState:  string(model.ReadyAssignmentState),
			ExpectedConfig: json.RawMessage(cfg2),
		},
		{
			Name:                  "Update State and Config when source subtype is supported",
			InputFa:               faWithConfig.Clone(),
			NewState:              &state,
			NewConfig:             &cfg2,
			OnlyForSourceSubtypes: []string{subtype},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceID, applicationTypeLabel).Return(&model.Label{Value: subtype}, nil).Once()
				return svc
			},
			ExpectedState:  string(model.ConfigPendingAssignmentState),
			ExpectedConfig: json.RawMessage(cfg2),
		},
		{
			Name:                  "No op when source subtype is not supported",
			InputFa:               faWithConfig.Clone(),
			NewState:              &state,
			NewConfig:             &cfg2,
			OnlyForSourceSubtypes: []string{otherSubtype},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceID, applicationTypeLabel).Return(&model.Label{Value: subtype}, nil).Once()
				return svc
			},
			ExpectedState:  string(model.ReadyAssignmentState),
			ExpectedConfig: json.RawMessage(cfg),
		},
		{
			Name:                  "Error while getting label",
			InputFa:               faWithConfig.Clone(),
			NewState:              &state,
			NewConfig:             &cfg2,
			OnlyForSourceSubtypes: []string{otherSubtype},
			LabelSvcFn: func() *automock.LabelService {
				svc := &automock.LabelService{}
				svc.On("GetByKey", ctx, testTenantID, model.ApplicationLabelableObject, sourceID, applicationTypeLabel).Return(nil, testErr).Once()
				return svc
			},
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
			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, labelService, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			fa := testCase.InputFa
			input := fixConfigMutatorInput(testCase.InputFa, testCase.NewState, testCase.NewConfig, testCase.OnlyForSourceSubtypes)
			result, err := engine.MutateConfig(ctx, input)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
				assert.Equal(t, false, result)
			} else {
				assert.Equal(t, true, result)
				assert.Equal(t, testCase.ExpectedConfig, fa.Value)
				assert.Equal(t, testCase.ExpectedState, fa.State)
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, labelService)
		})
	}

	t.Run("Error when incorrect input is provided", func(t *testing.T) {
		// GIVEN

		engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		input := &formationconstraintpkg.DestinationCreatorInput{}
		result, err := engine.MutateConfig(ctx, input)

		// THEN
		assert.Equal(t, false, result)
		assert.Equal(t, "Incompatible input for operator: ConfigMutator", err.Error())
	})
}
