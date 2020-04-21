package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestDocumentInput_ToDocument(t *testing.T) {
	// given
	packageID := "foo"
	id := "bar"
	tenant := "baz"
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
				PackageID:   packageID,
				ID:          id,
				Tenant:      tenant,
				Title:       title,
				DisplayName: displayName,
				Description: description,
				Format:      model.DocumentFormatMarkdown,
				Kind:        &kind,
				Data:        &data,
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
				PackageID:   packageID,
				ID:          id,
				Tenant:      tenant,
				Title:       title,
				DisplayName: displayName,
				Description: description,
				Format:      model.DocumentFormatMarkdown,
				Kind:        &kind,
				Data:        &data,
			},
		},
		{
			Name:  "Empty",
			Input: &model.DocumentInput{},
			Expected: &model.Document{
				PackageID: packageID,
				ID:        id,
				Tenant:    tenant,
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
			result := testCase.Input.ToDocumentWithinPackage(id, tenant, packageID)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
