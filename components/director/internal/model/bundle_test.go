package model_test

import (
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBundleCreateInput_ToBundle(t *testing.T) {
	// GIVEN
	id := "foo"
	appID := "bar"
	appTemplateVersionID := "naz"
	desc := "Sample"
	name := "sample"

	testCases := []struct {
		Name         string
		Input        *model.BundleCreateInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.Bundle
	}{
		{

			Name: "All properties given for App",
			Input: &model.BundleCreateInput{
				Name:        name,
				Description: &desc,
			},
			Expected: &model.Bundle{
				ApplicationID: &appID,
				Name:          name,
				Description:   &desc,
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
			Input: &model.BundleCreateInput{
				Name:        name,
				Description: &desc,
			},
			Expected: &model.Bundle{
				ApplicationTemplateVersionID: &appTemplateVersionID,
				Name:                         name,
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
			result := testCase.Input.ToBundle(id, testCase.ResourceType, testCase.ResourceID, 0)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
