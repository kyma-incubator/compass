package integrationsystem

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct{}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToGraphQL(in *model.IntegrationSystem) *graphql.IntegrationSystem {
	if in == nil {
		return nil
	}

	return &graphql.IntegrationSystem{
		ID:          in.ID,
		Name:        in.Name,
		Description: in.Description,
	}
}

func (c *converter) MultipleToGraphQL(in []*model.IntegrationSystem) []*graphql.IntegrationSystem {
	var intSys []*graphql.IntegrationSystem
	for _, r := range in {
		if r == nil {
			continue
		}

		intSys = append(intSys, c.ToGraphQL(r))
	}

	return intSys
}

func (c *converter) InputFromGraphQL(in graphql.IntegrationSystemInput) model.IntegrationSystemInput {
	return model.IntegrationSystemInput{
		Name:        in.Name,
		Description: in.Description,
	}
}

func (c *converter) ToEntity(in *model.IntegrationSystem) *Entity {
	if in == nil {
		return nil
	}
	return &Entity{
		ID:          in.ID,
		Name:        in.Name,
		Description: in.Description,
	}
}

func (c *converter) FromEntity(in *Entity) *model.IntegrationSystem {
	if in == nil {
		return nil
	}
	return &model.IntegrationSystem{
		ID:          in.ID,
		Name:        in.Name,
		Description: in.Description,
	}
}
