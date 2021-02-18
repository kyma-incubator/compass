package external_services_mock_integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/director/pkg/gql"
	"github.com/kyma-incubator/compass/tests/director/pkg/idtokenprovider"

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
	// Precondition - Register an application
	ctx := context.Background()
	appName := "app-for-testing-async-deletion-mock"
	appInput := graphql.ApplicationRegisterInput{
		Name:         appName,
		ProviderName: ptr.String("compass"),
		Webhooks: []*graphql.WebhookInput{
			{
				Type: graphql.WebhookTypeUnregisterApplication,
				Mode: webhookModePtr(graphql.WebhookModeAsync),
				URL:  str.Ptr(deleteFullPath),
				OutputTemplate: str.Ptr(fmt.Sprintf(`{
					"location": "%s",
					"success_status_code": 200,
					"error": "{{.Body.error}}"
				}`, operationFullPath)),
				StatusTemplate: str.Ptr(`{
					"status": "{{.Body.status}}",
					"success_status_code": 200,
					"success_status_identifier": "SUCCESS",
					"in_progress_status_identifier": "IN_PROGRESS",
					"failed_status_identifier": "FAILED",
					"error": "{{.Body.error}}"
				 }`),
			},
		},
	}

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	t.Log("Create request for registering application")
	appInputGQL, err := tc.Graphqlizer.ApplicationRegisterInputToGQL(appInput)
	require.NoError(t, err)

	registerRequest := fixRegisterApplicationRequest(appInputGQL)

	t.Log("Register application through Gateway with Dex id token")
	app := graphql.ApplicationExt{}
	err = tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, testConfig.DefaultTenant, registerRequest, &app)
	require.NoError(t, err)

	defer unregisterApplicationInTenant(t, ctx, dexGraphQLClient, app.ID, testConfig.DefaultTenant)

	t.Log("Unlock the mock application webhook")
	unlockWebhook(t, operationFullPath)

	t.Log("Verify the webhook is unlocked and operation is succedded")
	checkWebhookOperationState(t, operationFullPath, webhook.OperationResponseStatusOK)

	// Async Delete start
	t.Log("Start async Delete application")
	unregisterAsyncApplicationInTenant(t, ctx, dexGraphQLClient, app.ID, testConfig.DefaultTenant)

	// - check for operation CRD - (should exists) - retry 5 sec
	t.Log("Check check for operation CRD should exists: TODO")
	time.Sleep(time.Second * 5)

	t.Log("Verify the webhook is locked and operation is in progress")
	checkWebhookOperationState(t, operationFullPath, webhook.OperationResponseStatusINProgress)

	t.Log("Check the deleted application status in director")
	application := getApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)
	require.False(t, application.Ready, "Application is not in status ready:false")

	t.Log("Unlock application webhook")
	unlockWebhook(t, operationFullPath)

	// - check for operation - should not be available - retry 5 sec
	t.Log("Check check for operation CRD should not exists: TODO")
	time.Sleep(time.Second * 5)

	t.Log("Check in director that the deleted application should not exists")
	application = getApplication(t, ctx, dexGraphQLClient, testConfig.DefaultTenant, app.ID)
	require.Nil(t, application.BaseEntity, "Application is not deleted")
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

func checkWebhookOperationState(t *testing.T, operationFullPath, desiredState string) {
	httpClient := http.Client{}
	req, err := http.NewRequest(http.MethodGet, operationFullPath, nil)
	require.NoError(t, err)
	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	fullBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Contains(t, string(fullBody), desiredState)
}
