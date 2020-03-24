package scenarioassignment_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromInputGraphql(t *testing.T) {
	sut := scenarioassignment.NewConverter()
	t.Run("happy path", func(t *testing.T) {
		// WHEN
		actual, err := sut.FromInputGraphql(graphql.AutomaticScenarioAssignmentSetInput{
			ScenarioName: "scenario-A",
			Selector: &graphql.LabelInput{
				Key:   "my-label",
				Value: "my-value",
			},
		}, "tenant")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, model.AutomaticScenarioAssignment{
			Tenant:       "tenant",
			ScenarioName: "scenario-A",
			Selector: model.LabelSelector{
				Key:   "my-label",
				Value: "my-value",
			},
		}, actual)

	})

	t.Run("error on converting value which is not a string", func(t *testing.T) {
		_, err := sut.FromInputGraphql(graphql.AutomaticScenarioAssignmentSetInput{
			ScenarioName: "scenario-A",
			Selector: &graphql.LabelInput{
				Key:   "my-label",
				Value: 123,
			},
		}, "tenant")
		// THEN
		require.EqualError(t, err, "value has to be a string")
	})
}

func TestToGraphQL(t *testing.T) {
	// GIVEN
	sut := scenarioassignment.NewConverter()
	// WHEN
	actual := sut.ToGraphQL(model.AutomaticScenarioAssignment{
		ScenarioName: "scenario-A",
		Tenant:       "tenant",
		Selector: model.LabelSelector{
			Key:   "my-label",
			Value: "my-value",
		},
	})
	// THEN
	assert.Equal(t, graphql.AutomaticScenarioAssignment{
		ScenarioName: "scenario-A",
		Selector: &graphql.Label{
			Key:   "my-label",
			Value: "my-value",
		},
	}, actual)
}

func TestToEntity(t *testing.T) {
	// GIVEN
	sut := scenarioassignment.NewConverter()
	// WHEM
	actual := sut.ToEntity(model.AutomaticScenarioAssignment{
		ScenarioName: "scenario-A",
		Tenant:       "tenant",
		Selector: model.LabelSelector{
			Key:   "my-label",
			Value: "my-value",
		},
	})
	assert.Equal(t, scenarioassignment.Entity{
		Scenario:      "scenario-A",
		TenantID:      "tenant",
		SelectorKey:   "my-label",
		SelectorValue: "my-value",
	}, actual)
}
