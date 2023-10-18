package notifications

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"

	directordestinationcreator "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestFormationNotificationsWithApplicationSubscription(stdT *testing.T) {
	t := testingx.NewT(stdT)
	ctx := context.Background()

	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	subscriptionSubdomain := conf.SelfRegisterSubdomainPlaceholderValue
	subscriptionConsumerAccountID := conf.TestConsumerAccountID
	subscriptionProviderSubaccountID := conf.TestProviderSubaccountID
	subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID
	subscriptionConsumerTenantID := conf.TestConsumerTenantID

	applicationType1 := "app-subscription-notifications-type-1"
	applicationType2 := "app-subscription-notifications-type-2"
	providerFormationTmplName := "app-subscription-notifications-template-name"
	formationName := "app-subscription-notifications-name"

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
		},
	}

	certSubject := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.ExternalCertProviderConfig.TestExternalCertCN, "app-template-subscription-cn", -1)

	// We need an externally issued cert with a subject that is not part of the access level mappings
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceTestSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               certSubject,
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.CertificateService,
	}

	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(stdT, ctx, externalCertProviderConfig, false)
	appProviderDirectorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

	// The external cert secret created by the NewExternalCertFromConfig above is used by the external-services-mock for the async formation status API call,
	// that's why in the function above there is a false parameter that don't delete it and an explicit defer deletion func is added here
	// so, the secret could be deleted at the end of the test. Otherwise, it will remain as leftover resource in the cluster
	defer func() {
		k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
		require.NoError(t, err)
		k8s.DeleteSecret(stdT, ctx, k8sClient, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace)
	}()

	apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)

	t.Run("Formation Notifications With Application Subscription", func(t *testing.T) {
		// Create Application Template
		namePlaceholder := "name"
		displayNamePlaceholder := "display-name"
		appRegion := "test-app-region"
		appNamespace := "compass.test"
		localTenantID := "local-tenant-id"
		localTenantID2 := "local-tenant-id2"
		app1BaseURL := "http://e2e-test-app1-base-url"
		app2BaseURL := "http://e2e-test-app2-base-url"

		appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType1, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
		appTemplateInput.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue
		appTemplateInput.ApplicationInput.Labels[conf.GlobalSubaccountIDLabelKey] = subscriptionConsumerSubaccountID
		appTemplateInput.ApplicationInput.BaseURL = &app1BaseURL
		appTemplateInput.ApplicationInput.LocalTenantID = nil
		for i := range appTemplateInput.Placeholders {
			appTemplateInput.Placeholders[i].JSONPath = str.Ptr(fmt.Sprintf("$.%s", conf.SubscriptionProviderAppNameProperty))
		}

		appTmpl, err := fixtures.CreateApplicationTemplateFromInput(stdT, ctx, appProviderDirectorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(stdT, ctx, appProviderDirectorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), appTmpl)
		require.NoError(stdT, err)
		require.NotEmpty(stdT, appTmpl.ID)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

		selfRegLabelValue, ok := appTmpl.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
		require.True(stdT, ok)
		require.Contains(stdT, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+appTmpl.ID)

		deps, err := json.Marshal([]string{selfRegLabelValue, conf.ProviderDestinationConfig.Dependency})
		require.NoError(stdT, err)
		depConfigureReq, err := http.NewRequest(http.MethodPost, conf.ExternalServicesMockBaseURL+"/v1/dependencies/configure", bytes.NewBuffer(deps))
		require.NoError(stdT, err)
		response, err := httpClient.Do(depConfigureReq)
		defer func() {
			if err := response.Body.Close(); err != nil {
				stdT.Logf("Could not close response body %s", err)
			}
		}()
		require.NoError(stdT, err)
		require.Equal(stdT, http.StatusOK, response.StatusCode)

		subscriptionToken := token.GetClientCredentialsToken(t, ctx, conf.SubscriptionConfig.TokenURL+conf.TokenPath, conf.SubscriptionConfig.ClientID, conf.SubscriptionConfig.ClientSecret, claims.TenantFetcherClaimKey)

		defer subscription.BuildAndExecuteUnsubscribeRequest(t, appTmpl.ID, appTmpl.Name, httpClient, conf.SubscriptionConfig.URL, apiPath, subscriptionToken, conf.SubscriptionConfig.PropagatedProviderSubaccountHeader, subscriptionConsumerSubaccountID, subscriptionConsumerTenantID, subscriptionProviderSubaccountID, conf.SubscriptionConfig.StandardFlow, conf.SubscriptionConfig.SubscriptionFlowHeaderKey)
		subscription.CreateSubscription(t, conf.SubscriptionConfig, httpClient, appTmpl, apiPath, subscriptionToken, subscriptionConsumerTenantID, subscriptionConsumerSubaccountID, subscriptionProviderSubaccountID, conf.SubscriptionProviderAppNameValue, true, true, conf.SubscriptionConfig.StandardFlow)

		actualAppPage := graphql.ApplicationPage{}
		getSrcAppReq := fixtures.FixGetApplicationsRequestWithPagination()
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerSubaccountID, getSrcAppReq, &actualAppPage)
		require.NoError(t, err)

		require.Len(t, actualAppPage.Data, 1)
		require.Equal(t, appTmpl.ID, *actualAppPage.Data[0].ApplicationTemplateID)
		app1 := *actualAppPage.Data[0]
		t.Logf("app1 ID: %q", app1.ID)

		t.Log("Create integration system")
		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, "app-subscription-notifications")
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

		t.Logf("Create application template for type %q", applicationType2)
		appTemplateInput = fixtures.FixApplicationTemplateWithoutWebhook(applicationType2, localTenantID2, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
		appTemplateInput.ApplicationInput.Labels[conf.GlobalSubaccountIDLabelKey] = subscriptionConsumerSubaccountID
		appTemplateInput.ApplicationInput.BaseURL = &app2BaseURL
		appTmpl2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl2)
		require.NoError(t, err)

		// Create certificate subject mapping with custom subject that was used to create a certificate for the graphql client above
		internalConsumerID := appTmpl2.ID // add application templated ID as certificate subject mapping internal consumer to satisfy the authorization checks in the formation assignment status API
		certSubjectMappingCustomSubjectWithCommaSeparator := strings.ReplaceAll(strings.TrimLeft(certSubject, "/"), "/", ",")
		csmInput := fixtures.FixCertificateSubjectMappingInput(certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, &internalConsumerID, tenantAccessLevels)
		t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, tenantAccessLevels)

		var csmCreate graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
		defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
		csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)

		t.Logf("Sleeping for %s, so the hydrator component could update the certificate subject mapping cache with the new data", conf.CertSubjectMappingResyncInterval.String())
		time.Sleep(conf.CertSubjectMappingResyncInterval)

		t.Logf("Create application 2 from template %q", applicationType2)
		appFromTmplSrc2 := fixtures.FixApplicationFromTemplateInput(applicationType2, namePlaceholder, "app2-formation-notifications-tests", displayNamePlaceholder, "App 2 Display Name")
		appFromTmplSrc2GQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc2)
		require.NoError(t, err)
		createAppFromTmplSecondRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrc2GQL)
		app2 := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, createAppFromTmplSecondRequest, &app2)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, &app2)
		require.NoError(t, err)
		require.NotEmpty(t, app2.ID)
		t.Logf("app2 ID: %q", app2.ID)

		t.Logf("Creating formation template for the provider runtime type %q with name %q", conf.SubscriptionProviderAppNameValue, providerFormationTmplName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(stdT, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateFormationTemplateWithoutInput(stdT, ctx, certSecuredGraphQLClient, providerFormationTmplName, conf.SubscriptionProviderAppNameValue, []string{applicationType1, applicationType2}, graphql.ArtifactTypeSubscription)

		t.Run("Synchronous App to App Formation Assignment Notifications", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			webhookType := graphql.WebhookTypeApplicationTenantMapping
			webhookMode := graphql.WebhookModeSync
			urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"subdomain\\\": \\\"{{ if eq .TargetApplication.Tenant.Type \\\"subaccount\\\"}}{{ .TargetApplication.Tenant.Labels.subdomain }}{{end}}\\\"}], \\\"app-template\\\": \\\"{{ .TargetApplicationTemplate.Tenant.Labels}}\\\"}"
			outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, app1.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, subscriptionConsumerAccountID, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationWebhook.ID)

			// Create formation constraints for destination creator operator and attach them to a given formation template.
			// So we can verify the destination creator will not fail if in the configuration there is no destination information
			attachDestinationCreatorConstraints(t, ctx, ft, graphql.ResourceTypeApplication, graphql.ResourceTypeApplication)

			t.Logf("Creating formation with name: %q from template with name: %q", formationName, providerFormationTmplName)
			defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
			formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &providerFormationTmplName)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

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

			t.Logf("Assign application 2 to formation %s", formationName)
			defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, subscriptionConsumerAccountID)
			assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, assignReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 1)

			notificationsForApp1 := gjson.GetBytes(body, app1.ID)
			assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
			assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
			assertFormationAssignmentsNotificationSubdomainWithItemsStructure(t, assignNotificationAboutApp2, subscriptionSubdomain)

			t.Logf("Unassign Application 1 from formation %s", formationName)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var unassignFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 2)

			notificationsForApp1 = gjson.GetBytes(body, app1.ID)
			unassignNotificationFound := false
			for _, notification := range notificationsForApp1.Array() {
				op := notification.Get("Operation").String()
				if op == unassignOperation {
					unassignNotificationFound = true
					assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)
					assertFormationAssignmentsNotificationSubdomainWithItemsStructure(t, notification, subscriptionSubdomain)
				}
			}
			require.True(t, unassignNotificationFound, "notification for unassign app2 not found")

			t.Logf("Unassign Application 2 from formation %s", formationName)
			unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
		})

		t.Run("Formation assignment notifications with destination creation/deletion", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			cleanupDestinationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupDestinationsFromExternalSvcMock(t, certSecuredHTTPClient)

			cleanupDestnationCertificatesFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupDestnationCertificatesFromExternalSvcMock(t, certSecuredHTTPClient)

			// first app
			applicationTntMappingWebhookType := graphql.WebhookTypeApplicationTenantMapping
			asyncCallbackWebhookMode := graphql.WebhookModeAsyncCallback
			urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/destinations/{{.TargetApplication.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateApplication := "{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}"
			outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.configuration}}\\\",\\\"state\\\":\\\"{{.Body.state}}\\\",\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\":\\\"{{.Body.error}}\\\",\\\"success_status_code\\\":202}"

			applicationAsyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, asyncCallbackWebhookMode, urlTemplateAsyncApplication, inputTemplateApplication, outputTemplateAsyncApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", applicationTntMappingWebhookType, asyncCallbackWebhookMode, app2.ID)
			actualApplicationAsyncWebhookInput := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationAsyncWebhookInput, subscriptionConsumerAccountID, app2.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationAsyncWebhookInput.ID)

			// second app
			syncWebhookMode := graphql.WebhookModeSync
			urlTemplateSyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/destinations/configuration/{{.TargetApplication.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			outputTemplateSyncApplication := "{\\\"config\\\":\\\"{{.Body.configuration}}\\\",\\\"state\\\":\\\"{{.Body.state}}\\\",\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\":200}"

			applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, syncWebhookMode, urlTemplateSyncApplication, inputTemplateApplication, outputTemplateSyncApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", applicationTntMappingWebhookType, syncWebhookMode, app1.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, subscriptionConsumerAccountID, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, actualApplicationWebhook.ID)

			// create formation constraints and attach them to formation template
			firstConstraintInput := graphql.FormationConstraintInput{
				Name:            "e2e-destination-creator-notification-status-returned",
				ConstraintType:  graphql.ConstraintTypePre,
				TargetOperation: graphql.TargetOperationNotificationStatusReturned,
				Operator:        formationconstraintpkg.DestinationCreator,
				ResourceType:    graphql.ResourceTypeApplication,
				ResourceSubtype: applicationType1,
				InputTemplate:   "{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .FormationAssignment }}\\\"details_formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"details_reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}",
				ConstraintScope: graphql.ConstraintScopeFormationType,
			}

			firstConstraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, firstConstraintInput)
			defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, firstConstraint.ID)
			require.NotEmpty(t, firstConstraint.ID)

			fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, firstConstraint.ID, firstConstraint.Name, ft.ID, ft.Name)

			// second constraint
			secondConstraintInput := graphql.FormationConstraintInput{
				Name:            "e2e-destination-creator-send-notification",
				ConstraintType:  graphql.ConstraintTypePre,
				TargetOperation: graphql.TargetOperationSendNotification,
				Operator:        formationconstraintpkg.DestinationCreator,
				ResourceType:    graphql.ResourceTypeApplication,
				ResourceSubtype: applicationType1,
				InputTemplate:   "{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .FormationAssignment }}\\\"details_formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"details_reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}",
				ConstraintScope: graphql.ConstraintScopeFormationType,
			}

			secondConstraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, secondConstraintInput)
			defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, secondConstraint.ID)
			require.NotEmpty(t, secondConstraint.ID)

			fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, secondConstraint.ID, secondConstraint.Name, ft.ID, ft.Name)

			// create formation
			t.Logf("Creating formation with name: %q from template with name: %q", formationName, providerFormationTmplName)
			defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName)
			formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formationName, &providerFormationTmplName)
			require.NotEmpty(t, formation.ID)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			formationInput := graphql.FormationInput{Name: formationName}
			t.Logf("Assign application 2 with ID: %s to formation: %q", app2.ID, formationName)
			defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app2.ID, subscriptionConsumerAccountID)
			assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app2.ID, subscriptionConsumerAccountID)
			require.Equal(t, formation.ID, assignedFormation.ID)
			require.Equal(t, formation.State, assignedFormation.State)

			t.Log("Assert no formation assignment notifications are sent when there is only one app in formation")
			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 0)
			assertNotificationsCountForTenant(t, body, localTenantID2, 0)

			expectedAssignmentsBySourceID := map[string]map[string]fixtures.AssignmentState{
				app2.ID: {
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
			}
			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Assign application 1 with ID: %s to formation %s", app1.ID, formationName)
			defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app1.ID, subscriptionConsumerAccountID)
			assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app1.ID, subscriptionConsumerAccountID)
			require.Equal(t, formationName, assignedFormation.Name)

			assignmentWithDestDetails := fixtures.GetFormationAssignmentsBySourceAndTarget(t, ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, formation.ID, app2.ID, app1.ID)
			require.NotEmpty(t, assignmentWithDestDetails)

			t.Logf("Assert formation assignments during %s operation...", assignOperation)
			noAuthDestinationName := "e2e-design-time-destination-name"
			noAuthDestinationURL := "http://e2e-design-time-url-example.com"
			basicDestinationName := "e2e-basic-destination-name"
			basicDestinationURL := "http://e2e-basic-url-example.com"
			samlAssertionDestinationName := "e2e-saml-assertion-destination-name"
			samlAssertionDestinationURL := "http://e2e-saml-url-example.com"
			clientCertAuthDestinationName := "e2e-client-cert-auth-destination-name"
			clientCertAuthDestinationURL := "http://e2e-client-cert-auth-url-example.com"
			testDestinationInstanceID := conf.TestDestinationInstanceID

			samlAssertionDestinationCertName := fmt.Sprintf("%s-%s", directordestinationcreator.AuthTypeSAMLAssertion, assignmentWithDestDetails.ID)
			clientCertAuthDestinationCertName := fmt.Sprintf("%s-%s", directordestinationcreator.AuthTypeClientCertificate, assignmentWithDestDetails.ID)
			clientCertAuthDestinationCertName = clientCertAuthDestinationCertName[:directordestinationcreator.MaxDestinationNameLength] // due to the longer client cert auth destination prefix we exceed the maximum length of the name, that's we truncate it

			// NoAuthentication destination on 'provider' subaccount level
			// BasicDestination on 'provider' instance level
			// Client Certificate Authentication destination on 'consumer' subaccount(implicitly) level
			// SAML Assertion destination in the 'consumer' subaccount(implicitly) on provider instance level
			destinationDetailsConfigWithPlaceholders := "{\"destinations\":[{\"name\":\"%s\",\"type\":\"HTTP\",\"description\":\"e2e-design-time-destination description\",\"proxyType\":\"Internet\",\"authentication\":\"NoAuthentication\",\"url\":\"%s\", \"subaccountId\":\"%s\"}],\"credentials\":{\"inboundCommunication\":{\"basicAuthentication\":{\"correlationIds\":[\"e2e-basic-correlation-ids\"],\"destinations\":[{\"name\":\"%s\",\"description\":\"e2e-basic-destination description\",\"url\":\"%s\",\"authentication\":\"BasicAuthentication\",\"subaccountId\":\"%s\",\"instanceId\":\"%s\", \"additionalProperties\":{\"e2e-basic-testKey\":\"e2e-basic-testVal\"}}]},\"samlAssertion\":{\"correlationIds\":[\"e2e-saml-correlation-ids\"],\"destinations\":[{\"name\":\"%s\",\"description\":\"e2e saml assertion destination description\",\"url\":\"%s\",\"instanceId\":\"%s\",\"additionalProperties\":{\"e2e-samlTestKey\":\"e2e-samlTestVal\"}}]},\"clientCertificateAuthentication\":{\"correlationIds\":[\"e2e-client-cert-auth-correlation-ids\"],\"destinations\":[{\"name\":\"%s\",\"description\":\"e2e client cert auth destination description\",\"url\":\"%s\",\"additionalProperties\":{\"e2e-clientCertAuthTestKey\":\"e2e-clientCertAuthTestVal\"}}]}}},\"additionalProperties\":[{\"propertyName\":\"example-property-name\",\"propertyValue\":\"example-property-value\",\"correlationIds\":[\"correlation-ids\"]}]}"
			destinationDetailsConfig := fmt.Sprintf(destinationDetailsConfigWithPlaceholders, noAuthDestinationName, noAuthDestinationURL, conf.TestProviderSubaccountID, basicDestinationName, basicDestinationURL, conf.TestProviderSubaccountID, testDestinationInstanceID, samlAssertionDestinationName, samlAssertionDestinationURL, testDestinationInstanceID, clientCertAuthDestinationName, clientCertAuthDestinationURL)
			destinationCredentialsConfig := "{\"credentials\":{\"outboundCommunication\":{\"basicAuthentication\":{\"url\":\"https://e2e-basic-destination-url.com\",\"username\":\"e2e-basic-destination-username\",\"password\":\"e2e-basic-destination-password\"},\"samlAssertion\":{\"url\":\"http://e2e-saml-url-example.com\"},\"clientCertificateAuthentication\":{\"url\":\"http://e2e-client-cert-auth-url-example.com\"}}}}"
			expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app2.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil},
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "CONFIG_PENDING", Config: str.Ptr(destinationDetailsConfig)},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
			}
			assertFormationAssignmentsWithDestinationConfig(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID, app2.ID, app1.ID)
			// The aggregated formation status is IN_PROGRESS because of the FAs, but the Formation state should be READY
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})
			require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)
			require.Empty(t, formation.Error)

			expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app2.ID: fixtures.AssignmentState{State: "READY", Config: str.Ptr(destinationCredentialsConfig)},
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
			}

			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, subscriptionConsumerAccountID, formation.ID, 4, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Assert formation assignment notifications for %s operation...", assignOperation)
			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 2)
			notificationsForApp1Tenant := gjson.GetBytes(body, subscriptionConsumerTenantID)
			validateFormationNameInAssignmentNotification(t, notificationsForApp1Tenant.Array()[0], formationName)
			assertExpectationsForApplicationNotifications(t, notificationsForApp1Tenant.Array(), []*applicationFormationExpectations{
				{
					op:                  assignOperation,
					formationID:         formation.ID,
					objectID:            app1.ID,
					localTenantID:       subscriptionConsumerTenantID,
					objectRegion:        appRegion,
					configuration:       "",
					tenant:              subscriptionConsumerAccountID,
					customerID:          emptyParentCustomerID,
					receiverTenantState: initialAssignmentState,
					assignedTenantState: initialAssignmentState,
				},
				{
					op:                                     assignOperation,
					formationID:                            formation.ID,
					objectID:                               app1.ID,
					localTenantID:                          subscriptionConsumerTenantID,
					objectRegion:                           appRegion,
					configuration:                          destinationDetailsConfig,
					tenant:                                 subscriptionConsumerAccountID,
					customerID:                             emptyParentCustomerID,
					receiverTenantState:                    configPendingAssignmentState,
					assignedTenantState:                    readyAssignmentState,
					shouldRemoveDestinationCertificateData: true,
				},
			})

			notificationsForApp2Tenant := gjson.GetBytes(body, localTenantID2)
			validateFormationNameInAssignmentNotification(t, notificationsForApp2Tenant.Array()[0], formationName)
			assertExpectationsForApplicationNotifications(t, notificationsForApp2Tenant.Array(), []*applicationFormationExpectations{
				{
					op:                  assignOperation,
					formationID:         formation.ID,
					objectID:            app2.ID,
					localTenantID:       localTenantID2,
					objectRegion:        appRegion,
					configuration:       "",
					tenant:              subscriptionConsumerAccountID,
					customerID:          emptyParentCustomerID,
					receiverTenantState: initialAssignmentState,
					assignedTenantState: configPendingAssignmentState,
				},
			})

			// Configure destination service client
			region := conf.SubscriptionConfig.SelfRegRegion
			instance, ok := conf.DestinationsConfig.RegionToInstanceConfig[region]
			require.True(t, ok)
			destinationClient, err := clients.NewDestinationClient(instance, conf.DestinationAPIConfig, conf.DestinationConsumerSubdomainMtls)
			require.NoError(t, err)

			consumerTokenURL, err := buildConsumerTokenURL(conf.ProviderDestinationConfig.TokenURL, conf.DestinationConsumerSubdomain)
			require.NoError(t, err)

			destinationProviderToken := token.GetClientCredentialsToken(t, ctx, conf.ProviderDestinationConfig.TokenURL+conf.ProviderDestinationConfig.TokenPath, conf.ProviderDestinationConfig.ClientID, conf.ProviderDestinationConfig.ClientSecret, claims.DestinationProviderClaimKey)
			destinationProviderWithInstanceToken := token.GetClientCredentialsToken(t, ctx, conf.ProviderDestinationConfig.TokenURL+conf.ProviderDestinationConfig.TokenPath, conf.ProviderDestinationConfig.ClientID, conf.ProviderDestinationConfig.ClientSecret, claims.DestinationProviderWithInstanceClaimKey)
			destinationConsumerToken := token.GetClientCredentialsToken(t, ctx, consumerTokenURL+conf.ProviderDestinationConfig.TokenPath, conf.ProviderDestinationConfig.ClientID, conf.ProviderDestinationConfig.ClientSecret, claims.DestinationConsumerClaimKey)
			destinationConsumerWithInstanceToken := token.GetClientCredentialsToken(t, ctx, consumerTokenURL+conf.ProviderDestinationConfig.TokenPath, conf.ProviderDestinationConfig.ClientID, conf.ProviderDestinationConfig.ClientSecret, claims.DestinationConsumerWithInstanceClaimKey)

			t.Log("Assert destinations and destination certificates are created...")
			assertNoAuthDestination(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, noAuthDestinationName, noAuthDestinationURL, "", destinationProviderToken)
			assertBasicDestination(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, basicDestinationName, basicDestinationURL, testDestinationInstanceID, destinationProviderWithInstanceToken)
			assertSAMLAssertionDestination(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, samlAssertionDestinationName, samlAssertionDestinationCertName, samlAssertionDestinationURL, app2BaseURL, testDestinationInstanceID, destinationConsumerWithInstanceToken)
			assertClientCertAuthDestination(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, clientCertAuthDestinationName, clientCertAuthDestinationCertName, clientCertAuthDestinationURL, "", destinationConsumerToken)
			assertDestinationCertificate(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, samlAssertionDestinationCertName+directordestinationcreator.JavaKeyStoreFileExtension, testDestinationInstanceID, destinationConsumerWithInstanceToken)
			assertDestinationCertificate(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, clientCertAuthDestinationCertName+directordestinationcreator.JavaKeyStoreFileExtension, "", destinationConsumerToken)
			t.Log("Destinations and destination certificates have been successfully created")

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			unassignStartTime := time.Now()
			var unassignFormation graphql.Formation
			t.Logf("Unassign Application 1 from formation %s", formationName)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, graphql.FormationObjectTypeApplication.String(), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)
			unassignDuration := time.Since(unassignStartTime)

			// assert the intermediate FA state only if the unassign operation duration is within the fixed tenant mapping async delay
			if unassignDuration < time.Second*time.Duration(conf.TenantMappingAsyncResponseDelay) {
				expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
					app1.ID: {
						app2.ID: fixtures.AssignmentState{State: "DELETING", Config: nil},
					},
					app2.ID: {
						app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
					},
				}

				t.Logf("Assert formation assignments intermediate state during %s operation of the first app...", unassignOperation)
				assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 2, expectedAssignmentsBySourceID)
				// The aggregated formation status is IN_PROGRESS because of the FAs, but the Formation state should be READY
				assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})
				require.Equal(t, graphql.FormationStatusConditionReady.String(), formation.State)
				require.Empty(t, formation.Error)
			}

			t.Logf("Assert destinations and destination certificates are deleted as part of the unassign operation...")
			assertNoDestinationIsFound(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, noAuthDestinationName, "", destinationProviderToken)
			assertNoDestinationIsFound(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, basicDestinationName, testDestinationInstanceID, destinationProviderWithInstanceToken)
			assertNoDestinationIsFound(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, samlAssertionDestinationName, testDestinationInstanceID, destinationConsumerWithInstanceToken)
			assertNoDestinationIsFound(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, clientCertAuthDestinationName, "", destinationConsumerToken)
			assertNoDestinationCertificateIsFound(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, samlAssertionDestinationCertName+directordestinationcreator.JavaKeyStoreFileExtension, testDestinationInstanceID, destinationConsumerWithInstanceToken)
			assertNoDestinationCertificateIsFound(t, destinationClient, conf.ProviderDestinationConfig.ServiceURL, clientCertAuthDestinationCertName+directordestinationcreator.JavaKeyStoreFileExtension, "", destinationConsumerToken)
			t.Logf("Destinations and destination certificates are successfully deleted as part of the unassign operation")

			expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
				app2.ID: {
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil},
				},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, subscriptionConsumerAccountID, formation.ID, 1, expectedAssignmentsBySourceID)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Assert formation assignment notifications for %s operation of the first app...", unassignOperation)
			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, subscriptionConsumerTenantID, 1)
			notificationsForApp1 := gjson.GetBytes(body, subscriptionConsumerTenantID)
			unassignNotificationForApp1 := notificationsForApp1.Array()[0]
			validateFormationNameInAssignmentNotification(t, unassignNotificationForApp1, formationName)
			assertFormationAssignmentsNotification(t, unassignNotificationForApp1, unassignOperation, formation.ID, app2.ID, app1.ID, deletingAssignmentState, deletingAssignmentState, subscriptionConsumerTenantID, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			assertNotificationsCountForTenant(t, body, localTenantID2, 1)
			notificationsForApp2 := gjson.GetBytes(body, localTenantID2)
			unassignNotificationForApp2 := notificationsForApp2.Array()[0]
			validateFormationNameInAssignmentNotification(t, unassignNotificationForApp2, formationName)
			assertFormationAssignmentsNotification(t, unassignNotificationForApp2, unassignOperation, formation.ID, app1.ID, app2.ID, deletingAssignmentState, deletingAssignmentState, localTenantID2, appNamespace, appRegion, subscriptionConsumerAccountID, emptyParentCustomerID)

			t.Logf("Unassign Application 2 from formation %s", formationName)
			unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, graphql.FormationObjectTypeApplication.String(), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, subscriptionConsumerAccountID, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			assertFormationAssignments(t, ctx, subscriptionConsumerAccountID, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, subscriptionConsumerAccountID, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
		})
	})
}
