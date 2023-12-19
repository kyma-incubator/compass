package model_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDataProductInput_ToDataProduct(t *testing.T) {
	// GIVEN
	id := "foo"
	appID := "test-app-id"
	appTemplateVersionID := "test-app-tmpl-id"
	name := "Test Data Product Name"
	desc := "Test Description"
	pkgID := "test-pkg-id"

	testCases := []struct {
		Name         string
		Input        *model.DataProductInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.DataProduct
	}{
		{
			Name: "All properties given for App",
			Input: &model.DataProductInput{
				Title:       name,
				Description: &desc,
			},
			Expected: &model.DataProduct{
				ApplicationID: &appID,
				Title:         name,
				Description:   &desc,
				PackageID:     &pkgID,
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
			Input: &model.DataProductInput{
				Title:       name,
				Description: &desc,
			},
			Expected: &model.DataProduct{
				ApplicationTemplateVersionID: &appTemplateVersionID,
				Title:                        name,
				Description:                  &desc,
				PackageID:                    &pkgID,
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
			result := testCase.Input.ToDataProduct(id, testCase.ResourceType, testCase.ResourceID, &pkgID, 0)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
