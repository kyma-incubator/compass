package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/instance-creator/internal/client"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/automock"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/client/types"
	"github.com/kyma-incubator/compass/components/instance-creator/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testConfig               = fixConfig()
	testConfigWithInvalidURL = fixConfigWithInvalidURL()
	invalidResponseBody      = []byte(`{"key: value}`)
	emptyResponseBody        = []byte(`{}`)
)

func TestClient_RetrieveServiceOffering(t *testing.T) {
	ctx := context.TODO()

	respBody, err := json.Marshal(fixServiceOfferings())
	require.NoError(t, err)

	respBodyWithNoMatchingCatalogName, err := json.Marshal(fixServiceOfferingsWithNoMatchingCatalogName())
	require.NoError(t, err)

	testCases := []struct {
		Name             string
		CallerProvider   func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider
		Config           config.Config
		ExpectedResult   string
		ExpectedErrorMsg string
	}{
		{
			Name:           "Success",
			CallerProvider: callerThatGetsCalledOnce(http.StatusOK, respBody),
			Config:         testConfig,
			ExpectedResult: offeringID,
		},
		{
			Name:             "Error when building URL",
			CallerProvider:   callerThatDoesNotGetCalled,
			Config:           testConfigWithInvalidURL,
			ExpectedErrorMsg: "while building service offerings URL",
		},
		{
			Name:             "Error when getting caller fails",
			CallerProvider:   callerProviderThatFails,
			Config:           testConfig,
			ExpectedErrorMsg: "while getting caller for region:",
		},
		{
			Name:             "Error when caller fails",
			CallerProvider:   callerThatDoesNotSucceed,
			Config:           testConfig,
			ExpectedErrorMsg: "while executing request for retrieving service offerings for subaccount with ID:",
		},
		{
			Name:             "Error when response status code is not 200 OK",
			CallerProvider:   callerThatReturnsBadStatus,
			Config:           testConfig,
			ExpectedErrorMsg: "failed to get object(s), status:",
		},
		{
			Name:             "Error when unmarshalling response body",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusOK, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal service offerings:",
		},
		{
			Name:             "Error when offering ID is empty",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusOK, respBodyWithNoMatchingCatalogName),
			Config:           testConfig,
			ExpectedErrorMsg: "couldn't find service offering for catalog name:",
		},
	}

	for _, testCase := range testCases {
		// GIVEN
		callerProviderSvc := testCase.CallerProvider(t, testCase.Config, region)
		defer callerProviderSvc.AssertExpectations(t)

		testClient := client.NewClient(testCase.Config, callerProviderSvc)

		// WHEN
		resultOfferingID, err := testClient.RetrieveServiceOffering(ctx, region, catalogName, subaccountID)

		// THEN
		if testCase.ExpectedErrorMsg != "" {
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
		} else {
			assert.Equal(t, testCase.ExpectedResult, resultOfferingID)
			assert.NoError(t, err)
		}
	}
}

func TestClient_RetrieveServicePlan(t *testing.T) {
	ctx := context.TODO()

	respBody, err := json.Marshal(fixServicePlans())
	require.NoError(t, err)

	respBodyWithNoMatchingCatalogName, err := json.Marshal(fixServicePlansWithNoMatchingCatalogNameAndOfferingID())
	require.NoError(t, err)

	testCases := []struct {
		Name             string
		CallerProvider   func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider
		Config           config.Config
		ExpectedResult   string
		ExpectedErrorMsg string
	}{
		{
			Name:           "Success",
			CallerProvider: callerThatGetsCalledOnce(http.StatusOK, respBody),
			Config:         testConfig,
			ExpectedResult: planID,
		},
		{
			Name:             "Error when building URL",
			CallerProvider:   callerThatDoesNotGetCalled,
			Config:           testConfigWithInvalidURL,
			ExpectedErrorMsg: "while building service plans URL",
		},
		{
			Name:             "Error when getting caller fails",
			CallerProvider:   callerProviderThatFails,
			Config:           testConfig,
			ExpectedErrorMsg: "while getting caller for region:",
		},
		{
			Name:             "Error when caller fails",
			CallerProvider:   callerThatDoesNotSucceed,
			Config:           testConfig,
			ExpectedErrorMsg: "while executing request for retrieving service plans for subaccount with ID:",
		},
		{
			Name:             "Error when response status code is not 200 OK",
			CallerProvider:   callerThatReturnsBadStatus,
			Config:           testConfig,
			ExpectedErrorMsg: "failed to get object(s), status:",
		},
		{
			Name:             "Error when unmarshalling response body",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusOK, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal service plans:",
		},
		{
			Name:             "Error when plan ID is empty",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusOK, respBodyWithNoMatchingCatalogName),
			Config:           testConfig,
			ExpectedErrorMsg: "couldn't find service plan for catalog name:",
		},
	}

	for _, testCase := range testCases {
		// GIVEN
		callerProviderSvc := testCase.CallerProvider(t, testCase.Config, region)
		defer callerProviderSvc.AssertExpectations(t)

		testClient := client.NewClient(testCase.Config, callerProviderSvc)

		// WHEN
		resultPlanID, err := testClient.RetrieveServicePlan(ctx, region, planName, offeringID, subaccountID)

		// THEN
		if testCase.ExpectedErrorMsg != "" {
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
		} else {
			assert.Equal(t, testCase.ExpectedResult, resultPlanID)
			assert.NoError(t, err)
		}
	}
}

func TestClient_RetrieveServiceKeyByID(t *testing.T) {
	ctx := context.TODO()

	respBody, err := json.Marshal(fixServiceKey())
	require.NoError(t, err)

	testCases := []struct {
		Name             string
		CallerProvider   func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider
		Config           config.Config
		ExpectedResult   *types.ServiceKey
		ExpectedErrorMsg string
	}{
		{
			Name:           "Success",
			CallerProvider: callerThatGetsCalledOnce(http.StatusOK, respBody),
			Config:         testConfig,
			ExpectedResult: fixServiceKey(),
		},
		{
			Name:             "Error when building URL",
			CallerProvider:   callerThatDoesNotGetCalled,
			Config:           testConfigWithInvalidURL,
			ExpectedErrorMsg: "while building service binding URL",
		},
		{
			Name:             "Error when getting caller fails",
			CallerProvider:   callerProviderThatFails,
			Config:           testConfig,
			ExpectedErrorMsg: "while getting caller for region:",
		},
		{
			Name:             "Error when caller fails",
			CallerProvider:   callerThatDoesNotSucceed,
			Config:           testConfig,
			ExpectedErrorMsg: "while executing request for retrieving service key for subaccount with ID:",
		},
		{
			Name:             "Error when response status code is not 200 OK",
			CallerProvider:   callerThatReturnsBadStatus,
			Config:           testConfig,
			ExpectedErrorMsg: "failed to get object(s), status:",
		},
		{
			Name:             "Error when unmarshalling response body",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusOK, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal service key:",
		},
	}

	for _, testCase := range testCases {
		// GIVEN
		callerProviderSvc := testCase.CallerProvider(t, testCase.Config, region)
		defer callerProviderSvc.AssertExpectations(t)

		testClient := client.NewClient(testCase.Config, callerProviderSvc)

		// WHEN
		resultServiceKey, err := testClient.RetrieveServiceKeyByID(ctx, region, serviceKeyID, subaccountID)

		// THEN
		if testCase.ExpectedErrorMsg != "" {
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
		} else {
			assert.Equal(t, testCase.ExpectedResult, resultServiceKey)
			assert.NoError(t, err)
		}
	}
}

func TestClient_RetrieveServiceInstanceIDByName(t *testing.T) {
	ctx := context.TODO()

	respBody, err := json.Marshal(fixServiceInstances())
	require.NoError(t, err)

	respBodyWithNoMatchingName, err := json.Marshal(fixServiceInstancesWithNoMatchingName())
	require.NoError(t, err)

	testCases := []struct {
		Name             string
		CallerProvider   func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider
		Config           config.Config
		ExpectedResult   string
		ExpectedErrorMsg string
	}{
		{
			Name:           "Success",
			CallerProvider: callerThatGetsCalledOnce(http.StatusOK, respBody),
			Config:         testConfig,
			ExpectedResult: serviceInstanceID,
		},
		{
			Name:             "Error when building URL",
			CallerProvider:   callerThatDoesNotGetCalled,
			Config:           testConfigWithInvalidURL,
			ExpectedErrorMsg: "while building service instances URL",
		},
		{
			Name:             "Error when getting caller fails",
			CallerProvider:   callerProviderThatFails,
			Config:           testConfig,
			ExpectedErrorMsg: "while getting caller for region:",
		},
		{
			Name:             "Error when caller fails",
			CallerProvider:   callerThatDoesNotSucceed,
			Config:           testConfig,
			ExpectedErrorMsg: "while executing request for retrieving service instances for subaccount with ID:",
		},
		{
			Name:             "Error when response status code is not 200 OK",
			CallerProvider:   callerThatReturnsBadStatus,
			Config:           testConfig,
			ExpectedErrorMsg: "failed to get object(s), status:",
		},
		{
			Name:             "Error when unmarshalling response body",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusOK, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal service instances:",
		},
		{
			Name:           "Success when instance ID is empty",
			CallerProvider: callerThatGetsCalledOnce(http.StatusOK, respBodyWithNoMatchingName),
			Config:         testConfig,
			ExpectedResult: "",
		},
	}

	for _, testCase := range testCases {
		// GIVEN
		callerProviderSvc := testCase.CallerProvider(t, testCase.Config, region)
		defer callerProviderSvc.AssertExpectations(t)

		testClient := client.NewClient(testCase.Config, callerProviderSvc)

		// WHEN
		resultServiceInstanceID, err := testClient.RetrieveServiceInstanceIDByName(ctx, region, serviceInstanceName, subaccountID)

		// THEN
		if testCase.ExpectedErrorMsg != "" {
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
		} else {
			assert.Equal(t, testCase.ExpectedResult, resultServiceInstanceID)
			assert.NoError(t, err)
		}
	}
}

func TestClient_CreateServiceInstance(t *testing.T) {
	ctx := context.TODO()

	respBody, err := json.Marshal(fixServiceInstance(serviceInstanceID, serviceInstanceName))
	require.NoError(t, err)

	inProgressResponseBody, err := json.Marshal(fixOperationStatus(types.OperationStateInProgress, ""))
	require.NoError(t, err)

	opRespBody, err := json.Marshal(fixOperationStatus(types.OperationStateSucceeded, serviceInstanceID))
	require.NoError(t, err)

	opRespBodyWithFailedState, err := json.Marshal(fixOperationStatus(types.OperationStateFailed, serviceInstanceID))
	require.NoError(t, err)

	opRespBodyWithEmptyResourceID, err := json.Marshal(fixOperationStatusWithEmptyResourceID(types.OperationStateSucceeded))
	require.NoError(t, err)

	serviceInstanceWithNoID := fixServiceInstance(serviceInstanceID, serviceInstanceName)
	serviceInstanceWithNoID.ID = ""
	respBodyWithEmptyID, err := json.Marshal(serviceInstanceWithNoID)
	require.NoError(t, err)

	testCases := []struct {
		Name             string
		CallerProvider   func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider
		Config           config.Config
		InvalidParams    bool
		ExpectedResult   string
		ExpectedErrorMsg string
	}{
		// Sync flow
		{
			Name:           "Success in the synchronous case",
			CallerProvider: callerThatGetsCalledOnce(http.StatusCreated, respBody),
			Config:         testConfig,
			ExpectedResult: serviceInstanceID,
		},
		{
			Name:             "Error when marshalling request body",
			CallerProvider:   callerThatDoesNotGetCalled,
			Config:           testConfig,
			InvalidParams:    true,
			ExpectedErrorMsg: "failed to marshal service instance body",
		},
		{
			Name:             "Error when building URL",
			CallerProvider:   callerThatDoesNotGetCalled,
			Config:           testConfigWithInvalidURL,
			ExpectedErrorMsg: "while building service instances URL",
		},
		{
			Name:             "Error when getting caller fails",
			CallerProvider:   callerProviderThatFails,
			Config:           testConfig,
			ExpectedErrorMsg: "while getting caller for region:",
		},
		{
			Name:             "Error when caller fails",
			CallerProvider:   callerThatDoesNotSucceed,
			Config:           testConfig,
			ExpectedErrorMsg: testError.Error(),
		},
		{
			Name:             "Error when response status code is not 201 or 202",
			CallerProvider:   callerThatReturnsBadStatus,
			Config:           testConfig,
			ExpectedErrorMsg: "failed to create service instance, status:",
		},
		{
			Name:             "Error when unmarshalling response body",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusCreated, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal service instance:",
		},
		{
			Name:             "Error when service instance ID is empty",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusCreated, respBodyWithEmptyID),
			Config:           testConfig,
			ExpectedErrorMsg: "the service instance ID could not be empty",
		},
		// Async flow
		{
			Name:           "Success in the asynchronous case",
			CallerProvider: callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, respBody, opRespBody),
			Config:         testConfig,
			ExpectedResult: serviceInstanceID,
		},
		{
			Name:           "Success in the asynchronous case when initially operation state is 'In Progress'",
			CallerProvider: callerThatGetsCalledSeveralTimesInAsyncCase([]int{http.StatusAccepted, http.StatusOK, http.StatusOK}, [][]byte{respBody, inProgressResponseBody, opRespBody}, false),
			Config:         testConfig,
			ExpectedResult: serviceInstanceID,
		},
		{
			Name:             "Error when location header key is missing",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusAccepted, respBody),
			Config:           testConfig,
			ExpectedErrorMsg: fmt.Sprintf("the operation status path from %s header should not be empty", locationHeaderKey),
		},
		{
			Name:             "Error when caller fails the second time it is called",
			CallerProvider:   callerThatDoesNotSucceedTheLastTimeWhenCalled([]int{http.StatusAccepted}, [][]byte{respBody}),
			Config:           testConfig,
			ExpectedErrorMsg: "while handling asynchronous creation of service instance with name:",
		},
		{
			Name:             "Error when operation response status code is not 200 OK",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusBadRequest, respBody, opRespBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to get asynchronous object operation status. Received status:",
		},
		{
			Name:             "Error when unmarshalling operation status",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, respBody, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal object operation status:",
		},
		{
			Name:             "Error when operation state is not 'Succeeded' or 'In Progress'",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, respBody, opRespBodyWithFailedState),
			Config:           testConfig,
			ExpectedErrorMsg: fmt.Sprintf("the asynchronous object operation finished with state: %q", types.OperationStateFailed),
		},
		{
			Name:             "Error when returned service instance ID is empty",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, respBody, opRespBodyWithEmptyResourceID),
			Config:           testConfig,
			ExpectedErrorMsg: "the service instance ID could not be empty",
		},
	}

	for _, testCase := range testCases {
		// GIVEN
		callerProviderSvc := testCase.CallerProvider(t, testCase.Config, region)
		defer callerProviderSvc.AssertExpectations(t)

		testClient := client.NewClient(testCase.Config, callerProviderSvc)

		inputParams := parameters
		if testCase.InvalidParams {
			inputParams = invalidParameters
		}

		// WHEN
		resultServiceInstanceID, err := testClient.CreateServiceInstance(ctx, region, serviceInstanceName, planID, subaccountID, inputParams)

		// THEN
		if testCase.ExpectedErrorMsg != "" {
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
		} else {
			assert.Equal(t, testCase.ExpectedResult, resultServiceInstanceID)
			assert.NoError(t, err)
		}
	}
}

func TestClient_CreateServiceKey(t *testing.T) {
	ctx := context.TODO()

	respBody, err := json.Marshal(fixServiceKey())
	require.NoError(t, err)

	inProgressResponseBody, err := json.Marshal(fixOperationStatus(types.OperationStateInProgress, ""))
	require.NoError(t, err)

	opRespBody, err := json.Marshal(fixOperationStatus(types.OperationStateSucceeded, serviceKeyID))
	require.NoError(t, err)

	opRespBodyWithFailedState, err := json.Marshal(fixOperationStatus(types.OperationStateFailed, serviceKeyID))
	require.NoError(t, err)

	opRespBodyWithEmptyResourceID, err := json.Marshal(fixOperationStatusWithEmptyResourceID(types.OperationStateSucceeded))
	require.NoError(t, err)

	serviceKeyWithNoID := fixServiceKey()
	serviceKeyWithNoID.ID = ""
	respBodyWithEmptyID, err := json.Marshal(serviceKeyWithNoID)
	require.NoError(t, err)

	testCases := []struct {
		Name             string
		CallerProvider   func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider
		Config           config.Config
		InvalidParams    bool
		ExpectedResult   string
		ExpectedErrorMsg string
	}{
		// Sync flow
		{
			Name:           "Success in the synchronous case",
			CallerProvider: callerThatGetsCalledOnce(http.StatusCreated, respBody),
			Config:         testConfig,
			ExpectedResult: serviceKeyID,
		},
		{
			Name:             "Error when marshalling request body",
			CallerProvider:   callerThatDoesNotGetCalled,
			Config:           testConfig,
			InvalidParams:    true,
			ExpectedErrorMsg: "failed to marshal service key body",
		},
		{
			Name:             "Error when building URL",
			CallerProvider:   callerThatDoesNotGetCalled,
			Config:           testConfigWithInvalidURL,
			ExpectedErrorMsg: "while building service bindings URL",
		},
		{
			Name:             "Error when getting caller fails",
			CallerProvider:   callerProviderThatFails,
			Config:           testConfig,
			ExpectedErrorMsg: "while getting caller for region:",
		},
		{
			Name:             "Error when caller fails",
			CallerProvider:   callerThatDoesNotSucceed,
			Config:           testConfig,
			ExpectedErrorMsg: testError.Error(),
		},
		{
			Name:             "Error when response status code is not 201 or 202",
			CallerProvider:   callerThatReturnsBadStatus,
			Config:           testConfig,
			ExpectedErrorMsg: "failed to create service key, status:",
		},
		{
			Name:             "Error when unmarshalling response body",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusCreated, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal service key:",
		},
		{
			Name:             "Error when service instance ID is empty",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusCreated, respBodyWithEmptyID),
			Config:           testConfig,
			ExpectedErrorMsg: "the service key ID could not be empty",
		},
		// Async flow
		{
			Name:           "Success in the asynchronous case",
			CallerProvider: callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, respBody, opRespBody),
			Config:         testConfig,
			ExpectedResult: serviceKeyID,
		},
		{
			Name:           "Success in the asynchronous case when initially operation state is 'In Progress'",
			CallerProvider: callerThatGetsCalledSeveralTimesInAsyncCase([]int{http.StatusAccepted, http.StatusOK, http.StatusOK}, [][]byte{respBody, inProgressResponseBody, opRespBody}, false),
			Config:         testConfig,
			ExpectedResult: serviceKeyID,
		},
		{
			Name:             "Error when location header key is missing",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusAccepted, respBody),
			Config:           testConfig,
			ExpectedErrorMsg: fmt.Sprintf("the operation status path from %s header should not be empty", locationHeaderKey),
		},
		{
			Name:             "Error when caller fails the second time it is called",
			CallerProvider:   callerThatDoesNotSucceedTheLastTimeWhenCalled([]int{http.StatusAccepted}, [][]byte{respBody}),
			Config:           testConfig,
			ExpectedErrorMsg: "while handling asynchronous creation of service key for service instance with ID:",
		},
		{
			Name:             "Error when operation response status code is not 200 OK",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusBadRequest, respBody, opRespBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to get asynchronous object operation status. Received status:",
		},
		{
			Name:             "Error when unmarshalling operation status",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, respBody, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal object operation status:",
		},
		{
			Name:             "Error when operation state is not 'Succeeded' or 'In Progress'",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, respBody, opRespBodyWithFailedState),
			Config:           testConfig,
			ExpectedErrorMsg: fmt.Sprintf("the asynchronous object operation finished with state: %q", types.OperationStateFailed),
		},
		{
			Name:             "Error when returned service instance ID is empty",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, respBody, opRespBodyWithEmptyResourceID),
			Config:           testConfig,
			ExpectedErrorMsg: "the service key ID could not be empty",
		},
	}

	for _, testCase := range testCases {
		// GIVEN
		callerProviderSvc := testCase.CallerProvider(t, testCase.Config, region)
		defer callerProviderSvc.AssertExpectations(t)

		testClient := client.NewClient(testCase.Config, callerProviderSvc)

		inputParams := parameters
		if testCase.InvalidParams {
			inputParams = invalidParameters
		}

		// WHEN
		resultServiceKeyID, err := testClient.CreateServiceKey(ctx, region, serviceKeyName, serviceInstanceID, subaccountID, inputParams)

		// THEN
		if testCase.ExpectedErrorMsg != "" {
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
		} else {
			assert.Equal(t, testCase.ExpectedResult, resultServiceKeyID)
			assert.NoError(t, err)
		}
	}
}

func TestClient_DeleteServiceInstance(t *testing.T) {
	ctx := context.TODO()

	inProgressResponseBody, err := json.Marshal(fixOperationStatus(types.OperationStateInProgress, ""))
	require.NoError(t, err)

	opRespBody, err := json.Marshal(fixOperationStatusWithEmptyResourceID(types.OperationStateSucceeded))
	require.NoError(t, err)

	opRespBodyWithFailedState, err := json.Marshal(fixOperationStatusWithEmptyResourceID(types.OperationStateFailed))
	require.NoError(t, err)

	testCases := []struct {
		Name             string
		CallerProvider   func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider
		Config           config.Config
		ExpectedErrorMsg string
	}{
		// Sync flow
		{
			Name:           "Success in the synchronous case",
			CallerProvider: callerThatGetsCalledOnce(http.StatusOK, emptyResponseBody),
			Config:         testConfig,
		},
		{
			Name:             "Error when building URL",
			CallerProvider:   callerThatDoesNotGetCalled,
			Config:           testConfigWithInvalidURL,
			ExpectedErrorMsg: "while building service instances URL",
		},
		{
			Name:             "Error when getting caller fails",
			CallerProvider:   callerProviderThatFails,
			Config:           testConfig,
			ExpectedErrorMsg: "while getting caller for region:",
		},
		{
			Name:             "Error when caller fails",
			CallerProvider:   callerThatDoesNotSucceed,
			Config:           testConfig,
			ExpectedErrorMsg: testError.Error(),
		},
		{
			Name:             "Error when response status code is not 200 or 202",
			CallerProvider:   callerThatReturnsBadStatus,
			Config:           testConfig,
			ExpectedErrorMsg: "failed to delete service instance, status:",
		},
		// Async flow
		{
			Name:           "Success in the asynchronous case",
			CallerProvider: callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, emptyResponseBody, opRespBody),
			Config:         testConfig,
		},
		{
			Name:           "Success in the asynchronous case when initially operation state is 'In Progress'",
			CallerProvider: callerThatGetsCalledSeveralTimesInAsyncCase([]int{http.StatusAccepted, http.StatusOK, http.StatusOK}, [][]byte{emptyResponseBody, inProgressResponseBody, opRespBody}, false),
			Config:         testConfig,
		},
		{
			Name:             "Error when location header key is missing",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusAccepted, emptyResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: fmt.Sprintf("the operation status path from %s header should not be empty", locationHeaderKey),
		},
		{
			Name:             "Error when caller fails the second time it is called",
			CallerProvider:   callerThatDoesNotSucceedTheLastTimeWhenCalled([]int{http.StatusAccepted}, [][]byte{emptyResponseBody}),
			Config:           testConfig,
			ExpectedErrorMsg: "while deleting service instance with ID:",
		},
		{
			Name:             "Error when operation response status code is not 200 OK",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusBadRequest, emptyResponseBody, opRespBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to get asynchronous object operation status. Received status:",
		},
		{
			Name:             "Error when unmarshalling operation status",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, emptyResponseBody, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal object operation status:",
		},
		{
			Name:             "Error when operation state is not 'Succeeded' or 'In Progress'",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusAccepted, http.StatusOK, emptyResponseBody, opRespBodyWithFailedState),
			Config:           testConfig,
			ExpectedErrorMsg: fmt.Sprintf("the asynchronous object operation finished with state: %q", types.OperationStateFailed),
		},
	}

	for _, testCase := range testCases {
		// GIVEN
		callerProviderSvc := testCase.CallerProvider(t, testCase.Config, region)
		defer callerProviderSvc.AssertExpectations(t)

		testClient := client.NewClient(testCase.Config, callerProviderSvc)

		// WHEN
		err := testClient.DeleteServiceInstance(ctx, region, serviceInstanceID, serviceInstanceName, subaccountID)

		// THEN
		if testCase.ExpectedErrorMsg != "" {
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
		} else {
			assert.NoError(t, err)
		}
	}
}

func TestClient_DeleteServiceKeys(t *testing.T) {
	ctx := context.TODO()

	respBody, err := json.Marshal(fixServiceKeys())
	require.NoError(t, err)

	opRespBody, err := json.Marshal(fixOperationStatusWithEmptyResourceID(types.OperationStateSucceeded))
	require.NoError(t, err)

	inProgressResponseBody, err := json.Marshal(fixOperationStatus(types.OperationStateInProgress, ""))
	require.NoError(t, err)

	opRespBodyWithFailedState, err := json.Marshal(fixOperationStatusWithEmptyResourceID(types.OperationStateFailed))
	require.NoError(t, err)

	testCases := []struct {
		Name             string
		CallerProvider   func(t *testing.T, cfg config.Config, region string) *automock.ExternalSvcCallerProvider
		Config           config.Config
		ExpectedResult   error
		ExpectedErrorMsg string
	}{
		// Sync flow
		{
			Name:           "Success in the synchronous case",
			CallerProvider: callerThatGetsCalledTwice(http.StatusOK, http.StatusOK, respBody, emptyResponseBody),
			Config:         testConfig,
		},
		{
			Name:             "Error when getting caller fails",
			CallerProvider:   callerProviderThatFails,
			Config:           testConfig,
			ExpectedErrorMsg: "while getting caller for region:",
		},
		{
			Name:             "Error when building URL",
			CallerProvider:   callerThatDoesNotGetCalledButProviderIs,
			Config:           testConfigWithInvalidURL,
			ExpectedErrorMsg: "while building service binding URL",
		},
		{
			Name:             "Error when caller fails",
			CallerProvider:   callerThatDoesNotSucceed,
			Config:           testConfig,
			ExpectedErrorMsg: testError.Error(),
		},
		{
			Name:             "Error when response status code is not 200",
			CallerProvider:   callerThatReturnsBadStatus,
			Config:           testConfig,
			ExpectedErrorMsg: "failed to get service bindings, status:",
		},
		{
			Name:             "Error when unmarshalling service keys",
			CallerProvider:   callerThatGetsCalledOnce(http.StatusOK, invalidResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal service keys:",
		},
		{
			Name:             "Error when caller fails on the second call",
			CallerProvider:   callerThatDoesNotSucceedTheLastTimeWhenCalled([]int{http.StatusOK}, [][]byte{respBody}),
			Config:           testConfig,
			ExpectedErrorMsg: testError.Error(),
		},
		{
			Name:             "Error when caller does not return 200 or 202 on the second call",
			CallerProvider:   callerThatGetsCalledTwice(http.StatusOK, http.StatusBadRequest, respBody, emptyResponseBody),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to delete service binding, status:",
		},
		// Async flow
		{
			Name:           "Success in the asynchronous case",
			CallerProvider: callerThatGetsCalledSeveralTimesInAsyncCase([]int{http.StatusOK, http.StatusAccepted, http.StatusOK}, [][]byte{respBody, emptyResponseBody, opRespBody}, false),
			Config:         testConfig,
		},
		{
			Name:           "Success in the asynchronous case when initially operation state is 'In Progress'",
			CallerProvider: callerThatGetsCalledSeveralTimesInAsyncCase([]int{http.StatusOK, http.StatusAccepted, http.StatusOK, http.StatusOK}, [][]byte{respBody, emptyResponseBody, inProgressResponseBody, opRespBody}, false),
			Config:         testConfig,
		},
		{
			Name:             "Error when location header key is missing",
			CallerProvider:   callerThatGetsCalledSeveralTimesInAsyncCase([]int{http.StatusOK, http.StatusAccepted}, [][]byte{respBody, emptyResponseBody}, true),
			Config:           testConfig,
			ExpectedErrorMsg: fmt.Sprintf("the operation status path from %s header should not be empty", locationHeaderKey),
		},
		{
			Name:             "Error when caller fails the last time it is called",
			CallerProvider:   callerThatDoesNotSucceedTheLastTimeWhenCalled([]int{http.StatusOK, http.StatusAccepted}, [][]byte{respBody, emptyResponseBody}),
			Config:           testConfig,
			ExpectedErrorMsg: "while deleting service binding with ID:",
		},
		{
			Name:             "Error when operation response status code is not 200 OK",
			CallerProvider:   callerThatGetsCalledSeveralTimesInAsyncCase([]int{http.StatusOK, http.StatusAccepted, http.StatusBadRequest}, [][]byte{respBody, emptyResponseBody, emptyResponseBody}, false),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to get asynchronous object operation status. Received status:",
		},
		{
			Name:             "Error when unmarshalling operation status",
			CallerProvider:   callerThatGetsCalledSeveralTimesInAsyncCase([]int{http.StatusOK, http.StatusAccepted, http.StatusOK}, [][]byte{respBody, emptyResponseBody, invalidResponseBody}, false),
			Config:           testConfig,
			ExpectedErrorMsg: "failed to unmarshal object operation status:",
		},
		{
			Name:             "Error when operation state is not 'Succeeded' or 'In Progress'",
			CallerProvider:   callerThatGetsCalledSeveralTimesInAsyncCase([]int{http.StatusOK, http.StatusAccepted, http.StatusOK}, [][]byte{respBody, emptyResponseBody, opRespBodyWithFailedState}, false),
			Config:           testConfig,
			ExpectedErrorMsg: fmt.Sprintf("the asynchronous object operation finished with state: %q", types.OperationStateFailed),
		},
	}

	for _, testCase := range testCases {
		// GIVEN
		callerProviderSvc := testCase.CallerProvider(t, testCase.Config, region)
		defer callerProviderSvc.AssertExpectations(t)

		testClient := client.NewClient(testCase.Config, callerProviderSvc)

		// WHEN
		err := testClient.DeleteServiceKeys(ctx, region, serviceInstanceID, serviceInstanceName, subaccountID)

		// THEN
		if testCase.ExpectedErrorMsg != "" {
			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
		} else {
			assert.NoError(t, err)
		}
	}
}
