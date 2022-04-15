package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	directorErrors "github.com/kyma-incubator/compass/components/hydrator/internal/director"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

// NewSystemAuthContextProvider missing godoc
func NewSystemAuthContextProvider(clientProvider DirectorClient, scopesGetter ScopesGetter) *systemAuthContextProvider {
	return &systemAuthContextProvider{
		scopesGetter:   scopesGetter,
		directorClient: clientProvider,
		tenantKeys: KeysExtra{
			TenantKey:         tenantmapping.ConsumerTenantKey,
			ExternalTenantKey: tenantmapping.ExternalTenantKey,
		},
	}
}

type systemAuthContextProvider struct {
	scopesGetter   ScopesGetter
	tenantKeys     KeysExtra
	directorClient DirectorClient
}

// GetObjectContext missing godoc
func (m *systemAuthContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	sysAuth, opErr := m.directorClient.GetSystemAuthByID(ctx, authDetails.AuthID)
	if opErr != nil || sysAuth == nil {
		return ObjectContext{}, errors.Wrap(opErr, "while retrieving system auth from director")
	}

	refObjectType, err := sysAuth.GetReferenceObjectType()
	if err != nil {
		return ObjectContext{}, errors.Errorf("unknown reference object type for system auth with id %s", sysAuth.ID)
	}

	refObjectID, err := sysAuth.GetReferenceObjectID()
	if err != nil {
		return ObjectContext{}, errors.Errorf("unknown reference object ID for system auth with id %s", sysAuth.ID)
	}

	log.C(ctx).Debugf("Reference object id is %s", refObjectID)
	log.C(ctx).Infof("Reference object type is %s", refObjectType)

	if authDetails.AuthFlow.IsCertFlow() && sysAuth.Value != nil && sysAuth.Value.CertCommonName != authDetails.AuthID {
		sysAuth.Value.OneTimeToken = nil
		sysAuth.Value.CertCommonName = authDetails.AuthID

		if _, err := m.directorClient.UpdateSystemAuth(ctx, sysAuth); err != nil {
			return ObjectContext{}, errors.Wrap(err, "while updating system auth")
		}
	}

	var scopes string
	var tenantCtx TenantContext

	switch refObjectType {
	case model.IntegrationSystemReference:
		tenantCtx, scopes, err = m.getTenantAndScopesForIntegrationSystem(ctx, reqData)
	case model.ApplicationReference, model.RuntimeReference:
		tenantCtx, scopes, err = m.getTenantAndScopesForApplicationOrRuntime(ctx, sysAuth.TenantID, refObjectType, reqData, authDetails.AuthFlow)
	default:
		return ObjectContext{}, errors.Errorf("unsupported reference object type (%s)", refObjectType)
	}

	if err != nil {
		return ObjectContext{}, errors.Wrapf(err, "while fetching the tenant and scopes for system auth with id: %s, object type: %s, using auth flow: %s", sysAuth.ID, refObjectType, authDetails.AuthFlow)
	}
	log.C(ctx).Debugf("Successfully got tenant context - external ID: %s, internal ID: %s", tenantCtx.ExternalTenantID, tenantCtx.TenantID)

	consumerType, err := consumer.MapSystemAuthToConsumerType(refObjectType)
	if err != nil {
		return ObjectContext{}, apperrors.NewInternalError("while mapping reference type to consumer type")
	}

	objCxt := NewObjectContext(tenantCtx, m.tenantKeys, scopes, intersectWithOtherScopes, authDetails.Region, "", refObjectID, authDetails.AuthFlow, consumerType, tenantmapping.SystemAuthObjectContextProvider)
	log.C(ctx).Infof("Object context: %+v", objCxt)

	return objCxt, nil
}

func (m *systemAuthContextProvider) Match(_ context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error) {
	// Custom authenticator flow:
	// If that key is set, then the request has already passed by the authenticator mapping handler,
	// hence the context provider will be the one of that particular authenticator.
	if _, ok := data.Body.Extra[authenticator.CoordinatesKey]; ok {
		return false, nil, nil
	}

	// Certificate flow
	idVal := data.Body.Header.Get(oathkeeper.ClientIDCertKey)
	certIssuer := data.Body.Header.Get(oathkeeper.ClientIDCertIssuer)

	if idVal != "" && certIssuer != oathkeeper.ExternalIssuer {
		return true, &oathkeeper.AuthDetails{AuthID: idVal, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: certIssuer}, nil
	}

	// One-Time Token flow
	if idVal := data.Body.Header.Get(oathkeeper.ClientIDTokenKey); idVal != "" {
		return true, &oathkeeper.AuthDetails{AuthID: idVal, AuthFlow: oathkeeper.OneTimeTokenFlow}, nil
	}

	// Hydra Client Credentials OAuth flow
	if idVal, ok := data.Body.Extra[oathkeeper.ClientIDKey]; ok {
		authID, err := str.Cast(idVal)
		if err != nil {
			return false, nil, errors.Wrapf(err, "while parsing the value for %s", oathkeeper.ClientIDKey)
		}

		return true, &oathkeeper.AuthDetails{AuthID: authID, AuthFlow: oathkeeper.OAuth2Flow}, nil
	}

	return false, nil, nil
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
	tenantMapping, err := m.directorClient.GetTenantByExternalID(ctx, externalTenantID)
	if err != nil {
		if directorErrors.IsGQLNotFoundError(err) {
			log.C(ctx).Warningf("Could not find external tenant with ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewTenantContext(externalTenantID, ""), scopes, nil
		}
		return TenantContext{}, scopes, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	return NewTenantContext(externalTenantID, tenantMapping.InternalID), scopes, nil
}

func (m *systemAuthContextProvider) getTenantAndScopesForApplicationOrRuntime(ctx context.Context, tenantID *string, refObjType model.SystemAuthReferenceObjectType, reqData oathkeeper.ReqData, authFlow oathkeeper.AuthFlow) (TenantContext, string, error) {
	var externalTenantID, scopes string
	var err error

	if tenantID == nil {
		return TenantContext{}, scopes, apperrors.NewInternalError("system auth tenant id cannot be nil")
	}
	log.C(ctx).Infof("Internal tenant id is %s", *tenantID)

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

		log.C(ctx).Infof("Returning context with empty external tenant ID and internal tenant id: %s", *tenantID)
		return NewTenantContext("", *tenantID), scopes, nil
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := m.directorClient.GetTenantByExternalID(ctx, externalTenantID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Warningf("Could not find external tenant with ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Info("Returning tenant context with empty internal tenant ID...")
			return NewTenantContext(externalTenantID, ""), scopes, nil
		}
		return TenantContext{}, scopes, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	if tenantMapping.InternalID != *tenantID {
		log.C(ctx).Errorf("Tenant mismatch - tenant id %s and system auth tenant id %s, for object of type %s", tenantMapping.ID, *tenantID, refObjType)
		return NewTenantContext(externalTenantID, ""), scopes, nil
	}

	return NewTenantContext(externalTenantID, *tenantID), scopes, nil
}

func buildPath(refObjectType model.SystemAuthReferenceObjectType) string {
	lowerCaseType := strings.ToLower(string(refObjectType))
	transformedObjType := strings.ReplaceAll(lowerCaseType, " ", "_")
	return fmt.Sprintf("%s.%s", scopesPerConsumerTypePrefix, transformedObjType)
}
