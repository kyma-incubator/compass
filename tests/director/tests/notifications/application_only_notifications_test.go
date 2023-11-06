package notifications

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/kyma-incubator/compass/components/director/pkg/templatehelper"
)

func TestFormationNotificationsWithApplicationOnlyParticipants(t *testing.T) {
	ctx := context.Background()
	tnt := tenant.TestTenants.GetDefaultTenantID()
	tntParentCustomer := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant) // parent of `tenant.TestTenants.GetDefaultTenantID()` above

	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	formationTmplName := "e2e-tests-app-only-formation-template-name"

	certSubjcetMappingCN := "csm-async-callback-cn"
	certSubjcetMappingCNSecond := "csm-async-callback-cn-second"
	certSubjectMappingCustomSubject := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.TestExternalCertCN, certSubjcetMappingCN, -1)
	certSubjectMappingCustomSubjectSecond := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.TestExternalCertCN, certSubjcetMappingCNSecond, -1)

	// We need an externally issued cert with a custom subject that will be used to create a certificate subject mapping through the GraphQL API,
	// which later will be loaded in-memory from the hydrator component
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceTestSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               certSubjectMappingCustomSubject,
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.CertificateService,
	}

	// We need only to create the secret so in the external-services-mock an HTTP client with certificate to be created and used to call the formation status API
	_, _ = certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, false)

	// The external cert secret created by the NewExternalCertFromConfig above is used by the external-services-mock for the async formation status API call,
	// that's why in the function above there is a false parameter that don't delete it and an explicit defer deletion func is added here
	// so, the secret could be deleted at the end of the test. Otherwise, it will remain as leftover resource in the cluster
	defer func() {
		k8sClient, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
		require.NoError(t, err)
		k8s.DeleteSecret(t, ctx, k8sClient, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName, conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace)
	}()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, "int-system-app-to-app-notifications")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tnt, intSys.ID)
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

	applicationType1 := "e2e-tests-app-type-1"
	t.Logf("Create application template for type: %q", applicationType1)
	appTemplateInput := fixtures.FixApplicationTemplateWithCompositeLabelWithoutWebhook(applicationType1, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl)
	require.NoError(t, err)
	internalConsumerID := appTmpl.ID // add application templated ID as certificate subject mapping internal consumer to satisfy the authorization checks in the formation assignment status API

	// Create certificate subject mapping with custom subject that was used to create a certificate for the graphql client above
	certSubjectMappingCustomSubjectWithCommaSeparator := strings.ReplaceAll(strings.TrimLeft(certSubjectMappingCustomSubject, "/"), "/", ",")
	csmInput := fixtures.FixCertificateSubjectMappingInput(certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, &internalConsumerID, tenantAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", certSubjectMappingCustomSubjectWithCommaSeparator, consumerType, tenantAccessLevels)

	var csmCreate graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreate)
	csmCreate = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInput)

	// Create second certificate subject mapping with custom subject that was used to test that trust details are send to the target
	certSubjectMappingCustomSubjectWithCommaSeparatorSecond := strings.ReplaceAll(strings.TrimLeft(certSubjectMappingCustomSubjectSecond, "/"), "/", ",")
	csmInputSecond := fixtures.FixCertificateSubjectMappingInput(certSubjectMappingCustomSubjectWithCommaSeparatorSecond, consumerType, &internalConsumerID, tenantAccessLevels)
	t.Logf("Create certificate subject mapping with subject: %s, consumer type: %s and tenant access levels: %s", certSubjectMappingCustomSubjectWithCommaSeparatorSecond, consumerType, tenantAccessLevels)

	var csmCreateSecond graphql.CertificateSubjectMapping // needed so the 'defer' can be above the cert subject mapping creation
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreateSecond)
	csmCreateSecond = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInputSecond)

	t.Logf("Sleeping for %s, so the hydrator component could update the certificate subject mapping cache with the new data", conf.CertSubjectMappingResyncInterval.String())
	time.Sleep(conf.CertSubjectMappingResyncInterval)

	localTenantID2 := "local-tenant-id2"
	applicationType2 := "e2e-tests-app-type-2"
	t.Logf("Create application template for type %q", applicationType2)
	appTemplateInput = fixtures.FixApplicationTemplateWithCompositeLabelWithoutWebhook(applicationType2, localTenantID2, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
	appTmpl2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)

	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl2)
	require.NoError(t, err)

	leadingProductIDs := []string{internalConsumerID} // internalConsumerID is used in the certificate subject mapping created above with certificate data that will be used in the external-services-mock when a formation status API is called

	var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
	ft = fixtures.CreateAppOnlyFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, formationTmplName, []string{applicationType1, applicationType2, exceptionSystemType}, leadingProductIDs, supportReset)

	t.Logf("Create application 1 from template %q", applicationType1)
	appFromTmplSrc := fixtures.FixApplicationFromTemplateInput(applicationType1, namePlaceholder, "app1-formation-notifications-tests", displayNamePlaceholder, "App 1 Display Name")
	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	require.NoError(t, err)
	createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	app1 := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, createAppFromTmplFirstRequest, &app1)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &app1)
	require.NoError(t, err)
	require.NotEmpty(t, app1.ID)
	t.Logf("app1 ID: %q", app1.ID)

	t.Logf("Create application 2 from template %q", applicationType2)
	appFromTmplSrc2 := fixtures.FixApplicationFromTemplateInput(applicationType2, namePlaceholder, "app2-formation-notifications-tests", displayNamePlaceholder, "App 2 Display Name")
	appFromTmplSrc2GQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc2)
	require.NoError(t, err)
	createAppFromTmplSecondRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrc2GQL)
	app2 := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, createAppFromTmplSecondRequest, &app2)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &app2)
	require.NoError(t, err)
	require.NotEmpty(t, app2.ID)
	t.Logf("app2 ID: %q", app2.ID)

	// This test is adapted to the new formation in application_only_notifications_new_format_test.go
	// Leaving the old version commented for reference while adapting the other test cases

	//t.Run("Synchronous App to App Formation Assignment Notifications", func(t *testing.T) {
	//	cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	//	defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	//
	//	webhookType := graphql.WebhookTypeApplicationTenantMapping
	//	webhookMode := graphql.WebhookModeSync
	//	urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
	//	inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
	//	outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"
	//
	//	applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)
	//
	//	t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, app1.ID)
	//	actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app1.ID)
	//	defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)
	//
	//	formationName := "app-to-app-formation-name"
	//	t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
	//	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
	//	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)
	//
	//	assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
	//	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	//
	//	t.Logf("Assign application 1 to formation %s", formationName)
	//	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
	//	assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
	//	var assignedFormation graphql.Formation
	//	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
	//	require.NoError(t, err)
	//	require.Equal(t, formationName, assignedFormation.Name)
	//
	//	expectedAssignments := map[string]map[string]fixtures.AssignmentState{
	//		app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
	//	}
	//	assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
	//	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	//
	//	t.Logf("Assign application 2 to formation %s", formationName)
	//	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, tnt)
	//	assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
	//	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
	//	require.NoError(t, err)
	//	require.Equal(t, formationName, assignedFormation.Name)
	//
	//	expectedConfig := str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
	//	expectedAssignments = map[string]map[string]fixtures.AssignmentState{
	//		app1.ID: {
	//			app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
	//			app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
	//		},
	//		app2.ID: {
	//			app1.ID: fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
	//			app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
	//		},
	//	}
	//	assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
	//	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	//
	//	body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	//	assertNotificationsCountForTenant(t, body, app1.ID, 1)
	//
	//	notificationsForApp1 := gjson.GetBytes(body, app1.ID)
	//	assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
	//	assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
	//
	//	t.Logf("Unassign Application 1 from formation %s", formationName)
	//	unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
	//	var unassignFormation graphql.Formation
	//	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
	//	require.NoError(t, err)
	//	require.Equal(t, formationName, unassignFormation.Name)
	//
	//	expectedAssignments = map[string]map[string]fixtures.AssignmentState{
	//		app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
	//	}
	//	assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
	//	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	//
	//	body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	//	assertNotificationsCountForTenant(t, body, app1.ID, 2)
	//
	//	notificationsForApp1 = gjson.GetBytes(body, app1.ID)
	//	unassignNotificationFound := false
	//	for _, notification := range notificationsForApp1.Array() {
	//		op := notification.Get("Operation").String()
	//		if op == unassignOperation {
	//			unassignNotificationFound = true
	//			assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
	//		}
	//	}
	//	require.True(t, unassignNotificationFound, "notification for unassign app2 not found")
	//
	//	t.Logf("Assign application 1 to formation %s again", formationName)
	//	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
	//	assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
	//	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
	//	require.NoError(t, err)
	//	require.Equal(t, formationName, assignedFormation.Name)
	//
	//	expectedAssignments = map[string]map[string]fixtures.AssignmentState{
	//		app1.ID: {
	//			app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
	//			app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
	//		},
	//		app2.ID: {
	//			app1.ID: fixtures.AssignmentState{State: "READY", Config: expectedConfig, Value: expectedConfig, Error: nil},
	//			app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
	//		},
	//	}
	//
	//	assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
	//	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	//
	//	body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	//	assertNotificationsCountForTenant(t, body, app1.ID, 3)
	//
	//	notificationsForApp1 = gjson.GetBytes(body, app1.ID)
	//	assignNotificationsFound := 0
	//	for _, notification := range notificationsForApp1.Array() {
	//		op := notification.Get("Operation").String()
	//		if op == assignOperation {
	//			assignNotificationsFound++
	//			assertFormationAssignmentsNotificationWithItemsStructure(t, notification, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
	//		}
	//	}
	//	require.Equal(t, 2, assignNotificationsFound, "two notifications for assign app2 expected")
	//
	//	t.Logf("Unassign Application 2 from formation %s", formationName)
	//	unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
	//	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
	//	require.NoError(t, err)
	//	require.Equal(t, formationName, unassignFormation.Name)
	//
	//	expectedAssignments = map[string]map[string]fixtures.AssignmentState{
	//		app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
	//	}
	//	assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
	//	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	//
	//	body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	//	assertNotificationsCountForTenant(t, body, app1.ID, 4)
	//
	//	notificationsForApp1 = gjson.GetBytes(body, app1.ID)
	//	unassignNotificationsFound := 0
	//	for _, notification := range notificationsForApp1.Array() {
	//		op := notification.Get("Operation").String()
	//		if op == unassignOperation {
	//			unassignNotificationsFound++
	//			assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
	//		}
	//	}
	//	require.Equal(t, 2, unassignNotificationsFound, "two notifications for unassign app2 expected")
	//
	//	t.Logf("Unassign Application 1 from formation %s", formationName)
	//	unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
	//	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
	//	require.NoError(t, err)
	//	require.Equal(t, formationName, unassignFormation.Name)
	//
	//	assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
	//	assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	//})

	t.Run("Synchronous App to App Formation Assignment Notifications when state is in the response body", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		webhookType := graphql.WebhookTypeApplicationTenantMapping
		webhookMode := graphql.WebhookModeSync
		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/with-state/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, app1.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		// Create formation constraints for destination creator operator and attach them to a given formation template.
		// So we can verify the destination creator will not fail if in the configuration there is no destination information
		attachDestinationCreatorConstraints(t, ctx, ft, graphql.ResourceTypeApplication, graphql.ResourceTypeApplication)

		formationName := "app-to-app-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 1 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 2 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "CONFIG_PENDING", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 1)

		notificationsForApp1 := gjson.GetBytes(body, app1.ID)
		assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
		assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var unassignFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 2)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		unassignNotificationFound := false
		for _, notification := range notificationsForApp1.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				unassignNotificationFound = true
				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
			}
		}
		require.True(t, unassignNotificationFound, "notification for unassign app2 not found")

		t.Logf("Assign application 1 to formation %s again", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)
		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "CONFIG_PENDING", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}

		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 3)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		assignNotificationsFound := 0
		for _, notification := range notificationsForApp1.Array() {
			op := notification.Get("Operation").String()
			if op == assignOperation {
				assignNotificationsFound++
				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
			}
		}
		require.Equal(t, 2, assignNotificationsFound, "two notifications for assign app2 expected")

		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 4)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		unassignNotificationsFound := 0
		for _, notification := range notificationsForApp1.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				unassignNotificationsFound++
				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
			}
		}
		require.Equal(t, 2, unassignNotificationsFound, "two notifications for unassign app2 expected")

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	})

	t.Run("Asynchronous App to App Formation Assignment Notifications using the Default Tenant Mapping Handler", func(t *testing.T) {
		webhookType := graphql.WebhookTypeApplicationTenantMapping
		webhookMode := graphql.WebhookModeAsyncCallback
		urlTemplateAsync := "{\\\"path\\\":\\\"" + conf.CompassExternalMTLSGatewayURL + "/default-tenant-mapping-handler/v1/tenantMappings/{{.TargetApplication.ID}}\\\",\\\"method\\\":\\\"PATCH\\\"}"
		inputTemplate := "" // since the Default Tenant Mapping Handler does not take into account the request body because it always returns READY, there is no need to set an InputTemplate to the Webhook
		outputTemplate := "{\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"
		headerTemplate := "{\\\"Content-Type\\\": [\\\"application/json\\\"], \\\"Location\\\":[\\\"" + conf.CompassExternalMTLSGatewayURL + "/v1/businessIntegrations/{{.FormationID}}/assignments/{{.Assignment.ID}}/status\\\"]}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplateAsync, inputTemplate, outputTemplate)
		applicationWebhookInput.HeaderTemplate = &headerTemplate

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, app1.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		formationName := "e2e-test-app-to-app-formation"
		t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 1 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 2 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var unassignFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 1 to formation %s again", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}

		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 0, nil, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	})

	t.Run("Use Application Template Webhook if App does not have one for notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		webhookType := graphql.WebhookTypeApplicationTenantMapping
		webhookMode := graphql.WebhookModeSync
		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, app1.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		t.Logf("Add webhook with type %q and mode: %q to application template with ID %q", webhookType, webhookMode, appTmpl2.ID)

		actualApplicationTemplateWebhook := fixtures.AddWebhookToApplicationTemplate(t, ctx, oauthGraphQLClient, applicationWebhookInput, "", appTmpl2.ID)
		defer fixtures.CleanupWebhook(t, ctx, oauthGraphQLClient, "", actualApplicationTemplateWebhook.ID)

		// register a few more webhooks for the application template to verify that only the correct type of webhook is used when generation formation notifications
		actualUnregisterApplicationWebhook := fixtures.AddWebhookToApplicationTemplate(t, ctx, oauthGraphQLClient, fixtures.FixNonFormationNotificationWebhookInput(graphql.WebhookTypeUnregisterApplication), "", appTmpl2.ID)
		defer fixtures.CleanupWebhook(t, ctx, oauthGraphQLClient, "", actualUnregisterApplicationWebhook.ID)
		actualRegisterApplicationWebhook := fixtures.AddWebhookToApplicationTemplate(t, ctx, oauthGraphQLClient, fixtures.FixNonFormationNotificationWebhookInput(graphql.WebhookTypeRegisterApplication), "", appTmpl2.ID)
		defer fixtures.CleanupWebhook(t, ctx, oauthGraphQLClient, "", actualRegisterApplicationWebhook.ID)
		actualORDWebhook := fixtures.AddWebhookToApplicationTemplate(t, ctx, oauthGraphQLClient, fixtures.FixNonFormationNotificationWebhookInput(graphql.WebhookTypeOpenResourceDiscovery), "", appTmpl2.ID)
		defer fixtures.CleanupWebhook(t, ctx, oauthGraphQLClient, "", actualORDWebhook.ID)

		// Create formation constraints for destination creator operator and attach them to a given formation template.
		// So we can verify the destination creator will not fail if in the configuration there is no destination information
		attachDestinationCreatorConstraints(t, ctx, ft, graphql.ResourceTypeApplication, graphql.ResourceTypeApplication)

		formationName := "app-to-app-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 1 to formation %s", formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 2 to formation %s", formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
		fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 1)

		notificationsForApp1 := gjson.GetBytes(body, app1.ID)
		assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
		assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)

		assertNotificationsCountForTenant(t, body, app2.ID, 1)

		notificationsForApp2 := gjson.GetBytes(body, app2.ID)
		assignNotificationAboutApp1 := notificationsForApp2.Array()[0]
		assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutApp1, assignOperation, formation.ID, app1.ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer)
	})

	t.Run("Test only formation lifecycle synchronous notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		webhookType := graphql.WebhookTypeFormationLifecycle
		webhookMode := graphql.WebhookModeSync
		urlTemplateFormation := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/businessIntegration/{{.Formation.ID}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"createFormation\\\"}}POST{{else}}DELETE{{end}}\\\"}"
		inputTemplateFormation := "{\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"details\\\":{\\\"id\\\":\\\"{{.Formation.ID}}\\\",\\\"name\\\":\\\"{{.Formation.Name}}\\\"}}"
		outputTemplateFormation := "{\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}"
		formationTemplateWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplateFormation, inputTemplateFormation, outputTemplateFormation)

		t.Logf("Add webhook with type %q and mode: %q to formation template with ID %q", webhookType, webhookMode, ft.ID)
		actualFormationTemplateWebhook := fixtures.AddWebhookToFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateWebhookInput, "", ft.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, "", actualFormationTemplateWebhook.ID)

		formationName := "formation-name-from-template-with-webhook"
		t.Logf("Creating formation with name: %q from template with name: %q that has %q webhook", formationName, formationTmplName, graphql.WebhookTypeFormationLifecycle)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)
		require.NotEmpty(t, formation.ID)

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertFormationNotificationFromCreationOrDeletion(t, body, formation.ID, formation.Name, createFormationOperation, tnt, tntParentCustomer)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		delFormation := fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		require.NotEmpty(t, delFormation.ID)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForFormationID(t, body, formation.ID, 1)
		assertFormationNotificationFromCreationOrDeletion(t, body, formation.ID, formation.Name, deleteFormationOperation, tnt, tntParentCustomer)
	})

	t.Run("Formation lifecycle asynchronous notifications and asynchronous app to app formation assignment notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		applicationTntMappingWebhookType := graphql.WebhookTypeApplicationTenantMapping
		asyncCallbackWebhookMode := graphql.WebhookModeAsyncCallback
		urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateAsyncApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"source-trust-details\\\":[{{ Join  .SourceApplicationTemplate.TrustDetails.Subjects }}],\\\"target-trust-details\\\":[{{ Join  .TargetApplicationTemplate.TrustDetails.Subjects }}] }]}"
		outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, asyncCallbackWebhookMode, urlTemplateAsyncApplication, inputTemplateAsyncApplication, outputTemplateAsyncApplication)

		t.Logf("Add webhook with type %q and mode: %q to application template with ID %q", applicationTntMappingWebhookType, asyncCallbackWebhookMode, appTmpl.ID)
		actualApplicationTemplateWebhook := fixtures.AddWebhookToApplicationTemplate(t, ctx, oauthGraphQLClient, applicationWebhookInput, "", appTmpl.ID)
		defer fixtures.CleanupWebhook(t, ctx, oauthGraphQLClient, "", actualApplicationTemplateWebhook.ID)

		formationLifecycleWebhookType := graphql.WebhookTypeFormationLifecycle
		urlTemplateFormation := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/businessIntegration/async/{{.Formation.ID}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"createFormation\\\"}}POST{{else}}DELETE{{end}}\\\"}"
		inputTemplateFormation := "{\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"details\\\":{\\\"id\\\":\\\"{{.Formation.ID}}\\\",\\\"name\\\":\\\"{{.Formation.Name}}\\\"}}"
		outputTemplateFormation := "{\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

		formationTemplateWebhookInput := fixtures.FixFormationNotificationWebhookInput(formationLifecycleWebhookType, asyncCallbackWebhookMode, urlTemplateFormation, inputTemplateFormation, outputTemplateFormation)

		t.Logf("Add webhook with type %q and mode: %q to formation template with ID %q", formationLifecycleWebhookType, asyncCallbackWebhookMode, ft.ID)
		actualFormationTemplateWebhook := fixtures.AddWebhookToFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateWebhookInput, "", ft.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, "", actualFormationTemplateWebhook.ID)

		formationName := "formation-name-from-template-with-webhook"
		t.Logf("Creating formation with name: %q from template with name: %q that has %q webhook", formationName, formationTmplName, formationLifecycleWebhookType)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, "", actualFormationTemplateWebhook.ID) // Otherwise, FT wouldn't be able to be deleted because formation is stuck in DELETING state
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)
		require.Equal(t, "INITIAL", formation.State)
		require.Empty(t, formation.Error)

		// Assign both applications when the formation is still in INITIAL state and validate no notifications are sent and formation assignments are in INITIAL state
		t.Logf("Assign application 1 to formation: %q", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		t.Logf("Assign application 2 to formation: %q", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		// As part of the formation status API request, formation assignment synchronization will be executed.
		assertAsyncFormationNotificationFromCreationOrDeletion(t, ctx, body, formation.ID, formation.Name, "READY", createFormationOperation, tnt, tntParentCustomer)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, time.Second*8,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 1)

		notificationsForApp1 := gjson.GetBytes(body, app1.ID)
		assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
		assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)

		// Check that there are trust details for the target and there are no trust details for the source.
		assertTrustDetailsForTargetAndNoTrustDetailsForSource(t, assignNotificationAboutApp2, certSubjectMappingCustomSubjectWithCommaSeparator, certSubjectMappingCustomSubjectWithCommaSeparatorSecond)

		t.Logf("Unassign Application 1 from formation: %q", formationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var unassignFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 2)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		unassignNotificationFound := false
		for _, notification := range notificationsForApp1.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				unassignNotificationFound = true
				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
			}
		}
		require.True(t, unassignNotificationFound, "notification for unassign app2 not found")

		// Check that there are trust details for the target and there are no trust details for the source.
		assignNotificationAboutApp2 = notificationsForApp1.Array()[0]
		assertTrustDetailsForTargetAndNoTrustDetailsForSource(t, assignNotificationAboutApp2, certSubjectMappingCustomSubjectWithCommaSeparator, certSubjectMappingCustomSubjectWithCommaSeparatorSecond)

		t.Logf("Assign application 1 to formation: %q again", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}

		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 3)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		assignNotificationsFound := 0
		for _, notification := range notificationsForApp1.Array() {
			op := notification.Get("Operation").String()
			if op == assignOperation {
				assignNotificationsFound++
				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
			}
		}
		require.Equal(t, 2, assignNotificationsFound, "two notifications for assign app2 expected")

		// Check that there are trust details for the target and there are no trust details for the source.
		assignNotificationAboutApp2 = notificationsForApp1.Array()[0]
		assertTrustDetailsForTargetAndNoTrustDetailsForSource(t, assignNotificationAboutApp2, certSubjectMappingCustomSubjectWithCommaSeparator, certSubjectMappingCustomSubjectWithCommaSeparatorSecond)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 4)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		unassignNotificationsFound := 0
		for _, notification := range notificationsForApp1.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				unassignNotificationsFound++
				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
			}
		}
		require.Equal(t, 2, unassignNotificationsFound, "two notifications for unassign app2 expected")

		// Check that there are trust details for the target and there are no trust details for the source.
		assignNotificationAboutApp2 = notificationsForApp1.Array()[0]
		assertTrustDetailsForTargetAndNoTrustDetailsForSource(t, assignNotificationAboutApp2, certSubjectMappingCustomSubjectWithCommaSeparator, certSubjectMappingCustomSubjectWithCommaSeparatorSecond)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 0, nil, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		t.Logf("Deleting formation with name: %q from template with name: %q that has %q webhook", formationName, formationTmplName, formationLifecycleWebhookType)
		delFormation := fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		require.NotEmpty(t, delFormation.ID)
		require.Equal(t, "DELETING", delFormation.State)
		require.Empty(t, delFormation.Error)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForFormationID(t, body, formation.ID, 1)
		assertAsyncFormationNotificationFromCreationOrDeletion(t, ctx, body, formation.ID, formation.Name, "READY", deleteFormationOperation, tnt, tntParentCustomer)
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Verify the formation with name: %q is successfully deleted after READY status is reported on the status API...", formationName)
		formationPage := fixtures.ListFormationsWithinTenant(t, ctx, tnt, certSecuredGraphQLClient)
		require.Equal(t, 0, formationPage.TotalCount)
		require.Empty(t, formationPage.Data)
		t.Logf("Formation with name: %q is successfully deleted after READY status is reported on the status API", formationName)
	})

	t.Run("Resynchronize synchronous formation notifications with tenant mapping notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

		formationLifecycleWebhookType := graphql.WebhookTypeFormationLifecycle
		syncWebhookMode := graphql.WebhookModeSync
		urlTemplateFormation := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/businessIntegration/fail-once/{{.Formation.ID}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"createFormation\\\"}}POST{{else}}DELETE{{end}}\\\"}"
		inputTemplateFormation := "{\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"details\\\":{\\\"id\\\":\\\"{{.Formation.ID}}\\\",\\\"name\\\":\\\"{{.Formation.Name}}\\\"}}"
		outputTemplateFormation := "{\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}"
		formationTemplateWebhookInput := fixtures.FixFormationNotificationWebhookInput(formationLifecycleWebhookType, syncWebhookMode, urlTemplateFormation, inputTemplateFormation, outputTemplateFormation)

		t.Logf("Add webhook with type %q and mode: %q to formation template with ID: %q", formationLifecycleWebhookType, syncWebhookMode, ft.ID)
		actualFormationTemplateWebhook := fixtures.AddWebhookToFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateWebhookInput, "", ft.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, "", actualFormationTemplateWebhook.ID)

		urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateAsyncApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

		applicationTntMappingWebhookType := graphql.WebhookTypeApplicationTenantMapping
		asyncCallbacWebhookMode := graphql.WebhookModeAsyncCallback
		applicationAsyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, asyncCallbacWebhookMode, urlTemplateAsyncApplication, inputTemplateAsyncApplication, outputTemplateAsyncApplication)

		t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", applicationTntMappingWebhookType, asyncCallbacWebhookMode, app1.ID)
		actualApplicationAsyncWebhookInput := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationAsyncWebhookInput, tnt, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID)

		urlTemplateSyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateSyncApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplateSyncApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, syncWebhookMode, urlTemplateSyncApplication, inputTemplateSyncApplication, outputTemplateSyncApplication)

		t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", applicationTntMappingWebhookType, syncWebhookMode, app2.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app2.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		formationName := "formation-name-from-template-with-webhook"
		t.Logf("Creating formation with name: %q from template with name: %q that has %q webhook", formationName, formationTmplName, formationLifecycleWebhookType)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)
		require.NotEmpty(t, formation.ID)
		require.Equal(t, "CREATE_ERROR", formation.State)

		t.Logf("Assign application 1 to formation: %q", formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
		assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
		require.Equal(t, formation.ID, assignedFormation.ID)
		require.Equal(t, formation.State, assignedFormation.State)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{
			Condition: graphql.FormationStatusConditionError,
			Errors: []*graphql.FormationStatusError{{
				Message:   "failed to parse request",
				ErrorCode: 2,
			}},
		})

		t.Logf("Assign application 2 to formation: %q", formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
		assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
		require.Equal(t, formation.ID, assignedFormation.ID)
		require.Equal(t, formation.State, assignedFormation.State)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{
			Condition: graphql.FormationStatusConditionError,
			Errors: []*graphql.FormationStatusError{{
				Message:   "failed to parse request",
				ErrorCode: 2,
			}},
		})

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertFormationNotificationFromCreationOrDeletion(t, body, formation.ID, formation.Name, createFormationOperation, tnt, tntParentCustomer)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Resynchronize formation %q should retry and succeed", formation.Name)
		resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
		example.SaveExample(t, resynchronizeReq.Query(), "resynchronize formation notifications")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &formation)
		require.NoError(t, err)
		require.Equal(t, formationName, formation.Name)
		require.Equal(t, "READY", formation.State)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		var unassignFormation graphql.Formation
		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

		delFormation := fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		require.NotEmpty(t, delFormation.ID)
		require.Equal(t, "DELETE_ERROR", delFormation.State)

		t.Logf("Should get formation with name: %q by ID: %q", formationName, formation.ID)
		var gotFormation *graphql.Formation
		getFormationReq := fixtures.FixGetFormationRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, getFormationReq, &gotFormation)
		require.NoError(t, err)
		require.Equal(t, delFormation.ID, gotFormation.ID)
		require.Equal(t, delFormation.State, gotFormation.State)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Resynchronize formation %q should retry and succeed", formationName)
		resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &delFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, delFormation.Name)
		require.Equal(t, "READY", delFormation.State)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForFormationID(t, body, formation.ID, 1)
		assertFormationNotificationFromCreationOrDeletion(t, body, formation.ID, formation.Name, deleteFormationOperation, tnt, tntParentCustomer)

		t.Logf("Should fail while getting formation with name: %q by ID: %q because it is already deleted", formation.Name, formation.ID)
		var nonexistentFormation *graphql.Formation
		getNonexistentFormationReq := fixtures.FixGetFormationRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, getNonexistentFormationReq, nonexistentFormation)
		require.Error(t, err)
		require.Nil(t, nonexistentFormation)
	})

	t.Run("Resynchronize tenant mapping notifications with reset", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

		urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateAsyncApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

		applicationAsyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, urlTemplateAsyncApplication, inputTemplateAsyncApplication, outputTemplateAsyncApplication)

		t.Logf("Add webhook with application with ID %q", app1.ID)
		actualApplicationAsyncWebhookInput := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationAsyncWebhookInput, tnt, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID)

		urlTemplateApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, urlTemplateApplication, inputTemplateApplication, outputTemplateApplication)

		t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app2.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		formationName := "formation-name-from-template-with-webhook"
		t.Logf("Creating formation with name: %q from template with name: %q that has %q webhook", formationName, formationTmplName, graphql.WebhookTypeFormationLifecycle)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)
		require.NotEmpty(t, formation.ID)
		require.Equal(t, "READY", formation.State)

		t.Logf("Assign application 1 to formation: %q", formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
		assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
		require.Equal(t, formation.ID, assignedFormation.ID)
		require.Equal(t, formation.State, assignedFormation.State)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{
			Condition: graphql.FormationStatusConditionReady,
			Errors:    nil,
		})

		t.Logf("Assign application 2 to formation: %q", formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
		assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
		require.Equal(t, formation.ID, assignedFormation.ID)
		require.Equal(t, formation.State, assignedFormation.State)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Resynchronize formation %q without reset should not do anything", formation.Name)
		resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
		example.SaveExample(t, resynchronizeReq.Query(), "resynchronize formation notifications")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &formation)
		require.NoError(t, err)
		require.Equal(t, formationName, formation.Name)
		require.Equal(t, "READY", formation.State)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Resynchronize formation %q with reset should set assignments to initial state", formation.Name)
		resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequestWithResetOption(formation.ID, true)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &formation)
		require.NoError(t, err)
		require.Equal(t, formationName, formation.Name)
		require.Equal(t, "READY", formation.State)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Unassign Application 1 from formation %s", formationName)
		fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

		delFormation := fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		require.NotEmpty(t, delFormation.ID)
		require.Equal(t, "READY", delFormation.State)
	})

	t.Run("Resynchronize asynchronous formation notifications with tenant mapping notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

		formationTemplateWebhookType := graphql.WebhookTypeFormationLifecycle
		formationTemplateWebhookMode := graphql.WebhookModeAsyncCallback
		urlTemplateThatNeverResponds := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/businessIntegration/async-no-response/{{.Formation.ID}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"createFormation\\\"}}POST{{else}}DELETE{{end}}\\\"}"
		inputTemplateFormation := "{\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"details\\\":{\\\"id\\\":\\\"{{.Formation.ID}}\\\",\\\"name\\\":\\\"{{.Formation.Name}}\\\"}}"
		outputTemplateFormation := "{\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

		formationTemplateWebhookInput := fixtures.FixFormationNotificationWebhookInput(formationTemplateWebhookType, formationTemplateWebhookMode, urlTemplateThatNeverResponds, inputTemplateFormation, outputTemplateFormation)

		t.Logf("Add webhook with type %q and mode: %q to formation template with ID: %q", formationTemplateWebhookType, formationTemplateWebhookMode, ft.ID)
		actualFormationTemplateWebhook := fixtures.AddWebhookToFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateWebhookInput, "", ft.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, "", actualFormationTemplateWebhook.ID)

		urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateAsyncApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

		applicationAsyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, urlTemplateAsyncApplication, inputTemplateAsyncApplication, outputTemplateAsyncApplication)

		t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1.ID)
		actualApplicationAsyncWebhookInput := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationAsyncWebhookInput, tnt, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID)

		urlTemplateApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, urlTemplateApplication, inputTemplateApplication, outputTemplateApplication)

		t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app2.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		formationName := "formation-name-from-template-with-webhook"
		t.Logf("Creating formation with name: %q from template with name: %q that has %q webhook", formationName, formationTmplName, graphql.WebhookTypeFormationLifecycle)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)
		require.NotEmpty(t, formation.ID)
		require.Equal(t, "INITIAL", formation.State)
		require.Empty(t, formation.Error)

		t.Logf("Assign application 1 to formation: %q", formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
		assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
		require.Equal(t, formation.ID, assignedFormation.ID)
		require.Equal(t, formation.State, assignedFormation.State)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})

		t.Logf("Assign application 2 to formation: %q", formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
		assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
		require.Equal(t, formation.ID, assignedFormation.ID)
		require.Equal(t, formation.State, assignedFormation.State)

		assertNoNotificationsAreSentForTenant(t, certSecuredHTTPClient, app1.ID)
		assertNoNotificationsAreSentForTenant(t, certSecuredHTTPClient, app2.ID)

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Should get formation with name: %q by ID: %q", formationName, formation.ID)
		var gotFormation *graphql.Formation
		getFormationReq := fixtures.FixGetFormationRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, getFormationReq, &gotFormation)
		require.NoError(t, err)
		require.Equal(t, formation.ID, gotFormation.ID)
		require.Equal(t, "INITIAL", gotFormation.State)

		assertAsyncFormationNotificationFromCreationOrDeletion(t, ctx, body, formation.ID, formation.Name, "INITIAL", createFormationOperation, tnt, tntParentCustomer)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "INITIAL", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})

		urlTemplateThatFailsOnce := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/businessIntegration/async-fail-once/{{.Formation.ID}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"createFormation\\\"}}POST{{else}}DELETE{{end}}\\\"}"
		webhookThatFailsOnceInput := fixtures.FixFormationNotificationWebhookInput(formationTemplateWebhookType, formationTemplateWebhookMode, urlTemplateThatFailsOnce, inputTemplateFormation, outputTemplateFormation)

		t.Logf("Update webhook with type %q and mode: %q for formation template with ID: %q", formationTemplateWebhookType, formationTemplateWebhookMode, ft.ID)
		updatedFormationTemplateWebhook := fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, "", actualFormationTemplateWebhook.ID, webhookThatFailsOnceInput)
		require.Equal(t, updatedFormationTemplateWebhook.ID, actualFormationTemplateWebhook.ID)

		t.Logf("Resynchronize formation %q should retry and fail", formation.Name)
		resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &formation)
		require.NoError(t, err)
		require.Equal(t, formationName, formation.Name)
		require.Equal(t, "INITIAL", formation.State)
		require.Empty(t, formation.Error)

		// As part of the formation status API request, formation assignment synchronization will be executed.
		assertAsyncFormationNotificationFromCreationOrDeletion(t, ctx, body, formation.ID, formation.Name, "CREATE_ERROR", createFormationOperation, tnt, tntParentCustomer)
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{
			Condition: graphql.FormationStatusConditionError,
			Errors: []*graphql.FormationStatusError{{
				Message:   "failed to parse request",
				ErrorCode: 2,
			}},
		})

		t.Logf("Should get formation with name: %q by ID: %q", formationName, formation.ID)
		getFormationReq = fixtures.FixGetFormationRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, getFormationReq, &gotFormation)
		require.NoError(t, err)
		require.Equal(t, formation.ID, gotFormation.ID)
		require.Equal(t, "CREATE_ERROR", gotFormation.State)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Resynchronize formation %q should retry and succeed", formation.Name)
		resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &formation)
		require.NoError(t, err)
		require.Equal(t, formationName, formation.Name)
		require.Equal(t, "INITIAL", formation.State)
		require.Empty(t, formation.Error)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}

		assertAsyncFormationNotificationFromCreationOrDeletion(t, ctx, body, formation.ID, formation.Name, "READY", createFormationOperation, tnt, tntParentCustomer)
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Should get formation with name: %q by ID: %q", formationName, formation.ID)
		getFormationReq = fixtures.FixGetFormationRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, getFormationReq, &gotFormation)
		require.NoError(t, err)
		require.Equal(t, formation.ID, gotFormation.ID)
		require.Equal(t, "READY", gotFormation.State)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		var unassignFormation graphql.Formation
		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Update webhook with type %q and mode: %q to formation template with ID: %q", formationTemplateWebhookType, formationTemplateWebhookMode, ft.ID)
		updatedFormationTemplateWebhook = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, "", actualFormationTemplateWebhook.ID, formationTemplateWebhookInput)
		require.Equal(t, updatedFormationTemplateWebhook.ID, actualFormationTemplateWebhook.ID)

		delFormation := fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		require.NotEmpty(t, delFormation.ID)
		require.Equal(t, "DELETING", delFormation.State)
		require.Empty(t, delFormation.Error)

		t.Logf("Should get formation with name: %q by ID: %q", formationName, formation.ID)
		getFormationReq = fixtures.FixGetFormationRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, getFormationReq, &gotFormation)
		require.NoError(t, err)
		require.Equal(t, formation.ID, gotFormation.ID)
		require.Equal(t, delFormation.State, gotFormation.State)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForFormationID(t, body, formation.ID, 1)
		assertAsyncFormationNotificationFromCreationOrDeletionWithShouldExpectDeleted(t, ctx, body, formation.ID, formation.Name, "DELETING", deleteFormationOperation, tnt, tntParentCustomer, false)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Update webhook with type %q and mode: %q to formation template with ID: %q", formationTemplateWebhookType, formationTemplateWebhookMode, ft.ID)
		updatedFormationTemplateWebhook = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, "", actualFormationTemplateWebhook.ID, webhookThatFailsOnceInput)
		require.Equal(t, updatedFormationTemplateWebhook.ID, actualFormationTemplateWebhook.ID)

		t.Logf("Resynchronize formation %s should retry and fail", formation.Name)
		resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &formation)
		require.NoError(t, err)
		require.Equal(t, formationName, formation.Name)
		require.Equal(t, "DELETING", formation.State)
		require.Empty(t, formation.Error)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForFormationID(t, body, formation.ID, 1)
		assertAsyncFormationNotificationFromCreationOrDeletion(t, ctx, body, formation.ID, formation.Name, "DELETE_ERROR", deleteFormationOperation, tnt, tntParentCustomer)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Resynchronize formation %s should retry and succeed", formationName)
		resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &delFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, delFormation.Name)
		require.Equal(t, "DELETING", delFormation.State)
		require.Empty(t, delFormation.Error)

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForFormationID(t, body, formation.ID, 1)
		assertAsyncFormationNotificationFromCreationOrDeletion(t, ctx, body, formation.ID, formation.Name, delFormation.State, deleteFormationOperation, tnt, tntParentCustomer)

		t.Logf("Should fail while getting formation with name: %q by ID: %q because it is already deleted", formation.Name, formation.ID)
		var nonexistentFormation *graphql.Formation
		getNonexistentFormationReq := fixtures.FixGetFormationRequest(formation.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, getNonexistentFormationReq, nonexistentFormation)
		require.Error(t, err)
		require.Nil(t, nonexistentFormation)
	})

	t.Run("App to App Notifications are skipped if DoNotGenerateFormationAssignmentNotification constraints is attached to the formation type", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		webhookType := graphql.WebhookTypeApplicationTenantMapping
		webhookMode := graphql.WebhookModeSync
		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, app1.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		in := graphql.FormationConstraintInput{
			Name:            "TestDoNotGenerateFormationAssignmentNotifications",
			ConstraintType:  graphql.ConstraintTypePre,
			TargetOperation: graphql.TargetOperationGenerateFormationAssignmentNotification,
			Operator:        formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationOperator,
			ResourceType:    graphql.ResourceTypeApplication,
			ResourceSubtype: applicationType1,
			InputTemplate:   fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"source_resource_type\\\": \\\"{{if .SourceApplication}}APPLICATION{{else if .RuntimeContext}}RUNTIME_CONTEXT{{else}}RUNTIME{{end}}\\\",\\\"source_resource_id\\\": \\\"{{if .SourceApplication}}{{.SourceApplication.ID}}{{else if .RuntimeContext}}{{.RuntimeContext.ID}}{{else}}{{.Runtime.ID}}{{end}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\",\\\"formation_template_id\\\":\\\"{{.FormationTemplateID}}\\\",\\\"except_subtypes\\\": [\\\"%s\\\"]}", exceptionSystemType),
			ConstraintScope: graphql.ConstraintScopeFormationType,
		}
		constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)
		defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
		require.NotEmpty(t, constraint.ID)

		defer fixtures.DetachConstraintFromFormationTemplateNoCheckError(ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)
		fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, constraint.Name, ft.ID, ft.Name)

		formationName := "app-to-app-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 1 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 2 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 0)

		localTenantID3 := "local-tenant-id3"
		t.Logf("Create application template for type %q", exceptionSystemType)
		appTemplateInput = fixtures.FixApplicationTemplateWithCompositeLabelWithoutWebhook(exceptionSystemType, localTenantID3, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
		appTmpl3, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, "", appTemplateInput)

		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, "", appTmpl3)
		require.NoError(t, err)

		t.Log("Create third application with exception type")
		appFromTmplSrc := fixtures.FixApplicationFromTemplateInput(exceptionSystemType, namePlaceholder, "exception-system-formation-notifications-tests", displayNamePlaceholder, "Exception App Display Name")
		appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
		exceptionTypeApp := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, createAppFromTmplRequest, &exceptionTypeApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tnt, &exceptionTypeApp)
		require.NoError(t, err)
		require.NotEmpty(t, exceptionTypeApp.ID)
		t.Logf("Successfully created exception type application with ID %q", exceptionTypeApp.ID)

		t.Logf("Assign application 3 (exception one) to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, exceptionTypeApp.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(exceptionTypeApp.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID:             fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID:             fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				exceptionTypeApp.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID:             fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID:             fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				exceptionTypeApp.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			exceptionTypeApp.ID: {
				app1.ID:             fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				app2.ID:             fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				exceptionTypeApp.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}

		assertFormationAssignments(t, ctx, tnt, formation.ID, 9, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 1)

		notificationsForApp1 := gjson.GetBytes(body, app1.ID)
		assignNotificationAboutExceptionTypeApp := notificationsForApp1.Array()[0]
		assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutExceptionTypeApp, assignOperation, formation.ID, exceptionTypeApp.ID, localTenantID3, appNamespace, appRegion, tnt, tntParentCustomer)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var unassignFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {
				app2.ID:             fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				exceptionTypeApp.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			exceptionTypeApp.ID: {
				app2.ID:             fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				exceptionTypeApp.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 2)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		unassignNotificationFound := false
		for _, notification := range notificationsForApp1.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				unassignNotificationFound = true
				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, exceptionTypeApp.ID, localTenantID3, appNamespace, appRegion, tnt, tntParentCustomer)
			}
		}
		require.True(t, unassignNotificationFound, "notification for unassign exceptionTypeApp not found")

		t.Logf("Unassign application 3 (exception one) from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(exceptionTypeApp.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Detaching constraint from formation template")
		fixtures.DetachConstraintFromFormationTemplate(t, ctx, certSecuredGraphQLClient, constraint.ID, ft.ID)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Assign application 1 to formation %s again", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
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

		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 1)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		assignNotificationAboutExceptionTypeApp = notificationsForApp1.Array()[0]
		assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutExceptionTypeApp, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 2)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		unassignNotificationFound = false
		for _, notification := range notificationsForApp1.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				unassignNotificationFound = true
				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
			}
		}
		require.True(t, unassignNotificationFound, "notification for unassign app2 not found")

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	})

	t.Run("App to App Notifications are skipped if DoNotGenerateFormationAssignmentNotification constraints is globally attached", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		webhookType := graphql.WebhookTypeApplicationTenantMapping
		webhookMode := graphql.WebhookModeSync
		urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
		outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

		applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(webhookType, webhookMode, urlTemplate, inputTemplate, outputTemplate)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", webhookType, webhookMode, app1.ID)
		actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

		in := graphql.FormationConstraintInput{
			Name:            "TestDoNotGenerateFormationAssignmentNotifications",
			ConstraintType:  graphql.ConstraintTypePre,
			TargetOperation: graphql.TargetOperationGenerateFormationAssignmentNotification,
			Operator:        formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationOperator,
			ResourceType:    graphql.ResourceTypeApplication,
			ResourceSubtype: applicationType1,
			InputTemplate:   fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"source_resource_type\\\": \\\"{{if .SourceApplication}}APPLICATION{{else if .RuntimeContext}}RUNTIME_CONTEXT{{else}}RUNTIME{{end}}\\\",\\\"source_resource_id\\\": \\\"{{if .SourceApplication}}{{.SourceApplication.ID}}{{else if .RuntimeContext}}{{.RuntimeContext.ID}}{{else}}{{.Runtime.ID}}{{end}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\",\\\"formation_template_id\\\":\\\"{{.FormationTemplateID}}\\\",\\\"except_formation_types\\\": [\\\"%s\\\"]}", formationTmplName),
			ConstraintScope: graphql.ConstraintScopeGlobal,
		}
		constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)
		defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, constraint.ID)
		require.NotEmpty(t, constraint.ID)

		formationName := "app-to-app-formation-name"
		formationInput := graphql.FormationInput{Name: formationName}
		t.Logf("Creating formation with name: %q from the template with name: %q that is in the constraints exceptions", formationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 1 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, formationInput, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments := map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 2 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, formationInput, app2.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
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
		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 1)

		notificationsForApp1 := gjson.GetBytes(body, app1.ID)
		assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
		assertFormationAssignmentsNotificationWithItemsStructure(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var unassignFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 2)

		notificationsForApp1 = gjson.GetBytes(body, app1.ID)
		unassignNotificationFound := false
		for _, notification := range notificationsForApp1.Array() {
			op := notification.Get("Operation").String()
			if op == unassignOperation {
				unassignNotificationFound = true
				assertFormationAssignmentsNotificationWithItemsStructure(t, notification, unassignOperation, formation.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
			}
		}
		require.True(t, unassignNotificationFound, "notification for unassign app2 not found")

		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		exceptionFormationTemplateName := "exception-ft"
		t.Logf("Create another template with name: %q that is not in the constraints exceptions", exceptionFormationTemplateName)
		var ft graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
		defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &ft)
		ft = fixtures.CreateAppOnlyFormationTemplateWithoutInput(t, ctx, certSecuredGraphQLClient, exceptionFormationTemplateName, []string{applicationType1, applicationType2, exceptionSystemType}, leadingProductIDs, doesNotSupportReset)

		formationName = "app-to-app-formation-name-from-exception-formation-type"
		formationInput = graphql.FormationInput{Name: formationName}
		t.Logf("Creating formation with name: %q from the template with name: %q that is in the constraints exceptions", formationName, exceptionFormationTemplateName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation = fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &exceptionFormationTemplateName)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 1 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, formationInput, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 2 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, formationInput, app2.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 4, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 0)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignments = map[string]map[string]fixtures.AssignmentState{
			app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, app1.ID, 0)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
	})

	t.Run("Formation Assignment Notifications with status reset API", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		formationName := "app-to-app-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		t.Logf("Assign application 1 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq := fixtures.FixAssignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var assignedFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		t.Logf("Assign application 2 to formation %s", formationName)
		defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, graphql.FormationObjectTypeApplication, tnt)
		assignReq = fixtures.FixAssignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, assignReq, &assignedFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, assignedFormation.Name)

		t.Run("Synchronous App to App reset", func(t *testing.T) {
			applicationTntMappingWebhookType := graphql.WebhookTypeApplicationTenantMapping
			syncWebhookMode := graphql.WebhookModeSync
			urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, syncWebhookMode, urlTemplate, inputTemplate, outputTemplate)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", applicationTntMappingWebhookType, syncWebhookMode, app1.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", applicationTntMappingWebhookType, syncWebhookMode, app2.ID)
			actualApplicationWebhookApp2 := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app2.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhookApp2.ID)
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			expectedResetConfig := "{\"resetKey\":\"resetValue\",\"resetKey2\":{\"resetKey\":\"resetValue2\"}}"
			assignmentsPage := assertFormationAssignmentsCount(t, ctx, formation.ID, tnt, 4)
			formationAssignmentID := getFormationAssignmentIDBySourceAndTarget(t, assignmentsPage, app1.ID, app2.ID)
			reverseAssignmentID := getFormationAssignmentIDBySourceAndTarget(t, assignmentsPage, app2.ID, app1.ID)

			t.Logf("Calling FA status reset for formation assignment with source %q and target %q", app1.ID, app2.ID)
			executeFAStatusResetReqWithExpectedStatusCode(t, certSecuredHTTPClient, "CONFIG_PENDING", expectedResetConfig, tnt, formation.ID, formationAssignmentID, http.StatusOK)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			// We use the async method, because the status API is being called.
			// Even though the case is synchronous, the notification is executed in a separate goroutine and isn't guaranteed
			// to have been executed before we perform the checks otherwise.
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 1)
			notificationsForApp1Tenant := gjson.GetBytes(body, app1.ID)
			assignNotificationForApp1 := notificationsForApp1Tenant.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp1, assignOperation, formation.ID, app1.ID, "", appRegion, expectedResetConfig, tnt, tntParentCustomer)
			require.NoError(t, err)

			assertNotificationsCountForTenant(t, body, app2.ID, 1)
			notificationsForApp2Tenant := gjson.GetBytes(body, app2.ID)
			assignNotificationForApp2 := notificationsForApp2Tenant.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp2, assignOperation, formation.ID, app2.ID, "", appRegion, *fixtures.StatusAPISyncConfigJSON, tnt, tntParentCustomer)
			require.NoError(t, err)

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Calling FA status reset for formation assignment with source %q and target %q", app2.ID, app1.ID)
			executeFAStatusResetReqWithExpectedStatusCode(t, certSecuredHTTPClient, "CONFIG_PENDING", expectedResetConfig, tnt, formation.ID, reverseAssignmentID, http.StatusOK)
			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}

			// We use the async method, because the status API is being called.
			// Even though the case is synchronous, the notification is executed in a separate goroutine and isn't guaranteed
			// to have been executed before we perform the checks otherwise.
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 1)
			assignNotificationForApp1 = gjson.GetBytes(body, app1.ID)
			assignNotificationForApp1 = assignNotificationForApp1.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp1, assignOperation, formation.ID, app1.ID, "", appRegion, *fixtures.StatusAPISyncConfigJSON, tnt, tntParentCustomer)
			require.NoError(t, err)

			assertNotificationsCountForTenant(t, body, app2.ID, 1)
			notificationsForApp2Tenant = gjson.GetBytes(body, app2.ID)
			assignNotificationForApp2 = notificationsForApp2Tenant.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp2, assignOperation, formation.ID, app2.ID, "", appRegion, expectedResetConfig, tnt, tntParentCustomer)
			require.NoError(t, err)
		})

		t.Run("Asynchronous App to App reset", func(t *testing.T) {
			applicationTntMappingWebhookType := graphql.WebhookTypeApplicationTenantMapping
			asyncWebhookMode := graphql.WebhookModeAsyncCallback
			urlTemplateAsync := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateAsync := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplateAsync := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

			applicationWebhookInputAsync := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, asyncWebhookMode, urlTemplateAsync, inputTemplateAsync, outputTemplateAsync)

			syncWebhookMode := graphql.WebhookModeSync
			urlTemplateSync := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateSync := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplateSync := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookInputSync := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, syncWebhookMode, urlTemplateSync, inputTemplateSync, outputTemplateSync)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", applicationTntMappingWebhookType, asyncWebhookMode, app1.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInputAsync, tnt, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", applicationTntMappingWebhookType, syncWebhookMode, app2.ID)
			actualApplicationWebhookApp2 := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInputSync, tnt, app2.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhookApp2.ID)
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			expectedResetConfig := "{\"resetKeyAsync\":\"resetValue\",\"resetKeyAsync2\":{\"resetKeyAsync\":\"resetValue2\"}}"
			assignmentsPage := assertFormationAssignmentsCount(t, ctx, formation.ID, tnt, 4)
			formationAssignmentID := getFormationAssignmentIDBySourceAndTarget(t, assignmentsPage, app1.ID, app2.ID)
			reverseAssignmentID := getFormationAssignmentIDBySourceAndTarget(t, assignmentsPage, app2.ID, app1.ID)

			t.Logf("Calling FA status reset for formation assignment with source %q and target %q", app1.ID, app2.ID)
			executeFAStatusResetReqWithExpectedStatusCode(t, certSecuredHTTPClient, "CONFIG_PENDING", expectedResetConfig, tnt, formation.ID, formationAssignmentID, http.StatusOK)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			// We use the async method, because the status API is being called.
			// Even though the case is synchronous, the notification is executed in a separate goroutine and isn't guaranteed
			// to have been executed before we perform the checks otherwise.
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 1)
			notificationsForApp1Tenant := gjson.GetBytes(body, app1.ID)
			assignNotificationForApp1 := notificationsForApp1Tenant.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp1, assignOperation, formation.ID, app1.ID, "", appRegion, expectedResetConfig, tnt, tntParentCustomer)
			require.NoError(t, err)

			assertNotificationsCountForTenant(t, body, app2.ID, 1)
			notificationsForApp2Tenant := gjson.GetBytes(body, app2.ID)
			assignNotificationForApp2 := notificationsForApp2Tenant.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp2, assignOperation, formation.ID, app2.ID, "", appRegion, *fixtures.StatusAPIAsyncConfigJSON, tnt, tntParentCustomer)
			require.NoError(t, err)

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Calling FA status reset for formation assignment with source %q and target %q", app2.ID, app1.ID)
			executeFAStatusResetReqWithExpectedStatusCode(t, certSecuredHTTPClient, "CONFIG_PENDING", expectedResetConfig, tnt, formation.ID, reverseAssignmentID, http.StatusOK)
			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}

			// We use the async method, because the status API is being called.
			// Even though the case is synchronous, the notification is executed in a separate goroutine and isn't guaranteed
			// to have been executed before we perform the checks otherwise.
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 1)
			assignNotificationForApp1 = gjson.GetBytes(body, app1.ID)
			assignNotificationForApp1 = assignNotificationForApp1.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp1, assignOperation, formation.ID, app1.ID, "", appRegion, "", tnt, tntParentCustomer)
			require.NoError(t, err)

			assertNotificationsCountForTenant(t, body, app2.ID, 1)
			notificationsForApp2Tenant = gjson.GetBytes(body, app2.ID)
			assignNotificationForApp2 = notificationsForApp2Tenant.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp2, assignOperation, formation.ID, app2.ID, "", appRegion, "", tnt, tntParentCustomer)
			require.NoError(t, err)
		})

		t.Run("App to App reset with one webhook", func(t *testing.T) {
			applicationTntMappingWebhookType := graphql.WebhookTypeApplicationTenantMapping
			syncWebhookMode := graphql.WebhookModeSync
			urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/no-configuration/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplate := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplate := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, syncWebhookMode, urlTemplate, inputTemplate, outputTemplate)

			t.Logf("Add webhook with type %q and mode: %q to application with ID %q", applicationTntMappingWebhookType, syncWebhookMode, app1.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

			assignmentsPage := assertFormationAssignmentsCount(t, ctx, formation.ID, tnt, 4)
			formationAssignmentID := getFormationAssignmentIDBySourceAndTarget(t, assignmentsPage, app1.ID, app2.ID)

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Calling FA status reset for formation assignment with source %q and target %q", app1.ID, app2.ID)
			executeFAStatusResetReqWithExpectedStatusCode(t, certSecuredHTTPClient, "READY", *fixtures.StatusAPIResetConfigJSON, tnt, formation.ID, formationAssignmentID, http.StatusOK)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIResetConfigJSON, Value: fixtures.StatusAPIResetConfigJSON, Error: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}

			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 1)
			notificationsForApp1Tenant := gjson.GetBytes(body, app1.ID)
			assignNotificationForApp1 := notificationsForApp1Tenant.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp1, assignOperation, formation.ID, app1.ID, "", appRegion, *fixtures.StatusAPIResetConfigJSON, tnt, tntParentCustomer)

			assertNotificationsCountForTenant(t, body, app2.ID, 0)

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Calling FA status reset for formation assignment with source %q and target %q", app1.ID, app2.ID)
			executeFAStatusResetReqWithExpectedStatusCode(t, certSecuredHTTPClient, "CONFIG_PENDING", *fixtures.StatusAPIResetConfigJSON, tnt, formation.ID, formationAssignmentID, http.StatusOK)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "CONFIG_PENDING", Config: fixtures.StatusAPIResetConfigJSON, Value: fixtures.StatusAPIResetConfigJSON, Error: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
				},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionInProgress, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 1)
			notificationsForApp1Tenant = gjson.GetBytes(body, app1.ID)
			assignNotificationForApp1 = notificationsForApp1Tenant.Array()[0]
			err = verifyFormationNotificationForApplicationWithItemsStructure(assignNotificationForApp1, assignOperation, formation.ID, app1.ID, "", appRegion, *fixtures.StatusAPIResetConfigJSON, tnt, tntParentCustomer)

			assertNotificationsCountForTenant(t, body, app2.ID, 0)

		})

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
		var unassignFormation graphql.Formation
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)
	})

	t.Run("Formation Assignment notifications with self-referenced notifications", func(t *testing.T) {
		queryRequest := fixtures.FixQueryFormationConstraintsRequest()
		const constraintName = "DoNotGenerateFormationAssignmentNotificationForLoopsGlobalApplication"

		var actualFormationConstraints []*graphql.FormationConstraint
		require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryRequest, &actualFormationConstraints))

		originalConstraint := findConstraintByName(t, constraintName, actualFormationConstraints)

		originalInput := graphql.FormationConstraintUpdateInput{
			InputTemplate: strings.Trim(strconv.Quote(originalConstraint.InputTemplate), "\""),
		}
		output := formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationInput{}
		unquoted := strings.ReplaceAll(originalInput.InputTemplate, "\\", "")
		err = templatehelper.ParseTemplate(&unquoted, struct {
			SourceApplication *struct {
				ID string `json:"id"`
			} `json:"source_application"`
			ResourceID      string `json:"resource_id"`
			ResourceSubtype string `json:"resource_subtype"`
			ResourceType    string `json:"resource_type"`
			TenantID        string `json:"tenant_id"`
		}{}, &output)

		exceptSubtypes := `\"` + strings.Join(append(output.ExceptSubtypes, []string{applicationType1, applicationType2}...), "\\\", \\\"") + `\"`

		updateInput := graphql.FormationConstraintUpdateInput{
			InputTemplate: fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"source_resource_type\\\": \\\"APPLICATION\\\",\\\"source_resource_id\\\":\\\"{{ if .SourceApplication }}{{.SourceApplication.ID}}{{end}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\",\\\"except_subtypes\\\":[%s]}", exceptSubtypes),
		}

		defer fixtures.UpdateFormationConstraint(t, ctx, originalConstraint.ID, originalInput, certSecuredGraphQLClient)
		fixtures.UpdateFormationConstraint(t, ctx, originalConstraint.ID, updateInput, certSecuredGraphQLClient)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)
		defer resetShouldFailEndpointFromExternalSvcMock(t, certSecuredHTTPClient)

		formationName := "formation-name-from-template-with-webhook"
		t.Logf("Creating formation with name: %q from template with name: %q that has %q webhook", formationName, formationTmplName, graphql.WebhookTypeFormationLifecycle)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)
		require.NotEmpty(t, formation.ID)
		require.Equal(t, "READY", formation.State)
		require.Empty(t, formation.Error)

		t.Run("Assign and Unassign should correctly send notifications for sync participant", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			urlTemplateApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, urlTemplateApplication, inputTemplateApplication, outputTemplateApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tnt, app2.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

			t.Logf("Assign application 2 to formation: %q", formationName)
			defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
			assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
			require.Equal(t, formation.ID, assignedFormation.ID)
			require.Equal(t, formation.State, assignedFormation.State)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app2.ID: {
					app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
			}
			assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Unassign application 2 from formation: %q", formationName)
			unassignReq := fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var unassignFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})
		})

		t.Run("Assign and Unassign should correctly send notifications for async participant", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateAsyncApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

			applicationAsyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, urlTemplateAsyncApplication, inputTemplateAsyncApplication, outputTemplateAsyncApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1.ID)
			actualApplicationAsyncWebhookInput := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationAsyncWebhookInput, tnt, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID)

			t.Logf("Assign application 1 to formation: %q", formationName)
			defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
			assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
			require.Equal(t, formation.ID, assignedFormation.ID)
			require.Equal(t, formation.State, assignedFormation.State)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil}},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Unassign Application 1 from formation %s", formationName)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var unassignFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 0, nil, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

		})

		t.Run("Resynchronize should correctly send notifications for sync participant", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			urlTemplateApplicationFailsSync := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/fail/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookThatFailsInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, urlTemplateApplicationFailsSync, inputTemplateApplication, outputTemplateApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookThatFailsInput, tnt, app2.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

			t.Logf("Assign application 2 to formation: %q", formationName)
			defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
			assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
			require.Equal(t, formation.ID, assignedFormation.ID)
			require.Equal(t, formation.State, assignedFormation.State)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app2.ID: {app2.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON}},
			}
			assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionError, Errors: []*graphql.FormationStatusError{fixtures.StatusAPISyncError}})

			urlTemplateApplicationSucceedsSync := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			applicationWebhookThatSucceedsInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, urlTemplateApplicationSucceedsSync, inputTemplateApplication, outputTemplateApplication)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that succeeds", actualApplicationWebhook.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync)
			updatedWebhookSync := fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID, applicationWebhookThatSucceedsInput)
			require.Equal(t, updatedWebhookSync.ID, actualApplicationWebhook.ID)

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app2.ID: {app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil}},
			}
			assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that respnds with error", actualApplicationWebhook.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync)
			updatedWebhookSync = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID, applicationWebhookThatFailsInput)
			require.Equal(t, updatedWebhookSync.ID, actualApplicationWebhook.ID)

			t.Logf("Unassign Application 2 from formation %s", formationName)
			unassignReq := fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var unassignFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
			require.Error(t, err)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app2.ID: {app2.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON}},
			}
			assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignments)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{
				Condition: graphql.FormationStatusConditionError,
				Errors:    []*graphql.FormationStatusError{fixtures.StatusAPISyncError},
			})

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that succeeds", actualApplicationWebhook.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync)
			updatedWebhookSync = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID, applicationWebhookThatSucceedsInput)
			require.Equal(t, updatedWebhookSync.ID, actualApplicationWebhook.ID)

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)
			assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		})

		t.Run("Resynchronize should correctly send notifications for async participant", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-fail/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateAsyncApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

			applicationAsyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, urlTemplateAsyncApplication, inputTemplateAsyncApplication, outputTemplateAsyncApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1.ID)
			actualApplicationAsyncWebhookInput := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationAsyncWebhookInput, tnt, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID)

			t.Logf("Assign application 1 to formation: %q", formationName)
			defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
			assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
			require.Equal(t, formation.ID, assignedFormation.ID)
			require.Equal(t, formation.State, assignedFormation.State)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app1.ID: {app1.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPIAsyncErrorMessageJSON, Error: fixtures.StatusAPIAsyncErrorMessageJSON}},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)

			urlTemplateApplicationSucceedsAsync := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			applicationAsyncWebhookThatSucceedsInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, urlTemplateApplicationSucceedsAsync, inputTemplateAsyncApplication, outputTemplateAsyncApplication)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that succeeds", actualApplicationAsyncWebhookInput.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback)
			updatedWebhook := fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID, applicationAsyncWebhookThatSucceedsInput)
			require.Equal(t, updatedWebhook.ID, actualApplicationAsyncWebhookInput.ID)

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil}},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that respnds with error", actualApplicationAsyncWebhookInput.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback)
			updatedWebhook = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID, applicationAsyncWebhookInput)
			require.Equal(t, updatedWebhook.ID, actualApplicationAsyncWebhookInput.ID)

			t.Logf("Unassign Application 1 from formation %s", formationName)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var unassignFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, unassignFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {app1.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: fixtures.StatusAPIAsyncErrorMessageJSON, Error: fixtures.StatusAPIAsyncErrorMessageJSON}},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that succeds", actualApplicationAsyncWebhookInput.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback)
			updatedWebhook = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID, applicationAsyncWebhookThatSucceedsInput)
			require.Equal(t, updatedWebhook.ID, actualApplicationAsyncWebhookInput.ID)

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 0, nil, eventuallyTimeout,eventuallyTick)
		})

		t.Run("Resynchronize should correctly send notifications for multiple participants", func(t *testing.T) {
			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			urlTemplateApplicationFailsSync := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/fail/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplateApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}"

			applicationWebhookThatFailsInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, urlTemplateApplicationFailsSync, inputTemplateApplication, outputTemplateApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2.ID)
			actualApplicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookThatFailsInput, tnt, app2.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID)

			urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-fail/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			inputTemplateAsyncApplication := "{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}"
			outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

			applicationAsyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, urlTemplateAsyncApplication, inputTemplateAsyncApplication, outputTemplateAsyncApplication)

			t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1.ID)
			actualApplicationAsyncWebhookInput := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationAsyncWebhookInput, tnt, app1.ID)
			defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID)

			t.Logf("Assign application 1 to formation: %q", formationName)
			defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
			assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app1.ID, tnt)
			require.Equal(t, formation.ID, assignedFormation.ID)
			require.Equal(t, formation.State, assignedFormation.State)

			expectedAssignments := map[string]map[string]fixtures.AssignmentState{
				app1.ID: {app1.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPIAsyncErrorMessageJSON, Error: fixtures.StatusAPIAsyncErrorMessageJSON}},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignments, eventuallyTimeout,eventuallyTick)

			body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 1)

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Assign application 2 to formation: %q", formationName)
			defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
			assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app2.ID, tnt)
			require.Equal(t, formation.ID, assignedFormation.ID)
			require.Equal(t, formation.State, assignedFormation.State)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPIAsyncErrorMessageJSON, Error: fixtures.StatusAPIAsyncErrorMessageJSON},
					app2.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPIAsyncErrorMessageJSON, Error: fixtures.StatusAPIAsyncErrorMessageJSON},
					app2.ID: fixtures.AssignmentState{State: "CREATE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
				},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID,
				graphql.FormationStatus{
					Condition: graphql.FormationStatusConditionError,
					Errors: []*graphql.FormationStatusError{
						fixtures.StatusAPISyncError,
						fixtures.StatusAPIAsyncError,
						fixtures.StatusAPISyncError,
						fixtures.StatusAPIAsyncError,
					}})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			assertNotificationsCountForTenant(t, body, app1.ID, 1)
			assertNotificationsCountForTenant(t, body, app2.ID, 2)

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			urlTemplateApplicationSucceedsAsync := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			applicationAsyncWebhookThatSucceedsInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, urlTemplateApplicationSucceedsAsync, inputTemplateAsyncApplication, outputTemplateAsyncApplication)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that succeeds", actualApplicationAsyncWebhookInput.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback)
			updatedWebhook := fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID, applicationAsyncWebhookThatSucceedsInput)
			require.Equal(t, updatedWebhook.ID, actualApplicationAsyncWebhookInput.ID)

			urlTemplateApplicationSucceedsSync := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
			applicationWebhookThatSucceedsInput := fixtures.FixFormationNotificationWebhookInput(graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, urlTemplateApplicationSucceedsSync, inputTemplateApplication, outputTemplateApplication)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that succeeds", actualApplicationWebhook.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync)
			updatedWebhookSync := fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID, applicationWebhookThatSucceedsInput)
			require.Equal(t, updatedWebhookSync.ID, actualApplicationWebhook.ID)

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON, Value: fixtures.StatusAPIAsyncConfigJSON, Error: nil},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady, Errors: nil})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			assertNotificationsCountForTenant(t, body, app1.ID, 3)
			assertNotificationsCountForTenant(t, body, app2.ID, 2)

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that responds with error", actualApplicationAsyncWebhookInput.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback)
			updatedWebhook = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID, applicationAsyncWebhookInput)
			require.Equal(t, updatedWebhook.ID, actualApplicationAsyncWebhookInput.ID)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that responds with error", actualApplicationWebhook.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync)
			updatedWebhookSync = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID, applicationWebhookThatFailsInput)
			require.Equal(t, updatedWebhookSync.ID, actualApplicationWebhook.ID)

			t.Logf("Unassign Application 1 from formation %s", formationName)
			unassignReq := fixtures.FixUnassignFormationRequest(app1.ID, string(graphql.FormationObjectTypeApplication), formationName)
			var unassignFormation graphql.Formation
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
			require.Error(t, err)

			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: fixtures.StatusAPIAsyncErrorMessageJSON, Error: fixtures.StatusAPIAsyncErrorMessageJSON},
					app2.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: fixtures.StatusAPIAsyncErrorMessageJSON, Error: fixtures.StatusAPIAsyncErrorMessageJSON},
					app2.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPISyncConfigJSON, Value: fixtures.StatusAPISyncConfigJSON, Error: nil},
				},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID,
				graphql.FormationStatus{
					Condition: graphql.FormationStatusConditionError,
					Errors: []*graphql.FormationStatusError{
						fixtures.StatusAPISyncError,
						fixtures.StatusAPIAsyncError,
						fixtures.StatusAPISyncError,
					}})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
			assertNotificationsCountForTenant(t, body, app1.ID, 2)
			assertNotificationsCountForTenant(t, body, app2.ID, 1)

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Unassign Application 2 from formation %s", formationName)
			unassignReq = fixtures.FixUnassignFormationRequest(app2.ID, string(graphql.FormationObjectTypeApplication), formationName)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
			require.Error(t, err)
			expectedAssignments = map[string]map[string]fixtures.AssignmentState{
				app1.ID: {
					app1.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: fixtures.StatusAPIAsyncErrorMessageJSON, Error: fixtures.StatusAPIAsyncErrorMessageJSON},
					app2.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
				},
				app2.ID: {
					app1.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: fixtures.StatusAPIAsyncErrorMessageJSON, Error: fixtures.StatusAPIAsyncErrorMessageJSON},
					app2.ID: fixtures.AssignmentState{State: "DELETE_ERROR", Config: nil, Value: fixtures.StatusAPISyncErrorMessageJSON, Error: fixtures.StatusAPISyncErrorMessageJSON},
				},
			}
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignments, eventuallyTimeout,eventuallyTick)
			assertFormationStatus(t, ctx, tnt, formation.ID,
				graphql.FormationStatus{
					Condition: graphql.FormationStatusConditionError,
					Errors: []*graphql.FormationStatusError{
						fixtures.StatusAPISyncError,
						fixtures.StatusAPIAsyncError,
						fixtures.StatusAPISyncError,
						fixtures.StatusAPIAsyncError,
					}})

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			assertNotificationsCountForTenant(t, body, app1.ID, 1)
			assertNotificationsCountForTenant(t, body, app2.ID, 2)

			cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that succeeds", actualApplicationAsyncWebhookInput.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback)
			updatedWebhook = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID, applicationAsyncWebhookThatSucceedsInput)
			require.Equal(t, updatedWebhook.ID, actualApplicationAsyncWebhookInput.ID)

			t.Logf("Update webhook with ID: %q of type: %q and mode: %q to have URLTemlate that points to endpoint that succeeds", actualApplicationWebhook.ID, graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync)
			updatedWebhookSync = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationWebhook.ID, applicationWebhookThatSucceedsInput)
			require.Equal(t, updatedWebhookSync.ID, actualApplicationWebhook.ID)

			t.Logf("Resynchronize formation %s should retry and succeed", formation.Name)
			resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequest(formation.ID)
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, resynchronizeReq, &assignedFormation)
			require.NoError(t, err)
			require.Equal(t, formationName, assignedFormation.Name)
			assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 0, nil, eventuallyTimeout,eventuallyTick)

			body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

			assertNotificationsCountForTenant(t, body, app1.ID, 2)
			assertNotificationsCountForTenant(t, body, app2.ID, 2)
		})
	})

	t.Run("Formation assignment notifications validating redirect operator", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		urlTemplateSyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/configuration/redirect-notification/{{.TargetApplication.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		inputTemplateApplication := "{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}"
		outputTemplateSyncApplication := "{\\\"config\\\":\\\"{{.Body.configuration}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}"

		applicationTntMappingWebhookType := graphql.WebhookTypeApplicationTenantMapping
		syncWebhookMode := graphql.WebhookModeSync
		asyncCallbackWebhookMode := graphql.WebhookModeAsyncCallback
		applicationSyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, syncWebhookMode, urlTemplateSyncApplication, inputTemplateApplication, outputTemplateSyncApplication)

		t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", applicationTntMappingWebhookType, syncWebhookMode, app2.ID)
		applicationSyncWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationSyncWebhookInput, tnt, app2.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, applicationSyncWebhook.ID)

		urlTemplateAsyncApplication := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.TargetApplication.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
		outputTemplateAsyncApplication := "{\\\"config\\\":\\\"{{.Body.configuration}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}"

		applicationAsyncWebhookInput := fixtures.FixFormationNotificationWebhookInput(applicationTntMappingWebhookType, asyncCallbackWebhookMode, urlTemplateAsyncApplication, inputTemplateApplication, outputTemplateAsyncApplication)

		t.Logf("Add webhook with type %q and mode: %q to application with ID: %q", applicationTntMappingWebhookType, asyncCallbackWebhookMode, app1.ID)
		actualApplicationAsyncWebhookInput := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationAsyncWebhookInput, tnt, app1.ID)
		defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tnt, actualApplicationAsyncWebhookInput.ID)

		redirectedTntID := "custom-redirected-tenant-id"
		redirectPath := "/formation-callback/redirect-notification/" + redirectedTntID
		redirectURL := fmt.Sprintf("%s%s", conf.ExternalServicesMockMtlsSecuredURL, redirectPath)
		redirectURLTemplate := fmt.Sprintf("{\\\\\\\"path\\\\\\\":\\\\\\\"%s\\\\\\\",\\\\\\\"method\\\\\\\":\\\\\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\\\\\"}", redirectURL)
		// create formation constraints and attach them to formation template
		redirectConstraintInput := graphql.FormationConstraintInput{
			Name:            "e2e-redirect-operator-constraint",
			ConstraintType:  graphql.ConstraintTypePre,
			TargetOperation: graphql.TargetOperationSendNotification,
			Operator:        formationconstraintpkg.RedirectNotificationOperator,
			ResourceType:    graphql.ResourceTypeApplication,
			ResourceSubtype: applicationType2,
			InputTemplate:   fmt.Sprintf("{\\\"should_redirect\\\": {{ if contains .FormationAssignment.Value \\\"redirectProperties\\\" }}true{{else}}false{{end}},\\\"url_template\\\": \\\"%s\\\",\\\"url\\\": \\\"%s\\\",{{ if .Webhook }}\\\"webhook_memory_address\\\":{{ .Webhook.GetAddress }},{{ end }}\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}", redirectURLTemplate, redirectURL),
			ConstraintScope: graphql.ConstraintScopeFormationType,
		}

		firstConstraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, redirectConstraintInput)
		defer fixtures.CleanupFormationConstraint(t, ctx, certSecuredGraphQLClient, firstConstraint.ID)
		require.NotEmpty(t, firstConstraint.ID)

		fixtures.AttachConstraintToFormationTemplate(t, ctx, certSecuredGraphQLClient, firstConstraint.ID, firstConstraint.Name, ft.ID, ft.Name)

		// create formation
		formationName := "e2e-redirect-formation-name"
		t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName)
		formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tnt, formationName, &formationTmplName)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady})

		formationInput := graphql.FormationInput{Name: formationName}
		t.Logf("Assign application 2 with ID: %s to formation: %q", app2.ID, formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app2.ID, tnt)
		assignedFormation := fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app2.ID, tnt)
		require.Equal(t, formation.ID, assignedFormation.ID)
		require.Equal(t, formation.State, assignedFormation.State)

		t.Log("Assert no formation assignment notifications are sent when there is only one app in formation")
		body := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, tnt, 0)
		assertNotificationsCountForTenant(t, body, localTenantID2, 0)

		expectedAssignmentsBySourceID := map[string]map[string]fixtures.AssignmentState{
			app2.ID: {
				app2.ID: fixtures.AssignmentState{State: "READY"},
			},
		}
		assertFormationAssignments(t, ctx, tnt, formation.ID, 1, expectedAssignmentsBySourceID)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady})

		t.Logf("Assign application 1 with ID: %s to formation %s", app1.ID, formationName)
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app1.ID, tnt)
		assignedFormation = fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, formationInput, app1.ID, tnt)
		require.Equal(t, formationName, assignedFormation.Name)

		expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app2.ID: fixtures.AssignmentState{State: "READY"},
				app1.ID: fixtures.AssignmentState{State: "READY"},
			},
			app2.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY", Config: fixtures.StatusAPIAsyncConfigJSON},
				app2.ID: fixtures.AssignmentState{State: "READY"},
			},
		}

		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 4, expectedAssignmentsBySourceID, eventuallyTimeoutForDestinations, eventuallyTickForDestinations)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady})

		t.Logf("Assert formation assignment notifications for %s operation...", assignOperation)
		notifications := getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		// Normally, the app 1 tenant should have two formation assignment notifications
		// but due to redirect operator, one of them is redirected to a different receiver
		assertNotificationsCountForTenant(t, notifications, localTenantID2, 1)
		notificationsForApp2 := gjson.GetBytes(notifications, localTenantID2)
		assignNotificationAboutApp1 := notificationsForApp2.Array()[0]
		assertFormationAssignmentsNotification(t, assignNotificationAboutApp1, assignOperation, formation.ID, app1.ID, app2.ID, initialAssignmentState, initialAssignmentState, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)

		// validate the second(redirected) formation assignment notifications
		assertNotificationsCountForTenant(t, notifications, redirectedTntID, 1)
		redirectedNotifications := gjson.GetBytes(notifications, redirectedTntID)
		redirectedNotification := redirectedNotifications.Array()[0]
		assertFormationAssignmentsNotification(t, redirectedNotification, assignOperation, formation.ID, app1.ID, app2.ID, configPendingAssignmentState, readyAssignmentState, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)
		require.Equal(t, redirectPath, redirectedNotification.Get("RequestPath").String())

		assertNotificationsCountForTenant(t, notifications, localTenantID, 1)
		notificationsForApp1 := gjson.GetBytes(notifications, localTenantID)
		assignNotificationAboutApp2 := notificationsForApp1.Array()[0]
		assertFormationAssignmentsNotification(t, assignNotificationAboutApp2, assignOperation, formation.ID, app2.ID, app1.ID, initialAssignmentState, configPendingAssignmentState, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer)

		t.Logf("Deleting the stored notifications in the external services mock")
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		var unassignFormation graphql.Formation
		t.Logf("Unassign Application 2 from formation %s", formationName)
		unassignReq := fixtures.FixUnassignFormationRequest(app2.ID, graphql.FormationObjectTypeApplication.String(), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		expectedAssignmentsBySourceID = map[string]map[string]fixtures.AssignmentState{
			app1.ID: {
				app1.ID: fixtures.AssignmentState{State: "READY"},
			},
		}

		assertFormationAssignmentsAsynchronouslyWithEventually(t, ctx, tnt, formation.ID, 1, expectedAssignmentsBySourceID, eventuallyTimeoutForDestinations, eventuallyTickForDestinations)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady})

		t.Logf("Assert formation assignment notifications for %s operation...", unassignOperation)
		body = getNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		assertNotificationsCountForTenant(t, body, localTenantID, 1)
		unassignNotificationsForApp1 := gjson.GetBytes(body, localTenantID)
		unassignNotificationForApp1 := unassignNotificationsForApp1.Array()[0]
		assertFormationAssignmentsNotification(t, unassignNotificationForApp1, unassignOperation, formation.ID, app2.ID, app1.ID, deletingAssignmentState, deletingAssignmentState, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer)

		assertNotificationsCountForTenant(t, body, localTenantID2, 1)
		unassignNotificationsForApp2 := gjson.GetBytes(body, localTenantID2)
		unassignNotificationForApp2 := unassignNotificationsForApp2.Array()[0]
		assertFormationAssignmentsNotification(t, unassignNotificationForApp2, unassignOperation, formation.ID, app1.ID, app2.ID, deletingAssignmentState, deletingAssignmentState, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		unassignReq = fixtures.FixUnassignFormationRequest(app1.ID, graphql.FormationObjectTypeApplication.String(), formationName)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tnt, unassignReq, &unassignFormation)
		require.NoError(t, err)
		require.Equal(t, formationName, unassignFormation.Name)

		assertFormationAssignments(t, ctx, tnt, formation.ID, 0, nil)
		assertFormationStatus(t, ctx, tnt, formation.ID, graphql.FormationStatus{Condition: graphql.FormationStatusConditionReady})
	})
}
