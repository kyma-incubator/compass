package formationtemplateconstraintreferences

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// NewConverter creates a new formationTemplate-constraint references converter
func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

// ToEntity converts from internal model to entity
func (c *converter) ToEntity(in *model.FormationTemplateConstraintReference) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ConstraintID:        in.ConstraintID,
		FormationTemplateID: in.FormationTemplateID,
	}
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(e *Entity) *model.FormationTemplateConstraintReference {
	if e == nil {
		return nil
	}

	return &model.FormationTemplateConstraintReference{
		ConstraintID:        e.ConstraintID,
		FormationTemplateID: e.FormationTemplateID,
	}
}

// ToModel converts from graphql to internal model
func (c *converter) ToModel(in *graphql.ConstraintReference) *model.FormationTemplateConstraintReference {
	if in == nil {
		return nil
	}

	return &model.FormationTemplateConstraintReference{
		ConstraintID:        in.ConstraintID,
		FormationTemplateID: in.FormationTemplateID,
	}
}

// ToGraphql converts from internal model to graphql
func (c *converter) ToGraphql(in *model.FormationTemplateConstraintReference) *graphql.ConstraintReference {
	if in == nil {
		return nil
	}

	return &graphql.ConstraintReference{
		ConstraintID:        in.ConstraintID,
		FormationTemplateID: in.FormationTemplateID,
	}
}
