package scenarioassignment

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

func NewConverter() *converter{
	return &converter{}
}
type converter struct{}

func (c *converter) FromInputGraphql(in graphql.AutomaticScenarioAssignmentSetInput, tenant string) (model.AutomaticScenarioAssignment, error) {
	out := model.AutomaticScenarioAssignment{
		ScenarioName: in.ScenarioName,
		Tenant:       tenant,
	}

	if in.Selector != nil {
		out.Selector = model.LabelSelector{
			Key: in.Selector.Key,
		}

		strVal, ok := (in.Selector.Value).(string)
		if ok {
			out.Selector.Value = strVal
		} else {
			return model.AutomaticScenarioAssignment{}, errors.New("value has to be a string")
		}
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
