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
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping/automock"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCertServiceContextProvider(t *testing.T) {
	emptyCtx := context.TODO()
	tenantID := uuid.New().String()

	testError := errors.New("test error")
	notFoundErr := apperrors.NewNotFoundErrorWithType(resource.Tenant)
	subaccountRegionNotFoundErr := fmt.Errorf("region label not found for subaccount with ID: %q", tenantID)

	authDetails := oathkeeper.AuthDetails{AuthID: tenantID, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ExternalIssuer}

	scopes := []string{"runtime:read", "runtime:write", "tenant:read"}
	scopesString := "runtime:read runtime:write tenant:read"

	reqData := oathkeeper.ReqData{}

	internalConsumerID := "123"
	reqDataWithInternalConsumerID := oathkeeper.ReqData{
		Body: oathkeeper.ReqBody{Extra: map[string]interface{}{
			cert.InternalConsumerIDField: internalConsumerID,
		}},
	}

	internalSubaccount := "internalSubaccountID"

	testSubaccount := &graphql.Tenant{
		ID:         "externalTestSubaccount",
		InternalID: internalSubaccount,
		Name:       str.Ptr("testSubaccount"),
		Type:       "subaccount",
		Labels: map[string]interface{}{
			"region": "eu-1",
		},
	}

	testSubaccountWithoutRegion := *testSubaccount
	testSubaccountWithoutRegion.Labels = nil

	testCases := []struct {
		Name               string
		DirectorClient     func() *automock.DirectorClient
		ScopesGetterFn     func() *automock.ScopesGetter
		ReqDataInput       oathkeeper.ReqData
		AuthDetailsInput   oathkeeper.AuthDetails
		ExpectedScopes     string
		ExpectedInternalID string
		ExpectedConsumerID string
		ExpectedErr        error
	}{
		{
			Name: "Success when cannot find internal tenant",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, tenantID).Return(nil, notFoundErr).Once()
				return client
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.external_certificate").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:       reqData,
			AuthDetailsInput:   authDetails,
			ExpectedScopes:     scopesString,
			ExpectedInternalID: "",
			ExpectedErr:        nil,
		},
		{
			Name: "Error when the error from getting the internal tenant is different from not found",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, tenantID).Return(nil, testError).Once()
				return client
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.external_certificate").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:       reqData,
			AuthDetailsInput:   authDetails,
			ExpectedScopes:     "",
			ExpectedInternalID: "",
			ExpectedErr:        testError,
		},
		{
			Name: "Success when internal tenant exists",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, tenantID).Return(testSubaccount, nil).Once()
				return client
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.external_certificate").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:       reqData,
			AuthDetailsInput:   authDetails,
			ExpectedScopes:     scopesString,
			ExpectedInternalID: internalSubaccount,
			ExpectedErr:        nil,
		},
		{
			Name: "Success when internal consumer ID is provided",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, tenantID).Return(testSubaccount, nil).Once()
				return client
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.external_certificate").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:       reqDataWithInternalConsumerID,
			AuthDetailsInput:   authDetails,
			ExpectedScopes:     scopesString,
			ExpectedInternalID: internalSubaccount,
			ExpectedConsumerID: internalConsumerID,
			ExpectedErr:        nil,
		},
		{
			Name: "Error when can't extract tenant region",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, tenantID).Return(&testSubaccountWithoutRegion, nil).Once()
				return client
			},
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.external_certificate").Return(scopes, nil)
				return scopesGetter
			},
			ReqDataInput:     reqDataWithInternalConsumerID,
			AuthDetailsInput: authDetails,
			ExpectedScopes:   scopesString,
			ExpectedErr:      subaccountRegionNotFoundErr,
		},
		{
			Name:           "Error when can't get required scopes",
			DirectorClient: unusedDirectorClient,
			ScopesGetterFn: func() *automock.ScopesGetter {
				scopesGetter := &automock.ScopesGetter{}
				scopesGetter.On("GetRequiredScopes", "scopesPerConsumerType.external_certificate").Return(nil, testError)
				return scopesGetter
			},
			ReqDataInput:     reqData,
			AuthDetailsInput: authDetails,
			ExpectedErr:      errors.New("failed to extract scopes"),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			client := testCase.DirectorClient()
			scopesGetter := testCase.ScopesGetterFn()
			provider := tenantmapping.NewCertServiceContextProvider(client, scopesGetter)
			if testCase.ExpectedConsumerID == "" {
				testCase.ExpectedConsumerID = tenantID
			}
			// WHEN
			objectCtx, err := provider.GetObjectContext(emptyCtx, testCase.ReqDataInput, testCase.AuthDetailsInput)

			// THEN
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, consumer.ExternalCertificate, objectCtx.ConsumerType)
				require.Equal(t, testCase.ExpectedConsumerID, objectCtx.ConsumerID)
				require.Equal(t, testCase.ExpectedInternalID, objectCtx.TenantContext.TenantID)
				require.Equal(t, tenantID, objectCtx.TenantContext.ExternalTenantID)
				require.Equal(t, testCase.ExpectedScopes, objectCtx.Scopes)
			} else {
				require.Error(t, err)
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(testCase.ExpectedErr.Error()))
				require.Empty(t, objectCtx)
			}
			mock.AssertExpectationsForObjects(t, scopesGetter, client)
		})
	}
}

func TestCertServiceContextProviderMatch(t *testing.T) {
	t.Run("returns ID string and CertificateFlow when a client-id-from-certificate is specified in the Header map of request body", func(t *testing.T) {
		clientID := "de766a55-3abb-4480-8d4a-6d255990b159"
		provider := tenantmapping.NewCertServiceContextProvider(nil, nil)

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
		provider := tenantmapping.NewCertServiceContextProvider(nil, nil)

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

func unusedDirectorClient() *automock.DirectorClient {
	return &automock.DirectorClient{}
}
