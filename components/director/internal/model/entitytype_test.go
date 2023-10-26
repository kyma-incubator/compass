package model_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestEntityTypeInput_ToEntityType(t *testing.T) {
	// GIVEN
	id := "foo"
	appID := "bar"
	appTemplateVersionID := "naz"
	vendor := "Sample"
	name := "sample"
	labels := json.RawMessage("{}")

	testCases := []struct {
		Name         string
		Input        *model.PackageInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.Package
	}{
		{
			Name: "All properties given for App",
			Input: &model.PackageInput{
				Title:  name,
				Vendor: &vendor,
				Labels: labels,
			},
			Expected: &model.Package{
				ID:            id,
				ApplicationID: &appID,
				Title:         name,
				Vendor:        &vendor,
				Labels:        labels,
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
		},
		{
			Name: "All properties given for App Template Version",
			Input: &model.PackageInput{
				Title:  name,
				Vendor: &vendor,
				Labels: labels,
			},
			Expected: &model.Package{
				ID:                           id,
				ApplicationTemplateVersionID: &appTemplateVersionID,
				Title:                        name,
				Vendor:                       &vendor,
				Labels:                       labels,
			},
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateVersionID,
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
			result := testCase.Input.ToPackage(id, testCase.ResourceType, testCase.ResourceID, 0)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
