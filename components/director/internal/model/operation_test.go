package model_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestOperationInput_ToOperation(t *testing.T) {
	// GIVEN
	id := "foo"

	testCases := []struct {
		Name     string
		Input    *model.OperationInput
		Expected *model.Operation
	}{
		{
			Name: "All properties given",
			Input: &model.OperationInput{
				OpType:     "OP_TYPE",
				Status:     "OP_STATUS",
				Data:       json.RawMessage("{}"),
				Error:      json.RawMessage("{}"),
				Priority:   1,
				CreatedAt:  &time.Time{},
				FinishedAt: &time.Time{},
			},
			Expected: &model.Operation{
				ID:         id,
				OpType:     "OP_TYPE",
				Status:     "OP_STATUS",
				Data:       json.RawMessage("{}"),
				Error:      json.RawMessage("{}"),
				Priority:   1,
				CreatedAt:  &time.Time{},
				FinishedAt: &time.Time{},
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToOperation(id)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
