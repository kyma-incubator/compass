package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestEventAPIDefinitionInput_ToEventAPIDefinition(t *testing.T) {
	// given
	id := "foo"
	bundleID := "bar"
	desc := "Sample"
	name := "sample"
	group := "sampleGroup"
	tenant := "tenant"

	testCases := []struct {
		Name     string
		Input    *model.EventDefinitionInput
		Expected *model.EventDefinition
	}{
		{
			Name: "All properties given",
			Input: &model.EventDefinitionInput{
				Title:       name,
				Description: &desc,
				Group:       &group,
			},
			Expected: &model.EventDefinition{
				ID:          id,
				Tenant:      tenant,
				BundleID:    bundleID,
				Title:       name,
				Description: &desc,
				Group:       &group,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {

			// when
			result := testCase.Input.ToEventDefinitionWithinBundle(id, bundleID, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestEventAPIDefinitionInput_ToEventAPISpec(t *testing.T) {
	// given
	data := "bar"
	format := model.SpecFormat("Sample")
	specType := model.EventSpecType("sample")

	testCases := []struct {
		Name     string
		Input    *model.EventSpecInput
		Expected *model.EventSpec
	}{
		{
			Name: "All properties given",
			Input: &model.EventSpecInput{
				Data:          &data,
				EventSpecType: specType,
				Format:        format,
			},
			Expected: &model.EventSpec{
				Data:   &data,
				Format: format,
				Type:   specType,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {

			// when
			result := testCase.Input.ToEventSpec()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
