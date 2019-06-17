package application

import (
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct {
}

func (c *converter) ToGraphQL(in *model.Application) *graphql.Application {
	if in == nil {
		return nil
	}

	return &graphql.Application{
		ID:          in.ID,
		Status:      c.statusToGraphQL(in.Status),
		Name:        in.Name,
		Description: in.Description,
		Tenant:      graphql.Tenant(in.Tenant),
		Annotations: in.Annotations,
		Labels:      in.Labels,
	}
}

func (c *converter) MultipleToGraphQL(in []*model.Application) []*graphql.Application {
	var runtimes []*graphql.Application
	for _, r := range in {
		if r == nil {
			continue
		}

		runtimes = append(runtimes, c.ToGraphQL(r))
	}

	return runtimes
}

func (c *converter) InputFromGraphQL(in graphql.ApplicationInput) model.ApplicationInput {
	var annotations map[string]interface{}
	if in.Annotations != nil {
		annotations = *in.Annotations
	}

	var labels map[string][]string
	if in.Labels != nil {
		labels = *in.Labels
	}

	return model.ApplicationInput{
		Name:        in.Name,
		Description: in.Description,
		Annotations: annotations,
		Labels:      labels,
	}
}

func (c *converter) statusToGraphQL(in *model.ApplicationStatus) *graphql.ApplicationStatus {
	if in == nil {
		return &graphql.ApplicationStatus{
			Condition: graphql.ApplicationStatusConditionInitial,
		}
	}

	var condition graphql.ApplicationStatusCondition

	switch in.Condition {
	case model.ApplicationStatusConditionInitial:
		condition = graphql.ApplicationStatusConditionInitial
	case model.ApplicationStatusConditionFailed:
		condition = graphql.ApplicationStatusConditionFailed
	case model.ApplicationStatusConditionReady:
		condition = graphql.ApplicationStatusConditionReady
	default:
		condition = graphql.ApplicationStatusConditionInitial
	}

	return &graphql.ApplicationStatus{
		Condition: condition,
		Timestamp: graphql.Timestamp(in.Timestamp),
	}
}
