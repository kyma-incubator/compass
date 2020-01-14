package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

func NewMapperForSystemAuth(systemAuthSvc systemauth.SystemAuthService, scopesGetter ScopesGetter, tenantStorageService TenantStorageService) *mapperForSystemAuth {
	return &mapperForSystemAuth{
		systemAuthSvc:        systemAuthSvc,
		scopesGetter:         scopesGetter,
		tenantStorageService: tenantStorageService,
	}
}

type mapperForSystemAuth struct {
	systemAuthSvc        systemauth.SystemAuthService
	scopesGetter         ScopesGetter
	tenantStorageService TenantStorageService
}

func (m *mapperForSystemAuth) GetObjectContext(ctx context.Context, reqData ReqData, authID string, authFlow AuthFlow) (ObjectContext, error) {
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
		tenant, scopes, err = m.getTenantAndScopesForIntegrationSystem(ctx, reqData)
	case model.RuntimeReference, model.ApplicationReference:
		tenant, scopes, err = m.getTenantAndScopesForApplicationOrRuntime(ctx, sysAuth, refObjType, reqData, authFlow)
	default:
		return ObjectContext{}, errors.Errorf("unsupported reference object type (%s)", refObjType)
	}

	if err != nil {
		return ObjectContext{}, errors.Wrap(err, fmt.Sprintf("while fetching the tenant and scopes for object of type %s", refObjType))
	}

	refObjID, err := sysAuth.GetReferenceObjectID()
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while getting context object")
	}

	consumerType, err := consumer.MapSystemAuthToConsumerType(refObjType)
	if err != nil {
		return ObjectContext{}, errors.New("while mapping reference type to consumer type")
	}

	return NewObjectContext(scopes, tenant, refObjID, consumerType), nil
}

func (m *mapperForSystemAuth) getTenantAndScopesForIntegrationSystem(ctx context.Context, reqData ReqData) (string, string, error) {
	var externalTenant, scopes string

	externalTenant, err := reqData.GetExternalTenantID()
	if err != nil {
		return "", "", errors.Wrap(err, "while fetching tenant")
	}

	internalTenant, err := m.tenantStorageService.GetInternalTenant(ctx, externalTenant)
	if err != nil {
		return "", "", errors.Wrap(err, "while mapping external to internal tenant")
	}

	scopes, err = reqData.GetScopes()
	if err != nil {
		return "", "", errors.Wrap(err, "while fetching scopes")
	}

	return internalTenant, scopes, nil
}

func (m *mapperForSystemAuth) getTenantAndScopesForApplicationOrRuntime(ctx context.Context, sysAuth *model.SystemAuth, refObjType model.SystemAuthReferenceObjectType, reqData ReqData, authFlow AuthFlow) (string, string, error) {
	var externalTenant, scopes string
	hasTenant := true

	externalTenant, err := reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return "", "", errors.Wrap(err, "while fetching tenant")
		}

		hasTenant = false
	}

	internalTenant, err := m.tenantStorageService.GetInternalTenant(ctx, externalTenant)
	if err != nil {
		return "", "", errors.Wrap(err, "while mapping external to internal tenant")
	}

	if hasTenant && internalTenant != sysAuth.TenantID {
		return "", "", errors.New("tenant missmatch")
	}

	internalTenant = sysAuth.TenantID

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

	return internalTenant, scopes, nil
}

func buildPath(refObjectType model.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(refObjectType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s.%s", clientCredentialScopesPrefix, transformedObjType)
}
