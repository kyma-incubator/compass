package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestTombstoneInput_ToTombstone(t *testing.T) {
	// given
	id := "test"
	ordID := "foo"
	appID := "bar"
	removalDate := "Sample"
	tenant := "tenant"

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
				TenantID:      tenant,
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

			// when
			result := testCase.Input.ToTombstone(id, tenant, appID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
