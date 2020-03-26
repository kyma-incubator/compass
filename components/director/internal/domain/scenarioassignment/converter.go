package scenarioassignment

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

func (c *converter) FromInputGraphQL(in graphql.AutomaticScenarioAssignmentSetInput) model.AutomaticScenarioAssignment {
	out := model.AutomaticScenarioAssignment{
		ScenarioName: in.ScenarioName,
	}

	if in.Selector != nil {
		out.Selector = c.LabelSelectorFromInput(*in.Selector)
	}
	return out
}

func (c *converter) LabelSelectorFromInput(in graphql.LabelSelectorInput) model.LabelSelector {
	return model.LabelSelector{
		Key:   in.Key,
		Value: in.Value,
	}
}

func (c *converter) ToGraphQL(in model.AutomaticScenarioAssignment) graphql.AutomaticScenarioAssignment {
	return graphql.AutomaticScenarioAssignment{
		ScenarioName: in.ScenarioName,
		Selector: &graphql.Label{
			Key:   in.Selector.Key,
			Value: in.Selector.Value,
		},
	}
}

func (c *converter) ToEntity(in model.AutomaticScenarioAssignment) Entity {
	return Entity{
		TenantID:      in.Tenant,
		Scenario:      in.ScenarioName,
		SelectorKey:   in.Selector.Key,
		SelectorValue: in.Selector.Value,
	}
}

func (c *converter) FromEntity(in Entity) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
		ScenarioName: in.Scenario,
		Tenant:       in.TenantID,
		Selector: model.LabelSelector{
			Key:   in.SelectorKey,
			Value: in.SelectorValue,
		},
	}
}

func (c *converter) MultipleToGraphQL(assignments []*model.AutomaticScenarioAssignment) []*graphql.AutomaticScenarioAssignment {
	var gqlAssignments []*graphql.AutomaticScenarioAssignment

	for _, v := range assignments {
		assignment := c.ToGraphQL(*v)
		gqlAssignments = append(gqlAssignments, &assignment)
	}
	return gqlAssignments
}
