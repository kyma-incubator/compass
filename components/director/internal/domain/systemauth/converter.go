package systemauth

import (
	"database/sql"
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/systemauth"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/pkg/errors"
)

// AuthConverter missing godoc
//go:generate mockery --name=AuthConverter --output=automock --outpkg=automock --case=underscore
type AuthConverter interface {
	ToGraphQL(in *model.Auth) (*graphql.Auth, error)
	ModelFromGraphQLInput(in graphql.AuthInput) (*model.Auth, error)
}

type converter struct {
	authConverter AuthConverter
}

// NewConverter missing godoc
func NewConverter(authConverter AuthConverter) *converter {
	return &converter{
		authConverter: authConverter,
	}
}

// ToGraphQL missing godoc
func (c *converter) ToGraphQL(in *systemauth.SystemAuth) (graphql.SystemAuth, error) {
	if in == nil {
		return nil, nil
	}

	auth, err := c.authConverter.ToGraphQL(in.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Auth")
	}

	objectType, err := in.GetReferenceObjectType()
	if err != nil {
		return nil, err
	}

	systemAuthTypeApplication := graphql.SystemAuthReferenceTypeApplication
	systemAuthTypeRuntime := graphql.SystemAuthReferenceTypeRuntime
	systemAuthTypeIntSystem := graphql.SystemAuthReferenceTypeIntegrationSystem
	switch objectType {
	case systemauth.ApplicationReference:
		return &graphql.AppSystemAuth{
			ID:                in.ID,
			Auth:              auth,
			Type:              &systemAuthTypeApplication,
			TenantID:          in.TenantID,
			ReferenceObjectID: in.AppID,
		}, nil
	case systemauth.IntegrationSystemReference:
		return &graphql.IntSysSystemAuth{
			ID:                in.ID,
			Auth:              auth,
			Type:              &systemAuthTypeIntSystem,
			TenantID:          in.TenantID,
			ReferenceObjectID: in.IntegrationSystemID,
		}, nil
	case systemauth.RuntimeReference:
		return &graphql.RuntimeSystemAuth{
			ID:                in.ID,
			Auth:              auth,
			Type:              &systemAuthTypeRuntime,
			TenantID:          in.TenantID,
			ReferenceObjectID: in.RuntimeID,
		}, nil
	default:
		return nil, errors.New("invalid object type")
	}
}

// ToEntity missing godoc
func (c *converter) ToEntity(in systemauth.SystemAuth) (Entity, error) {
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

// FromEntity missing godoc
func (c *converter) FromEntity(in Entity) (systemauth.SystemAuth, error) {
	var value *model.Auth
	if in.Value.Valid {
		var tmpAuth model.Auth
		err := json.Unmarshal([]byte(in.Value.String), &tmpAuth)
		if err != nil {
			return systemauth.SystemAuth{}, err
		}
		value = &tmpAuth
	}

	return systemauth.SystemAuth{
		ID:                  in.ID,
		TenantID:            repo.StringPtrFromNullableString(in.TenantID),
		AppID:               repo.StringPtrFromNullableString(in.AppID),
		RuntimeID:           repo.StringPtrFromNullableString(in.RuntimeID),
		IntegrationSystemID: repo.StringPtrFromNullableString(in.IntegrationSystemID),
		Value:               value,
	}, nil
}
