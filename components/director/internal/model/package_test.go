package model_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestPackageInput_ToPackage(t *testing.T) {
	// given
	id := "foo"
	appID := "bar"
	vendor := "Sample"
	name := "sample"
	labels := json.RawMessage("{}")

	testCases := []struct {
		Name     string
		Input    *model.PackageInput
		Expected *model.Package
	}{
		{
			Name: "All properties given",
			Input: &model.PackageInput{
				Title:  name,
				Vendor: &vendor,
				Labels: labels,
			},
			Expected: &model.Package{
				ID:            id,
				ApplicationID: appID,
				Title:         name,
				Vendor:        &vendor,
				Labels:        labels,
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
			result := testCase.Input.ToPackage(id, appID, 0)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
