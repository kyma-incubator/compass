package systemauth

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/pkg/errors"
)

//go:generate mockery -name=AuthConverter -output=automock -outpkg=automock -case=underscore
type AuthConverter interface {
	ToGraphQL(in *model.Auth) (*graphql.Auth, error)
}

type converter struct {
	authConverter AuthConverter
}

func NewConverter(authConverter AuthConverter) *converter {
	return &converter{
		authConverter: authConverter,
	}
}

func (c *converter) ToGraphQL(in *model.SystemAuth) (*graphql.SystemAuth, error) {
	if in == nil {
		return nil, nil
	}

	auth, err := c.authConverter.ToGraphQL(in.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Auth")
	}

	return &graphql.SystemAuth{
		ID:   in.ID,
		Auth: auth,
	}, nil
}

func (c *converter) ToEntity(in model.SystemAuth) (Entity, error) {
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
		ID:                  in.ID,
		TenantID:            repo.NewNullableString(in.TenantID),
		AppID:               repo.NewNullableString(in.AppID),
		RuntimeID:           repo.NewNullableString(in.RuntimeID),
		IntegrationSystemID: repo.NewNullableString(in.IntegrationSystemID),
		Value:               value,
	}, nil
}

func (c *converter) FromEntity(in Entity) (model.SystemAuth, error) {
	var value *model.Auth
	if in.Value.Valid {
		var tmpAuth model.Auth
		err := json.Unmarshal([]byte(in.Value.String), &tmpAuth)
		if err != nil {
			return model.SystemAuth{}, err
		}
		value = &tmpAuth
	}

	return model.SystemAuth{
		ID:                  in.ID,
		TenantID:            repo.StringPtrFromNullableString(in.TenantID),
		AppID:               repo.StringPtrFromNullableString(in.AppID),
		RuntimeID:           repo.StringPtrFromNullableString(in.RuntimeID),
		IntegrationSystemID: repo.StringPtrFromNullableString(in.IntegrationSystemID),
		Value:               value,
	}, nil
}
