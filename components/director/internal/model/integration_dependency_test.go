package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestIntegrationDependencyInput_ToIntegrationDependency(t *testing.T) {
	// GIVEN
	id := "foo"
	appID := "test-app-id"
	appTemplateVersionID := "test-app-tmpl-id"
	name := "Test Name"
	desc := "Test Description"
	pkgId := "test-pkg-id"

	testCases := []struct {
		Name         string
		Input        *model.IntegrationDependencyInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.IntegrationDependency
	}{
		{
			Name: "All properties given for App",
			Input: &model.IntegrationDependencyInput{
				Title:       name,
				Description: &desc,
			},
			Expected: &model.IntegrationDependency{
				ApplicationID: &appID,
				Title:         name,
				Description:   &desc,
				PackageID:     &pkgId,
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
			Input: &model.IntegrationDependencyInput{
				Title:       name,
				Description: &desc,
			},
			Expected: &model.IntegrationDependency{
				ApplicationTemplateVersionID: &appTemplateVersionID,
				Title:                        name,
				Description:                  &desc,
				PackageID:                    &pkgId,
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
			result := testCase.Input.ToIntegrationDependency(id, testCase.ResourceType, testCase.ResourceID, &pkgId, 0)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
