package tenantmapping

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"
)

// NewCertServiceContextProvider implements the ObjectContextProvider interface by looking for tenant information directly populated in the certificate.
func NewCertServiceContextProvider(tenantRepo TenantRepository, scopesGetter ScopesGetter, consumerExistsFuncs map[model.SystemAuthReferenceObjectType]func(context.Context, string) (bool, error)) *certServiceContextProvider {
	return &certServiceContextProvider{
		tenantRepo: tenantRepo,
		tenantKeys: KeysExtra{
			TenantKey:         ProviderTenantKey,
			ExternalTenantKey: ProviderExternalTenantKey,
		},
		scopesGetter:        scopesGetter,
		consumerExistsFuncs: consumerExistsFuncs,
	}
}

type certServiceContextProvider struct {
	tenantRepo TenantRepository
	tenantKeys KeysExtra

	scopesGetter ScopesGetter

	consumerExistsFuncs map[model.SystemAuthReferenceObjectType]func(context.Context, string) (bool, error)
}

// GetObjectContext is the certServiceContextProvider implementation of the ObjectContextProvider interface
// By using trusted external certificate issuer we assume that we will receive the tenant information extracted from the certificate.
// There we should only convert the tenant identifier from external to internal.
func (p *certServiceContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	// the authID in this flow is an OU selected by the Connector
	//TODO move in getExternalTenantID as last option
	//TODO this will match internal_consumer_id or subaccountid if internal_consumer_id is missing // this is either m.InternalConsumerID if set in mapping or last OU
	externalTenantID := authDetails.AuthID
	//TODO idea - authid regex ot each mapping (says where from the  subject to find the auth id)
	//TODO this way in authid we get the actual caller id and provider only flow will work for dest certs (no tenant header flow)
	//
	extraData := reqData.GetExtraDataWithDefaults()

	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": extraData.ConsumerType,
	})
	ctx = log.ContextWithLogger(ctx, logger)

	if extraData.AccessLevel != "" {
		var err error
		//TODO will fail if header is not provided ?? provider-flow subject mapping fails ?
		externalTenantID, err = reqData.GetExternalTenantID() // will return tenant ID from header if it exists
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Failed to get external tenant ID: %v", err)
			return ObjectContext{}, err
		}
	}

	scopes, err := p.directorScopes(reqData.GetConsumerTypeFromExtra())
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to get scopes for consumer type %s: %v", extraData.ConsumerType, err)
		return ObjectContext{}, apperrors.NewInternalError(fmt.Sprintf("Failed to extract scopes for consumer with type %s", extraData.ConsumerType))
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

	if extraData.AccessLevel != "" && tenantMapping.Type != extraData.AccessLevel {
		return ObjectContext{}, apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to %s with ID %s", authDetails.AuthID, extraData.AccessLevel, tenantMapping.ExternalTenant))
	}

	if extraData.InternalConsumerID != "" {
		found, err := p.consumerExistsFuncs[extraData.ConsumerType](ctx, extraData.InternalConsumerID)
		if err != nil {
			return ObjectContext{}, errors.Wrapf(err, "while getting %s with ID %s", extraData.ConsumerType, extraData.InternalConsumerID)
		}
		if !found {
			return ObjectContext{}, apperrors.NewUnauthorizedError(fmt.Sprintf("%s with ID %s does not exist", extraData.ConsumerType, extraData.InternalConsumerID))
		}
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), p.tenantKeys, scopes,
		//TODO consumerid is not authid but internal consumer id, authid should be put in tenant?
		authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.ConsumerType(extraData.ConsumerType), CertServiceObjectContextProvider)
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
	if consumerType == "" {
		consumerType = "default"
	}

	declaredScopes, err := p.scopesGetter.GetRequiredScopes(buildPath(consumerType))
	if err != nil {
		return "", errors.Wrap(err, "while fetching scopes")
	}
	return strings.Join(declaredScopes, " "), nil
}
