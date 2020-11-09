package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeContextInput_ToRuntimeContext(t *testing.T) {
	// given
	id := "foo"
	runtimeID := "bar"
	tenant := "sample"
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
				Labels: map[string]interface{}{
					"test": []string{"val", "val2"},
				},
			},
			Expected: &model.RuntimeContext{
				ID:        id,
				RuntimeID: runtimeID,
				Tenant:    tenant,
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
			// when
			result := testCase.Input.ToRuntimeContext(id, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
