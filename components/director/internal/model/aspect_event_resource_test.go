package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func TestAspectEventResourceInput_ToAspectEventResource(t *testing.T) {
	id := "foo"
	appID := "test-app-id"
	appTemplateVersionID := "test-app-tmpl-id"
	ordID := "Test ord id"
	minVersion := "1.0.0"
	aspectID := "test-aspect-id"

	testCases := []struct {
		Name         string
		Input        *model.AspectEventResourceInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.AspectEventResource
	}{
		{
			Name: "All properties given for App",
			Input: &model.AspectEventResourceInput{
				OrdID:      ordID,
				MinVersion: &minVersion,
			},
			Expected: &model.AspectEventResource{
				ApplicationID: &appID,
				AspectID:      aspectID,
				OrdID:         ordID,
				MinVersion:    &minVersion,
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
			Input: &model.AspectEventResourceInput{
				OrdID:      ordID,
				MinVersion: &minVersion,
			},
			Expected: &model.AspectEventResource{
				ApplicationTemplateVersionID: &appTemplateVersionID,
				AspectID:                     aspectID,
				OrdID:                        ordID,
				MinVersion:                   &minVersion,
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
			result := testCase.Input.ToAspectEventResource(id, testCase.ResourceType, testCase.ResourceID, aspectID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
