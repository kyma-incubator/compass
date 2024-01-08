package model_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestVendorInput_ToVendor(t *testing.T) {
	// GIVEN
	id := "test"
	ordID := "foo"
	appID := "bar"
	appTemplateID := "nar"
	partners := json.RawMessage(`["microsoft:vendor:Microsoft:"]`)
	name := "sample"
	labels := json.RawMessage("{}")

	testCases := []struct {
		Name         string
		Input        *model.VendorInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.Vendor
	}{
		{
			Name: "All properties given for App",
			Input: &model.VendorInput{
				OrdID:    ordID,
				Title:    name,
				Partners: partners,
				Labels:   labels,
			},
			Expected: &model.Vendor{
				ID:            id,
				OrdID:         ordID,
				ApplicationID: &appID,
				Title:         name,
				Partners:      partners,
				Labels:        labels,
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
		},
		{
			Name: "All properties given for App Template Version",
			Input: &model.VendorInput{
				OrdID:    ordID,
				Title:    name,
				Partners: partners,
				Labels:   labels,
			},
			Expected: &model.Vendor{
				ID:                           id,
				OrdID:                        ordID,
				ApplicationTemplateVersionID: &appTemplateID,
				Title:                        name,
				Partners:                     partners,
				Labels:                       labels,
			},
			ResourceType: resource.ApplicationTemplateVersion,
			ResourceID:   appTemplateID,
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
			result := testCase.Input.ToVendor(id, testCase.ResourceType, testCase.ResourceID)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
