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
		Name:        in.Name,
		Description: in.Description,
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

func (c *converter) metadataToGraphQL(creationTimestamp time.Time) *graphql.RuntimeMetadata {
	return &graphql.RuntimeMetadata{
		CreationTimestamp: graphql.Timestamp(creationTimestamp),
	}
}
