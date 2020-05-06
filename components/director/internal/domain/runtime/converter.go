package runtime

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToGraphQL(in *model.Runtime) *graphql.Runtime {
	if in == nil {
		return nil
	}

	return &graphql.Runtime{
		ID:          in.ID,
		Status:      c.statusToGraphQL(in.Status),
		Name:        in.Name,
		Description: in.Description,
		Metadata:    c.metadataToGraphQL(in.CreationTimestamp),
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
	var labels map[string]interface{}
	if in.Labels != nil {
		labels = *in.Labels
	}

	return model.RuntimeInput{
		Name:            in.Name,
		Description:     in.Description,
		Labels:          labels,
		StatusCondition: c.statusConditionToModel(in.StatusCondition),
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
	case model.RuntimeStatusConditionProvisioning:
		condition = graphql.RuntimeStatusConditionProvisioning
	case model.RuntimeStatusConditionFailed:
		condition = graphql.RuntimeStatusConditionFailed
	case model.RuntimeStatusConditionConnected:
		condition = graphql.RuntimeStatusConditionConnected
	default:
		condition = graphql.RuntimeStatusConditionInitial
	}

	return &graphql.RuntimeStatus{
		Condition: condition,
		Timestamp: graphql.Timestamp(in.Timestamp),
	}
}

func (c *converter) metadataToGraphQL(creationTimestamp time.Time) *graphql.RuntimeMetadata {
	return &graphql.RuntimeMetadata{
		CreationTimestamp: graphql.Timestamp(creationTimestamp),
	}
}

func (c *converter) statusConditionToModel(in *graphql.RuntimeStatusCondition) *model.RuntimeStatusCondition {
	if in == nil {
		return nil
	}

	var condition model.RuntimeStatusCondition
	switch *in {
	case graphql.RuntimeStatusConditionConnected:
		condition = model.RuntimeStatusConditionConnected
	case graphql.RuntimeStatusConditionFailed:
		condition = model.RuntimeStatusConditionFailed
	case graphql.RuntimeStatusConditionProvisioning:
		condition = model.RuntimeStatusConditionProvisioning
	case graphql.RuntimeStatusConditionInitial:
		fallthrough
	default:
		condition = model.RuntimeStatusConditionInitial
	}

	return &condition
}
