package systemauth

import (
	"errors"
	authConv "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) GraphQLToModel(appSysAuth *graphql.AppSystemAuth) (*SystemAuth, error) {
	if appSysAuth.Type == nil {
		return nil, errors.New("no system auth type provided")
	}

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
			ID:                  appSysAuth.ID,
			TenantID:            appSysAuth.TenantID,
			IntegrationSystemID: appSysAuth.ReferenceObjectID,
			Value:               auth,
		}, nil
	case graphql.SystemAuthReferenceTypeRuntime:
		auth, err := authConv.ToModel(appSysAuth.Auth)
		if err != nil {
			return nil, err
		}
		return &SystemAuth{
			ID:        appSysAuth.ID,
			TenantID:  appSysAuth.TenantID,
			RuntimeID: appSysAuth.ReferenceObjectID,
			Value:     auth,
		}, nil
	default:
		return nil, errors.New("could not determine system auth")
	}
}
