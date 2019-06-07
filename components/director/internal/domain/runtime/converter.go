package runtime

import (
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type Converter struct{}

func (c *Converter) ToGraphQL(in *model.Runtime) *graphql.Runtime {
	if in == nil {
		return nil
	}

	return &graphql.Runtime{
		ID:          in.ID,
		Status:      c.statusToGraphQL(in.Status),
		Name:        in.Name,
		Description: in.Description,
		// Tenant:      in.Tenant, //TODO: Wait for scalar marshalling
		//Annotations:in.Annotations, //TODO: Wait for scalar marshalling
		//Labels:in.Labels, //TODO: Wait for scalar marshalling
	}
}

func (c *Converter) MultipleToGraphQL(in []*model.Runtime) []*graphql.Runtime {
	var runtimes []*graphql.Runtime
	for _, r := range in {
		if r == nil {
			continue
		}

		runtimes = append(runtimes, c.ToGraphQL(r))
	}

	return runtimes
}

func (c *Converter) InputFromGraphQL(in graphql.RuntimeInput) model.RuntimeInput {
	return model.RuntimeInput{
		Name:        in.Name,
		Description: in.Description,
		//Annotations: in.Annotations, //TODO: Wait for scalar unmarshalling
		//Labels: in.Labels, //TODO: Wait for scalar unmarshalling
	}
}

func (c *Converter) statusToGraphQL(in *model.RuntimeStatus) *graphql.RuntimeStatus {
	if in == nil {
		return &graphql.RuntimeStatus{
			Condition: graphql.RuntimeStatusConditionInitial,
		}
	}

	var condition graphql.RuntimeStatusCondition

	switch in.Condition {
	case model.RuntimeStatusConditionInitial:
		condition = graphql.RuntimeStatusConditionInitial
	case model.RuntimeStatusConditionFailed:
		condition = graphql.RuntimeStatusConditionFailed
	case model.RuntimeStatusConditionReady:
		condition = graphql.RuntimeStatusConditionReady
	default:
		condition = graphql.RuntimeStatusConditionInitial
	}

	return &graphql.RuntimeStatus{
		Condition: condition,
		//Timestamp: in.Timestamp, //TODO: Wait for scalar unmarshalling
	}
}
