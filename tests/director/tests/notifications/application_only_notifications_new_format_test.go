package notifications

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	mock_data "github.com/kyma-incubator/compass/tests/pkg/notifications/expectations-builders"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/operations"
	resource_providers "github.com/kyma-incubator/compass/tests/pkg/notifications/resource-providers"
	"strings"
	"testing"
	"time"

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

	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	formationTemplateName := "app-only-formation-template-name"

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
	applicationType1 := "app-type-1"

	t.Logf("Create application template for type: %q", applicationType1)
	appTemplateProvider := resource_providers.NewApplicationTemplateProvider(applicationType1, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder, tnt, nil)
	defer appTemplateProvider.Cleanup(t, ctx, oauthGraphQLClient)
	appTplID := appTemplateProvider.Provide(t, ctx, oauthGraphQLClient)
	internalConsumerID := appTplID // add application templated ID as certificate subject mapping internal consumer to satisfy the authorization checks in the formation assignment status API

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
	t.Logf("Create application template for type %q", applicationType2)
	appTemplateProvider2 := resource_providers.NewApplicationTemplateProvider(applicationType2, localTenantID2, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder, tnt, nil)
	defer appTemplateProvider2.Cleanup(t, ctx, oauthGraphQLClient)
	appTemplateProvider2.Provide(t, ctx, oauthGraphQLClient)

	ftProvider := resource_providers.NewFormationTemplateCreator(formationTemplateName)
	defer ftProvider.Cleanup(t, ctx, certSecuredGraphQLClient)
	ftplID := ftProvider.WithSupportedResources(appTemplateProvider.GetResource(), appTemplateProvider2.GetResource()).WithLeadingProductIDs([]string{internalConsumerID}).Provide(t, ctx, certSecuredGraphQLClient)
	ctx = context.WithValue(ctx, context_keys.FormationTemplateIDKey, ftplID)

	t.Logf("Create application 1 from template %q", applicationType1)
	appProvider1 := resource_providers.NewApplicationProvider(applicationType1, namePlaceholder, "app1-formation-notifications-tests", displayNamePlaceholder, "App 1 Display Name", tnt)
	defer appProvider1.Cleanup(t, ctx, certSecuredGraphQLClient)
	app1ID := appProvider1.Provide(t, ctx, certSecuredGraphQLClient)

	t.Logf("Create application 2 from template %q", applicationType2)
	appProvider2 := resource_providers.NewApplicationProvider(applicationType2, namePlaceholder, "app2-formation-notifications-tests", displayNamePlaceholder, "App 2 Display Name", tnt)
	defer appProvider2.Cleanup(t, ctx, certSecuredGraphQLClient)
	app2ID := appProvider2.Provide(t, ctx, certSecuredGraphQLClient)

	formationName := "app-to-app-formation-name"
	t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTemplateName)
	formationProvider := resource_providers.NewFormationProvider(formationName, tnt, &formationTemplateName)
	defer formationProvider.Cleanup(t, ctx, certSecuredGraphQLClient)
	formationID := formationProvider.Provide(t, ctx, certSecuredGraphQLClient)
	ctx = context.WithValue(ctx, context_keys.FormationIDKey, formationID)
	ctx = context.WithValue(ctx, context_keys.FormationNameKey, formationName)

	expectationsBuilder := mock_data.NewFAExpectationsBuilder()
	asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
		AssertExpectations(t, ctx)
	asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient).AssertExpectations(t, ctx)

	t.Run("Synchronous App to App Formation Assignment Notifications", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app1ID)
		op := operations.NewAddWebhookToApplicationOperation(graphql.WebhookTypeApplicationTenantMapping, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}").Operation()
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
		expectedConfig := str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, "READY", expectedConfig, nil),
			})
		faAsserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsAsserter := asserters.NewNotificationsAsserter(1, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s again", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app2ID, app1ID, "READY", expectedConfig, nil),
			})
		asserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsAsserter = asserters.NewNotificationsAsserter(3, assignOperation, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter, notificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID)
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter = asserters.NewUnassignNotificationsAsserter(2, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder()
		faAsserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})
	t.Run("Config mutator operator test", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		op := operations.NewAddConstraintOperation("mutate").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator("ConfigMutator").
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType1).
			WithInputTemplate(`{ \"tenant\":\"{{.Tenant}}\", \"last_formation_assignment_state\":\"{{.LastFormationAssignmentState}}\"{{if ne .FormationAssignment.State \"CREATE_ERROR\"}},\"modified_configuration\":\"{\\\"tmp\\\":\\\"{{.FormationAssignmentTemplateInput.SourceApplication.Application.LocalTenantID}}\\\"}\",\"state\":{{if and (eq .LastFormationAssignmentState \"INITIAL\") (eq .LastFormationAssignmentConfiguration \"\")}}\"CONFIG_PENDING\"{{ else }}\"{{.FormationAssignment.State}}\"{{ end }}{{ end }},\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"operation\": \"{{.Operation}}\",{{ if .FormationAssignment }}\"details_formation_assignment_memory_address\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\"details_reverse_formation_assignment_memory_address\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\"join_point_location\": {\"OperationName\":\"{{.Location.OperationName}}\",\"ConstraintType\":\"{{.Location.ConstraintType}}\"}}`).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		op = operations.NewAddConstraintOperation("mutateTwo").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator("ConfigMutator").
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType2).
			WithInputTemplate(`{ \"tenant\":\"{{.Tenant}}\",\"only_for_source_subtypes\":[\"app-type-1\"],\"last_formation_assignment_state\":\"{{.LastFormationAssignmentState}}\"{{if ne .FormationAssignment.State \"CREATE_ERROR\"}},\"modified_configuration\": {{ $slice := mkslice \"{\\\"key\\\":\\\"key\\\",\\\"value\\\": \\\"example.com\\\"}\" \"{\\\"key\\\":\\\"ID\\\", \\\"value\\\":\\\"0000-0000000-0000\\\"}\" \"{\\\"key\\\":\\\"config\\\":\\\"value\\\":{\\\"clientID\\\":\\\"1111-11111\\\", \\\"clientSecret\\\":\\\"secret\\\"}}\" }} {{updateAndCopy .FormationAssignment.Value \"key2\" $slice}}{{if eq .LastFormationAssignmentState \"INITIAL\"}},\"state\":\"CONFIG_PENDING\"{{ end }}{{ end }},\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"operation\": \"{{.Operation}}\",{{ if .FormationAssignment }}\"details_formation_assignment_memory_address\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\"details_reverse_formation_assignment_memory_address\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\"join_point_location\": {\"OperationName\":\"{{.Location.OperationName}}\",\"ConstraintType\":\"{{.Location.ConstraintType}}\"}}`).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		// EXPECTED BEHAVIOUR
		// Case SYNC notification is executed first
		// Assignment1 State: INITIAL -> Execute SYNC notification -> Response { READY, someCfg } -> Set Assignment1 to { CONFIG_PENDING, mutated_configuration
		// Assignment2 State: INITIAL -> Execute ASYNC notification
		//
		// Process ASYNC status update with Response { READY, someCfg }: Assignment2 State: INITIAL ->  Set Assignment2 to { CONFIG_PENDING, mutated_configuration }
		// Assignment1 State: CONFIG_PENDING -> Execute SYNC notification -> Response { READY, someCfg } -> Set Assignment1 to { READY, mutated_configuration }
		// Assignment2 State: CONFIG_PENDING -> Execute ASYNC notification
		//
		// Process ASYNC status update with Response { READY, someCfg }: Assignment2 State: CONFIG_PENDING ->  Set Assignment2 to { READY, mutated_configuration }

		// Case ASYNC notification is executed first
		// Assignment2 State: INITIAL -> Execute ASYNC notification
		// Assignment1 State: INITIAL -> Execute SYNC notification -> Response { READY, someCfg } -> Set Assignment1 to { CONFIG_PENDING, mutated_configuration
		// Assignment2 State: INITIAL -> Execute ASYNC notification
		//
		// The two ASYNC status updates will be processed almost simultaneously
		//
		// Process ASYNC status update with Response { READY, someCfg }: Assignment2 State: INITIAL ->  Set Assignment2 to { CONFIG_PENDING, mutated_configuration }
		// Assignment1 State: CONFIG_PENDING -> Execute SYNC notification -> Response { READY, someCfg } -> Set Assignment1 to { READY, mutated_configuration }
		// Assignment2 State: CONFIG_PENDING -> Execute ASYNC notification
		// Process ASYNC status update with Response { READY, someCfg }: Assignment2 State: CONFIG_PENDING ->  Set Assignment2 to { READY, mutated_configuration }
		// Assignment1 State: CONFIG_PENDING -> Execute SYNC notification -> Response { READY, someCfg } -> Set Assignment1 to { READY, mutated_configuration }
		// Assignment2 State: READY
		//
		// Process ASYNC status update with Response { READY, someCfg }: Assignment2 State: CONFIG_PENDING ->  Set Assignment2 to { READY, mutated_configuration }

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2ID)
		op = operations.NewAddWebhookToApplicationOperation(graphql.WebhookTypeApplicationTenantMapping, app2ID, tnt).
			WithWebhookMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op = operations.NewAddWebhookToApplicationOperation(graphql.WebhookTypeApplicationTenantMapping, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").Operation()
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
				mock_data.NewNotificationData(app1ID, app2ID, "READY", expectedConfig, nil),
				mock_data.NewNotificationData(app2ID, app1ID, "READY", expectedConfig2, nil),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, (conf.TenantMappingAsyncResponseDelay+5)*2)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsCountAsserter := asserters.NewNotificationsCountAsserter(2, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 := asserters.NewNotificationsCountAsserter(2, assignOperation, app2ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, notificationsCountAsserter, notificationsCountAsserter2).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, conf.TenantMappingAsyncResponseDelay+5)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppToFormationOperation(app1ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Update URL template for webhook with type %q and mode: %q for application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op = operations.NewUpdateWebhookOperation().
			WithWebhookType(graphql.WebhookTypeApplicationTenantMapping).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async-fail/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").
			WithApplicationID(app1ID).
			WithTenantID(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s again", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1ID, app2ID, "CONFIG_PENDING", expectedConfig, nil),
				mock_data.NewNotificationData(app2ID, app1ID, "CREATE_ERROR", nil, fixtures.StatusAPIAsyncErrorMessageJSON),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, conf.TenantMappingAsyncResponseDelay+5)
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
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").
			WithApplicationID(app1ID).
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
				mock_data.NewNotificationData(app1ID, app2ID, "READY", expectedConfig, nil),
				mock_data.NewNotificationData(app2ID, app1ID, "READY", expectedConfig2, nil),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, (conf.TenantMappingAsyncResponseDelay+5)*2)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		notificationsCountAsserter = asserters.NewNotificationsCountAsserter(2, assignOperation, app1ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 = asserters.NewNotificationsCountAsserter(2, assignOperation, app2ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewResynchronizeFormationOperation().WithTenantID(tnt).WithAsserters()

		t.Logf("Unassign Application 2 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, conf.TenantMappingAsyncResponseDelay+5)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter = asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppToFormationOperation(app2ID, tnt).WithAsserters(faAsyncAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder()
		asserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppToFormationOperation(app1ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})
}
