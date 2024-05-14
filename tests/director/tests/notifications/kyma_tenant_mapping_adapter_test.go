package notifications

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
)

const (
	OpenAPISpec = "https://raw.githubusercontent.com/kyma-incubator/github-slack-connectors/beb8e5b6d8f3a644b8380e667a9376bc353e54dd/github-connector/internal/registration/configs/githubopenAPI.json"
)

func TestKymaTenantMappingAdapter(t *testing.T) {
	ctx := context.Background()

	tenantId := conf.TestProviderAccountID

	// Create formation template
	applicationType := "App1"
	formationTemplateName := "test-kyma-formation-template"
	runtimeArtifactKind := graphql.ArtifactTypeEnvironmentInstance
	supportReset := true
	formationTemplateRegisterInput := graphql.FormationTemplateRegisterInput{
		Name:                   formationTemplateName,
		ApplicationTypes:       []string{applicationType},
		RuntimeTypes:           []string{"kyma"},
		SupportsReset:          &supportReset,
		RuntimeTypeDisplayName: str.Ptr("test-display-name"),
		RuntimeArtifactKind:    &runtimeArtifactKind,
	}

	var formationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplate)
	formationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateRegisterInput)

	// Create formation from template
	formationName := "test-kyma-adapter-formation"
	t.Logf("Should create formation: %q", formationName)
	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, formationName)
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, formationName, &formationTemplate.Name)

	// Creating application template
	t.Log("Create application template")
	namePlaceholder := "name"
	displayNamePlaceholder := "display-name"
	appRegion := "test-app-region"
	appNamespace := "compass.test"
	localTenantID := "local-tenant-id"
	baseUrl := "url"
	t.Logf("Create application template for type %q", applicationType)
	appTemplateInput := fixtures.FixApplicationTemplateWithoutWebhook(applicationType, localTenantID, appRegion, appNamespace, namePlaceholder, displayNamePlaceholder)
	appTemplateInput.ApplicationInput.BaseURL = &baseUrl
	appTemplateInput.ApplicationInput.Bundles = []*graphql.BundleCreateInput{{
		Name: "bndl-1",
		APIDefinitions: []*graphql.APIDefinitionInput{{
			Name:      "api-def-1",
			TargetURL: "https://target.url",
			Spec: &graphql.APISpecInput{
				Format: graphql.SpecFormatJSON,
				Type:   graphql.APISpecTypeOpenAPI,
				FetchRequest: &graphql.FetchRequestInput{
					URL: OpenAPISpec,
				},
			},
		}},
	},
		{
			Name: "bndl-2",
			APIDefinitions: []*graphql.APIDefinitionInput{{
				Name:      "api-def-2",
				TargetURL: "https://target.url",
				Spec: &graphql.APISpecInput{
					Format: graphql.SpecFormatJSON,
					Type:   graphql.APISpecTypeOpenAPI,
					FetchRequest: &graphql.FetchRequestInput{
						URL: OpenAPISpec,
					},
				},
			}},
		},
	}

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, conf.TestProviderSubaccountID, "int-kyma-tma-notifications")
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, conf.TestProviderSubaccountID, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, conf.TestProviderSubaccountID, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	appTemplate, err := fixtures.CreateApplicationTemplateFromInputWithoutTenant(t, ctx, oauthGraphQLClient, appTemplateInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTemplate)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Name)

	// Create application from template
	appName := "kyma-adapter-app-name"
	appFromTemplate := graphql.ApplicationFromTemplateInput{TemplateName: appTemplate.Name, Values: []*graphql.TemplateValueInput{{Placeholder: "name", Value: appName}, {Placeholder: "display-name", Value: "display-name"}}}
	appFromTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTemplate)
	require.NoError(t, err)
	createAppFromTemplateRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTemplateGQL)
	app := graphql.ApplicationExt{}

	t.Logf("Creating application with name %q from application template with id %s", appName, appTemplate.ID)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, createAppFromTemplateRequest, &app)
	defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenantId, app.ID)
	require.NoError(t, err)

	// Add webhook to the application pointing to the external services mock for basic credentials
	webhookType := graphql.WebhookTypeConfigurationChanged
	urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/tenants/emptyCredentials\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
	inputTemplate := "{\\\"context\\\":{\\\"platform\\\":\\\"{{if .CustomerTenantContext.AccountID}}btp{{else}}unified-services{{end}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"accountId\\\":\\\"{{if .CustomerTenantContext.AccountID}}{{.CustomerTenantContext.AccountID}}{{else}}{{.CustomerTenantContext.Path}}{{end}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .Application.Labels.region}}{{.Application.Labels.region}}{{else}}{{.ApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{if .Application.ApplicationNamespace}}{{.Application.ApplicationNamespace}}{{else}}{{.ApplicationTemplate.ApplicationNamespace}}{{end}}\\\",\\\"applicationUrl\\\":\\\"{{.Application.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.Application.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.Application.ID}}\\\",{{if .ApplicationTemplate.Labels.parameters}}\\\"parameters\\\":{{.ApplicationTemplate.Labels.parameters}},{{end}}\\\"configuration\\\":{{.ReverseAssignment.Value}}},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if and .RuntimeContext .RuntimeContext.Labels.region}}{{.RuntimeContext.Labels.region}}{{else}}{{.Runtime.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.Runtime.ApplicationNamespace}}\\\",\\\"applicationTenantId\\\":\\\"{{if .RuntimeContext}}{{.RuntimeContext.Value}}{{else}}{{.Runtime.Labels.global_subaccount_id}}{{end}}\\\",\\\"uclSystemTenantId\\\":\\\"{{if .RuntimeContext}}{{.RuntimeContext.ID}}{{else}}{{.Runtime.ID}}{{end}}\\\",{{if .Runtime.Labels.parameters}}\\\"parameters\\\":{{.Runtime.Labels.parameters}},{{end}}\\\"configuration\\\":{{.Assignment.Value}}}}"
	outputTemplate := "{\\\"error\\\":\\\"{{.Body.error}}\\\",\\\"state\\\":\\\"{{.Body.state}}\\\",\\\"config\\\":\\\"{{.Body.configuration}}\\\",\\\"success_status_code\\\": 200,\\\"incomplete_status_code\\\": 422}"
	headerTemplate := "{\\\"Content-Type\\\": [\\\"application/json\\\"]}"
	webhookMode := graphql.WebhookModeSync

	applicationWebhookInput := &graphql.WebhookInput{
		Mode: &webhookMode,
		Type: webhookType,
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
		},
		URLTemplate:    &urlTemplate,
		InputTemplate:  &inputTemplate,
		OutputTemplate: &outputTemplate,
		HeaderTemplate: &headerTemplate,
	}

	t.Logf("Add webhook with type %q to application with ID %q", webhookType, app.ID)
	applicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tenantId, app.ID)
	defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tenantId, applicationWebhook.ID)

	// Create Kyma runtime which should have webhook added to it pointing to the Kyma Adapter
	runtimeName := "runtime-test"
	t.Log(fmt.Sprintf("Registering runtime %q", runtimeName))
	runtimeRegInput := fixtures.FixRuntimeRegisterInputWithoutLabels(runtimeName)
	runtimeRegInput.Labels[conf.GlobalSubaccountIDLabelKey] = conf.TestProviderSubaccountID

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, runtimeRegInput, conf.GatewayOauth)

	// Assign the Kyma runtime to formation
	t.Logf("Assigning runtime with name %q to formation with name %q", runtime.Name, formationName)
	newFormationInput := graphql.FormationInput{Name: formationName}
	defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, newFormationInput, runtime.ID, tenantId)
	fixtures.AssignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, newFormationInput, runtime.ID, tenantId)

	// Check that there are no bundle instance auths
	t.Log("Assert that there are no bundle instance auths for application bundles")
	queryAPIForApplication := fixtures.FixGetApplicationWithInstanceAuths(app.ID)

	returnedApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, queryAPIForApplication, &returnedApp)
	require.NoError(t, err)

	require.Equal(t, 2, returnedApp.Bundles.TotalCount)
	require.Equal(t, 0, len(returnedApp.Bundles.Data[0].InstanceAuths))
	require.Equal(t, 0, len(returnedApp.Bundles.Data[1].InstanceAuths))

	// Assign the application to the formation
	t.Logf("Assigning application with name %q to formation with name %q", app.Name, formationName)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, newFormationInput, app.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, newFormationInput, app.ID, tenantId)

	// Check that there are no bundle instance auths because the external services mock returns empty config
	t.Log("Assert that there are no bundle instance auths for application bundles")
	returnedApp = graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, queryAPIForApplication, &returnedApp)
	require.NoError(t, err)

	require.Equal(t, 2, returnedApp.Bundles.TotalCount)
	require.Equal(t, 0, len(returnedApp.Bundles.Data[0].InstanceAuths))
	require.Equal(t, 0, len(returnedApp.Bundles.Data[1].InstanceAuths))

	// Assert the assignments - there should be an assignment with source APP, target RUNTIME and state CONFIG_PENDING
	t.Logf("Assert formation assignments for formation with ID: %q", formation.ID)
	expectedAssignments := map[string]map[string]fixtures.AssignmentState{
		app.ID: {
			app.ID:     fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			runtime.ID: fixtures.AssignmentState{State: "CONFIG_PENDING", Config: nil, Value: nil, Error: nil},
		},
		runtime.ID: {
			app.ID:     fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
			runtime.ID: fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil},
		},
	}
	assertFormationAssignments(t, ctx, tenantId, formation.ID, 4, expectedAssignments)

	// Update the application webhook to point to the external services mock for basic credentials
	updatedUrlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/tenants/basicCredentials\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"

	updatedApplicationWebhookInput := &graphql.WebhookInput{
		Mode: &webhookMode,
		Type: webhookType,
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
		},
		URLTemplate:    &updatedUrlTemplate,
		InputTemplate:  &inputTemplate,
		OutputTemplate: &outputTemplate,
		HeaderTemplate: &headerTemplate,
	}
	t.Log("Update the application webhook to point to basic credentials external services mock endpoint")
	updatedWebhook := fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tenantId, applicationWebhook.ID, updatedApplicationWebhookInput)
	require.Equal(t, updatedWebhook.ID, applicationWebhook.ID)

	// Reset and resync
	t.Logf("Resynchronize formation %q with reset", formationName)
	resynchronizeReq := fixtures.FixResynchronizeFormationNotificationsRequestWithResetOption(formation.ID, true)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, resynchronizeReq, &formation)
	require.NoError(t, err)
	// Check that there are bundle instance auths created for each application bundle by the Kyma Adapter
	t.Log("Assert that there are bundle instance auths for application bundles")
	returnedApp = graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, queryAPIForApplication, &returnedApp)
	require.NoError(t, err)

	require.Equal(t, 2, returnedApp.Bundles.TotalCount)
	require.Equal(t, 1, len(returnedApp.Bundles.Data[0].InstanceAuths))
	require.Equal(t, "user", returnedApp.Bundles.Data[0].InstanceAuths[0].Auth.Credential.(*graphql.BasicCredentialData).Username)
	require.Equal(t, "pass", returnedApp.Bundles.Data[0].InstanceAuths[0].Auth.Credential.(*graphql.BasicCredentialData).Password)
	require.Equal(t, 1, len(returnedApp.Bundles.Data[1].InstanceAuths))
	require.Equal(t, "user", returnedApp.Bundles.Data[1].InstanceAuths[0].Auth.Credential.(*graphql.BasicCredentialData).Username)
	require.Equal(t, "pass", returnedApp.Bundles.Data[1].InstanceAuths[0].Auth.Credential.(*graphql.BasicCredentialData).Password)

	// Update the application webhook to point to the external services mock for oauth credentials
	updatedUrlTemplate = "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/tenants/oauthCredentials\\\",\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"

	updatedApplicationWebhookInput = &graphql.WebhookInput{
		Mode: &webhookMode,
		Type: webhookType,
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr("sap:cmp-mtls:v1"),
		},
		URLTemplate:    &updatedUrlTemplate,
		InputTemplate:  &inputTemplate,
		OutputTemplate: &outputTemplate,
		HeaderTemplate: &headerTemplate,
	}
	t.Log("Update the application webhook to point to oauth credentials external services mock endpoint")
	updatedWebhook = fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tenantId, applicationWebhook.ID, updatedApplicationWebhookInput)
	require.Equal(t, updatedWebhook.ID, applicationWebhook.ID)

	// Reset and resync
	t.Logf("Resynchronize formation %q with reset", formationName)
	resynchronizeReq = fixtures.FixResynchronizeFormationNotificationsRequestWithResetOption(formation.ID, true)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, resynchronizeReq, &formation)
	require.NoError(t, err)

	// Check there are the updated instance auths for the application bundles
	t.Log("Assert that there are the updated bundle instance auths for application bundles")
	returnedApp = graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, queryAPIForApplication, &returnedApp)
	require.NoError(t, err)

	require.Equal(t, 2, returnedApp.Bundles.TotalCount)
	require.Equal(t, 1, len(returnedApp.Bundles.Data[0].InstanceAuths))
	require.Equal(t, "url", returnedApp.Bundles.Data[0].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).URL)
	require.Equal(t, "id", returnedApp.Bundles.Data[0].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).ClientID)
	require.Equal(t, "secret", returnedApp.Bundles.Data[0].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).ClientSecret)
	require.Equal(t, 1, len(returnedApp.Bundles.Data[1].InstanceAuths))
	require.Equal(t, "url", returnedApp.Bundles.Data[1].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).URL)
	require.Equal(t, "id", returnedApp.Bundles.Data[1].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).ClientID)
	require.Equal(t, "secret", returnedApp.Bundles.Data[1].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).ClientSecret)

	// Unassign application from formation
	t.Logf("Unassigning application with name %q from formation with name %q", app.Name, formationName)
	fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app.ID, tenantId)

	// Check that there are no instance auths for application bundles
	t.Log("Assert that there are no bundle instance auths for application bundles")
	returnedApp = graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, queryAPIForApplication, &returnedApp)
	require.NoError(t, err)

	require.Equal(t, 2, returnedApp.Bundles.TotalCount)
	require.Equal(t, 0, len(returnedApp.Bundles.Data[0].InstanceAuths))
	require.Equal(t, 0, len(returnedApp.Bundles.Data[1].InstanceAuths))

	// Assign application again
	t.Logf("Assigning application with name %q to formation with name %q", app.Name, formationName)
	defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, newFormationInput, app.ID, tenantId)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, newFormationInput, app.ID, tenantId)

	// Check there are the updated instance auths for the application bundles
	t.Log("Assert that there are the bundle instance auths for application bundles")
	returnedApp = graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, queryAPIForApplication, &returnedApp)
	require.NoError(t, err)

	require.Equal(t, 2, returnedApp.Bundles.TotalCount)
	require.Equal(t, 1, len(returnedApp.Bundles.Data[0].InstanceAuths))
	require.Equal(t, "url", returnedApp.Bundles.Data[0].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).URL)
	require.Equal(t, "id", returnedApp.Bundles.Data[0].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).ClientID)
	require.Equal(t, "secret", returnedApp.Bundles.Data[0].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).ClientSecret)
	require.Equal(t, 1, len(returnedApp.Bundles.Data[1].InstanceAuths))
	require.Equal(t, "url", returnedApp.Bundles.Data[1].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).URL)
	require.Equal(t, "id", returnedApp.Bundles.Data[1].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).ClientID)
	require.Equal(t, "secret", returnedApp.Bundles.Data[1].InstanceAuths[0].Auth.Credential.(*graphql.OAuthCredentialData).ClientSecret)

	// Unassign runtime
	t.Logf("Unassigning runtime with name %q from formation with name %q", runtime.Name, formationName)
	fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, newFormationInput, runtime.ID, tenantId)

	t.Log("Assert that there are no bundle instance auths for application bundles")
	returnedApp = graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, queryAPIForApplication, &returnedApp)
	require.NoError(t, err)

	require.Equal(t, 2, returnedApp.Bundles.TotalCount)
	require.Equal(t, 0, len(returnedApp.Bundles.Data[0].InstanceAuths))
	require.Equal(t, 0, len(returnedApp.Bundles.Data[1].InstanceAuths))
}
