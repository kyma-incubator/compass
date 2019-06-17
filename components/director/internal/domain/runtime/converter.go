package runtime

import (
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct{}

func (c *converter) ToGraphQL(in *model.Runtime) *graphql.Runtime {
	if in == nil {
		return nil
	}

	return &graphql.Runtime{
		ID:          in.ID,
		Status:      c.statusToGraphQL(in.Status),
		Name:        in.Name,
		Description: in.Description,
		Tenant:      graphql.Tenant(in.Tenant),
		Annotations: in.Annotations,
		Labels:      in.Labels,
	}
}

func (c *converter) MultipleToGraphQL(in []*model.Runtime) []*graphql.Runtime {
	var runtimes []*graphql.Runtime
	for _, r := range in {
		if r == nil {
			continue
		}

		runtimes = append(runtimes, c.ToGraphQL(r))
	}

	return runtimes
}

func (c *converter) InputFromGraphQL(in graphql.RuntimeInput) model.RuntimeInput {
	var annotations map[string]interface{}
	if in.Annotations != nil {
		annotations = *in.Annotations
	}

	var labels map[string][]string
	if in.Labels != nil {
		labels = *in.Labels
	}

	return model.RuntimeInput{
		Name:        in.Name,
		Description: in.Description,
		Annotations: annotations,
		Labels:      labels,
	}
}

func (c *converter) statusToGraphQL(in *model.RuntimeStatus) *graphql.RuntimeStatus {
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
		Timestamp: graphql.Timestamp(in.Timestamp),
	}
}
