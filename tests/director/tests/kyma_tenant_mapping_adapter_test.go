package tests

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKymaTenantMappingAdapter(t *testing.T) {
	ctx := context.Background()

	tenant := tenant.TestTenants.GetDefaultTenantID()

	// Create formation template
	applicationType := "App1"
	formationTemplateName := "test-kyma-formation-template"
	formationTemplateInput := graphql.FormationTemplateInput{
		Name:             formationTemplateName,
		ApplicationTypes: []string{applicationType},
		RuntimeTypes:     []string{"kyma"},
	}

	var formationTemplate graphql.FormationTemplate // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &formationTemplate)
	formationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateInput)

	// Create formation from template
	formationName := "test-formation"
	t.Logf("Should create formation: %q", formationName)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, formationName)
	fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, certSecuredGraphQLClient, tenant, formationName, &formationTemplate.Name)

	// Creating application template
	t.Log("Create application template")
	appTemplateName := "SAP app-template"
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

	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant, appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant, appTemplate)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Name)

	// Create application from template
	appFromTemplate := graphql.ApplicationFromTemplateInput{TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{{Placeholder: "name", Value: "name"}, {Placeholder: "display-name", Value: "display-name"}}}
	appFromTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTemplate)
	require.NoError(t, err)
	createAppFromTemplateRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTemplateGQL)
	app := graphql.ApplicationExt{}

	t.Logf("Creating application from application template with id %s", appTemplate.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createAppFromTemplateRequest, &app)
	defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenant, app.ID)
	require.NoError(t, err)

	// Add webhook to the application pointing to the external services mock for basic credentials
	webhookType := graphql.WebhookTypeConfigurationChanged
	urlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/tenants/basicCredentials,\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"
	inputTemplate := "{\\\"context\\\":{\\\"platform\\\":\\\"{{if .CustomerTenantContext.AccountID}}btp{{else}}unified-services{{end}}\\\",\\\"uclFormationId\\\":\\\"{{.FormationID}}\\\",\\\"accountId\\\":\\\"{{if .CustomerTenantContext.AccountID}}{{.CustomerTenantContext.AccountID}}{{else}}{{.CustomerTenantContext.Path}}{{end}}\\\",\\\"crmId\\\":\\\"{{.CustomerTenantContext.CustomerID}}\\\",\\\"operation\\\":\\\"{{.Operation}}\\\"},\\\"assignedTenant\\\":{\\\"state\\\":\\\"{{.Assignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.Assignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if .Application.Labels.region}}{{.Application.Labels.region}}{{else}}{{.ApplicationTemplate.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{if .Application.ApplicationNamespace}}{{.Application.ApplicationNamespace}}{{else}}{{.ApplicationTemplate.ApplicationNamespace}}{{end}}\\\",\\\"applicationUrl\\\":\\\"{{.Application.BaseURL}}\\\",\\\"applicationTenantId\\\":\\\"{{.Application.LocalTenantID}}\\\",\\\"uclSystemName\\\":\\\"{{.Application.Name}}\\\",\\\"uclSystemTenantId\\\":\\\"{{.Application.ID}}\\\",{{if .ApplicationTemplate.Labels.parameters}}\\\"parameters\\\":{{.ApplicationTemplate.Labels.parameters}},{{end}}\\\"configuration\\\":{{.ReverseAssignment.Value}}},\\\"receiverTenant\\\":{\\\"state\\\":\\\"{{.ReverseAssignment.State}}\\\",\\\"uclAssignmentId\\\":\\\"{{.ReverseAssignment.ID}}\\\",\\\"deploymentRegion\\\":\\\"{{if and .RuntimeContext .RuntimeContext.Labels.region}}{{.RuntimeContext.Labels.region}}{{else}}{{.Runtime.Labels.region}}{{end}}\\\",\\\"applicationNamespace\\\":\\\"{{.Runtime.ApplicationNamespace}}\\\",\\\"applicationTenantId\\\":\\\"{{if .RuntimeContext}}{{.RuntimeContext.Value}}{{else}}{{.Runtime.Labels.global_subaccount_id}}{{end}}\\\",\\\"uclSystemTenantId\\\":\\\"{{if .RuntimeContext}}{{.RuntimeContext.ID}}{{else}}{{.Runtime.ID}}{{end}}\\\",{{if .Runtime.Labels.parameters}}\\\"parameters\\\":{{.Runtime.Labels.parameters}},{{end}}\\\"configuration\\\":{{.Assignment.Value}}}}"
	outputTemplate := "{\\\"error\\\":\\\"{{.Body.error}}\\\",\\\"state\\\":\\\"{{.Body.state}}\\\",\\\"success_status_code\\\": 200,\\\"incomplete_status_code\\\": 422}"
	headerTemplate := "{\\\"Content-Type\\\": [\\\"application/json\\\"]}"

	applicationWebhookInput := &graphql.WebhookInput{
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
	applicationWebhook := fixtures.AddWebhookToApplication(t, ctx, certSecuredGraphQLClient, applicationWebhookInput, tenant, app.ID)
	defer fixtures.CleanupWebhook(t, ctx, certSecuredGraphQLClient, tenant, applicationWebhook.ID)

	// Create Kyma runtime which should have webhook added to it pointing to the Kyma Adapter
	runtimeName := "runtime-test"
	t.Log(fmt.Sprintf("Registering runtime %q", runtimeName))
	runtimeRegInput := fixRuntimeInput(runtimeName)

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenant, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenant, runtimeRegInput, conf.GatewayOauth)

	// Assign the application to the formation
	t.Logf("Assigning application with name %q to formation with name %q", app.Name, formationName)
	assignReq := fixtures.FixAssignFormationRequest(app.ID, string(graphql.FormationObjectTypeApplication), formationName)
	var assignedFormation graphql.Formation
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant, assignReq, &assignedFormation)
	defer fixtures.CleanupFormation(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: formationName}, app.ID, graphql.FormationObjectTypeApplication, tenant)
	require.NoError(t, err)
	require.Equal(t, formationName, assignedFormation.Name)

	// Check that there are no bundle instance auths

	// Assign the Kyma runtime to formation
	t.Logf("Assigning runtime with name %q to formation with name %q", runtime.Name, formationName)
	newFormationInput := graphql.FormationInput{Name: formationName}
	defer fixtures.UnassignFormationWithRuntimeObjectType(t, ctx, certSecuredGraphQLClient, newFormationInput, runtime.ID, tenant)
	assignReq = fixtures.FixAssignFormationRequest(runtime.ID, string(graphql.FormationObjectTypeRuntime), formationName)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, formationName, assignFormation.Name)

	// Check that there are bundle instance auths created for each bundle by the Kyma Adapter

	// Update the application webhook to point to the external services mock for oauth credentials
	updatedUrlTemplate := "{\\\"path\\\":\\\"" + conf.ExternalServicesMockMtlsSecuredURL + "/v1/tenants/oauthCredentials,\\\"method\\\":\\\"{{if eq .Operation \\\"assign\\\"}}PATCH{{else}}DELETE{{end}}\\\"}"

	updatedApplicationWebhookInput := &graphql.WebhookInput{
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
	updatedWebhook := fixtures.UpdateWebhook(t, ctx, certSecuredGraphQLClient, tenant, applicationWebhook.ID, updatedApplicationWebhookInput)
	require.Equal(t, updatedWebhook.ID, applicationWebhook.ID)

	// reset and resync

	// check there are the new instance auths

	// unassign app from formation

	// check there are no auths

	// cleanup
}
