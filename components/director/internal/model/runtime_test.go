package model_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeRegisterInput_ToRuntime(t *testing.T) {
	// GIVEN
	desc := "Sample"
	id := "foo"
	creationTimestamp := time.Now()
	conditionTimestamp := time.Now()
	conditionStatus := model.RuntimeStatusConditionConnected
	testCases := []struct {
		Name     string
		Input    *model.RuntimeRegisterInput
		Expected *model.Runtime
	}{
		{
			Name: "All properties given",
			Input: &model.RuntimeRegisterInput{
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

func TestRuntimeUpdateInput_ToRuntime(t *testing.T) {
	// GIVEN
	desc := "Sample"
	id := "foo"
	creationTimestamp := time.Now()
	conditionTimestamp := time.Now()
	conditionStatus := model.RuntimeStatusConditionConnected
	testCases := []struct {
		Name     string
		Input    *model.RuntimeUpdateInput
		Expected *model.Runtime
	}{
		{
			Name: "All properties given",
			Input: &model.RuntimeUpdateInput{
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
