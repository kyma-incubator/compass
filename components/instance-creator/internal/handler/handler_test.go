package handler_test

import (
	"bytes"
	"fmt"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/handler"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/handler/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func Test_HandlerFunc(t *testing.T) {
	url := "https://target-url.com"
	apiPath := fmt.Sprintf("/")
	statusUrl := "localhost"

	formationID := "formation-id"
	assignmentID := "assignment-id"
	region := "region"
	subaccount := "subaccount"

	reqBodyFormatter := `{
	 "context": %s,
	 "receiverTenant": %s,
	 "assignedTenant": %s
	}`

	reqBodyContextFormatter := `{"uclFormationId": %q, "operation": %q}`
	reqBodyContextWithAssign := fmt.Sprintf(reqBodyContextFormatter, formationID, "assign")
	reqBodyContextWithUnassign := fmt.Sprintf(reqBodyContextFormatter, formationID, "unassign")

	assignedTenantFormatter := `{
		"uclAssignmentId": %q,
		"configuration": %s
	}`

	receiverTenantFormatter := `{
		"deploymentRegion": %q,
		"subaccountId": %q,
		"configuration": %s
	}`

	testCases := []struct {
		name                 string
		smClientFn           func() *automock.Client
		mtlsClientFn         func() *automock.MtlsHTTPClient
		requestBody          string
		expectedResponseCode int
	}{
		{
			name:        "Wrong json - fails on decoding",
			requestBody: `wrong json`,
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Request body contains badly-formed JSON")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing config(empty json) - fails on validation",
			requestBody: `{}`,
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Formation ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing config(empty context, receiverTenant and assignedTenant) - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, "{}", "{}", "{}"),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("while validating the request body")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing formation ID in the context - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, `{"operation": "assign"}`, "{}", "{}"),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Formation ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Missing operation in the context - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, `{"uclFormationId": "formation-id", "operation": ""}`, "{}", "{}"),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Operation is invalid")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Wrong operation in the context - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, `{"uclFormationId": "formation-id", "operation": "wrong-operation"}`, "{}", "{}"),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Context's Operation is invalid")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Formation assignment is missing in the assignedTenant - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, "{}", "{}"),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Assigned Tenant Assignment ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Region is missing in the receiverTenant - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, "{}", fmt.Sprintf(assignedTenantFormatter, assignmentID, "{}")),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Receiver Tenant Region should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Subaccount ID is missing in the receiverTenant - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, `{"deploymentRegion": "region"}`, fmt.Sprintf(assignedTenantFormatter, assignmentID, `{}`)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Receiver Tenant Subaccount ID should be provided")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is assign and inboundCommunication is missing in the assignedTenant configuration - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, "{}"), fmt.Sprintf(assignedTenantFormatter, assignmentID, "{}")),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Assigned tenant inbound communication is missing in the configuration")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is assign and outboundCommunication is missing in the receiverTenant configuration - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, "{}"), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody("Receiver tenant outbound communication is missing")).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Operation is assign and receiverTenant has outboundCommunication but not in the same path as assignedTenant inboundCommunication - fails on validation",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithAssign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(`Receiver tenant outbound communication is missing - should be in \"credentials\" in the configuration`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
		{
			name:        "Success for unassign",
			requestBody: fmt.Sprintf(reqBodyFormatter, reqBodyContextWithUnassign, fmt.Sprintf(receiverTenantFormatter, region, subaccount, `{"credentials": {"another-field":{"credentials": {"outboundCommunication":{}}}}}`), fmt.Sprintf(assignedTenantFormatter, assignmentID, `{"credentials": {"inboundCommunication":{}}}`)),
			smClientFn: func() *automock.Client {
				client := &automock.Client{}
				client.On("DeleteMultipleResources", mock.Anything, region, subaccount, mock.Anything, &types.ServiceInstanceMatchParameters{ServiceInstanceName: assignmentID}).Return(nil).Once()
				return client
			},
			mtlsClientFn: func() *automock.MtlsHTTPClient {
				client := &automock.MtlsHTTPClient{}
				client.On("Do", requestThatHasBody(`{"state":"READY","configuration":null}`)).Return(fixHTTPResponse(http.StatusOK, ""), nil).Once()
				return client
			},
			expectedResponseCode: http.StatusAccepted,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			//GIVEN
			smClient := &automock.Client{}
			if testCase.smClientFn != nil {
				smClient = testCase.smClientFn()
			}
			mtlsClient := &automock.MtlsHTTPClient{}
			if testCase.mtlsClientFn != nil {
				mtlsClient = testCase.mtlsClientFn()
			}
			defer mock.AssertExpectationsForObjects(t, smClient)

			req, err := http.NewRequest(http.MethodPost, url+apiPath, bytes.NewBuffer([]byte(testCase.requestBody)))
			require.NoError(t, err)
			req.Header.Set("Location", statusUrl)

			h := handler.NewHandler(smClient, mtlsClient)
			recorder := httptest.NewRecorder()

			//WHEN
			h.HandlerFunc(recorder, req)
			resp := recorder.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			require.Equal(t, testCase.expectedResponseCode, resp.StatusCode, string(body))

			require.Eventually(t, func() bool {
				return mtlsClient.AssertExpectations(t)
			}, time.Second*5, 50*time.Millisecond)
		})
	}
}

func requestThatHasBody(expectedBody string) interface{} {
	return mock.MatchedBy(func(actualReq *http.Request) bool {
		bytes, err := io.ReadAll(actualReq.Body)
		if err != nil {
			return false
		}
		fmt.Printf("Expected Body %q\n", string(bytes))
		return strings.Contains(string(bytes), expectedBody)
	})
}

func fixHTTPResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
