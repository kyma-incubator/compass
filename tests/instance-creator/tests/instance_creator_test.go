package tests

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
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/claims"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	test_json "github.com/kyma-incubator/compass/tests/pkg/json"
	"github.com/kyma-incubator/compass/tests/pkg/k8s"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	mock_data "github.com/kyma-incubator/compass/tests/pkg/notifications/expectations-builders"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/operations"
	resource_providers "github.com/kyma-incubator/compass/tests/pkg/notifications/resource-providers"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	consumerType          = "Integration System" // should be a valid consumer type
	readyAssignmentState  = "READY"
	emptyParentCustomerID = ""

	uriKey      = "uri"
	usernameKey = "username"

	eventuallyTimeoutForInstances = 60 * time.Second
	eventuallyTickForInstances    = 2 * time.Second
)

var (
	tenantAccessLevels = []string{"account", "global"} // should be a valid tenant access level
)

func TestInstanceCreator(t *testing.T) {
	ctx := context.Background()
	tnt := conf.TestConsumerAccountID

	certSecuredHTTPClient := fixtures.FixCertSecuredHTTPClient(cc, conf.ExternalClientCertSecretName, conf.SkipSSLValidation)

	subscriptionProviderSubaccountID := conf.TestProviderSubaccountID
	subscriptionConsumerSubaccountID := conf.TestConsumerSubaccountID
	subscriptionConsumerTenantID := conf.TestConsumerTenantID

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: conf.SkipSSLValidation},
		},
	}

	formationTemplateName := "instance-creator-formation-template-name"

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
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, false)
	appProviderDirectorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)

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
	appRegion := conf.InstanceCreatorRegion
	appNamespace := "compass.test"
	localTenantID := "local-tenant-id"
	applicationType1 := "app-type-1"
	app1BaseURL := "http://e2e-test-app1-base-url"

	t.Logf("Create application template for type: %q", applicationType1)
	appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType1, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
	appTemplateInput.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue
	appTemplateInput.ApplicationInput.Labels[conf.GlobalSubaccountIDLabelKey] = subscriptionConsumerSubaccountID
	appTemplateInput.ApplicationInput.BaseURL = &app1BaseURL
	appTemplateInput.ApplicationInput.LocalTenantID = nil
	for i := range appTemplateInput.Placeholders {
		appTemplateInput.Placeholders[i].JSONPath = str.Ptr(fmt.Sprintf("$.%s", conf.SubscriptionProviderAppNameProperty))
	}

	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, appProviderDirectorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, appProviderDirectorCertSecuredClient, tenant.TestTenants.GetDefaultTenantID(), appTmpl)
	require.NoError(t, err)
	require.NotEmpty(t, appTmpl.ID)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

	selfRegLabelValue, ok := appTmpl.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey].(string)
	require.True(t, ok)
	require.Contains(t, selfRegLabelValue, conf.SubscriptionConfig.SelfRegisterLabelValuePrefix+appTmpl.ID)

	deps, err := json.Marshal([]string{selfRegLabelValue, conf.ProviderDestinationConfig.Dependency})
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

	appTplID := appTmpl.ID
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
	t.Logf("Create application template for type: %q", applicationType2)
	appTemplateProvider2 := resource_providers.NewApplicationTemplateProvider(applicationType2, localTenantID2, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder, tnt, nil, graphql.ApplicationStatusConditionConnected)
	defer appTemplateProvider2.Cleanup(t, ctx, oauthGraphQLClient)
	appTemplateProvider2.Provide(t, ctx, oauthGraphQLClient)

	ftProvider := resource_providers.NewFormationTemplateCreator(formationTemplateName)
	defer ftProvider.Cleanup(t, ctx, certSecuredGraphQLClient)
	ftplID := ftProvider.WithSupportedResources(resource_providers.NewApplicationTemplateResource(appTmpl), appTemplateProvider2.GetResource()).WithLeadingProductIDs([]string{internalConsumerID}).Provide(t, ctx, certSecuredGraphQLClient)
	ctx = context.WithValue(ctx, context_keys.FormationTemplateIDKey, ftplID)
	ctx = context.WithValue(ctx, context_keys.FormationTemplateNameKey, ftplID)

	apiPath := fmt.Sprintf("/saas-manager/v1/applications/%s/subscription", conf.SubscriptionProviderAppNameValue)

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
	app1ID := app1.ID
	t.Logf("app1 ID: %q", app1.ID)

	t.Logf("Create application 2 from template: %q", applicationType2)
	appProvider2 := resource_providers.NewApplicationFromTemplateProvider(applicationType2, namePlaceholder, "app2-formation-notifications-tests", displayNamePlaceholder, "App 2 Display Name", tnt)
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

	t.Run("Asynchronous App to App Formation Assignment Notifications with Instances creation", func(t *testing.T) {
		t.Logf("Cleanup notifications")
		cleanupOp := operations.NewCleanupNotificationsOperation().WithExternalServicesMockMtlsSecuredURL(conf.ExternalServicesMockMtlsSecuredURL).WithHTTPClient(certSecuredHTTPClient).Operation()
		defer cleanupOp.Cleanup(t, ctx, certSecuredGraphQLClient)
		cleanupOp.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type: %q and mode: %q to application with ID: %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app2ID)
		headerTemplate := "{\\\"Content-Type\\\": [\\\"application/json\\\"], \\\"Location\\\":[\\\"" + conf.CompassExternalMTLSGatewayURL + "/v1/businessIntegrations/{{.FormationID}}/assignments/{{.Assignment.ID}}/status\\\"]}"
		op := operations.NewAddWebhookToApplicationOperation(graphql.WebhookTypeApplicationTenantMapping, app2ID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.CompassExternalMTLSGatewayURL + "/default-tenant-mapping-handler/v1/tenantMappings/{{.TargetApplication.ID}}\\\",\\\"method\\\":\\\"PATCH\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"subaccountId\\\":\\\"{{ .TargetApplication.Labels.global_subaccount_id }}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").
			WithHeaderTemplate(&headerTemplate).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Add webhook with type %q and mode: %q to application with ID %q", graphql.WebhookTypeApplicationTenantMapping, graphql.WebhookModeAsyncCallback, app1ID)
		op = operations.NewAddWebhookToApplicationOperation(graphql.WebhookTypeApplicationTenantMapping, app1ID, tnt).
			WithWebhookMode(graphql.WebhookModeAsyncCallback).
			WithURLTemplate("{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/formation-callback/async/{{.TargetApplication.ID}}{{if eq .Operation \\\"unassign\\\"}}/{{.SourceApplication.ID}}{{end}}\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}").
			WithInputTemplate("{\\\"context\\\":{\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"globalAccountId\\\":\\\"{{.CustomerTenantContext.AccountID}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"uclFormationName\\\":\\\"{{.Formation.Name}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"subaccountId\\\":\\\"{{ .TargetApplication.Labels.global_subaccount_id }}\\\",\\\"deploymentRegion\\\":\\\"{{if .TargetApplication.Labels.region}}{{.TargetApplication.Labels.region}}{{else}}{{.TargetApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.TargetApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.TargetApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.TargetApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.TargetApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.TargetApplication.ID}}\\\",\\\"configuration\\\":{{.Assignment.Value}}},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .SourceApplication.Labels.region}}{{.SourceApplication.Labels.region}}{{else}}{{.SourceApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.SourceApplicationTemplate.ApplicationNamespace}}\\\",\\\"applicationUrl\\\":\\\"{{.SourceApplication.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.SourceApplication.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.SourceApplication.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.SourceApplication.ID}}\\\",\\\"configuration\\\":{{.ReverseAssignment.Value}}}}").
			WithOutputTemplate("{\\\"config\\\":\\\"{{.Body.config}}\\\", \\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"error\\\": \\\"{{.Body.error}}\\\",\\\"success_status_code\\\": 202}").
			WithHeaderTemplate(&headerTemplate).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		// Config Mutator
		op = operations.NewAddConstraintOperation("mutate").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator(formationconstraintpkg.ConfigMutatorOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType2).
			WithInputTemplate(`{ \"tenant\":\"{{.Tenant}}\",\"only_for_source_subtypes\":[\"app-type-1\"],\"source_resource_type\": \"{{.FormationAssignment.SourceType}}\",\"source_resource_id\": \"{{.FormationAssignment.Source}}\"{{if ne .NotificationStatusReport.State \"CREATE_ERROR\"}},\"modified_configuration\": \"{\\\"credentials\\\":{\\\"inboundCommunication\\\":{\\\"basicAuthentication\\\":{\\\"uri\\\":\\\"$.credentials.inboundCommunication.basicAuthentication.serviceInstances[0].serviceBinding.credentials.uri\\\",\\\"username\\\":\\\"$.credentials.inboundCommunication.basicAuthentication.serviceInstances[0].serviceBinding.credentials.username\\\",\\\"password\\\":\\\"$.credentials.inboundCommunication.basicAuthentication.serviceInstances[0].serviceBinding.credentials.password\\\",\\\"serviceInstances\\\":[{\\\"service\\\":\\\"feature-flags\\\",\\\"plan\\\":\\\"standard\\\",\\\"serviceBinding\\\":{}}]}}}}\"{{if eq .FormationAssignment.State \"INITIAL\"}},\"state\":\"CONFIG_PENDING\"{{ end }}{{ end }},\"resource_type\": \"{{.ResourceType}}\",\"resource_subtype\": \"{{.ResourceSubtype}}\",\"operation\": \"{{.Operation}}\",{{ if .NotificationStatusReport }}\"notification_status_report_memory_address\":{{ .NotificationStatusReport.GetAddress }},{{ end }}\"join_point_location\": {\"OperationName\":\"{{.Location.OperationName}}\",\"ConstraintType\":\"{{.Location.ConstraintType}}\"}}`).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		// Redirect Operator
		redirectedTntID := "1234"
		redirectPath := "/instance-creator/v1/tenantMappings/" + redirectedTntID
		redirectURL := fmt.Sprintf("%s%s", conf.CompassExternalMTLSGatewayURL, redirectPath)
		op = operations.NewAddConstraintOperation("e2e-flow-control-operator-constraint-status-returned").
			WithTargetOperation(graphql.TargetOperationNotificationStatusReturned).
			WithOperator(formationconstraintpkg.AsynchronousFlowControlOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType1).
			WithInputTemplate(fmt.Sprintf("{\\\"url_template\\\": \\\"{\\\\\\\"path\\\\\\\":\\\\\\\"%s\\\\\\\",\\\\\\\"method\\\\\\\":\\\\\\\"PATCH\\\\\\\"}\\\",{{ if .NotificationStatusReport }}\\\"notification_status_report_memory_address\\\":{{ .NotificationStatusReport.GetAddress }},{{ end }}{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}{{ if .ReverseFormationAssignment }}\\\"reverse_formation_assignment_memory_address\\\":{{ .ReverseFormationAssignment.GetAddress }},{{ end }}\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}", redirectURL)).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		redirectURLTemplate := fmt.Sprintf("{\\\\\\\"path\\\\\\\":\\\\\\\"%s\\\\\\\",\\\\\\\"method\\\\\\\":\\\\\\\"PATCH\\\\\\\"}", redirectURL)
		redirectInputTemplate := fmt.Sprintf("{\\\"should_redirect\\\": {{ if and .ReverseFormationAssignment .ReverseFormationAssignment.Value (contains .ReverseFormationAssignment.Value \\\"serviceInstances\\\") }}true{{else}}false{{end}},\\\"url_template\\\": \\\"%s\\\",{{ if .Webhook }}\\\"webhook_memory_address\\\":{{ .Webhook.GetAddress }},{{ end }}{{ if .FormationAssignment }}\\\"formation_assignment_memory_address\\\":{{ .FormationAssignment.GetAddress }},{{ end }}\\\"resource_type\\\": \\\"{{.ResourceType}}\\\",\\\"resource_subtype\\\": \\\"{{.ResourceSubtype}}\\\",\\\"operation\\\": \\\"{{.Operation}}\\\",\\\"join_point_location\\\": {\\\"OperationName\\\":\\\"{{.Location.OperationName}}\\\",\\\"ConstraintType\\\":\\\"{{.Location.ConstraintType}}\\\"}}", redirectURLTemplate)
		op = operations.NewAddConstraintOperation("e2e-flow-control-operator-constraint-send-notification").
			WithTargetOperation(graphql.TargetOperationSendNotification).
			WithOperator(formationconstraintpkg.AsynchronousFlowControlOperator).
			WithResourceType(graphql.ResourceTypeApplication).
			WithResourceSubtype(applicationType1).
			WithInputTemplate(redirectInputTemplate).
			WithTenant(tnt).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 2 to formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		asserter := asserters.NewFormationAssignmentAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter := asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app2ID, tnt).WithAsserters(asserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Assign application 1 to formation: %s", formationName)
		expectedInstanceCreatorConfig := str.Ptr(`{"credentials": {"outboundCommunication": {"basicAuthentication": {"password": "password", "uri": "uri", "username": "username"}}}}`)
		expectedNotSubstitutedConfig := str.Ptr(`{"credentials": {"inboundCommunication": {"basicAuthentication": {"uri": "$.credentials.inboundCommunication.basicAuthentication.serviceInstances[0].serviceBinding.credentials.uri", "password": "$.credentials.inboundCommunication.basicAuthentication.serviceInstances[0].serviceBinding.credentials.password", "username": "$.credentials.inboundCommunication.basicAuthentication.serviceInstances[0].serviceBinding.credentials.username", "serviceInstances": [{"plan": "standard", "service": "feature-flags", "serviceBinding": {}}]}}}}`)

		expectationsBuilder = mock_data.NewFAExpectationsBuilder().
			WithParticipant(app1ID).
			WithParticipant(app2ID).
			WithNotifications([]*mock_data.NotificationData{
				mock_data.NewNotificationData(app1ID, app1ID, readyAssignmentState, nil, nil),
				mock_data.NewNotificationData(app1ID, app2ID, readyAssignmentState, expectedNotSubstitutedConfig, nil),
				mock_data.NewNotificationData(app2ID, app2ID, readyAssignmentState, nil, nil),
				mock_data.NewNotificationData(app2ID, app1ID, readyAssignmentState, expectedInstanceCreatorConfig, nil),
			}).
			WithOperations([]*fixtures.Operation{
				fixtures.NewOperation(app1ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app1ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app1ID, "ASSIGN", "ASSIGN_OBJECT", true),
				fixtures.NewOperation(app2ID, app2ID, "ASSIGN", "ASSIGN_OBJECT", true),
			})

		configMatcher := func(t require.TestingT, expectedConfig, actualConfig *string) bool {
			if expectedConfig != nil && strings.Contains(*expectedConfig, "outboundCommunication") {
				return assertSubstitutedConfig(t, expectedConfig, actualConfig)
			}
			return assertExactConfig(t, expectedConfig, actualConfig)
		}

		asserterWithCustomConfigMatcher := asserters.NewFormationAssignmentsAsyncCustomConfigMatcherAsserter(configMatcher, expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
			WithTimeout(eventuallyTimeoutForInstances).
			WithTick(eventuallyTickForInstances)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewAssignAppToFormationOperation(app1ID, tnt).WithAsserters(asserterWithCustomConfigMatcher, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 1 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder().WithParticipant(app2ID)
		faAsserter := asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt).
			WithTimeout(eventuallyTimeoutForInstances).
			WithTick(eventuallyTickForInstances)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		unassignNotificationsAsserter := asserters.NewUnassignNotificationsAsserter(1, app1ID, app2ID, subscriptionConsumerTenantID, appNamespace, appRegion, tnt, emptyParentCustomerID, "", conf.ExternalServicesMockMtlsSecuredURL, certSecuredHTTPClient)
		op = operations.NewUnassignAppToFormationOperationGlobal(app1ID).WithAsserters(faAsserter, statusAsserter, unassignNotificationsAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)

		t.Logf("Unassign Application 2 from formation: %s", formationName)
		expectationsBuilder = mock_data.NewFAExpectationsBuilder()
		faAsserter = asserters.NewFormationAssignmentAsyncAsserter(expectationsBuilder.GetExpectations(), expectationsBuilder.GetExpectedAssignmentsCount(), certSecuredGraphQLClient, tnt)
		statusAsserter = asserters.NewFormationStatusAsserter(tnt, certSecuredGraphQLClient)
		op = operations.NewUnassignAppToFormationOperationGlobal(app2ID).WithAsserters(faAsserter, statusAsserter).Operation()
		defer op.Cleanup(t, ctx, certSecuredGraphQLClient)
		op.Execute(t, ctx, certSecuredGraphQLClient)
	})
}

func assertExactConfig(t require.TestingT, expectedConfig, actualConfig *string) bool {
	return test_json.AssertJSONStringEquality(t, expectedConfig, actualConfig)
}

func assertSubstitutedConfig(t require.TestingT, _, actualConfig *string) bool {
	found := 0

	var iterate func(key string, value gjson.Result)
	iterate = func(key string, value gjson.Result) {
		if value.IsObject() {
			for k, v := range value.Map() {
				iterate(k, v)
			}
		} else if value.IsArray() {
			for i, el := range value.Array() {
				strI := fmt.Sprint(i)
				iterate(strI, el)
			}
		} else {
			if (key == uriKey || key == usernameKey || key == "password") && !strings.Contains(key, "$.") {
				found++
			}
		}
	}
	iterate("", gjson.Parse(*actualConfig))

	if found != 3 {
		t.Errorf("The actual assignment config %s don't have substituted %q, %q and %q", *actualConfig, uriKey, usernameKey, "password")
		return false
	}

	return true
}
