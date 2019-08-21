package runtime_auth

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/pkg/errors"
)

//go:generate mockery -name=AuthConverter -output=automock -outpkg=automock -case=underscore
type AuthConverter interface {
	ToGraphQL(in *model.Auth) *graphql.Auth
}

type converter struct {
	authConverter AuthConverter
}

func NewConverter(authConverter AuthConverter) *converter {
	return &converter{
		authConverter: authConverter,
	}
}

func (c *converter) ToGraphQL(in *model.RuntimeAuth) *graphql.RuntimeAuth {
	if in == nil {
		return nil
	}

	return &graphql.RuntimeAuth{
		RuntimeID: in.RuntimeID,
		Auth:      c.authConverter.ToGraphQL(in.Value),
	}
}

func (c *converter) ToEntity(in model.RuntimeAuth) (Entity, error) {
	value := sql.NullString{}
	if in.Value != nil {
		valueMarshalled, err := json.Marshal(in.Value)
		if err != nil {
			return Entity{}, errors.Wrap(err, "while marshalling Value")
		}
		value.Valid = true
		value.String = string(valueMarshalled)
	}

	return Entity{
		ID:        repo.NewNullableString(in.ID),
		TenantID:  in.TenantID,
		RuntimeID: in.RuntimeID,
		APIDefID:  in.APIDefID,
		Value:     value,
	}, nil
}

func (c *converter) FromEntity(in Entity) (model.RuntimeAuth, error) {
	out := model.RuntimeAuth{
		TenantID:  in.TenantID,
		RuntimeID: in.RuntimeID,
		APIDefID:  in.APIDefID,
	}

	if in.ID.Valid {
		out.ID = &in.ID.String
	}
	if in.Value.Valid {
		var auth model.Auth
		err := json.Unmarshal([]byte(in.Value.String), &auth)
		if err != nil {
			return model.RuntimeAuth{}, err
		}
		out.Value = &auth
	}

	return out, nil
}
