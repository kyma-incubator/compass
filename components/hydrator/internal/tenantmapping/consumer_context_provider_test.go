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
	authID := "test-user-name@sap.com"
	testRegion := "eu-1"
	testSubdomain := "consumerSubdomain"

	testError := errors.New("test error")
	notFoudErr := apperrors.NewNotFoundError(resource.Tenant, consumerTenantID)

	consumerClaimsKeysConfig := config.ConsumerClaimsKeysConfig{
		ClientIDKey:  "client_id",
		TenantIDKey:  "tenantid",
		UserNameKey:  "user_name",
		SubdomainKey: "subdomain",
	}

	authDetails := oathkeeper.AuthDetails{AuthID: authID, AuthFlow: oathkeeper.ConsumerProviderFlow}

	userCtxHeaderWithAllProperties := fmt.Sprintf(`{"client_id":"%s","exp":1659618593,"tenantid":"%s","subdomain":"%s","identity":"subscription-flow-identity","iss":"http://compass-external-services-mock.compass-system.svc.cluster.local:8080","subsc-key-test":"subscription-flow","tenant":"%s","user_name":"%s","x-zid":""}`, clientID, consumerTenantID, testSubdomain, consumerTenantID, authID)
	userCtxHeaderWithoutClientID := fmt.Sprintf(`{"exp":1659618593,"tenantid":"%s","subdomain":"%s","identity":"subscription-flow-identity","iss":"http://compass-external-services-mock.compass-system.svc.cluster.local:8080","subsc-key-test":"subscription-flow","tenant":"%s","user_name":"%s","x-zid":""}`, consumerTenantID, testSubdomain, consumerTenantID, authID)
	userCtxHeaderWithoutTenantID := fmt.Sprintf(`{"client_id":"%s","exp":1659618593,"subdomain":"%s","identity":"subscription-flow-identity","iss":"http://compass-external-services-mock.compass-system.svc.cluster.local:8080","subsc-key-test":"subscription-flow","tenant":"%s","user_name":"%s","x-zid":""}`, clientID, testSubdomain, consumerTenantID, authID)
	userCtxHeaderWithoutSubdomain := fmt.Sprintf(`{"client_id":"%s","exp":1659618593,"tenantid":"%s","identity":"subscription-flow-identity","iss":"http://compass-external-services-mock.compass-system.svc.cluster.local:8080","subsc-key-test":"subscription-flow","tenant":"%s","user_name":"%s","x-zid":""}`, clientID, consumerTenantID, consumerTenantID, authID)

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

	expectedObjectContextFunc := func(externalTenantID, internalTenantID, region string) tenantmapping.ObjectContext {
		return tenantmapping.ObjectContext{
			TenantContext: tenantmapping.NewTenantContext(externalTenantID, internalTenantID),
			KeysExtra: tenantmapping.KeysExtra{
				TenantKey:         tenantmapping_pkg.ConsumerTenantKey,
				ExternalTenantKey: tenantmapping_pkg.ExternalTenantKey,
			},
			Scopes:              "",
			ScopesMergeStrategy: mergeWithOtherScopes,
			Region:              region,
			OauthClientID:       clientID,
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

	testTenantWithoutRegion := &graphql.Tenant{
		ID:         consumerTenantID,
		InternalID: consumerInternalTenantID,
		Name:       &tenantName,
		Type:       "subaccount",
		Labels: map[string]interface{}{
			"subdomain": "consumer-subdomain",
		},
		Provider: "provider-tenant",
	}

	testTenantWithoutSubdomain := &graphql.Tenant{
		ID:         consumerTenantID,
		InternalID: consumerInternalTenantID,
		Name:       &tenantName,
		Type:       "subaccount",
		Labels: map[string]interface{}{
			"region": testRegion,
		},
		Provider: "provider-tenant",
	}

	region := testTenantWithoutSubdomain.Labels["region"].(string)
	tenantToUpdateWithSubdomainFromHeader := graphql.BusinessTenantMappingInput{
		Name:           *testTenantWithoutSubdomain.Name,
		ExternalTenant: testTenantWithoutSubdomain.ID,
		Parent:         &testTenantWithoutSubdomain.ParentID,
		Subdomain:      &testSubdomain,
		Region:         &region,
		Type:           testTenantWithoutSubdomain.Type,
		Provider:       testTenantWithoutSubdomain.Provider,
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
			ExpectedObjectContext: expectedObjectContextFunc(clientID, consumerInternalTenantID, testRegion),
		},
		{
			Name: "Success when tenant subdomain label is missing and tenant update is successful",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testTenantWithoutSubdomain, nil).Once()
				client.On("WriteTenants", mock.Anything, []graphql.BusinessTenantMappingInput{tenantToUpdateWithSubdomainFromHeader}).Return(nil).Once()
				return client
			},
			ReqDataInput:          reqDataFunc(userCtxHeaderWithAllProperties),
			ExpectedObjectContext: expectedObjectContextFunc(clientID, consumerInternalTenantID, testRegion),
		},
		{
			Name:                  "Returns error when client_id property is missing from user_context header",
			ReqDataInput:          reqDataFunc(userCtxHeaderWithoutClientID),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        "while getting user context data from \"user_context\" header: invalid data [reason=property \"client_id\" is mandatory",
		},
		{
			Name:                  "Returns error when tenantid property is missing from user_context header",
			ReqDataInput:          reqDataFunc(userCtxHeaderWithoutTenantID),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        "while getting user context data from \"user_context\" header: invalid data [reason=property \"tenantid\" is mandatory",
		},
		{
			Name:                  "Returns error when subdomain property is missing from user_context header",
			ReqDataInput:          reqDataFunc(userCtxHeaderWithoutSubdomain),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        "while getting user context data from \"user_context\" header: invalid data [reason=property \"subdomain\" is mandatory",
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
			ExpectedObjectContext: expectedObjectContextFunc(consumerTenantID, "", ""),
		},
		{
			Name: "Returns error when region label is missing from the fetched tenant",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testTenantWithoutRegion, nil).Once()
				return client
			},
			ReqDataInput:          reqDataFunc(userCtxHeaderWithAllProperties),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        fmt.Sprintf("region label not found for tenant with ID: %q", consumerTenantID),
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
		{
			Name: "Returns error while updating tenant when subdomain label is missing",
			DirectorClient: func() *automock.DirectorClient {
				client := &automock.DirectorClient{}
				client.On("GetTenantByExternalID", mock.Anything, consumerTenantID).Return(testTenantWithoutSubdomain, nil).Once()
				client.On("WriteTenants", mock.Anything, []graphql.BusinessTenantMappingInput{tenantToUpdateWithSubdomainFromHeader}).Return(testError).Once()
				return client
			},
			ReqDataInput:          reqDataFunc(userCtxHeaderWithAllProperties),
			ExpectedObjectContext: tenantmapping.ObjectContext{},
			ExpectedErrMsg:        fmt.Sprintf("an error occurred while write tenant with external ID: %q:", consumerTenantID),
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
				require.Equal(t, clientID, objectCtx.OauthClientID)
				require.Equal(t, oathkeeper.ConsumerProviderFlow, objectCtx.AuthFlow)
				//require.Equal(t, consumerInternalTenantID, objectCtx.TenantContext.TenantID)
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
		ClientIDKey:  "client_id",
		TenantIDKey:  "tenantid",
		UserNameKey:  "user_name",
		SubdomainKey: "subdomain",
	}
	provider := tenantmapping.NewConsumerContextProvider(nil, consumerClaimsKeysConfig)

	userCtxHeader := `{"client_id":"id-value!t12345","exp":1659618593,"tenantid":"f8075207-1478-4a80-bd26-24a4785a2bfd","subdomain":"consumerSubdomain","identity":"subscription-flow-identity","iss":"http://compass-external-services-mock.compass-system.svc.cluster.local:8080","subsc-key-test":"subscription-flow","tenant":"1f538f34-30bf-4d3d-aeaa-02e69eef84ae","user_name":"test-user-name@sap.com","x-zid":""}`
	userCtxHeaderWithoutUserNameProperty := `{"client_id":"id-value!t12345","exp":1659618593,"tenantid":"f8075207-1478-4a80-bd26-24a4785a2bfd","subdomain":"consumerSubdomain","identity":"subscription-flow-identity","iss":"http://compass-external-services-mock.compass-system.svc.cluster.local:8080","subsc-key-test":"subscription-flow","tenant":"1f538f34-30bf-4d3d-aeaa-02e69eef84ae","x-zid":""}`

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
				AuthID:   "test-user-name@sap.com",
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
			ExpectedErrMsg:      "the key does not exist in the source object [key=User_context]",
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
			ExpectedErrMsg:      "could not find user_name property",
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
