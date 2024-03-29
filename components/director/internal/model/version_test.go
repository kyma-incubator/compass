package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestVersionInput_ToVersion(t *testing.T) {
	// GIVEN
	value := "foo"
	deprecated := true
	deprecatedSince := "bar"
	forRemoval := false

	testCases := []struct {
		Name     string
		Input    *model.VersionInput
		Expected *model.Version
	}{
		{
			Name: "All properties given",
			Input: &model.VersionInput{
				Value:           value,
				Deprecated:      &deprecated,
				DeprecatedSince: &deprecatedSince,
				ForRemoval:      &forRemoval,
			},
			Expected: &model.Version{
				Value:           value,
				Deprecated:      &deprecated,
				DeprecatedSince: &deprecatedSince,
				ForRemoval:      &forRemoval,
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
			result := testCase.Input.ToVersion()

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
