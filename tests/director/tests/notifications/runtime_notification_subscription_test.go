package notifications

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestFormationNotificationsWithRuntimeAndApplicationParticipants(stdT *testing.T) {
	ctx := context.Background()
	t := testingx.NewT(stdT)

	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	applicationType1 := "subscription-notifications-app-type-1"
	applicationType2 := "subscription-notifications-app-type-2"
	formationTemplateName := "subscription-notifications-template-name"
	formationName := "subscription-notifications-formation-name"

	subscriptionSubdomain := conf.SelfRegisterSubdomainPlaceholderValue
	subscriptionConsumerAccountID := conf.TestConsumerAccountID
	subscriptionProviderSubaccountID := conf.TestProviderSubaccountID // in local set up the parent is testDefaultTenant
	subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID // in local set up the parent is ApplicationsForRuntimeTenantName
	subscriptionConsumerTenantID := conf.TestConsumerTenantID

	t.Run("Formation Notifications With Subscriptions", func(t *testing.T) {
		// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
		providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, false)
		directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

		defer func() {
			k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
			require.NoError(t, err)
			k8s.DeleteSecret(t, ctx, k8sClient, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace)
		}()

		// Register provider runtime
		providerRuntimeInput := graphql.RuntimeRegisterInput{
			Name:        "subscription-notifications-provider-runtime",
			Description: ptr.String("subscription-notifications-provider-runtime-description"),
			Labels: graphql.Labels{
				conf.SubscriptionConfig.SelfRegDistinguishLabelKey: conf.SubscriptionConfig.SelfRegDistinguishLabelValue,
			},
			ApplicationNamespace: ptr.String("e2e.namespace"),
		}

		var providerRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntimeWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntime)
		providerRuntime = fixtures.RegisterRuntimeFromInputWithoutTenant(t, ctx, directorCertSecuredClient, &providerRuntimeInput)
		require.NotEmpty(t, providerRuntime.ID)

		selfRegLabelValue, ok := providerRuntime.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(t, ok)
		require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+providerRuntime.ID)

		saasAppLbl, ok := providerRuntime.Labels[conf.SaaSAppNameLabelKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, saasAppLbl)

		regionLbl, ok := providerRuntime.Labels[tenantfetcher.RegionKey].(string)
		require.True(t, ok)
		require.NotEmpty(t, regionLbl)

		httpClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
			},
		}

		deps, err := json.Marshal([]string{selfRegLabelValue})
		require.NoError(t, err)
		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer(deps))
		require.NoError(t, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				t.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)

		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, claims.TenantFetcherClaimKey)
		apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)
		defer subscription.BuildAndExecuteUnsubscribeRequest(t, providerRuntime.ID, providerRuntime.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)
		subscription.CreateRuntimeSubscription(t, conf.SubscriptionConfig, httpClient, providerRuntime, subscriptionToken, apiPath, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID, conf.SubscriptionProviderAppNameValue, true, conf.SubscriptionConfig.StandardFlow)

		t.Log("Assert provider runtime is visible in the consumer's subaccount after successful subscription")
		consumerSubaccountRuntime := fixtures.GetRuntime(t, ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, providerRuntime.ID)
		require.Equal(t, providerRuntime.ID, consumerSubaccountRuntime.ID)

		t.Log("Assert there is a runtime context(subscription) as part of the provider runtime")
		require.Len(t, consumerSubaccountRuntime.RuntimeContexts.Data, 1)
		require.NotEmpty(t, consumerSubaccountRuntime.RuntimeContexts.Data[0].ID)
		require.Equal(t, conf.SubscriptionLabelKey, consumerSubaccountRuntime.RuntimeContexts.Data[0].Key)
		require.Equal(t, subscriptionConsumerTenantID, consumerSubaccountRuntime.RuntimeContexts.Data[0].Value)
		rtCtx := consumerSubaccountRuntime.RuntimeContexts.Data[0]

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-template-test")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		namePlaceholder := "name"
		displayNamePlaceholder := "display-name"
		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		t.Logf("Create application template for type %q", applicationType1)
		appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType1, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		localTenantID2 := "local-tenant-id2"
		t.Logf("Create application template for type %q", applicationType2)
		appTemplateInput = fixtures.FixApplicationTemplateWithoutWebhook(applicationType2, localTenantID2, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
		appTmpl, err = fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
		require.NoError(t, err)

		appFromTmplSrc := fixtures.FixApplicationFromTemplateInput(applicationType1, namePlaceholder, "app1-formation-notifications-tests", displayNamePlaceholder, "App 1 Display Name")
		t.Logf("Create application 1 from template %q", applicationType1)
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		app1 := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplFirstRequest, &app1)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &app1)
		require.NoError(t, err)
		require.NotEmpty(t, app1.ID)
		t.Logf("app1 ID: %q", app1.ID)
		appFromTmplSrc2 := fixtures.FixApplicationFromTemplateInput(applicationType2, namePlaceholder, "app2-formation-notifications-tests", displayNamePlaceholder, "App 2 Display Name")

		t.Logf("Create application 2 from template %q", applicationType2)
		appFromTmplSrc2GQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc2)
		require.NoError(t, err)
		createAppFromTmplSecondRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrc2GQL)
		app2 := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplSecondRequest, &app2)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &app2)
		require.NoError(t, err)
		require.NotEmpty(t, app2.ID)
		t.Logf("app2 ID: %q", app2.ID)

		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, formationTemplateName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(stdT, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(stdT, ctx, certSecuredGraphQLClient, formationTemplateName, conf.SubscriptionProviderAppNameValue, []string{applicationType1, applicationType2}, graphql.ArtifactTypeSubscription)

		t.Run("Formation Assignment Notifications For Runtime With Synchronous Webhook", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			webhookType := graphql.WebhookTypeConfigurationChanged
			webhookMode := graphql.WebhookModeSync
			urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"ucl-formation-name\\\":\\\"{{.Formation.Name}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
			outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			runtimeWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

			t.Logf("Add webhook with type %q and mode: %q to provider runtime with ID %q", webhookType, webhookMode, providerRuntime.ID)
			actualWebhook := fixtures.AddWebhookToRuntime(t, ctx, directorCertSecuredClient, runtimeWebhookInput, "", providerRuntime.ID)
			defer fixtures.CleanupWebhook(t, ctx, directorCertSecuredClient, "", actualWebhook.ID)

			t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
			defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
			formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &formationTemplateName)
			require.NotEmpty(t, formation.ID)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)

			t.Logf("Assign application 1 to formation %s", formationName)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
			assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var assignedFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
			assignReq = fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedConfig := str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
				},
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)

			notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]
			assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			t.Logf("Assign application 2 to formation %s", formationName)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
			assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
					app2.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				app2.ID: {
					app2.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 9, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
			validateJSONStringProperty(t, notificationsForConsumerTenant.Array()[0], "RequestBody.ucl-formation-name", formationName)
			validateJSONStringProperty(t, assignNotificationForApp1, "RequestBody.ucl-formation-name", formationName)

			notificationForApp2Found := false
			for _, notification := range notificationsForConsumerTenant.Array() {
				appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
				t.Logf("Found notification for app %q", appIDFromNotification)
				if appIDFromNotification == app2.ID {
					notificationForApp2Found = true
					assertFormationAssignmentsNotificationWithItemsStructure(t, notification, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
				}
			}
			require.True(t, notificationForApp2Found, "notification for assign app2 not found")

			t.Logf("Unassign Application 1 from formation %s", formationName)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var unassignFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app2.ID: {
					app2.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
				},
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)

			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
			unassignNotificationFound := false
			for _, notification := range notificationsForConsumerTenant.Array() {
				op := notification.Get("Operation").String()
				if op == unassignOperation {
					unassignNotificationFound = true
					assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
					validateJSONStringProperty(t, notification, "RequestBody.ucl-formation-name", formationName)
				}
			}
			require.True(t, unassignNotificationFound, "notification for unassign app1 not found")

			t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, formationName)
			unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 4)

			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
			unassignNotificationForApp2Found := false
			for _, notification := range notificationsForConsumerTenant.Array() {
				op := notification.Get("Operation").String()
				appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
				t.Logf("Found %q notification for app %q", op, appIDFromNotification)
				if appIDFromNotification == app2.ID && op == unassignOperation {
					unassignNotificationForApp2Found = true
					assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
				}
			}
			require.True(t, unassignNotificationForApp2Found, "notification for unassign app2 not found")

			t.Logf("Unassign Application 2 from formation %s", formationName)
			unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
		})

		t.Run("Runtime Context to Application formation assignment notifications", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
			defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
			formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &formationTemplateName)

			webhookType := graphql.WebhookTypeConfigurationChanged
			webhookMode := graphql.WebhookModeSync
			urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.Application.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.RuntimeContext.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{.Runtime.Labels.region }}\\\",\\\"application-namespace\\\":\\\"{{.Runtime.ApplicationNamespace}}\\\",\\\"application-tenant-id\\\":\\\"{{.RuntimeContext.Value}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.RuntimeContext.ID}}\\\"}]}"
			outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, app1.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, subscriptionConsumerAccountID, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationWebhook.ID)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)

			t.Logf("Assign application to formation %s", formationName)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
			assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var assignedFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
			assignReq = fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedConfig := str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
				},
			}

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)

			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, localTenantID, 1)

			notificationsForConsumerTenant := gjson.GetBytes(body, localTenantID)
			assignNotificationForApp := notificationsForConsumerTenant.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp, assignOperation, formation.ID, rtCtx.ID, rtCtx.Value, regionLbl, "", subscriptionConsumerAccountID, emptyParentCustomerID)
			assert.NoError(t, err)

			t.Logf("Unassign Application from formation %s", formationName)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var unassignFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, localTenantID, 2)

			notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
			assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, unassignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 1)

			t.Logf("Assign application to formation %s", formationName)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
			assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var secondAssignedFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &secondAssignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
				},
			}

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, localTenantID, 3)

			notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
			assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, assignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 2)

			t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, formationName)
			unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, localTenantID, 4)

			notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
			assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, unassignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 2)

			t.Logf("Unassign Application from formation %s", formationName)
			unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
		})

		t.Run("Formation Assignment Notifications for Runtime with AsyncCallback Webhook and application with Synchronous Webhook", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			webhookTypeRuntime := graphql.WebhookTypeConfigurationChanged
			webhookModeRuntime := graphql.WebhookModeAsyncCallback
			urlTemplateRuntime := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateRuntime := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\",\\\"subdomain\\\":\\\"{{ if eq .RuntimeContext.Tenant.Type \\\"subaccount\\\" }}{{ .RuntimeContext.Tenant.Labels.subdomain }}{{end}}\\\"}]}"
			outputTemplateRuntime := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

			runtimeWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookTypeRuntime, webhookModeRuntime, urlTemplateRuntime, inputTemplateRuntime, outputTemplateRuntime)

			t.Logf("Add webhook with type %q and mode: %q to provider runtime with ID %q", webhookTypeRuntime, webhookModeRuntime, providerRuntime.ID)
			actualRuntimeWebhook := fixtures.AddWebhookToRuntime(t, ctx, directorCertSecuredClient, runtimeWebhookInput, "", providerRuntime.ID)
			defer fixtures.CleanupWebhook(t, ctx, directorCertSecuredClient, "", actualRuntimeWebhook.ID)

			webhookTypeApplication := graphql.WebhookTypeConfigurationChanged
			webhookModeApplication := graphql.WebhookModeSync
			urlTemplateApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/configuration/{{.Application.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.RuntimeContext.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{.Runtime.Labels.region }}\\\",\\\"application-namespace\\\":\\\"\\\",\\\"application-tenant-id\\\":\\\"{{.RuntimeContext.Value}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.RuntimeContext.ID}}\\\",\\\"subdomain\\\":\\\"{{ if eq .RuntimeContext.Tenant.Type \\\"subaccount\\\" }}{{ .RuntimeContext.Tenant.Labels.subdomain }}{{end}}\\\"}]}"
			outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookTypeApplication, webhookModeApplication, urlTemplateApplication, inputTemplateApplication, outputTemplateApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookTypeApplication, webhookModeApplication, app1.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, subscriptionConsumerAccountID, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationWebhook.ID)

			// Create formation constraints for destination creator operator and attach them to a given formation template.
			// So we can verify the destination creator will not fail if in the configuration there is no destination information
			attachDestinationCreatorConstraints(t, ctx, ft, graphql.ResourceTypeRuntimeContext, graphql.ResourceTypeApplication)

			t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
			defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
			formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &formationTemplateName)
			require.NotEmpty(t, formation.ID)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			syncConfig := str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
			asyncConfig := str.Ptr("{\"asyncKey\":\"asyncValue\",\"asyncKey2\":{\"asyncNestedKey\":\"asyncNestedValue\"}}")
			expectedAssignmentsBySourceID := map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: syncConfig, Value: syncConfig, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: asyncConfig, Value: asyncConfig, Error: nil},
				},
			}

			t.Run("Normal case notifications are sent and formation assignments are correct", func(t *testing.T) {
				var assignedFormation graphql.Formation

				t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
				assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
				defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				t.Logf("Assign application to formation %s", formation.Name)
				defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
				assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

				body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

				// rtCtx <- App notifications
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)
				notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
				assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]
				assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

				// rtCtx -> App notifications
				assertNotificationsCountForTenant(t, body, localTenantID, 2)
				notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
				assertExpectationsForApplicationNotificationsWithItemsStructure(t, notificationsForConsumerTenant.Array(), []*applicationFormationExpectations{
					{
						op:            assignOperation,
						formationID:   formation.ID,
						objectID:      rtCtx.ID,
						localTenantID: rtCtx.Value,
						objectRegion:  regionLbl,
						configuration: "",
						tenant:        subscriptionConsumerAccountID,
						customerID:    emptyParentCustomerID,
					},
					{
						op:            assignOperation,
						formationID:   formation.ID,
						objectID:      rtCtx.ID,
						localTenantID: rtCtx.Value,
						objectRegion:  regionLbl,
						configuration: "{\"asyncKey\":\"asyncValue\",\"asyncKey2\":{\"asyncNestedKey\":\"asyncNestedValue\"}}",
						tenant:        subscriptionConsumerAccountID,
						customerID:    emptyParentCustomerID,
					},
				})
				assertFormationAssignmentsNotificationSubdomainWithItemsStructure(t, notificationsForConsumerTenant.Array()[0], subscriptionSubdomain)
				assertFormationAssignmentsNotificationSubdomainWithItemsStructure(t, notificationsForConsumerTenant.Array()[1], subscriptionSubdomain)

				var unassignFormation graphql.Formation
				t.Logf("Unassign application from formation %s", formation.Name)
				unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
				require.NoError(t, err)
				require.Equal(t, formation.Name, assignedFormation.Name)

				application := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
				scenarios, hasScenarios := application.Labels["scenarios"]
				assert.True(t, hasScenarios)
				assert.Len(t, scenarios, 1)
				assert.Contains(t, scenarios, formationName)

				body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

				// rtCtx <- App notifications
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)
				notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

				unassignNotificationFound := false
				for _, notification := range notificationsForConsumerTenant.Array() {
					op := notification.Get("Operation").String()
					if op == unassignOperation {
						appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
						t.Logf("Found notification for app %q", appIDFromNotification)
						if appIDFromNotification == app1.ID {
							unassignNotificationFound = true
							assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t, notification, unassignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID, nil)
							assertFormationAssignmentsNotificationSubdomainWithItemsStructure(t, notification, subscriptionSubdomain)
						}
					}
				}
				require.True(t, unassignNotificationFound)

				// rtCtx -> App notifications
				assertNotificationsCountForTenant(t, body, localTenantID, 3)
				notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
				assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, unassignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 1)

				expectedAssignments := map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

				t.Logf("Check that application with ID %q is unassigned from formation %s", app1.ID, formationName)
				app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
				scenarios, hasScenarios = app.Labels["scenarios"]
				assert.False(t, hasScenarios)

				t.Logf("Check that runtime context with ID %q is still assigned to formation %s", subscriptionConsumerSubaccountID, formationName)
				actualRtmCtx := fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
				scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
				assert.True(t, hasScenarios)
				assert.Len(t, scenarios, 1)
				assert.Contains(t, scenarios, formationName)

				t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, formationName)
				unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, unassignFormation.Name)

				t.Logf("Check that runtime context with ID %q is actually unassigned from formation %s", subscriptionConsumerSubaccountID, formationName)
				actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
				scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
				assert.False(t, hasScenarios)

			})

			t.Run("Consecutive participants unassignment are still in formation before the formation assignments are processed by the async API call and removed afterwards", func(t *testing.T) {
				var assignedFormation graphql.Formation

				t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
				assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
				defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				t.Logf("Assign application with ID: %s to formation %s", app1.ID, formation.Name)
				defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
				assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID, 300)

				t.Logf("Check that the runtime context with ID: %s is assigned to formation: %s", rtCtx.ID, formationName)
				actualRtmCtx := fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
				scenarios, hasScenarios := actualRtmCtx.Labels["scenarios"]
				assert.True(t, hasScenarios)
				assert.Len(t, scenarios, 1)
				assert.Contains(t, scenarios, formationName)

				t.Logf("Check that the application with ID: %q is assigned to formation: %s", app1.ID, formationName)
				app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
				scenarios, hasScenarios = app.Labels["scenarios"]
				assert.True(t, hasScenarios)
				assert.Len(t, scenarios, 1)
				assert.Contains(t, scenarios, formationName)

				var unassignFormation graphql.Formation

				t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, formationName)
				unassignReq := fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, unassignFormation.Name)

				t.Logf("Unassign application with ID: %s from formation %s", app1.ID, formation.Name)
				unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
				if err != nil {
					require.Contains(t, err.Error(), "Object not found")
				}
				require.Equal(t, formation.Name, assignedFormation.Name)

				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil, 300)

				actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
				scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
				assert.False(t, hasScenarios)

				app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
				scenarios, hasScenarios = app.Labels["scenarios"]
				assert.False(t, hasScenarios)
			})

			t.Run("Application is not unassigned when only tenant is unassigned", func(t *testing.T) {
				var assignedFormation graphql.Formation

				t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
				assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
				defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				t.Logf("Assign application to formation %s", formation.Name)
				defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
				assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID, 300)

				t.Logf("Check that runtime context with ID %q is assigned from formation %s", subscriptionConsumerAccountID, formationName)
				actualRtmCtx := fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
				scenarios, hasScenarios := actualRtmCtx.Labels["scenarios"]
				assert.True(t, hasScenarios)
				assert.Len(t, scenarios, 1)
				assert.Contains(t, scenarios, formationName)

				t.Logf("Check that application with ID %q is assigned from formation %s", app1.ID, formationName)
				app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
				scenarios, hasScenarios = app.Labels["scenarios"]
				assert.True(t, hasScenarios)
				assert.Len(t, scenarios, 1)
				assert.Contains(t, scenarios, formationName)

				var unassignFormation graphql.Formation

				t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, formationName)
				unassignReq := fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, unassignFormation.Name)

				expectedAssignments := map[string]map[string]fixtures.AssignmentState{
					app1.ID: {
						app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments, 300)

				actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
				scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
				assert.False(t, hasScenarios)

				t.Logf("Check that application with ID %q is still assigned to formation %s", app1.ID, formationName)
				app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
				scenarios, hasScenarios = app.Labels["scenarios"]
				assert.True(t, hasScenarios)
				assert.Len(t, scenarios, 1)
				assert.Contains(t, scenarios, formationName)
			})
		})

		t.Run("Fail Processing formation assignments while assigning from formation", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
			defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

			webhookMode := graphql.WebhookModeSync
			webhookType := graphql.WebhookTypeConfigurationChanged
			urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{if eq .Operation \\\"assign\\\"}}fail-once/{{end}}{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
			outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			runtimeWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

			t.Logf("Add webhook with type %q and mode: %q to provider runtime with ID %q", webhookType, webhookMode, providerRuntime.ID)
			actualWebhook := fixtures.AddWebhookToRuntime(t, ctx, directorCertSecuredClient, runtimeWebhookInput, "", providerRuntime.ID)
			defer fixtures.CleanupWebhook(t, ctx, directorCertSecuredClient, "", actualWebhook.ID)

			t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
			defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
			formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &formationTemplateName)
			require.NotEmpty(t, formation.ID)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			var assignedFormation graphql.Formation

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
			assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			// notification mock API should return error
			t.Logf("Assign application to formation %s should fail", formation.Name)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
			assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedError := str.Ptr("{\"error\":{\"message\":\"failed to parse request\",\"errorCode\":2}}")
			// target:source:state
			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: expectedError, Error: expectedError},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{
				Condition: graphql.FormationStatusConditionError,
				Errors: []*graphql.FormationStatusError{{
					Message:   "failed to parse request",
					ErrorCode: 2,
				}},
			})
			// The aggregated formation status is ERROR because of the FAs, but the Formation state should be READY
			require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)

			notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]

			assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			t.Logf("Assign application to formation %s should succeed on retry", formation.Name)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
			assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedConfig := str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 = notificationsForConsumerTenant.Array()[1]

			assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			var unassignFormation graphql.Formation
			t.Logf("Unassign application from formation %s", formation.Name)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formation.Name, assignedFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)

			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

			unassignNotificationFound := false
			for _, notification := range notificationsForConsumerTenant.Array() {
				op := notification.Get("Operation").String()
				if op == unassignOperation {
					appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
					t.Logf("Found notification for app %q", appIDFromNotification)
					if appIDFromNotification == app1.ID {
						unassignNotificationFound = true
						assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
					}
				}
			}
			require.True(t, unassignNotificationFound)

			t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, formationName)
			unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
		})

		t.Run("Fail Processing formation assignments while unassigning from formation", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
			defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

			webhookMode := graphql.WebhookModeSync
			webhookType := graphql.WebhookTypeConfigurationChanged
			urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{if eq .Operation \\\"unassign\\\"}}fail-once/{{end}}{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
			outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			runtimeWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

			t.Logf("Add webhook with type %q and mode: %q to provider runtime with ID %q", webhookType, webhookMode, providerRuntime.ID)
			actualWebhook := fixtures.AddWebhookToRuntime(t, ctx, directorCertSecuredClient, runtimeWebhookInput, "", providerRuntime.ID)
			defer fixtures.CleanupWebhook(t, ctx, directorCertSecuredClient, "", actualWebhook.ID)

			t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
			defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
			formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &formationTemplateName)
			require.NotEmpty(t, formation.ID)

			var assignedFormation graphql.Formation
			// Expect no formation assignments to be created
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
			assignReq := fixtures.FixAssignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, assignedFormation.Name, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			// Expect one formation assignment to be created
			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Assign application to formation %s", formation.Name)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerTenantID)
			assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)
			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
			}

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)

			notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]

			assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			var unassignFormation graphql.Formation
			// notification mock api should return error
			t.Logf("Unassign application from formation %s should fail.", formation.Name)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.Error(t, err)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
				},
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 2, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{
				Condition: graphql.FormationStatusConditionError,
				Errors: []*graphql.FormationStatusError{{
					Message:   "failed to parse request",
					ErrorCode: 2,
				}},
			})
			// The aggregated formation status is ERROR because of the FAs, but the Formation state should be READY
			require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 = notificationsForConsumerTenant.Array()[1]

			assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, unassignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			t.Logf("Unassign application from formation %s should succeed on retry", formation.Name)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formation.Name, assignedFormation.Name)

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)

			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

			unassignNotificationFound := false
			for _, notification := range notificationsForConsumerTenant.Array() {
				op := notification.Get("Operation").String()
				if op == unassignOperation {
					appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
					t.Logf("Found notification for app %q", appIDFromNotification)
					if appIDFromNotification == app1.ID {
						unassignNotificationFound = true
						assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
					}
				}
			}
			require.True(t, unassignNotificationFound)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, formationName)
			unassignReq = fixtures.FixUnassignFormationRequest(subscriptionConsumerSubaccountID, string(graphql.FormationObjectTypeTenant), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			// Expect formation assignments to be cleared
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
		})

		t.Run("Formation Assignment Notification Synchronous Resynchronization", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
			defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

			configurationChangedWebhookType := graphql.WebhookTypeConfigurationChanged
			webhookSyncMode := graphql.WebhookModeSync
			urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/fail-once/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
			outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			runtimeWebhookInput := fixtures.FixFormationNotificationWebhookInput(configurationChangedWebhookType, webhookSyncMode, urlTemplate, inputTemplate, outputTemplate)

			t.Logf("Add webhook with type %q and mode: %q to provider runtime with ID %q", configurationChangedWebhookType, webhookSyncMode, providerRuntime.ID)
			actualWebhook := fixtures.AddWebhookToRuntime(t, ctx, directorCertSecuredClient, runtimeWebhookInput, "", providerRuntime.ID)
			defer fixtures.CleanupWebhook(t, ctx, directorCertSecuredClient, "", actualWebhook.ID)

			urlTemplateApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/fail/{{.Application.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.RuntimeContext.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{.Runtime.Labels.region }}\\\",\\\"application-namespace\\\":\\\"\\\",\\\"application-tenant-id\\\":\\\"{{.RuntimeContext.Value}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.RuntimeContext.ID}}\\\"}]}"
			outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(configurationChangedWebhookType, webhookSyncMode, urlTemplateApplication, inputTemplateApplication, outputTemplateApplication)
			urlTemplateApplicationSucceeds := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/configuration/{{.Application.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.RuntimeContext.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			applicationWebhookInputThatSucceedsInput := fixtures.FixFormationNotificationWebhookInput(configurationChangedWebhookType, webhookSyncMode, urlTemplateApplicationSucceeds, inputTemplateApplication, outputTemplateApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", configurationChangedWebhookType, webhookSyncMode, app1.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, subscriptionConsumerAccountID, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationWebhook.ID)

			t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
			defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
			formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &formationTemplateName)
			require.NotEmpty(t, formation.ID)
			require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)   // Asserting only the formation state
			require.Equal(t, graphql.FormationStatusConditionReady, formation.Status.Condition) // Asserting the aggregated formation status

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
			defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationName, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
			assignedFormation := fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			// notification mock API should return error
			t.Logf("Assign application to formation %s should fail", formation.Name)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
			assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, subscriptionConsumerAccountID)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					app1.ID:  fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{
				Condition: graphql.FormationStatusConditionError,
				Errors: []*graphql.FormationStatusError{
					{
						Message:   "failed to parse request",
						ErrorCode: 2,
					},
					{
						Message:   "failed to parse request",
						ErrorCode: 2,
					},
				},
			})
			// The aggregated formation status is ERROR because of the FAs, but the Formation state should be READY
			require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)

			notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]

			assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			t.Logf("Resynchronize formation %s should retry and succeed for the runtime context", formation.Name)
			resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					app1.ID:  fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				app1.ID: {
					app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionError, Errors: []*graphql.FormationStatusError{{
				Message:   "failed to parse request",
				ErrorCode: 2,
			}}})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
			assignNotificationForApp1 = notificationsForConsumerTenant.Array()[1]

			assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
			defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Update application webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint which succeeds", actualApplicationWebhook.ID, configurationChangedWebhookType, webhookSyncMode)
			updatedWebhook := fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationWebhook.ID, applicationWebhookInputThatSucceedsInput)
			require.Equal(t, updatedWebhook.ID, actualApplicationWebhook.ID)

			var unassignFormation graphql.Formation
			t.Logf("Unassign application with ID: %s from formation: %s should fail due to runtime context notification failure", app1.ID, formation.Name)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.Error(t, err)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				app1.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 2, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{
				Condition: graphql.FormationStatusConditionError,
				Errors: []*graphql.FormationStatusError{{
					Message:   "failed to parse request",
					ErrorCode: 2,
				}},
			})
			// The aggregated formation status is ERROR because of the FAs, but the Formation state should be READY
			require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)

			notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

			unassignNotificationFound := false
			for _, notification := range notificationsForConsumerTenant.Array() {
				op := notification.Get("Operation").String()
				if op == unassignOperation {
					appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
					t.Logf("Found notification for app %q", appIDFromNotification)
					if appIDFromNotification == app1.ID {
						unassignNotificationFound = true
						assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
					}
				}
			}
			require.True(t, unassignNotificationFound)

			t.Logf("Check that the application with ID: %q is still assigned to formation: %s", app1.ID, formationName)
			app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
			scenarios, hasScenarios := app.Labels["scenarios"]
			assert.True(t, hasScenarios)
			assert.Len(t, scenarios, 1)

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				rtCtx.ID: {
					rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 4)

			t.Logf("Check that the application with ID: %q is unassigned from formation %s from formation after resyonchronization", app1.ID, formationName)
			assert.Contains(t, scenarios, formationName)
			app = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
			scenarios, hasScenarios = app.Labels["scenarios"]
			assert.False(t, hasScenarios)

			t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, formationName)
			unassignedFormation := fixtures.UnassignFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, subscriptionConsumerAccountID, subscriptionConsumerSubaccountID, graphql.FormationObjectTypeTenant)
			require.Equal(t, formation.ID, unassignedFormation.ID)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
		})

		t.Run("Formation Assignment Notification Asynchronous Resynchronization", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			webhookTypeRuntime := graphql.WebhookTypeConfigurationChanged
			webhookModeRuntime := graphql.WebhookModeAsyncCallback
			urlTemplateRuntime := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-fail-once/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateRuntime := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .Application.Labels.region }}{{.Application.Labels.region}}{{ else }}{{.ApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.ApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.Application.ID}}\\\"}]}"
			outputTemplateRuntime := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

			webhookThatFailsOnceInput := fixtures.FixFormationNotificationWebhookInput(webhookTypeRuntime, webhookModeRuntime, urlTemplateRuntime, inputTemplateRuntime, outputTemplateRuntime)

			t.Logf("Add webhook with type %q and mode: %q to provider runtime with ID %q", webhookTypeRuntime, webhookModeRuntime, providerRuntime.ID)
			actualRuntimeWebhook := fixtures.AddWebhookToRuntime(t, ctx, directorCertSecuredClient, webhookThatFailsOnceInput, "", providerRuntime.ID)
			defer fixtures.CleanupWebhook(t, ctx, directorCertSecuredClient, "", actualRuntimeWebhook.ID)

			webhookTypeApplication := graphql.WebhookTypeConfigurationChanged
			webhookModeApplication := graphql.WebhookModeSync
			urlTemplateApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/configuration/{{.Application.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.RuntimeContext.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{.Runtime.Labels.region }}\\\",\\\"application-namespace\\\":\\\"\\\",\\\"application-tenant-id\\\":\\\"{{.RuntimeContext.Value}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.RuntimeContext.ID}}\\\"}]}"
			outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookTypeApplication, webhookModeApplication, urlTemplateApplication, inputTemplateApplication, outputTemplateApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookTypeApplication, webhookModeApplication, app1.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, subscriptionConsumerAccountID, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationWebhook.ID)

			t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
			defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
			formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &formationTemplateName)
			require.NotEmpty(t, formation.ID)
			require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State) // Asserting only the formation state

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Run("Resynchronize when in CREATE_ERROR and DELETE_ERROR should resend notifications and succeed", func(t *testing.T) {
				cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
				defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
				resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
				defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

				t.Logf("Assign application to formation %s", formation.Name)
				defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
				assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, subscriptionConsumerAccountID)

				t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
				defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationName, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
				assignedFormation = fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)

				expectedError := str.Ptr(`{"error":{"message":"test error","errorCode":2}}`)
				expectedAssignmentsBySourceID := map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						app1.ID:  fixtures.AssignmentState{State: "CONFIG_PENDING", Config: nil, Value: nil, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
					app1.ID: {
						app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: expectedError, Error: expectedError},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionError,
					Errors: []*graphql.FormationStatusError{{
						Message:   "test error",
						ErrorCode: 2,
					}},
				})
				// The aggregated formation status is ERROR because of the FAs, but the Formation state should be READY
				require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)

				body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

				// rtCtx <- App notifications
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)
				notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
				assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]
				assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

				// rtCtx -> App notifications
				assertNotificationsCountForTenant(t, body, localTenantID, 1)
				notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
				assertExpectationsForApplicationNotificationsWithItemsStructure(t, notificationsForConsumerTenant.Array(), []*applicationFormationExpectations{
					{
						op:            assignOperation,
						formationID:   formation.ID,
						objectID:      rtCtx.ID,
						localTenantID: rtCtx.Value,
						objectRegion:  regionLbl,
						configuration: "",
						tenant:        subscriptionConsumerAccountID,
						customerID:    emptyParentCustomerID,
					},
				})

				t.Logf("Resynchronize formation %s should retry and succeed for the runtime context", formation.Name)
				resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						app1.ID:  fixtures.AssignmentState{State: "CONFIG_PENDING", Config: nil, Value: nil, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
					app1.ID: {
						app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
					},
				}

				assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})

				expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						app1.ID:  fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
					app1.ID: {
						app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

				body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

				// rtCtx <- App notifications
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)
				notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)
				assignNotificationForApp1 = notificationsForConsumerTenant.Array()[1]
				assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

				assertNotificationsCountForTenant(t, body, localTenantID, 2)
				notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
				assertExpectationsForApplicationNotificationsWithItemsStructure(t, notificationsForConsumerTenant.Array(), []*applicationFormationExpectations{
					{
						op:            assignOperation,
						formationID:   formation.ID,
						objectID:      rtCtx.ID,
						localTenantID: rtCtx.Value,
						objectRegion:  regionLbl,
						configuration: "",
						tenant:        subscriptionConsumerAccountID,
						customerID:    emptyParentCustomerID,
					},
					{
						op:            assignOperation,
						formationID:   formation.ID,
						objectID:      rtCtx.ID,
						localTenantID: rtCtx.Value,
						objectRegion:  regionLbl,
						configuration: `{"asyncKey":"asyncValue","asyncKey2":{"asyncNestedKey":"asyncNestedValue"}}`,
						tenant:        subscriptionConsumerAccountID,
						customerID:    emptyParentCustomerID,
					},
				})

				resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
				defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

				t.Logf("Unassign application from formation %s", formation.Name)
				unassignedFormation := fixtures.UnassignFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, subscriptionConsumerAccountID, app1.ID, graphql.FormationObjectTypeApplication)
				require.Equal(t, formation.ID, unassignedFormation.ID)

				application := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
				scenarios, hasScenarios := application.Labels["scenarios"]
				require.True(t, hasScenarios)
				require.Len(t, scenarios, 1)
				require.Contains(t, scenarios, formationName)

				body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

				// rtCtx <- App notifications
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 3)
				notificationsForConsumerTenant = gjson.GetBytes(body, subscriptionConsumerTenantID)

				unassignNotificationFound := false
				for _, notification := range notificationsForConsumerTenant.Array() {
					op := notification.Get("Operation").String()
					if op == unassignOperation {
						appIDFromNotification := notification.Get("RequestBody.items.0.ucl-system-tenant-id").String()
						t.Logf("Found notification for app %q", appIDFromNotification)
						if appIDFromNotification == app1.ID {
							unassignNotificationFound = true
							assertFormationAssignmentsNotificationWithConfigContainingItemsStructure(t, notification, unassignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID, fixtures.StatusAPISyncConfigJSON)
						}
					}
				}
				require.True(t, unassignNotificationFound)

				// rtCtx -> App notifications
				assertNotificationsCountForTenant(t, body, localTenantID, 3)
				notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
				assertSeveralFormationAssignmentsNotifications(t, notificationsForConsumerTenant, rtCtx, formation.ID, regionLbl, unassignOperation, subscriptionConsumerAccountID, emptyParentCustomerID, 1)

				expectedAssignments := map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
					app1.ID: {
						rtCtx.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: expectedError, Error: expectedError},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 2, expectedAssignments, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionError,
					Errors: []*graphql.FormationStatusError{{
						Message:   "test error",
						ErrorCode: 2,
					}},
				})
				// The aggregated formation status is ERROR because of the FAs, but the Formation state should be READY
				require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)

				t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
				resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				expectedAssignments = map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady})

				body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 4)

				t.Logf("Check that application with ID %q is unassigned from formation %s", app1.ID, formationName)
				app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, app1.ID)
				scenarios, hasScenarios = app.Labels["scenarios"]
				require.False(t, hasScenarios)

				t.Logf("Check that runtime context with ID %q is still assigned to formation %s", subscriptionConsumerSubaccountID, formationName)
				actualRtmCtx := fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
				scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
				require.True(t, hasScenarios)
				require.Len(t, scenarios, 1)
				require.Contains(t, scenarios, formationName)

				t.Logf("Unassign tenant %s from formation %s", subscriptionConsumerSubaccountID, formationName)
				unassignedFormation = fixtures.UnassignFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, subscriptionConsumerAccountID, subscriptionConsumerSubaccountID, graphql.FormationObjectTypeTenant)
				require.Equal(t, formation.ID, unassignedFormation.ID)

				t.Logf("Check that runtime context with ID %q is actually unassigned from formation %s", subscriptionConsumerSubaccountID, formationName)
				actualRtmCtx = fixtures.GetRuntimeContext(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, consumerSubaccountRuntime.ID, rtCtx.ID)
				scenarios, hasScenarios = actualRtmCtx.Labels["scenarios"]
				require.False(t, hasScenarios)
			})

			t.Run("Resynchronize when in INITIAL and DELETING should resend notifications and succeed", func(t *testing.T) {
				cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
				defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
				resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
				defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

				urlTemplateThatNeverResponds := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-no-response/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
				webhookThatNeverRespondsInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeConfigurationChanged, graphql.WebhookModeAsyncCallback, urlTemplateThatNeverResponds, inputTemplateRuntime, outputTemplateRuntime)

				t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint which never responds", actualRuntimeWebhook.ID, graphql.WebhookTypeConfigurationChanged, graphql.WebhookModeAsyncCallback)
				updatedWebhook := fixtures.UpdateWebhook(t, ctx, directorCertSecuredClient, "", actualRuntimeWebhook.ID, webhookThatNeverRespondsInput)
				require.Equal(t, updatedWebhook.ID, actualRuntimeWebhook.ID)

				t.Logf("Assign tenant %s to formation %s", subscriptionConsumerSubaccountID, formationName)
				defer fixtures.CleanupFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, formationName, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)
				assignedFormation := fixtures.AssignFormationWithTenantObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, subscriptionConsumerSubaccountID, subscriptionConsumerAccountID)

				t.Logf("Assign application to formation %s", formation.Name)
				defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
				assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, subscriptionConsumerAccountID)

				expectedAssignmentsBySourceID := map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						app1.ID:  fixtures.AssignmentState{State: "CONFIG_PENDING", Config: nil, Value: nil, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
					app1.ID: {
						app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})
				// The aggregated formation status is IN_PROGRESS because of the FAs, but the Formation state should be READY
				require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)

				body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

				// rtCtx <- App notifications
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)
				notificationsForConsumerTenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
				assignNotificationForApp1 := notificationsForConsumerTenant.Array()[0]
				assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationForApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

				// rtCtx -> App notifications
				assertNotificationsCountForTenant(t, body, localTenantID, 1)
				notificationsForConsumerTenant = gjson.GetBytes(body, localTenantID)
				assertExpectationsForApplicationNotificationsWithItemsStructure(t, notificationsForConsumerTenant.Array(), []*applicationFormationExpectations{
					{
						op:            assignOperation,
						formationID:   formation.ID,
						objectID:      rtCtx.ID,
						localTenantID: rtCtx.Value,
						objectRegion:  regionLbl,
						configuration: "",
						tenant:        subscriptionConsumerAccountID,
						customerID:    emptyParentCustomerID,
					},
				})

				urlTemplateThatSucceeds := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.RuntimeContext.Value}}{{if eq .Operation \\\"unassign\\\"}}/{{.Application.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"

				webhookThatSucceeds := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeConfigurationChanged, graphql.WebhookModeAsyncCallback, urlTemplateThatSucceeds, inputTemplateRuntime, outputTemplateRuntime)

				t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that responds with success", actualRuntimeWebhook.ID, graphql.WebhookTypeConfigurationChanged, graphql.WebhookModeAsyncCallback)
				updatedWebhook = fixtures.UpdateWebhook(t, ctx, directorCertSecuredClient, "", actualRuntimeWebhook.ID, webhookThatSucceeds)
				require.Equal(t, updatedWebhook.ID, actualRuntimeWebhook.ID)

				t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
				resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				expectedConfig := str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
				asyncConfig := str.Ptr("{\"asyncKey\":\"asyncValue\",\"asyncKey2\":{\"asyncNestedKey\":\"asyncNestedValue\"}}")
				expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						app1.ID:  fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
					app1.ID: {
						app1.ID:  fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: asyncConfig, Value: asyncConfig, Error: nil},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

				body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)

				t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint which never responds", actualRuntimeWebhook.ID, graphql.WebhookTypeConfigurationChanged, graphql.WebhookModeAsyncCallback)
				updatedWebhook = fixtures.UpdateWebhook(t, ctx, directorCertSecuredClient, "", actualRuntimeWebhook.ID, webhookThatNeverRespondsInput)
				require.Equal(t, updatedWebhook.ID, actualRuntimeWebhook.ID)

				t.Logf("Unassign application from formation %s", formation.Name)
				unassignedFormation := fixtures.UnassignFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, subscriptionConsumerAccountID, app1.ID, graphql.FormationObjectTypeApplication)
				require.Equal(t, formation.ID, unassignedFormation.ID)

				expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
					app1.ID: {
						rtCtx.ID: fixtures.AssignmentState{State: "DELETING", Config: nil, Value: nil, Error: nil},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 2, expectedAssignmentsBySourceID, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})
				// The aggregated formation status is IN_PROGRESS because of the FAs, but the Formation state should be READY
				require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)

				t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that responds with success", actualRuntimeWebhook.ID, graphql.WebhookTypeConfigurationChanged, graphql.WebhookModeAsyncCallback)
				updatedWebhook = fixtures.UpdateWebhook(t, ctx, directorCertSecuredClient, "", actualRuntimeWebhook.ID, webhookThatSucceeds)
				require.Equal(t, updatedWebhook.ID, actualRuntimeWebhook.ID)

				t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
				resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
				err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, resynchronizeReq, &assignedFormation)
				require.NoError(t, err)
				require.Equal(t, formationName, assignedFormation.Name)

				expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
					rtCtx.ID: {
						rtCtx.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					},
				}
				assertFormationAssignmentsAsynchronously(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignmentsBySourceID, 300)
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

				body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
				assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 4)
			})
		})
	})
}
