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
	bndlID := "bar"
	appID := "baz"
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
				Name:        name,
				Description: &desc,
				Group:       &group,
			},
			Expected: &model.EventDefinition{
				ApplicationID: appID,
				Tenant:        tenant,
				Name:          name,
				Description:   &desc,
				Group:         &group,
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
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {

			// when
			result := testCase.Input.ToEventDefinitionWithinBundle(id, appID, bndlID, tenant, 0)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
