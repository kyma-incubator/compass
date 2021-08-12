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
func (m *certServiceContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (ObjectContext, error) {
	logger := log.C(ctx).WithFields(logrus.Fields{
		"consumer_type": consumer.TechnicalCustomer,
	})

	ctx = log.ContextWithLogger(ctx, logger)

	externalTenantID := authDetails.AuthID

	// TODO: Uncomment once we start storing subaccounts as tenants
	//log.C(ctx).Infof("Getting the tenant with external ID: %s", externalTenantID)
	//tenantMapping, err := m.tenantRepo.GetByExternalTenant(ctx, externalTenantID)
	//if err != nil {
	//	if apperrors.IsNotFoundError(err) {
	//		log.C(ctx).Warningf("Could not find tenant with external ID: %s, error: %s", externalTenantID, err.Error())
	//
	//		log.C(ctx).Infof("Returning tenant context with empty internal tenant ID and external ID %s", externalTenantID)
	//		return NewObjectContext(NewTenantContext(externalTenantID, ""), "", authDetails.AuthID, consumer.TechnicalCustomer), nil
	//	}
	//	return ObjectContext{}, errors.Wrapf(err, "while getting external tenant mapping [ExternalTenantId=%s]", externalTenantID)
	//}
	//
	//objCtx := NewObjectContext(NewTenantContext(externalTenantID, tenantMapping.ID), "", authDetails.AuthID, consumer.TechnicalCustomer)

	objCtx := NewObjectContext(NewTenantContext(externalTenantID, externalTenantID), "", authDetails.AuthID, consumer.TechnicalCustomer)

	log.C(ctx).Infof("Successfully got object context: %+v", objCtx)

	return objCtx, nil
}
