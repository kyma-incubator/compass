package operators_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_RedirectNotification(t *testing.T) {
	testCases := []struct {
		Name                  string
		Input                 operators.OperatorInput
		DestinationSvc        func() *automock.DestinationService
		DestinationCreatorSvc func() *automock.DestinationCreatorService
		ExpectedResult        bool
		ExpectedErrorMsg      string
	}{
		{
			Name:           "Successfully changed URL and URL Template",
			Input:          fixRedirectNotificationOperatorInput(true),
			ExpectedResult: true,
		},
		{
			Name:           "Success(no-op) when the operator condition is not met",
			Input:          fixRedirectNotificationOperatorInput(false),
			ExpectedResult: true,
		},
		{
			Name:             "Error when parsing operator input",
			Input:            "wrong input",
			ExpectedErrorMsg: "Incompatible input for operator:",
		},
		{
			Name:             "Error when retrieving webhook pointer fails",
			Input:            inputWithoutWebhookMemoryAddress,
			ExpectedErrorMsg: "The webhook memory address cannot be 0",
		},
	}

	for _, ts := range testCases {
		t.Run(ts.Name, func(t *testing.T) {
			// GIVEN
			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			result, err := engine.RedirectNotification(ctx, ts.Input)

			// THEN
			if ts.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), ts.ExpectedErrorMsg)
			} else {
				assert.Equal(t, ts.ExpectedResult, result)
				assert.NoError(t, err)
			}
		})
	}
}
