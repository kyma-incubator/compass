package tenantmapping

import (
	"context"
	"fmt"

	directorErrors "github.com/kyma-incubator/compass/components/hydrator/internal/director"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type accessLevelContextProvider struct {
	directorClient DirectorClient
	tenantKeys     KeysExtra
}

// NewAccessLevelContextProvider implements the ObjectContextProvider interface by looking tenant header and access levels defined in the auth session extra.
func NewAccessLevelContextProvider(clientProvider DirectorClient) *accessLevelContextProvider {
	return &accessLevelContextProvider{
		directorClient: clientProvider,
		tenantKeys: KeysExtra{
			TenantKey:         tenantmapping.ConsumerTenantKey,
			ExternalTenantKey: tenantmapping.ExternalTenantKey,
		},
	}
}

// GetObjectContext is the accessLevelContextProvider implementation of the ObjectContextProvider interface.
// It receives the tenant information extracted from the tenant header in this case, and it verifies that the caller has access to this tenant.
func (p *accessLevelContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	consumerType := reqData.ConsumerType()
	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumerType,
	})
	ctx = log.ContextWithLogger(ctx, logger)

	externalTenantID, err := reqData.GetExternalTenantID()
	if err != nil {
		return ObjectContext{}, err
	}

	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := p.directorClient.GetTenantByExternalID(ctx, externalTenantID)
	if err != nil {
		if directorErrors.IsGQLNotFoundError(err) {
			// tenant not in DB yet, might be because we have not imported all subaccounts yet
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())
			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), p.tenantKeys, "", mergeWithOtherScopes, authDetails.Region,
				"", authDetails.AuthID, authDetails.AuthFlow, consumer.ConsumerType(consumerType), tenantmapping.CertServiceObjectContextProvider), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	if err := p.verifyTenantAccessLevels(tenantMapping.Type, authDetails, reqData); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to verify tenant access level: %v", err)
		return ObjectContext{}, err
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.InternalID), p.tenantKeys, "", mergeWithOtherScopes,
		authDetails.Region, "", authDetails.AuthID, authDetails.AuthFlow, consumer.ConsumerType(consumerType), tenantmapping.CertServiceObjectContextProvider)
	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)
	return objCtx, nil
}

// Match will only match requests coming from certificate consumers that are able to access only a subset of tenants.
func (p *accessLevelContextProvider) Match(_ context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error) {
	if len(data.TenantAccessLevels()) == 0 {
		return false, nil, nil
	}

	idVal := data.Body.Header.Get(oathkeeper.ClientIDCertKey)
	certIssuer := data.Body.Header.Get(oathkeeper.ClientIDCertIssuer)

	if idVal == "" || certIssuer != oathkeeper.ExternalIssuer {
		return false, nil, nil
	}

	if _, err := data.GetExternalTenantID(); err != nil {
		if apperrors.IsKeyDoesNotExist(err) {
			return false, nil, nil
		}
		return false, nil, err
	}

	return true, &oathkeeper.AuthDetails{AuthID: idVal, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: certIssuer}, nil
}

func (p *accessLevelContextProvider) verifyTenantAccessLevels(tenantType string, authDetails oathkeeper.AuthDetails, reqData oathkeeper.ReqData) error {
	grantedAccessLevels := reqData.TenantAccessLevels()
	var accessLevelExists bool
	for _, al := range grantedAccessLevels {
		if tenantEntity.Type(tenantType) == al {
			accessLevelExists = true
			break
		}
	}

	externalTenantID, err := reqData.GetExternalTenantID()
	if err != nil {
		return err
	}

	if !accessLevelExists {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to %s tenant with ID %s", authDetails.AuthID, tenantType, externalTenantID))
	}

	return nil
}
