package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"

	"github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
	"github.com/kyma-incubator/compass/components/operations-controller/client"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testPkg "github.com/kyma-incubator/compass/tests/pkg/webhook"
	gcli "github.com/machinebox/graphql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/webhook"
)

func TestAsyncAPIUnpairApplicationWithAppWebhook(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("TestAsyncAPIUnpairApplicationWithAppWebhook", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		appName := fmt.Sprintf("app-async-unpair-%s", time.Now().Format("060102150405"))
		appInput := graphql.ApplicationRegisterInput{
			Name:         appName,
			ProviderName: ptr.String("compass"),
			Webhooks:     []*graphql.WebhookInput{testPkg.BuildMockedWebhook(testConfig.ExternalServicesMockBaseURL, graphql.WebhookTypeUnregisterApplication)},
		}

		t.Log(fmt.Sprintf("Registering application: %s", appName))
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(appInput)
		require.NoError(t, err)

		registerRequest := fixtures.FixRegisterApplicationRequest(appInputGQL)
		app := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, registerRequest, &app)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &app)
		require.NoError(t, err)

		require.Equal(t, app.Status.Condition, graphql.ApplicationStatusConditionInitial)
		require.Len(t, app.Webhooks, 1)
		nearCreationTime := time.Now().Add(-1 * time.Second)

		triggerAsyncUnpair(t, ctx, app, nearCreationTime, app.Webhooks[0].ID, certSecuredGraphQLClient)
	})
}

func triggerAsyncUnpair(t *testing.T, ctx context.Context, app graphql.ApplicationExt, appNearCreationTime time.Time, expectedWebhookID string, gqlClient *gcli.Client) {
	operationFullPath := testPkg.BuildOperationFullPath(testConfig.ExternalServicesMockBaseURL)

	t.Log("Unlock the mock application webhook")
	testPkg.UnlockWebhook(t, operationFullPath)
	require.True(t, isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusOK), fmt.Sprintf("Expected state: %s", webhook.OperationResponseStatusOK))

	t.Log("Start async Unpair of application")
	fixtures.UnpairAsyncApplicationInTenant(t, ctx, gqlClient, testConfig.DefaultTestTenant, app.ID)

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

	if len(expectedWebhookID) == 0 {
		require.Len(t, operation.Spec.WebhookIDs, 0)
	} else {
		require.Len(t, operation.Spec.WebhookIDs, 1)
		require.Equal(t, expectedWebhookID, operation.Spec.WebhookIDs[0])

		t.Log(fmt.Sprintf("Verify operation CR with name %s is in progress", operationName))
		require.Eventually(t, func() bool {
			return isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusINProgress)
		}, time.Minute*3, time.Second*5, "Waiting for state change timed out.")

		t.Log("Verify the application status in director is 'ready:false'")
		unpairedApp := fixtures.GetApplication(t, ctx, gqlClient, testConfig.DefaultTestTenant, app.ID)
		require.NoError(t, err)
		require.Equal(t, unpairedApp.Status.Condition, graphql.ApplicationStatusConditionUnpairing)
		require.Empty(t, unpairedApp.Error, "Application Error is not empty")

		t.Log("Verify UpdatedAt in director is set and is in expected range")
		require.NotEmpty(t, unpairedApp.UpdatedAt, "Application Update time is not empty")
		updatedAt := time.Time(*unpairedApp.UpdatedAt)
		require.True(t, appNearCreationTime.Before(updatedAt), "Updated time is before creation time")

		t.Log("Unlock application webhook")
		testPkg.UnlockWebhook(t, operationFullPath)
		require.True(t, isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusOK), fmt.Sprintf("Expected state: %s", webhook.OperationResponseStatusOK))
	}

	t.Log(fmt.Sprintf("Verify operation CR with name %s status condition is ConditionTypeReady", operationName))
	require.Eventually(t, func() bool {
		operation, err = operationsK8sClient.Get(ctx, operationName, metav1.GetOptions{})
		require.NoError(t, err)
		t.Log(fmt.Sprintf("The operation state is: %s", operation.Status.Phase))
		return isOperationUnpairCompleted(operation)
	}, time.Minute*3, time.Second*10, "Waiting for operation unpair timed out.")

	t.Log("Verify the unpaired application still exists in director")
	existingApp := fixtures.GetApplication(t, ctx, gqlClient, testConfig.DefaultTestTenant, app.ID)
	require.NotEmpty(t, existingApp.Name, "Application is deleted")
}

func isOperationUnpairCompleted(operation *v1alpha1.Operation) bool {
	if operation.Status.Phase == v1alpha1.StateSuccess {
		return true
	}
	return false
}
