package subscription

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	UserContextHeader  = "user_context"
	LocationHeader     = "Location"
	JobSucceededStatus = "COMPLETED"
	EventuallyTimeout  = 60 * time.Second
	EventuallyTick     = 2 * time.Second
)

func BuildAndExecuteUnsubscribeRequest(t *testing.T, resourceID, resourceName string, httpClient *http.Client, subscriptionURL, apiPath, subscriptionToken, propagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID, subscriptionFlow, subscriptionFlowHeaderKey string) {
	unsubscribeReq, err := http.NewRequest(http.MethodDelete, subscriptionURL+apiPath, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	unsubscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
	unsubscribeReq.Header.Add(propagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)
	unsubscribeReq.Header.Add(subscriptionFlowHeaderKey, subscriptionFlow)

	t.Logf("Removing subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, resourceName, resourceID, subscriptionProviderSubaccountID)
	unsubscribeResp, err := httpClient.Do(unsubscribeReq)
	require.NoError(t, err)
	unsubscribeBody, err := ioutil.ReadAll(unsubscribeResp.Body)
	require.NoError(t, err)

	body := string(unsubscribeBody)
	if strings.Contains(body, "job-not-created-yet") { // Check in the body if subscription is already removed if yes, do not perform unsubscription again because it will fail
		t.Logf("Subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q, and subaccount id: %q is alredy removed", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, resourceName, resourceID, subscriptionProviderSubaccountID)
		return
	}

	require.Equal(t, http.StatusAccepted, unsubscribeResp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", unsubscribeResp.StatusCode, http.StatusAccepted, body))
	unsubJobStatusPath := unsubscribeResp.Header.Get(LocationHeader)
	require.NotEmpty(t, unsubJobStatusPath)
	unsubJobStatusURL := subscriptionURL + unsubJobStatusPath
	require.Eventually(t, func() bool {
		return GetSubscriptionJobStatus(t, httpClient, unsubJobStatusURL, subscriptionToken) == JobSucceededStatus
	}, EventuallyTimeout, EventuallyTick)
	t.Logf("Successfully removed subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q, and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, resourceName, resourceID, subscriptionProviderSubaccountID)
}

func GetSubscriptionJobStatus(t *testing.T, httpClient *http.Client, jobStatusURL, token string) string {
	getJobReq, err := http.NewRequest(http.MethodGet, jobStatusURL, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	getJobReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", token))
	getJobReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)

	resp, err := httpClient.Do(getJobReq)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	status := gjson.GetBytes(respBody, "status")
	require.True(t, status.Exists())
	t.Logf("the status of the asynchronous job is: %s", status.String())

	jobErr := gjson.GetBytes(respBody, "error")
	if jobErr.Exists() {
		t.Errorf("Error occurred while executing asynchronous subscription job: %s", jobErr.String())
	}

	return status.String()
}

func BuildSubscriptionRequest(t *testing.T, subscriptionToken, subscriptionUrl, subscriptionProviderSubaccountID, subscriptionProviderAppNameValue, propagatedProviderSubaccountHeader, subscriptionFlow, subscriptionFlowHeaderKey string) *http.Request {
	apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", subscriptionProviderAppNameValue)
	subscribeReq, err := http.NewRequest(http.MethodPost, subscriptionUrl+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
	require.NoError(t, err)
	subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
	subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
	subscribeReq.Header.Add(propagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

	// This header is used by external services mock to determine to which saas application is the subscription.
	subscribeReq.Header.Add(subscriptionFlowHeaderKey, subscriptionFlow)

	return subscribeReq
}

func CreateSubscription(t *testing.T, conf Config, httpClient *http.Client, appTmpl graphql.ApplicationTemplate, apiPath, subscriptionToken, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID, subscriptionProviderAppNameValue string, expectedToPass, unsubscribeFirst bool, subscriptionFlow string) {
	subscribeReq := BuildSubscriptionRequest(t, subscriptionToken, conf.URL, subscriptionProviderSubaccountID, subscriptionProviderAppNameValue, conf.PropagatedProviderSubaccountHeader, subscriptionFlow, conf.SubscriptionFlowHeaderKey)

	if unsubscribeFirst {
		//unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		//In case there isn't subscription it will fail-safe without error
		BuildAndExecuteUnsubscribeRequest(t, appTmpl.ID, appTmpl.Name, httpClient, conf.URL, apiPath, subscriptionToken, conf.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID, subscriptionFlow, conf.SubscriptionFlowHeaderKey)
	}

	t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, appTmpl.Name, appTmpl.ID, subscriptionProviderSubaccountID)
	resp, err := httpClient.Do(subscribeReq)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	if !expectedToPass {
		require.Equal(t, http.StatusInternalServerError, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))
		t.Logf("As expected subscription was not created between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, appTmpl.Name, appTmpl.ID, subscriptionProviderSubaccountID)
		return
	}
	require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

	subJobStatusPath := resp.Header.Get(LocationHeader)
	require.NotEmpty(t, subJobStatusPath)
	subJobStatusURL := conf.URL + subJobStatusPath
	require.Eventually(t, func() bool {
		return GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == JobSucceededStatus
	}, EventuallyTimeout, EventuallyTick)
	t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, appTmpl.Name, appTmpl.ID, subscriptionProviderSubaccountID)
}

func CreateRuntimeSubscription(t *testing.T, conf Config, httpClient *http.Client, providerRuntime graphql.RuntimeExt, subscriptionToken, apiPath, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID, subscriptionProviderAppNameValue string, shouldUnsubscribeFirst bool, subscriptionFlow string) {
	subscribeReq := BuildSubscriptionRequest(t, subscriptionToken, conf.URL, subscriptionProviderSubaccountID, subscriptionProviderAppNameValue, conf.PropagatedProviderSubaccountHeader, subscriptionFlow, conf.SubscriptionFlowHeaderKey)

	if shouldUnsubscribeFirst {
		// unsubscribe request execution to ensure no resources/subscriptions are left unintentionally due to old unsubscribe failures or broken tests in the middle.
		// In case there isn't subscription it will fail-safe without error
		BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.URL, apiPath, subscriptionToken, conf.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID, conf.StandardFlow, conf.SubscriptionFlowHeaderKey)
	}

	t.Logf("Creating a subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
	resp, err := httpClient.Do(subscribeReq)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()
	require.NoError(t, err)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, http.StatusAccepted, string(body)))

	subJobStatusPath := resp.Header.Get(LocationHeader)
	require.NotEmpty(t, subJobStatusPath)
	subJobStatusURL := conf.URL + subJobStatusPath
	require.Eventually(t, func() bool {
		return GetSubscriptionJobStatus(t, httpClient, subJobStatusURL, subscriptionToken) == JobSucceededStatus
	}, EventuallyTimeout, EventuallyTick)
	t.Logf("Successfully created subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, providerRuntime.Name, providerRuntime.ID, subscriptionProviderSubaccountID)
}
