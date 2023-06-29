package subscription

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	UserContextHeader = "user_context"
	// SubscriptionFlowHeaderKey is the key for header in which the subscription flow for local test execution is specified
	SubscriptionFlowHeaderKey = "subscriptionFlow"
	LocationHeader            = "Location"
	JobSucceededStatus        = "COMPLETED"
	EventuallyTimeout         = 60 * time.Second
	EventuallyTick            = 2 * time.Second
	SubscriptionsLabelKey     = "subscriptions"
	subscriptionGUIDPath      = "subscriptionGUID"
	// StandardFlow subscribes to saas-instance which has CMP declared as dependency
	StandardFlow = "standard"
	// DirectDependencyFlow is used in subscription tests for subscribing to a SAAS applications that have CMP as indirect dependency.
	// This subscription flow is used when subscribing to saas-direct-dependency-instance which has CMP declared as dependency
	DirectDependencyFlow = "direct dependency"
	// IndirectDependencyFlow is used in subscription tests for subscribing to a SAAS applications that have CMP as indirect dependency.
	// This subscription flow is used when subscribing to saas-direct-dependency-instance through saas-root-instance
	IndirectDependencyFlow = "indirect dependency"
)

func BuildAndExecuteUnsubscribeRequest(t *testing.T, resourceID, resourceName string, httpClient *http.Client, subscriptionURL, apiPath, subscriptionToken, propagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID, subscriptionFlow string) string {
	unsubscribeReq, err := http.NewRequest(http.MethodDelete, subscriptionURL+apiPath, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	unsubscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
	unsubscribeReq.Header.Add(propagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)
	unsubscribeReq.Header.Add(SubscriptionFlowHeaderKey, subscriptionFlow)

	t.Logf("Removing subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, resourceName, resourceID, subscriptionProviderSubaccountID)
	unsubscribeResp, err := httpClient.Do(unsubscribeReq)
	require.NoError(t, err)
	unsubscribeBody, err := ioutil.ReadAll(unsubscribeResp.Body)
	require.NoError(t, err)

	body := string(unsubscribeBody)
	if strings.Contains(body, "job-not-created-yet") { // Check in the body if subscription is already removed if yes, do not perform unsubscription again because it will fail
		t.Logf("Subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q, and subaccount id: %q is alredy removed", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, resourceName, resourceID, subscriptionProviderSubaccountID)
		return ""
	}

	require.Equal(t, http.StatusAccepted, unsubscribeResp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", unsubscribeResp.StatusCode, http.StatusAccepted, body))
	unsubJobStatusPath := unsubscribeResp.Header.Get(LocationHeader)
	require.NotEmpty(t, unsubJobStatusPath)
	unsubJobStatusURL := subscriptionURL + unsubJobStatusPath
	require.Eventually(t, func() bool {
		return GetSubscriptionJobStatus(t, httpClient, unsubJobStatusURL, subscriptionToken) == JobSucceededStatus
	}, EventuallyTimeout, EventuallyTick)
	t.Logf("Successfully removed subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q, and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, resourceName, resourceID, subscriptionProviderSubaccountID)
	return gjson.Get(body, subscriptionGUIDPath).String()
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

func BuildSubscriptionRequest(t *testing.T, subscriptionToken, subscriptionUrl, subscriptionProviderSubaccountID, subscriptionProviderAppNameValue, propagatedProviderSubaccountHeader, subscriptionFlow string) *http.Request {
	apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", subscriptionProviderAppNameValue)
	subscribeReq, err := http.NewRequest(http.MethodPost, subscriptionUrl+apiPath, bytes.NewBuffer([]byte("{\"subscriptionParams\": {}}")))
	require.NoError(t, err)
	subscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
	subscribeReq.Header.Add(util.ContentTypeHeader, util.ContentTypeApplicationJSON)
	subscribeReq.Header.Add(propagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

	// This header is used when a subscription is made for an application and CMP is an indirect dependency
	subscribeReq.Header.Add(SubscriptionFlowHeaderKey, subscriptionFlow)

	return subscribeReq
}
