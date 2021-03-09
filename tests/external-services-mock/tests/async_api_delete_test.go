package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/client"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	gcli "github.com/machinebox/graphql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/webhook"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
)

func TestAsyncAPIDeleteApplication(t *testing.T) {
	var (
		operationFullPath = fmt.Sprintf("%s%s", testConfig.ExternalServicesMockBaseURL, "webhook/delete/operation")
		deleteFullPath    = fmt.Sprintf("%s%s", testConfig.ExternalServicesMockBaseURL, "webhook/delete")
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	appName := fmt.Sprintf("app-async-del-%s", time.Now().Format("060102150405"))
	appInput := graphql.ApplicationRegisterInput{
		Name:         appName,
		ProviderName: ptr.String("compass"),
		Webhooks: []*graphql.WebhookInput{
			{
				Type:           graphql.WebhookTypeUnregisterApplication,
				Mode:           webhookModePtr(graphql.WebhookModeAsync),
				URLTemplate:    str.Ptr(fmt.Sprintf("{ \\\"method\\\": \\\"DELETE\\\", \\\"path\\\": \\\"%s\\\" }", deleteFullPath)),
				OutputTemplate: str.Ptr(fmt.Sprintf("{ \\\"location\\\": \\\"%s\\\", \\\"success_status_code\\\": 200, \\\"error\\\": \\\"{{.Body.error}}\\\" }", operationFullPath)),
				StatusTemplate: str.Ptr("{ \\\"status\\\": \\\"{{.Body.status}}\\\", \\\"success_status_code\\\": 200, \\\"success_status_identifier\\\": \\\"SUCCEEDED\\\", \\\"in_progress_status_identifier\\\": \\\"IN_PROGRESS\\\", \\\"failed_status_identifier\\\": \\\"FAILED\\\", \\\"error\\\": \\\"{{.Body.error}}\\\" }"),
			},
		},
	}

	t.Log("Get Dex token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)
	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log(fmt.Sprintf("Registering application: %s", appName))
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(appInput)
	require.NoError(t, err)
	registerRequest := fixtures.FixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, testConfig.DefaultTenant, registerRequest, &app)
	require.NoError(t, err)
	nearCreationTime := time.Now().Add(-1 * time.Second)

	defer deleteApplicationOnExit(t, ctx, dexGraphQLClient, app.ID, testConfig.DefaultTenant)

	t.Log("Unlock the mock application webhook")
	unlockWebhook(t, operationFullPath)
	require.True(t, isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusOK), fmt.Sprintf("Expected state: %s", webhook.OperationResponseStatusOK))

	t.Log("Start async Delete of application")
	fixtures.UnregisterAsyncApplicationInTenant(t, ctx, dexGraphQLClient, app.ID, testConfig.DefaultTenant)

	t.Log("Prepare operation client for compass-system namespace")
	cfg, err := rest.InClusterConfig()
	require.NoError(t, err)
	k8sClient, err := client.NewForConfig(cfg)
	require.NoError(t, err)
	operationsK8sClient := k8sClient.Operations("compass-system")

	operationName := fmt.Sprintf("application-%s", app.ID)
	t.Log(fmt.Sprintf("Check operation CR with name %s is created", operationName))
	operation, err := operationsK8sClient.Get(ctx, operationName, metav1.GetOptions{})
	require.NoError(t, err)
	require.NotEmpty(t, operation)

	t.Log(fmt.Sprintf("Verify operation CR with name %s is in progress", operationName))
	require.Eventually(t, func() bool {
		return isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusINProgress)
	}, time.Minute*3, time.Second*5, "Waiting for state change timed out.")

	t.Log("Verify the application status in director is 'ready:false'")
	deletedApp := fixtures.GetApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)
	require.NoError(t, err)
	require.False(t, deletedApp.Ready, "Application is not in status ready:false")
	require.Empty(t, deletedApp.Error, "Application Error is not empty")

	t.Log("Verify DeletedAt in director is set and is in expected range")
	require.NotEmpty(t, deletedApp.DeletedAt, "Application Deletion time is not empty")
	deletedAtTime := time.Time(*deletedApp.DeletedAt)
	require.True(t, nearCreationTime.Before(deletedAtTime), "Deleted time is before creation time")
	require.True(t, time.Now().After(deletedAtTime), "Deleted time is in the future")

	t.Log("Unlock application webhook")
	unlockWebhook(t, operationFullPath)
	require.True(t, isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusOK), fmt.Sprintf("Expected state: %s", webhook.OperationResponseStatusOK))

	t.Log(fmt.Sprintf("Verify operation CR with name %s status condition is ConditionTypeReady", operationName))
	require.Eventually(t, func() bool {
		operation, err = operationsK8sClient.Get(ctx, operationName, metav1.GetOptions{})
		require.NoError(t, err)
		t.Log(fmt.Sprintf("The operation state is: %s", operation.Status.Phase))
		return isOperationDeletionCompleted(operation)
	}, time.Minute*3, time.Second*10, "Waiting for operation deletion timed out.")

	t.Log("Verify the deleted application do not exists in director")
	missingApp := fixtures.GetApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)
	require.Empty(t, missingApp.Name, "Application is not deleted")
}

func webhookModePtr(mode graphql.WebhookMode) *graphql.WebhookMode {
	return &mode
}

func unlockWebhook(t *testing.T, operationFullPath string) {
	httpClient := http.Client{}
	requestData := webhook.OperationStatusRequestData{
		InProgress: false,
	}
	jsonRequestData, err := json.Marshal(requestData)
	require.NoError(t, err)
	reqPost, err := http.NewRequest(http.MethodPost, operationFullPath, bytes.NewBuffer(jsonRequestData))
	require.NoError(t, err)
	respPost, err := httpClient.Do(reqPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respPost.StatusCode)
}

func isWebhookOperationInDesiredState(t *testing.T, operationFullPath, desiredState string) (isInState bool) {
	httpClient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, operationFullPath, nil)
	require.NoError(t, err)
	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	fullBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	fullBodyString := string(fullBody)
	responseState := strings.Contains(fullBodyString, desiredState)
	if !responseState {
		t.Log(fullBodyString)
	}
	return responseState
}

func isOperationDeletionCompleted(operation *v1alpha1.Operation) bool {
	if operation.Status.Phase == v1alpha1.StateSuccess || operation.Status.Phase == v1alpha1.StateFailed {
		return true
	}
	return false
}

func deleteApplicationOnExit(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string, tenant string) {
	application := fixtures.GetApplication(t, ctx, gqlClient, testConfig.DefaultTenant, id)
	if application.Name != "" {
		fixtures.UnregisterApplication(t, ctx, gqlClient, id, tenant)
	}
}
