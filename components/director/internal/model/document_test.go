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
				FetchRequest:  nil,
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
			result := testCase.Input.ToDocument(id, applicationID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
