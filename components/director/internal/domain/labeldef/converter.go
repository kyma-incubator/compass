package labeldef

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
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

func (c *converter) ToEntity(in model.LabelDefinition) (Entity, error) {
	out := Entity{
		ID:       in.ID,
		Key:      in.Key,
		TenantID: in.Tenant,
	}
	if in.Schema != nil {
		b, err := json.Marshal(in.Schema)
		if err != nil {
			return Entity{}, errors.Wrap(err, "while marshaling schema to JSON")
		}
		out.SchemaJSON = string(b)
	}
	return out, nil
}
