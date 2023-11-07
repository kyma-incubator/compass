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
	packageID := "pack"
	labels := json.RawMessage("{}")

	testCases := []struct {
		Name         string
		Input        *model.EntityTypeInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.EntityType
	}{
		{
			Name: "All properties given for App",
			Input: &model.EntityTypeInput{
				Labels: labels,
			},
			Expected: &model.EntityType{
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
				},
				ApplicationID: &appID,
				PackageID:     packageID,
				Labels:        labels,
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
		},
		{
			Name: "All properties given for App Template Version",
			Input: &model.EntityTypeInput{
				Labels: labels,
			},
			Expected: &model.EntityType{
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
				},
				ApplicationTemplateVersionID: &appTemplateVersionID,
				PackageID:                    packageID,
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
			result := testCase.Input.ToEntityType(id, testCase.ResourceType, testCase.ResourceID, packageID, 0)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
