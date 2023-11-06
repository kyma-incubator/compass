package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestAspectInput_ToAspect(t *testing.T) {
	id := "foo"
	appID := "test-app-id"
	appTemplateVersionID := "test-app-tmpl-id"
	name := "Test Name"
	desc := "Test Description"
	integrationDependencyID := "test-integration-dependency-id"

	testCases := []struct {
		Name         string
		Input        *model.AspectInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.Aspect
	}{
		{
			Name: "All properties given for App",
			Input: &model.AspectInput{
				Title:       name,
				Description: &desc,
			},
			Expected: &model.Aspect{
				ApplicationID:           &appID,
				IntegrationDependencyID: integrationDependencyID,
				Title:                   name,
				Description:             &desc,
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
			Input: &model.AspectInput{
				Title:       name,
				Description: &desc,
			},
			Expected: &model.Aspect{
				ApplicationTemplateVersionID: &appTemplateVersionID,
				IntegrationDependencyID:      integrationDependencyID,
				Title:                        name,
				Description:                  &desc,
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
			result := testCase.Input.ToAspect(id, testCase.ResourceType, testCase.ResourceID, integrationDependencyID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
