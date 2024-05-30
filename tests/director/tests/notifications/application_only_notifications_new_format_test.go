package notifications

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"

	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	mock_data "github.com/kyma-incubator/compass/tests/pkg/notifications/expectations-builders"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/operations"
	resource_providers "github.com/kyma-incubator/compass/tests/pkg/notifications/resource-providers"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
)

func TestFormationNotificationsWithApplicationOnlyParticipantsNewFormat(t *testing.T) {
	ctx := context.Background()
	tnt := tenant.TestTenants.GetDefaultTenantID()
	tntParentCustomer := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant) // parent of `tenant.TestTenants.GetDefaultTenantID()` above

	emptyConfig := ""

	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	formationTemplateName := "app-only-formation-template-name"

	certSubjectMappingCustomSubject := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.TestExternalCertCN, "instance-creator", -1)

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

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, true)
	instanceCreatorCertClient := gql.NewCertAuthorizedHTTPClient(pk, cert, conf.SkipSSLValidation)

	certSubjectMappingCN := "csm-async-new-format-callback-cn"
	certSubjectMappingCNSecond := "csm-async-new-format-callback-cn-second"
	certSubjectMappingCustomSubject = strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.TestExternalCertCN, certSubjectMappingCN, -1)
	certSubjectMappingCustomSubjectSecond := strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.TestExternalCertCN, certSubjectMappingCNSecond, -1)

	// We need an externally issued cert with a custom subject that will be used to create a certificate subject mapping through the GraphQL API,
	// which later will be loaded in-memory from the hydrator component
	externalCertProviderConfig = certprovider.ExternalCertProviderConfig{
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
	appPk, appCert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, false)
	appCertClient := gql.NewCertAuthorizedHTTPClient(appPk, appCert, conf.SkipSSLValidation)

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
	applicationType1 := "app-type-1"

	appTemplateProvider := resource_providers.NewApplicationTemplateProvider(applicationType1, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder, tnt, nil, graphql.ApplicationStatusConditionConnected)
	defer appTemplateProvider.Cleanup(t, ctx, oauthGraphQLClient)
	appTpl := appTemplateProvider.Provide(t, ctx, oauthGraphQLClient)
	appTplID := appTpl.ID
	internalConsumerID := appTplID // add application templated ID as certificate subject mapping internal consumer to satisfy the authorization checks in the formation assignment status API
	t.Logf("Created application template for type: %q", applicationType1)

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
	applicationType2 := "app-type-2"
	appTemplateProvider2 := resource_providers.NewApplicationTemplateProvider(applicationType2, localTenantID2, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder, tnt, nil, graphql.ApplicationStatusConditionConnected)
	defer appTemplateProvider2.Cleanup(t, ctx, oauthGraphQLClient)
	appTpl2 := appTemplateProvider2.Provide(t, ctx, oauthGraphQLClient)
	appTplID2 := appTpl2.ID
	t.Logf("Created application template for type: %q", applicationType2)

	ftProvider := resource_providers.NewFormationTemplateCreator(formationTemplateName)
	defer ftProvider.Cleanup(t, ctx, certSecuredGraphQLClient)
	ftplID := ftProvider.WithSupportedResources(appTemplateProvider.GetResource(), appTemplateProvider2.GetResource()).WithLeadingProductIDs([]string{internalConsumerID}).Provide(t, ctx, certSecuredGraphQLClient)
	ctx = context.WithValue(ctx, context_keys.FormationTemplateIDKey, ftplID)
	ctx = context.WithValue(ctx, context_keys.FormationTemplateNameKey, ftplID)
	t.Logf("Created Formation Template with ID: %q and name: %q", ftplID, formationTemplateName)

	appProvider1 := resource_providers.NewApplicationFromTemplateProvider(applicationType1, namePlaceholder, "app1-formation-notifications-tests", displayNamePlaceholder, "App 1 Display Name", tnt)
	defer appProvider1.Cleanup(t, ctx, certSecuredGraphQLClient)
	app1ID := appProvider1.Provide(t, ctx, certSecuredGraphQLClient)
	t.Logf("Created application 1 with ID: %q from template: %q", app1ID, applicationType1)

	appProvider2 := resource_providers.NewApplicationFromTemplateProvider(applicationType2, namePlaceholder, "app2-formation-notifications-tests", displayNamePlaceholder, "App 2 Display Name", tnt)
	defer appProvider2.Cleanup(t, ctx, certSecuredGraphQLClient)
	app2ID := appProvider2.Provide(t, ctx, certSecuredGraphQLClient)
	t.Logf("Created application 2 with ID: %q from template: %q", app2ID, applicationType2)

	formationName := "app-to-app-formation-name"
	formationProvider := resource_providers.NewFormationProvider(formationName, tnt, &formationTemplateName)
	defer formationProvider.Cleanup(t, ctx, certSecuredGraphQLClient)
	formationID := formationProvider.Provide(t, ctx, certSecuredGraphQLClient)
	t.Logf("Created formation with name: %q and ID: %q from template with name: %q", formationName, formationID, formationTemplateName)
	ctx = context.WithValue(ctx, context_keys.FormationIDKey, formationID)
	ctx = context.WithValue(ctx, context_keys.FormationNameKey, formationName)

	expectationsBuilder := mock_data.NewFAExpectationsBuilder()
	asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
		AssertExpectations(t, ctx)
	asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).AssertExpectations(t, ctx)

	t.Run("Synchronous App to App Formation Assignment Notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Add webhook with type: %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app1ID)
		op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		asserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, fixtures.StatusAPISyncConfigJSON, nil),
			})
		faAsserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsAsserter := asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s again", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, fixtures.StatusAPISyncConfigJSON, nil),
			})
		asserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsAsserter = asserters.NewNotificationsAsserter(3, assignOperation, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID)
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter = asserters.NewUnassignNotificationsAsserter(2, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder()
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Synchronous App to App Formation Assignment Notifications when state is in the response body", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app1ID)
		op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate(`{\"path\":\"` + conf.ExternalServicesMockMtlsSecuredURL + `/formation-callback/with-state/{{.TargetApplication.ID}}{{if eq .Operation \"unassign\"}}/{{.SourceApplication.ID}}{{end}}\",\"method\":\"{{if eq .Operation \"assign\"}}PATCH{{else}}DELETE{{end}}\"}`).
			WithInputTemplate(`{\"ucl-formation-id\":\"{{.FormationID}}\",\"globalAccountId\":\"{{.CustomerTenantContext.AccountID}}\",\"crmId\":\"{{.CustomerTenantContext.CustomerID}}\",\"config\":{{ .ReverseAssignment.Value }},\"items\":[{\"region\":\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\",\"application-namespace\":\"{{.SourceApplicationTemplate.ApplicationNamespace}}\",\"tenant-id\":\"{{.SourceApplication.LocalTenantID}}\",\"ucl-system-tenant-id\":\"{{.SourceApplication.ID}}\"}]}`).
			WithOutputTemplate(`{\"config\":\"{{.Body.config}}\", \"state\":\"{{.Body.state}}\", \"location\":\"{{.Headers.Location}}\",\"error\": \"{{.Body.error}}\",\"success_status_code\": 200}`).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAddConstraintOperation("e2e-destination-creator-notification-status-returned").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator(formationconstraintpkg.DestinationCreator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype("ANY").
			WithInputTemplate("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .NotificationStatusReport }}\\\"notification_status_report_memory_address\\\":{{ .NotificationStatusReport.GetAddress }},{{ end }}{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}").
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAddConstraintOperation("e2e-destination-creator-send-notification").
			WithTargetOperation(graphql.TargetOperationSendNotification).
			WithOperator(formationconstraintpkg.DestinationCreator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype("ANY").
			WithInputTemplate("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}").
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		asserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, configPendingAssignmentState, fixtures.StatusAPISyncConfigJSON, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", false),
			})
		faAsserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).WithCondition(graphql.FormationStatusConditionInProgress)
		notificationsAsserter := asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).WithUseItemsStruct(true)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).WithUseItemsStruct(true)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, configPendingAssignmentState, fixtures.StatusAPISyncConfigJSON, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", false),
			})
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).WithCondition(graphql.FormationStatusConditionInProgress)
		notificationsAsserter = asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).WithUseItemsStruct(true)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID)
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter = asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).WithUseItemsStruct(true)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder()
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Asynchronous App to App Formation Assignment Notifications using the Default Tenant Mapping Handler", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate(`{\"path\":\"` + conf.CompassExternalMTLSGatewayURL + `/default-tenant-mapping-handler/v1/tenantMappings/{{.TargetApplication.ID}}\",\"method\":\"PATCH\"}`).
			WithInputTemplate("").
			WithOutputTemplate(`{\"error\": \"{{.Body.error}}\",\"success_status_code\": 202}`).
			WithHeaderTemplate(str.Ptr(`{\"Content-Type\": [\"application/json\"], \"Location\":[\"` + conf.CompassExternalMTLSGatewayURL + `/v1/businessIntegrations/{{.FormationID}}/assignments/{{.Assignment.ID}}/status\"]}`)).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		faAsserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID)
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s again", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder()
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Synchronous App to App Formation Assignment Notifications with CREATE_READY and DELETE_READY states", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Add webhook with type: %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app1ID)
		op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/sync-{{if eq .Operation \\\"assign\\\"}}create{{else}}delete{{end}}-ready/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		faAsserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, fixtures.StatusAPISyncConfigJSON, nil),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsAsserter := asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder()
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Asynchronous App to App Formation Assignment Notifications with CREATE_READY and DELETE_READY states", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Add webhook with type: %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-{{if eq .Operation \\\"assign\\\"}}create{{else}}delete{{end}}-ready/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\",\\\"operationId\\\":\\\"{{.AssignmentOperation.ID}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		faAsserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, nil, nil),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsAsserter := asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Update URL template for webhook with type %q and mode: %q for application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op = operations.NewUpdateWebhookOperation().
			WithWebhookType(graphql.WebhookTypeApplicationTenantMapping).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-no-response/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\",\\\"operationId\\\":\\\"{{.AssignmentOperation.ID}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").
			WithObjectID(app1ID).
			WithObjectType(operations.WebhookReferenceObjectTypeApplication).
			WithTenantID(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, initialAssignmentState, nil, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", false),
			})
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).WithCondition(graphql.FormationStatusConditionInProgress)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewExecuteStatusReportOperation().WithTenant(tnt).
			WithFormationAssignment(app2ID, app1ID).
			WithStatusCode(http.StatusBadRequest).
			WithState(deleteReadyAssignmentState).
			WithHTTPClient(appCertClient).
			WithExternalServicesMockMtlsSecuredURL(conf.DirectorExternalCertFAAsyncStatusURL).
			WithAsserters(faAsserter, statusAsserter).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, nil, nil),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewExecuteStatusReportOperation().WithTenant(tnt).
			WithFormationAssignment(app2ID, app1ID).
			WithState(createReadyAssignmentState).
			WithHTTPClient(appCertClient).
			WithExternalServicesMockMtlsSecuredURL(conf.DirectorExternalCertFAAsyncStatusURL).
			WithAsserters(faAsyncAsserter, statusAsserter).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithCustomParticipants([]string{app2ID}).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, deletingAssignmentState, nil, nil),
				mock_data.NewNotificationData(app2ID, app2ID, readyAssignmentState, nil, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app1ID, "UNASSIGN", "UNASSIGN_OBJECT", false),
			})
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).WithCondition(graphql.FormationStatusConditionInProgress)
		unassignNotificationsAsserter = asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewExecuteStatusReportOperation().WithTenant(tnt).
			WithFormationAssignment(app2ID, app1ID).
			WithStatusCode(http.StatusBadRequest).
			WithState(createReadyAssignmentState).
			WithHTTPClient(appCertClient).
			WithExternalServicesMockMtlsSecuredURL(conf.DirectorExternalCertFAAsyncStatusURL).
			WithAsserters(faAsserter, statusAsserter).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app2ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewExecuteStatusReportOperation().WithTenant(tnt).
			WithFormationAssignment(app2ID, app1ID).
			WithState(deleteReadyAssignmentState).
			WithHTTPClient(appCertClient).
			WithExternalServicesMockMtlsSecuredURL(conf.DirectorExternalCertFAAsyncStatusURL).
			WithAsserters(faAsyncAsserter, statusAsserter).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder()
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Use Application Template Webhook if App does not have one for notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app1ID)
		op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate(`{\"path\":\"` + conf.ExternalServicesMockMtlsSecuredURL + `/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \"unassign\"}}/{{.SourceApplication.ID}}{{end}}\",\"method\":\"{{if eq .Operation \"assign\"}}PATCH{{else}}DELETE{{end}}\"}`).
			WithInputTemplate(`{\"ucl-formation-id\":\"{{.FormationID}}\",\"globalAccountId\":\"{{.CustomerTenantContext.AccountID}}\",\"crmId\":\"{{.CustomerTenantContext.CustomerID}}\",\"config\":{{ .ReverseAssignment.Value }},\"items\":[{\"region\":\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\",\"application-namespace\":\"{{.SourceApplicationTemplate.ApplicationNamespace}}\",\"tenant-id\":\"{{.SourceApplication.LocalTenantID}}\",\"ucl-system-tenant-id\":\"{{.SourceApplication.ID}}\"}]}`).
			WithOutputTemplate(`{\"config\":\"{{.Body.config}}\", \"location\":\"{{.Headers.Location}}\",\"error\": \"{{.Body.error}}\",\"success_status_code\": 200, \"incomplete_status_code\": 204}`).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, appTplID2)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplicationTemplate, appTplID2, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate(`{\"path\":\"` + conf.ExternalServicesMockMtlsSecuredURL + `/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \"unassign\"}}/{{.SourceApplication.ID}}{{end}}\",\"method\":\"{{if eq .Operation \"assign\"}}PATCH{{else}}DELETE{{end}}\"}`).
			WithInputTemplate(`{\"ucl-formation-id\":\"{{.FormationID}}\",\"globalAccountId\":\"{{.CustomerTenantContext.AccountID}}\",\"crmId\":\"{{.CustomerTenantContext.CustomerID}}\",\"config\":{{ .ReverseAssignment.Value }},\"items\":[{\"region\":\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\",\"application-namespace\":\"{{.SourceApplicationTemplate.ApplicationNamespace}}\",\"tenant-id\":\"{{.SourceApplication.LocalTenantID}}\",\"ucl-system-tenant-id\":\"{{.SourceApplication.ID}}\"}]}`).
			WithOutputTemplate(`{\"config\":\"{{.Body.config}}\", \"location\":\"{{.Headers.Location}}\",\"error\": \"{{.Body.error}}\",\"success_status_code\": 200, \"incomplete_status_code\": 204}`).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		// register a few more webhooks for the application template to verify that only the correct type of webhook is used when generation formation notifications
		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeUnregisterApplication, graphql.WebhookModeSync, appTplID2)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeUnregisterApplication, operations.WebhookReferenceObjectTypeApplicationTemplate, appTplID2, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURL("http://new-webhook.url").
			WithOutputTemplate(`{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}`).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeRegisterApplication, graphql.WebhookModeSync, appTplID2)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeRegisterApplication, operations.WebhookReferenceObjectTypeApplicationTemplate, appTplID2, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURL("http://new-webhook.url").
			WithOutputTemplate(`{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}`).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeOpenResourceDiscovery, graphql.WebhookModeSync, appTplID2)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeOpenResourceDiscovery, operations.WebhookReferenceObjectTypeApplicationTemplate, appTplID2, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURL("http://new-webhook.url").
			WithOutputTemplate(`{\"location\":\"{{.Headers.Location}}\",\"success_status_code\": 202,\"error\": \"{{.Body.error}}\"}`).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		// Create formation constraints for destination creator operator and attach them to a given formation template.
		// So we can verify the destination creator will not fail if in the configuration there is no destination information
		op = operations.NewAddConstraintOperation("e2e-destination-creator-notification-status-returned").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator(formationconstraintpkg.DestinationCreator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype("ANY").
			WithInputTemplate("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .NotificationStatusReport }}\\\"notification_status_report_memory_address\\\":{{ .NotificationStatusReport.GetAddress }},{{ end }}{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}").
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAddConstraintOperation("e2e-destination-creator-send-notification").
			WithTargetOperation(graphql.TargetOperationSendNotification).
			WithOperator(formationconstraintpkg.DestinationCreator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype("ANY").
			WithInputTemplate("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}").
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		asserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, fixtures.StatusAPISyncConfigJSON, nil),
				mock_data.NewNotificationData(app1ID, app2ID, readyAssignmentState, fixtures.StatusAPISyncConfigJSON, nil),
			})
		faAsserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).WithCondition(graphql.FormationStatusConditionReady)
		notificationsAsserter := asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, nil, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).WithUseItemsStruct(true)
		notificationsAsserter2 := asserters.NewNotificationsAsserter(1, assignOperation, app2ID, app1ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, nil, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).WithUseItemsStruct(true)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter, notificationsAsserter, notificationsAsserter2).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Formation lifecycle asynchronous notifications and asynchronous app to app formation assignment notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Add webhook with type %q and mode: %q to application template with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, appTplID)
		op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplicationTemplate, appTplID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate(`{\"path\":\"` + conf.ExternalServicesMockMtlsSecuredURL + `/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \"unassign\"}}/{{.SourceApplication.ID}}{{end}}\",\"method\":\"{{if eq .Operation \"assign\"}}PATCH{{else}}DELETE{{end}}\"}`).
			WithInputTemplate(`{\"ucl-formation-id\":\"{{.FormationID}}\",\"operation-id\":\"{{.AssignmentOperation.ID}}\",\"globalAccountId\":\"{{.CustomerTenantContext.AccountID}}\",\"crmId\":\"{{.CustomerTenantContext.CustomerID}}\",\"formation-assignment-id\":\"{{ .Assignment.ID }}\", \"config\":{{ .ReverseAssignment.Value }},\"items\":[{\"region\":\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\",\"application-namespace\":\"{{.SourceApplicationTemplate.ApplicationNamespace}}\",\"tenant-id\":\"{{.SourceApplication.LocalTenantID}}\",\"ucl-system-tenant-id\":\"{{.SourceApplication.ID}}\",\"source-trust-details\":[{{ Join  .SourceApplicationTemplate.TrustDetails.Subjects }}],\"target-trust-details\":[{{ Join  .TargetApplicationTemplate.TrustDetails.Subjects }}] }]}`).
			WithOutputTemplate(`{\"config\":\"{{.Body.config}}\", \"location\":\"{{.Headers.Location}}\",\"error\": \"{{.Body.error}}\",\"success_status_code\": 202}`).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to formation template with ID %q", graphql.WebhookTypeFormationLifecycle, graphql.WebhookModeAsyncCallback, ftplID)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeFormationLifecycle, operations.WebhookReferenceObjectTypeFormationTemplate, ftplID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate(`{\"path\":\"` + conf.ExternalServicesMockMtlsSecuredURL + `/v1/businessIntegration/async-no-response/{{.Formation.ID}}\",\"method\":\"{{if eq .Operation \"createFormation\"}}POST{{else}}DELETE{{end}}\"}`).
			WithInputTemplate(`{\"globalAccountId\":\"{{.CustomerTenantContext.AccountID}}\",\"crmId\":\"{{.CustomerTenantContext.CustomerID}}\",\"details\":{\"id\":\"{{.Formation.ID}}\",\"name\":\"{{.Formation.Name}}\"}}`).
			WithOutputTemplate(`{\"error\": \"{{.Body.error}}\",\"success_status_code\": 202}`).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		formationName := "formation-name-from-template-with-webhook"
		t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).
			WithFormationName(formationName).
			WithCondition(graphql.FormationStatusConditionInProgress).
			WithState(initialAssignmentState)
		createFormation := operations.NewCreateFormationOperation(tnt).
			WithName(formationName).
			WithFormationTemplateName(formationTemplateName).
			WithAsserters(statusAsserter)

		// Assign both applications when the formation is still in INITIAL state and validate no notifications are sent and formation assignments are in INITIAL state
		t.Logf("Assign application 1 to formation: %s", formationName)
		assignApp1 := operations.NewAssignAppToFormationOperation(app1ID, tnt).
			WithFormationName(formationName).
			Operation()

		t.Logf("Assign application 2 to formation: %s", formationName)
		assignApp2 := operations.NewAssignAppToFormationOperation(app2ID, tnt).
			WithFormationName(formationName).
			Operation()

		lifecycleAsserter := asserters.NewLifecycleNotificationsAsserter(conf.ExternalServicesMockMtlsSecuredURL, certSecuredGraphQLClient, certSecuredHTTPClient).
			WithOperation(createFormationOperation).
			WithFormationName(formationName).
			WithState(initialAssignmentState).
			WithTenantID(tnt).
			WithParentTenantID(tntParentCustomer)
		op = operations.NewMultiOperation().
			WithOperation(createFormation).
			WithOperation(assignApp1).
			WithOperation(assignApp2).
			WithAsserters(lifecycleAsserter).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Update webhook with type %q and mode: %q for formation template with ID: %q", graphql.WebhookTypeFormationLifecycle, graphql.WebhookModeAsyncCallback, ftplID)
		op = operations.NewUpdateWebhookOperation().
			WithWebhookType(graphql.WebhookTypeFormationLifecycle).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate(`{\"path\":\"` + conf.ExternalServicesMockMtlsSecuredURL + `/v1/businessIntegration/async/{{.Formation.ID}}\",\"method\":\"{{if eq .Operation \"createFormation\"}}POST{{else}}DELETE{{end}}\"}`).
			WithInputTemplate(`{\"globalAccountId\":\"{{.CustomerTenantContext.AccountID}}\",\"crmId\":\"{{.CustomerTenantContext.CustomerID}}\",\"details\":{\"id\":\"{{.Formation.ID}}\",\"name\":\"{{.Formation.Name}}\"}}`).
			WithOutputTemplate(`{\"error\": \"{{.Body.error}}\",\"success_status_code\": 202}`).
			WithObjectType(operations.WebhookReferenceObjectTypeFormationTemplate).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Resynchronize formation")
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, fixtures.StatusAPIAsyncConfigJSON, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app1ID, app2ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "RESYNC", true),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
			WithFormationName(formationName)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).
			WithFormationName(formationName)
		notificationsAsserter := asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).
			WithUseItemsStruct(true).
			WithAssertTrustDetails(true).
			WithFormationName(formationName).
			WithGQLClient(certSecuredGraphQLClient).
			WithExpectedSubjects([]string{certSubjectMappingCustomSubjectWithCommaSeparator, certSubjectMappingCustomSubjectWithCommaSeparatorSecond})
		lifecycleAsserter = asserters.NewLifecycleNotificationsAsserter(conf.ExternalServicesMockMtlsSecuredURL, certSecuredGraphQLClient, certSecuredHTTPClient).
			WithOperation(createFormationOperation).
			WithFormationName(formationName).
			WithState("READY").
			WithTenantID(tnt).
			WithParentTenantID(tntParentCustomer)
		op = operations.NewResynchronizeFormationOperation().
			WithTenantID(tnt).
			WithFormationName(formationName).
			WithAsserters(faAsyncAsserter, statusAsserter, notificationsAsserter, lifecycleAsserter)
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID).WithOperations([]*fixtures.Operation{
			fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "RESYNC", true),
		})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
			WithFormationName(formationName)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).
			WithFormationName(formationName)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).
			WithFormationName(formationName).
			WithUseItemsStruct(true).
			WithGQLClient(certSecuredGraphQLClient)
		unassignNotificationsAsserter2 := asserters.NewUnassignNotificationsAsserter(0, app2ID, app1ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).
			WithFormationName(formationName).
			WithUseItemsStruct(true).
			WithGQLClient(certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).
			WithFormationName(formationName).
			WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter, unassignNotificationsAsserter2).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s again", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, fixtures.StatusAPIAsyncConfigJSON, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app1ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "RESYNC", true),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
			WithFormationName(formationName)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).
			WithFormationName(formationName)
		notificationsAsserter = asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).
			WithUseItemsStruct(true).
			WithAssertTrustDetails(true).
			WithFormationName(formationName).
			WithGQLClient(certSecuredGraphQLClient).
			WithExpectedSubjects([]string{certSubjectMappingCustomSubjectWithCommaSeparator, certSubjectMappingCustomSubjectWithCommaSeparatorSecond})
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).
			WithFormationName(formationName).
			WithAsserters(faAsyncAsserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
			WithFormationName(formationName)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).
			WithFormationName(formationName)
		unassignNotificationsAsserter = asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).
			WithFormationName(formationName).
			WithUseItemsStruct(true).
			WithGQLClient(certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).
			WithFormationName(formationName).
			WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		emptyExpectationsBuilder := mock_data.NewFAExpectationsBuilder()
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(emptyExpectationsBuilder.GetExpectations(), emptyExpectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
			WithFormationName(formationName)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).
			WithFormationName(formationName)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).
			WithFormationName(formationName).
			WithAsserters(faAsyncAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		lifecycleAsserter = asserters.NewLifecycleNotificationsAsserter(conf.ExternalServicesMockMtlsSecuredURL, certSecuredGraphQLClient, certSecuredHTTPClient).
			WithOperation(deleteFormationOperation).
			WithTenantID(tnt).
			WithParentTenantID(tntParentCustomer)
		formationIsDeletedAsserter := asserters.NewFormationIsDeletedAsserter(certSecuredGraphQLClient).
			WithTenantID(tnt)
		op = operations.NewDeleteFormationOperation(tnt).
			WithName(formationName).
			WithFormationTemplateName(formationTemplateName).
			WithAsserters(lifecycleAsserter, formationIsDeletedAsserter)
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Formation with name: %q is successfully deleted after READY status is reported on the status API", formationName)
	})

	t.Run("Draft Formation with Formation lifecycle asynchronous notifications and asynchronous app to app formation assignment notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Add webhook with type %q and mode: %q to application template with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, appTplID)
		op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplicationTemplate, appTplID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate(`{\"path\":\"` + conf.ExternalServicesMockMtlsSecuredURL + `/formation-callback/async-old/{{.TargetApplication.ID}}{{if eq .Operation \"unassign\"}}/{{.SourceApplication.ID}}{{end}}\",\"method\":\"{{if eq .Operation \"assign\"}}PATCH{{else}}DELETE{{end}}\"}`).
			WithInputTemplate(`{\"ucl-formation-id\":\"{{.FormationID}}\",\"operation-id\":\"{{.AssignmentOperation.ID}}\",\"globalAccountId\":\"{{.CustomerTenantContext.AccountID}}\",\"crmId\":\"{{.CustomerTenantContext.CustomerID}}\",\"formation-assignment-id\":\"{{ .Assignment.ID }}\", \"config\":{{ .ReverseAssignment.Value }},\"items\":[{\"region\":\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\",\"application-namespace\":\"{{.SourceApplicationTemplate.ApplicationNamespace}}\",\"tenant-id\":\"{{.SourceApplication.LocalTenantID}}\",\"ucl-system-tenant-id\":\"{{.SourceApplication.ID}}\",\"source-trust-details\":[{{ Join  .SourceApplicationTemplate.TrustDetails.Subjects }}],\"target-trust-details\":[{{ Join  .TargetApplicationTemplate.TrustDetails.Subjects }}] }]}`).
			WithOutputTemplate(`{\"config\":\"{{.Body.config}}\", \"location\":\"{{.Headers.Location}}\",\"error\": \"{{.Body.error}}\",\"success_status_code\": 202}`).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to formation template with ID %q", graphql.WebhookTypeFormationLifecycle, graphql.WebhookModeAsyncCallback, ftplID)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeFormationLifecycle, operations.WebhookReferenceObjectTypeFormationTemplate, ftplID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate(`{\"path\":\"` + conf.ExternalServicesMockMtlsSecuredURL + `/v1/businessIntegration/async/{{.Formation.ID}}\",\"method\":\"{{if eq .Operation \"createFormation\"}}POST{{else}}DELETE{{end}}\"}`).
			WithInputTemplate(`{\"globalAccountId\":\"{{.CustomerTenantContext.AccountID}}\",\"crmId\":\"{{.CustomerTenantContext.CustomerID}}\",\"details\":{\"id\":\"{{.Formation.ID}}\",\"name\":\"{{.Formation.Name}}\"}}`).
			WithOutputTemplate(`{\"error\": \"{{.Body.error}}\",\"success_status_code\": 202}`).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		formationName := "draft-formation-e2e-test"
		t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).
			WithFormationName(formationName).
			WithCondition(graphql.FormationStatusConditionDraft).
			WithState(draftFormationState)
		createFormation := operations.NewCreateFormationOperation(tnt).
			WithName(formationName).
			WithState(draftFormationState).
			WithFormationTemplateName(formationTemplateName).
			WithAsserters(statusAsserter)

		// Assign both applications when the formation is still in DRAFT state and validate no notifications are sent and formation assignments are in INITIAL state
		t.Logf("Assign application 1 to formation: %s", formationName)
		assignApp1 := operations.NewAssignAppToFormationOperation(app1ID, tnt).
			WithFormationName(formationName).
			Operation()

		t.Logf("Assign application 2 to formation: %s", formationName)
		assignApp2 := operations.NewAssignAppToFormationOperation(app2ID, tnt).
			WithFormationName(formationName).
			Operation()

		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipantAndStates(app1ID, initialAssignmentState, initialAssignmentState, initialAssignmentState).
			WithParticipantAndStates(app2ID, initialAssignmentState, initialAssignmentState, initialAssignmentState).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", false),
				fixtures.NewOperation(app1ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", false),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
			WithFormationName(formationName)
		lifecycleAsserter := asserters.NewLifecycleNotificationsAsserter(conf.ExternalServicesMockMtlsSecuredURL, certSecuredGraphQLClient, certSecuredHTTPClient).
			WithOperation(createFormationOperation).
			WithFormationName(formationName).
			WithExpectNotifications(false).
			WithTenantID(tnt).
			WithParentTenantID(tntParentCustomer)
		notificationsAsserter := asserters.NewNotificationsAsserter(0, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).
			WithUseItemsStruct(true).
			WithAssertTrustDetails(true).
			WithFormationName(formationName).
			WithGQLClient(certSecuredGraphQLClient)
		notificationsAsserter2 := asserters.NewNotificationsAsserter(0, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).
			WithUseItemsStruct(true).
			WithAssertTrustDetails(true).
			WithFormationName(formationName).
			WithGQLClient(certSecuredGraphQLClient)
		op = operations.NewMultiOperation().
			WithOperation(createFormation).
			WithOperation(assignApp1).
			WithOperation(assignApp2).
			WithAsserters(faAsyncAsserter, lifecycleAsserter, notificationsAsserter, notificationsAsserter2).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Finalize formation")
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, fixtures.StatusAPIAsyncConfigJSON, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app1ID, app2ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "RESYNC", true),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
			WithFormationName(formationName)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).
			WithFormationName(formationName)
		notificationsAsserter = asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).
			WithUseItemsStruct(true).
			WithAssertTrustDetails(true).
			WithFormationName(formationName).
			WithGQLClient(certSecuredGraphQLClient).
			WithExpectedSubjects([]string{certSubjectMappingCustomSubjectWithCommaSeparator, certSubjectMappingCustomSubjectWithCommaSeparatorSecond})
		lifecycleAsserter = asserters.NewLifecycleNotificationsAsserter(conf.ExternalServicesMockMtlsSecuredURL, certSecuredGraphQLClient, certSecuredHTTPClient).
			WithOperation(createFormationOperation).
			WithFormationName(formationName).
			WithState("READY").
			WithTenantID(tnt).
			WithParentTenantID(tntParentCustomer)
		op = operations.NewFinalizeFormationOperation().
			WithTenantID(tnt).
			WithFormationName(formationName).
			WithAsserters(faAsyncAsserter, statusAsserter, notificationsAsserter, lifecycleAsserter)
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Contains Scenario Groups Operator", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		scenarioGroup := "testScenarioGroup"
		scenarioGroup2 := "testScenarioGroup2"
		scenarioGroups := fmt.Sprintf(`{"key": "%s","description": "some description for key"}, {"key": "%s","description": "some description for key 2"}`, scenarioGroup, scenarioGroup2)

		op := operations.NewAddConstraintOperation("TestContainsScenarioGroupsFANotification").
			WithTargetOperation(graphql.TargetOperationGenerateFormationAssignmentNotification).
			WithOperator(formationconstraintpkg.ContainsScenarioGroups).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType1).
			WithInputTemplate(fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", scenarioGroup2)).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAddConstraintOperation("TestContainsScenarioGroupsAssign").
			WithTargetOperation(graphql.TargetOperationAssignFormation).
			WithOperator(formationconstraintpkg.ContainsScenarioGroups).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType2).
			WithInputTemplate(fmt.Sprintf("{\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"resource_id\\\": \\\"{{.ResourceID}}\\\",\\\"tenant\\\": \\\"{{.TenantID}}\\\", \\\"requiredScenarioGroups\\\": [\\\"%s\\\"]}", scenarioGroup)).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAssignAppToFormationErrorOperation(app2ID, tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewGenerateOnetimeTokenForApplicationOperation(app2ID, tnt).WithScenarioGroups(`{"key": "someOtherGroup", "description": "someOtherDescription"}`).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAssignAppToFormationErrorOperation(app2ID, tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewGenerateOnetimeTokenForApplicationOperation(app2ID, tnt).WithScenarioGroups(scenarioGroups).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type: %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app1ID)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/with-state/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		notificationsCountAsserter := asserters.NewNotificationsCountAsserter(0, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 := asserters.NewNotificationsCountAsserter(0, assignOperation, app2ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(notificationsCountAsserter, notificationsCountAsserter2).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewGenerateOnetimeTokenForApplicationOperation(app1ID, tnt).WithScenarioGroups(scenarioGroups).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		notificationsCountAsserter = asserters.NewNotificationsCountAsserter(1, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(notificationsCountAsserter, notificationsCountAsserter2).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Config mutator operator test - sync empty response", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		op := operations.NewAddConstraintOperation("mutate").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator(formationconstraintpkg.ConfigMutatorOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType1).
			WithInputTemplate(`{ \"tenant\":\"{{.Tenant}}\",\"source_resource_type\": \"{{.FormationAssignment.SourceType}}\",\"source_resource_id\": \"{{.FormationAssignment.Source}}\"{{if ne .NotificationStatusReport.State \"CREATE_ERROR\"}},\"modified_configuration\":\"{\\\"tmp\\\":\\\"{{.FormationAssignmentTemplateInput.SourceApplication.Application.LocalTenantID}}\\\"}\",\"state\":{{ $assignmentconfig := printf \"%s\" .FormationAssignment.Value }}{{if and (eq .FormationAssignment.State \"INITIAL\") (eq $assignmentconfig \"\")}}\"CONFIG_PENDING\"{{ else }}\"{{.NotificationStatusReport.State}}\"{{ end }}{{ end }},\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"operation\": \"{{.Operation}}\",{{ if .NotificationStatusReport }}\"notification_status_report_memory_address\":{{ .NotificationStatusReport.GetAddress }},{{ end }}\"join_point_location\": {\"OperationName\":\"{{.Location.OperationName}}\",\"ConstraintType\":\"{{.Location.ConstraintType}}\"}}`).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAddConstraintOperation("mutateTwo").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator(formationconstraintpkg.ConfigMutatorOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType2).
			WithInputTemplate(`{ \"tenant\":\"{{.Tenant}}\",\"only_for_source_subtypes\":[\"app-type-1\"],\"source_resource_type\": \"{{.FormationAssignment.SourceType}}\",\"source_resource_id\": \"{{.FormationAssignment.Source}}\"{{if ne .NotificationStatusReport.State \"CREATE_ERROR\"}},\"modified_configuration\": {{ $slice := mkslice \"{\\\"key\\\":\\\"key\\\",\\\"value\\\": \\\"example.com\\\"}\" \"{\\\"key\\\":\\\"ID\\\", \\\"value\\\":\\\"0000-0000000-0000\\\"}\" \"{\\\"key\\\":\\\"config\\\":\\\"value\\\":{\\\"clientID\\\":\\\"1111-11111\\\", \\\"clientSecret\\\":\\\"secret\\\"}}\" }} {{updateAndCopy .NotificationStatusReport.Configuration \"key2\" $slice}}{{if eq .FormationAssignment.State \"INITIAL\"}},\"state\":\"CONFIG_PENDING\"{{ end }}{{ end }},\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"operation\": \"{{.Operation}}\",{{ if .NotificationStatusReport }}\"notification_status_report_memory_address\":{{ .NotificationStatusReport.GetAddress }},{{ end }}\"join_point_location\": {\"OperationName\":\"{{.Location.OperationName}}\",\"ConstraintType\":\"{{.Location.ConstraintType}}\"}}`).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app1ID)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{if eq .Operation \\\"assign\\\"}}async-no-config{{else}}async{{end}}/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2ID)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app2ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/no-configuration/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		asserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation %s", formationName)
		expectedConfig := str.Ptr("{\"key\":\"example.com\",\"ID\":\"0000-0000000-0000\",\"config\":{\"clientID\":\"1111-11111\", \"clientSecret\":\"secret\"}}")
		expectedConfig2 := str.Ptr(fmt.Sprintf("{\"tmp\":\"%s\"}", localTenantID2))
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1ID, app2ID, readyAssignmentState, expectedConfig, nil),
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, expectedConfig2, nil),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsCountAsserter := asserters.NewNotificationsCountAsserter(2, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 := asserters.NewNotificationsCountAsserter(2, assignOperation, app2ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationsCountAsserter, notificationsCountAsserter2).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Update URL template for webhook with type %q and mode: %q for application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2ID)
		op = operations.NewUpdateWebhookOperation().
			WithWebhookType(graphql.WebhookTypeApplicationTenantMapping).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/fail/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}").
			WithObjectID(app2ID).
			WithObjectType(operations.WebhookReferenceObjectTypeApplication).
			WithTenantID(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s again", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1ID, app2ID, "CREATE_ERROR", nil, fixtures.StatusAPISyncErrorMessageJSON),
				mock_data.NewNotificationData(app2ID, app1ID, "CONFIG_PENDING", expectedConfig2, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app1ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", true),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).WithCondition(graphql.FormationStatusConditionError).WithErrors([]*graphql.FormationStatusError{fixtures.StatusAPISyncError})
		notificationsCountAsserter = asserters.NewNotificationsCountAsserter(3, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 = asserters.NewNotificationsCountAsserter(3, assignOperation, app2ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(notificationsCountAsserter, notificationsCountAsserter2, faAsyncAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Update URL template for webhook with type %q and mode: %q for application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2ID)
		op = operations.NewUpdateWebhookOperation().
			WithWebhookType(graphql.WebhookTypeApplicationTenantMapping).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}").
			WithObjectID(app2ID).
			WithObjectType(operations.WebhookReferenceObjectTypeApplication).
			WithTenantID(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Resynchronize formation")
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1ID, app2ID, readyAssignmentState, expectedConfig, nil),
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, expectedConfig2, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app1ID, app2ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", true),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsCountAsserter = asserters.NewNotificationsCountAsserter(2, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 = asserters.NewNotificationsCountAsserter(2, assignOperation, app2ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewResynchronizeFormationOperation().WithTenantID(tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationsCountAsserter, notificationsCountAsserter2)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter = asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		emptyExpectationsBuilder := mock_data.NewFAExpectationsBuilder()
		asserter = asserters.NewFormationAssignmentAsserter(emptyExpectationsBuilder.GetExpectations(), emptyExpectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Config mutator operator test - async empty response", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		op := operations.NewAddConstraintOperation("mutate").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator(formationconstraintpkg.ConfigMutatorOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType1).
			WithInputTemplate(`{ \"tenant\":\"{{.Tenant}}\",\"only_for_source_subtypes\":[\"app-type-2\"],\"source_resource_type\": \"{{.FormationAssignment.SourceType}}\",\"source_resource_id\": \"{{.FormationAssignment.Source}}\"{{if ne .NotificationStatusReport.State \"CREATE_ERROR\"}},\"modified_configuration\": {{ $slice := mkslice \"{\\\"key\\\":\\\"key\\\",\\\"value\\\": \\\"example.com\\\"}\" \"{\\\"key\\\":\\\"ID\\\", \\\"value\\\":\\\"0000-0000000-0000\\\"}\" \"{\\\"key\\\":\\\"config\\\":\\\"value\\\":{\\\"clientID\\\":\\\"1111-11111\\\", \\\"clientSecret\\\":\\\"secret\\\"}}\" }} {{updateAndCopy .NotificationStatusReport.Configuration \"asyncKey2\" $slice}}{{if eq .FormationAssignment.State \"INITIAL\"}},\"state\":\"CONFIG_PENDING\"{{ end }}{{ end }},\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"operation\": \"{{.Operation}}\",{{ if .NotificationStatusReport }}\"notification_status_report_memory_address\":{{ .NotificationStatusReport.GetAddress }},{{ end }}\"join_point_location\": {\"OperationName\":\"{{.Location.OperationName}}\",\"ConstraintType\":\"{{.Location.ConstraintType}}\"}}`).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAddConstraintOperation("mutateTwo").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator(formationconstraintpkg.ConfigMutatorOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType2).
			WithInputTemplate(`{ \"tenant\":\"{{.Tenant}}\",\"source_resource_type\": \"{{.FormationAssignment.SourceType}}\",\"source_resource_id\": \"{{.FormationAssignment.Source}}\"{{if ne .NotificationStatusReport.State \"CREATE_ERROR\"}},\"modified_configuration\":\"{\\\"tmp\\\":\\\"{{.FormationAssignmentTemplateInput.SourceApplication.Application.LocalTenantID}}\\\"}\",\"state\":{{ $assignmentconfig := printf \"%s\" .FormationAssignment.Value }}{{if and (eq .FormationAssignment.State \"INITIAL\") (eq $assignmentconfig \"\")}}\"CONFIG_PENDING\"{{ else }}\"{{.NotificationStatusReport.State}}\"{{ end }}{{ end }},\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"operation\": \"{{.Operation}}\",{{ if .NotificationStatusReport }}\"notification_status_report_memory_address\":{{ .NotificationStatusReport.GetAddress }},{{ end }}\"join_point_location\": {\"OperationName\":\"{{.Location.OperationName}}\",\"ConstraintType\":\"{{.Location.ConstraintType}}\"}}`).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2ID)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app2ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/no-configuration/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\",\\\"operationId\\\":\\\"{{.AssignmentOperation.ID}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		asserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation %s", formationName)
		expectedConfig := str.Ptr("{\"asyncNestedKey\": \"asyncNestedValue\", \"key\":\"example.com\",\"ID\":\"0000-0000000-0000\",\"config\":{\"clientID\":\"1111-11111\", \"clientSecret\":\"secret\"}}")
		expectedConfig2 := str.Ptr(fmt.Sprintf("{\"tmp\":\"%s\"}", localTenantID))
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1ID, app2ID, readyAssignmentState, expectedConfig2, nil),
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, expectedConfig, nil),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsCountAsserter := asserters.NewNotificationsCountAsserter(2, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 := asserters.NewNotificationsCountAsserter(2, assignOperation, app2ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationsCountAsserter, notificationsCountAsserter2).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Update URL template for webhook with type %q and mode: %q for application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op = operations.NewUpdateWebhookOperation().
			WithWebhookType(graphql.WebhookTypeApplicationTenantMapping).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-fail/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\",\\\"operationId\\\":\\\"{{.AssignmentOperation.ID}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").
			WithObjectID(app1ID).
			WithObjectType(operations.WebhookReferenceObjectTypeApplication).
			WithTenantID(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s again", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1ID, app2ID, "CONFIG_PENDING", expectedConfig2, nil),
				mock_data.NewNotificationData(app2ID, app1ID, "CREATE_ERROR", nil, fixtures.StatusAPIAsyncErrorMessageJSON),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app1ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", true),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).WithCondition(graphql.FormationStatusConditionError).WithErrors([]*graphql.FormationStatusError{fixtures.StatusAPIAsyncError})
		notificationsCountAsserter = asserters.NewNotificationsCountAsserter(3, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 = asserters.NewNotificationsCountAsserter(3, assignOperation, app2ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(notificationsCountAsserter, notificationsCountAsserter2, faAsyncAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Update URL template for webhook with type %q and mode: %q for application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op = operations.NewUpdateWebhookOperation().
			WithWebhookType(graphql.WebhookTypeApplicationTenantMapping).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\",\\\"operationId\\\":\\\"{{.AssignmentOperation.ID}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").
			WithObjectID(app1ID).
			WithObjectType(operations.WebhookReferenceObjectTypeApplication).
			WithTenantID(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Resynchronize formation")
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1ID, app2ID, readyAssignmentState, expectedConfig2, nil),
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, expectedConfig, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app1ID, app2ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "RESYNC", true),
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", true),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsCountAsserter = asserters.NewNotificationsCountAsserter(2, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 = asserters.NewNotificationsCountAsserter(2, assignOperation, app2ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewResynchronizeFormationOperation().WithTenantID(tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationsCountAsserter, notificationsCountAsserter2)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter = asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		emptyExpectationsBuilder := mock_data.NewFAExpectationsBuilder()
		asserter = asserters.NewFormationAssignmentAsserter(emptyExpectationsBuilder.GetExpectations(), emptyExpectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Formation assignment notifications validating redirect operator", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		webhookInputTemplate := "{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\",\\\"operationId\\\":\\\"{{.AssignmentOperation.ID}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}"
		t.Logf("Add webhook with type: %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2ID)
		op := operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app2ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/configuration/redirect-notification/{{.TargetApplication.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate(webhookInputTemplate).
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.configuration}}\\\", \\\"state\\\":\\\"{{.Body.state}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type: %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.TargetApplication.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate(webhookInputTemplate).
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.configuration}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		redirectedTntID := "custom-redirected-tenant-id"
		redirectPath := "/formation-callback/redirect-notification/" + redirectedTntID
		redirectURL := fmt.Sprintf("%s%s", conf.ExternalServicesMockMtlsSecuredURL, redirectPath)
		redirectURLTemplate := fmt.Sprintf("{\\\\\\\"path\\\\\\\":\\\\\\\"%s\\\\\\\",\\\\\\\"method\\\\\\\":\\\\\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\\\\\"}", redirectURL)
		redirectInputTemplate := fmt.Sprintf("{\\\"should_redirect\\\": {{ if contains .FormationAssignment.Value \\\"redirectProperties\\\" }}true{{else}}false{{end}},\\\"url_template\\\": \\\"%s\\\",\\\"url\\\": \\\"%s\\\",{{ if .Webhook }}\\\"webhook_memory_address\\\":{{ .Webhook.GetAddress }},{{ end }}\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}", redirectURLTemplate, redirectURL)

		// create redirect formation constraint and attach it to formation template
		op = operations.NewAddConstraintOperation("e2e-redirect-operator-constraint").
			WithTargetOperation(graphql.TargetOperationSendNotification).
			WithOperator(formationconstraintpkg.RedirectNotificationOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType2).
			WithInputTemplate(redirectInputTemplate).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		notificationsCountAsserterApp1Tenant := asserters.NewNotificationsCountAsserter(0, assignOperation, tnt, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserterApp2Tenant := asserters.NewNotificationsCountAsserter(0, assignOperation, localTenantID2, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter, notificationsCountAsserterApp1Tenant, notificationsCountAsserterApp2Tenant).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1ID, app2ID, readyAssignmentState, nil, nil),
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, fixtures.StatusAPIAsyncConfigJSON, nil),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)

		// Normally, the app 2 tenant should have two formation assignment notifications
		// but due to redirect operator, one of them is redirected to a different receiver(redirect tenant)
		notificationsCountAsserterApp2Tenant = asserters.NewNotificationsCountAsserter(1, assignOperation, localTenantID2, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserterRedirectTenant := asserters.NewNotificationsCountAsserter(1, assignOperation, redirectedTntID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserterApp1Tenant = asserters.NewNotificationsCountAsserter(1, assignOperation, localTenantID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)

		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationsCountAsserterApp2Tenant, notificationsCountAsserterRedirectTenant, notificationsCountAsserterApp1Tenant).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Deleting the stored notifications in the external services mock")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app1ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserterApp1Tenant := asserters.NewUnassignNotificationsAsserter(1, localTenantID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		unassignNotificationsAsserterApp2Tenant := asserters.NewUnassignNotificationsAsserter(1, localTenantID2, app1ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserterApp1Tenant, unassignNotificationsAsserterApp2Tenant).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		emptyExpectationsBuilder := mock_data.NewFAExpectationsBuilder()
		faAsserter = asserters.NewFormationAssignmentAsserter(emptyExpectationsBuilder.GetExpectations(), emptyExpectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})

	t.Run("Flow control operator test", func(t *testing.T) {
		redirectedTntID := "custom-redirected-tenant-id"
		redirectPath := "/formation-callback/async-no-response/" + redirectedTntID + "/" + app2ID
		redirectURL := fmt.Sprintf("%s%s", conf.ExternalServicesMockMtlsSecuredURL, redirectPath)
		redirectURLTemplate := fmt.Sprintf("{\\\\\\\"path\\\\\\\":\\\\\\\"%s\\\\\\\",\\\\\\\"method\\\\\\\":\\\\\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\\\\\"}", redirectURL)
		op := operations.NewAddConstraintOperation("e2e-flow-control-operator-constraint-status-returned").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator(formationconstraintpkg.AsynchronousFlowControlOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType1).
			WithInputTemplate(fmt.Sprintf("{\\\"url_template\\\": \\\"%s\\\",\\\"url\\\": \\\"%s\\\",{{ if .NotificationStatusReport }}\\\"notification_status_report_memory_address\\\":{{ .NotificationStatusReport.GetAddress }},{{ end }}{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}", redirectURLTemplate, redirectURL)).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAddConstraintOperation("e2e-flow-control-operator-constraint-send-notification").
			WithTargetOperation(graphql.TargetOperationSendNotification).
			WithOperator(formationconstraintpkg.AsynchronousFlowControlOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType1).
			WithInputTemplate(fmt.Sprintf("{\\\"url_template\\\": \\\"%s\\\",\\\"url\\\": \\\"%s\\\",{{ if .Webhook }}\\\"webhook_memory_address\\\":{{ .Webhook.GetAddress }},{{ end }}{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}", redirectURLTemplate, redirectURL)).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type: %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app1ID)
		op = operations.NewAddWebhookToObjectOperation(graphql.WebhookTypeApplicationTenantMapping, operations.WebhookReferenceObjectTypeApplication, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-old/{{.TargetApplication.LocalTenantID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\",\\\"operationId\\\":\\\"{{.AssignmentOperation.ID}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.configuration}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		asserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, fixtures.StatusAPIAsyncConfigJSON, nil),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsAsserter := asserters.NewNotificationsAsserter(1, assignOperation, localTenantID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, &emptyConfig, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign application 2 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithCustomParticipants([]string{app1ID, app2ID}).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, deletingAssignmentState, nil, nil),
				mock_data.NewNotificationData(app1ID, app1ID, readyAssignmentState, nil, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app1ID, "UNASSIGN", "UNASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app1ID, "INSTANCE_CREATOR_UNASSIGN", "UNASSIGN_OBJECT", false),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).WithCondition(graphql.FormationStatusConditionInProgress)
		notificationCountAsyncAsserter := asserters.NewNotificationsCountAsyncAsserter(1, unassignOperation, localTenantID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsUnassignAsserter := asserters.NewUnassignNotificationsAsserter(1, localTenantID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationCountRedirectNotificationAsyncAsserter := asserters.NewNotificationsCountAsyncAsserter(1, unassignOperation, redirectedTntID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsUnassignRedirectedAsserter := asserters.NewUnassignNotificationsAsserter(1, redirectedTntID, app2ID, localTenantID, appNamespace, appRegion, tnt, tntParentCustomer, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient).WithState(deletingAssignmentState)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationCountAsyncAsserter, notificationsUnassignAsserter, notificationCountRedirectNotificationAsyncAsserter, notificationsUnassignRedirectedAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign application 2 from formation: %s should only send notification to instance creator", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithCustomParticipants([]string{app1ID, app2ID}).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, deletingAssignmentState, nil, nil),
				mock_data.NewNotificationData(app1ID, app1ID, readyAssignmentState, nil, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app1ID, "UNASSIGN", "UNASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app1ID, "INSTANCE_CREATOR_UNASSIGN", "UNASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app1ID, "UNASSIGN", "UNASSIGN_OBJECT", false),
				fixtures.NewOperation(app2ID, app1ID, "INSTANCE_CREATOR_UNASSIGN", "UNASSIGN_OBJECT", false),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		op = operations.NewUnassignAppFromFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationCountRedirectNotificationAsyncAsserter, notificationsUnassignRedirectedAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Cleanup notifications")
		op = operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewExecuteStatusReportOperation().WithTenant(tnt).
			WithFormationAssignment(app2ID, app1ID).
			WithHTTPClient(appCertClient).
			WithExternalServicesMockMtlsSecuredURL(conf.DirectorExternalCertFAAsyncStatusURL).
			WithAsserters(faAsyncAsserter, statusAsserter).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)

		op = operations.NewExecuteStatusReportOperation().WithTenant(tnt).
			WithFormationAssignment(app2ID, app1ID).
			WithHTTPClient(instanceCreatorCertClient).
			WithExternalServicesMockMtlsSecuredURL(conf.DirectorExternalCertFAAsyncStatusURL).
			WithAsserters(faAsyncAsserter, statusAsserter).
			Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder()
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppFromFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})
}
