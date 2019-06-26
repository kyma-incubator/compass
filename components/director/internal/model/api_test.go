package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)
func TestAPIDefinitionInput_ToAPIDefinition(t *testing.T) {
	// given
	url := "https://foo.bar"
	desc := "Sample"
	id := "foo"
	tenant := "sample"
	testCases := []struct {
		Name     string
		Input    *model.ApplicationInput
		Expected *model.Application
	}{
		{
			Name: "All properties given",
			Input: &model.ApplicationInput{
				Name:        "Foo",
				Description: &desc,
				Annotations: map[string]interface{}{
					"key": "value",
				},
				Labels: map[string][]string{
					"test": {"val", "val2"},
				},
				HealthCheckURL: &url,
			},
			Expected: &model.Application{
				Name:        "Foo",
				ID:          id,
				Tenant:      tenant,
				Description: &desc,
				Annotations: map[string]interface{}{
					"key": "value",
				},
				Labels: map[string][]string{
					"test": {"val", "val2"},
				},
				HealthCheckURL: &url,
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
			result := testCase.Input.ToApplication(id, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
