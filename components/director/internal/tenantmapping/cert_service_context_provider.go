package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	oathkeeper2 "github.com/kyma-incubator/compass/components/director/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/systemauth"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
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

type certServiceContextProvider struct {
	tenantRepo   TenantRepository
	tenantKeys   KeysExtra
	scopesGetter ScopesGetter
}

// GetObjectContext is the certServiceContextProvider implementation of the ObjectContextProvider interface
// By using trusted external certificate issuer we assume that we will receive the tenant information extracted from the certificate.
// There we should only convert the tenant identifier from external to internal.
func (p *certServiceContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper2.ReqData, authDetails oathkeeper2.AuthDetails) (ObjectContext, error) {
	externalTenantID := authDetails.AuthID

	consumerType := reqData.ConsumerType()
	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumerType,
	})
	ctx = log.ContextWithLogger(ctx, logger)

	scopes, err := p.directorScopes(consumerType)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to get scopes for consumer type %s: %v", consumerType, err)
		return ObjectContext{}, apperrors.NewInternalError(fmt.Sprintf("Failed to extract scopes for consumer with type %s", consumerType))
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := p.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			// tenant not in DB yet, might be because we have not imported all subaccounts yet
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())
			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), p.tenantKeys, scopes, mergeWithOtherScopes, authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.Runtime, CertServiceObjectContextProvider), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), p.tenantKeys, scopes, mergeWithOtherScopes,
		authDetails.Region, "", getConsumerID(reqData, authDetails), authDetails.AuthFlow, consumer.ConsumerType(consumerType), CertServiceObjectContextProvider)
	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)
	return objCtx, nil
}

// Match checks if there is "client-id-from-certificate" Header with nonempty value and "client-certificate-issuer" Header with value "certificate-service".
// If so AuthDetails object is build.
func (p *certServiceContextProvider) Match(_ context.Context, data oathkeeper2.ReqData) (bool, *oathkeeper2.AuthDetails, error) {
	idVal := data.Body.Header.Get(oathkeeper2.ClientIDCertKey)
	certIssuer := data.Body.Header.Get(oathkeeper2.ClientIDCertIssuer)

	if idVal != "" && certIssuer == oathkeeper2.ExternalIssuer {
		return true, &oathkeeper2.AuthDetails{AuthID: idVal, AuthFlow: oathkeeper2.CertificateFlow, CertIssuer: certIssuer}, nil
	}

	return false, nil, nil
}

func (p *certServiceContextProvider) directorScopes(consumerType systemauth.SystemAuthReferenceObjectType) (string, error) {
	declaredScopes, err := p.scopesGetter.GetRequiredScopes(buildPath(consumerType))
	if err != nil {
		return "", errors.Wrap(err, "while fetching scopes")
	}
	return strings.Join(declaredScopes, " "), nil
}

func getConsumerID(data oathkeeper2.ReqData, details oathkeeper2.AuthDetails) string {
	if id := data.InternalConsumerID(); id != "" {
		return id
	}
	return details.AuthID
}
