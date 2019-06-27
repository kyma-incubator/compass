package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestAPIDefinitionInput_ToAPIDefinition(t *testing.T) {
	// given
	id := "foo"
	appID := "bar"
	desc := "Sample"
	name := "sample"
	targetUrl := "https://foo.bar"
	group := "sampleGroup"

	testCases := []struct {
		Name     string
		Input    *model.APIDefinitionInput
		Expected *model.APIDefinition
	}{
		{
			Name: "All properties given",
			Input: &model.APIDefinitionInput{
				ApplicationID: appID,
				Name:          name,
				Description:   &desc,
				TargetURL:     targetUrl,
				Group:         &group,
			},
			Expected: &model.APIDefinition{
				ID:            id,
				ApplicationID: appID,
				Name:          name,
				Description:   &desc,
				TargetURL:     targetUrl,
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
			result := testCase.Input.ToAPIDefinition(id)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestAPIDefinitionInput_ToAPISpec(t *testing.T) {
	// given
	data := []byte("bar")
	format := model.SpecFormat("Sample")
	specType := model.APISpecType("sample")

	testCases := []struct {
		Name     string
		Input    *model.APISpecInput
		Expected *model.APISpec
	}{
		{
			Name: "All properties given",
			Input: &model.APISpecInput{
				Data:   &data,
				Format: &format,
				Type:   specType,
			},
			Expected: &model.APISpec{
				Data:   &data,
				Format: &format,
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
			result := testCase.Input.ToAPISpec()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}

func TestAPIDefinitionInput_ToVersion(t *testing.T) {
	// given
	value := "foo"
	deprecated := true
	deprecatedSince := "bar"
	forRemoval := false

	testCases := []struct {
		Name     string
		Input    *model.VersionInput
		Expected *model.Version
	}{
		{
			Name: "All properties given",
			Input: &model.VersionInput{
				Value:           value,
				Deprecated:      &deprecated,
				DeprecatedSince: &deprecatedSince,
				ForRemoval:      &forRemoval,
			},
			Expected: &model.Version{
				Value:           value,
				Deprecated:      &deprecated,
				DeprecatedSince: &deprecatedSince,
				ForRemoval:      &forRemoval,
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
			result := testCase.Input.ToVersion()

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
