package tests

//
//import (
//	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
//	"github.com/kyma-incubator/compass/components/director/pkg/str"
//	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
//	"github.com/kyma-incubator/compass/tests/pkg/clients"
//	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
//	"github.com/kyma-incubator/compass/tests/pkg/gql"
//	"github.com/kyma-incubator/compass/tests/pkg/k8s"
//	"github.com/kyma-incubator/compass/tests/pkg/tenant"
//	"github.com/kyma-incubator/compass/tests/pkg/testctx"
//	"github.com/kyma-incubator/compass/tests/pkg/token"
//	"github.com/stretchr/testify/require"
//	"github.com/tidwall/gjson"
//	"strings"
//	"testing"
//	"time"
//)
//
//func TestFormationNotificationsWithApplicationOnlyParticipantss(t *testing.T) {
//	tnt := tenant.TestTenants.GetDefaultTenantID()
//	tntParentCustomer := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant) // parent of `tenant.TestTenants.GetDefaultTenantID()` above
//
//	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)
//
//	formationTmplName := "app-only-formation-template-name"
//
//	certSubjcetMappingCN := "csm-async-callback-cn"
//	certSubjcetMappingCNSecond := "csm-async-callback-cn-second"
//	certSubjectMappingCustomSubject := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.TestExternalCertCN, certSubjcetMappingCN, -1)
//	certSubjectMappingCustomSubjectSecond := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.TestExternalCertCN, certSubjcetMappingCNSecond, -1)
//
//	// We need an externally issued cert with a custom subject that will be used to create a certificate subject mapping through the GraphQL API,
//	// which later will be loaded in-memory from the hydrator component
//	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
//		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
//		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
//		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceTestSecretName,
//		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
//		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
//		TestExternalCertSubject:               certSubjectMappingCustomSubject,
//		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
//		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
//		ExternalCertProvider:                  certprovider.CertificateService,
//	}
//
//	// We need only to create the secret so in the external-services-mock an HTTP client with certificate to be created and used to call the formation status API
//	_, _ = certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, false)
//
//	// The external cert secret created by the NewExternalCertFromConfig above is used by the external-services-mock for the async formation status API call,
//	// that's why in the function above there is a false parameter that don't delete it and an explicit defer deletion func is added here
//	// so, the secret could be deleted at the end of the test. Otherwise, it will remain as leftover resource in the cluster
//	defer func() {
//		k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
//		require.NoError(t, err)
//		k8s.DeleteSecret(t, ctx, k8sClient, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace)
//	}()
//
//	t.Log("Create integration system")
//	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, "int-system-app-to-app-notifications")
//	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, intSys)
//	require.NoError(t, err)
//	require.NotEmpty(t, intSys.ID)
//
//	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, intSys.ID)
//	require.NotEmpty(t, intSysAuth)
//	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)
//
//	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
//	require.True(t, ok)
//
//	t.Log("Issue a Hydra token with Client Credentials")
//	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
//	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)
//
//	namePlaceholder := "name"
//	displayNamePlaceholder := "display-name"
//	appRegion := "test-app-region"
//	appNamespace := "compass.test"
//	localTenantID := "local-tenant-id"
//
//	applicationType1 := "app-type-1"
//	t.Logf("Create application template for type: %q", applicationType1)
//	appTemplateInput := fixtures.FixApplicationTemplateWithCompositeLabelWithoutWebhook(applicationType1, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
//	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
//	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
//	require.NoError(t, err)
//	internalConsumerID := appTmpl.ID // add application templated ID as certificate subject mapping internal consumer to satisfy the authorization checks in the formation assignment status API
//
//	// Create certificate subject mapping with custom subject that was used to create a certificate for the graphql client above
//	certSubjectMappingCustomSubjectWithCommaSeparator := strings.ReplaceAll(strings.TrimLeft(certSubjectMappingCustomSubject, "/"), "/", ",")
//	csmInput := fixtures.FixCertificateSubjectMappingInput(certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, &internalConsumerID, tenantAccessLevels)
//	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, tenantAccessLevels)
//
//	var csmCreate graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
//	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
//	csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)
//
//	// Create second certificate subject mapping with custom subject that was used to test that trust details are send to the target
//	certSubjectMappingCustomSubjectWithCommaSeparatorSecond := strings.ReplaceAll(strings.TrimLeft(certSubjectMappingCustomSubjectSecond, "/"), "/", ",")
//	csmInputSecond := fixtures.FixCertificateSubjectMappingInput(certSubjectMappingCustomSubjectWithCommaSeparatorSecond, consumerType, &internalConsumerID, tenantAccessLevels)
//	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", certSubjectMappingCustomSubjectWithCommaSeparatorSecond, consumerType, tenantAccessLevels)
//
//	var csmCreateSecond graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
//	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreateSecond)
//	csmCreateSecond = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInputSecond)
//
//	t.Logf("Sleeping for %s, so the hydrator component could update the certificate subject mapping cache with the new data", conf.CertSubjectMappingResyncInterval.String())
//	time.Sleep(conf.CertSubjectMappingResyncInterval)
//
//	localTenantID2 := "local-tenant-id2"
//	applicationType2 := "app-type-2"
//	t.Logf("Create application template for type %q", applicationType2)
//	appTemplateInput = fixtures.FixApplicationTemplateWithCompositeLabelWithoutWebhook(applicationType2, localTenantID2, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
//	appTmpl2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
//
//	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl2)
//	require.NoError(t, err)
//
//	leadingProductIDs := []string{internalConsumerID} // internalConsumerID is used in the certificate subject mapping created above with certificate data that will be used in the external-services-mock when a formation status API is called
//
//	var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
//	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
//	ft = fixtures.CreateAppOnlyFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, formationTmplName, []string{applicationType1, applicationType2, exceptionSystemType}, leadingProductIDs, supportReset)
//
//	t.Logf("Create application 1 from template %q", applicationType1)
//	appFromTmplSrc := fixtures.FixApplicationFromTemplateInput(applicationType1, namePlaceholder, "app1-formation-notifications-tests", displayNamePlaceholder, "App 1 Display Name")
//	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
//	require.NoError(t, err)
//	createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
//	app1 := graphql.ApplicationExt{}
//	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, createAppFromTmplFirstRequest, &app1)
//	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &app1)
//	require.NoError(t, err)
//	require.NotEmpty(t, app1.ID)
//	t.Logf("app1 ID: %q", app1.ID)
//
//	t.Logf("Create application 2 from template %q", applicationType2)
//	appFromTmplSrc2 := fixtures.FixApplicationFromTemplateInput(applicationType2, namePlaceholder, "app2-formation-notifications-tests", displayNamePlaceholder, "App 2 Display Name")
//	appFromTmplSrc2GQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc2)
//	require.NoError(t, err)
//	createAppFromTmplSecondRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrc2GQL)
//	app2 := graphql.ApplicationExt{}
//	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, createAppFromTmplSecondRequest, &app2)
//	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &app2)
//	require.NoError(t, err)
//	require.NotEmpty(t, app2.ID)
//	t.Logf("app2 ID: %q", app2.ID)
//
//	t.Run("Synchronous App to App Formation Assignment Notificationss", func(t *testing.T) {
//		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
//		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
//
//		webhookType := graphql.WebhookTypeApplicationTenantMapping
//		webhookMode := graphql.WebhookModeSync
//		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
//		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
//		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"
//
//		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)
//
//		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, app1.ID)
//		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app1.ID)
//		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)
//
//		formationName := "app-to-app-formation-name"
//		t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
//		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
//		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)
//
//		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
//		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
//
//		t.Logf("Assign application 1 to formation %s", formationName)
//		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
//		assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
//		var assignedFormation graphql.Formation
//		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
//		require.NoError(t, err)
//		require.Equal(t, formationName, assignedFormation.Name)
//
//		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
//			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
//		}
//		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
//		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
//
//		t.Logf("Assign application 2 to formation %s", formationName)
//		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, tnt)
//		assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
//		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
//		require.NoError(t, err)
//		require.Equal(t, formationName, assignedFormation.Name)
//
//		expectedConfig := str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
//		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
//			app1.ID: {
//				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
//				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
//			},
//			app2.ID: {
//				app1.ID: fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
//				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
//			},
//		}
//		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
//		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
//
//		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
//		assertNotificationsCountForTenant(t, body, app1.ID, 1)
//
//		notificationsForApp1 := gjson.GetBytes(body, app1.ID)
//		assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
//		assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
//
//		t.Logf("Unassign Application 1 from formation %s", formationName)
//		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
//		var unassignFormation graphql.Formation
//		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
//		require.NoError(t, err)
//		require.Equal(t, formationName, unassignFormation.Name)
//
//		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
//			app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
//		}
//		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
//		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
//
//		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
//		assertNotificationsCountForTenant(t, body, app1.ID, 2)
//
//		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
//		unassignNotificationFound := false
//		for _, notification := range notificationsForApp1.Array() {
//			op := notification.Get("Operation").String()
//			if op == unassignOperation {
//				unassignNotificationFound = true
//				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
//			}
//		}
//		require.True(t, unassignNotificationFound, "notification for unassign app2 not found")
//
//		t.Logf("Assign application 1 to formation %s again", formationName)
//		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
//		assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
//		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
//		require.NoError(t, err)
//		require.Equal(t, formationName, assignedFormation.Name)
//
//		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
//			app1.ID: {
//				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
//				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
//			},
//			app2.ID: {
//				app1.ID: fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
//				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
//			},
//		}
//
//		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
//		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
//
//		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
//		assertNotificationsCountForTenant(t, body, app1.ID, 3)
//
//		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
//		assignNotificationsFound := 0
//		for _, notification := range notificationsForApp1.Array() {
//			op := notification.Get("Operation").String()
//			if op == assignOperation {
//				assignNotificationsFound++
//				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
//			}
//		}
//		require.Equal(t, 2, assignNotificationsFound, "two notifications for assign app2 expected")
//
//		t.Logf("Unassign Application 2 from formation %s", formationName)
//		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
//		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
//		require.NoError(t, err)
//		require.Equal(t, formationName, unassignFormation.Name)
//
//		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
//			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
//		}
//		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
//		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
//
//		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
//		assertNotificationsCountForTenant(t, body, app1.ID, 4)
//
//		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
//		unassignNotificationsFound := 0
//		for _, notification := range notificationsForApp1.Array() {
//			op := notification.Get("Operation").String()
//			if op == unassignOperation {
//				unassignNotificationsFound++
//				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
//			}
//		}
//		require.Equal(t, 2, unassignNotificationsFound, "two notifications for unassign app2 expected")
//
//		t.Logf("Unassign Application 1 from formation %s", formationName)
//		unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
//		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
//		require.NoError(t, err)
//		require.Equal(t, formationName, unassignFormation.Name)
//
//		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
//		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
//	})
//}
