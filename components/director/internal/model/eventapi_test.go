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
	appID := "bar"
	desc := "Sample"
	name := "sample"
	group := "sampleGroup"

	testCases := []struct {
		Name     string
		Input    *model.EventAPIDefinitionInput
		Expected *model.EventAPIDefinition
	}{
		{
			Name: "All properties given",
			Input: &model.EventAPIDefinitionInput{
				Name:        name,
				Description: &desc,
				Group:       &group,
			},
			Expected: &model.EventAPIDefinition{
				ID:            id,
				ApplicationID: appID,
				Name:          name,
				Description:   &desc,
				Group:         &group,
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
			result := testCase.Input.ToEventAPIDefinition(id, appID, nil)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestEventAPIDefinitionInput_ToEventAPISpec(t *testing.T) {
	// given
	data := "bar"
	format := model.SpecFormat("Sample")
	specType := model.EventAPISpecType("sample")

	testCases := []struct {
		Name     string
		Input    *model.EventAPISpecInput
		Expected *model.EventAPISpec
	}{
		{
			Name: "All properties given",
			Input: &model.EventAPISpecInput{
				Data:          &data,
				EventSpecType: specType,
				Format:        format,
			},
			Expected: &model.EventAPISpec{
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
			result := testCase.Input.ToEventAPISpec(nil)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
