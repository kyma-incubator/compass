package model_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

func TestIntegrationSystemInput_ToIntegrationSystem(t *testing.T) {
	// given
	id := "foo"

	testCases := []struct {
		Name     string
		Input    *model.IntegrationSystemInput
		Expected model.IntegrationSystem
	}{
		{
			Name: "All properties given",
			Input: &model.IntegrationSystemInput{
				Name: "TestSystem",
			},
			Expected: model.IntegrationSystem{
				ID:   id,
				Name: "TestSystem",
			},
		},
		{
			Name:  "Empty",
			Input: &model.IntegrationSystemInput{},
			Expected: model.IntegrationSystem{
				ID: id,
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: model.IntegrationSystem{},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			result := testCase.Input.ToIntegrationSystem(id)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
