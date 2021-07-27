package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestCreateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template-name"
	appTemplateInput := fixtures.FixApplicationTemplate(name)
	appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
	output := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Create application template")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createApplicationTemplateRequest, &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, createApplicationTemplateRequest.Query(), "create application template")

	t.Log("Check if application template was created")

	getApplicationTemplateRequest := fixtures.FixApplicationTemplateRequest(output.ID)
	appTemplateOutput := graphql.ApplicationTemplate{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, getApplicationTemplateRequest, &appTemplateOutput)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOutput)
	assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
	saveExample(t, getApplicationTemplateRequest.Query(), "query application template")
}

func TestUpdateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template"
	newName := "new-app-template"
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationRegisterInput{
		Name:           "new-app-create-input",
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTemplate := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenantId, fixtures.FixApplicationTemplate(name))
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTemplate.ID)

	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateAppTemplateRequest, &updateOutput)
	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	saveExample(t, updateAppTemplateRequest.Query(), "update application template")
}

func TestDeleteApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTemplate := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenantId, fixtures.FixApplicationTemplate(name))

	deleteApplicationTemplateRequest := fixtures.FixDeleteApplicationTemplateRequest(appTemplate.ID)
	deleteOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Delete application template")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteApplicationTemplateRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application template was deleted")

	out := fixtures.GetApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTemplate.ID)

	require.Empty(t, out)
	saveExample(t, deleteApplicationTemplateRequest.Query(), "delete application template")
}

func TestQueryApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTemplate := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenantId, fixtures.FixApplicationTemplate(name))
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTemplate.ID)

	getApplicationTemplateRequest := fixtures.FixApplicationTemplateRequest(appTemplate.ID)
	output := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Get application template")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getApplicationTemplateRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	//THEN
	t.Log("Check if application template was received")
	assert.Equal(t, name, output.Name)
}

func TestQueryApplicationTemplates(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name1 := "app-template-1"
	name2 := "app-template-2"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application templates")
	appTemplate1 := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenantId, fixtures.FixApplicationTemplate(name1))
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTemplate1.ID)

	appTemplate2 := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenantId, fixtures.FixApplicationTemplate(name2))
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTemplate2.ID)

	first := 100
	after := ""

	getApplicationTemplatesRequest := fixtures.FixGetApplicationTemplatesWithPagination(first, after)
	output := graphql.ApplicationTemplatePage{}

	// WHEN
	t.Log("List application templates")
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, getApplicationTemplatesRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application templates were received")
	assert.Subset(t, output.Data, []*graphql.ApplicationTemplate{&appTemplate1, &appTemplate2})
	saveExample(t, getApplicationTemplatesRequest.Query(), "query application templates")
}

func TestRegisterApplicationFromTemplate(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	tmplName := "template"
	placeholderKey := "new-placeholder"
	appTmplInput := fixtures.FixApplicationTemplate(tmplName)
	appTmplInput.ApplicationInput.Description = ptr.String("test {{new-placeholder}}")
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        placeholderKey,
			Description: ptr.String("description"),
		}}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appTmpl := fixtures.CreateApplicationTemplateFromInput(t, ctx, dexGraphQLClient, tenantId, appTmplInput)
	defer fixtures.DeleteApplicationTemplate(t, ctx, dexGraphQLClient, tenantId, appTmpl.ID)

	appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: tmplName, Values: []*graphql.TemplateValueInput{
		{
			Placeholder: placeholderKey,
			Value:       "new-value",
		}}}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}
	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createAppFromTmplRequest, &outputApp)

	//THEN
	require.NoError(t, err)
	fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, outputApp.ID)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-value", *outputApp.Application.Description)
	saveExample(t, createAppFromTmplRequest.Query(), "register application from template")
}

func TestAddWebhookToApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create integration system")
	intSys := fixtures.RegisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, name)
	require.NotEmpty(t, intSys)
	defer fixtures.UnregisterIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, dexGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, dexGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTemplate := fixtures.CreateApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, name)
	defer fixtures.DeleteApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate.ID)

	// add
	url := "http://new-webhook.url"
	urlUpdated := "http://updated-webhook.url"
	outputTemplate := "{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"

	webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &url,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})

	require.NoError(t, err)
	addReq := fixtures.FixAddWebhookToTemplateRequest(appTemplate.ID, webhookInStr)
	saveExampleInCustomDir(t, addReq.Query(), addWebhookCategory, "add application template webhook")

	actualWebhook := graphql.Webhook{}
	t.Run("fails when tenant is present", func(t *testing.T) {
		t.Log("Trying to Webhook to application template with tenant")
		err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, addReq, &actualWebhook)
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
	t.Run("succeeds with no tenant", func(t *testing.T) {

		t.Log("Add Webhook to application template")
		err = testctx.Tc.RunOperationWithoutTenant(ctx, oauthGraphQLClient, addReq, &actualWebhook)
		require.NoError(t, err)
		assert.NotNil(t, actualWebhook.URL)
		assert.Equal(t, "http://new-webhook.url", *actualWebhook.URL)
		assert.Equal(t, graphql.WebhookTypeUnregisterApplication, actualWebhook.Type)
		id := actualWebhook.ID
		require.NotNil(t, id)

	})

	updatedAppTemplate := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate.ID)
	assert.Len(t, updatedAppTemplate.Webhooks, 1)

	webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &urlUpdated,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})
	require.NoError(t, err)

	t.Log("Getting Webhooks for application template")
	updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, oauthGraphQLClient, updateReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, urlUpdated, *actualWebhook.URL)

	// delete

	//GIVEN
	deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)

	//WHEN
	err = testctx.Tc.RunOperationWithoutTenant(ctx, oauthGraphQLClient, deleteReq, &actualWebhook)

	//THEN
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, urlUpdated, *actualWebhook.URL)
}
