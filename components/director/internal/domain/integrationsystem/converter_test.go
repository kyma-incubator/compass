package integrationsystem_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	converter := integrationsystem.NewConverter()

	testCases := []struct {
		Name     string
		Input    *model.IntegrationSystem
		Expected *graphql.IntegrationSystem
	}{
		{
			Name:     "All properties given",
			Input:    fixModelIntegrationSystem(testID, testName),
			Expected: fixGQLIntegrationSystem(testID, testName),
		},
		{
			Name:     "Empty",
			Input:    &model.IntegrationSystem{},
			Expected: &graphql.IntegrationSystem{},
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
			res := converter.ToGraphQL(testCase.Input)

			// THEN
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// GIVEN
	input := []*model.IntegrationSystem{
		fixModelIntegrationSystem("id1", "name1"),
		fixModelIntegrationSystem("id2", "name2"),
		{},
		nil,
	}
	expected := []*graphql.IntegrationSystem{
		fixGQLIntegrationSystem("id1", "name1"),
		fixGQLIntegrationSystem("id2", "name2"),
		{},
	}
	converter := integrationsystem.NewConverter()

	// WHEN
	res := converter.MultipleToGraphQL(input)

	// THEN
	assert.Equal(t, expected, res)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// GIVEN
	converter := integrationsystem.NewConverter()

	testCases := []struct {
		Name     string
		Input    graphql.IntegrationSystemInput
		Expected model.IntegrationSystemInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLIntegrationSystemInput(testName),
			Expected: fixModelIntegrationSystemInput(testName),
		},
		{
			Name:     "Empty",
			Input:    graphql.IntegrationSystemInput{},
			Expected: model.IntegrationSystemInput{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			res := converter.InputFromGraphQL(testCase.Input)

			// THEN
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	// GIVEN
	converter := integrationsystem.NewConverter()

	testCases := []struct {
		Name     string
		Input    *model.IntegrationSystem
		Expected *integrationsystem.Entity
	}{
		{
			Name:     "All properties given",
			Input:    fixModelIntegrationSystem(testID, testName),
			Expected: fixEntityIntegrationSystem(testID, testName),
		},
		{
			Name:     "Empty",
			Input:    &model.IntegrationSystem{},
			Expected: &integrationsystem.Entity{},
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
			res := converter.ToEntity(testCase.Input)

			// THEN
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	// GIVEN
	converter := integrationsystem.NewConverter()

	testCases := []struct {
		Name     string
		Input    *integrationsystem.Entity
		Expected *model.IntegrationSystem
	}{
		{
			Name:     "All properties given",
			Input:    fixEntityIntegrationSystem(testID, testName),
			Expected: fixModelIntegrationSystem(testID, testName),
		},
		{
			Name:     "Empty",
			Input:    &integrationsystem.Entity{},
			Expected: &model.IntegrationSystem{},
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
			res := converter.FromEntity(testCase.Input)

			// THEN
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
