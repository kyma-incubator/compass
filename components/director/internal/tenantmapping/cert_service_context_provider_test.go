package tenantmapping_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/sirupsen/logrus"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	tenantmappingmock "github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/stretchr/testify/require"
)

var internalTenant = "internalTenantID"

func TestCertServiceContextProvider(t *testing.T) {
	testError := errors.New("test error")
	notFoundErr := apperrors.NewNotFoundErrorWithType(resource.Tenant)
	emptyCtx := context.TODO()
	logger := log.C(emptyCtx).WithFields(logrus.Fields{"consumer_type": consumer.Runtime})
	ctxWithLogger := log.ContextWithLogger(emptyCtx, logger)
	subaccount := uuid.New().String()
	authDetails := oathkeeper.AuthDetails{AuthID: subaccount, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ExternalIssuer}
	directorScopes := "runtime:read runtime:write tenant:read"
	tenantRepo := &tenantmappingmock.TenantRepository{}
	provider := tenantmapping.NewCertServiceContextProvider(tenantRepo)

	testTenant := &model.BusinessTenantMapping{
		ID:             internalTenant,
		Name:           "testTenant",
		ExternalTenant: "externalTestTenant",
		Type:           "subaccount",
	}

	t.Run("Success when cannot find internal tenant", func(t *testing.T) {
		tenantRepo.On("GetByExternalTenant", ctxWithLogger, subaccount).Return(nil, notFoundErr).Once()
		provider = tenantmapping.NewCertServiceContextProvider(tenantRepo)

		objectCtx, err := provider.GetObjectContext(emptyCtx, oathkeeper.ReqData{}, authDetails)
		require.NoError(t, err)
		require.Empty(t, objectCtx.TenantContext.TenantID)
		require.Equal(t, subaccount, objectCtx.TenantContext.ExternalTenantID)
		require.Equal(t, subaccount, objectCtx.ConsumerID)
		require.Equal(t, directorScopes, objectCtx.Scopes)

		tenantRepo.AssertExpectations(t)
	})

	t.Run("Error when the error is different from not found", func(t *testing.T) {
		tenantRepo.On("GetByExternalTenant", ctxWithLogger, subaccount).Return(nil, testError).Once()
		provider = tenantmapping.NewCertServiceContextProvider(tenantRepo)

		objectCtx, err := provider.GetObjectContext(emptyCtx, oathkeeper.ReqData{}, authDetails)
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("while getting external tenant mapping [ExternalTenantID=%s]", subaccount))
		require.Empty(t, objectCtx)

		tenantRepo.AssertExpectations(t)
	})

	t.Run("Success when internal tenant exists", func(t *testing.T) {
		tenantRepo.On("GetByExternalTenant", ctxWithLogger, subaccount).Return(testTenant, nil).Once()
		provider = tenantmapping.NewCertServiceContextProvider(tenantRepo)

		objectCtx, err := provider.GetObjectContext(emptyCtx, oathkeeper.ReqData{}, authDetails)

		require.NoError(t, err)
		require.Equal(t, consumer.Runtime, objectCtx.ConsumerType)
		require.Equal(t, subaccount, objectCtx.ConsumerID)
		require.Equal(t, internalTenant, objectCtx.TenantContext.TenantID)
		require.Equal(t, subaccount, objectCtx.TenantContext.ExternalTenantID)
		require.Equal(t, directorScopes, objectCtx.Scopes)

		tenantRepo.AssertExpectations(t)
	})
}

func TestCertServiceContextProviderMatch(t *testing.T) {
	t.Run("returns ID string and CertificateFlow when a client-id-from-certificate is specified in the Header map of request body", func(t *testing.T) {
		clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
		provider := tenantmapping.NewCertServiceContextProvider(nil)

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{clientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.CertificateFlow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("returns nil when does not match", func(t *testing.T) {
		provider := tenantmapping.NewCertServiceContextProvider(nil)

		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.NoError(t, err)
	})
}
