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
	UserContextHeader  = "user_context"
	LocationHeader     = "Location"
	JobSucceededStatus = "SUCCEEDED"
	EventuallyTimeout  = 60 * time.Second
	EventuallyTick     = 2 * time.Second
)

func BuildAndExecuteUnsubscribeRequest(t *testing.T, resourceID, resourceName string, httpClient *http.Client, subscriptionURL, apiPath, subscriptionToken, propagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID string) {
	unsubscribeReq, err := http.NewRequest(http.MethodDelete, subscriptionURL+apiPath, bytes.NewBuffer([]byte{}))
	require.NoError(t, err)
	unsubscribeReq.Header.Add(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", subscriptionToken))
	unsubscribeReq.Header.Add(propagatedProviderSubaccountHeader, subscriptionProviderSubaccountID)

	t.Logf("Removing subscription between consumer with subaccount id: %q and tenant id: %q, and provider with name: %q, id: %q and subaccount id: %q", subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, resourceName, resourceID, subscriptionProviderSubaccountID)
	unsubscribeResp, err := httpClient.Do(unsubscribeReq)
	require.NoError(t, err)
	unsubscribeBody, err := ioutil.ReadAll(unsubscribeResp.Body)
	require.NoError(t, err)

	body := string(unsubscribeBody)
	if strings.Contains(body, "does not exist") { // Check in the body if subscription is already removed if yes, do not perform unsubscription again because it will fail
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

	id := gjson.GetBytes(respBody, "id")
	state := gjson.GetBytes(respBody, "state")
	require.True(t, id.Exists())
	require.True(t, state.Exists())
	t.Logf("state of the asynchronous job with id: %s is: %s", id.String(), state.String())

	jobErr := gjson.GetBytes(respBody, "error")
	if jobErr.Exists() {
		t.Errorf("Error occurred while executing asynchronous subscription job: %s", jobErr.String())
	}

	return state.String()
}
