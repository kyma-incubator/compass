package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

func NewSystemAuthContextProvider(systemAuthSvc systemauth.SystemAuthService, scopesGetter ScopesGetter, tenantRepo TenantRepository) *systemAuthContextProvider {
	return &systemAuthContextProvider{
		systemAuthSvc: systemAuthSvc,
		scopesGetter:  scopesGetter,
		tenantRepo:    tenantRepo,
	}
}

type systemAuthContextProvider struct {
	systemAuthSvc systemauth.SystemAuthService
	scopesGetter  ScopesGetter
	tenantRepo    TenantRepository
}

func (m *systemAuthContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	sysAuth, err := m.systemAuthSvc.GetGlobal(ctx, authDetails.AuthID)
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while retrieving system auth from database")
	}

	refObjType, err := sysAuth.GetReferenceObjectType()
	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while getting reference object type for system auth id %s", sysAuth.ID)
	}
	log.C(ctx).Infof("Reference object type is %s", refObjType)

	var scopes string
	var tenantCtx TenantContext

	switch refObjType {
	case model.IntegrationSystemReference:
		tenantCtx, scopes, err = m.getTenantAndScopesForIntegrationSystem(ctx, reqData)
	case model.RuntimeReference, model.ApplicationReference:
		tenantCtx, scopes, err = m.getTenantAndScopesForApplicationOrRuntime(ctx, sysAuth, refObjType, reqData, authDetails.AuthFlow)
	default:
		return ObjectContext{}, errors.Errorf("unsupported reference object type (%s)", refObjType)
	}

	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while fetching the tenant and scopes for system auth with id: %s, object type: %s, using auth flow: %s", sysAuth.ID, refObjType, authDetails.AuthFlow)
	}
	log.C(ctx).Debugf("Successfully got tenant context - external ID: %s, internal ID: %s", tenantCtx.ExternalTenantID, tenantCtx.TenantID)

	refObjID, err := sysAuth.GetReferenceObjectID()
	if err != nil {
		return ObjectContext{}, errors.Wrap(err, "while getting reference object id")
	}
	log.C(ctx).Debugf("Reference object id is %s", refObjID)

	consumerType, err := consumer.MapSystemAuthToConsumerType(refObjType)
	if err != nil {
		return ObjectContext{}, apperrors.NewInternalError("while mapping reference type to consumer type")
	}

	objCxt := NewObjectContext(tenantCtx, scopes, refObjID, consumerType)
	log.C(ctx).Infof("Object context: %+v", objCxt)

	return objCxt, nil
}

func (m *systemAuthContextProvider) getTenantAndScopesForIntegrationSystem(ctx context.Context, reqData oathkeeper.ReqData) (TenantContext, string, error) {
	var externalTenantID, scopes string

	scopes, err := reqData.GetScopes()
	if err != nil {
		return TenantContext{}, scopes, errors.Wrap(err, "while fetching scopes")
	}
	log.C(ctx).Debugf("Found scopes are: %s", scopes)

	externalTenantID, err = reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return TenantContext{}, scopes, errors.Wrap(err, "while fetching tenant external id")
		}
		log.C(ctx).Warningf("Could not get tenant external id, error: %s", err.Error())

		log.C(ctx).Info("Could not create tenant context, returning empty context...")
		return TenantContext{}, scopes, nil
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Warningf("Could not find external tenant with ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Info("Returning tenant context with empty internal tenant ID...")
			return NewTenantContext(externalTenantID, ""), scopes, nil
		}
		return TenantContext{}, scopes, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	return NewTenantContext(externalTenantID, tenantMapping.ID), scopes, nil
}

func (m *systemAuthContextProvider) getTenantAndScopesForApplicationOrRuntime(ctx context.Context, sysAuth *model.SystemAuth, refObjType model.SystemAuthReferenceObjectType, reqData oathkeeper.ReqData, authFlow oathkeeper.AuthFlow) (TenantContext, string, error) {
	var externalTenantID, scopes string
	var err error

	if sysAuth.TenantID == nil {
		return TenantContext{}, scopes, apperrors.NewInternalError("system auth tenant id cannot be nil")
	}
	log.C(ctx).Infof("Internal tenant id is %s", *sysAuth.TenantID)

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
	log.C(ctx).Debugf("Found scopes are: %s", scopes)

	externalTenantID, err = reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			return TenantContext{}, scopes, errors.Wrap(err, "while fetching tenant external id")
		}
		log.C(ctx).Warningf("Could not get tenant external id, error: %s", err.Error())

		log.C(ctx).Infof("Returning context with empty external tenant ID and internal tenant id: %s", *sysAuth.TenantID)
		return NewTenantContext("", *sysAuth.TenantID), scopes, nil
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Warningf("Could not find external tenant with ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Info("Returning tenant context with empty internal tenant ID...")
			return NewTenantContext(externalTenantID, ""), scopes, nil
		}
		return TenantContext{}, scopes, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	}

	if tenantMapping.ID != *sysAuth.TenantID {
		log.C(ctx).Errorf("Tenant mismatch - tenant id %s and system auth tenant id %s, for object of type %s", tenantMapping.ID, *sysAuth.TenantID, refObjType)
		return NewTenantContext(externalTenantID, ""), scopes, nil
	}

	return NewTenantContext(externalTenantID, *sysAuth.TenantID), scopes, nil
}

func buildPath(refObjectType model.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(refObjectType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s.%s", clientCredentialScopesPrefix, transformedObjType)
}
