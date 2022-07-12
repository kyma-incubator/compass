package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

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
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		appTemplateName := createAppTemplateName("app-template-name")
		appTemplateInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(appTemplateName)
		appTemplate, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput)
		require.NoError(t, err)

		createApplicationTemplateRequest := fixtures.FixCreateApplicationTemplateRequest(appTemplate)
		output := graphql.ApplicationTemplate{}

		// WHEN
		t.Log("Create application template")
		err = testctx.Tc.RunOperationNoTenant(ctx, certSecuredGraphQLClient, createApplicationTemplateRequest, &output)
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

		appTemplateInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
		appTemplateInput.Labels["global_subaccount_id"] = conf.ConsumerID
		appTemplateInput.ApplicationInput.Labels["applicationType"] = fmt.Sprintf("%s (%s)", appTemplateName, conf.SubscriptionConfig.SelfRegRegion)

		require.NoError(t, err)
		require.NotEmpty(t, appTemplateOutput)
		assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
		saveExample(t, getApplicationTemplateRequest.Query(), "query application template")
	})

	t.Run("Error when Self Registered Application Template already exists for a given region and distinguished label key", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		appTemplateName1 := createAppTemplateName("app-template-name-self-reg-1")
		appTemplateInput1 := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(appTemplateName1)
		appTemplate1, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput1)
		require.NoError(t, err)

		createApplicationTemplateRequest1 := fixtures.FixCreateApplicationTemplateRequest(appTemplate1)
		output1 := graphql.ApplicationTemplate{}

		appTemplateName2 := createAppTemplateName("app-template-name-self-reg-2")
		appTemplateInput2 := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(appTemplateName2)
		appTemplate2, err := testctx.Tc.Graphqlizer.ApplicationTemplateInputToGQL(appTemplateInput2)
		require.NoError(t, err)

		createApplicationTemplateRequest2 := fixtures.FixCreateApplicationTemplateRequest(appTemplate2)
		output2 := graphql.ApplicationTemplate{}

		t.Log("Create first application template")
		err = testctx.Tc.RunOperationNoTenant(ctx, certSecuredGraphQLClient, createApplicationTemplateRequest1, &output1)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &output1)

		require.NoError(t, err)
		require.NotEmpty(t, output1.ID)

		// WHEN
		t.Log("Create second application template")
		err = testctx.Tc.RunOperationNoTenant(ctx, certSecuredGraphQLClient, createApplicationTemplateRequest2, &output2)
		defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &output2)

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
	appTemplateInput := fixAppTemplateInputWithRegion(appTemplateName, conf.SubscriptionConfig.SelfRegRegion)
	appTemplateInput.ApplicationInput.Labels["applicationType"] = fmt.Sprintf("%s (%s)", appTemplateName, conf.SubscriptionConfig.SelfRegRegion)

	// WHEN
	t.Log("Create application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &appTemplate)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)
	require.NotEmpty(t, appTemplate.Name)

	t.Log("Check if application template was created")
	appTemplateOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplate.ID)
	appTemplateInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
	appTemplateInput.Labels["global_subaccount_id"] = conf.ConsumerID

	require.NotEmpty(t, appTemplateOutput)
	assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
}

func TestCreateApplicationTemplate_InvalidApplicationTypeLabel(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	appTemplateInput := fixAppTemplateInputWithRegion("SAP app-template", conf.SubscriptionConfig.SelfRegRegion)
	appTemplateInput.ApplicationInput.Labels["applicationType"] = "random-app-type"

	// WHEN
	t.Log("Create application template")
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &appTemplate)

	// THEN
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "\"applicationType\" label value does not follow \"<app_template_name> (<region>)\" schema")
}

func TestCreateApplicationTemplate_SameNames(t *testing.T) {
	testCases := []struct {
		Name                 string
		AppTemplateOneName   string
		AppTemplateTwoName   string
		AppTemplateOneRegion string
		AppTemplateTwoRegion string
		ExpectError          bool
		ExpectedErrMessage   string
	}{
		{
			Name:                 "Create two application templates with same names and region",
			AppTemplateOneName:   "SAP app-template",
			AppTemplateTwoName:   "SAP app-template",
			AppTemplateOneRegion: conf.SubscriptionConfig.SelfRegRegion,
			AppTemplateTwoRegion: conf.SubscriptionConfig.SelfRegRegion,
			ExpectError:          true,
			ExpectedErrMessage:   "application template with name \"SAP app-template\" already exists",
		},
		{
			Name:                 "Create two application templates with same names and different regions",
			AppTemplateOneName:   "SAP app-template",
			AppTemplateTwoName:   "SAP app-template",
			AppTemplateOneRegion: conf.SubscriptionConfig.SelfRegRegion,
			AppTemplateTwoRegion: conf.SubscriptionConfig.SelfRegRegion2,
			ExpectError:          false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			appTemplateOneInput := fixAppTemplateInputWithRegion(testCase.AppTemplateOneName, testCase.AppTemplateOneRegion)

			t.Log("Create first application template")
			appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOneInput)
			defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &appTemplateOne)

			//THEN
			require.NoError(t, err)
			require.NotEmpty(t, appTemplateOne.ID)
			require.NotEmpty(t, appTemplateOne.Name)

			t.Log("Check if application template one was created")
			appTemplateOneOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOne.ID)

			appTemplateOneInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateOneOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
			appTemplateOneInput.Labels["global_subaccount_id"] = conf.ConsumerID
			appTemplateOneInput.ApplicationInput.Labels["applicationType"] = fmt.Sprintf("%s (%s)", testCase.AppTemplateOneName, testCase.AppTemplateOneRegion)

			require.NotEmpty(t, appTemplateOneOutput)
			assertions.AssertApplicationTemplate(t, appTemplateOneInput, appTemplateOneOutput)

			appTemplateTwoInput := fixAppTemplateInputWithRegionAndDistinguishLabel(testCase.AppTemplateTwoName, testCase.AppTemplateTwoRegion, "other-distinguished-label")

			t.Log("Create second application template")
			appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwoInput)
			defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &appTemplateTwo)

			if testCase.ExpectError {
				require.NotNil(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, appTemplateTwo.ID)
				require.NotEmpty(t, appTemplateTwo.Name)

				t.Log("Check if application template two was created")
				appTemplateTwoOutput := fixtures.GetApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwo.ID)

				appTemplateTwoInput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey] = appTemplateTwoOutput.Labels[conf.SubscriptionConfig.SelfRegisterLabelKey]
				appTemplateTwoInput.Labels["global_subaccount_id"] = conf.ConsumerID
				appTemplateTwoInput.ApplicationInput.Labels["applicationType"] = fmt.Sprintf("%s (%s)", testCase.AppTemplateTwoName, testCase.AppTemplateTwoRegion)

				require.NotEmpty(t, appTemplateTwoOutput)
				assertions.AssertApplicationTemplate(t, appTemplateTwoInput, appTemplateTwoOutput)
			}
		})
	}
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
			AppTemplateName: fmt.Sprintf("SAP %s", "app-template-name"),
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			appTemplateInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(testCase.AppTemplateName)
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
	newName := createAppTemplateName("new-app-template")
	newDescription := "new description"
	newAppCreateInput := &graphql.ApplicationRegisterInput{
		Name:           "new-app-create-input",
		Description:    ptr.String("{{name}} {{display-name}}"),
		HealthCheckURL: ptr.String("http://url.valid"),
	}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log("Create application template")
	appTmplInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(appTemplateName)
	appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTemplate)
	require.NoError(t, err)
	require.NotEmpty(t, appTemplate.ID)

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
	appTemplateInput.ApplicationInput.Labels = map[string]interface{}{"applicationType": fmt.Sprintf("%s (%s)", newName, conf.SubscriptionConfig.SelfRegRegion)}

	require.NoError(t, err)
	require.NotEmpty(t, updateOutput.ID)

	//THEN
	t.Log("Check if application template was updated")
	assertions.AssertUpdateApplicationTemplate(t, appTemplateInput, updateOutput)

	saveExample(t, updateAppTemplateRequest.Query(), "update application template")
}

func TestUpdateApplicationTemplate_AlreadyExistsInTheSameRegion(t *testing.T) {
	ctx := context.Background()
	appTemplateOneInput := fixAppTemplateInputWithRegion("SAP app-template", conf.SubscriptionConfig.SelfRegRegion)

	t.Log("Create first application template")
	appTemplateOne, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateOneInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &appTemplateOne)

	require.NoError(t, err)
	require.NotEmpty(t, appTemplateOne.ID)
	require.NotEmpty(t, appTemplateOne.Name)

	appTemplateTwoInput := fixAppTemplateInputWithRegionAndDistinguishLabel("SAP app-template-two", conf.SubscriptionConfig.SelfRegRegion, "other-label")

	t.Log("Create second application template")
	appTemplateTwo, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appTemplateTwoInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &appTemplateTwo)

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
	require.Contains(t, err.Error(), "application template with name \"SAP app-template\" already exists")
}

func TestUpdateApplicationTemplate_NotValid(t *testing.T) {
	namePlaceholder := "name-placeholder"
	displayNamePlaceholder := "display-name-placeholder"

	testCases := []struct {
		Name                       string
		NewAppTemplateName         string
		NewAppTemplatePlaceholders []*graphql.PlaceholderDefinitionInput
		AppInputDescription        *string
		ExpectedErrMessage         string
	}{
		{
			Name:               "not compliant name",
			NewAppTemplateName: "not-compliant-name",
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
				{
					Name:        "name",
					Description: &namePlaceholder,
				},
				{
					Name:        "display-name",
					Description: &displayNamePlaceholder,
				},
			},
			AppInputDescription: ptr.String("test {{display-name}}"),
			ExpectedErrMessage:  "application template name \"not-compliant-name\" does not comply with the following naming convention",
		},
		{
			Name:               "not compliant placeholders",
			NewAppTemplateName: fmt.Sprintf("SAP %s (%s)", "app-template-name", conf.SubscriptionConfig.SelfRegRegion),
			NewAppTemplatePlaceholders: []*graphql.PlaceholderDefinitionInput{
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.Background()
			appTemplateName := createAppTemplateName("app-template")
			tenantId := tenant.TestTenants.GetDefaultTenantID()

			t.Log("Create application template")
			appTmplInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(appTemplateName)
			appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
			defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTemplate)

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
			appRegisterInput.Description = testCase.AppInputDescription
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
	appTmplInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(appTemplateName)
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
	appTmplInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(name)
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
	appTmplInput1 := fixAppTemplateInputWithRegion(name1, conf.SubscriptionConfig.SelfRegRegion)
	appTemplate1, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput1)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTemplate1)
	require.NoError(t, err)

	appTmplInput2 := fixAppTemplateInputWithRegion(name2, conf.SubscriptionConfig.SelfRegRegion2)
	appTemplate2, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput2)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTemplate2)
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
	assert.Subset(t, output.Data, []*graphql.ApplicationTemplate{&appTemplate1, &appTemplate2})
	saveExample(t, getApplicationTemplatesRequest.Query(), "query application templates")
}

func TestRegisterApplicationFromTemplate(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	appTemplateName := createAppTemplateName("template")
	appTmplInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(appTemplateName)
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

func TestRegisterApplicationFromTemplate_DifferentSubaccount(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	appTemplateName := createAppTemplateName("template")
	appTmplInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(appTemplateName)
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

	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.ExternalCertProviderConfig.CertSvcInstanceTestSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, conf.ExternalCertProviderConfig.TestExternalCertOU, conf.ExternalCertProviderConfig.TestExternalCertOU2, -1),
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
	}
	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)

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
	require.Contains(t, err.Error(), "application template with name \"SAP template\" and consumer id \"bad76f69-e5c2-4d55-bca5-240944824b83\" not found")
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
	appTmplInput := fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(name)
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

func fixAppTemplateInputWithDefaultRegionAndDistinguishLabel(name string) graphql.ApplicationTemplateInput {
	return fixAppTemplateInputWithRegion(name, conf.SubscriptionConfig.SelfRegRegion)
}

func fixAppTemplateInputWithRegion(name, region string) graphql.ApplicationTemplateInput {
	input := fixtures.FixApplicationTemplate(name)
	input.Labels[tenantfetcher.RegionKey] = region
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = conf.SubscriptionConfig.SelfRegDistinguishLabelValue

	return input
}

func fixAppTemplateInputWithRegionAndDistinguishLabel(name, region, distinguishedLabel string) graphql.ApplicationTemplateInput {
	input := fixAppTemplateInputWithRegion(name, region)
	input.Labels[conf.SubscriptionConfig.SelfRegDistinguishLabelKey] = distinguishedLabel

	return input
}

func createAppTemplateName(name string) string {
	return fmt.Sprintf("SAP %s", name)
}
