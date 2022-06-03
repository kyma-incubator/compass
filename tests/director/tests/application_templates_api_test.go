package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

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
	appTemplateName := createAppTemplateName("app-template-name")
	appTemplateInput := fixAppTemplateInput(appTemplateName)
	appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
	output := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Create application template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createApplicationTemplateRequest, &output)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &output)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, createApplicationTemplateRequest.Query(), "create application template")

	t.Log("Check if application template was created")

	getApplicationTemplateRequest := fixtures.FixApplicationTemplateRequest(output.ID)
	appTemplateOutput := graphql.ApplicationTemplate{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getApplicationTemplateRequest, &appTemplateOutput)

	appTemplateInput.Labels[conf.SelfRegLabelKey] = appTemplateOutput.Labels[conf.SelfRegLabelKey]
	appTemplateInput.Labels["global_subaccount_id"] = conf.ConsumerID

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOutput)
	assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
	saveExample(t, getApplicationTemplateRequest.Query(), "query application template")
}

func TestCreateApplicationTemplate_NotValid(t *testing.T) {
	namePlaceholder := "name-placeholder"
	displayNamePlaceholder := "display-name-placeholder"

	testCases := []struct {
		Name                    string
		AppTemplateName         string
		AppTemplatePlaceholders []*graphql.PlaceholderDefinitionInput
		AppInputDescription     *string
		ExpectedErrMessage      string
	}{
		{
			Name:            "not compliant name",
			AppTemplateName: "not-compliant-name",
			AppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
				},
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
				},
			},
			AppInputDescription: nil,
			ExpectedErrMessage:  "application template name \"not-compliant-name\" does not comply with the following naming convention",
		},
		{
			Name:            "not compliant placeholders",
			AppTemplateName: fmt.Sprintf("SAP %s (%s)", "app-template-name", conf.SelfRegRegion),
			AppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
				},
				{
					Name:        "not-compliant",
					Description: &displayNamePlaceholder,
				},
			},
			AppInputDescription: ptr.String("test {{not-compliant}}"),
			ExpectedErrMessage:  "unexpected placeholder with name \"not-compliant\" found",
		},
		{
			Name:            "not matching region",
			AppTemplateName: fmt.Sprintf("SAP %s (%s)", "app-template-name-2", "random-region"),
			AppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
				},
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
				},
			},
			AppInputDescription: nil,
			ExpectedErrMessage:  "the region specified in the application template name does not match",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			appTemplateInput := fixAppTemplateInput(testCase.AppTemplateName)
			if testCase.AppInputDescription != nil {
				appTemplateInput.ApplicationInput.Description = testCase.AppInputDescription
			}
			appTemplateInput.Placeholders = testCase.AppTemplatePlaceholders
			appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
			require.NoError(t, err)

			createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
			output := graphql.ApplicationTemplate{}

			// WHEN
			t.Log("Create application template")
			err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createApplicationTemplateRequest, &output)
			defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &output)

			//THEN
			require.NotNil(t, err)
			if testCase.ExpectedErrMessage != "" {
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestUpdateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := createAppTemplateName("app-template")
	newName := "new-app-template"
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationRegisterInput{
		Name:           "new-app-create-input",
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInput(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)
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
	appTemplateName := createAppTemplateName("app-template")

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInput(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	deleteApplicationTemplateRequest := fixtures.FixDeleteApplicationTemplateRequest(appTemplate.ID)
	deleteOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Delete application template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteApplicationTemplateRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application template was deleted")

	out := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate.ID)

	require.Empty(t, out)
	saveExample(t, deleteApplicationTemplateRequest.Query(), "delete application template")
}

func TestQueryApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := createAppTemplateName("app-template")

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInput(name)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTemplate)

	getApplicationTemplateRequest := fixtures.FixApplicationTemplateRequest(appTemplate.ID)
	output := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Get application template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getApplicationTemplateRequest, &output)
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	//THEN
	t.Log("Check if application template was received")
	assert.Equal(t, name, output.Name)
}

func TestQueryApplicationTemplates(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name1 := createAppTemplateName("app-template-1")
	name2 := createAppTemplateName("app-template-2")

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application templates")
	appTmplInput1 := fixAppTemplateInput(name1)
	appTemplate1, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput1)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTemplate1)

	appTmplInput2 := fixAppTemplateInput(name2)
	appTemplate2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTemplate2)

	first := 100
	after := ""

	getApplicationTemplatesRequest := fixtures.FixGetApplicationTemplatesWithPagination(first, after)
	output := graphql.ApplicationTemplatePage{}

	// WHEN
	t.Log("List application templates")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getApplicationTemplatesRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application templates were received")
	assert.Subset(t, output.Data, []*graphql.ApplicationTemplate{&appTemplate1, &appTemplate2})
	saveExample(t, getApplicationTemplatesRequest.Query(), "query application templates")
}

func TestRegisterApplicationFromTemplate(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	appTemplateName := createAppTemplateName("template")
	appTmplInput := fixAppTemplateInput(appTemplateName)
	appTmplInput.ApplicationInput.Description = ptr.String("test {{display-name}}")
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "name",
			Description: ptr.String("name"),
		},
		{
			Name:        "display-name",
			Description: ptr.String("display-name"),
		},
	}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTmpl)
	require.NoError(t, err)

	appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{
		{
			Placeholder: "name",
			Value:       "new-name",
		},
		{
			Placeholder: "display-name",
			Value:       "new-display-name",
		}}}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}
	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createAppFromTmplRequest, &outputApp)

	//THEN
	require.NoError(t, err)
	fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenantId, outputApp.ID)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-display-name", *outputApp.Application.Description)
	saveExample(t, createAppFromTmplRequest.Query(), "register application from template")
}

func TestAddWebhookToApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := "app-template"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, name)
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInput(name)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, &appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

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
		require.Contains(t, err.Error(), "unknown parent for entity type webhook")
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

func fixAppTemplateInput(name string) graphql.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[conf.SelfRegDistinguishLabelKey] = []interface{}{conf.SelfRegDistinguishLabelValue}
	input.Labels[tenantfetcher.RegionKey] = conf.SelfRegRegion

	return input
}

func createAppTemplateName(name string) string {
	return fmt.Sprintf("SAP %s (%s)", name, conf.SelfRegRegion)
}
