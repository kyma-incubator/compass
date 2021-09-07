package tenantmapping

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/sirupsen/logrus"
)

// NewCertServiceContextProvider implements the ObjectContextProvider interface by looking for tenant information directly populated in the certificate.
func NewCertServiceContextProvider(tenantRepo TenantRepository) *certServiceContextProvider {
	return &certServiceContextProvider{
		tenantRepo: tenantRepo,
	}
}

type certServiceContextProvider struct {
	tenantRepo TenantRepository
}

// GetObjectContext is the certServiceContextProvider implementation of the ObjectContextProvider interface
// By using trusted external certificate issuer we assume that we will receive the tenant information extracted from the certificate.
// There we should only convert the tenant identifier from external to internal. Additionally, we mark the consumer in this flow as Runtime.
func (m *certServiceContextProvider) GetObjectContext(ctx context.Context, _ oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumer.Runtime,
	})

	ctx = log.ContextWithLogger(ctx, logger)

	externalTenantID := authDetails.AuthID

	/* TODO: Uncomment once we start storing subaccounts as tenants
	log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())

			log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
			return NewObjectContext(NewTenantContext(externalTenantID, ""), "", authDetails.AuthID, consumer.Runtime), nil
		}
		return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantID=%s]", externalTenantID)
	}

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), "", authDetails.AuthID, consumer.Runtime)

	*/

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, externalTenantID), "", authDetails.AuthID, consumer.Runtime)

	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)

	return objCtx, nil
}
