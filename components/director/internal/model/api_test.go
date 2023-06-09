package model_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestAPIDefinitionInput_ToAPIDefinitionWithBundleID(t *testing.T) {
	// GIVEN
	id := "foo"
	appID := "baz"
	appTemplateVersionID := "naz"
	desc := "Sample"
	name := "sample"
	targetURL := "https://foo.bar"
	group := "sampleGroup"

	testCases := []struct {
		Name         string
		Input        *model.APIDefinitionInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.APIDefinition
	}{
		{
			Name: "All properties given for App",
			Input: &model.APIDefinitionInput{
				Name:        name,
				Description: &desc,
				TargetURLs:  api.ConvertTargetURLToJSONArray(targetURL),
				Group:       &group,
			},
			Expected: &model.APIDefinition{
				ApplicationID: &appID,
				Name:          name,
				Description:   &desc,
				TargetURLs:    api.ConvertTargetURLToJSONArray(targetURL),
				Group:         &group,
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
			Input: &model.APIDefinitionInput{
				Name:        name,
				Description: &desc,
				TargetURLs:  api.ConvertTargetURLToJSONArray(targetURL),
				Group:       &group,
			},
			Expected: &model.APIDefinition{
				ApplicationTemplateVersionID: &appTemplateVersionID,
				Name:                         name,
				Description:                  &desc,
				TargetURLs:                   api.ConvertTargetURLToJSONArray(targetURL),
				Group:                        &group,
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
			result := testCase.Input.ToAPIDefinition(id, testCase.ResourceType, testCase.ResourceID, nil, 0)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
