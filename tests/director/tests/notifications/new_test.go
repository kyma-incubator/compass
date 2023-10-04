package notifications

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	mock_data "github.com/kyma-incubator/compass/tests/pkg/mock-data"
	"github.com/kyma-incubator/compass/tests/pkg/operations"
	"github.com/kyma-incubator/compass/tests/pkg/resource-providers"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestNewFormat(t *testing.T) {
	ctx := context.Background()
	tnt := tenant.TestTenants.GetDefaultTenantID()
	tntParentCustomer := tenant.TestTenants.GetIDByName(t, tenant.TestDefaultCustomerTenant) // parent of `tenant.TestTenants.GetDefaultTenantID()` above

	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	formationTmplName := "app-only-formation-template-name"

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
	provider := resource_providers.NewApplicationTemplateProvider(applicationType1, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder, nil)
	defer provider.TearDown(t, ctx, oauthGraphQLClient, "")
	appTmpl := provider.Provide(t, ctx, oauthGraphQLClient, "")
	internalConsumerID := appTmpl.GetID() // add application templated ID as certificate subject mapping internal consumer to satisfy the authorization checks in the formation assignment status API

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
	provider2 := resource_providers.NewApplicationTemplateProvider(applicationType2, localTenantID2, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder, nil)
	defer provider2.TearDown(t, ctx, oauthGraphQLClient, "")
	appTpl2 := provider2.Provide(t, ctx, oauthGraphQLClient, "")

	ftProvider := resource_providers.NewFormationTemplateCreator(formationTmplName)
	defer ftProvider.TearDown(t, ctx, certSecuredGraphQLClient)
	ftplID := ftProvider.WithParticipant(appTmpl).WithParticipant(appTpl2).WithLeadingProductIDs([]string{internalConsumerID}).Provide(t, ctx, certSecuredGraphQLClient)

	t.Logf("Create application 1 from template %q", applicationType1)
	appProvider1 := resource_providers.NewApplicationProvider(applicationType1, namePlaceholder, "app1-formation-notifications-tests", displayNamePlaceholder, "App 1 Display Name")
	defer appProvider1.TearDown(t, ctx, certSecuredGraphQLClient, tnt)
	app1 := appProvider1.Provide(t, ctx, certSecuredGraphQLClient, tnt)

	t.Logf("Create application 2 from template %q", applicationType2)
	appProvider2 := resource_providers.NewApplicationProvider(applicationType2, namePlaceholder, "app2-formation-notifications-tests", displayNamePlaceholder, "App 2 Display Name")
	defer appProvider2.TearDown(t, ctx, certSecuredGraphQLClient, tnt)
	app2 := appProvider2.Provide(t, ctx, certSecuredGraphQLClient, tnt)

	formationName := "app-to-app-formation-name"
	t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
	createFormationOp := operations.NewCreateFormationOperation(formationName, tnt, &formationTmplName)
	defer createFormationOp.Cleanup(t, ctx, certSecuredGraphQLClient)
	createFormationOp.Execute(t, ctx, certSecuredGraphQLClient)
	formationID := createFormationOp.GetFormationID()
	expectationsBuilder := mock_data.NewFANNotificationExpectationsBuilder()
	asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID).
		AssertExpectations(t, ctx)
	asserters.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient).AssertExpectations(t, ctx)

	//t.Run("Synchronous App to App Formation Assignment Notifications demo", func(t *testing.T) {
	//	cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	//	defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
	//
	//	formationName := "app-to-app-formation-name"
	//	t.Logf("Creating formation with name: %q from template with name: %q", formationName, formationTmplName)
	//	createFormationOp := operations.NewCreateFormationOperation(formationName, tnt, &formationTmplName)
	//	defer createFormationOp.Cleanup(t, ctx, certSecuredGraphQLClient)
	//	createFormationOp.Execute(t, ctx, certSecuredGraphQLClient)
	//	formationID := createFormationOp.GetFormationID()
	//
	//	t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app1.ID)
	//	op := operations.NewAddWebhookToApplicationOperation(graphql.WebhookTypeApplicationTenantMapping, app1.ID, tnt).
	//		WithMode(graphql.WebhookModeSync).
	//		WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
	//		WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
	//		WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}")
	//	defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
	//	op.Execute(t, ctx, certSecuredGraphQLClient)
	//
	//	expectationsBuilder := mock_data.NewFANNotificationExpectationsBuilder()
	//	pocsetup.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID).
	//		AssertExpectations(t, ctx)
	//	pocsetup.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient).AssertExpectations(t, ctx)
	//
	//	t.Logf("Assign application 1 to formation %s", formationName)
	//	expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().WithParticipant(app1.ID)
	//	asserter := pocsetup.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID)
	//	statusAsserter := pocsetup.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
	//	asOp := operations.NewAssignAppToFormationOperation(formationName, app1.ID, tnt).WithAsserter(asserter).WithAsserter(statusAsserter)
	//	defer asOp.Cleanup(t, ctx, certSecuredGraphQLClient)
	//	asOp.Execute(t, ctx, certSecuredGraphQLClient)
	//
	//	t.Logf("Assign application 2 to formation %s", formationName)
	//	expectedConfig := str.Ptr("{\"key\":\"value\",\"key2\":{\"key\":\"value2\"}}")
	//	expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().
	//		WithParticipant(app1.ID).
	//		WithParticipant(app2.ID).
	//		WithNotifications([]*mock_data.NotificationData{
	//			mock_data.NewNotificationData(app1.ID, app2.ID, "READY", expectedConfig, nil),
	//		})
	//	faAsserter := pocsetup.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID)
	//	statusAsserter = pocsetup.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
	//	notificationsAsserter := pocsetup.NewNotificationsAsserter(1, assignOperation, formationID, app1.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
	//	as2Op := operations.NewAssignAppToFormationOperation(formationName, app2.ID, tnt).WithAsserter(faAsserter).WithAsserter(statusAsserter).WithAsserter(notificationsAsserter)
	//	defer as2Op.Cleanup(t, ctx, certSecuredGraphQLClient)
	//	as2Op.Execute(t, ctx, certSecuredGraphQLClient)
	//
	//	t.Logf("Unassign Application 1 from formation %s", formationName)
	//	expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().WithParticipant(app2.ID)
	//	faAsserter = pocsetup.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID)
	//	statusAsserter = pocsetup.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
	//	unassignNotificationsAsserter := pocsetup.NewUnassignNotificationsAsserter(2, unassignOperation, 1, formationID, app1.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
	//	unasOP := operations.NewUnassignAppToFormationOperation(formationName, app1.ID, tnt).WithAsserter(faAsserter).WithAsserter(statusAsserter).WithAsserter(unassignNotificationsAsserter)
	//	defer unasOP.Cleanup(t, ctx, certSecuredGraphQLClient)
	//	unasOP.Execute(t, ctx, certSecuredGraphQLClient)
	//
	//	t.Logf("Assign application 1 to formation %s again", formationName)
	//	expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().
	//		WithParticipant(app1.ID).
	//		WithParticipant(app2.ID).
	//		WithNotifications([]*mock_data.NotificationData{
	//			mock_data.NewNotificationData(app1.ID, app2.ID, "READY", expectedConfig, nil),
	//		})
	//	asserter = pocsetup.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID)
	//	statusAsserter = pocsetup.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
	//	notificationsAsserter = pocsetup.NewNotificationsAsserter(3, assignOperation, formationID, app1.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
	//	assignOp := operations.NewAssignAppToFormationOperation(formationName, app1.ID, tnt).WithAsserter(asserter).WithAsserter(statusAsserter).WithAsserter(notificationsAsserter)
	//	defer assignOp.Cleanup(t, ctx, certSecuredGraphQLClient)
	//	assignOp.Execute(t, ctx, certSecuredGraphQLClient)
	//
	//	t.Logf("Unassign Application 2 from formation %s", formationName)
	//	expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().
	//		WithParticipant(app1.ID)
	//	faAsserter = pocsetup.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID)
	//	statusAsserter = pocsetup.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
	//	unassignNotificationsAsserter = pocsetup.NewUnassignNotificationsAsserter(4, unassignOperation, 2, formationID, app1.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
	//	unass2Op := operations.NewUnassignAppToFormationOperation(formationName, app2.ID, tnt).WithAsserter(faAsserter).WithAsserter(statusAsserter).WithAsserter(unassignNotificationsAsserter)
	//	defer unass2Op.Cleanup(t, ctx, certSecuredGraphQLClient)
	//	unass2Op.Execute(t, ctx, certSecuredGraphQLClient)
	//
	//	t.Logf("Unassign Application 1 from formation %s", formationName)
	//	expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder()
	//	faAsserter = pocsetup.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID)
	//	statusAsserter = pocsetup.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
	//	unas3Op := operations.NewUnassignAppToFormationOperation(formationName, app1.ID, tnt).WithAsserter(faAsserter).WithAsserter(statusAsserter)
	//	defer unas3Op.Cleanup(t, ctx, certSecuredGraphQLClient)
	//	unas3Op.Execute(t, ctx, certSecuredGraphQLClient)
	//})
	t.Run("Config mutator operator test", func(t *testing.T) {
		cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)
		defer cleanupNotificationsFromExternalSvcMock(t, certSecuredHTTPClient)

		//constraint := fixtures.CreateFormationConstraint(t, ctx, certSecuredGraphQLClient, in)
		inputTpl := "{ \\\"last_formation_assignment_state\\\":\\\"{{.LastFormationAssignmentState}}\\\",\\\"configuration\\\":\\\"{\\\\\\\"tmp\\\\\\\":\\\\\\\"tmpval\\\\\\\"}\\\",\\\"state\\\":{{if eq .LastFormationAssignmentState \\\"INITIAL\\\"}}\\\"CONFIG_PENDING\\\"{{ else }}\\\"{{.FormationAssignment.State}}\\\"{{ end }},\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .FormationAssignment }}\\\"details_formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"details_reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}"
		op := operations.NewAddConstraintOperation("mutate", graphql.ConstraintTypePre, graphql.TargetOperationNotificationStatusReturned, "ConfigMutator", graphql.ResourceTypeApplication, applicationType1, inputTpl, graphql.ConstraintScopeFormationType, ftplID, tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		inputTpl = "{ \\\"last_formation_assignment_state\\\":\\\"{{.LastFormationAssignmentState}}\\\",\\\"configuration\\\":\\\"{\\\\\\\"tmp\\\\\\\":\\\\\\\"tmpvalll\\\\\\\"}\\\",\\\"state\\\":{{if eq .LastFormationAssignmentState \\\"INITIAL\\\"}}\\\"CONFIG_PENDING\\\"{{ else }}\\\"{{.FormationAssignment.State}}\\\"{{ end }},\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",{{ if .FormationAssignment }}\\\"details_formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"details_reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}"
		op = operations.NewAddConstraintOperation("mutateTwo", graphql.ConstraintTypePre, graphql.TargetOperationNotificationStatusReturned, "ConfigMutator", graphql.ResourceTypeApplication, applicationType2, inputTpl, graphql.ConstraintScopeFormationType, ftplID, tnt).Operation()
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

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeSync, app2.ID)
		op = operations.NewAddWebhookToApplicationOperation(graphql.WebhookTypeApplicationTenantMapping, app2.ID, tnt).
			WithMode(graphql.WebhookModeSync).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\", \\\"config\\\":{{ .ReverseAssignment.Value }},\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\"{{ if .SourceApplicationTemplate.Labels.composite }},\\\"composite-label\\\":{{.SourceApplicationTemplate.Labels.composite}}{{end}},\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 200, \\\"incomplete_status_code\\\": 204}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1.ID)
		op = operations.NewAddWebhookToApplicationOperation(graphql.WebhookTypeApplicationTenantMapping, app1.ID, tnt).
			WithMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"ucl-formation-id\\\":\\\"{{.FormationID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"formation-assignment-id\\\":\\\"{{ .Assignment.ID }}\\\",\\\"items\\\":[{\\\"region\\\":\\\"{{ if .SourceApplication.Labels.region }}{{.SourceApplication.Labels.region}}{{ else }}{{.SourceApplicationTemplate.Labels.region}}{{ end }}\\\",\\\"application-namespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"tenant-id\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"ucl-system-tenant-id\\\":\\\"{{.SourceApplication.ID}}\\\"}]}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.Config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s", formationName)
		expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().WithParticipant(app1.ID)
		asserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID)
		statusAsserter := asserters.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(formationName, app1.ID, tnt).WithAsserter(asserter).WithAsserter(statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation %s", formationName)
		expectedConfig := str.Ptr("{\"tmp\":\"tmpvalll\"}")
		expectedConf2 := str.Ptr("{\"tmp\":\"tmpval\"}")
		expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().
			WithParticipant(app1.ID).
			WithParticipant(app2.ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1.ID, app2.ID, "READY", expectedConf2, nil),
				mock_data.NewNotificationData(app2.ID, app1.ID, "READY", expectedConfig, nil),
			})
		faAsyncAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID, (conf.TenantMappingAsyncResponseDelay+5)*2)
		statusAsserter = asserters.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
		notificationsCountAsserter := asserters.NewNotificationsCountAsserter(2, assignOperation, app1.ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 := asserters.NewNotificationsCountAsserter(2, assignOperation, app2.ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(formationName, app2.ID, tnt).WithAsserter(faAsyncAsserter).WithAsserter(statusAsserter).WithAsserter(notificationsCountAsserter).WithAsserter(notificationsCountAsserter2).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().WithParticipant(app2.ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID, conf.TenantMappingAsyncResponseDelay+5)
		statusAsserter = asserters.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(unassignOperation, 1, formationID, app1.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppToFormationOperation(formationName, app1.ID, tnt).WithAsserter(faAsyncAsserter).WithAsserter(statusAsserter).WithAsserter(unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation %s again", formationName)
		expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().
			WithParticipant(app1.ID).
			WithParticipant(app2.ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1.ID, app2.ID, "READY", expectedConf2, nil),
				mock_data.NewNotificationData(app2.ID, app1.ID, "READY", expectedConfig, nil),
			})
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID, conf.TenantMappingAsyncResponseDelay+5)
		statusAsserter = asserters.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
		notificationsCountAsserter = asserters.NewNotificationsCountAsserter(4, assignOperation, app1.ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		notificationsCountAsserter2 = asserters.NewNotificationsCountAsserter(4, assignOperation, app2.ID, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewAssignAppToFormationOperation(formationName, app1.ID, tnt).WithAsserter(faAsyncAsserter).WithAsserter(statusAsserter).WithAsserter(notificationsCountAsserter).WithAsserter(notificationsCountAsserter2).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder().
			WithParticipant(app1.ID)
		faAsyncAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID, conf.TenantMappingAsyncResponseDelay+5)
		statusAsserter = asserters.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter = asserters.NewUnassignNotificationsAsserter(unassignOperation, 2, formationID, app1.ID, app2.ID, localTenantID2, appNamespace, appRegion, tnt, tntParentCustomer, conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppToFormationOperation(formationName, app2.ID, tnt).WithAsserter(faAsyncAsserter).WithAsserter(statusAsserter).WithAsserter(unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation %s", formationName)
		expectationsBuilder = mock_data.NewFANNotificationExpectationsBuilder()
		asserter = asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt, formationID)
		statusAsserter = asserters.NewFormationStatusAsserter(formationID, tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppToFormationOperation(formationName, app1.ID, tnt).WithAsserter(asserter).WithAsserter(statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})
}
