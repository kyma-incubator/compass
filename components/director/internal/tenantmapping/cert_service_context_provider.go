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
	tenantRepo          TenantRepository
	tenantKeys          KeysExtra
	scopesGetter        ScopesGetter
	consumerExistsFuncs map[model.SystemAuthReferenceObjectType]func(context.Context, string) (bool, error)
}

// GetObjectContext is the certServiceContextProvider implementation of the ObjectContextProvider interface
// By using trusted external certificate issuer we assume that we will receive the tenant information extracted from the certificate.
// There we should only convert the tenant identifier from external to internal.
func (p *certServiceContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	consumerType := reqData.ConsumerType()
	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumerType,
	})
	ctx = log.ContextWithLogger(ctx, logger)

	externalTenantID, err := getExternalTenant(ctx, reqData, authDetails)
	if err != nil {
		return ObjectContext{}, err
	}

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
			return NewObjectContext(NewTenantContext(externalTenantID, ""), p.tenantKeys, scopes, authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.Runtime, CertServiceObjectContextProvider), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	if reqData.IsIntegrationSystemFlow() {
		if err := p.verifyTenantAccess(ctx, tenantMapping, authDetails, reqData); err != nil {
			return ObjectContext{}, err
		}
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), p.tenantKeys, scopes,
		authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.ConsumerType(consumerType), CertServiceObjectContextProvider)
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

func getExternalTenant(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (string, error) {
	if reqData.ConsumerType() != model.IntegrationSystemReference {
		return authDetails.AuthID, nil // the authID in this flow is an OU selected by the Connector - matches a subaccount tenant
	}

	id, err := reqData.GetExternalTenantID()
	if err != nil {
		if !apperrors.IsKeyDoesNotExist(err) {
			log.C(ctx).WithError(err).Errorf("Failed to get external tenant ID: %v", err)
			return "", errors.Wrap(err, "while fetching tenant external id")
		}
		id = authDetails.AuthID
	}

	return id, nil
}

func (p *certServiceContextProvider) directorScopes(consumerType model.SystemAuthReferenceObjectType) (string, error) {
	declaredScopes, err := p.scopesGetter.GetRequiredScopes(buildPath(consumerType))
	if err != nil {
		return "", errors.Wrap(err, "while fetching scopes")
	}
	return strings.Join(declaredScopes, " "), nil
}

func (p *certServiceContextProvider) verifyTenantAccess(ctx context.Context, tenant *model.BusinessTenantMapping, authDetails oathkeeper.AuthDetails, reqData oathkeeper.ReqData) error {
	data := reqData.GetExtraDataWithDefaults()
	var accessLevelExists bool
	for _, al := range data.AccessLevels {
		if tenant.Type == al {
			accessLevelExists = true
			break
		}
	}

	if !accessLevelExists {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to %s tenant with ID %s", authDetails.AuthID, tenant.Type, tenant.ExternalTenant))
	}

	return p.verifyConsumerExists(ctx, data)
}

func (p *certServiceContextProvider) verifyConsumerExists(ctx context.Context, data oathkeeper.ExtraData) error {
	if data.InternalConsumerID == "" {
		return nil
	}

	found, err := p.consumerExistsFuncs[data.ConsumerType](ctx, data.InternalConsumerID)
	if err != nil {
		return errors.Wrapf(err, "while getting %s with ID %s", data.ConsumerType, data.InternalConsumerID)
	}
	if !found {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("%s with ID %s does not exist", data.ConsumerType, data.InternalConsumerID))
	}
	return nil
}
