package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"
)

// NewCertServiceContextProvider implements the ObjectContextProvider interface by looking for tenant information directly populated in the certificate.
func NewCertServiceContextProvider(tenantRepo TenantRepository, scopesGetter ScopesGetter) *certServiceContextProvider {
	return &certServiceContextProvider{
		tenantRepo: tenantRepo,
		tenantKeys: KeysExtra{
			TenantKey:         ProviderTenantKey,
			ExternalTenantKey: ProviderExternalTenantKey,
		},
		scopesGetter: scopesGetter,
	}
}

// TODO include integrationsyssvc
type certServiceContextProvider struct {
	tenantRepo TenantRepository
	tenantKeys KeysExtra

	scopesGetter ScopesGetter
}

// GetObjectContext is the certServiceContextProvider implementation of the ObjectContextProvider interface
// By using trusted external certificate issuer we assume that we will receive the tenant information extracted from the certificate.
// There we should only convert the tenant identifier from external to internal. Additionally, we mark the consumer in this flow as Runtime.
func (p *certServiceContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	matchedComponentName, ok := reqData.Body.Header[authenticator.ComponentName]
	if !ok || len(matchedComponentName) == 0 {
		return ObjectContext{}, errors.New("empty matched component header")
	}

	// This if is needed to separate the director from ord flow because for the director flow we need to use the internal ID of the subaccount
	// whereas in the ord flow we expect external IDs in order ord views to work properly(using Automatic Scenario Assignments)
	log.C(ctx).Infof("Matched component name is %s", matchedComponentName[0])

	if matchedComponentName[0] != "director" { // ORD Flow, set the external tenant ID both for internal and external tenants
		externalTenantID := authDetails.AuthID
		objCtx := NewObjectContext(NewTenantContext(externalTenantID, externalTenantID), p.tenantKeys, "", authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.Runtime, CertServiceObjectContextProvider)
		log.C(ctx).Infof("Successfully got object context: %+v", objCtx)
		return objCtx, nil
	}

	consumerType := reqData.GetConsumerTypeExtraFieldFromExtra()
	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumerType,
	})

	ctx = log.ContextWithLogger(ctx, logger)
	scopes, err := p.directorScopes(consumerType)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to get scopes for consumer type %s: %v", consumerType, err)
		return ObjectContext{}, apperrors.NewInternalError("failed to extract scopes") // TODO improve msg
	}

	externalTenantID, err := reqData.GetExternalTenantID()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to get external tenant ID: %v", err)
		return ObjectContext{}, apperrors.NewInternalError("failed to extract external tenant") // TODO improve msg
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := p.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			// tenant not in DB yet, might be because we have not imported all subaccounts yet
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())
			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), p.tenantKeys, scopes, authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.Runtime, CertServiceObjectContextProvider), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	accessLvl := reqData.GetAccessLevelFromExtra()
	if accessLvl != "" && tenantMapping.Type != accessLvl {
		// TODO improve message
		return ObjectContext{}, apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to tenant with ID %s", authDetails.AuthID, tenantMapping.ExternalTenant))
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), p.tenantKeys, scopes, authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.Runtime, CertServiceObjectContextProvider)
	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)
	return objCtx, nil
}

// Match checks if there is "client-id-from-certificate" Header with nonempty value and "client-certificate-issuer" Header with value "certificate-service".
// If so AuthDetails object is build.
func (p *certServiceContextProvider) Match(_ context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error) {
	idVal := data.Body.Header.Get(oathkeeper.ClientIDCertKey)
	certIssuer := data.Body.Header.Get(oathkeeper.ClientIDCertIssuer)

	if idVal != "" && certIssuer == oathkeeper.ExternalIssuer {
		return true, &oathkeeper.AuthDetails{AuthID: idVal, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: certIssuer}, nil
	}

	return false, nil, nil
}

func (p *certServiceContextProvider) directorScopes(consumerType model.SystemAuthReferenceObjectType) (string, error) {
	if consumerType == "" { // TODO improve
		consumerType = model.SystemAuthReferenceObjectType("default")
	}

	declaredScopes, err := p.scopesGetter.GetRequiredScopes(buildPath(consumerType))
	if err != nil {
		return "", errors.Wrap(err, "while fetching scopes")
	}
	return strings.Join(declaredScopes, " "), nil
}
