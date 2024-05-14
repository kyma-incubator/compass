package application

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tnt "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/tests/director/tests/example"
	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	registerApplicationCategory = "register application"
	queryApplicationCategory    = "query application"
	managedLabel                = "managed"
	sccLabel                    = "scc"
	ScenariosLabel              = "scenarios"
	testScenario                = "test-scenario"
	IsNormalizedLabel           = "isNormalized"
)

func TestRegisterApplicationWithAllSimpleFieldsProvided(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, testScenario)

	in := graphql.ApplicationRegisterInput{
		Name:           "wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group":                      []interface{}{"production", "experimental"},
			ScenariosLabel:               []interface{}{testScenario},
			conf.ApplicationTypeLabelKey: fixtures.CreateAppTemplateName("Cloud for Customer"),
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	example.SaveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application")

	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID, []string{testScenario})
	require.NoError(t, err)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)
	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualApp.Status.Condition)
}

func TestRegisterApplicationWithExternalCertificate(t *testing.T) {
	ctx := context.Background()

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)

	in := fixtures.FixSampleApplicationRegisterInputWithName("test", "register-app-with-external-cert")
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixtures.FixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, createRequest, &app)
	defer fixtures.CleanupApplication(t, ctx, directorCertSecuredClient, "", &app)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotEmpty(t, app.ID)
}

func TestRegisterApplicationWithOrdWebhook(t *testing.T) {
	ctx := context.Background()

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, conf.ExternalCertProviderConfig, true)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)

	in := fixtures.FixSampleApplicationRegisterInputWithORDWebhooks("test", "register-app-with-external-cert", "http://test.test", nil)
	in.LocalTenantID = nil
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixtures.FixRegisterApplicationRequest(appInputGQL)
	app := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperationWithoutTenant(ctx, directorCertSecuredClient, createRequest, &app)
	defer fixtures.CleanupApplication(t, ctx, directorCertSecuredClient, "", &app)
	require.NoError(t, err)
	require.NotNil(t, app)
	require.NotEmpty(t, app.ID)
	require.Equal(t, 1, len(app.Operations))
	require.Equal(t, graphql.ScheduledOperationTypeOrdAggregation, app.Operations[0].OperationType)
}

// TODO: Uncomment the bellow test once the authentication for last operation is in place

// func TestAsyncRegisterApplication(t *testing.T) {
// 	// GIVEN
// 	ctx := context.Background()

// 	in := graphql.ApplicationRegisterInput{
// 		Name:           "wordpress_async",
// 		ProviderName:   ptr.String("provider name"),
// 		Description:    ptr.String("my first wordpress application"),
// 		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
// 		Labels: graphql.Labels{
// 			"group":     []interface{}{"production", "experimental"},
// 			"scenarios": []interface{}{conf.DefaultScenario},
// 		},
// 	}

// 	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
// 	require.NoError(t, err)

// 	t.Log("DIRECTOR URL: ", gql.GetDirectorGraphQLURL())

// 	// WHEN
// 	request := fixtures.FixAsyncRegisterApplicationRequest(appInputGQL)
// 	var result map[string]interface{}
// 	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), request, &result)
// 	require.NoError(t, err)

// 	request = fixtures.FixGetApplicationsRequestWithPagination()
// 	actualAppPage := graphql.ApplicationPage{}
// 	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), request, &actualAppPage)
// 	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualAppPage.Data[0].ID)

// 	require.NoError(t, err)
// 	assert.Len(t, actualAppPage.Data, 1)

// 	directorURL := gql.GetDirectorURL()
// 	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/last_operation/application/%s", directorURL, actualAppPage.Data[0].ID), nil)
// 	req.Header.Set("Tenant", tenant.TestTenants.GetDefaultTenantID())
// 	require.NoError(t, err)
// 	resp, err := directorHTTPClient.Do(req)
// 	require.NoError(t, err)

// 	responseBytes, err := ioutil.ReadAll(resp.Body)
// 	require.NoError(t, err)
// 	var opResponse operation.OperationResponse
// 	err = json.Unmarshal(responseBytes, &opResponse)
// 	require.NoError(t, err)

// 	//THEN
// 	assert.Equal(t, operation.OperationTypeCreate, opResponse.OperationType)
// 	assert.Equal(t, actualAppPage.Data[0].ID, opResponse.ResourceID)
// 	assert.Equal(t, resource.Application, opResponse.ResourceType)
// 	assert.Equal(t, operation.OperationStatusSucceeded, opResponse.Status)
// }

func TestRegisterApplicationNormalizationValidation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	firstAppName := "app@wordpress"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	actualApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, firstAppName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	//THEN
	require.NotEmpty(t, actualApp.ID)
	require.Equal(t, actualApp.Name, firstAppName)

	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualApp.Status.Condition)

	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, testScenario)

	// SECOND APP WITH SAME APP NAME WHEN NORMALIZED
	inSecond := graphql.ApplicationRegisterInput{
		Name:           "app!wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group":        []interface{}{"production", "experimental"},
			ScenariosLabel: []interface{}{testScenario},
		},
	}
	appSecondInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(inSecond)
	require.NoError(t, err)
	actualSecondApp := graphql.ApplicationExt{}

	// WHEN

	request := fixtures.FixRegisterApplicationRequest(appSecondInputGQL)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualSecondApp)

	//THEN
	require.EqualError(t, err, "graphql: Object name is not unique [object=application]")
	require.Empty(t, actualSecondApp.BaseEntity)

	// THIRD APP WITH DIFFERENT APP NAME WHEN NORMALIZED
	actualThirdApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "appwordpress", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &actualThirdApp)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantId, actualThirdApp.ID, []string{testScenario})
	require.NoError(t, err)
	require.NotEmpty(t, actualThirdApp.ID)

	//THEN
	require.NotEmpty(t, actualThirdApp.ID)

	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualThirdApp.Status.Condition)

	// FOURTH APP WITH DIFFERENT ALREADY NORMALIZED NAME WHICH MATCHES EXISTING APP WHEN NORMALIZED
	inFourth := graphql.ApplicationRegisterInput{
		Name:           "mp-appwordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group":        []interface{}{"production", "experimental"},
			ScenariosLabel: []interface{}{testScenario},
		},
	}
	appFourthInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(inFourth)
	require.NoError(t, err)
	actualFourthApp := graphql.ApplicationExt{}
	// WHEN
	request = fixtures.FixRegisterApplicationRequest(appFourthInputGQL)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualFourthApp)
	//THEN
	require.EqualError(t, err, "graphql: Object name is not unique [object=application]")
	require.Empty(t, actualFourthApp.BaseEntity)

	// FIFTH APP WITH DIFFERENT ALREADY NORMALIZED NAME WHICH DOES NOT MATCH ANY EXISTING APP WHEN NORMALIZED
	fifthAppName := "mp-application"
	actualFifthApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, fifthAppName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &actualFifthApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualFifthApp.ID)

	//THEN
	require.NotEmpty(t, actualFifthApp.ID)
	require.Equal(t, actualFifthApp.Name, fifthAppName)

	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualFifthApp.Status.Condition)
}

func TestRegisterApplicationWithStatusCondition(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, testScenario)

	statusCond := graphql.ApplicationStatusConditionConnected
	in := graphql.ApplicationRegisterInput{
		Name:           "wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group":                      []interface{}{"production", "experimental"},
			ScenariosLabel:               []interface{}{testScenario},
			conf.ApplicationTypeLabelKey: fixtures.CreateAppTemplateName("Cloud for Customer"),
		},
		StatusCondition: &statusCond,
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	example.SaveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with status")

	// WHEN
	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID, []string{testScenario})

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	assertions.AssertApplication(t, in, actualApp)
	assert.Equal(t, statusCond, actualApp.Status.Condition)
}

func TestRegisterApplicationWithWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	url := "http://mywordpress.com/webhooks1"

	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, testScenario)

	in := graphql.ApplicationRegisterInput{
		Name:         "wordpress",
		ProviderName: ptr.String("compass"),
		Webhooks: []*graphql.WebhookInput{
			{
				Type: graphql.WebhookTypeConfigurationChanged,
				Auth: fixtures.FixBasicAuth(t),
				URL:  &url,
			},
		},
		Labels: graphql.Labels{
			ScenariosLabel:               []interface{}{testScenario},
			conf.ApplicationTypeLabelKey: fixtures.CreateAppTemplateName("Cloud for Customer"),
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	example.SaveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with webhooks")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID, []string{testScenario})

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)
}

func TestRegisterApplicationWithBundles(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := fixtures.FixApplicationRegisterInputWithBundles(t)
	in.Labels = graphql.Labels{
		"scenarios":                  []interface{}{testScenario},
		conf.ApplicationTypeLabelKey: fixtures.CreateAppTemplateName("Cloud for Customer"),
	}
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, testScenario)

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	example.SaveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with bundles")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID, []string{testScenario})

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)
}

// TODO: Delete after bundles are adopted

func TestRegisterApplicationWithPackagesBackwardsCompatibility(t *testing.T) {
	ctx := context.Background()
	expectedAppName := "create-app-with-packages"

	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, testScenario)

	type ApplicationWithPackagesExt struct {
		graphql.Application
		Labels                graphql.Labels                           `json:"labels"`
		Webhooks              []graphql.Webhook                        `json:"webhooks"`
		Auths                 []*graphql.SystemAuth                    `json:"auths"`
		Package               graphql.BundleExt                        `json:"package"`
		Packages              graphql.BundlePageExt                    `json:"packages"`
		EventingConfiguration graphql.ApplicationEventingConfiguration `json:"eventingConfiguration"`
	}

	t.Run("Register Application with Packages should succeed", func(t *testing.T) {
		var actualApp ApplicationWithPackagesExt
		request := fixtures.FixRegisterApplicationWithPackagesRequest(expectedAppName, conf.ApplicationTypeLabelKey, fixtures.CreateAppTemplateName("Cloud for Customer"))
		err := testctx.Tc.NewOperation(ctx).Run(request, certSecuredGraphQLClient, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &graphql.ApplicationExt{Application: actualApp.Application})
		defer fixtures.UnassignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, graphql.FormationInput{Name: testScenario}, actualApp.ID, tenant.TestTenants.GetDefaultTenantID())
		require.NoError(t, err)

		appID := actualApp.ID
		require.NotEmpty(t, appID)

		require.NotNil(t, actualApp.Packages.Data[0])
		packageID := actualApp.Packages.Data[0].ID
		require.NotEmpty(t, packageID)
		require.Equal(t, expectedAppName, actualApp.Name)

		t.Run("Get Application with Package should succeed", func(t *testing.T) {
			var actualAppWithPackage ApplicationWithPackagesExt

			request := fixtures.FixGetApplicationWithPackageRequest(appID, packageID)
			err := testctx.Tc.NewOperation(ctx).Run(request, certSecuredGraphQLClient, &actualAppWithPackage)

			require.NoError(t, err)
			require.NotEmpty(t, actualAppWithPackage.ID)
			require.NotEmpty(t, actualAppWithPackage.Package.ID)
		})

		tenantID := tenant.TestTenants.GetDefaultTenantID()
		runtimeInput := fixtures.FixRuntimeRegisterInputWithoutLabels("test-runtime")
		runtimeInput.Labels[ScenariosLabel] = []string{testScenario}
		var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtime)
		runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantID, runtimeInput, conf.GatewayOauth)

		t.Run("Get ApplicationForRuntime with Package should succeed", func(t *testing.T) {
			applicationPage := struct {
				Data []*ApplicationWithPackagesExt `json:"data"`
			}{}
			request := fixtures.FixApplicationsForRuntimeWithPackagesRequest(runtime.ID)

			rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), runtime.ID)
			require.NotNil(t, rtmAuth.Auth)
			rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
			require.True(t, ok)
			require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
			require.NotEmpty(t, rtmOauthCredentialData.ClientID)

			t.Log("Issue a Hydra token with Client Credentials")
			accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
			require.NotEmpty(t, accessToken)
			oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

			err = testctx.Tc.NewOperation(ctx).Run(request, oauthGraphQLClient, &applicationPage)

			require.NoError(t, err)
			require.Len(t, applicationPage.Data, 1)

			actualAppWithPackage := applicationPage.Data[0]

			require.NotEmpty(t, actualAppWithPackage.ID)
			require.Equal(t, actualAppWithPackage.Name, "mp-"+actualApp.Name)
			require.Equal(t, actualAppWithPackage.Description, actualApp.Description)
			require.Equal(t, actualAppWithPackage.HealthCheckURL, actualApp.HealthCheckURL)
			require.Equal(t, actualAppWithPackage.ProviderName, actualApp.ProviderName)
			require.Equal(t, len(actualAppWithPackage.Webhooks), len(actualApp.Webhooks))
			require.Equal(t, len(actualAppWithPackage.Packages.Data), len(actualApp.Packages.Data))
		})
	})
}

func TestCreateApplicationWithNonExistentIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	in := fixtures.FixSampleApplicationCreateInputWithIntegrationSystem("placeholder")
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	request := fixtures.FixRegisterApplicationRequest(appInputGQL)

	// WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)

	//THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "Object not found")
}

func TestUpdateApplication(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		actualApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "before", tenant.TestTenants.GetDefaultTenantID())
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)

		updateStatusCond := graphql.ApplicationStatusConditionConnected

		expectedApp := actualApp
		expectedApp.Name = "before"
		expectedApp.ProviderName = ptr.String("after")
		expectedApp.Description = ptr.String("after")
		expectedApp.HealthCheckURL = ptr.String(conf.WebhookUrl)
		expectedApp.BaseURL = ptr.String("after")
		expectedApp.Status.Condition = updateStatusCond
		expectedApp.Labels["name"] = "before"

		updateInput := fixtures.FixSampleApplicationUpdateInput("after")
		updateInput.BaseURL = ptr.String("after")
		updateInput.StatusCondition = &updateStatusCond
		updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
		require.NoError(t, err)
		request := fixtures.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
		updatedApp := graphql.ApplicationExt{}

		//WHEN
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &updatedApp)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, expectedApp.ID, updatedApp.ID)
		assert.Equal(t, expectedApp.Name, updatedApp.Name)
		assert.Equal(t, expectedApp.ProviderName, updatedApp.ProviderName)
		assert.Equal(t, expectedApp.Description, updatedApp.Description)
		assert.Equal(t, expectedApp.HealthCheckURL, updatedApp.HealthCheckURL)
		assert.Equal(t, expectedApp.BaseURL, updatedApp.BaseURL)
		assert.Equal(t, expectedApp.Status.Condition, updatedApp.Status.Condition)

		example.SaveExample(t, request.Query(), "update application")
	})

	t.Run("Create webhook when updating application with baseURL", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		appInput := fixtures.FixSampleApplicationRegisterInputWithAppType("before", "before", conf.ApplicationTypeLabelKey, conf.SupportedORDApplicationType)
		actualApp, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)

		updateStatusCond := graphql.ApplicationStatusConditionConnected

		require.Empty(t, actualApp.Webhooks)

		updateInput := fixtures.FixSampleApplicationUpdateInput("after")
		updateInput.BaseURL = ptr.String("https://local.com")
		updateInput.StatusCondition = &updateStatusCond
		updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
		require.NoError(t, err)
		request := fixtures.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
		updatedApp := graphql.ApplicationExt{}

		//WHEN
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &updatedApp)

		//THEN
		require.NoError(t, err)
		require.Len(t, updatedApp.Webhooks, 1)
	})

	t.Run("Does not create webhook when updating application without baseURL", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		appInput := fixtures.FixSampleApplicationRegisterInputWithAppType("before", "before", conf.ApplicationTypeLabelKey, conf.SupportedORDApplicationType)
		actualApp, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)

		require.Empty(t, actualApp.Webhooks)

		updateInput := fixtures.FixSampleApplicationUpdateInput("after")
		updateStatusCond := graphql.ApplicationStatusConditionConnected
		updateInput.StatusCondition = &updateStatusCond
		updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
		require.NoError(t, err)
		request := fixtures.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
		updatedApp := graphql.ApplicationExt{}

		//WHEN
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &updatedApp)

		//THEN
		require.NoError(t, err)
		require.Empty(t, updatedApp.Webhooks)
	})

	t.Run("Does not create webhook when updating application that has no matching applicationType from configuration", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()

		appInput := fixtures.FixSampleApplicationRegisterInputWithAppType("before", "before", "applicationType", "SAP unsupported")
		actualApp, err := fixtures.RegisterApplicationFromInput(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appInput)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)

		require.Empty(t, actualApp.Webhooks)

		updateInput := fixtures.FixSampleApplicationUpdateInput("after")
		updateInput.BaseURL = ptr.String("https://local.com")
		updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
		require.NoError(t, err)
		request := fixtures.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
		updatedApp := graphql.ApplicationExt{}

		//WHEN
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &updatedApp)

		//THEN
		require.NoError(t, err)
		require.Empty(t, updatedApp.Webhooks)
	})
}

func TestUpdateApplicationWithLocalTenantIDShouldBeAllowedOnlyForIntegrationSystems(t *testing.T) {
	ctx := context.Background()

	actualApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "before", tenant.TestTenants.GetDefaultTenantID())
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	updateStatusCond := graphql.ApplicationStatusConditionConnected
	updateInput := fixtures.FixSampleApplicationUpdateInput("after")
	updateInput.BaseURL = ptr.String("after")
	updateInput.StatusCondition = &updateStatusCond
	updateInput.LocalTenantID = ptr.String("localTenantID")
	updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)
	request := fixtures.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}

	t.Run("should fail for non-integration system", func(t *testing.T) {
		runtime, err := fixtures.RegisterRuntime(t, ctx, certSecuredGraphQLClient, "test-runtime", tenant.TestTenants.GetDefaultTenantID())
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &runtime)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		runtimeAuth := fixtures.RequestClientCredentialsForRuntime(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), runtime.ID)
		require.NotEmpty(t, runtimeAuth.ID)
		defer fixtures.DeleteSystemAuthForRuntime(t, ctx, certSecuredGraphQLClient, runtimeAuth.ID)

		runtimeOauthCredentialData, ok := runtimeAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, runtimeOauthCredentialData, token.RuntimeScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, request, &updatedApp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient scopes provided")
	})

	t.Run("should be allowed for integration systems", func(t *testing.T) {
		expectedApp := actualApp
		expectedApp.Name = "before"
		expectedApp.ProviderName = ptr.String("after")
		expectedApp.Description = ptr.String("after")
		expectedApp.HealthCheckURL = ptr.String(conf.WebhookUrl)
		expectedApp.BaseURL = ptr.String("after")
		expectedApp.Status.Condition = updateStatusCond
		expectedApp.Labels["name"] = "before"
		expectedApp.LocalTenantID = ptr.String("localTenantID")

		intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), "test-update-local-tenant")
		defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys)
		require.NoError(t, err)
		require.NotEmpty(t, intSys.ID)

		intSysAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), intSys.ID)
		require.NotEmpty(t, intSysAuth)
		defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, intSysAuth.ID)

		intSysOauthCredentialData, ok := intSysAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, intSysOauthCredentialData, token.IntegrationSystemScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, request, &updatedApp)

		require.NoError(t, err)
		assert.Equal(t, expectedApp.ID, updatedApp.ID)
		assert.Equal(t, expectedApp.Name, updatedApp.Name)
		assert.Equal(t, expectedApp.ProviderName, updatedApp.ProviderName)
		assert.Equal(t, expectedApp.Description, updatedApp.Description)
		assert.Equal(t, expectedApp.HealthCheckURL, updatedApp.HealthCheckURL)
		assert.Equal(t, expectedApp.BaseURL, updatedApp.BaseURL)
		assert.Equal(t, expectedApp.Status.Condition, updatedApp.Status.Condition)
	})
}

func TestUpdateApplicationWithNonExistentIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	actualApp, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "before", tenant.TestTenants.GetDefaultTenantID())
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	updateInput := fixtures.FixSampleApplicationUpdateInputWithIntegrationSystem("after")
	updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)
	request := fixtures.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}

	//WHEN
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &updatedApp)

	//THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "Object not found")
}

func TestDeleteApplication(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		in := fixtures.FixSampleApplicationRegisterInputWithWebhooks("app")

		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)

		createReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		actualApp := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)

		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)

		// WHEN
		delReq := fixtures.FixUnregisterApplicationRequest(actualApp.ID)
		example.SaveExample(t, delReq.Query(), "unregister application")
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, delReq, &actualApp)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Error when application is in scenario but not with runtime", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		tenantID := tenant.TestTenants.GetIDByName(t, "TestDeleteApplicationIfInScenario")

		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)
		fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)

		applicationInput := fixtures.FixSampleApplicationRegisterInputWithAppTypeAndScenarios("first", "first", conf.ApplicationTypeLabelKey, fixtures.CreateAppTemplateName("Cloud for Customer"), ScenariosLabel, []string{testScenario})
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &application)
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, application.ID, []string{testScenario})

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		//WHEN
		req := fixtures.FixUnregisterApplicationRequest(application.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, nil)

		//THEN
		require.ErrorContains(t, err, "System first is part of the following formations : test-scenario")
	})

	t.Run("Error when application is in scenario with runtime", func(t *testing.T) {
		//GIVEN
		expectedErrorMsg := "graphql: The operation is not allowed [reason=System first is part of the following formations : test-scenario]"

		ctx := context.Background()
		tenantID := tenant.TestTenants.GetIDByName(t, "TestDeleteApplicationIfInScenario")

		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)
		fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)

		runtimeInput := fixtures.FixRuntimeRegisterInputWithoutLabels("one-runtime")
		runtimeInput.Labels[ScenariosLabel] = []string{testScenario}

		var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtime)
		runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantID, runtimeInput, conf.GatewayOauth)

		applicationInput := fixtures.FixSampleApplicationRegisterInputWithAppTypeAndScenarios("first", "first", conf.ApplicationTypeLabelKey, fixtures.CreateAppTemplateName("Cloud for Customer"), ScenariosLabel, []string{testScenario})
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &application)
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, application.ID, []string{testScenario})

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		//WHEN
		req := fixtures.FixUnregisterApplicationRequest(application.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, nil)

		//THEN
		require.EqualError(t, err, expectedErrorMsg)
	})
}

func TestUnpairApplication(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		in := fixtures.FixSampleApplicationRegisterInputWithWebhooks("app")

		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)

		createReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		actualApp := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)

		require.NoError(t, err)
		require.NotEmpty(t, actualApp.ID)

		// WHEN
		unpairRequest := fixtures.FixUnpairApplicationRequest(actualApp.ID)
		example.SaveExample(t, unpairRequest.Query(), "unpair application")
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unpairRequest, &actualApp)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Success when application is in scenario but not with runtime", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		tenantID := tenant.TestTenants.GetIDByName(t, "TestDeleteApplicationIfInScenario")

		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)
		fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)

		applicationInput := fixtures.FixSampleApplicationRegisterInputWithAppTypeAndScenarios("first", "first", conf.ApplicationTypeLabelKey, fixtures.CreateAppTemplateName("Cloud for Customer"), ScenariosLabel, []string{testScenario})
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &application)
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, application.ID, []string{testScenario})
		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		//WHEN
		req := fixtures.FixUnpairApplicationRequest(application.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, nil)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Error when application is in scenario with runtime", func(t *testing.T) {
		//GIVEN
		expectedErrorMsg := "graphql: The operation is not allowed [reason=System first is still used and cannot be deleted. Unassign the system from the following formations first: test-scenario. Then, unassign the system from the following runtimes, too: one-runtime]"

		ctx := context.Background()
		tenantID := tenant.TestTenants.GetIDByName(t, "TestDeleteApplicationIfInScenario")

		runtimeInput := fixtures.FixRuntimeRegisterInputWithoutLabels("one-runtime")

		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)
		fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, testScenario)

		runtimeInput.Labels[ScenariosLabel] = []string{testScenario}

		var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtime)
		runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantID, runtimeInput, conf.GatewayOauth)

		applicationInput := fixtures.FixSampleApplicationRegisterInputWithAppTypeAndScenarios("first", "first", conf.ApplicationTypeLabelKey, fixtures.CreateAppTemplateName("Cloud for Customer"), ScenariosLabel, []string{testScenario})
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &application)
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, application.ID, []string{testScenario})

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		//WHEN
		req := fixtures.FixUnpairApplicationRequest(application.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, nil)

		//THEN
		require.EqualError(t, err, expectedErrorMsg)
	})
}

func TestUpdateApplicationParts(t *testing.T) {
	ctx := context.Background()
	placeholder := "app"
	in := fixtures.FixSampleApplicationRegisterInputWithWebhooks(placeholder)

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &actualApp)

	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	t.Run("labels manipulation", func(t *testing.T) {
		expectedLabel := graphql.Label{Key: "brand_new_label", Value: []interface{}{"aaa", "bbb"}}

		// add label
		createdLabel := &graphql.Label{}

		addReq := fixtures.FixSetApplicationLabelRequest(actualApp.ID, expectedLabel.Key, []string{"aaa", "bbb"})
		example.SaveExample(t, addReq.Query(), "set application label")

		err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addReq, &createdLabel)
		require.NoError(t, err)
		assert.Equal(t, &expectedLabel, createdLabel)

		actualApp := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, actualApp.ID)
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "aaa")
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "bbb")

		// delete label value
		deletedLabel := graphql.Label{}
		delReq := fixtures.FixDeleteApplicationLabelRequest(actualApp.ID, expectedLabel.Key)
		example.SaveExample(t, delReq.Query(), "delete application label")
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, delReq, &deletedLabel)
		require.NoError(t, err)
		assert.Equal(t, expectedLabel, deletedLabel)
		actualApp = fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, actualApp.ID)
		assert.Nil(t, actualApp.Labels[expectedLabel.Key])

	})

	t.Run("manage webhooks", func(t *testing.T) {
		// add
		outputTemplate := "{\\\"location\\\":\\\"{{.Headers.Location}}\\\",\\\"success_status_code\\\": 202,\\\"error\\\": \\\"{{.Body.error}}\\\"}"
		url := "http://new-webhook.url"
		urlUpdated := "http://updated-webhook.url"
		webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL:            &url,
			Type:           graphql.WebhookTypeUnregisterApplication,
			OutputTemplate: &outputTemplate,
		})

		require.NoError(t, err)
		addReq := fixtures.FixAddWebhookToApplicationRequest(actualApp.ID, webhookInStr)
		example.SaveExampleInCustomDir(t, addReq.Query(), example.AddWebhookCategory, "add application webhook")

		actualWebhook := graphql.Webhook{}
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addReq, &actualWebhook)
		require.NoError(t, err)

		assert.NotNil(t, actualWebhook.URL)
		assert.NotNil(t, actualWebhook.CreatedAt)
		assert.Equal(t, "http://new-webhook.url", *actualWebhook.URL)
		assert.Equal(t, graphql.WebhookTypeUnregisterApplication, actualWebhook.Type)
		id := actualWebhook.ID
		require.NotNil(t, id)

		// get all webhooks
		updatedApp := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, actualApp.ID)
		assert.Len(t, updatedApp.Webhooks, 2)

		// update
		webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL: &urlUpdated, Type: graphql.WebhookTypeUnregisterApplication, OutputTemplate: &outputTemplate})

		require.NoError(t, err)
		updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
		example.SaveExampleInCustomDir(t, updateReq.Query(), example.UpdateWebhookCategory, "update webhook")
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateReq, &actualWebhook)
		require.NoError(t, err)
		assert.NotNil(t, actualWebhook.URL)
		assert.Equal(t, urlUpdated, *actualWebhook.URL)

		// delete

		//GIVEN
		deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
		example.SaveExampleInCustomDir(t, deleteReq.Query(), example.DeleteWebhookCategory, "delete webhook")

		//WHEN
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteReq, &actualWebhook)

		//THEN
		require.NoError(t, err)
		assert.NotNil(t, actualWebhook.URL)
		assert.Equal(t, urlUpdated, *actualWebhook.URL)

	})

	t.Run("refetch API", func(t *testing.T) {
		// TODO later
	})

	t.Run("refetch Event Spec", func(t *testing.T) {
		// TODO later
	})
}

func TestQueryApplications(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		in := graphql.ApplicationRegisterInput{
			Name: fmt.Sprintf("app-%d", i),
		}

		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)

		actualApp := graphql.ApplicationExt{}
		request := fixtures.FixRegisterApplicationRequest(appInputGQL)

		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
		require.NoError(t, err)
	}
	actualAppPage := graphql.ApplicationPage{}

	// WHEN
	queryReq := fixtures.FixGetApplicationsRequestWithPagination()
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryReq, &actualAppPage)
	example.SaveExampleInCustomDir(t, queryReq.Query(), example.QueryApplicationsCategory, "query applications")

	//THEN
	require.NoError(t, err)
	assert.Len(t, actualAppPage.Data, 3)
	assert.Equal(t, 3, actualAppPage.TotalCount)
}

func TestQueryApplicationsPageable(t *testing.T) {
	// GIVEN
	appAmount := 7
	after := 3
	cursor := ""
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	apps := make(map[string]*graphql.ApplicationExt)
	for i := 0; i < appAmount; i++ {
		app, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, fmt.Sprintf("app-%d", i), tenantId)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &app)
		require.NoError(t, err)
		require.NotEmpty(t, app.ID)
		apps[app.ID] = &app
	}
	appsPage := graphql.ApplicationPageExt{}

	// WHEN
	queriesForFullPage := appAmount / after
	for i := 0; i < queriesForFullPage; i++ {
		appReq := fixtures.FixApplicationsPageableRequest(after, cursor)
		err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, appReq, &appsPage)
		require.NoError(t, err)

		//THEN
		assert.Equal(t, cursor, string(appsPage.PageInfo.StartCursor))
		assert.True(t, appsPage.PageInfo.HasNextPage)
		assert.Len(t, appsPage.Data, after)
		assert.Equal(t, appAmount, appsPage.TotalCount)
		for _, app := range appsPage.Data {
			assert.Equal(t, app, apps[app.ID])
			delete(apps, app.ID)
		}
		cursor = string(appsPage.PageInfo.EndCursor)
	}

	appReq := fixtures.FixApplicationsPageableRequest(after, cursor)
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, appReq, &appsPage)
	require.NoError(t, err)

	assert.False(t, appsPage.PageInfo.HasNextPage)
	assert.Empty(t, appsPage.PageInfo.EndCursor)
	assert.Equal(t, appAmount, appsPage.TotalCount)
	require.Len(t, appsPage.Data, 1)
	delete(apps, appsPage.Data[0].ID)
	assert.Len(t, apps, 0)
}

func TestQueryApplicationsGlobalPageable(t *testing.T) {
	// GIVEN
	pageSize := 10
	cursor := ""
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	t.Log(fmt.Sprintf("Creating application with label with key %s and value %s", conf.ApplicationTypeLabelKey, "filter-me"))
	appOne, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "app-one", conf.ApplicationTypeLabelKey, "filter-me", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &appOne)
	require.NoError(t, err)
	require.NotEmpty(t, appOne.ID)

	t.Log(fmt.Sprintf("Creating application with label with key %s and value %s", conf.ApplicationTypeLabelKey, "unknown"))
	appTwo, err := fixtures.RegisterApplicationWithApplicationType(t, ctx, certSecuredGraphQLClient, "app-two", conf.ApplicationTypeLabelKey, "unknown", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &appTwo)
	require.NoError(t, err)
	require.NotEmpty(t, appTwo.ID)

	// WHEN
	labelFilter := graphql.LabelFilter{
		Key:   conf.ApplicationTypeLabelKey,
		Query: str.Ptr(fmt.Sprintf("\"%s\"", "filter-me")),
	}

	labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	t.Log("Executing applicationsGlobal query with label filter")
	appWithTenantsPage := graphql.ApplicationExtWithTenantsPage{}
	req := fixtures.FixApplicationsGlobalFilteredPageableRequest(labelFilterGQL, pageSize, cursor)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &appWithTenantsPage)
	require.NoError(t, err)

	//THEN
	assert.Equal(t, cursor, string(appWithTenantsPage.PageInfo.StartCursor))
	assert.False(t, appWithTenantsPage.PageInfo.HasNextPage)
	assert.Equal(t, 1, appWithTenantsPage.TotalCount)
	assert.Len(t, appWithTenantsPage.Data, 1)
	assert.Equal(t, appWithTenantsPage.Data[0].Application.ID, appOne.ID)
	val, exists := appWithTenantsPage.Data[0].Application.Labels[conf.ApplicationTypeLabelKey]
	assert.True(t, exists)
	assert.Equal(t, "filter-me", val.(string))
	assert.Len(t, appWithTenantsPage.Data[0].Tenants, 1)
	assert.Equal(t, tnt.TypeToStr(tnt.Customer), appWithTenantsPage.Data[0].Tenants[0].Type)
}

func TestQuerySpecificApplication(t *testing.T) {
	// GIVEN
	in := graphql.ApplicationRegisterInput{
		Name:         "app",
		ProviderName: ptr.String("compass"),
		Labels: graphql.Labels{
			conf.ApplicationTypeLabelKey: fixtures.CreateAppTemplateName("Cloud for Customer"),
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	actualApp := graphql.ApplicationExt{}
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	err = testctx.Tc.RunOperation(context.Background(), certSecuredGraphQLClient, request, &actualApp)
	defer fixtures.CleanupApplication(t, context.Background(), certSecuredGraphQLClient, tenantId, &actualApp)

	require.NotEmpty(t, actualApp.ID)
	appID := actualApp.ID

	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)

	t.Run("Query Application With Consumer User", func(t *testing.T) {
		actualApp := graphql.Application{}

		// WHEN
		queryAppReq := fixtures.FixGetApplicationRequest(appID)
		err = testctx.Tc.RunOperation(context.Background(), certSecuredGraphQLClient, queryAppReq, &actualApp)
		example.SaveExampleInCustomDir(t, queryAppReq.Query(), queryApplicationCategory, "query application")

		//THE
		require.NoError(t, err)
		assert.Equal(t, appID, actualApp.ID)
	})

	ctx := context.Background()

	input := fixtures.FixRuntimeRegisterInputWithoutLabels("runtime-test")

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, input, conf.GatewayOauth)
	require.NoError(t, err)
	require.NotEmpty(t, runtime.ID)

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenantId, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	testScenarioSecond := "test-scenario-2"
	scenarios := []string{testScenario, testScenarioSecond}
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, testScenario)
	defer fixtures.DeleteFormation(t, ctx, certSecuredGraphQLClient, testScenarioSecond)
	fixtures.CreateFormation(t, ctx, certSecuredGraphQLClient, testScenarioSecond)

	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	t.Run("Query Application With Consumer Runtime in same scenario", func(t *testing.T) {
		// set application scenarios label

		scenariosToAssign := scenarios[1:]
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantId, appID, scenariosToAssign)
		for i := range scenariosToAssign {
			fi := graphql.FormationInput{Name: scenariosToAssign[i]}
			fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, fi, appID, tenantId)
		}

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.DeleteRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, ScenariosLabel)

		actualApp := graphql.Application{}

		// WHEN
		queryAppReq := fixtures.FixGetApplicationRequest(appID)
		err = runtimeConsumer.Run(queryAppReq, oauthGraphQLClient, &actualApp)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, appID, actualApp.ID)
	})

	t.Run("Query Application With Consumer Runtime not in same scenario", func(t *testing.T) {
		scenariosToAssign := scenarios[:1]
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantId, appID, scenariosToAssign)
		for i := range scenariosToAssign {
			fi := graphql.FormationInput{Name: scenariosToAssign[i]}
			fixtures.AssignFormationWithApplicationObjectType(t, ctx, certSecuredGraphQLClient, fi, appID, tenantId)
		}

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.DeleteRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, ScenariosLabel)

		actualApp := graphql.Application{}

		// WHEN
		queryAppReq := fixtures.FixGetApplicationRequest(appID)
		err = runtimeConsumer.Run(queryAppReq, oauthGraphQLClient, &actualApp)
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "The operation is not allowed")
	})
}

func TestApplicationsForRuntime(t *testing.T) {
	//GIVEN
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetIDByName(t, tenant.TenantSeparationTenantName)
	otherTenant := tenant.TestTenants.GetIDByName(t, tenant.ApplicationsForRuntimeTenantName)
	tenantUnnormalizedApplications := []*graphql.Application{}
	tenantNormalizedApplications := []*graphql.Application{}
	scenarios := []string{"black-friday-campaign", "christmas-campaign", "summer-campaign"}

	for _, scenario := range scenarios {
		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenario)
		fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenario)

		defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, otherTenant, scenario)
		fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, otherTenant, scenario)
	}

	//create runtime without normalization
	runtimeInputWithoutNormalization := fixtures.FixRuntimeRegisterInputWithoutLabels("unnormalized-runtime")
	runtimeInputWithoutNormalization.Labels[ScenariosLabel] = scenarios
	runtimeInputWithoutNormalization.Labels[IsNormalizedLabel] = "false"
	var runtimeWithoutNormalization graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtimeWithoutNormalization)
	runtimeWithoutNormalization = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantID, runtimeInputWithoutNormalization, conf.GatewayOauth)

	// create an oauth graphql client for requesting bundle instance auth on behalf of the runtime
	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenantID, runtimeWithoutNormalization.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)
	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	applications := []struct {
		ApplicationName string
		BundlesData     []struct {
			Name           string
			AuthsClientIDs []string
		}
		Tenant       string
		WithinTenant bool
		Scenarios    []string
	}{
		{
			ApplicationName: "second",
			BundlesData: []struct {
				Name           string
				AuthsClientIDs []string
			}{
				{Name: "bundleWithTwoAuths", AuthsClientIDs: []string{"test2", "test3"}},
				{Name: "bundleWithNoAuths", AuthsClientIDs: nil},
			},
			Tenant:       tenantID,
			WithinTenant: true,
			Scenarios:    scenarios[:1],
		},
		{
			ApplicationName: "third",
			BundlesData: []struct {
				Name           string
				AuthsClientIDs []string
			}{
				{Name: "bundleWithOneAuth", AuthsClientIDs: []string{"test1"}},
			},
			Tenant:       tenantID,
			WithinTenant: true,
			Scenarios:    scenarios,
		},
		{
			ApplicationName: "allscenarios",
			Tenant:          tenantID,
			WithinTenant:    true,
			Scenarios:       scenarios,
		},
		{
			ApplicationName: "test",
			Tenant:          otherTenant,
			WithinTenant:    false,
			Scenarios:       scenarios[:1],
		},
	}

	var expectedBundles []*graphql.Bundle
	for i, testApp := range applications {
		applicationInput := fixtures.FixSampleApplicationRegisterInput(testApp.ApplicationName)
		applicationInput.Labels = graphql.Labels{ScenariosLabel: testApp.Scenarios, conf.ApplicationTypeLabelKey: fixtures.CreateAppTemplateName("Cloud for Customer")}
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.Application{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testApp.Tenant, createApplicationReq, &application)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testApp.Tenant, &graphql.ApplicationExt{Application: application})
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, testApp.Tenant, application.ID, applications[i].Scenarios)

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		if testApp.WithinTenant {
			tenantUnnormalizedApplications = append(tenantUnnormalizedApplications, &application)

			normalizedApp := application
			normalizedApp.Name = conf.DefaultNormalizationPrefix + normalizedApp.Name
			tenantNormalizedApplications = append(tenantNormalizedApplications, &normalizedApp)
		}

		for _, data := range testApp.BundlesData {
			bundleExt := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, testApp.Tenant, application.ID, data.Name)

			var expectedBIA *graphql.BundleInstanceAuth
			for _, clientID := range data.AuthsClientIDs {
				currentBundleInstanceAuth := fixtures.CreateBundleInstanceAuthForRuntime(t, ctx, oauthGraphQLClient, tenantID, bundleExt.ID)
				currentBundleInstanceAuth = fixtures.SetBundleInstanceAuthForRuntime(t, ctx, certSecuredGraphQLClient, tenantID, currentBundleInstanceAuth.ID, clientID)
				// For a single bundle there may be more than one bundle instance auth created for specific runtime.
				// When the Kyma runtime lists the bundles it reads only the defaultInstanceAuth property of the bundle.
				// We are overriding defaultInstanceAuth field of the bundle with on of the BIAs created for the runtime.
				// When fetching the BIAs they are ordered by ID so that every time one and the same BIA is returned to
				// the Kyma runtime
				if expectedBIA == nil || strings.Compare(expectedBIA.ID, currentBundleInstanceAuth.ID) == 1 {
					expectedBIA = currentBundleInstanceAuth
				}
			}

			expectedBundle := &bundleExt.Bundle
			if expectedBIA != nil {
				expectedBundle.DefaultInstanceAuth = expectedBIA.Auth
			}
			expectedBundles = append(expectedBundles, expectedBundle)
		}
	}

	t.Run("Applications For Runtime Query without normalization", func(t *testing.T) {
		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithoutNormalization.ID)
		applicationPage := graphql.ApplicationPageExt{}

		rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenantID, runtimeWithoutNormalization.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		err := testctx.Tc.NewOperation(ctx).WithTenant(tenantID).Run(request, oauthGraphQLClient, &applicationPage)
		example.SaveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantUnnormalizedApplications))

		var actualApplications []*graphql.Application
		var actualBundles []*graphql.Bundle
		for _, applicationExt := range applicationPage.Data {
			actualApplications = append(actualApplications, &applicationExt.Application)
			for _, bundleExt := range applicationExt.Bundles.Data {
				actualBundles = append(actualBundles, &bundleExt.Bundle)
			}
		}
		assert.ElementsMatch(t, tenantUnnormalizedApplications, actualApplications)
		assert.ElementsMatch(t, expectedBundles, actualBundles)

	})

	t.Run("Applications For Runtime Query without normalization due to missing label", func(t *testing.T) {
		//create runtime without normalization
		unlabeledRuntimeInput := fixtures.FixRuntimeRegisterInputWithoutLabels("unlabeled-runtime")
		unlabeledRuntimeInput.Labels[ScenariosLabel] = scenarios
		unlabeledRuntimeInput.Labels[IsNormalizedLabel] = "false"
		var unlabeledRuntime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &unlabeledRuntime)
		unlabeledRuntime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantID, unlabeledRuntimeInput, conf.GatewayOauth)

		deleteLabelRuntimeResp := graphql.Runtime{}
		deleteLabelRequest := fixtures.FixDeleteRuntimeLabelRequest(unlabeledRuntime.ID, IsNormalizedLabel)
		err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, deleteLabelRequest, &deleteLabelRuntimeResp)
		require.NoError(t, err)

		request := fixtures.FixApplicationForRuntimeRequest(unlabeledRuntime.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, request, &applicationPage)
		example.SaveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantNormalizedApplications))
		assert.ElementsMatch(t, tenantNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications For Runtime Query with normalization", func(t *testing.T) {
		//create runtime without normalization
		runtimeInputWithNormalization := fixtures.FixRuntimeRegisterInputWithoutLabels("normalized-runtime")
		runtimeInputWithNormalization.Labels[ScenariosLabel] = scenarios
		runtimeInputWithNormalization.Labels[IsNormalizedLabel] = "true"

		var runtimeWithNormalization graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtimeWithNormalization)
		runtimeWithNormalization = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantID, runtimeInputWithNormalization, conf.GatewayOauth)

		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, request, &applicationPage)
		example.SaveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantNormalizedApplications))
		assert.ElementsMatch(t, tenantNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications Query With Consumer Runtime", func(t *testing.T) {
		request := fixtures.FixGetApplicationsRequestWithPagination()
		applicationPage := graphql.ApplicationPage{}

		rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenantID, runtimeWithoutNormalization.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		err := testctx.Tc.NewOperation(ctx).WithTenant(tenantID).Run(request, oauthGraphQLClient, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantUnnormalizedApplications))
		assert.ElementsMatch(t, tenantUnnormalizedApplications, applicationPage.Data)
	})
}

func TestApplicationsForRuntimeWithHiddenApps(t *testing.T) {
	//GIVEN
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetIDByName(t, tenant.ApplicationsForRuntimeWithHiddenAppsTenantName)
	expectedApplications := []*graphql.Application{}
	expectedNormalizedApplications := []*graphql.Application{}

	scenarios := []string{testScenario}

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenarios[0])
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantID, scenarios[0])

	applications := []struct {
		ApplicationName string
		Scenarios       []string
		Hidden          bool
	}{
		{
			ApplicationName: "second",
			Scenarios:       scenarios,
			Hidden:          false,
		},
		{
			ApplicationName: "third",
			Scenarios:       scenarios,
			Hidden:          true,
		},
	}

	applicationHideSelectorKey := "applicationHideSelectorKey"
	applicationHideSelectorValue := "applicationHideSelectorValue"

	for i, testApp := range applications {
		applicationInput := fixtures.FixSampleApplicationRegisterInput(testApp.ApplicationName)
		applicationInput.Labels = graphql.Labels{ScenariosLabel: scenarios, conf.ApplicationTypeLabelKey: fixtures.CreateAppTemplateName("Cloud for Customer")}
		if testApp.Hidden {
			(applicationInput.Labels)[applicationHideSelectorKey] = applicationHideSelectorValue
		}
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.Application{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		require.NoError(t, err)
		require.NotEmpty(t, application.ID)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &graphql.ApplicationExt{Application: application})
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, application.ID, applications[i].Scenarios)

		if !testApp.Hidden {
			expectedApplications = append(expectedApplications, &application)

			normalizedApp := application
			normalizedApp.Name = conf.DefaultNormalizationPrefix + normalizedApp.Name
			expectedNormalizedApplications = append(expectedNormalizedApplications, &normalizedApp)
		}
	}

	//create runtime without normalization
	runtimeWithoutNormalizationInput := fixtures.FixRuntimeRegisterInputWithoutLabels("unnormalized-runtime")
	runtimeWithoutNormalizationInput.Labels[ScenariosLabel] = scenarios
	runtimeWithoutNormalizationInput.Labels[IsNormalizedLabel] = "false"
	var runtimeWithoutNormalization graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtimeWithoutNormalization)
	runtimeWithoutNormalization = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantID, runtimeWithoutNormalizationInput, conf.GatewayOauth)

	t.Run("Applications For Runtime Query without normalization", func(t *testing.T) {
		//WHEN
		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithoutNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, request, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedApplications))
		assert.ElementsMatch(t, expectedApplications, applicationPage.Data)
	})

	t.Run("Applications For Runtime Query with normalization", func(t *testing.T) {
		//create runtime with normalization
		runtimeWithNormalizationInput := fixtures.FixRuntimeRegisterInputWithoutLabels("normalized-runtime")
		runtimeWithNormalizationInput.Labels[ScenariosLabel] = scenarios
		runtimeWithNormalizationInput.Labels[IsNormalizedLabel] = "true"
		var runtimeWithNormalization graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtimeWithNormalization)
		runtimeWithNormalization = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantID, runtimeWithNormalizationInput, conf.GatewayOauth)

		//WHEN
		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err := testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, request, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedNormalizedApplications))
		assert.ElementsMatch(t, expectedNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications Query With Consumer Runtime", func(t *testing.T) {
		//WHEN
		request := fixtures.FixGetApplicationsRequestWithPagination()
		applicationPage := graphql.ApplicationPage{}

		rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenantID, runtimeWithoutNormalization.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		err := testctx.Tc.NewOperation(ctx).WithTenant(tenantID).Run(request, oauthGraphQLClient, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedApplications))
		assert.ElementsMatch(t, expectedApplications, applicationPage.Data)
	})
}

func TestDeleteApplicationWithNoScenarios(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	in := graphql.ApplicationRegisterInput{
		Name:           "wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &actualApp)
	require.NoError(t, err)

	app := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, actualApp.ID)

	fixtures.DeleteApplicationLabel(t, ctx, certSecuredGraphQLClient, actualApp.ID, "integrationSystemID")
	fixtures.DeleteApplicationLabel(t, ctx, certSecuredGraphQLClient, actualApp.ID, "name")
	if _, found := app.Labels[ScenariosLabel]; found {
		fixtures.DeleteApplicationLabel(t, ctx, certSecuredGraphQLClient, actualApp.ID, ScenariosLabel)
	}
}

func TestApplicationDeletionInScenario(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	scenarios := []string{testScenario}

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, scenarios[0])
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, scenarios[0])

	in := graphql.ApplicationRegisterInput{
		Name:           "wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			ScenariosLabel:               scenarios,
			conf.ApplicationTypeLabelKey: fixtures.CreateAppTemplateName("Cloud for Customer"),
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &actualApp)
	defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantId, actualApp.ID, scenarios)
	require.NoError(t, err)

	inRuntime := fixtures.FixRuntimeRegisterInputWithoutLabels("test-runtime")
	inRuntime.Labels[ScenariosLabel] = scenarios
	inRuntime.Description = nil

	var runtime graphql.RuntimeExt // needed so the 'defer' can be above the runtime registration
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	runtime = fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, tenantId, inRuntime, conf.GatewayOauth)
	require.NoError(t, err)

	request = fixtures.FixUnregisterApplicationRequest(actualApp.ID)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, request, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), fmt.Sprintf("The operation is not allowed [reason=System wordpress is part of the following formations : %s", testScenario))

	request = fixtures.FixDeleteRuntimeLabel(runtime.ID, ScenariosLabel)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, request, nil)
	require.NoError(t, err)
}

func TestMergeApplications(t *testing.T) {
	// GIVEN

	ctx := context.Background()
	baseURL := ptr.String("http://base.com")
	healthURL := ptr.String("http://health.com")
	providerName := ptr.String("test-provider")
	description := "app description"
	nameJSONPath := "$.name-json-path"
	tenantId := tenant.TestTenants.GetDefaultTenantID()
	namePlaceholder := "name"
	displayNamePlaceholder := "display-name"
	displayNameJSONPath := "$.display-name-json-path"
	managedLabelValue := "true"
	sccLabelValue := "cloud connector"
	expectedProductType := fixtures.CreateAppTemplateName("MergeTemplate")
	newFormation := "formation-merge-applications-e2e"
	formationTemplateName := "merge-applications-template"

	appTmplInput := fixtures.FixApplicationTemplate(expectedProductType)
	appTmplInput.ApplicationInput.Name = "{{name}}"
	appTmplInput.ApplicationInput.BaseURL = baseURL
	appTmplInput.ApplicationInput.ProviderName = nil
	appTmplInput.ApplicationInput.Description = ptr.String("{{display-name}}")
	appTmplInput.ApplicationInput.HealthCheckURL = nil
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        namePlaceholder,
			Description: ptr.String("description"),
			JSONPath:    ptr.String(nameJSONPath),
		},
		{
			Name:        displayNamePlaceholder,
			Description: ptr.String(description),
			JSONPath:    ptr.String(displayNameJSONPath),
		},
	}

	t.Log("Create integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "app-template")
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

	// Create Application Template
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, oauthGraphQLClient, tenantId, appTmpl)
	require.NoError(t, err)

	appFromTmplSrc := graphql.ApplicationFromTemplateInput{
		TemplateName: expectedProductType, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: namePlaceholder,
				Value:       "app1-e2e-merge",
			},
			{
				Placeholder: displayNamePlaceholder,
				Value:       description,
			},
		},
	}

	appFromTmplDest := graphql.ApplicationFromTemplateInput{
		TemplateName: expectedProductType, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: namePlaceholder,
				Value:       "app2-e2e-merge",
			},
			{
				Placeholder: displayNamePlaceholder,
				Value:       description,
			},
		},
	}

	t.Logf("Should create source application")
	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	require.NoError(t, err)
	createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	outputSrcApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, createAppFromTmplFirstRequest, &outputSrcApp)
	defer fixtures.CleanupApplication(t, ctx, oauthGraphQLClient, tenantId, &outputSrcApp)
	require.NoError(t, err)

	t.Logf("Should create destination application")
	appFromTmplDestGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplDest)
	require.NoError(t, err)
	createAppFromTmplSecondRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplDestGQL)
	outputDestApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, createAppFromTmplSecondRequest, &outputDestApp)
	defer fixtures.CleanupApplication(t, ctx, oauthGraphQLClient, tenantId, &outputDestApp)
	require.NoError(t, err)

	t.Logf("Should update source application with more data")
	updateInput := fixtures.FixSampleApplicationUpdateInput("after")
	updateInput.ProviderName = providerName
	updateInput.HealthCheckURL = healthURL
	updateInput.Description = ptr.String(description)
	updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)

	updateRequest := fixtures.FixUpdateApplicationRequest(outputSrcApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, updateRequest, &updatedApp)
	require.NoError(t, err)

	fixtures.SetApplicationLabelWithTenant(t, ctx, oauthGraphQLClient, tenantId, outputSrcApp.ID, managedLabel, managedLabelValue)
	fixtures.SetApplicationLabelWithTenant(t, ctx, oauthGraphQLClient, tenantId, outputSrcApp.ID, sccLabel, sccLabelValue)

	t.Logf("Should create formation template: %s", formationTemplateName)
	formationTemplateRegisterInput := fixtures.FixFormationTemplateRegisterInputWithTypes(formationTemplateName, []string{conf.KymaRuntimeTypeLabelValue}, []string{expectedProductType})
	actualFormationTemplate := graphql.FormationTemplate{} // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &actualFormationTemplate)
	actualFormationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateRegisterInput)
	t.Logf("Should create formation: %s", newFormation)
	var formation graphql.Formation
	formationInput := fixtures.FixFormationInput(newFormation, str.Ptr(formationTemplateName))
	formationInputGQL, err := testctx.Tc.Graphqlizer.FormationInputToGQL(formationInput)
	require.NoError(t, err)

	createReq := fixtures.FixCreateFormationWithTemplateRequest(formationInputGQL)
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, createReq, &formation)
	defer fixtures.DeleteFormation(t, ctx, oauthGraphQLClient, newFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, formation.Name)

	t.Logf("Assign application %s to formation %s", outputDestApp.Name, newFormation)
	assignReq := fixtures.FixAssignFormationRequest(outputDestApp.ID, string(graphql.FormationObjectTypeApplication), newFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, assignReq, &assignFormation)
	defer func() {
		t.Logf("Unassigning %s from formation %s", outputDestApp.Name, newFormation)
		request := fixtures.FixUnassignFormationRequest(outputDestApp.ID, string(graphql.FormationObjectTypeApplication), newFormation)
		var response graphql.Formation
		err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, request, &response)
		if nil == err {
			t.Logf("%s was unassigned from formation %s", outputDestApp.Name, newFormation)
		} else {
			t.Logf("%s was not removed from formation %s: %v", outputDestApp.Name, newFormation, err)
		}
	}()
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	// WHEN
	t.Logf("Should be able to merge application %s into %s", outputSrcApp.Name, outputDestApp.Name)
	destApp := graphql.ApplicationExt{}
	mergeRequest := fixtures.FixMergeApplicationsRequest(outputSrcApp.ID, outputDestApp.ID)
	example.SaveExample(t, mergeRequest.Query(), "merge applications")
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, mergeRequest, &destApp)

	// THEN
	require.NoError(t, err)

	assert.Equal(t, description, str.PtrStrToStr(destApp.Description))
	assert.Equal(t, healthURL, destApp.HealthCheckURL)
	assert.Equal(t, providerName, destApp.ProviderName)
	assert.Equal(t, managedLabelValue, destApp.Labels[managedLabel])
	assert.Equal(t, sccLabelValue, destApp.Labels[sccLabel])
	assert.Contains(t, destApp.Labels[ScenariosLabel], newFormation)

	srcApp := graphql.ApplicationExt{}
	getSrcAppReq := fixtures.FixGetApplicationRequest(outputSrcApp.ID)
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, getSrcAppReq, &srcApp)
	require.NoError(t, err)

	// Source application is deleted
	assert.Empty(t, srcApp.BaseEntity)
}

func TestMergeApplicationsWithSelfRegDistinguishLabelKey(t *testing.T) {
	// GIVEN

	ctx := context.Background()
	baseURL := ptr.String("http://base.com")
	healthURL := ptr.String("http://health.com")
	providerName := ptr.String("test-provider")
	description := "app description"
	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()
	namePlaceholder := "name"
	displayNamePlaceholder := "display-name"
	managedLabelValue := "true"
	sccLabelValue := "cloud connector"
	expectedProductType := fixtures.CreateAppTemplateName("MergeTemplate")
	newFormation := "formation-merge-applications-e2e"
	formationTemplateName := "merge-applications-template"
	nameJSONPath := "$.name-json-path"
	displayNameJSONPath := "$.display-name-json-path"

	// Make clients that use a certificate which has a OU value for subaccount the same as the OU in the certSecuredGraphQLClient
	// in order to maintain the tenant isolation
	appTechnicalProviderDirectorCertSecuredClient := directorCertSecuredClientWithExternalCertSubaccount(t, ctx, "app-template-merge-technical-cn")

	appTmplInput := fixtures.FixAppTemplateInputWithDefaultDistinguishLabel(expectedProductType, conf.SubscriptionConfig.SelfRegDistinguishLabelKey, conf.SubscriptionConfig.SelfRegDistinguishLabelValue)
	appTmplInput.ApplicationInput.Name = "{{name}}"
	appTmplInput.ApplicationInput.BaseURL = baseURL
	appTmplInput.ApplicationInput.ProviderName = nil
	appTmplInput.ApplicationInput.Description = ptr.String("{{display-name}}")
	appTmplInput.ApplicationInput.HealthCheckURL = nil
	appTmplInput.ApplicationInput.Labels = map[string]interface{}{"displayName": "{{display-name}}"}
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        namePlaceholder,
			Description: ptr.String("description"),
			JSONPath:    ptr.String(nameJSONPath),
		},
		{
			Name:        displayNamePlaceholder,
			Description: ptr.String(description),
			JSONPath:    ptr.String(displayNameJSONPath),
		},
	}
	// Create Application Template
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, appTechnicalProviderDirectorCertSecuredClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, appTechnicalProviderDirectorCertSecuredClient, appTmpl)
	require.NoError(t, err)
	require.Equal(t, conf.SubscriptionConfig.SelfRegRegion, appTmpl.Labels[tenantfetcher.RegionKey])

	appFromTmplSrc := graphql.ApplicationFromTemplateInput{
		TemplateName: expectedProductType, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: namePlaceholder,
				Value:       "app1-e2e-merge",
			},
			{
				Placeholder: displayNamePlaceholder,
				Value:       description,
			},
		},
	}

	appFromTmplDest := graphql.ApplicationFromTemplateInput{
		TemplateName: expectedProductType, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: namePlaceholder,
				Value:       "app2-e2e-merge",
			},
			{
				Placeholder: displayNamePlaceholder,
				Value:       description,
			},
		},
	}

	t.Logf("Should create source application")
	appFromTmplSrcGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplSrc)
	require.NoError(t, err)
	createAppFromTmplFirstRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplSrcGQL)
	outputSrcApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, createAppFromTmplFirstRequest, &outputSrcApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &outputSrcApp)
	require.NoError(t, err)

	t.Logf("Should create destination application")
	appFromTmplDestGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmplDest)
	require.NoError(t, err)
	createAppFromTmplSecondRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplDestGQL)
	outputDestApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, createAppFromTmplSecondRequest, &outputDestApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &outputDestApp)
	require.NoError(t, err)

	t.Logf("Should update source application with more data")
	updateInput := fixtures.FixSampleApplicationUpdateInput("after")
	updateInput.ProviderName = providerName
	updateInput.HealthCheckURL = healthURL
	updateInput.Description = ptr.String(description)
	updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)

	updateRequest := fixtures.FixUpdateApplicationRequest(outputSrcApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateRequest, &updatedApp)
	require.NoError(t, err)

	fixtures.SetApplicationLabelWithTenant(t, ctx, certSecuredGraphQLClient, tenantId, outputSrcApp.ID, managedLabel, managedLabelValue)
	fixtures.SetApplicationLabelWithTenant(t, ctx, certSecuredGraphQLClient, tenantId, outputSrcApp.ID, sccLabel, sccLabelValue)

	formationTemplateRegisterInput := fixtures.FixFormationTemplateRegisterInputWithTypes(formationTemplateName, []string{conf.KymaRuntimeTypeLabelValue}, []string{expectedProductType})
	actualFormationTemplate := graphql.FormationTemplate{} // needed so the 'defer' can be above the formation template creation
	defer fixtures.CleanupFormationTemplate(t, ctx, certSecuredGraphQLClient, &actualFormationTemplate)
	actualFormationTemplate = fixtures.CreateFormationTemplate(t, ctx, certSecuredGraphQLClient, formationTemplateRegisterInput)

	t.Logf("Should create formation: %s", newFormation)
	var formation graphql.Formation
	formationInput := fixtures.FixFormationInput(newFormation, str.Ptr(formationTemplateName))
	formationInputGQL, err := testctx.Tc.Graphqlizer.FormationInputToGQL(formationInput)
	require.NoError(t, err)

	createReq := fixtures.FixCreateFormationWithTemplateRequest(formationInputGQL)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	defer func() {
		t.Logf("Cleaning up formation: %s", newFormation)
		var response graphql.Formation
		deleteFormationReq := fixtures.FixDeleteFormationRequest(newFormation)
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteFormationReq, &response)
		require.NoError(t, err)
		require.Equal(t, newFormation, response.Name)
		t.Logf("Deleted formation with name: %s", response.Name)
	}()
	require.NoError(t, err)
	require.Equal(t, newFormation, formation.Name)
	t.Logf("Assign application to formation %s", newFormation)
	assignReq := fixtures.FixAssignFormationRequest(outputDestApp.ID, string(graphql.FormationObjectTypeApplication), newFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	defer func() {
		t.Logf("Unassigning dest-app from formation %s", newFormation)
		request := fixtures.FixUnassignFormationRequest(outputDestApp.ID, string(graphql.FormationObjectTypeApplication), newFormation)
		var response graphql.Formation
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &response)
		if nil == err {
			t.Logf("Dest-app was unassigned from formation %s", newFormation)
		} else {
			t.Logf("Dest-app was not removed from formation %s: %v", newFormation, err)
		}
	}()
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	// WHEN
	t.Logf("Should not be able to merge application %s into %s", outputSrcApp.Name, outputDestApp.Name)
	destApp := graphql.ApplicationExt{}
	mergeRequest := fixtures.FixMergeApplicationsRequest(outputSrcApp.ID, outputDestApp.ID)
	example.SaveExample(t, mergeRequest.Query(), "merge applications with self register distinguish label key")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, mergeRequest, &destApp)

	// THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), fmt.Sprintf("app template: %s has label %s", *outputSrcApp.ApplicationTemplateID, conf.SubscriptionConfig.SelfRegDistinguishLabelKey))

	srcApp := graphql.ApplicationExt{}
	getSrcAppReq := fixtures.FixGetApplicationRequest(outputSrcApp.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getSrcAppReq, &srcApp)
	require.NoError(t, err)

	// Source application is not deleted
	t.Logf("Source application should not be deleted")
	assert.NotEmpty(t, srcApp.BaseEntity)
}

func TestGetApplicationsAPIEventDefinitions(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()
	appName := "app-test-get-api-event"

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	api := fixtures.AddAPIToApplication(t, ctx, certSecuredGraphQLClient, application.ID)
	event := fixtures.AddEventToApplication(t, ctx, certSecuredGraphQLClient, application.ID)
	require.NotEmpty(t, api.ID)
	require.NotEmpty(t, event.ID)

	queryAPIForApplication := fixtures.FixGetApplicationWithAPIEventDefinitionRequest(application.ID, api.ID, event.ID)

	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryAPIForApplication, &app)
	require.NoError(t, err)

	require.NotEmpty(t, app.APIDefinition)
	require.NotEmpty(t, app.EventDefinition)
	assert.Equal(t, app.APIDefinition.ID, api.ID)
	assert.Equal(t, app.EventDefinition.ID, event.ID)
}

func TestListApplicationsByLocalTenantID(t *testing.T) {
	//GIVEN
	ctx := context.TODO()
	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()
	localTenantID := "local-tenant-id-1234"

	nameField := "name"
	displayNameField := "displayName"
	localTenantIDField := "localTenantId"

	createIntegrationSystem := func() *graphql.IntegrationSystemExt {
		integrationSystem, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "integration-system-name")
		require.NoError(t, err)
		require.NotEmpty(t, integrationSystem.ID)
		return integrationSystem
	}

	createIntegrationSystemCredentials := func(integrationSystem *graphql.IntegrationSystemExt) *graphql.IntSysSystemAuth {
		integrationSystemAuth := fixtures.RequestClientCredentialsForIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, integrationSystem.ID)
		require.NotEmpty(t, integrationSystemAuth)
		return integrationSystemAuth
	}

	createIntegrationSystemOauthClient := func(integrationSystemAuth *graphql.IntSysSystemAuth) *gcli.Client {
		integrationSystemOauthCredentialData, ok := integrationSystemAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		return gql.NewAuthorizedGraphQLClientWithCustomURL(token.GetAccessToken(t, integrationSystemOauthCredentialData, token.IntegrationSystemScopes), conf.GatewayOauth)
	}

	createAppTemplate := func(client *gcli.Client, appTemplateName string) graphql.ApplicationTemplate {
		input := fixtures.FixApplicationTemplate(appTemplateName)
		input.Placeholders = []*graphql.PlaceholderDefinitionInput{
			{
				Name:     "name",
				JSONPath: ptr.String(fmt.Sprintf("$.%s", nameField)),
			},
			{
				Name:     "display-name",
				JSONPath: ptr.String(fmt.Sprintf("$.%s", displayNameField)),
			},
			{
				Name:     "tenant-id",
				JSONPath: ptr.String(fmt.Sprintf("$.%s", localTenantIDField)),
			},
		}
		input.ApplicationInput.LocalTenantID = ptr.String("{{tenant-id}}")

		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, client, tenantId, input)
		require.NoError(t, err)
		return appTemplate
	}

	createApp := func(client *gcli.Client, appTemplate graphql.ApplicationTemplate, name, localTenantID string) graphql.ApplicationExt {
		input := graphql.ApplicationFromTemplateInput{
			ID:           &appTemplate.ID,
			TemplateName: appTemplate.Name,
			PlaceholdersPayload: str.Ptr(fmt.Sprintf(`{\"%s\": \"%s\", \"%s\": \"%s\", \"%s\": \"%s\"}`,
				nameField, name,
				displayNameField, name,
				localTenantIDField, localTenantID)),
		}
		inputGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(input)
		require.NoError(t, err)

		request := fixtures.FixRegisterApplicationFromTemplateWithLocalTenantID(inputGQL)
		app := graphql.ApplicationExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, client, tenantId, request, &app)
		require.NoError(t, err)
		return app
	}

	listApplications := func(localTenantID, appTemplateName string) (graphql.ApplicationPageExt, error) {
		var filter string
		var err error
		if appTemplateName != "" {
			filter, err = testctx.Tc.Graphqlizer.LabelFilterToGQL(graphql.LabelFilter{Key: conf.ApplicationTypeLabelKey, Query: str.Ptr(fmt.Sprintf("\"%s\"", appTemplateName))})
			require.NoError(t, err)
		}

		request := fixtures.FixListApplicationsByLocalTenantID(localTenantID, filter, 200, "")
		page := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &page)
		if appTemplateName != "" {
			example.SaveExampleInCustomDir(t, request.Query(), example.QueryApplicationsCategory, "query applications by local tenant id and an optional filter")
		}

		return page, err
	}

	integrationSystem := createIntegrationSystem()
	defer fixtures.CleanupIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, integrationSystem)

	integrationSystemAuth := createIntegrationSystemCredentials(integrationSystem)
	defer fixtures.DeleteSystemAuthForIntegrationSystem(t, ctx, certSecuredGraphQLClient, integrationSystemAuth.ID)

	integrationSystemOAuthClient := createIntegrationSystemOauthClient(integrationSystemAuth)

	appTemplateOneName := fixtures.CreateAppTemplateName("template one")
	appTemplateOne := createAppTemplate(integrationSystemOAuthClient, appTemplateOneName)
	defer fixtures.CleanupApplicationTemplate(t, ctx, integrationSystemOAuthClient, tenantId, appTemplateOne)

	appTemplateTwoName := fixtures.CreateAppTemplateName("template two")
	appTemplateTwo := createAppTemplate(integrationSystemOAuthClient, appTemplateTwoName)
	defer fixtures.CleanupApplicationTemplate(t, ctx, integrationSystemOAuthClient, tenantId, appTemplateTwo)

	appOne := createApp(integrationSystemOAuthClient, appTemplateOne, "app-one-name", localTenantID)
	defer fixtures.UnregisterApplication(t, ctx, integrationSystemOAuthClient, tenantId, appOne.ID)

	appTwo := createApp(integrationSystemOAuthClient, appTemplateTwo, "app-two-name", localTenantID)
	defer fixtures.UnregisterApplication(t, ctx, integrationSystemOAuthClient, tenantId, appTwo.ID)

	//WHEN
	page, err := listApplications(localTenantID, "")

	//THEN
	require.NoError(t, err)
	require.Equal(t, 2, page.TotalCount)
	require.Equal(t, 2, len(page.Data))

	localTenantID1 := page.Data[0].LocalTenantID
	require.NotNil(t, localTenantID1)
	require.Equal(t, localTenantID, *localTenantID1)

	localTenantID2 := page.Data[1].LocalTenantID
	require.NotNil(t, localTenantID2)
	require.Equal(t, localTenantID, *localTenantID2)

	actualName1 := page.Data[0].Name
	actualName2 := page.Data[1].Name
	require.NotEqual(t, actualName1, actualName2)

	appNames := []string{actualName1, actualName2}
	require.Subset(t, appNames, []string{appOne.Name})
	require.Subset(t, appNames, []string{appTwo.Name})

	//WHEN
	page, err = listApplications(localTenantID, appTemplateOne.Name)

	//THEN
	require.NoError(t, err)
	require.Equal(t, 1, page.TotalCount)
	require.Equal(t, 1, len(page.Data))

	localTenantID3 := page.Data[0].LocalTenantID
	require.NotNil(t, localTenantID3)
	require.Equal(t, localTenantID, *localTenantID3)
	require.Equal(t, appOne.Name, page.Data[0].Name)
}

func TestGetApplicationByLocalTenantIDAndAppTemplateID(t *testing.T) {
	//GIVEN
	ctx := context.TODO()

	nameJSONPath := "$.name"
	displayNameJSONPath := "$.displayName"
	tenantIDJSONPath := "$.localTenantId"

	appName := "appName"
	localTenantID := "local-tenant-id-1234"
	appTemplateID := "app-template-id-1234"
	placeholdersPayload := fmt.Sprintf(`{\"name\": \"%s\", \"displayName\":\"appDisplayName\", \"localTenantId\":\"%s\"}`, appName, localTenantID)

	appTemplateName := fixtures.CreateAppTemplateName("template")
	appTmplInput := fixtures.FixApplicationTemplate(appTemplateName)
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
		{
			Name:        "tenant-id",
			Description: ptr.String("tenant-id"),
			JSONPath:    &tenantIDJSONPath,
		},
	}
	appTmplInput.ApplicationInput.LocalTenantID = ptr.String("{{tenant-id}}")

	tenantId := tenant.TestTenants.GetDefaultSubaccountTenantID()

	t.Log("Creating integration system")
	intSys, err := fixtures.RegisterIntegrationSystem(t, ctx, certSecuredGraphQLClient, tenantId, "update-app-template-with-override")
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

	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, oauthGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplateWithoutTenant(t, ctx, oauthGraphQLClient, appTmpl)
	require.NoError(t, err)

	appFromTmpl := graphql.ApplicationFromTemplateInput{ID: &appTemplateID, TemplateName: appTemplateName, PlaceholdersPayload: &placeholdersPayload}
	appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
	require.NoError(t, err)
	createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplateWithLocalTenantID(appFromTmplGQL)

	outputApp := graphql.ApplicationExt{}
	//WHEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, oauthGraphQLClient, tenantId, createAppFromTmplRequest, &outputApp)
	defer fixtures.UnregisterApplication(t, ctx, oauthGraphQLClient, tenantId, outputApp.ID)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, outputApp)
	require.Equal(t, appName, outputApp.Application.Name)
	require.Equal(t, appTmpl.ID, *outputApp.Application.ApplicationTemplateID)
	require.Equal(t, localTenantID, *outputApp.Application.LocalTenantID)

	getAppRequest := fixtures.FixGetApplicationByLocalTenantIDAndAppTemplateIDRequest(localTenantID, *outputApp.ApplicationTemplateID)
	newApp := graphql.ApplicationExt{}
	//WHEN
	err = testctx.Tc.RunOperation(ctx, oauthGraphQLClient, getAppRequest, &newApp)
	example.SaveExampleInCustomDir(t, getAppRequest.Query(), queryApplicationCategory, "query application by local tenant id and app template id")

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, newApp)
	require.Equal(t, appName, newApp.Application.Name)
	require.Equal(t, outputApp.Application.ID, newApp.Application.ID)
	require.Equal(t, appTmpl.ID, *newApp.Application.ApplicationTemplateID)
	require.Equal(t, localTenantID, *newApp.Application.LocalTenantID)
}

func directorCertSecuredClientWithExternalCertSubaccount(t *testing.T, ctx context.Context, cn string) *gcli.Client {
	replacer := strings.NewReplacer(conf.TestProviderSubaccountID, conf.ExternalCertTestIntSystemOUSubaccount, conf.TestExternalCertCN, cn)
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertTestJobName,
		TestExternalCertSubject:               replacer.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject),
		ExternalClientCertCertKey:             conf.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.CertificateService,
	}

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig, true)
	return gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)
}
