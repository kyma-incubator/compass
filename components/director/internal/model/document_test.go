package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDocumentInput_ToDocument(t *testing.T) {
	// GIVEN
	bundleID := "foo"
	appID := "appID"
	id := "bar"
	kind := "fookind"
	data := "foodata"
	displayName := "foodisplay"
	description := "foodescription"
	title := "footitle"
	testCases := []struct {
		Name     string
		Input    *model.DocumentInput
		Expected *model.Document
	}{
		{
			Name: "All properties given",
			Input: &model.DocumentInput{
				Title:       title,
				DisplayName: displayName,
				Description: description,
				Format:      model.DocumentFormatMarkdown,
				Kind:        &kind,
				Data:        &data,
				FetchRequest: &model.FetchRequestInput{
					URL: "foo.bar",
				},
			},
			Expected: &model.Document{
				BundleID:    bundleID,
				AppID:       appID,
				Title:       title,
				DisplayName: displayName,
				Description: description,
				Format:      model.DocumentFormatMarkdown,
				Kind:        &kind,
				Data:        &data,
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
				},
			},
		},
		{
			Name: "No FetchRequest",
			Input: &model.DocumentInput{
				Title:        title,
				DisplayName:  displayName,
				Description:  description,
				Format:       model.DocumentFormatMarkdown,
				Kind:         &kind,
				Data:         &data,
				FetchRequest: nil,
			},
			Expected: &model.Document{
				BundleID:    bundleID,
				AppID:       appID,
				Title:       title,
				DisplayName: displayName,
				Description: description,
				Format:      model.DocumentFormatMarkdown,
				Kind:        &kind,
				Data:        &data,
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
				},
			},
		},
		{
			Name:  "Empty",
			Input: &model.DocumentInput{},
			Expected: &model.Document{
				BundleID: bundleID,
				AppID:    appID,
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
				},
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToDocumentWithinBundle(id, bundleID, appID)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
