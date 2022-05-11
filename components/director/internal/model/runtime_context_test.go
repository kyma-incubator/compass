package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeContextInput_ToRuntimeContext(t *testing.T) {
	// GIVEN
	id := "foo"
	runtimeID := "bar"
	key := "key"
	val := "val"
	testCases := []struct {
		Name     string
		Input    *model.RuntimeContextInput
		Expected *model.RuntimeContext
	}{
		{
			Name: "All properties given",
			Input: &model.RuntimeContextInput{
				Key:       key,
				Value:     val,
				RuntimeID: runtimeID,
			},
			Expected: &model.RuntimeContext{
				ID:        id,
				RuntimeID: runtimeID,
				Key:       key,
				Value:     val,
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
			result := testCase.Input.ToRuntimeContext(id)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
