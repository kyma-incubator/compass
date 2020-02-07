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

func NewMapperForSystemAuth(systemAuthSvc systemauth.SystemAuthService, scopesGetter ScopesGetter, tenantRepo TenantRepository) *mapperForSystemAuth {
	return &mapperForSystemAuth{
		systemAuthSvc: systemAuthSvc,
		scopesGetter:  scopesGetter,
		tenantRepo:    tenantRepo,
	}
}

type mapperForSystemAuth struct {
	systemAuthSvc systemauth.SystemAuthService
	scopesGetter  ScopesGetter
	tenantRepo    TenantRepository
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
	var tenantCtx TenantContext

	switch refObjType {
	case model.IntegrationSystemReference:
		tenantCtx, scopes, err = m.getTenantAndScopesForIntegrationSystem(ctx, reqData)
	case model.RuntimeReference, model.ApplicationReference:
		tenantCtx, scopes, err = m.getTenantAndScopesForApplicationOrRuntime(ctx, sysAuth, refObjType, reqData, authFlow)
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

	return NewObjectContext(tenantCtx, scopes, refObjID, consumerType), nil
}

func (m *mapperForSystemAuth) getTenantAndScopesForIntegrationSystem(ctx context.Context, reqData ReqData) (TenantContext, string, error) {
	var externalTenantID, scopes string

	scopes, err := reqData.GetScopes()
	if err != nil {
		return TenantContext{}, scopes, errors.Wrap(err, "while fetching scopes")
	}

	externalTenantID, err = reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return TenantContext{}, scopes, errors.Wrap(err, "while fetching external tenant")
		}

		return TenantContext{}, scopes, nil
	}

	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		return TenantContext{}, scopes, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	return NewTenantContext(externalTenantID, tenantMapping.ID), scopes, nil
}

func (m *mapperForSystemAuth) getTenantAndScopesForApplicationOrRuntime(ctx context.Context, sysAuth *model.SystemAuth, refObjType model.SystemAuthReferenceObjectType, reqData ReqData, authFlow AuthFlow) (TenantContext, string, error) {
	var externalTenantID, scopes string
	var err error

	if sysAuth.TenantID == nil {
		return TenantContext{}, scopes, errors.New("system auth tenant id cannot be nil")
	}

	if authFlow.IsOAuth2Flow() {
		scopes, err = reqData.GetScopes()
		if err != nil {
			return TenantContext{}, scopes, errors.Wrap(err, "while fetching scopes")
		}
	}

	if authFlow.IsCertFlow() || authFlow.IsOneTimeTokenFlow() {
		declaredScopes, err := m.scopesGetter.GetRequiredScopes(buildPath(refObjType))
		if err != nil {
			return TenantContext{}, scopes, errors.Wrap(err, "while fetching scopes")
		}

		scopes = strings.Join(declaredScopes, " ")
	}

	externalTenantID, err = reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return TenantContext{}, scopes, errors.Wrap(err, "while fetching tenant")
		}

		return NewTenantContext(externalTenantID, *sysAuth.TenantID), scopes, nil
	}

	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		return TenantContext{}, scopes, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	if tenantMapping.ID != *sysAuth.TenantID {
		return TenantContext{}, scopes, errors.New("tenant mismatch")
	}

	return NewTenantContext(externalTenantID, *sysAuth.TenantID), scopes, nil
}

func buildPath(refObjectType model.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(refObjectType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s.%s", clientCredentialScopesPrefix, transformedObjType)
}
