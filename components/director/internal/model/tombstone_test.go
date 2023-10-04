package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestTombstoneInput_ToTombstone(t *testing.T) {
	// GIVEN
	id := "test"
	ordID := "foo"
	appID := "bar"
	appTemplateVersionID := "nar"
	removalDate := "Sample"
	description := "desc"

	testCases := []struct {
		Name         string
		Input        *model.TombstoneInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.Tombstone
	}{
		{
			Name: "All properties given for App",
			Input: &model.TombstoneInput{
				OrdID:       ordID,
				RemovalDate: removalDate,
				Description: &description,
			},
			Expected: &model.Tombstone{
				ID:            id,
				OrdID:         ordID,
				ApplicationID: &appID,
				RemovalDate:   removalDate,
				Description:   &description,
			},
			ResourceType: resource.Application,
			ResourceID:   appID,
		},
		{
			Name: "All properties given for App Template Version",
			Input: &model.TombstoneInput{
				OrdID:       ordID,
				RemovalDate: removalDate,
				Description: &description,
			},
			Expected: &model.Tombstone{
				ID:                           id,
				OrdID:                        ordID,
				ApplicationTemplateVersionID: &appTemplateVersionID,
				RemovalDate:                  removalDate,
				Description:                  &description,
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
			result := testCase.Input.ToTombstone(id, testCase.ResourceType, testCase.ResourceID)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
