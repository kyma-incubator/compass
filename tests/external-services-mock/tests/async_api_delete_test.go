package tests

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

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

func TestAsyncAPIDeleteApplicationWithAppWebhook(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("TestAsyncAPIDeleteApplicationWithAppWebhook", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		appName := fmt.Sprintf("app-async-del-%s", time.Now().Format("060102150405"))
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

		triggerAsyncDeletion(t, ctx, app, nearCreationTime, app.Webhooks[0].ID, certSecuredGraphQLClient)
	})
}

func TestAsyncAPIDeleteApplicationWithMTLSAppWebhook(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("TestAsyncAPIDeleteApplicationWithMTLSAppWebhook", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		appName := fmt.Sprintf("app-async-del-%s", time.Now().Format("060102150405"))
		mtlsWebhook := testPkg.BuildMockedWebhook(testConfig.ExternalServicesMockMTLSSecuredURL, graphql.WebhookTypeUnregisterApplication)
		mtlsWebhook.Auth = &graphql.AuthInput{
			AccessStrategy: str.Ptr(string(accessstrategy.CMPmTLSAccessStrategy)),
		}
		appInput := graphql.ApplicationRegisterInput{
			Name:         appName,
			ProviderName: ptr.String("compass"),
			Webhooks:     []*graphql.WebhookInput{mtlsWebhook},
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

		triggerAsyncDeletion(t, ctx, app, nearCreationTime, app.Webhooks[0].ID, certSecuredGraphQLClient)
	})
}

func TestAsyncAPIDeleteApplicationWithAppTemplateWebhook(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("TestAsyncAPIDeleteApplicationWithAppTemplateWebhook", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		appName := fmt.Sprintf("app-async-del-%s", time.Now().Format("060102150405"))
		appTemplateName := fmt.Sprintf("test-app-tmpl-%s", time.Now().Format("060102150405"))
		appTemplateName = fmt.Sprintf("SAP %s", appTemplateName)
		appTemplateInput := graphql.ApplicationTemplateInput{
			Name: appTemplateName,
			ApplicationInput: &graphql.ApplicationRegisterInput{
				Name:        "{{name}}",
				Description: ptr.String("test {{display-name}}"),
			},
			Placeholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &appName,
				},
				{
					Name:        "display-name",
					Description: ptr.String("display-name"),
				},
			},
			Labels: graphql.Labels{
				testConfig.AppSelfRegDistinguishLabelKey: []interface{}{testConfig.AppSelfRegDistinguishLabelValue},
				tenantfetcher.RegionKey:                  testConfig.AppSelfRegRegion,
			},
			AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
			Webhooks:    []*graphql.WebhookInput{testPkg.BuildMockedWebhook(testConfig.ExternalServicesMockBaseURL, graphql.WebhookTypeUnregisterApplication)},
		}

		t.Log(fmt.Sprintf("Registering application template: %s", appTemplateName))
		appTemplateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
		require.NoError(t, err)

		registerTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplateInputGQL)
		appTemplate := graphql.ApplicationTemplate{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, registerTemplateRequest, &appTemplate)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &appTemplate)
		require.NoError(t, err)

		require.Len(t, appTemplate.Webhooks, 1)

		t.Log(fmt.Sprintf("Registering application from template: %s", appTemplateName))
		appFromAppTemplateInput := graphql.ApplicationFromTemplateInput{
			TemplateName: appTemplateName,
			Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       appName,
				},
				{
					Placeholder: "display-name",
					Value:       "display-name",
				},
			},
		}
		appFromTemplateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromAppTemplateInput)
		require.NoError(t, err)

		registerAppRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTemplateInputGQL)
		app := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, registerAppRequest, &app)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &appTemplate)
		require.NoError(t, err)

		require.Equal(t, app.Status.Condition, graphql.ApplicationStatusConditionInitial)
		nearCreationTime := time.Now().Add(-1 * time.Second)

		triggerAsyncDeletion(t, ctx, app, nearCreationTime, appTemplate.Webhooks[0].ID, certSecuredGraphQLClient)
	})
}

func TestAsyncAPIDeleteApplicationPrioritizationWithBothAppTemplateAndAppWebhook(stdT *testing.T) {
	t := testingx.NewT(stdT)
	t.Run("TestAsyncAPIDeleteApplicationPrioritizationWithBothAppTemplateAndAppWebhook", func(t *testing.T) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		appName := fmt.Sprintf("app-async-del-%s", time.Now().Format("060102150405"))
		appTemplateName := fmt.Sprintf("test-app-tmpl-%s", time.Now().Format("060102150405"))
		appTemplateName = fmt.Sprintf("SAP %s", appTemplateName)
		appTemplateInput := graphql.ApplicationTemplateInput{
			Name: appTemplateName,
			ApplicationInput: &graphql.ApplicationRegisterInput{
				Name:        "{{name}}",
				Description: ptr.String("test {{display-name}}"),
			},
			Placeholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &appName,
				},
				{
					Name:        "display-name",
					Description: ptr.String("display-name"),
				},
			},
			Labels: graphql.Labels{
				testConfig.AppSelfRegDistinguishLabelKey: []interface{}{testConfig.AppSelfRegDistinguishLabelValue},
				tenantfetcher.RegionKey:                  testConfig.AppSelfRegRegion,
			},
			AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal,
			Webhooks:    []*graphql.WebhookInput{testPkg.BuildMockedWebhook(testConfig.ExternalServicesMockBaseURL, graphql.WebhookTypeUnregisterApplication)},
		}

		t.Log(fmt.Sprintf("Registering application template: %s", appTemplateName))
		appTemplateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
		require.NoError(t, err)

		registerTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplateInputGQL)
		appTemplate := graphql.ApplicationTemplate{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, registerTemplateRequest, &appTemplate)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &appTemplate)
		require.NoError(t, err)

		require.Len(t, appTemplate.Webhooks, 1)

		t.Log(fmt.Sprintf("Registering application from template: %s", appName))
		appFromAppTemplateInput := graphql.ApplicationFromTemplateInput{
			TemplateName: appTemplateName,
			Values: []*graphql.TemplateValueInput{
				{
					Placeholder: "name",
					Value:       appName,
				},
				{
					Placeholder: "display-name",
					Value:       "display-name",
				},
			},
		}
		appFromTemplateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromAppTemplateInput)
		require.NoError(t, err)

		registerAppRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTemplateInputGQL)
		app := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, registerAppRequest, &app)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, &app)
		require.NoError(t, err)

		require.Equal(t, app.Status.Condition, graphql.ApplicationStatusConditionInitial)
		require.Len(t, app.Webhooks, 1)
		nearCreationTime := time.Now().Add(-1 * time.Second)

		t.Log(fmt.Sprintf("Registering webhook for application: %s", appName))
		appWebhookInputGQL, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(testPkg.BuildMockedWebhook(testConfig.ExternalServicesMockBaseURL, graphql.WebhookTypeUnregisterApplication))
		require.NoError(t, err)

		registerAppWebhookRequest := fixtures.FixAddWebhookToApplicationRequest(app.ID, appWebhookInputGQL)
		webhookResult := graphql.Webhook{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testConfig.DefaultTestTenant, registerAppWebhookRequest, &webhookResult)
		require.NoError(t, err)

		triggerAsyncDeletion(t, ctx, app, nearCreationTime, webhookResult.ID, certSecuredGraphQLClient)
	})
}

func triggerAsyncDeletion(t *testing.T, ctx context.Context, app graphql.ApplicationExt, appNearCreationTime time.Time, expectedWebhookID string, gqlClient *gcli.Client) {
	operationFullPath := testPkg.BuildOperationFullPath(testConfig.ExternalServicesMockBaseURL)

	t.Log("Unlock the mock application webhook")
	testPkg.UnlockWebhook(t, operationFullPath)
	require.True(t, isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusOK), fmt.Sprintf("Expected state: %s", webhook.OperationResponseStatusOK))

	t.Log("Start async Delete of application")
	fixtures.UnregisterAsyncApplicationInTenant(t, ctx, gqlClient, testConfig.DefaultTestTenant, app.ID)

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
		deletedApp := fixtures.GetApplication(t, ctx, gqlClient, testConfig.DefaultTestTenant, app.ID)
		require.NoError(t, err)
		require.Equal(t, deletedApp.Status.Condition, graphql.ApplicationStatusConditionDeleting)
		require.Empty(t, deletedApp.Error, "Application Error is not empty")

		t.Log("Verify DeletedAt in director is set and is in expected range")
		require.NotEmpty(t, deletedApp.DeletedAt, "Application Deletion time is not empty")
		deletedAtTime := time.Time(*deletedApp.DeletedAt)
		require.True(t, appNearCreationTime.Before(deletedAtTime), "Deleted time is before creation time")
		require.True(t, time.Now().After(deletedAtTime), "Deleted time is in the future")

		t.Log("Unlock application webhook")
		testPkg.UnlockWebhook(t, operationFullPath)
		require.True(t, isWebhookOperationInDesiredState(t, operationFullPath, webhook.OperationResponseStatusOK), fmt.Sprintf("Expected state: %s", webhook.OperationResponseStatusOK))

	}

	t.Log(fmt.Sprintf("Verify operation CR with name %s status condition is ConditionTypeReady", operationName))
	require.Eventually(t, func() bool {
		operation, err = operationsK8sClient.Get(ctx, operationName, metav1.GetOptions{})
		require.NoError(t, err)
		t.Log(fmt.Sprintf("The operation state is: %s", operation.Status.Phase))
		return isOperationDeletionCompleted(operation)
	}, time.Minute*3, time.Second*10, "Waiting for operation deletion timed out.")

	t.Log("Verify the deleted application do not exists in director")
	missingApp := fixtures.GetApplication(t, ctx, gqlClient, testConfig.DefaultTestTenant, app.ID)
	require.Empty(t, missingApp.Name, "Application is not deleted")
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
