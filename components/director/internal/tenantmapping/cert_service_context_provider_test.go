package tenantmapping_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	tenantmappingmock "github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/stretchr/testify/require"
)

func TestCertServiceContextProvider(t *testing.T) {
	subaccount := uuid.New().String()
	authDetails := oathkeeper.AuthDetails{AuthID: subaccount, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ExternalIssuer}

	tenantRepo := &tenantmappingmock.TenantRepository{}
	provider := tenantmapping.NewCertServiceContextProvider(tenantRepo)

	objectCtx, err := provider.GetObjectContext(context.TODO(), oathkeeper.ReqData{}, authDetails)
	require.NoError(t, err)

	require.Equal(t, consumer.TechnicalCustomer, objectCtx.ConsumerType)
	require.Equal(t, subaccount, objectCtx.ConsumerID)
	require.Equal(t, subaccount, objectCtx.TenantContext.TenantID)
	require.Equal(t, subaccount, objectCtx.TenantContext.ExternalTenantID)
	require.Empty(t, objectCtx.Scopes)

	tenantRepo.AssertExpectations(t)
}
