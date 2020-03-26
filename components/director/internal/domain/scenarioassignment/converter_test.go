package scenarioassignment_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestFromInputGraphql(t *testing.T) {
	sut := scenarioassignment.NewConverter()
	t.Run("happy path", func(t *testing.T) {
		// WHEN
		actual := sut.FromInputGraphQL(graphql.AutomaticScenarioAssignmentSetInput{
			ScenarioName: scenarioName,
			Selector: &graphql.LabelSelectorInput{
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

func TestLabelSelectorFromInput(t *testing.T) {
	//GIVEN
	sut := scenarioassignment.NewConverter()
	//WHEN
	actual := sut.LabelSelectorFromInput(graphql.LabelSelectorInput{
		Key:   "test-key",
		Value: "test-value",
	})
	//THEN
	assert.Equal(t, model.LabelSelector{
		Key:   "test-key",
		Value: "test-value",
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

func TestConverter_MultipleToGraphQL(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		// GIVEN
		in := []*model.AutomaticScenarioAssignment{
			{
				ScenarioName: "Scenario-A",
				Tenant:       "4e7b4cc2-09d7-44e8-9e88-d70b8e7adef4",
				Selector: model.LabelSelector{
					Key:   "A-Key",
					Value: "A-Value",
				},
			},
			{
				ScenarioName: "Scenario-B",
				Tenant:       "475107c3-8938-4cec-b4f2-1b22df90a264",
				Selector: model.LabelSelector{
					Key:   "B-Key",
					Value: "B-Value",
				},
			},
			{
				ScenarioName: "Scenario-C",
				Tenant:       "4e3c3ae0-61f9-414f-b88d-7328c2bf4550",
				Selector: model.LabelSelector{
					Key:   "C-Key",
					Value: "C-Value",
				},
			},
		}
		expected := []*graphql.AutomaticScenarioAssignment{
			{
				ScenarioName: "Scenario-A",
				Selector: &graphql.Label{
					Key:   "A-Key",
					Value: "A-Value",
				},
			},
			{
				ScenarioName: "Scenario-B",
				Selector: &graphql.Label{
					Key:   "B-Key",
					Value: "B-Value",
				},
			},
			{
				ScenarioName: "Scenario-C",
				Selector: &graphql.Label{
					Key:   "C-Key",
					Value: "C-Value",
				},
			},
		}
		sut := scenarioassignment.NewConverter()

		// WHEN
		actual := sut.MultipleToGraphQL(in)
		// THEN
		assert.Equal(t, expected, actual)
	})
}
