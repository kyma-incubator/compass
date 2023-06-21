package model_test

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApplicationTemplateVersionInput_ToApplicationTemplateVersion(t *testing.T) {
	// GIVEN
	version := "2306"
	title := "Release 2306"
	releaseDate := "2023-06-21T06:42:08+00:00"
	correlationIDs := json.RawMessage("[]")
	testID := "id-1"
	appTemplateID := "id-2"

	testCases := []struct {
		Name     string
		Input    *model.ApplicationTemplateVersionInput
		Expected model.ApplicationTemplateVersion
	}{
		{
			Name: "All properties given",
			Input: &model.ApplicationTemplateVersionInput{
				Version:        version,
				Title:          &title,
				ReleaseDate:    &releaseDate,
				CorrelationIDs: correlationIDs,
			},
			Expected: model.ApplicationTemplateVersion{
				ID:                    testID,
				Version:               version,
				Title:                 &title,
				ReleaseDate:           &releaseDate,
				CorrelationIDs:        correlationIDs,
				ApplicationTemplateID: appTemplateID,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: model.ApplicationTemplateVersion{},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToApplicationTemplateVersion(testID, appTemplateID)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
