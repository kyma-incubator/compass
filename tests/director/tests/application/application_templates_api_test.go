package application

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/tests/director/tests/example"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

	"github.com/kyma-incubator/compass/tests/pkg/token"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

var (
	subject            = "C=DE, L=E2E-test, O=E2E-Org, OU=TestRegion, OU=E2E-Org-Unit, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=E2E-test-compass"
	subjectTwo         = "C=DE, L=E2E-test, O=E2E-Org, OU=TestRegion, OU=E2E-Org-Unit, OU=3c0fe289-bb13-4814-ac49-ac88c4a76b10, CN=E2E-test-compass"
	consumerType       = "Integration System"          // should be a valid consumer type
	tenantAccessLevels = []string{"account", "global"} // should be a valid tenant access level
)

func TestCreateApplicationTemplate(t *testing.T) {
	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()
	t.Run("Success for global template", func(t *testing.T) {
		// GIVEN

		// Our graphql Timestamp object parses data to RFC3339 which does not include milliseconds. This may cause the test to fail if it executes in less than a second
		testStartTime := time.Now().Add(-1 * time.Minute)

		ctx := context.Background()
		appTemplateName := fixtures.CreateAppTemplateName("app-template-name")
		appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)

		appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
		require.NoError(t, err)

		createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
		output := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create application template")
		err = testctx.Tc.RunOperationNoTenant(ctx, certSecuredGraphQLClient, createApplicationTemplateRequest, &output)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, "", output)

		//THEN
		require.NoError(t, err)
		require.NotEmpty(t, output.ID)
		require.NotEmpty(t, output.Name)

		t.Log("Check if application template was created")

		getApplicationTemplateRequest := fixtures.FixApplicationTemplateRequest(output.ID)
		appTemplateOutput := graphql.ApplicationTemplate{}

		err = testctx.Tc.RunOperationNoTenant(ctx, certSecuredGraphQLClient, getApplicationTemplateRequest, &appTemplateOutput)

		appTemplateInput.ApplicationInput.Labels["applicationType"] = appTemplateName

		require.NoError(t, err)
		require.NotEmpty(t, appTemplateOutput)
		assert.True(t, time.Time(appTemplateOutput.CreatedAt).After(testStartTime))
		assert.True(t, time.Time(appTemplateOutput.UpdatedAt).After(testStartTime))

		assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
	})

	t.Run("Success for template with product label created with certificate", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		productLabelValue := "productLabelValue"
		appTemplateName := fixtures.CreateAppTemplateName("app-template-name-product")
		appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)
		appTemplateInput.Labels[conf.ApplicationTemplateProductLabel] = productLabelValue
		appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
		require.NoError(t, err)

		createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
		output := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationTemplateRequest, &output)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, output)

		// THEN
		require.Equal(t, output.Labels[conf.ApplicationTemplateProductLabel], productLabelValue)
		require.NoError(t, err)
		require.NotEmpty(t, output.ID)
	})

	t.Run("Error for self register when distinguished label or product label have not been defined and the call is made with a certificate", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		appTemplateName := fixtures.CreateAppTemplateName("app-template-name-invalid")
		appTemplateInput := fixtures.FixApplicationTemplate(appTemplateName)
		appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
		require.NoError(t, err)

		createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
		output := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationTemplateRequest, &output)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, output)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("missing %q or %q label", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.ApplicationTemplateProductLabel))
	})

	t.Run("Error for self register when distinguished label and product label have been defined and the call is made with a certificate", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		appTemplateName := fixtures.CreateAppTemplateName("app-template-name-invalid")
		appTemplateInputInvalid := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		appTemplateInputInvalid.Labels[conf.ApplicationTemplateProductLabel] = "test1"

		appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInputInvalid)
		require.NoError(t, err)

		createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
		output := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationTemplateRequest, &output)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, output)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("should provide either %q or %q label - providing both at the same time is not allowed", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.ApplicationTemplateProductLabel))
	})

	t.Run("Error when Self Registered Application Template already exists for a given region and distinguished label key", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		appTemplateName1 := fixtures.CreateAppTemplateName("app-template-name-self-reg-1")
		appTemplateInput1 := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName1, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		appTemplate1, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput1)
		require.NoError(t, err)

		createApplicationTemplateRequest1 := fixtures.FixCreateApplicationTemplateRequest(appTemplate1)
		output1 := graphql.ApplicationTemplate{}

		appTemplateName2 := fixtures.CreateAppTemplateName("app-template-name-self-reg-2")
		appTemplateInput2 := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName2, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
		appTemplate2, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput2)
		require.NoError(t, err)

		createApplicationTemplateRequest2 := fixtures.FixCreateApplicationTemplateRequest(appTemplate2)
		output2 := graphql.ApplicationTemplate{}

		t.Log("Create first application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationTemplateRequest1, &output1)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, output1)

		require.NoError(t, err)
		require.NotEmpty(t, output1.ID)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, output1.Labels[tenantfetcher.RegionKey])

		// WHEN
		t.Log("Create second application template")
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationTemplateRequest2, &output2)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, output2)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("Cannot have more than one application template with labels %q: %q and %q: %q", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue, tenantfetcher.RegionKey, conf.SubscriptionConfig.SelfRegRegion))
		require.Empty(t, output2.ID)
	})

	t.Run("Error for self register when not using the certificate flow", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		tenantId := tenant.TestTenants.GetDefaultTenantID()
		name := "app-template-name-invalid-flow"

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

		appTemplateName := fixtures.CreateAppTemplateName(name)
		appTemplateInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)

		// WHEN
		t.Log("Creating application template")
		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)

		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("label %s is forbidden when creating Application Template in a non-cert flow", conf.SubscriptionConfig.SelfRegDistinguishLabelKey))
	})
}

func TestCreateApplicationTemplate_ValidApplicationTypeLabel(t *testing.T) {

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()
	// GIVEN
	ctx := context.Background()
	appTemplateName := "SAP app-template"
	appTemplateInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTemplateInput.ApplicationInput.Labels["applicationType"] = appTemplateName

	// WHEN
	t.Log("Create application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplate)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Name)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTemplate.Labels[tenantfetcher.RegionKey])

	t.Log("Check if application template was created")
	appTemplateOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplate.ID)
	appTemplateInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateInput.Labels[conf.GlobalSubaccountIDLabelKey] = conf.ConsumerID
	appTemplateInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion

	require.NotEmpty(t, appTemplateOutput)
	assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
}

func TestCreateApplicationTemplate_InvalidApplicationTypeLabel(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel("SAP app-template", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTemplateInput.ApplicationInput.Labels["applicationType"] = "random-app-type"

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	// WHEN
	t.Log("Create application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplate)

	// THEN
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"applicationType\" label value does not match the application template name")
}

func TestCreateApplicationTemplate_SameNamesAndRegion(t *testing.T) {
	ctx := context.Background()
	appTemplateName := "SAP app-template"
	appTemplateRegion := conf.SubscriptionConfig.SelfRegRegion
	appTemplateOneInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateOne)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	t.Log("Check if application template one was created")
	appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateOne.ID)

	appTemplateOneInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOneOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateOneInput.Labels[conf.GlobalSubaccountIDLabelKey] = conf.ConsumerID
	appTemplateOneInput.ApplicationInput.Labels["applicationType"] = appTemplateName
	appTemplateOneInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion

	require.NotEmpty(t, appTemplateOneOutput)
	assertions.AssertApplicationTemplate(t, appTemplateOneInput, appTemplateOneOutput)

	appTemplateTwoInput := fixAppTemplateInputWithDistinguishLabel(appTemplateName, "other-distinguished-label")

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateTwo)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("application template with name \"SAP app-template\" and region %s already exists", appTemplateRegion))
}

func TestCreateApplicationTemplate_SameNamesAndDifferentRegions(t *testing.T) {
	ctx := context.Background()
	appTemplateName := "SAP app-template"
	appTemplateOneInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateOne)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	t.Log("Check if application template one was created")
	appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateOne.ID)

	appTemplateOneInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOneOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateOneInput.Labels[conf.GlobalSubaccountIDLabelKey] = conf.ConsumerID
	appTemplateOneInput.ApplicationInput.Labels["applicationType"] = appTemplateName
	appTemplateOneInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion

	require.NotEmpty(t, appTemplateOneOutput)
	assertions.AssertApplicationTemplate(t, appTemplateOneInput, appTemplateOneOutput)

	appTemplateTwoInput := fixAppTemplateInputWithDistinguishLabel(appTemplateName, "other-distinguished-label")

	directorCertClientForAnotherRegion := createDirectorCertClientForAnotherRegion(t, ctx)

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertClientForAnotherRegion, tenantID, appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertClientForAnotherRegion, tenantID, appTemplateTwo)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateTwo.ID)
	require.NotEmpty(t, appTemplateTwo.Name)

	t.Log("Check if application template two was created")
	appTemplateTwoOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateTwo.ID)

	appTemplateTwoInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateTwoOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateTwoInput.Labels[conf.GlobalSubaccountIDLabelKey] = conf.TestProviderSubaccountIDRegion2
	appTemplateTwoInput.ApplicationInput.Labels["applicationType"] = appTemplateName
	appTemplateTwoInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion2

	require.NotEmpty(t, appTemplateTwoOutput)
	assertions.AssertApplicationTemplate(t, appTemplateTwoInput, appTemplateTwoOutput)
}

func TestCreateApplicationTemplate_NotValid(t *testing.T) {
	namePlaceholder := "name-placeholder"
	displayNamePlaceholder := "display-name-placeholder"
	sapProvider := "SAP"
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	testCases := []struct {
		Name                                  string
		AppTemplateName                       string
		AppTemplateAppInputJSONNameProperty   *string
		AppTemplateAppInputJSONLabelsProperty *map[string]interface{}
		AppTemplatePlaceholders               []*graphql.PlaceholderDefinitionInput
		AppInputDescription                   *string
		ExpectedErrMessage                    string
	}{
		{
			Name:                                  "not compliant name",
			AppTemplateName:                       "not-compliant-name",
			AppTemplateAppInputJSONNameProperty:   str.Ptr("not-compliant-name"),
			AppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name"), "name": "{{name}}", "displayName": "{{display-name}}"},
			AppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
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
			AppInputDescription: nil,
			ExpectedErrMessage:  "application template name \"not-compliant-name\" does not comply with the following naming convention",
		},
		{
			Name:                                  "missing mandatory applicationInput name property",
			AppTemplateName:                       fmt.Sprintf("SAP %s", "app-template-name"),
			AppTemplateAppInputJSONNameProperty:   str.Ptr(""),
			AppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name"), "displayName": "{{display-name}}"},
			AppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
					JSONPath:    &displayNamePlaceholder,
				},
			},
			AppInputDescription: ptr.String("test {{not-compliant}}"),
			ExpectedErrMessage:  "Invalid data ApplicationTemplateInput [appInput=name: cannot be blank.]",
		},
		{
			Name:                                  "missing mandatory applicationInput displayName label property",
			AppTemplateName:                       fmt.Sprintf("SAP %s", "app-template-name"),
			AppTemplateAppInputJSONNameProperty:   str.Ptr("test-app"),
			AppTemplateAppInputJSONLabelsProperty: &map[string]interface{}{"applicationType": fmt.Sprintf("SAP %s", "app-template-name")},
			AppTemplatePlaceholders:               []*graphql.PlaceholderDefinitionInput{},
			AppInputDescription:                   ptr.String("test {{not-compliant}}"),
			ExpectedErrMessage:                    "applicationInputJSON name property or applicationInputJSON displayName label is missing. They must be present in order to proceed.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			appTemplateInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(testCase.AppTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
			if testCase.AppInputDescription != nil {
				appTemplateInput.ApplicationInput.Description = testCase.AppInputDescription
			}
			appTemplateInput.Placeholders = testCase.AppTemplatePlaceholders
			if testCase.AppTemplateAppInputJSONNameProperty != nil {
				appTemplateInput.ApplicationInput.Name = *testCase.AppTemplateAppInputJSONNameProperty
			}
			appTemplateInput.ApplicationInput.ProviderName = &sapProvider

			if testCase.AppTemplateAppInputJSONLabelsProperty != nil {
				appTemplateInput.ApplicationInput.Labels = *testCase.AppTemplateAppInputJSONLabelsProperty
			}
			appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
			require.NoError(t, err)

			createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
			output := graphql.ApplicationTemplate{}

			// WHEN
			t.Log("Create application template")
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationTemplateRequest, &output)
			defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, output)

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
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create application template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTmplInput.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  ptr.String("http://url.com"),
	}}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Webhooks)
	oldWebhokCount := len(appTemplate.Webhooks)
	oldWebhokID := appTemplate.Webhooks[0].ID
	oldWebhokUrl := appTemplate.Webhooks[0].URL

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
	appTemplateInput.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  ptr.String("http://url2.com"),
	}}

	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template without override and one webhook")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	require.NotEmpty(t, updateOutput.Webhooks)
	newWebhokCount := len(updateOutput.Webhooks)
	newWebhokID := updateOutput.Webhooks[0].ID
	newWebhokUrl := updateOutput.Webhooks[0].URL

	require.Equal(t, oldWebhokCount, newWebhokCount)
	require.NotEqual(t, oldWebhokID, newWebhokID)
	require.NotEqual(t, oldWebhokUrl, newWebhokUrl)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	// Our graphql Timestamp object parses data to RFC3339 which does not include milliseconds. This may cause the test
	// to fail if it executes in less than a second. We add 1 second in order to insure a difference in the timestamps
	assert.True(t, time.Time(updateOutput.UpdatedAt).Add(1*time.Second).After(time.Time(updateOutput.CreatedAt)))

	example.SaveExample(t, updateAppTemplateRequest.Query(), "update application template")
}

func TestUpdateApplicationTemplateWithOverride(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create application template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTmplInput.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  ptr.String("http://url.com"),
	}}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Webhooks)

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
	appTemplateInput.Webhooks = []*graphql.WebhookInput{}

	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateWithOverrideRequest(appTemplate.ID, true, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template with override and empty list of webhooks")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)
	require.Empty(t, updateOutput.Webhooks)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	// Our graphql Timestamp object parses data to RFC3339 which does not include milliseconds. This may cause the test
	// to fail if it executes in less than a second. We add 1 second in order to insure a difference in the timestamps
	assert.True(t, time.Time(updateOutput.UpdatedAt).Add(1*time.Second).After(time.Time(updateOutput.CreatedAt)))
}

func TestUpdateApplicationTemplateWithoutOverrideWithoutWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create application template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTmplInput.Webhooks = []*graphql.WebhookInput{{
		Type: graphql.WebhookTypeConfigurationChanged,
		URL:  ptr.String("http://url.com"),
	}}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Webhooks)
	oldWebhokCount := len(appTemplate.Webhooks)
	oldWebhokID := appTemplate.Webhooks[0].ID
	oldWebhokUrl := appTemplate.Webhooks[0].URL

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
	appTemplateInput.Webhooks = []*graphql.WebhookInput{}

	appTemplateGQL, err := testctx.Tc.Graphqlizer.ApplicationTemplateUpdateInputToGQL(appTemplateInput)
	require.NoError(t, err)

	updateAppTemplateRequest := fixtures.FixUpdateApplicationTemplateRequest(appTemplate.ID, appTemplateGQL)
	updateOutput := graphql.ApplicationTemplate{}

	// WHEN
	t.Log("Update application template without override and empty list of webhooks")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateAppTemplateRequest, &updateOutput)
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": newName, "displayName": "{{display-name}}"}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	require.NotEmpty(t, updateOutput.Webhooks)
	newWebhokCount := len(updateOutput.Webhooks)
	newWebhokID := updateOutput.Webhooks[0].ID
	newWebhokUrl := updateOutput.Webhooks[0].URL

	require.Equal(t, oldWebhokCount, newWebhokCount)
	require.Equal(t, oldWebhokID, newWebhokID)
	require.Equal(t, oldWebhokUrl, newWebhokUrl)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	// Our graphql Timestamp object parses data to RFC3339 which does not include milliseconds. This may cause the test
	// to fail if it executes in less than a second. We add 1 second in order to insure a difference in the timestamps
	assert.True(t, time.Time(updateOutput.UpdatedAt).Add(1*time.Second).After(time.Time(updateOutput.CreatedAt)))
}

func TestUpdateLabelsOfApplicationTemplateFailsWithInsufficientScopes(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		Labels:         map[string]interface{}{"displayName": "{{display-name}}"},
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create application template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
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
	require.NoError(t, err)

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
	appTemplateName := fixtures.CreateAppTemplateName("app-template")
	newName := fixtures.CreateAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationJSONInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		Labels:         map[string]interface{}{"displayName": "{{display-name}}"},
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	firstTenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()
	secondTenantId := tenant.TestTenants.List()[1].ExternalTenant

	t.Log("Create application template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
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
	assert.Equal(t, outputAppSecondTenant.ID, app2.ID)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	t.Log("Check if applicationType label of application for the first tenant was updated")
	assert.Equal(t, app1.Labels["applicationType"], newName)

	t.Log("Check if applicationType label of application for the second tenant was updated")
	assert.Equal(t, app2.Labels["applicationType"], newName)
}

func TestUpdateApplicationTemplate_AlreadyExistsInTheSameRegion(t *testing.T) {
	ctx := context.Background()
	appTemplateRegion := conf.SubscriptionConfig.SelfRegRegion
	appTemplateOneInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel("SAP app-template", conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)

	tenantID := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateOne)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	appTemplateTwoInput := fixAppTemplateInputWithDistinguishLabel("SAP app-template-two", "other-label")

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantID, appTemplateTwo)

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
	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

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
			appTemplateName := fixtures.CreateAppTemplateName("app-template")

			t.Log("Create application template")
			appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
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
	appTemplateName := fixtures.CreateAppTemplateName("app-template")

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create application template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
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
	example.SaveExample(t, deleteApplicationTemplateRequest.Query(), "delete application template")
}

func TestDeleteApplicationTemplateWithCertSubjMapping(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("app-template")

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Logf("Create application template with name %q", appTemplateName)
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	t.Logf("Create certificate subject mapping one for application template with name: %q and id: %q", appTemplate.Name, appTemplate.ID)
	csmInputOne := fixtures.FixCertificateSubjectMappingInput(subject, consumerType, &appTemplate.ID, tenantAccessLevels)

	var csmCreateOne graphql.CertificateSubjectMapping
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreateOne)
	csmCreateOne = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInputOne)

	t.Logf("Create certificate subject mapping two for application template with name: %q and id: %q", appTemplate.Name, appTemplate.ID)
	csmInputTwo := fixtures.FixCertificateSubjectMappingInput(subjectTwo, consumerType, &appTemplate.ID, tenantAccessLevels)

	var csmCreateTwo graphql.CertificateSubjectMapping
	defer fixtures.CleanupCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, &csmCreateTwo)
	csmCreateTwo = fixtures.CreateCertificateSubjectMapping(t, ctx, certSecuredGraphQLClient, csmInputTwo)

	// WHEN
	t.Logf("Delete application template with id %q", appTemplate.ID)

	deleteApplicationTemplateRequest := fixtures.FixDeleteApplicationTemplateRequest(appTemplate.ID)
	deleteOutput := graphql.ApplicationTemplate{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteApplicationTemplateRequest, &deleteOutput)
	require.NoError(t, err)

	//THEN
	t.Logf("Check if application template with id %q was deleted", appTemplate.ID)

	out := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate.ID)

	require.Empty(t, out)

	t.Log("Check if all certificate subject mappings were deleted")

	t.Logf("Query certificate subject mapping by ID: %s", csmCreateOne.ID)
	queryCertSubjectMappingReq := fixtures.FixQueryCertificateSubjectMappingRequest(csmCreateOne.ID)
	csm := graphql.CertificateSubjectMapping{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryCertSubjectMappingReq, &csm)

	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "Object not found")

	t.Logf("Query certificate subject mapping by ID: %s", csmCreateTwo.ID)
	queryCertSubjectMappingReq = fixtures.FixQueryCertificateSubjectMappingRequest(csmCreateTwo.ID)
	csm = graphql.CertificateSubjectMapping{}
	err = testctx.Tc.RunOperationWithoutTenant(ctx, certSecuredGraphQLClient, queryCertSubjectMappingReq, &csm)

	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "Object not found")
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

	appTemplateName := fixtures.CreateAppTemplateName(name)
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
	name := fixtures.CreateAppTemplateName("app-template")

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create application template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(name, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
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
	name1 := fixtures.CreateAppTemplateName("app-template-1")
	name2 := fixtures.CreateAppTemplateName("app-template-2")

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create application templates")
	appTmplInput1 := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(name1, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTemplate1, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput1)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplate1)
	require.NoError(t, err)

	directorCertClientRegion2 := createDirectorCertClientForAnotherRegion(t, ctx)

	appTmplInput2 := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(name2, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
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
	example.SaveExample(t, getApplicationTemplatesRequest.Query(), "query application templates")
}

func TestRegisterApplicationFromTemplate(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"
	appTemplateName := fixtures.CreateAppTemplateName("template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
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

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

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
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-display-name", *outputApp.Application.Description)
	example.SaveExample(t, createAppFromTmplRequest.Query(), "register application from template")
}

func TestRegisterApplicationFromTemplateWithTemplateID(t *testing.T) {
	//GIVEN
	ctx := context.Background()
	appTemplateName := fixtures.CreateAppTemplateName("template")
	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Create application template in the first region")
	appTemplateOneInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplateOne)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.Equal(t, appTemplateName, appTemplateOne.Name)

	t.Log("Check if application template in the first region was created")
	appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOneOutput)

	t.Log("Create application template in the second region")
	appTemplateTwoInput := fixAppTemplateInputWithDistinguishLabel(appTemplateName, "other-distinguished-label")
	directorCertClientForAnotherRegion := createDirectorCertClientForAnotherRegion(t, ctx)
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertClientForAnotherRegion, tenantId, appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertClientForAnotherRegion, tenantId, appTemplateTwo)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateTwo.ID)
	require.Equal(t, appTemplateName, appTemplateTwo.Name)

	t.Log("Check if application template in the second region was created")
	appTemplateTwoOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, appTemplateTwo.ID)
	require.NotEmpty(t, appTemplateTwoOutput)

	require.NotEqual(t, appTemplateOne.ID, appTemplateTwo.ID)

	t.Log("Create application using template id")
	appFromTmpl := graphql.ApplicationFromTemplateInput{
		ID:           &appTemplateTwo.ID,
		TemplateName: appTemplateTwo.Name,
		Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "name",
				Value:       "app-name",
			},
			{
				Placeholder: "display-name",
				Value:       "app-display-name",
			},
		},
	}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputWithTemplateIDToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
	outputApp := graphql.ApplicationExt{}

	//WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.Equal(t, appTemplateTwo.ID, *outputApp.Application.ApplicationTemplateID)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test app-display-name", *outputApp.Application.Description)
	example.SaveExample(t, createAppFromTmplRequest.Query(), "register application from template using template name and id")
}

func TestRegisterApplicationFromTemplatewithPlaceholderPayload(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	nameJSONPath := "$.name"
	displayNameJSONPath := "$.displayName"
	placeholdersPayload := `{\"name\": \"appName\", \"displayName\":\"appDisplayName\"}`
	appTemplateName := fixtures.CreateAppTemplateName("templateForPlaceholdersPayload")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
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

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

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
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "appName", outputApp.Application.Name)
	require.Equal(t, "test appDisplayName", *outputApp.Application.Description)
	example.SaveExample(t, createAppFromTmplRequest.Query(), "register application from template with placeholder payload")
}

func TestRegisterApplicationFromTemplate_DifferentSubaccount(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"
	appTemplateName := fixtures.CreateAppTemplateName("template")
	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
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

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

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
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, directorCertSecuredClient, tenantId, createAppFromTmplRequest, &outputApp)

	// THEN
	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("application template with name %q and consumer id %q not found", appTemplateName, conf.TestProviderSubaccountIDRegion2))
}

func TestAddWebhookToApplicationTemplateWithTenant(t *testing.T) {
	ctx := context.Background()
	name := "app-template"
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, name)
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	t.Log("Request Client Credentials for Integration System")
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
	appTmplInput.Labels = graphql.Labels{
		"a":                             []string{"b", "c"},
		"d":                             []string{"e", "f"},
		"displayName":                   "{{display-name}}",
		conf.GlobalSubaccountIDLabelKey: tenantId,
	}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	t.Log("Add Webhook to application template with invalid tenant")
	// add
	url := "http://new-webhook.url"
	outputTemplate := "{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"

	webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &url,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})
	require.NoError(t, err)

	actualWebhook := graphql.Webhook{}
	addReq := fixtures.FixAddWebhookToTemplateRequest(appTemplate.ID, webhookInStr)
	example.SaveExampleInCustomDir(t, addReq.Query(), example.AddWebhookCategory, "add application template webhook")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenant.TestTenants.GetSystemFetcherTenantID(), addReq, &actualWebhook)
	require.Error(t, err)
	require.Contains(t, err.Error(), "the provided tenant c395681d-11dd-4cde-bbcf-570b4a153e79 and the parent tenant 5577cf46-4f78-45fa-b55f-a42a3bdba868 do not match")

	t.Log("Add Webhook to application template with valid tenant")
	actualWebhook = graphql.Webhook{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, addReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, "http://new-webhook.url", *actualWebhook.URL)
	assert.Equal(t, graphql.WebhookTypeUnregisterApplication, actualWebhook.Type)
	id := actualWebhook.ID
	require.NotNil(t, id)

	t.Log("Get Application Template")
	updatedAppTemplate := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate.ID)
	assert.Len(t, updatedAppTemplate.Webhooks, 1)

	t.Log("Update Application Template webhook with tenant")
	urlUpdated := "http://updated-webhook.url"
	webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &urlUpdated,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})
	require.NoError(t, err)

	updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, updateReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, urlUpdated, *actualWebhook.URL)

	t.Log("Delete Application Template webhook with tenant")
	deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, deleteReq, &actualWebhook)
	require.NoError(t, err)
}

func TestAddWebhookToApplicationTemplateWithoutTenant(t *testing.T) {
	ctx := context.Background()
	name := "app-template"
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, name)
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, intSys)
	require.NoError(t, err)
	require.NotEmpty(t, intSys.ID)

	t.Log("Request Client Credentials for Integration System")
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
	appTmplInput.Labels = graphql.Labels{
		"a":                             []string{"b", "c"},
		"d":                             []string{"e", "f"},
		"displayName":                   "{{display-name}}",
		conf.GlobalSubaccountIDLabelKey: tenantId,
	}
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

	t.Log("Add Webhook to application template without tenant")
	url := "http://new-webhook.url"
	outputTemplate := "{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"
	webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &url,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})
	require.NoError(t, err)

	actualWebhook := graphql.Webhook{}
	addReq := fixtures.FixAddWebhookToTemplateRequest(appTemplate.ID, webhookInStr)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, oauthGraphQLClient, addReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, "http://new-webhook.url", *actualWebhook.URL)
	assert.Equal(t, graphql.WebhookTypeUnregisterApplication, actualWebhook.Type)
	id := actualWebhook.ID
	require.NotNil(t, id)

	t.Log("Get Application Template")
	updatedAppTemplate := fixtures.GetApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTemplate.ID)
	assert.Len(t, updatedAppTemplate.Webhooks, 1)

	t.Log("Update Application Template webhook without tenant")
	urlUpdated := "http://updated-webhook.url"
	webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
		URL:            &urlUpdated,
		Type:           graphql.WebhookTypeUnregisterApplication,
		OutputTemplate: &outputTemplate,
	})
	require.NoError(t, err)

	updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, oauthGraphQLClient, updateReq, &actualWebhook)
	require.NoError(t, err)
	assert.NotNil(t, actualWebhook.URL)
	assert.Equal(t, urlUpdated, *actualWebhook.URL)

	t.Log("Delete Application Template webhook without tenant")
	deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
	err = testctx.Tc.RunOperationWithoutTenant(ctx, oauthGraphQLClient, deleteReq, &actualWebhook)
	require.NoError(t, err)
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
}

func fixAppTemplateInputWithDefaultDistinguishLabelAndSubdomainRegion(name string) graphql.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue
	input.ApplicationInput.BaseURL = str.Ptr(fmt.Sprintf(baseURLTemplate, "{{subdomain}}", "{{region}}"))
	input.Placeholders = append(input.Placeholders, &graphql.PlaceholderDefinitionInput{Name: "subdomain"}, &graphql.PlaceholderDefinitionInput{Name: "region"})
	return input
}

func fixAppTemplateInputWithDistinguishLabel(name, distinguishedLabel string) graphql.ApplicationTemplateInput {
	return fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(name, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, distinguishedLabel)
}
