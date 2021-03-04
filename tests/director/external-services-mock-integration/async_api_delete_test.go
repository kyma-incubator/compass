package external_services_mock_integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/client"
	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"
	gcli "github.com/machinebox/graphql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/webhook"
	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"
)

const (
	operationPath = "webhook/delete/operation"
	deletePath    = "webhook/delete"
)

func TestAsyncAPIDeleteApplication(t *testing.T) {
	var (
		operationFullPath = fmt.Sprintf("%s%s", testConfig.ExternalServicesMockBaseURL, operationPath)
		deleteFullPath    = fmt.Sprintf("%s%s", testConfig.ExternalServicesMockBaseURL, deletePath)
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
				URLTemplate:    str.Ptr(fmt.Sprintf("{ \\\"method\\\": \\\"GET\\\", \\\"path\\\": \\\"%s\\\" }", deleteFullPath)),
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
	appInputGQL, err := tc.Graphqlizer.ApplicationRegisterInputToGQL(appInput)
	require.NoError(t, err)
	registerRequest := fixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}
	err = tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, testConfig.DefaultTenant, registerRequest, &app)
	require.NoError(t, err)

	defer deleteApplicationOnExit(t, ctx, dexGraphQLClient, app.ID, testConfig.DefaultTenant)

	t.Log("Unlock the mock application webhook")
	unlockWebhook(t, operationFullPath)
	isInDesiredState, webhookResponse := isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusOK)
	require.True(t, isInDesiredState, fmt.Sprintf("Webhook responded with: %s. Expected state: %s", webhookResponse, webhook.OperationResponseStatusOK))

	t.Log("Start async Delete of application")
	unregisterAsyncApplicationInTenant(t, ctx, dexGraphQLClient, app.ID, testConfig.DefaultTenant)

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
	counter := 0
	isInDesiredState, webhookResponse = isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusINProgress)
	for (!isInDesiredState) && counter < 100 {
		t.Log(fmt.Sprintf("[%s] Webhook responded with: %s. Expected state: %s", time.Now().Format("06.01.02 15:04:05"), webhookResponse, webhook.OperationResponseStatusINProgress))
		time.Sleep(time.Second * 5)
		isInDesiredState, webhookResponse = isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusINProgress)
		counter++
	}
	isInDesiredState, webhookResponse = isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusINProgress)
	require.True(t, isInDesiredState, fmt.Sprintf("Webhook responded with: %s. Expected state: %s", webhookResponse, webhook.OperationResponseStatusINProgress))

	t.Log("Verify the application status in director is 'ready:false'")
	application := getApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)
	require.False(t, application.Ready, "Application is not in status ready:false")

	t.Log("Unlock application webhook")
	unlockWebhook(t, operationFullPath)
	isInDesiredState, webhookResponse = isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusOK)
	require.True(t, isInDesiredState, fmt.Sprintf("Webhook responded with: %s. Expected state: %s", webhookResponse, webhook.OperationResponseStatusOK))

	t.Log(fmt.Sprintf("Verify operation CR with name %s status condition should be ConditionTypeReady", operationName))
	operation, err = operationsK8sClient.Get(ctx, operationName, metav1.GetOptions{})
	require.NoError(t, err)
	counter = 0
	for !isOperationDeletionCompleted(operation) && counter < 100 {
		t.Log(fmt.Sprintf("[%s] Operation deletion is not completed yet. The operation state is: %s", time.Now().Format("06.01.02 15:04:05"), operation.Status.Phase))
		time.Sleep(time.Second * 5)
		operation, err = operationsK8sClient.Get(ctx, operationName, metav1.GetOptions{})
		require.NoError(t, err)
		counter++
	}
	t.Log(fmt.Sprintf("The operation state is: %s", operation.Status.Phase))

	t.Log("Verify the deleted application do not exists in director")
	application = getApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)
	require.Empty(t, application.Name, "Application is not deleted")
}

func webhookModePtr(mode graphql.WebhookMode) *graphql.WebhookMode {
	return &mode
}

func unlockWebhook(t *testing.T, operationFullPath string) {
	httpClient := http.Client{}
	okRequestData := webhook.OKRequestData{
		OK: true,
	}
	jsonValueoOKRequestData, err := json.Marshal(okRequestData)
	require.NoError(t, err)
	reqPost, err := http.NewRequest(http.MethodPost, operationFullPath, bytes.NewBuffer(jsonValueoOKRequestData))
	require.NoError(t, err)
	respPost, err := httpClient.Do(reqPost)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, respPost.StatusCode)
}

func isWebhookOperationInDesiredState(t *testing.T, operationFullPath, desiredState string) (isInState bool, response string) {
	httpClient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, operationFullPath, nil)
	require.NoError(t, err)
	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	fullBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	fullBodyString := string(fullBody)
	return strings.Contains(fullBodyString, desiredState), fullBodyString
}

func isOperationDeletionCompleted(operation *v1alpha1.Operation) bool {
	if operation.Status.Phase == v1alpha1.StateSuccess || operation.Status.Phase == v1alpha1.StateFailed {
		return true
	}
	return false
}

func deleteApplicationOnExit(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string, tenant string) {
	application := getApplication(t, ctx, gqlClient, testConfig.DefaultTenant, id)
	if application.Name != "" {
		unregisterApplicationInTenant(t, ctx, gqlClient, id, tenant)
	}
}
