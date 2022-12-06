package scenariogroups_test

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenariogroups"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadFromContext(t *testing.T) {
	value := []string{"foo"}

	testCases := []struct {
		Name    string
		Context context.Context

		ExpectedResult []string
	}{
		{
			Name:           "Success",
			Context:        context.WithValue(context.TODO(), scenariogroups.ScenarioGroupsContextKey, value),
			ExpectedResult: value,
		},
		{
			Name:           "Nil value",
			Context:        context.TODO(),
			ExpectedResult: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := scenariogroups.LoadFromContext(testCase.Context)

			assert.Equal(t, testCase.ExpectedResult, result)
		})
	}
}

func TestSaveToLoadFromContext(t *testing.T) {
	// GIVEN
	value := []string{"foo"}
	ctx := context.TODO()

	// WHEN
	result := scenariogroups.SaveToContext(ctx, value)

	// then
	assert.Equal(t, value, result.Value(scenariogroups.ScenarioGroupsContextKey))
}
