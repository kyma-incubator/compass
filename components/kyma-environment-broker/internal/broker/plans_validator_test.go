package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlansSchemaValidatorErrors(t *testing.T) {
	tests := map[string]struct {
		againstPlans []string
		inputJSON    string
		expErr       string
	}{
		"missing name, not valid components list": {
			againstPlans: []string{GcpPlanID, AzurePlanID},
			inputJSON:    `{"components": ["wrong component name"]}`,
			expErr:       `(root): name is required, components.0: components.0 must be one of the following: "Kiali", "Jaeger"`,
		},
		"not valid node count": {
			againstPlans: []string{AzurePlanID},
			inputJSON:    `{"name": "wrong-node", "nodeCount": 123123}`,
			expErr:       "nodeCount: Must be less than or equal to 20",
		},
		"missing name, not valid machine type": {
			againstPlans: []string{AzurePlanID},
			inputJSON:    `{"name": "wrong-machType", "machineType": "WrongName"}`,
			expErr:       `machineType: machineType must be one of the following: "Standard_D8_v3"`,
		},
	}
	for tN, tC := range tests {
		t.Run(tN, func(t *testing.T) {
			// given
			validator, err := NewPlansSchemaValidator()
			require.NoError(t, err)

			for _, id := range tC.againstPlans {
				// when
				result, err := validator[id].ValidateString(tC.inputJSON)
				require.NoError(t, err)

				// then
				assert.False(t, result.Valid)
				assert.EqualError(t, result.Error, tC.expErr)
			}
		})
	}
}

func TestNewPlansSchemaValidatorSuccess(t *testing.T) {
	// given
	validJSON := `{"name": "only-name-is-required"}`

	validator, err := NewPlansSchemaValidator()
	require.NoError(t, err)

	for _, id := range []string{GcpPlanID, AzurePlanID} {
		// when
		result, err := validator[id].ValidateString(validJSON)
		require.NoError(t, err)

		// then
		assert.True(t, result.Valid)

		// Currently there is a "bug" in /kyma-incubator/compass/components/director/pkg/jsonschema/validator.go:84
		// which missing executing method `.ErrorOrNil()` so we cannot use `assert.NoError`
		assert.Nil(t, result.Error)
	}
}
