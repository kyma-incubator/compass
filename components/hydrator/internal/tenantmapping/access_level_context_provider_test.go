package tenantmapping_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAccessLevelContextProvider_GetObjectContext(t *testing.T) {
	emptyCtx := context.TODO()
	consumerTenantID := "3944c2f9-f614-4680-b4a9-0f07315bc982"
	providerTenantID := uuid.New().String()

	testError := errors.New("test error")
	notFoundErr := apperrors.NewNotFoundErrorWithType(resource.Tenant)

	authDetails := oathkeeper.AuthDetails{AuthID: providerTenantID, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ExternalIssuer}
	scopes := []string{"runtime:read", "runtime:write", "tenant:read"}

	reqData := oathkeeper.ReqData{
		Body: oathkeeper.ReqBody{
			Extra: map[string]interface{}{
				cert.ConsumerTypeExtraField:  model.IntegrationSystemReference,
				cert.AccessLevelsExtraField:  []interface{}{tenantEntity.Subaccount},
				oathkeeper.ExternalTenantKey: consumerTenantID,
			},
		},
	}

	noTenantReqData := oathkeeper.ReqData{
		Body: oathkeeper.ReqBody{
			Extra: map[string]interface{}{
				cert.ConsumerTypeExtraField: model.IntegrationSystemReference,
				cert.AccessLevelsExtraField: []interface{}{tenantmapping.GlobalAccessLevel},
			},
		},
	}

	internalSubaccount := "internalSubaccountID"
	region := "eu-1"
	labels := map[string]interface{}{
		"region": region,
	}

	testSubaccount := &graphql.Tenant{
		ID:         consumerTenantID,
		InternalID: internalSubaccount,
		Type:       "subaccount",
		Labels:     labels,
	}

	testAccount := &graphql.Tenant{
		ID:         internalSubaccount,
		InternalID: "externalAccount",
		Type:       "account",
		Labels:     labels,
	}

	testEmptyTenant := &graphql.Tenant{}

	testTenantWithOnlyExternalID := &graphql.Tenant{ID: consumerTenantID}

	testCases := []struct {
		Name             string
		DirectorClient   func() *automock.DirectorClient
		ScopesGetterFn   func() *automock.ScopesGetter
		ReqDataInput     oathkeeper.ReqData
		AuthDetailsInput oathkeeper.AuthDetails
		ExpectedScopes   string
		ExpectedTenant   *graphql.Tenant
		ExpectedErr      error
	}{
		{
			Name: "Success when cannot find internal tenant",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(nil, notFoundErr).Once()
				return client
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:     reqData,
			AuthDetailsInput: authDetails,
			ExpectedTenant:   testTenantWithOnlyExternalID,
			ExpectedScopes:   strings.Join(scopes, " "),
			ExpectedErr:      nil,
		},
		{
			Name: "Success on global calls without tenant",
			DirectorClient: func() *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:     noTenantReqData,
			AuthDetailsInput: authDetails,
			ExpectedScopes:   strings.Join(scopes, " "),
			ExpectedTenant:   testEmptyTenant,
			ExpectedErr:      nil,
		},
		{
			Name: "Error when the error from getting the internal tenant is different from not found",
			DirectorClient: func() *automock.DirectorClient {
				tenantRepo := &automock.DirectorClient{}
				tenantRepo.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(nil, testError).Once()
				return tenantRepo
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:     reqData,
			AuthDetailsInput: authDetails,
			ExpectedTenant:   testTenantWithOnlyExternalID,
			ExpectedErr:      testError,
		},
		{
			Name: "Success when internal tenant exists",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testSubaccount, nil).Once()
				return client
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:     reqData,
			AuthDetailsInput: authDetails,
			ExpectedScopes:   strings.Join(scopes, " "),
			ExpectedTenant:   testSubaccount,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns empty region when tenant is subaccount and tenant region label is missing",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testSubaccount, nil).Once()
				return client
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:     reqData,
			AuthDetailsInput: authDetails,
			ExpectedScopes:   strings.Join(scopes, " "),
			ExpectedTenant:   testSubaccount,
			ExpectedErr:      nil,
		},
		{
			Name: "Error when consumer don't have access",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testAccount, nil).Once()
				return client
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:     reqData,
			AuthDetailsInput: authDetails,
			ExpectedErr:      apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s has no access to %s tenant with ID %s", authDetails.AuthID, testAccount.Type, consumerTenantID)),
		},
		{
			Name: "Error when consumer don't have global access",
			DirectorClient: func() *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Extra: map[string]interface{}{
						cert.ConsumerTypeExtraField: model.IntegrationSystemReference,
						cert.AccessLevelsExtraField: []interface{}{tenantEntity.Subaccount},
					},
				},
			},
			AuthDetailsInput: authDetails,
			ExpectedErr:      apperrors.NewUnauthorizedError(fmt.Sprintf("Certificate with auth ID %s does not have global access", authDetails.AuthID)),
		},
		{
			Name: "Error while getting scopes",
			DirectorClient: func() *automock.DirectorClient {
				return &automock.DirectorClient{}
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.integration_system").Return([]string{}, testError)
				return scopesGetter
			},
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Extra: map[string]interface{}{
						cert.ConsumerTypeExtraField: model.IntegrationSystemReference,
						cert.AccessLevelsExtraField: []interface{}{tenantEntity.Subaccount},
					},
				},
			},
			AuthDetailsInput: authDetails,
			ExpectedErr:      apperrors.NewInternalError("failed to extract scopes for consumer with type integration system"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			client := testCase.DirectorClient()
			scopesGetter := testCase.ScopesGetterFn()
			provider := tenantmapping.NewAccessLevelContextProvider(client, scopesGetter)
			// WHEN
			objectCtx, err := provider.GetObjectContext(emptyCtx, testCase.ReqDataInput, testCase.AuthDetailsInput)

			// THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, consumer.IntegrationSystem, objectCtx.ConsumerType)
				require.Equal(t, providerTenantID, objectCtx.ConsumerID)
				require.Equal(t, testCase.ExpectedTenant, objectCtx.Tenant)
				require.Equal(t, testCase.ExpectedScopes, objectCtx.Scopes)
				if objectCtx.Tenant.InternalID != "" {
					require.Equal(t, region, objectCtx.Region)
				}
			} else {
				require.Error(t, err)
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(testCase.ExpectedErr.Error()))
				require.Empty(t, objectCtx)
			}
			mock.AssertExpectationsForObjects(t, client, scopesGetter)
		})
	}
}

func TestAccessLevelContextProvider_Match(t *testing.T) {
	provider := tenantmapping.NewAccessLevelContextProvider(nil, nil)
	clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
	tenantHeader := "123"
	accessLevels := []interface{}{"account"}
	t.Run("returns ID string and CertificateFlow when a client-id-from-certificate is specified in the Header map of request body", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ExternalTenantKey):  []string{tenantHeader},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{clientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
				},
				Extra: map[string]interface{}{
					cert.AccessLevelsExtraField: accessLevels,
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.CertificateFlow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})

	t.Run("does not match when consumer type is not provided and does not match", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ExternalTenantKey):  []string{tenantHeader},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{clientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.False(t, match)
		require.Nil(t, authDetails)
		require.NoError(t, err)
	})
	t.Run("does not match when cert issuer is not external and does not match", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ExternalTenantKey): []string{tenantHeader},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):   []string{clientID},
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
	t.Run("matches when tenant header is missing", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{clientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
				},
				Extra: map[string]interface{}{
					cert.AccessLevelsExtraField: accessLevels,
				},
			},
		}

		match, authDetails, err := provider.Match(context.TODO(), reqData)

		require.True(t, match)
		require.NoError(t, err)
		require.Equal(t, oathkeeper.CertificateFlow, authDetails.AuthFlow)
		require.Equal(t, clientID, authDetails.AuthID)
	})
}
