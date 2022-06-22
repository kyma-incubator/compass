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
	return &graphql.Formation{
		ID:                  i.ID,
		Name:                i.Name,
		FormationTemplateID: i.FormationTemplateID,
	}
}

// MultipleToGraphQL converts multiple internal models to GraphQL models
func (c *converter) MultipleToGraphQL(in []*model.Formation) []*graphql.Formation {
	if in == nil {
		return nil
	}
	formations := make([]*graphql.Formation, 0, len(in))
	for _, r := range in {
		if r == nil {
			continue
		}

		formations = append(formations, c.ToGraphQL(r))
	}

	return formations
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
