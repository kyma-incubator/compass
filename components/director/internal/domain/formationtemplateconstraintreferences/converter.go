package formationtemplateconstraintreferences

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
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
		Constraint:        in.Constraint,
		FormationTemplate: in.FormationTemplate,
	}
}

// FromEntity converts from entity to internal model
func (c *converter) FromEntity(e *Entity) *model.FormationTemplateConstraintReference {
	if e == nil {
		return nil
	}

	return &model.FormationTemplateConstraintReference{
		Constraint:        e.Constraint,
		FormationTemplate: e.FormationTemplate,
	}
}

// MultipleFromEntity converts multiple entities to internal models
func (c *converter) MultipleFromEntity(in EntityCollection) []*model.FormationTemplateConstraintReference {
	if in == nil {
		return nil
	}
	references := make([]*model.FormationTemplateConstraintReference, 0, len(in))
	for _, e := range in {
		if e == nil {
			continue
		}

		fc := c.FromEntity(e)
		references = append(references, fc)
	}

	return references
}
