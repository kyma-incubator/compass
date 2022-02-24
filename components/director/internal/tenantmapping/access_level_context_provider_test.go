package tenantmapping_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	oathkeeper2 "github.com/kyma-incubator/compass/components/director/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/systemauth"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccessLevelContextProvider_GetObjectContext(t *testing.T) {
	testError := errors.New("test error")
	notFoundErr := apperrors.NewNotFoundErrorWithType(resource.Tenant)
	tenantKeyNotFoundErr := apperrors.NewKeyDoesNotExistError(string(resource.Tenant))

	emptyCtx := context.TODO()
	consumerTenantID := uuid.New().String()
	providerTenantID := uuid.New().String()
	authDetails := oathkeeper2.AuthDetails{AuthID: providerTenantID, AuthFlow: oathkeeper2.CertificateFlow, CertIssuer: oathkeeper2.ExternalIssuer}

	reqData := oathkeeper2.ReqData{
		Body: oathkeeper2.ReqBody{
			Extra: map[string]interface{}{
				cert.ConsumerTypeExtraField:   systemauth.IntegrationSystemReference,
				cert.AccessLevelsExtraField:   []interface{}{tenantEntity.Subaccount},
				oathkeeper2.ExternalTenantKey: consumerTenantID,
			},
		},
	}

	internalSubaccount := "internalSubaccountID"

	testSubaccount := &model.BusinessTenantMapping{
		ID:             internalSubaccount,
		ExternalTenant: "externalSubaccount",
		Type:           tenantEntity.Subaccount,
	}
	testAccount := &model.BusinessTenantMapping{
		ID:             internalSubaccount,
		ExternalTenant: "externalAccount",
		Type:           tenantEntity.Account,
	}

	testCases := []struct {
		Name               string
		TenantRepoFn       func() *automock.TenantRepository
		ReqDataInput       oathkeeper2.ReqData
		AuthDetailsInput   oathkeeper2.AuthDetails
		ExpectedInternalID string
		ExpectedConsumerID string
		ExpectedErr        error
	}{
		{
			Name: "Success when cannot find internal tenant",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, consumerTenantID).Return(nil, notFoundErr).Once()
				return tenantRepo
			},
			ReqDataInput:       reqData,
			AuthDetailsInput:   authDetails,
			ExpectedInternalID: "",
			ExpectedErr:        nil,
		},
		{
			Name: "Error when the error from getting the internal tenant is different from not found",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, consumerTenantID).Return(nil, testError).Once()
				return tenantRepo
			},
			ReqDataInput:       reqData,
			AuthDetailsInput:   authDetails,
			ExpectedInternalID: "",
			ExpectedErr:        testError,
		},
		{
			Name: "Success when internal tenant exists",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, consumerTenantID).Return(testSubaccount, nil).Once()
				return tenantRepo
			},
			ReqDataInput:       reqData,
			AuthDetailsInput:   authDetails,
			ExpectedInternalID: internalSubaccount,
			ExpectedErr:        nil,
		},
		{
			Name:             "Error when can't extract external tenant id",
			TenantRepoFn:     unusedTenantRepo,
			ReqDataInput:     oathkeeper2.ReqData{},
			AuthDetailsInput: authDetails,
			ExpectedErr:      tenantKeyNotFoundErr,
		},
		{
			Name: "Error when consumer don't have access",
			TenantRepoFn: func() *automock.TenantRepository {
				tenantRepo := &automock.TenantRepository{}
				tenantRepo.On("GetByExternalTenant", mock.Anything, consumerTenantID).Return(testAccount, nil).Once()
				return tenantRepo
			},
			ReqDataInput:     reqData,
			AuthDetailsInput: authDetails,
			ExpectedErr:      apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to %s tenant with ID %s", authDetails.AuthID, testAccount.Type, testAccount.ExternalTenant)),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			tenantRepo := testCase.TenantRepoFn()
			provider := tenantmapping.NewAccessLevelContextProvider(tenantRepo)
			// WHEN
			objectCtx, err := provider.GetObjectContext(emptyCtx, testCase.ReqDataInput, testCase.AuthDetailsInput)

			// THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, consumer.IntegrationSystem, objectCtx.ConsumerType)
				require.Equal(t, providerTenantID, objectCtx.ConsumerID)
				require.Equal(t, testCase.ExpectedInternalID, objectCtx.TenantContext.TenantID)
				require.Equal(t, consumerTenantID, objectCtx.TenantContext.ExternalTenantID)
				require.Equal(t, "", objectCtx.Scopes)
			} else {
				require.Error(t, err)
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(testCase.ExpectedErr.Error()))
				require.Empty(t, objectCtx)
			}
			mock.AssertExpectationsForObjects(t, tenantRepo)
		})
	}
}

func TestAccessLevelContextProvider_Match(t *testing.T) {
	provider := tenantmapping.NewAccessLevelContextProvider(nil)
	clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
	tenantHeader := "123"
	accessLevels := []interface{}{"account"}
	t.Run("returns ID string and CertificateFlow when a client-id-from-certificate is specified in the Header map of request body", func(t *testing.T) {
		reqData := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ExternalTenantKey):  []string{tenantHeader},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertKey):    []string{clientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertIssuer): []string{oathkeeper2.ExternalIssuer},
				},
				Extra: map[string]interface{}{
					cert.AccessLevelsExtraField: accessLevels,
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper2.CertificateFlow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("does not match when consumer type is not provided and does not match", func(t *testing.T) {
		reqData := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ExternalTenantKey):  []string{tenantHeader},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertKey):    []string{clientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertIssuer): []string{oathkeeper2.ExternalIssuer},
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.NoError(t, err)
	})
	t.Run("does not match when cert issuer is not external and does not match", func(t *testing.T) {
		reqData := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ExternalTenantKey): []string{tenantHeader},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertKey):   []string{clientID},
				},
				Extra: map[string]interface{}{
					cert.AccessLevelsExtraField: accessLevels,
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.NoError(t, err)
	})
	t.Run("does not match when tenant header is missing", func(t *testing.T) {
		reqData := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertKey):    []string{clientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertIssuer): []string{oathkeeper2.ExternalIssuer},
				},
				Extra: map[string]interface{}{
					cert.ConsumerTypeExtraField: accessLevels,
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.NoError(t, err)
	})
}
