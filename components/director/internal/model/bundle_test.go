package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBundleCreateInput_ToBundle(t *testing.T) {
	// GIVEN
	id := "foo"
	appID := "bar"
	desc := "Sample"
	name := "sample"

	testCases := []struct {
		Name     string
		Input    *model.BundleCreateInput
		Expected *model.Bundle
	}{
		{

			Name: "All properties given",
			Input: &model.BundleCreateInput{
				Name:        name,
				Description: &desc,
			},
			Expected: &model.Bundle{
				ApplicationID: appID,
				Name:          name,
				Description:   &desc,
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

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result := testCase.Input.ToBundle(id, appID, 0)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
