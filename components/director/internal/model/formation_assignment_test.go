package model_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/require"
)

func TestFormationAssignmentInput_ToModel(t *testing.T) {
	// GIVEN
	id := "id"
	tenantID := "tenant-id"

	faInput := &model.FormationAssignmentInput{
		FormationID: "formation-id",
		Source:      "source",
		SourceType:  "source-type",
		Target:      "target",
		TargetType:  "target-type",
		State:       "state",
		Value:       json.RawMessage(`{"testKey":"testValue"}`),
	}

	fa := &model.FormationAssignment{
		ID:          id,
		FormationID: "formation-id",
		TenantID:    tenantID,
		Source:      "source",
		SourceType:  "source-type",
		Target:      "target",
		TargetType:  "target-type",
		State:       "state",
		Value:       json.RawMessage(`{"testKey":"testValue"}`),
	}

	testCases := []struct {
		Name     string
		Input    *model.FormationAssignmentInput
		Expected *model.FormationAssignment
	}{
		{
			Name:     "Success",
			Input:    faInput,
			Expected: fa,
		},
		{
			Name:  "Empty",
			Input: &model.FormationAssignmentInput{},
			Expected: &model.FormationAssignment{
				ID:       id,
				TenantID: tenantID,
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
			result := testCase.Input.ToModel(id, tenantID)

			// THEN
			require.Equal(t, testCase.Expected, result)
		})
	}
}
