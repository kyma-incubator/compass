package formation

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct {
}

// NewConverter creates new formation converter
func NewConverter() *converter {
	return &converter{}
}

// FromGraphQL converts graphql.FormationInput to model.Formation
func (c *converter) FromGraphQL(i graphql.FormationInput) model.Formation {
	return model.Formation{Name: i.Name}
}

// ToGraphQL converts model.Formation to graphql.Formation
func (c *converter) ToGraphQL(i *model.Formation) *graphql.Formation {
	return &graphql.Formation{Name: i.Name}
}

func (c *converter) ToEntity(in *model.Formation) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ID:                  in.ID,
		TenantID:            in.TenantID,
		FormationTemplateID: in.FormationTemplateID,
		Name:                in.Name,
	}
}

func (c *converter) FromEntity(entity *Entity) *model.Formation {
	if entity == nil {
		return nil
	}

	return &model.Formation{
		ID:                  entity.ID,
		TenantID:            entity.TenantID,
		FormationTemplateID: entity.FormationTemplateID,
		Name:                entity.Name,
	}
}
