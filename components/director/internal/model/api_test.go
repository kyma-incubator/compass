package model_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestAPIDefinitionInput_ToAPIDefinitionWithBundleID(t *testing.T) {
	// given
	id := "foo"
	bndlID := "bar"
	desc := "Sample"
	name := "sample"
	targetUrl := "https://foo.bar"
	group := "sampleGroup"
	tenant := "tenant"

	testCases := []struct {
		Name     string
		Input    *model.APIDefinitionInput
		Expected *model.APIDefinition
	}{
		{
			Name: "All properties given",
			Input: &model.APIDefinitionInput{
				Name:        name,
				Description: &desc,
				TargetURL:   targetUrl,
				Group:       &group,
			},
			Expected: &model.APIDefinition{
				BundleID:    &bndlID,
				Name:        name,
				Description: &desc,
				TargetURL:   targetUrl,
				Group:       &group,
				Tenant:      tenant,
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
			result := testCase.Input.ToAPIDefinitionWithinBundle(id, bndlID, tenant)

			// then
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
