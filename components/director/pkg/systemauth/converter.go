package systemauth

import (
	authConv "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) GraphQLToModel(in *graphql.AppSystemAuth) (*SystemAuth, error) {
	if in.Type == nil {
		return nil, errors.New("cannot get system auth type")

	}

	switch *in.Type {
	case graphql.SystemAuthReferenceTypeApplication:
		auth, err := authConv.ToModel(in.Auth)
		if err != nil {
			return nil, err
		}
		return &SystemAuth{
			ID:       in.ID,
			TenantID: in.TenantID,
			AppID:    in.ReferenceObjectID,
			Value:    auth,
		}, nil
	case graphql.SystemAuthReferenceTypeIntegrationSystem:
		auth, err := authConv.ToModel(in.Auth)
		if err != nil {
			return nil, err
		}
		return &SystemAuth{
			ID:                  in.ID,
			TenantID:            in.TenantID,
			IntegrationSystemID: in.ReferenceObjectID,
			Value:               auth,
		}, nil
	case graphql.SystemAuthReferenceTypeRuntime:
		auth, err := authConv.ToModel(in.Auth)
		if err != nil {
			return nil, err
		}
		return &SystemAuth{
			ID:        in.ID,
			TenantID:  in.TenantID,
			RuntimeID: in.ReferenceObjectID,
			Value:     auth,
		}, nil
	default:
		return nil, errors.New("could not determine system auth")
	}
}
