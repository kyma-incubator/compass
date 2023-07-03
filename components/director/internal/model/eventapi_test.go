package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestEventAPIDefinitionInput_ToEventAPIDefinition(t *testing.T) {
	// GIVEN
	id := "foo"
	bndlID := "bar"
	appID := "baz"
	appTemplateVersionID := "naz"
	desc := "Sample"
	name := "sample"
	group := "sampleGroup"

	testCases := []struct {
		Name         string
		Input        *model.EventDefinitionInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.EventDefinition
	}{
		{
			Name: "All properties given for App",
			Input: &model.EventDefinitionInput{
				Name:        name,
				Description: &desc,
				Group:       &group,
			},
			Expected: &model.EventDefinition{
				ApplicationID: &appID,
				Name:          name,
				Description:   &desc,
				Group:         &group,
				PackageID:     &bndlID,
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
				},
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
		},
		{
			Name: "All properties given for App Template Version",
			Input: &model.EventDefinitionInput{
				Name:        name,
				Description: &desc,
				Group:       &group,
			},
			Expected: &model.EventDefinition{
				ApplicationTemplateVersionID: &appTemplateVersionID,
				Name:                         name,
				Description:                  &desc,
				Group:                        &group,
				PackageID:                    &bndlID,
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
				},
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
			result := testCase.Input.ToEventDefinition(id, testCase.ResourceType, testCase.ResourceID, &bndlID, 0)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
