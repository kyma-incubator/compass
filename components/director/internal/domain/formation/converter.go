package formation

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) FromGraphQL(i graphql.FormationInput) model.Formation {
	return model.Formation{Name: i.Name}
}

func (c *converter) ToGraphQL(i *model.Formation) *graphql.Formation {
	return &graphql.Formation{Name: i.Name}
}
