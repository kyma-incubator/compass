package handler_test

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/handler"
	"github.com/kyma-incubator/compass/components/kyma-adapter/internal/handler/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_HandlerFunc(t *testing.T) {
	tenantID := "tenant-id"
	url := "https://target-url.com"
	apiPath := fmt.Sprintf("/v1/tenantMappings/%s", tenantID)

	testErr := errors.New("test err")

	platform := "unit-tests"
	receiverTenantID := "receiver-tenant-id"
	receiverOwnerTenantID := "receiver-owner-tenant-id"
	assignedTenantID := "assigned-tenant-id"
	assignOperation := "assign"
	unassignOperation := "unassign"
	username := "user"
	password := "pass"
	basicCredentials := fmt.Sprintf(`{"credentials":{"outboundCommunication":{"basicAuthentication":{"password":%q,"username":%q}}}}`, password, username)
	clientID := "id"
	clientSecret := "secret"
	tokenServiceURL := "token-url"
	oauthCredentials := fmt.Sprintf(`{"credentials":{"outboundCommunication":{"oauth2ClientCredentials":{"clientId":%q,"clientSecret":%q,"tokenServiceUrl":%q}}}}`, clientID, clientSecret, tokenServiceURL)

	bodyFormatterBasic := `{"context":{"platform":%q,"operation":%q},"receiverTenant":{"ownerTenant":%q,"uclSystemTenantId":%q},"assignedTenant":{"uclSystemTenantId":%q,"configuration":%s}}`

	bodyWithConfigPendingState := "{\"state\":\"CONFIG_PENDING\"}\n"
	bodyWithReadyState := "{\"state\":\"READY\"}\n"

	bundles := []*graphql.BundleExt{{Bundle: graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: "bndl-1"}}}, {Bundle: graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: "bndl-1"}}}}
	bundlesWithAuths := []*graphql.BundleExt{
		{Bundle: graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: "bndl-1"}}, InstanceAuths: []*graphql.BundleInstanceAuth{{ID: "auth-1", RuntimeID: &receiverTenantID}}},
		{Bundle: graphql.Bundle{BaseEntity: &graphql.BaseEntity{ID: "bndl-1"}}, InstanceAuths: []*graphql.BundleInstanceAuth{{ID: "auth-1", RuntimeID: &receiverTenantID}}},
	}

	testCases := []struct {
		name                 string
		clientFn             func() *automock.Client
		requestBody          string
		expectedBody         string
		expectedResponseCode int
	}{
		{
			name:                 "Success - assign with missing config",
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, `{}`),
			expectedBody:         bodyWithConfigPendingState,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "Success - assign for application with no bundles",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return([]*graphql.BundleExt{}, nil).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedBody:         bodyWithReadyState,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "Success - assign with provided basic config",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundles, nil).Once()
				client.On("CreateBasicBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundles[0].Bundle.ID, receiverTenantID, username, password).Return(nil).Once()
				client.On("CreateBasicBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundles[1].Bundle.ID, receiverTenantID, username, password).Return(nil).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedBody:         bodyWithReadyState,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "Success - assign with provided oauth config",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundles, nil).Once()
				client.On("CreateOauthBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundles[0].Bundle.ID, receiverTenantID, tokenServiceURL, clientID, clientSecret).Return(nil).Once()
				client.On("CreateOauthBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundles[1].Bundle.ID, receiverTenantID, tokenServiceURL, clientID, clientSecret).Return(nil).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, oauthCredentials),
			expectedBody:         bodyWithReadyState,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "Success - resync with provided basic config",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundlesWithAuths, nil).Once()
				client.On("UpdateBasicBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundlesWithAuths[0].InstanceAuths[0].ID, bundlesWithAuths[0].Bundle.ID, username, password).Return(nil).Once()
				client.On("UpdateBasicBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundlesWithAuths[1].InstanceAuths[0].ID, bundlesWithAuths[1].Bundle.ID, username, password).Return(nil).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedBody:         bodyWithReadyState,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "Success - resync with provided oauth config",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundlesWithAuths, nil).Once()
				client.On("UpdateOauthBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundlesWithAuths[0].InstanceAuths[0].ID, bundlesWithAuths[0].Bundle.ID, tokenServiceURL, clientID, clientSecret).Return(nil).Once()
				client.On("UpdateOauthBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundlesWithAuths[1].InstanceAuths[0].ID, bundlesWithAuths[1].Bundle.ID, tokenServiceURL, clientID, clientSecret).Return(nil).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, oauthCredentials),
			expectedBody:         bodyWithReadyState,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "Success - unassign for application with no bundles",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return([]*graphql.BundleExt{}, nil).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, unassignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedBody:         bodyWithReadyState,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "Success - unassign auths",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundlesWithAuths, nil).Once()
				client.On("DeleteBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundlesWithAuths[0].InstanceAuths[0].ID).Return(nil).Once()
				client.On("DeleteBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundlesWithAuths[1].InstanceAuths[0].ID).Return(nil).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, unassignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedBody:         bodyWithReadyState,
			expectedResponseCode: http.StatusOK,
		},
		{
			name: "Success - unassign with no auths",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundles, nil).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, unassignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedBody:         bodyWithReadyState,
			expectedResponseCode: http.StatusOK,
		},
		{
			name:                 "Error - invalid json request body",
			requestBody:          `{`,
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name:                 "Error - body can't be validated",
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, "", receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "Error - assign can't get application bundles",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(nil, testErr).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "Error - assign can't create auth with basic config",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundles, nil).Once()
				client.On("CreateBasicBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundles[0].Bundle.ID, receiverTenantID, username, password).Return(testErr).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "Error - assign can't create auth with oauth config",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundles, nil).Once()
				client.On("CreateOauthBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundles[0].Bundle.ID, receiverTenantID, tokenServiceURL, clientID, clientSecret).Return(testErr).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, oauthCredentials),
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "Error - resync can't update auth with basic config",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundlesWithAuths, nil).Once()
				client.On("UpdateBasicBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundlesWithAuths[0].InstanceAuths[0].ID, bundlesWithAuths[0].Bundle.ID, username, password).Return(testErr).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "Error - resync can't update oauth config",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundlesWithAuths, nil).Once()
				client.On("UpdateOauthBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundlesWithAuths[0].InstanceAuths[0].ID, bundlesWithAuths[0].Bundle.ID, tokenServiceURL, clientID, clientSecret).Return(testErr).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, assignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, oauthCredentials),
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "Error - unassign can't get application bundles",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(nil, testErr).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, unassignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedResponseCode: http.StatusBadRequest,
		},
		{
			name: "Error - unassign can't delete auth",
			clientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("GetApplicationBundles", mock.Anything, assignedTenantID, receiverOwnerTenantID).Return(bundlesWithAuths, nil).Once()
				client.On("DeleteBundleInstanceAuth", mock.Anything, receiverOwnerTenantID, bundlesWithAuths[0].InstanceAuths[0].ID).Return(testErr).Once()
				return client
			},
			requestBody:          fmt.Sprintf(bodyFormatterBasic, platform, unassignOperation, receiverOwnerTenantID, receiverTenantID, assignedTenantID, basicCredentials),
			expectedResponseCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//GIVEN
			client := &automock.Client{}
			if testCase.clientFn != nil {
				client = testCase.clientFn()
			}
			defer client.AssertExpectations(t)

			req, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(testCase.requestBody)))
			require.NoError(t, err)

			h := handler.NewHandler(client)
			recorder := httptest.NewRecorder()

			//WHEN
			h.HandlerFunc(recorder, req)
			resp := recorder.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			//THEN
			if testCase.expectedResponseCode == http.StatusOK {
				require.Equal(t, testCase.expectedBody, string(body), string(body))
			}
			require.Equal(t, testCase.expectedResponseCode, resp.StatusCode, string(body))
		})
	}
}
