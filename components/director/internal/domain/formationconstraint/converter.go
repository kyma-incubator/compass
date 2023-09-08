package formationconstraint

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// NewConverter creates a new formation constraint converter
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// ToGraphQL converts from internal model to GraphQL output
func (c *converter) ToGraphQL(in *model.FormationConstraint) *graphql.FormationConstraint {
	if in == nil {
		return nil
	}

	createdAt := graphql.Timestamp{}
	if in.CreatedAt != nil {
		createdAt = graphql.Timestamp(*in.CreatedAt)
	}

	return &graphql.FormationConstraint{
		ID:              in.ID,
		Name:            in.Name,
		Description:     in.Description,
		ConstraintType:  string(in.ConstraintType),
		TargetOperation: string(in.TargetOperation),
		Operator:        in.Operator,
		ResourceType:    string(in.ResourceType),
		ResourceSubtype: in.ResourceSubtype,
		InputTemplate:   in.InputTemplate,
		ConstraintScope: string(in.ConstraintScope),
		Priority:        in.Priority,
		CreatedAt:       createdAt,
	}
}

// MultipleToGraphQL converts multiple internal models to GraphQL models
func (c *converter) MultipleToGraphQL(in []*model.FormationConstraint) []*graphql.FormationConstraint {
	if in == nil {
		return nil
	}
	formationConstraints := make([]*graphql.FormationConstraint, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		fc := c.ToGraphQL(r)
		formationConstraints = append(formationConstraints, fc)
	}

	return formationConstraints
}

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in *model.FormationConstraint) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ID:              in.ID,
		Name:            in.Name,
		Description:     in.Description,
		ConstraintType:  string(in.ConstraintType),
		TargetOperation: string(in.TargetOperation),
		Operator:        in.Operator,
		ResourceType:    string(in.ResourceType),
		ResourceSubtype: in.ResourceSubtype,
		InputTemplate:   in.InputTemplate,
		ConstraintScope: string(in.ConstraintScope),
		Priority:        in.Priority,
		CreatedAt:       in.CreatedAt,
	}
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(e *Entity) *model.FormationConstraint {
	if e == nil {
		return nil
	}

	return &model.FormationConstraint{
		ID:              e.ID,
		Name:            e.Name,
		Description:     e.Description,
		ConstraintType:  model.FormationConstraintType(e.ConstraintType),
		TargetOperation: model.TargetOperation(e.TargetOperation),
		Operator:        e.Operator,
		ResourceType:    model.ResourceType(e.ResourceType),
		ResourceSubtype: e.ResourceSubtype,
		InputTemplate:   e.InputTemplate,
		ConstraintScope: model.FormationConstraintScope(e.ConstraintScope),
		Priority:        e.Priority,
		CreatedAt:       e.CreatedAt,
	}
}

// FromInputGraphQL converts from GraphQL input to internal model input
func (c *converter) FromInputGraphQL(in *graphql.FormationConstraintInput) *model.FormationConstraintInput {
	priority := 0
	if in.Priority != nil {
		priority = *in.Priority
	}
	return &model.FormationConstraintInput{
		Name:            in.Name,
		Description:     in.Description,
		ConstraintType:  model.FormationConstraintType(in.ConstraintType),
		TargetOperation: model.TargetOperation(in.TargetOperation),
		Operator:        in.Operator,
		ResourceType:    model.ResourceType(in.ResourceType),
		ResourceSubtype: in.ResourceSubtype,
		InputTemplate:   in.InputTemplate,
		ConstraintScope: model.FormationConstraintScope(in.ConstraintScope),
		Priority:        priority,
	}
}

// FromModelInputToModel converts from internal model input to internal model
func (c *converter) FromModelInputToModel(in *model.FormationConstraintInput, id string) *model.FormationConstraint {
	return &model.FormationConstraint{
		ID:              id,
		Name:            in.Name,
		Description:     in.Description,
		ConstraintType:  in.ConstraintType,
		TargetOperation: in.TargetOperation,
		Operator:        in.Operator,
		ResourceType:    in.ResourceType,
		ResourceSubtype: in.ResourceSubtype,
		InputTemplate:   in.InputTemplate,
		ConstraintScope: in.ConstraintScope,
		Priority:        in.Priority,
	}
}
