package scenarioassignment

import (
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

func (c *converter) FromInputGraphQL(in graphql.AutomaticScenarioAssignmentSetInput, tenant string) (model.AutomaticScenarioAssignment, error) {
	out := model.AutomaticScenarioAssignment{
		ScenarioName: in.ScenarioName,
		Tenant:       tenant,
	}

	if in.Selector != nil {
		out.Selector = model.LabelSelector{
			Key: in.Selector.Key,
		}

		strVal, ok := (in.Selector.Value).(string)
		if !ok {
			return model.AutomaticScenarioAssignment{}, errors.New("value has to be a string")
		}
		out.Selector.Value = strVal
	}
	return out, nil
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
