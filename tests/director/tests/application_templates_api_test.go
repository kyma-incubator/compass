package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"

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
	t.Run("Success for global template", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		appTemplateName := createAppTemplateName("app-template-name")
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
		assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
	})

	t.Run("Error when Self Registered Application Template already exists for a given region and distinguished label key", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		appTemplateName1 := createAppTemplateName("app-template-name-self-reg-1")
		appTemplateInput1 := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName1)
		appTemplate1, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput1)
		require.NoError(t, err)

		createApplicationTemplateRequest1 := fixtures.FixCreateApplicationTemplateRequest(appTemplate1)
		output1 := graphql.ApplicationTemplate{}

		appTemplateName2 := createAppTemplateName("app-template-name-self-reg-2")
		appTemplateInput2 := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName2)
		appTemplate2, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput2)
		require.NoError(t, err)

		createApplicationTemplateRequest2 := fixtures.FixCreateApplicationTemplateRequest(appTemplate2)
		output2 := graphql.ApplicationTemplate{}

		t.Log("Create first application template")
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createApplicationTemplateRequest1, &output1)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), output1)

		require.NoError(t, err)
		require.NotEmpty(t, output1.ID)
		require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, output1.Labels[tenantfetcher.RegionKey])

		// WHEN
		t.Log("Create second application template")
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createApplicationTemplateRequest2, &output2)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), output2)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("Cannot have more than one application template with labels %q: %q and %q: %q", tenantfetcher.RegionKey, conf.SubscriptionConfig.SelfRegRegion, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue))
		require.Empty(t, output2.ID)
	})
}

func TestCreateApplicationTemplate_ValidApplicationTypeLabel(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateName := "SAP app-template"
	appTemplateInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)
	appTemplateInput.ApplicationInput.Labels["applicationType"] = appTemplateName

	// WHEN
	t.Log("Create application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplate)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Name)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTemplate.Labels[tenantfetcher.RegionKey])

	t.Log("Check if application template was created")
	appTemplateOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplate.ID)
	appTemplateInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateInput.Labels["global_subaccount_id"] = conf.ConsumerID
	appTemplateInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion

	require.NotEmpty(t, appTemplateOutput)
	assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
}

func TestCreateApplicationTemplate_InvalidApplicationTypeLabel(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateInput := fixAppTemplateInputWithDefaultDistinguishLabel("SAP app-template")
	appTemplateInput.ApplicationInput.Labels["applicationType"] = "random-app-type"

	// WHEN
	t.Log("Create application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplate)

	// THEN
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"applicationType\" label value does not match the application template name")
}

func TestCreateApplicationTemplate_SameNamesAndRegion(t *testing.T) {
	ctx := context.Background()
	appTemplateName := "SAP app-template"
	appTemplateRegion := conf.SubscriptionConfig.SelfRegRegion
	appTemplateOneInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOne)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	t.Log("Check if application template one was created")
	appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOne.ID)

	appTemplateOneInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOneOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateOneInput.Labels["global_subaccount_id"] = conf.ConsumerID
	appTemplateOneInput.ApplicationInput.Labels["applicationType"] = appTemplateName
	appTemplateOneInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion

	require.NotEmpty(t, appTemplateOneOutput)
	assertions.AssertApplicationTemplate(t, appTemplateOneInput, appTemplateOneOutput)

	appTemplateTwoInput := fixAppTemplateInputWithDistinguishLabel(appTemplateName, "other-distinguished-label")

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwo)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), fmt.Sprintf("application template with name \"SAP app-template\" and region %s already exists", appTemplateRegion))
}

func TestCreateApplicationTemplate_SameNamesAndDifferentRegions(t *testing.T) {
	ctx := context.Background()
	appTemplateName := "SAP app-template"
	appTemplateOneInput := fixAppTemplateInputWithDefaultDistinguishLabel(appTemplateName)

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOne)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	t.Log("Check if application template one was created")
	appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOne.ID)

	appTemplateOneInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOneOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateOneInput.Labels["global_subaccount_id"] = conf.ConsumerID
	appTemplateOneInput.ApplicationInput.Labels["applicationType"] = appTemplateName
	appTemplateOneInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion

	require.NotEmpty(t, appTemplateOneOutput)
	assertions.AssertApplicationTemplate(t, appTemplateOneInput, appTemplateOneOutput)

	appTemplateTwoInput := fixAppTemplateInputWithDistinguishLabel(appTemplateName, "other-distinguished-label")

	directorCertClientForAnotherRegion := createDirectorCertClientForAnotherRegion(t, ctx)

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertClientForAnotherRegion, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertClientForAnotherRegion, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwo)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateTwo.ID)
	require.NotEmpty(t, appTemplateTwo.Name)

	t.Log("Check if application template two was created")
	appTemplateTwoOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwo.ID)

	appTemplateTwoInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateTwoOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateTwoInput.Labels["global_subaccount_id"] = conf.TestProviderSubaccountIDRegion2
	appTemplateTwoInput.ApplicationInput.Labels["applicationType"] = appTemplateName
	appTemplateTwoInput.Labels[tenantfetcher.RegionKey] = conf.SubscriptionConfig.SelfRegRegion2

	require.NotEmpty(t, appTemplateTwoOutput)
	assertions.AssertApplicationTemplate(t, appTemplateTwoInput, appTemplateTwoOutput)
}

func TestCreateApplicationTemplate_NotValid(t *testing.T) {
	namePlaceholder := "name-placeholder"
	displayNamePlaceholder := "display-name-placeholder"
	sapProvider := "SAP"
	nameJSONPath := "name-json-path"
	displayNameJSONPath := "display-name-json-path"

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
			appTemplateInput := fixAppTemplateInputWithDefaultDistinguishLabel(testCase.AppTemplateName)
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
			err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createApplicationTemplateRequest, &output)
			defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), output)

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
	newName := createAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationRegisterInput{
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
	nameJSONPath := "name-json-path"
	displayNameJSONPath := "display-name-json-path"

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
			appRegisterInput := &graphql.ApplicationRegisterInput{
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
				appRegisterInput.Name = *testCase.NewAppTemplateAppInputJSONNameProperty
			}
			appRegisterInput.Description = testCase.AppInputDescription
			appRegisterInput.ProviderName = &sapProvider
			if testCase.NewAppTemplateAppInputJSONLabelsProperty != nil {
				appRegisterInput.Labels = *testCase.NewAppTemplateAppInputJSONLabelsProperty
			}

			appTemplateInput := graphql.ApplicationTemplateUpdateInput{Name: testCase.NewAppTemplateName, ApplicationInput: appRegisterInput, Placeholders: testCase.NewAppTemplatePlaceholders, AccessLevel: graphql.ApplicationTemplateAccessLevelGlobal}
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
	appTemplateIDs := []string{appTemplate1.ID, appTemplate2.ID}
	found := 0
	for _, tmpl := range output.Data {
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
	nameJSONPath := "name-json-path"
	displayNameJSONPath := "display-name-json-path"
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

	//THEN
	require.NoError(t, err)
	fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenantId, outputApp.ID)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "test new-display-name", *outputApp.Application.Description)
	saveExample(t, createAppFromTmplRequest.Query(), "register application from template")
}

func TestRegisterApplicationFromTemplatewithPlaceholderPayload(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	nameJSONPath := "name"
	displayNameJSONPath := "displayName"
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

	//THEN
	require.NoError(t, err)
	fixtures.UnregisterApplication(t, ctx, certSecuredGraphQLClient, tenantId, outputApp.ID)
	require.NotEmpty(t, outputApp)
	require.NotNil(t, outputApp.Application.Description)
	require.Equal(t, "appName", outputApp.Application.Name)
	require.Equal(t, "test appDisplayName", *outputApp.Application.Description)
	saveExample(t, createAppFromTmplRequest.Query(), "register application from template with placeholder payload")
}

func TestRegisterApplicationFromTemplate_DifferentSubaccount(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	nameJSONPath := "name-json-path"
	displayNameJSONPath := "display-name-json-path"
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
}

func fixAppTemplateInputWithDefaultDistinguishLabel(name string) graphql.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue

	return input
}

func fixAppTemplateInputWithDefaultDistinguishLabelAndSubdomainRegion(name string) graphql.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue
	input.ApplicationInput.BaseURL = str.Ptr(fmt.Sprintf(baseURLTemplate, "{{subdomain}}", "{{region}}"))
	input.Placeholders = append(input.Placeholders, &graphql.PlaceholderDefinitionInput{Name: "subdomain"}, &graphql.PlaceholderDefinitionInput{Name: "region"})
	return input
}

func fixAppTemplateInputWithDistinguishLabel(name, distinguishedLabel string) graphql.ApplicationTemplateInput {
	input := fixAppTemplateInputWithDefaultDistinguishLabel(name)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = distinguishedLabel

	return input
}

func createAppTemplateName(name string) string {
	return fmt.Sprintf("SAP %s", name)
}
