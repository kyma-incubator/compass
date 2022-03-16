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

func (c *converter) GraphQLToModel(appSysAuth *graphql.AppSystemAuth) (*SystemAuth, error) {
	switch *appSysAuth.Type {
	case graphql.SystemAuthReferenceTypeApplication:
		auth, err := authConv.ToModel(appSysAuth.Auth)
		if err != nil {
			return nil, err
		}
		return &SystemAuth{
			ID:       appSysAuth.ID,
			TenantID: appSysAuth.TenantID,
			AppID:    appSysAuth.ReferenceObjectID,
			Value:    auth,
		}, nil
	case graphql.SystemAuthReferenceTypeIntegrationSystem:
		auth, err := authConv.ToModel(appSysAuth.Auth)
		if err != nil {
			return nil, err
		}
		return &SystemAuth{
			ID:       appSysAuth.ID,
			TenantID: appSysAuth.TenantID,
			AppID:    appSysAuth.ReferenceObjectID,
			Value:    auth,
		}, nil
	case graphql.SystemAuthReferenceTypeRuntime:
		auth, err := authConv.ToModel(appSysAuth.Auth)
		if err != nil {
			return nil, err
		}
		return &SystemAuth{
			ID:       appSysAuth.ID,
			TenantID: appSysAuth.TenantID,
			AppID:    appSysAuth.ReferenceObjectID,
			Value:    auth,
		}, nil
	default:
		return nil, errors.New("could not determine system auth")
	}
}
