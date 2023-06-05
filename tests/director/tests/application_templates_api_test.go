package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func TestUpdateApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := createAppTemplateName("app-template")
	newName := createAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	newAppCreateInput.Labels = map[string]interface{}{"displayName": "{{display-name}}"}
	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name: "name",
		},
		{
			Name: "display-name",
		},
	}
	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	saveExample(t, updateAppTemplateRequest.Query(), "update application template")
}

func TestUpdateLabelsOfApplicationTemplateFailsWithInsufficientScopes(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := createAppTemplateName("app-template")
	newName := createAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		Labels:         map[string]interface{}{"displayName": "{{display-name}}"},
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, Labels: map[string]interface{}{"label1": "test"}, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name: "name",
		},
		{
			Name: "display-name",
		},
	}
	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	t.Log("Should return error because there is no application_template.labels:write scope")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient scopes provided")
}

func TestUpdateApplicationTypeLabelOfApplicationsWhenAppTemplateNameIsUpdated(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := createAppTemplateName("app-template")
	newName := createAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationRegisterInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		Labels:         map[string]interface{}{"displayName": "{{display-name}}"},
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	firstTenantId := tenant.TestTenants.GetDefaultTenantID()
	secondTenantId := tenant.TestTenants.List()[1].ID

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, firstTenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, firstTenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	t.Log("Create application from template for the first tenant")
	appFromTmplFirstTenant := graphql.ApplicationFromTemplateInput{
		TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "name",
				Value:       "app1-e2e-update-applicationType-label",
			},
			{
				Placeholder: "display-name",
				Value:       "app1 description",
			},
		},
	}

	appFromTmplGQLFirstTenant, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplFirstTenant)
	require.NoError(t, err)

	createAppFromTmplRequestFirstTenant := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQLFirstTenant)
	outputAppFirstTenant := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, firstTenantId, createAppFromTmplRequestFirstTenant, &outputAppFirstTenant)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, firstTenantId, &outputAppFirstTenant)
	require.NoError(t, err)

	t.Log("Create application from template for the second tenant")
	appFromTmplSecondTenant := graphql.ApplicationFromTemplateInput{
		TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "name",
				Value:       "app2-e2e-update-applicationType-label",
			},
			{
				Placeholder: "display-name",
				Value:       "app2 description",
			},
		},
	}

	appFromTmplGQLSecondTenant, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSecondTenant)
	require.NoError(t, err)

	createAppFromTmplRequestSecondTenant := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQLSecondTenant)
	outputAppSecondTenant := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, secondTenantId, createAppFromTmplRequestSecondTenant, &outputAppSecondTenant)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, secondTenantId, &outputAppSecondTenant)
	require.NoError(t, err)

	appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: newName, ApplicationInput: newAppCreateInput, Description: &newDescription, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
	appTemplateInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name: "name",
		},
		{
			Name: "display-name",
		},
	}
	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	t.Log("Get updated application for the first tenant")
	app1 := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, firstTenantId, outputAppFirstTenant.ID)
	assert.Equal(t, outputAppFirstTenant.ID, app1.ID)

	t.Log("Get updated application for the second tenant")
	app2 := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, secondTenantId, outputAppSecondTenant.ID)
	assert.Equal(t, outputAppSecondTenant.ID, app1.ID)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	t.Log("Check if applicationType label of application for the first tenant was updated")
	assert.Equal(t, app1.Labels["applicationType"], newName)

<<<<<<< Updated upstream
	saveExample(t, updateAppTemplateRequest.Query(), "update application template")
}

func TestUpdateApplicationTemplate_AlreadyExistsInTheSameRegion(t *testing.T) {
	ctx := context.Background()
	appTemplateRegion := conf.SubscriptionConfig.SelfRegRegion
	appTemplateOneInput := fixAppTemplateInputWithDefaultDistinguishLabel("SAP app-template")

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOne)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	appTemplateTwoInput := fixAppTemplateInputWithDistinguishLabel("SAP app-template-two", "other-label")

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwo)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateTwo.ID)
	require.NotEmpty(t, appTemplateTwo.Name)

	t.Log("Update the name of second application template to be the same as the name of the first one")
	appTemplateInput := graphql.ApplicationTemplateUpdateInput{
		Name:             appTemplateOne.Name,
		Description:      appTemplateOne.Description,
		ApplicationInput: appTemplateOneInput.ApplicationInput,
		Placeholders:     appTemplateOneInput.Placeholders,
		AccessLevel:      appTemplateOne.AccessLevel,
	}

	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)
	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplateTwo.ID, appTemplateGQL)

	updateOutput := graphql.ApplicationTemplate{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("application template with name \"SAP app-template\" and region %s already exists", appTemplateRegion))
}

func TestUpdateApplicationTemplate_NotValid(t *testing.T) {
	namePlaceholder := "name-placeholder"
	displayNamePlaceholder := "display-name-placeholder"
	sapProvider := "SAP"
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"

	testCases := []struct {
		Name                                     string
		NewAppTemplateName                       string
		NewAppTemplateAppInputJSONNameProperty   *string
		NewAppTemplateAppInputJSONLabelsProperty *map[string]interface{}
		NewAppTemplatePlaceholders               []*graphql.PlaceholderDefinitionInput
		AppInputDescription                      *string
		ExpectedErrMessage                       string
	}{
		{
			Name:               "not compliant name",
			NewAppTemplateName: "not-compliant-name",
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
					JSONPath:    &nameJSONPath,
				},
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNameJSONPath,
				},
			},
			AppInputDescription: ptr.String("test {{display-name}}"),
			ExpectedErrMessage:  "application template name \"not-compliant-name\" does not comply with the following naming convention",
		},
		{
			Name:                                     "missing mandatory applicationInput name property",
			NewAppTemplateName:                       fmt.Sprintf("SAP %s (%s)", "app-template-name", conf.SubscriptionConfig.SelfRegRegion),
			NewAppTemplateAppInputJSONNameProperty:   str.Ptr(""),
			NewAppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"name": "{{name}}", "displayName": "{{display-name}}"},
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
					JSONPath:    &nameJSONPath,
				},
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNameJSONPath,
				},
			},
			AppInputDescription: ptr.String("test {{not-compliant}}"),
			ExpectedErrMessage:  "Invalid data ApplicationTemplateUpdateInput [appInput=name: cannot be blank.]",
		},
		{
			Name:                                     "missing mandatory applicationInput displayName label property",
			NewAppTemplateName:                       fmt.Sprintf("SAP %s (%s)", "app-template-name", conf.SubscriptionConfig.SelfRegRegion),
			NewAppTemplateAppInputJSONNameProperty:   str.Ptr("test-app"),
			NewAppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"name": "{{name}}"},
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
					JSONPath:    &nameJSONPath,
				},
			},
			AppInputDescription: ptr.String("test {{not-compliant}}"),
			ExpectedErrMessage:  "applicationInputJSON name property or applicationInputJSON displayName label is missing. They must be present in order to proceed.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			appTemplateName := createAppTemplateName("app-template")
			tenantId := tenant.TestTenants.GetDefaultTenantID()

			t.Log("Create application template")
			appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
			appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
			defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)

			require.NoError(t, err)
			require.NotEmpty(t, appTemplate.ID)

			// WHEN
			t.Log("Update application template")
			appJSONInput := &graphql.ApplicationJSONInput{
				Name:         "{{name}}",
				ProviderName: ptr.String("compass-tests"),
				Labels: graphql.Labels{
					"a": []string{"b", "c"},
					"d": []string{"e", "f"},
				},
				Webhooks: []*graphql.WebhookInput{{
					Type: graphql.WebhookTypeConfigurationChanged,
					URL:  ptr.String("http://url.com"),
				}},
				HealthCheckURL: ptr.String("http://url.valid"),
			}

			if testCase.NewAppTemplateAppInputJSONNameProperty != nil {
				appJSONInput.Name = *testCase.NewAppTemplateAppInputJSONNameProperty
			}
			appJSONInput.Description = testCase.AppInputDescription
			appJSONInput.ProviderName = &sapProvider
			if testCase.NewAppTemplateAppInputJSONLabelsProperty != nil {
				appJSONInput.Labels = *testCase.NewAppTemplateAppInputJSONLabelsProperty
			}

			appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: testCase.NewAppTemplateName, ApplicationInput: appJSONInput, Placeholders: testCase.NewAppTemplatePlaceholders, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
			appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)

			updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
			updateOutput := graphql.ApplicationTemplate{}

			err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)

			//THEN
			require.NotNil(t, err)
			if testCase.ExpectedErrMessage != "" {
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}
		})
	}
}

func TestDeleteApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := createAppTemplateName("app-template")

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)
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

func TestDeleteApplicationTemplateBeforeDeletingAssociatedApplicationsWithIt(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	name := "template"

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, name)
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys.ID)
	require.NotEmpty(t, intSysAuth)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

	intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)

	t.Log("Issuing a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	appTemplateName := createAppTemplateName(name)
	appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)

	t.Log("Creating application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)

	appFromTemplate := graphql.ApplicationFromTemplateInput{TemplateName: appTemplateName, Values: []*graphql.TemplateValueInput{
		{
			Placeholder: "name",
			Value:       "new-name",
		},
		{
			Placeholder: "display-name",
			Value:       "new-display-name",
		}}}
	appFromTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTemplate)
	require.NoError(t, err)
	createAppFromTemplateRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTemplateGQL)
	outputApp := graphql.ApplicationExt{}

	t.Logf("Creating application from application template with id %s", appTemplate.ID)
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, createAppFromTemplateRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, tenantId, outputApp.ID)
	require.NoError(t, err)

	deleteApplicationTemplateRequest := fixtures.FixDeleteApplicationTemplateRequest(appTemplate.ID)
	deleteOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Logf("Deleting application template with id %s when application with id %s is associated with it", appTemplate.ID, outputApp.ID)
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, deleteApplicationTemplateRequest, &deleteOutput)
	require.Error(t, err)

	//THEN
	t.Logf("Checking if application template with id %s was deleted", appTemplate.ID)

	out := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate.ID)
	require.NotEmpty(t, out)

	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-display-name", *outputApp.Application.Description)
	require.Equal(t, "new-name", outputApp.Application.Name)
}

func TestQueryApplicationTemplate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	name := createAppTemplateName("app-template")

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(name)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)

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
	appTmplInput1 := fixAppTemplateInputWithDefaultDistinguishLabel(name1)
	appTemplate1, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput1)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate1)
	require.NoError(t, err)

	directorCertClientRegion2 := createDirectorCertClientForAnotherRegion(t, ctx)

	appTmplInput2 := fixAppTemplateInputWithDefaultDistinguishLabel(name2)
	appTemplate2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertClientRegion2, tenantId, appTmplInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertClientRegion2, tenantId, appTemplate2)
	require.NoError(t, err)

	first := 199
	after := ""

	getApplicationTemplatesRequest := fixtures.FixGetApplicationTemplatesWithPagination(first, after)
	output := graphql.ApplicationTemplatePage{}

	// WHEN
	t.Log("List application templates")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getApplicationTemplatesRequest, &output)
	require.NoError(t, err)

	//THEN
	t.Log("Check if application templates were received")
	appTemplateIDs := []string{appTemplate1.ID, appTemplate2.ID}
	t.Logf("Created templates are with IDs: %v ", appTemplateIDs)
	found := 0
	for _, tmpl := range output.Data {
		t.Logf("Checked template from query response is: %s ", tmpl.ID)
		if str.ContainsInSlice(appTemplateIDs, tmpl.ID) {
			found++
		}
	}
	assert.Equal(t, 2, found)
	saveExample(t, getApplicationTemplatesRequest.Query(), "query application templates")
}

func TestRegisterApplicationFromTemplate(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"
	appTemplateName := createAppTemplateName("template")
	appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
	appTmplInput.ApplicationInput.Description = ptr.String("test {{display-name}}")
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "name",
			Description: ptr.String("name"),
			JSONPath:    &nameJSONPath,
		},
		{
			Name:        "display-name",
			Description: ptr.String("display-name"),
			JSONPath:    &displayNameJSONPath,
		},
	}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTmpl)
	require.NoError(t, err)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

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
	defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-display-name", *outputApp.Application.Description)
	saveExample(t, createAppFromTmplRequest.Query(), "register application from template")
}

func TestRegisterApplicationFromTemplatewithPlaceholderPayload(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	nameJSONPath := "$.name"
	displayNameJSONPath := "$.displayName"
	placeholdersPayload := `{\"name\": \"appName\", \"displayName\":\"appDisplayName\"}`
	appTemplateName := createAppTemplateName("templateForPlaceholdersPayload")
	appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "name",
			Description: ptr.String("name"),
			JSONPath:    &nameJSONPath,
		},
		{
			Name:        "display-name",
			Description: ptr.String("display-name"),
			JSONPath:    &displayNameJSONPath,
		},
	}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTmpl)
	require.NoError(t, err)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

	appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: appTemplateName, PlaceholdersPayload: &placeholdersPayload}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}
	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "appName", outputApp.Application.Name)
	require.Equal(t, "test appDisplayName", *outputApp.Application.Description)
	saveExample(t, createAppFromTmplRequest.Query(), "register application from template with placeholder payload")
}

func TestRegisterApplicationFromTemplate_DifferentSubaccount(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"
	appTemplateName := createAppTemplateName("template")
	appTmplInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
	appTmplInput.ApplicationInput.Description = ptr.String("test {{display-name}}")
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        "name",
			Description: ptr.String("name"),
			JSONPath:    &nameJSONPath,
		},
		{
			Name:        "display-name",
			Description: ptr.String("display-name"),
			JSONPath:    &displayNameJSONPath,
		},
	}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTmpl)
	require.NoError(t, err)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

	directorCertSecuredClient := createDirectorCertClientForAnotherRegion(t, ctx)

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
	// WHEN
	err = testctx.Tc.RunOperation(ctx, directorCertSecuredClient, createAppFromTmplRequest, &outputApp)

	// THEN
	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("application template with name %q and consumer id %q not found", appTemplateName, conf.TestProviderSubaccountIDRegion2))
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
	appTmplInput := fixtures.FixApplicationTemplate(name)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
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

func createDirectorCertClientForAnotherRegion(t *testing.T, ctx context.Context) *gcli.Client {
	// Prepare provider external client certificate and secret and Build graphql director client configured with certificate
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:         conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace:    conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestRegion2SecretName:     conf.ExternalCertProviderConfig.CertSvcInstanceTestRegion2SecretName,
		ExternalCertCronjobContainerName:         conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:                  conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:                  conf.ExternalCertProviderConfig.TestExternalCertSubjectRegion2,
		ExternalClientCertCertKey:                conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:                 conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalClientCertExpectedIssuerLocality: &conf.ExternalClientCertExpectedIssuerLocalityRegion2,
		ExternalCertProvider:                     certprovider.CertificateService,
	}
	providerClientKey, providerRawCertChain := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, true)
	return gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, providerClientKey, providerRawCertChain, conf.SkipSSLValidation)
=======
	t.Log("Check if applicationType label of application for the second tenant was updated")
	assert.Equal(t, app2.Labels["applicationType"], newName)
>>>>>>> Stashed changes
}

func fixAppTemplateInputWithDefaultDistinguishLabel(name string) graphql.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue

	return input
}

func createAppTemplateName(name string) string {
	return fmt.Sprintf("SAP %s", name)
}
