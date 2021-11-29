package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestTombstoneInput_ToTombstone(t *testing.T) {
	// GIVEN
	id := "test"
	ordID := "foo"
	appID := "bar"
	removalDate := "Sample"

	testCases := []struct {
		Name     string
		Input    *model.TombstoneInput
		Expected *model.Tombstone
	}{
		{
			Name: "All properties given",
			Input: &model.TombstoneInput{
				OrdID:       ordID,
				RemovalDate: removalDate,
			},
			Expected: &model.Tombstone{
				ID:            id,
				OrdID:         ordID,
				ApplicationID: appID,
				RemovalDate:   removalDate,
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
			result := testCase.Input.ToTombstone(id, appID)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
