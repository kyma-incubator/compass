package model_test

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestProductInput_ToProduct(t *testing.T) {
	// GIVEN
	id := "test"
	ordID := "foo"
	appID := "bar"
	appTemplateVersionID := "vaz"
	vendor := "Sample"
	name := "sample"
	labels := json.RawMessage("{}")

	testCases := []struct {
		Name         string
		Input        *model.ProductInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.Product
	}{
		{
			Name: "All properties given for App",
			Input: &model.ProductInput{
				OrdID:  ordID,
				Title:  name,
				Vendor: vendor,
				Labels: labels,
			},
			Expected: &model.Product{
				ID:            id,
				OrdID:         ordID,
				ApplicationID: &appID,
				Title:         name,
				Vendor:        vendor,
				Labels:        labels,
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
		},
		{
			Name: "All properties given for App Template Version",
			Input: &model.ProductInput{
				OrdID:  ordID,
				Title:  name,
				Vendor: vendor,
				Labels: labels,
			},
			Expected: &model.Product{
				ID:                           id,
				OrdID:                        ordID,
				ApplicationTemplateVersionID: &appTemplateVersionID,
				Title:                        name,
				Vendor:                       vendor,
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
			result := testCase.Input.ToProduct(id, testCase.ResourceType, testCase.ResourceID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
