package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDocumentInput_ToDocument(t *testing.T) {
	// given
	applicationID := "foo"
	id := "bar"
	fetchRequestID := "frID"
	kind := "fookind"
	data := "foodata"
	displayName := "foodisplay"
	description := "foodescription"
	title := "footitle"
	testCases := []struct {
		Name     string
		Input    *model.DocumentInput
		FetchRequestID *string
		Expected *model.Document
	}{
		{
			FetchRequestID: &fetchRequestID,
			Name: "All properties given",
			Input: &model.DocumentInput{
				Title:        title,
				DisplayName:  displayName,
				Description:  description,
				Format:       model.DocumentFormatMarkdown,
				Kind:         &kind,
				Data:         &data,
				FetchRequest: &model.FetchRequestInput{
					URL: "foo.bar",
				},
			},
			Expected: &model.Document{
				ApplicationID: applicationID,
				ID:            id,
				Title:         title,
				DisplayName:   displayName,
				Description:   description,
				Format:        model.DocumentFormatMarkdown,
				Kind:          &kind,
				Data:          &data,
				FetchRequestID: &fetchRequestID,
			},
		},
		{
			Name: "No FetchRequest",
			FetchRequestID: nil,
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
				ApplicationID: applicationID,
				ID:            id,
				Title:         title,
				DisplayName:   displayName,
				Description:   description,
				Format:        model.DocumentFormatMarkdown,
				Kind:          &kind,
				Data:          &data,
				FetchRequestID: nil,
			},
		},
		{
			Name:  "Empty",
			Input: &model.DocumentInput{},
			Expected: &model.Document{
				ApplicationID: applicationID,
				ID:            id,
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

			// when
			result := testCase.Input.ToDocument(id, applicationID, testCase.FetchRequestID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
