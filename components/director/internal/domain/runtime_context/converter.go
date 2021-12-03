package runtimectx

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct{}

// NewConverter missing godoc
func NewConverter() *converter {
	return &converter{}
}

// ToGraphQL missing godoc
func (c *converter) ToGraphQL(in *model.RuntimeContext) *graphql.RuntimeContext {
	if in == nil {
		return nil
	}

	return &graphql.RuntimeContext{
		ID:    in.ID,
		Key:   in.Key,
		Value: in.Value,
	}
}

// MultipleToGraphQL missing godoc
func (c *converter) MultipleToGraphQL(in []*model.RuntimeContext) []*graphql.RuntimeContext {
	runtimeContexts := make([]*graphql.RuntimeContext, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		runtimeContexts = append(runtimeContexts, c.ToGraphQL(r))
	}

	return runtimeContexts
}

// InputFromGraphQL missing godoc
func (c *converter) InputFromGraphQL(in graphql.RuntimeContextInput, runtimeID string) model.RuntimeContextInput {
	var labels map[string]interface{}
	if in.Labels != nil {
		labels = in.Labels
	}

	return model.RuntimeContextInput{
		Key:       in.Key,
		Value:     in.Value,
		RuntimeID: runtimeID,
		Labels:    labels,
	}
}

func (c *converter) ToEntity(model *model.RuntimeContext) *RuntimeContext {
	return &RuntimeContext{
		ID:        model.ID,
		RuntimeID: model.RuntimeID,
		Key:       model.Key,
		Value:     model.Value,
	}
}

func (c *converter) FromEntity(e *RuntimeContext) *model.RuntimeContext {
	return &model.RuntimeContext{
		ID:        e.ID,
		RuntimeID: e.RuntimeID,
		Key:       e.Key,
		Value:     e.Value,
	}
}
