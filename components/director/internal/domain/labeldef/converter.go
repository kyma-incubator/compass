package labeldef

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

func (c *converter) FromGraphQL(input graphql.LabelDefinitionInput, tenant string) model.LabelDefinition {
	return model.LabelDefinition{
		Key:    input.Key,
		Schema: input.Schema,
		Tenant: tenant,
	}
}

func (c *converter) ToGraphQL(in model.LabelDefinition) graphql.LabelDefinition {
	return graphql.LabelDefinition{
		Key:    in.Key,
		Schema: in.Schema,
	}
}
