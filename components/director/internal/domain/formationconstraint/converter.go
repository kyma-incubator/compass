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

	return &graphql.FormationConstraint{
		ID:              in.ID,
		Name:            in.Name,
		ConstraintType:  string(in.ConstraintType),
		TargetOperation: string(in.TargetOperation),
		Operator:        in.Operator,
		ResourceType:    string(in.ResourceType),
		ResourceSubtype: in.ResourceSubtype,
		OperatorScope:   string(in.OperatorScope),
		InputTemplate:   in.InputTemplate,
		ConstraintScope: string(in.ConstraintScope),
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
		ConstraintType:  string(in.ConstraintType),
		TargetOperation: string(in.TargetOperation),
		Operator:        in.Operator,
		ResourceType:    string(in.ResourceType),
		ResourceSubtype: in.ResourceSubtype,
		OperatorScope:   string(in.OperatorScope),
		InputTemplate:   in.InputTemplate,
		ConstraintScope: string(in.ConstraintScope),
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
		ConstraintType:  model.FormationConstraintType(e.ConstraintType),
		TargetOperation: model.TargetOperation(e.TargetOperation),
		Operator:        e.Operator,
		ResourceType:    model.ResourceType(e.ResourceType),
		ResourceSubtype: e.ResourceSubtype,
		OperatorScope:   model.OperatorScopeType(e.OperatorScope),
		InputTemplate:   e.InputTemplate,
		ConstraintScope: model.FormationConstraintScope(e.ConstraintScope),
	}
}

// MultipleFromEntity converts multiple entities to internal models
func (c *converter) MultipleFromEntity(in EntityCollection) []*model.FormationConstraint {
	if in == nil {
		return nil
	}
	formationConstraints := make([]*model.FormationConstraint, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		fc := c.FromEntity(r)
		formationConstraints = append(formationConstraints, fc)
	}

	return formationConstraints
}

// FromInputGraphQL converts from GraphQL input to internal model input
func (c *converter) FromInputGraphQL(in *graphql.FormationConstraintInput) *model.FormationConstraintInput {
	return &model.FormationConstraintInput{
		Name:                in.Name,
		ConstraintType:      model.FormationConstraintType(in.ConstraintType),
		TargetOperation:     model.TargetOperation(in.TargetOperation),
		Operator:            in.Operator,
		ResourceType:        model.ResourceType(in.ResourceType),
		ResourceSubtype:     in.ResourceSubtype,
		OperatorScope:       model.OperatorScopeType(in.OperatorScope),
		InputTemplate:       in.InputTemplate,
		ConstraintScope:     model.FormationConstraintScope(in.ConstraintScope),
		FormationTemplateID: in.FormationTemplateID,
	}
}

// FromModelInputToModel converts from internal model input to internal model
func (c *converter) FromModelInputToModel(in *model.FormationConstraintInput, id string) *model.FormationConstraint {
	return &model.FormationConstraint{
		ID:              id,
		Name:            in.Name,
		ConstraintType:  in.ConstraintType,
		TargetOperation: in.TargetOperation,
		Operator:        in.Operator,
		ResourceType:    in.ResourceType,
		ResourceSubtype: in.ResourceSubtype,
		OperatorScope:   in.OperatorScope,
		InputTemplate:   in.InputTemplate,
		ConstraintScope: in.ConstraintScope,
	}
}
