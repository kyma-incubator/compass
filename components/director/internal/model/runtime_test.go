package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeInput_ToRuntime(t *testing.T) {
	// GIVEN
	desc := "Sample"
	id := "foo"
	creationTimestamp := time.Now()
	conditionTimestamp := time.Now()
	conditionStatus := model.RuntimeStatusConditionConnected
	testCases := []struct {
		Name     string
		Input    *model.RuntimeInput
		Expected *model.Runtime
	}{
		{
			Name: "All properties given",
			Input: &model.RuntimeInput{
				Name:        "Foo",
				Description: &desc,
				Labels: map[string]interface{}{
					"test": []string{"val", "val2"},
				},
				StatusCondition: &conditionStatus,
			},
			Expected: &model.Runtime{
				Name:        "Foo",
				ID:          id,
				Description: &desc,
				Status: &model.RuntimeStatus{
					Condition: conditionStatus,
					Timestamp: conditionTimestamp,
				},
				CreationTimestamp: creationTimestamp,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToRuntime(id, creationTimestamp, conditionTimestamp)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
