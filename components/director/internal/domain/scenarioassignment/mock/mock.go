package mock

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

func FixAssignmentForScenario(scenarioName string) *graphql.AutomaticScenarioAssignment {
	selector := &graphql.Label{
		Key:   "selector",
		Value: "dummy-value",
	}
	return FixAssignmentForScenarioWithSelector(scenarioName, selector)
}

func FixAssignmentForScenarioWithSelector(scenarioName string, selector *graphql.Label) *graphql.AutomaticScenarioAssignment {
	return &graphql.AutomaticScenarioAssignment{
		ScenarioName: scenarioName,
		Selector:     selector,
	}
}
