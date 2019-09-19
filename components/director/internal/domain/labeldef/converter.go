package labeldef

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

func NewConverter() *converter {
	return &converter{}
}

type converter struct{}

func (c *converter) FromGraphQL(input graphql.LabelDefinitionInput, tenant string) (model.LabelDefinition, error) {
	schema, err := input.Schema.Unmarshal()
	if err != nil {
		return model.LabelDefinition{}, err
	}

	return model.LabelDefinition{
		Key:    input.Key,
		Schema: schema,
		Tenant: tenant,
	}, nil
}

func (c *converter) ToGraphQL(in model.LabelDefinition) (graphql.LabelDefinition, error) {
	schema, err := graphql.MarshalSchema(in.Schema)
	if err != nil {
		return graphql.LabelDefinition{}, err
	}
	return graphql.LabelDefinition{
		Key:    in.Key,
		Schema: schema,
	}, nil
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
		out.SchemaJSON = sql.NullString{String: string(b), Valid: true}
	} else {
		out.SchemaJSON = sql.NullString{Valid: false}
	}
	return out, nil
}

func (c *converter) FromEntity(in Entity) (model.LabelDefinition, error) {
	out := model.LabelDefinition{
		ID:     in.ID,
		Key:    in.Key,
		Tenant: in.TenantID,
	}
	if in.SchemaJSON.Valid {
		mapDest := map[string]interface{}{}
		var tmp interface{}
		err := json.Unmarshal([]byte(in.SchemaJSON.String), &mapDest)
		if err != nil {
			return model.LabelDefinition{}, err
		}
		tmp = mapDest
		out.Schema = &tmp
	}
	return out, nil
}
