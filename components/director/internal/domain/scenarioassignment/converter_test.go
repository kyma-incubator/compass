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
		actual, err := sut.FromInputGraphQL(graphql.AutomaticScenarioAssignmentSetInput{
			ScenarioName: scenarioName,
			Selector: &graphql.LabelInput{
				Key:   "my-label",
				Value: "my-value",
			},
		})
		// THEN
		assert.Equal(t, model.AutomaticScenarioAssignment{
			ScenarioName: scenarioName,
			Selector: model.LabelSelector{
				Key:   "my-label",
				Value: "my-value",
			},
		}, actual)

	})

	t.Run("error on converting value which is not a string", func(t *testing.T) {
		_, err := sut.FromInputGraphQL(graphql.AutomaticScenarioAssignmentSetInput{
			ScenarioName: "scenario-A",
			Selector: &graphql.LabelInput{
				Key:   "my-label",
				Value: 123,
			},
		})
		// THEN
		require.EqualError(t, err, "value has to be a string")
	})
}

func TestToGraphQL(t *testing.T) {
	// GIVEN
	sut := scenarioassignment.NewConverter()
	// WHEN
	actual := sut.ToGraphQL(model.AutomaticScenarioAssignment{
		ScenarioName: scenarioName,
		Tenant:       tenantID,
		Selector: model.LabelSelector{
			Key:   "my-label",
			Value: "my-value",
		},
	})
	// THEN
	assert.Equal(t, graphql.AutomaticScenarioAssignment{
		ScenarioName: scenarioName,
		Selector: &graphql.Label{
			Key:   "my-label",
			Value: "my-value",
		},
	}, actual)
}

func TestToEntity(t *testing.T) {
	// GIVEN
	sut := scenarioassignment.NewConverter()
	// WHEN
	actual := sut.ToEntity(model.AutomaticScenarioAssignment{
		ScenarioName: scenarioName,
		Tenant:       tenantID,
		Selector: model.LabelSelector{
			Key:   "my-label",
			Value: "my-value",
		},
	})

	// THEN
	assert.Equal(t, scenarioassignment.Entity{
		Scenario:      scenarioName,
		TenantID:      tenantID,
		SelectorKey:   "my-label",
		SelectorValue: "my-value",
	}, actual)
}

func TestFromEntity(t *testing.T) {
	// GIVEN
	sut := scenarioassignment.NewConverter()
	// WHEN
	actual := sut.FromEntity(scenarioassignment.Entity{
		Scenario:      scenarioName,
		TenantID:      tenantID,
		SelectorKey:   "my-label",
		SelectorValue: "my-value",
	})

	// THEN
	assert.Equal(t, model.AutomaticScenarioAssignment{
		ScenarioName: scenarioName,
		Tenant:       tenantID,
		Selector: model.LabelSelector{
			Key:   "my-label",
			Value: "my-value",
		},
	}, actual)
}
