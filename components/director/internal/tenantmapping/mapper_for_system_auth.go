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

func (m *mapperForSystemAuth) GetTenantAndScopes(ctx context.Context, reqData ReqData, authID string, authFlow AuthFlow) (string, string, error) {
	sysAuth, err := m.systemAuthSvc.GetGlobal(ctx, authID)
	if err != nil {
		return "", "", errors.Wrap(err, "while retrieving system auth from database")
	}

	refObjType, err := sysAuth.GetReferenceObjectType()
	if err != nil {
		return "", "", errors.Wrap(err, "while getting reference object type")
	}

	var tenantAndScopesFunc func(sysAuth *model.SystemAuth, refObjType model.SystemAuthReferenceObjectType, reqData ReqData, authFlow AuthFlow) (string, string, error)
	switch refObjType {
	case model.IntegrationSystemReference:
		tenantAndScopesFunc = m.getTenantAndScopesForIntegrationSystem
		break
	case model.RuntimeReference, model.ApplicationReference:
		tenantAndScopesFunc = m.getTenantAndScopesForApplicationOrRuntime
		break
	}

	return tenantAndScopesFunc(sysAuth, refObjType, reqData, authFlow)
}

func (m *mapperForSystemAuth) getTenantAndScopesForIntegrationSystem(sysAuth *model.SystemAuth, refObjType model.SystemAuthReferenceObjectType, reqData ReqData, authFlow AuthFlow) (string, string, error) {
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
		return "", "", fmt.Errorf("tenant missmatch")
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
