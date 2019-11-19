package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

func NewMapperForSystemAuth(systemAuthSvc systemauth.SystemAuthService, scopesGetter ScopesGetter) *mapperForSystemAuth {
	return &mapperForSystemAuth{
		systemAuthSvc: systemAuthSvc,
		scopesGetter:  scopesGetter,
	}
}

type mapperForSystemAuth struct {
	systemAuthSvc systemauth.SystemAuthService
	scopesGetter  ScopesGetter
}

func (m *mapperForSystemAuth) GetTenantAndScopes(ctx context.Context, reqData ReqData, authID string, authFlow AuthFlow) (ObjectContext, error) {
	sysAuth, err := m.systemAuthSvc.GetGlobal(ctx, authID)
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while retrieving system auth from database")
	}

	refObjType, err := sysAuth.GetReferenceObjectType()
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while getting reference object type")
	}

	var scopes string
	var tenant string

	switch refObjType {
	case model.IntegrationSystemReference:
		tenant, scopes, err = m.getTenantAndScopesForIntegrationSystem(reqData)
	case model.RuntimeReference, model.ApplicationReference:
		tenant, scopes, err = m.getTenantAndScopesForApplicationOrRuntime(sysAuth, refObjType, reqData, authFlow)
	default:
		return ObjectContext{}, errors.Errorf("unsupported reference object type (%s)", refObjType)
	}

	if err != nil {
		return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while fetching the tenant and scopes for object of type %s", refObjType))
	}

	objID, objType, err := getContextObj(refObjType, sysAuth)
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while getting context object")
	}

	return NewObjectContext(scopes, tenant, objID, objType), nil
}

func (m *mapperForSystemAuth) getTenantAndScopesForIntegrationSystem(reqData ReqData) (string, string, error) {
	var tenant, scopes string

	tenant, err := reqData.GetTenantID()
	if err != nil {
		return "", "", errors.Wrap(err, "while fetching tenant")
	}

	scopes, err = reqData.GetScopes()
	if err != nil {
		return "", "", errors.Wrap(err, "while fetching scopes")
	}

	return tenant, scopes, nil
}

func (m *mapperForSystemAuth) getTenantAndScopesForApplicationOrRuntime(sysAuth *model.SystemAuth, refObjType model.SystemAuthReferenceObjectType, reqData ReqData, authFlow AuthFlow) (string, string, error) {
	var tenant, scopes string
	hasTenant := true

	tenant, err := reqData.GetTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return "", "", errors.Wrap(err, "while fetching tenant")
		}

		hasTenant = false
	}

	if hasTenant && tenant != sysAuth.TenantID {
		return "", "", errors.New("tenant missmatch")
	}

	tenant = sysAuth.TenantID

	if authFlow.IsOAuth2Flow() {
		scopes, err = reqData.GetScopes()
		if err != nil {
			return "", "", errors.Wrap(err, "while fetching scopes")
		}
	}

	if authFlow.IsCertFlow() {
		declaredScopes, err := m.scopesGetter.GetRequiredScopes(buildPath(refObjType))
		if err != nil {
			return "", "", errors.Wrap(err, "while fetching scopes")
		}

		scopes = strings.Join(declaredScopes, " ")
	}

	return tenant, scopes, nil
}

func buildPath(refObjectType model.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(refObjectType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s.%s", clientCredentialScopesPrefix, transformedObjType)
}

func getContextObj(refObjType model.SystemAuthReferenceObjectType, sysAuth *model.SystemAuth) (string, string, error) {
	switch refObjType {
	case model.IntegrationSystemReference:
		return *sysAuth.IntegrationSystemID, string(model.IntegrationSystemReference), nil
	case model.RuntimeReference:
		return *sysAuth.RuntimeID, string(model.RuntimeReference), nil
	case model.ApplicationReference:
		return *sysAuth.AppID, string(model.ApplicationReference), nil
	default:
		return "", "", fmt.Errorf("unable to determine context details for object of type %s", refObjType)
	}
}
