package operators_test

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/stretchr/testify/assert"
)

func TestConstraintOperators_ConfigMutator(t *testing.T) {
	cfg := "{\"config\": {\"description\": \"dummy description\", \"credentials\": {\"url\":\"test.test\", \"mode\":\"SYNC\"}}}"
	cfg2 := "{\"config2\": {\"description\": \"dummy description\"}}"
	state := string(model.ConfigPendingAssignmentState)

	testCases := []struct {
		Name           string
		InputFa        *model.FormationAssignment
		NewState       *string
		NewConfig      *string
		ExpectedState  string
		ExpectedConfig json.RawMessage
	}{
		{
			Name:           "Update State and Config",
			InputFa:        fixFormationAssignmentWithConfig(json.RawMessage(cfg)),
			NewState:       &state,
			NewConfig:      &cfg2,
			ExpectedState:  string(model.ConfigPendingAssignmentState),
			ExpectedConfig: json.RawMessage(cfg2),
		},
		{
			Name:           "Update State only",
			InputFa:        fixFormationAssignmentWithConfig(json.RawMessage(cfg)),
			NewState:       &state,
			ExpectedState:  string(model.ConfigPendingAssignmentState),
			ExpectedConfig: json.RawMessage(cfg),
		},
		{
			Name:           "Update Config only",
			InputFa:        fixFormationAssignmentWithConfig(json.RawMessage(cfg)),
			NewConfig:      &cfg2,
			ExpectedState:  string(model.ReadyAssignmentState),
			ExpectedConfig: json.RawMessage(cfg2),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN

			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			fa := testCase.InputFa
			input := fixConfigMutatorInput(testCase.InputFa, testCase.NewState, testCase.NewConfig)
			result, err := engine.MutateConfig(ctx, input)

			// THEN
			assert.Equal(t, true, result)
			assert.Equal(t, testCase.ExpectedConfig, fa.Value)
			assert.Equal(t, testCase.ExpectedState, fa.State)
			assert.NoError(t, err)
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
		assert.Equal(t, "Incompatible input", err.Error())
	})
}
