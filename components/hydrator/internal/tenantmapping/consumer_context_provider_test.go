package tenantmapping_test

import (
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/hydrator/internal/config"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping/automock"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	tenantmapping_pkg "github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConsumerContextProvider_GetObjectContext(t *testing.T) {
	ctx := context.Background()
	certClientID := "d008c9db-2469-4d0b-af2f-880a6d0ba096"
	consumerTenantID := "1f538f34-30bf-4d3d-aeaa-02e69eef84ae"
	consumerInternalTenantID := "9a4d24e6-3ff6-464f-8efb-b167d1bdfcb6"
	tenantName := "test-tenant-name"
	clientID := "id-value!t12345"
	invalidClientID := "invalid-id%"
	authID := "user-name@test.com"
	testRegion := "eu-1"

	testError := errors.New("test error")
	notFoudErr := apperrors.NewNotFoundError(resource.Tenant, consumerTenantID)

	consumerClaimsKeysConfig := config.ConsumerClaimsKeysConfig{
		ClientIDKey: "client_id",
		TenantIDKey: "tenantid",
		UserNameKey: "user_name",
	}

	authDetails := oathkeeper.AuthDetails{AuthID: authID, AuthFlow: oathkeeper.ConsumerProviderFlow}

	userCtxHeaderWithAllProperties := fmt.Sprintf(`{"client_id":"%s","tenantid":"%s","user_name":"%s"}`, clientID, consumerTenantID, authID)
	userCtxHeaderWithoutClientID := fmt.Sprintf(`{"tenantid":"%s","user_name":"%s"}`, consumerTenantID, authID)
	userCtxHeaderWithoutTenantID := fmt.Sprintf(`{"client_id":"%s","user_name":"%s"}`, clientID, authID)
	userCtxHeaderWithInvalidASCIICharacter := fmt.Sprintf(`{"client_id":"%s","tenantid":"%s","user_name":"%s"}`, invalidClientID, consumerTenantID, authID)
	userCtxHeaderWithInvalidASCIICharacterAndMissingClientID := `{"tenantid":"1f538f34-30bf-4d3d-aeaa-02e69eef84ae","user_name":"test%UserName@test.com"}`
	userCtxHeaderWithInvalidASCIICharacterAndMissingTenantID := fmt.Sprintf(`{"client_id":"%s","user_name":"%s"}`, invalidClientID, authID)

	reqDataFunc := func(userContextHeader string) oathkeeper.ReqData {
		return oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{certClientID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
				},
			},
			Header: http.Header{
				oathkeeper.UserContextKey: []string{userContextHeader},
			},
		}
	}

	expectedObjectContextFunc := func(externalTenantID, internalTenantID, region, oauthClientID string) tenantmapping.ObjectContext {
		return tenantmapping.ObjectContext{
			TenantContext: tenantmapping.NewTenantContext(externalTenantID, internalTenantID),
			KeysExtra: tenantmapping.KeysExtra{
				TenantKey:         tenantmapping_pkg.ConsumerTenantKey,
				ExternalTenantKey: tenantmapping_pkg.ExternalTenantKey,
			},
			Scopes:              "",
			ScopesMergeStrategy: mergeWithOtherScopes,
			Region:              region,
			OauthClientID:       oauthClientID,
			ConsumerID:          authID,
			AuthFlow:            oathkeeper.ConsumerProviderFlow,
			ConsumerType:        consumer.User,
			ContextProvider:     tenantmapping_pkg.ConsumerProviderObjectContextProvider,
		}
	}

	testTenant := &graphql.Tenant{
		ID:         consumerTenantID,
		InternalID: consumerInternalTenantID,
		Name:       &tenantName,
		Type:       "subaccount",
		Labels: map[string]interface{}{
			"region":    testRegion,
			"subdomain": "consumer-subdomain",
		},
		Provider: "provider-tenant",
	}

	testTenantWithIncorrectRegionLabelType := &graphql.Tenant{
		ID:         consumerTenantID,
		InternalID: consumerInternalTenantID,
		Name:       &tenantName,
		Type:       "subaccount",
		Labels: map[string]interface{}{
			"region":    []string{testRegion},
			"subdomain": "consumer-subdomain",
		},
		Provider: "provider-tenant",
	}

	testCases := []struct {
		Name                  string
		DirectorClient        func() *automock.DirectorClient
		ReqDataInput          oathkeeper.ReqData
		ExpectedObjectContext tenantmapping.ObjectContext
		ExpectedErrMsg        string
	}{
		{
			Name: "Success",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testTenant, nil).Once()
				return client
			},
			ReqDataInput:          reqDataFunc(userCtxHeaderWithAllProperties),
			ExpectedObjectContext: expectedObjectContextFunc(consumerTenantID, consumerInternalTenantID, testRegion, clientID),
		},
		{
			Name: "Success when unescape fails and the original header value is used successfully",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testTenant, nil).Once()
				return client
			},
			ReqDataInput:          reqDataFunc(userCtxHeaderWithInvalidASCIICharacter),
			ExpectedObjectContext: expectedObjectContextFunc(consumerTenantID, consumerInternalTenantID, testRegion, invalidClientID),
		},
		{
			Name:                  "Returns error when unescape fails and the client_id property from the original header value is empty",
			ReqDataInput:          reqDataFunc(userCtxHeaderWithInvalidASCIICharacterAndMissingClientID),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        fmt.Sprintf("while getting user context data from %q header: invalid data [reason=property %q is mandatory", oathkeeper.UserContextKey, consumerClaimsKeysConfig.ClientIDKey),
		},
		{
			Name:                  "Returns error when unescape fails and the tenantid property from the original header value is empty",
			ReqDataInput:          reqDataFunc(userCtxHeaderWithInvalidASCIICharacterAndMissingTenantID),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        fmt.Sprintf("while getting user context data from %q header: invalid data [reason=property %q is mandatory", oathkeeper.UserContextKey, consumerClaimsKeysConfig.TenantIDKey),
		},
		{
			Name:                  "Returns error when client_id property is missing from user_context header",
			ReqDataInput:          reqDataFunc(userCtxHeaderWithoutClientID),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        fmt.Sprintf("while getting user context data from %q header: invalid data [reason=property %q is mandatory", oathkeeper.UserContextKey, consumerClaimsKeysConfig.ClientIDKey),
		},
		{
			Name:                  "Returns error when tenantid property is missing from user_context header",
			ReqDataInput:          reqDataFunc(userCtxHeaderWithoutTenantID),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        fmt.Sprintf("while getting user context data from %q header: invalid data [reason=property %q is mandatory", oathkeeper.UserContextKey, consumerClaimsKeysConfig.TenantIDKey),
		},
		{
			Name: "Returns error while getting tenant by external ID",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(nil, testError).Once()
				return client
			},
			ReqDataInput:          reqDataFunc(userCtxHeaderWithAllProperties),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        fmt.Sprintf("while getting external tenant mapping [ExternalTenantID=%s]", consumerTenantID),
		},
		{
			Name: "Returns object context with empty internal ID when getting tenant returns not found error",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(nil, notFoudErr).Once()
				return client
			},
			ReqDataInput:          reqDataFunc(userCtxHeaderWithAllProperties),
			ExpectedObjectContext: expectedObjectContextFunc(consumerTenantID, "", "", clientID),
		},
		{
			Name: "Returns empty region when tenant is subaccount and tenant region label is missing",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testTenant, nil).Once()
				return client
			},
			ReqDataInput:          reqDataFunc(userCtxHeaderWithAllProperties),
			ExpectedObjectContext: expectedObjectContextFunc(consumerTenantID, consumerInternalTenantID, "", clientID),
		},
		{
			Name: "Returns error when region label type is not the expected one",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testTenantWithIncorrectRegionLabelType, nil).Once()
				return client
			},
			ReqDataInput:          reqDataFunc(userCtxHeaderWithAllProperties),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        fmt.Sprintf("unexpected region label type: %T, should be string", []string{}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			var client *automock.DirectorClient
			if testCase.DirectorClient != nil {
				client = testCase.DirectorClient()
			} else {
				client = &automock.DirectorClient{}
			}

			provider := tenantmapping.NewConsumerContextProvider(client, consumerClaimsKeysConfig)
			// WHEN
			objectCtx, err := provider.GetObjectContext(ctx, testCase.ReqDataInput, authDetails)

			// THEN
			if testCase.ExpectedErrMsg == "" {
				require.NoError(t, err)
				require.Equal(t, consumer.User, objectCtx.ConsumerType)
				require.Equal(t, authID, objectCtx.ConsumerID)
				require.Equal(t, testCase.ExpectedObjectContext.OauthClientID, objectCtx.OauthClientID)
				require.Equal(t, oathkeeper.ConsumerProviderFlow, objectCtx.AuthFlow)
				require.Equal(t, testCase.ExpectedObjectContext.TenantContext.TenantID, objectCtx.TenantContext.TenantID)
				require.Equal(t, consumerTenantID, objectCtx.TenantContext.ExternalTenantID)
				require.Equal(t, "", objectCtx.Scopes)
			} else {
				require.Error(t, err)
				require.Contains(t, strings.ToLower(err.Error()), strings.ToLower(testCase.ExpectedErrMsg))
				require.Empty(t, objectCtx)
			}
			mock.AssertExpectationsForObjects(t, client)
		})
	}
}

func TestConsumerContextProvider_Match(t *testing.T) {
	ctx := context.Background()
	certClientID := "d008c9db-2469-4d0b-af2f-880a6d0ba096"

	consumerClaimsKeysConfig := config.ConsumerClaimsKeysConfig{
		ClientIDKey: "client_id",
		TenantIDKey: "tenantid",
		UserNameKey: "user_name",
	}
	provider := tenantmapping.NewConsumerContextProvider(nil, consumerClaimsKeysConfig)

	userCtxHeader := `{"client_id":"id-value!t12345","tenantid":"f8075207-1478-4a80-bd26-24a4785a2bfd","user_name":"user-name@test.com"}`
	userCtxHeaderWithoutUserNameProperty := `{"client_id":"id-value!t12345","tenantid":"f8075207-1478-4a80-bd26-24a4785a2bfd"}`
	userCtxHeaderWithNonASCIICharacters := `{"client_id":"test nøn asçii chå®acte®","tenantid":"f8075207-1478-4a80-bd26-24a4785a2bfd","user_name":"user-name@test.com"}`
	userCtxHeaderWithEncodedNonASCIICharacters := `{"client_id":"test+n%C3%B8n+as%C3%A7ii+ch%C3%A5%C2%AEacte%C2%AE","tenantid":"f8075207-1478-4a80-bd26-24a4785a2bfd","user_name":"user-name@test.com"}`
	userCtxHeaderWithInvalidASCIICharacter := `{"client_id":"invalid-id%","tenantid":"f8075207-1478-4a80-bd26-24a4785a2bfd","user_name":"user-name@test.com"}`
	userCtxHeaderWithInvalidASCIICharacterAndMissingUserNameProperty := `{"client_id":"invalid-id%","tenantid":"f8075207-1478-4a80-bd26-24a4785a2bfd"}`

	testCases := []struct {
		Name                string
		ReqDataInput        oathkeeper.ReqData
		ExpectedMatch       bool
		ExpectedAuthDetails *oathkeeper.AuthDetails
		ExpectedErrMsg      string
	}{
		{
			Name: "Success",
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{certClientID},
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
					},
				},
				Header: http.Header{
					oathkeeper.UserContextKey: []string{userCtxHeader},
				},
			},
			ExpectedMatch: true,
			ExpectedAuthDetails: &oathkeeper.AuthDetails{
				AuthID:   "user-name@test.com",
				AuthFlow: oathkeeper.ConsumerProviderFlow,
			},
			ExpectedErrMsg: "",
		},
		{
			Name: "Success when user_context header contains non ascii characters",
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{certClientID},
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
					},
				},
				Header: http.Header{
					oathkeeper.UserContextKey: []string{userCtxHeaderWithNonASCIICharacters},
				},
			},
			ExpectedMatch: true,
			ExpectedAuthDetails: &oathkeeper.AuthDetails{
				AuthID:   "user-name@test.com",
				AuthFlow: oathkeeper.ConsumerProviderFlow,
			},
			ExpectedErrMsg: "",
		},
		{
			Name: "Success when user_context header contains encoded non ascii characters",
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{certClientID},
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
					},
				},
				Header: http.Header{
					oathkeeper.UserContextKey: []string{userCtxHeaderWithEncodedNonASCIICharacters},
				},
			},
			ExpectedMatch: true,
			ExpectedAuthDetails: &oathkeeper.AuthDetails{
				AuthID:   "user-name@test.com",
				AuthFlow: oathkeeper.ConsumerProviderFlow,
			},
			ExpectedErrMsg: "",
		},
		{
			Name: "Returns error when user_context header is missing",
			ReqDataInput: oathkeeper.ReqData{
				Header: http.Header{},
			},
			ExpectedMatch:       false,
			ExpectedAuthDetails: nil,
			ExpectedErrMsg:      fmt.Sprintf("the key does not exist in the source object [key=%s]", oathkeeper.UserContextKey),
		},
		{
			Name: "Do not match when cert ID is empty",
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{""},
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
					},
				},
				Header: http.Header{
					oathkeeper.UserContextKey: []string{userCtxHeader},
				},
			},
			ExpectedMatch:       false,
			ExpectedAuthDetails: nil,
			ExpectedErrMsg:      "",
		},
		{
			Name: "Do not match when certificate issuer is not the correct one",
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{certClientID},
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ConnectorIssuer},
					},
				},
				Header: http.Header{
					oathkeeper.UserContextKey: []string{userCtxHeader},
				},
			},
			ExpectedMatch:       false,
			ExpectedAuthDetails: nil,
			ExpectedErrMsg:      "",
		},
		{
			Name: "Returns error when user_name property from the user_context header is empty",
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{certClientID},
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
					},
				},
				Header: http.Header{
					oathkeeper.UserContextKey: []string{userCtxHeaderWithoutUserNameProperty},
				},
			},
			ExpectedMatch:       false,
			ExpectedAuthDetails: nil,
			ExpectedErrMsg:      fmt.Sprintf("could not find %s property", consumerClaimsKeysConfig.UserNameKey),
		},
		{
			Name: "Success when unescape fails and the original header value is used successfully",
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{certClientID},
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
					},
				},
				Header: http.Header{
					oathkeeper.UserContextKey: []string{userCtxHeaderWithInvalidASCIICharacter},
				},
			},
			ExpectedMatch: true,
			ExpectedAuthDetails: &oathkeeper.AuthDetails{
				AuthID:   "user-name@test.com",
				AuthFlow: oathkeeper.ConsumerProviderFlow,
			},
			ExpectedErrMsg: "",
		},
		{
			Name: "Error when unescape fails and the user_name property from the original header value is empty",
			ReqDataInput: oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Header: http.Header{
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{certClientID},
						textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
					},
				},
				Header: http.Header{
					oathkeeper.UserContextKey: []string{userCtxHeaderWithInvalidASCIICharacterAndMissingUserNameProperty},
				},
			},
			ExpectedMatch:       false,
			ExpectedAuthDetails: nil,
			ExpectedErrMsg:      fmt.Sprintf("could not find %s property", consumerClaimsKeysConfig.UserNameKey),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			match, authDetails, err := provider.Match(ctx, testCase.ReqDataInput)
			require.Equal(t, testCase.ExpectedMatch, match)
			require.EqualValues(t, testCase.ExpectedAuthDetails, authDetails)

			if testCase.ExpectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
