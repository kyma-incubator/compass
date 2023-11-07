package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestEntityTypeMappingInput_ToEntityTypeMapping(t *testing.T) {
	// GIVEN
	id := "foo"
	testAPIDefinitionID := "bar"
	testEventDefinitionID := "bar"

	testCases := []struct {
		Name         string
		Input        *model.EntityTypeMappingInput
		ResourceType resource.Type
		ResourceID   string
		Expected     *model.EntityTypeMapping
	}{
		{
			Name:  "All properties given for API Definition",
			Input: &model.EntityTypeMappingInput{},
			Expected: &model.EntityTypeMapping{
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
				},
				APIDefinitionID: &testAPIDefinitionID,
			},
			ResourceType: resource.API,
			ResourceID:   testAPIDefinitionID,
		},
		{
			Name:  "All properties given for Event Definition",
			Input: &model.EntityTypeMappingInput{},
			Expected: &model.EntityTypeMapping{
				BaseEntity: &model.BaseEntity{
					ID:    id,
					Ready: true,
				},
				EventDefinitionID: &testEventDefinitionID,
			},
			ResourceType: resource.EventDefinition,
			ResourceID:   testEventDefinitionID,
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
			result := testCase.Input.ToEntityTypeMapping(id, testCase.ResourceType, testCase.ResourceID)

			// THEN
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
