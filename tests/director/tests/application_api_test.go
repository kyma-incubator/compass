package tests

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/json"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	eventingCategory            = "eventing"
	registerApplicationCategory = "register application"
	queryApplicationsCategory   = "query applications"
	queryApplicationCategory    = "query application"
	deleteWebhookCategory       = "delete webhook"
	addWebhookCategory          = "add webhook"
	updateWebhookCategory       = "update webhook"
	managedLabel                = "managed"
	sccLabel                    = "scc"
)

func TestRegisterApplicationWithAllSimpleFieldsProvided(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	in := graphql.ApplicationRegisterInput{
		Name:           "wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application")

	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)
	require.NoError(t, err)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)
	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualApp.Status.Condition)
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
// 			"scenarios": []interface{}{"DEFAULT"},
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

	// SECOND APP WITH SAME APP NAME WHEN NORMALIZED
	inSecond := graphql.ApplicationRegisterInput{
		Name:           "app!wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
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
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
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
	statusCond := graphql.ApplicationStatusConditionConnected
	in := graphql.ApplicationRegisterInput{
		Name:           "wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
		},
		StatusCondition: &statusCond,
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with status")

	// WHEN
	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)

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
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with webhooks")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)
}

func TestRegisterApplicationWithBundles(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := fixtures.FixApplicationRegisterInputWithBundles(t)
	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with bundles")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)
}

// TODO: Delete after bundles are adopted

func TestRegisterApplicationWithPackagesBackwardsCompatibility(t *testing.T) {
	ctx := context.Background()
	expectedAppName := "create-app-with-packages"

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
		request := fixtures.FixRegisterApplicationWithPackagesRequest(expectedAppName)
		err := testctx.Tc.NewOperation(ctx).Run(request, certSecuredGraphQLClient, &actualApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &graphql.ApplicationExt{Application: actualApp.Application})
		appID := actualApp.ID
		packageID := actualApp.Packages.Data[0].ID

		require.NoError(t, err)
		require.NotEmpty(t, appID)

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

		runtimeInput := fixtures.FixRuntimeRegisterInput("test-runtime")
		(runtimeInput.Labels)[ScenariosLabel] = []string{"DEFAULT"}
		runtimeInputGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeInput)

		require.NoError(t, err)
		registerRuntimeRequest := fixtures.FixRegisterRuntimeRequest(runtimeInputGQL)

		runtime := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, registerRuntimeRequest, &runtime)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &runtime)

		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		t.Run("Get ApplicationForRuntime with Package should succeed", func(t *testing.T) {
			applicationPage := struct {
				Data []*ApplicationWithPackagesExt `json:"data"`
			}{}
			request := fixtures.FixApplicationsForRuntimeWithPackagesRequest(runtime.ID)

			rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), runtime.ID)
			rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
			require.True(t, ok)
			require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
			require.NotEmpty(t, rtmOauthCredentialData.ClientID)

			t.Log("Issue a Hydra token with Client Credentials")
			accessToken := token.GetAccessToken(t, rtmOauthCredentialData, token.RuntimeScopes)
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

	saveExample(t, request.Query(), "update application")
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
		saveExample(t, delReq.Query(), "unregister application")
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, delReq, &actualApp)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Success when application is in scenario but not with runtime", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		tenantID := tenant.TestTenants.GetIDByName(t, "TestDeleteApplicationIfInScenario")

		defaultValue := "DEFAULT"
		scenarios := []string{defaultValue, "test-scenario"}

		jsonSchema := map[string]interface{}{
			"type":        "array",
			"minItems":    1,
			"uniqueItems": true,
			"items": map[string]interface{}{
				"type": "string",
				"enum": scenarios,
			},
		}
		var schema interface{} = jsonSchema

		fixtures.CreateLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, ScenariosLabel, schema, tenantID)

		applicationInput := fixtures.FixSampleApplicationRegisterInput("first")
		applicationInput.Labels = graphql.Labels{ScenariosLabel: scenarios}
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), &application)

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		//WHEN
		req := fixtures.FixUnregisterApplicationRequest(application.ID)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, req, nil)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Error when application is in scenario with runtime", func(t *testing.T) {
		//GIVEN
		expectedErrorMsg := "graphql: The operation is not allowed [reason=System first is still used and cannot be deleted. Unassign the system from the following formations first: test-scenario. Then, unassign the system from the following runtimes, too: one-runtime]"

		ctx := context.Background()
		tenantID := tenant.TestTenants.GetIDByName(t, "TestDeleteApplicationIfInScenario")

		runtimeInput := fixtures.FixRuntimeRegisterInput("one-runtime")
		defaultValue := "DEFAULT"
		scenarios := []string{defaultValue, "test-scenario"}
		(runtimeInput.Labels)[ScenariosLabel] = scenarios
		runtimeInputWithNormalizationGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeInput)
		require.NoError(t, err)
		registerRuntimeRequest := fixtures.FixRegisterRuntimeRequest(runtimeInputWithNormalizationGQL)

		runtime := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, registerRuntimeRequest, &runtime)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtime)

		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		applicationInput := fixtures.FixSampleApplicationRegisterInput("first")
		applicationInput.Labels = graphql.Labels{ScenariosLabel: scenarios}
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &application)

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)
		defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, application.ID, conf.DefaultScenarioEnabled)

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
		saveExample(t, unpairRequest.Query(), "unpair application")
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, unpairRequest, &actualApp)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Success when application is in scenario but not with runtime", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		tenantID := tenant.TestTenants.GetIDByName(t, "TestDeleteApplicationIfInScenario")

		defaultValue := "DEFAULT"
		scenarios := []string{defaultValue, "test-scenario"}

		applicationInput := fixtures.FixSampleApplicationRegisterInput("first")
		applicationInput.Labels = graphql.Labels{ScenariosLabel: scenarios}
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		defer func() {
			defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &application)
			defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, application.ID, conf.DefaultScenarioEnabled)
		}()

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

		runtimeInput := fixtures.FixRuntimeRegisterInput("one-runtime")
		defaultValue := "DEFAULT"
		scenarios := []string{defaultValue, "test-scenario"}
		(runtimeInput.Labels)[ScenariosLabel] = scenarios
		runtimeInputWithNormalizationGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeInput)
		require.NoError(t, err)
		registerRuntimeRequest := fixtures.FixRegisterRuntimeRequest(runtimeInputWithNormalizationGQL)

		runtime := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, registerRuntimeRequest, &runtime)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtime)

		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		applicationInput := fixtures.FixSampleApplicationRegisterInput("first")
		applicationInput.Labels = graphql.Labels{ScenariosLabel: scenarios}
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.ApplicationExt{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &application)
		defer func() {
			defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &application)
			defer fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, application.ID, conf.DefaultScenarioEnabled)
		}()

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
		saveExample(t, addReq.Query(), "set application label")

		err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addReq, &createdLabel)
		require.NoError(t, err)
		assert.Equal(t, &expectedLabel, createdLabel)

		actualApp := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, actualApp.ID)
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "aaa")
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "bbb")

		// delete label value
		deletedLabel := graphql.Label{}
		delReq := fixtures.FixDeleteApplicationLabelRequest(actualApp.ID, expectedLabel.Key)
		saveExample(t, delReq.Query(), "delete application label")
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
		saveExampleInCustomDir(t, addReq.Query(), addWebhookCategory, "add application webhook")

		actualWebhook := graphql.Webhook{}
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addReq, &actualWebhook)
		require.NoError(t, err)

		assert.NotNil(t, actualWebhook.URL)
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
		saveExampleInCustomDir(t, updateReq.Query(), updateWebhookCategory, "update webhook")
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateReq, &actualWebhook)
		require.NoError(t, err)
		assert.NotNil(t, actualWebhook.URL)
		assert.Equal(t, urlUpdated, *actualWebhook.URL)

		// delete

		//GIVEN
		deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
		saveExampleInCustomDir(t, deleteReq.Query(), deleteWebhookCategory, "delete webhook")

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
	saveExampleInCustomDir(t, queryReq.Query(), queryApplicationsCategory, "query applications")

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

func TestQuerySpecificApplication(t *testing.T) {
	// GIVEN
	in := graphql.ApplicationRegisterInput{
		Name:         "app",
		ProviderName: ptr.String("compass"),
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
		saveExampleInCustomDir(t, queryAppReq.Query(), queryApplicationCategory, "query application")

		//THE
		require.NoError(t, err)
		assert.Equal(t, appID, actualApp.ID)
	})

	ctx := context.Background()

	input := fixtures.FixRuntimeRegisterInput("runtime-test")

	runtime, err := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, &input)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
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

	scenarios := []string{"test-scenario", "test-scenario-2"}
	defaultScenarios := []string{conf.DefaultScenario}
	// update label definitions
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, append([]string{conf.DefaultScenario}, scenarios...))
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, tenantId, defaultScenarios)

	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	t.Run("Query Application With Consumer Runtime in same scenario", func(t *testing.T) {
		// set application scenarios label
		fixtures.SetApplicationLabel(t, ctx, certSecuredGraphQLClient, appID, ScenariosLabel, scenarios[1:])
		defer fixtures.DeleteApplicationLabel(t, ctx, certSecuredGraphQLClient, appID, "scenarios")

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer func() {
			deleteLabelRequest := fixtures.FixDeleteRuntimeLabel(runtime.ID, "scenarios")
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteLabelRequest, nil)
			require.NoError(t, err)
		}()

		actualApp := graphql.Application{}

		// WHEN
		queryAppReq := fixtures.FixGetApplicationRequest(appID)
		err = runtimeConsumer.Run(queryAppReq, oauthGraphQLClient, &actualApp)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, appID, actualApp.ID)
	})

	t.Run("Query Application With Consumer Runtime not in same scenario", func(t *testing.T) {
		// set application scenarios label
		fixtures.SetApplicationLabel(t, ctx, certSecuredGraphQLClient, appID, ScenariosLabel, scenarios[:1])
		defer fixtures.DeleteApplicationLabel(t, ctx, certSecuredGraphQLClient, appID, "scenarios")

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, certSecuredGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer func() {
			deleteLabelRequest := fixtures.FixDeleteRuntimeLabel(runtime.ID, "scenarios")
			err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, deleteLabelRequest, nil)
			require.NoError(t, err)
		}()

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
	scenarios := []string{conf.DefaultScenario, "black-friday-campaign", "christmas-campaign", "summer-campaign"}

	jsonSchema := map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type": "string",
			"enum": scenarios,
		},
	}
	var schema interface{} = jsonSchema

	fixtures.CreateLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, ScenariosLabel, schema, tenantID)
	fixtures.CreateLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, ScenariosLabel, schema, otherTenant)

	applications := []struct {
		ApplicationName string
		Tenant          string
		WithinTenant    bool
		Scenarios       []string
	}{
		{
			Tenant:          tenantID,
			ApplicationName: "second",
			WithinTenant:    true,
			Scenarios:       []string{"black-friday-campaign"},
		},
		{
			Tenant:          tenantID,
			ApplicationName: "third",
			WithinTenant:    true,
			Scenarios:       []string{"black-friday-campaign", "christmas-campaign", "summer-campaign"},
		},
		{
			Tenant:          tenantID,
			ApplicationName: "allscenarios",
			WithinTenant:    true,
			Scenarios:       []string{"black-friday-campaign", "christmas-campaign", "summer-campaign"},
		},
		{
			Tenant:          otherTenant,
			ApplicationName: "test",
			WithinTenant:    false,
			Scenarios:       []string{"black-friday-campaign"},
		},
	}

	if conf.DefaultScenarioEnabled {
		applications = append(applications, struct {
			ApplicationName string
			Tenant          string
			WithinTenant    bool
			Scenarios       []string
		}{
			Tenant:          tenantID,
			ApplicationName: "first",
			WithinTenant:    true,
			Scenarios:       []string{conf.DefaultScenario},
		})
	}

	for _, testApp := range applications {
		applicationInput := fixtures.FixSampleApplicationRegisterInputWithWebhooks(testApp.ApplicationName)
		applicationInput.Labels = graphql.Labels{ScenariosLabel: testApp.Scenarios}
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.Application{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testApp.Tenant, createApplicationReq, &application)
		defer func(applicationID, tenant string) {
			fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenant, applicationID, conf.DefaultScenarioEnabled)
			fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenant, &graphql.ApplicationExt{Application: application})
		}(application.ID, testApp.Tenant)

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		if testApp.WithinTenant {
			tenantUnnormalizedApplications = append(tenantUnnormalizedApplications, &application)

			normalizedApp := application
			normalizedApp.Name = conf.DefaultNormalizationPrefix + normalizedApp.Name
			tenantNormalizedApplications = append(tenantNormalizedApplications, &normalizedApp)
		}
	}

	//create runtime without normalization
	runtimeInputWithoutNormalization := fixtures.FixRuntimeRegisterInput("unnormalized-runtime")
	(runtimeInputWithoutNormalization.Labels)[ScenariosLabel] = scenarios
	(runtimeInputWithoutNormalization.Labels)[IsNormalizedLabel] = "false"
	runtimeInputWithoutNormalizationGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeInputWithoutNormalization)
	require.NoError(t, err)
	registerRuntimeWithNormalizationRequest := fixtures.FixRegisterRuntimeRequest(runtimeInputWithoutNormalizationGQL)

	runtimeWithoutNormalization := graphql.RuntimeExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, registerRuntimeWithNormalizationRequest, &runtimeWithoutNormalization)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtimeWithoutNormalization)

	require.NoError(t, err)
	require.NotEmpty(t, runtimeWithoutNormalization.ID)

	t.Run("Applications For Runtime Query without normalization", func(t *testing.T) {
		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithoutNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, request, &applicationPage)
		saveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantUnnormalizedApplications))
		assert.ElementsMatch(t, tenantUnnormalizedApplications, applicationPage.Data)

	})

	t.Run("Applications For Runtime Query without normalization due to missing label", func(t *testing.T) {
		//create runtime without normalization
		unlabeledRuntimeInput := fixtures.FixRuntimeRegisterInput("unlabeled-runtime")
		(unlabeledRuntimeInput.Labels)[ScenariosLabel] = scenarios
		(unlabeledRuntimeInput.Labels)[IsNormalizedLabel] = "false"
		unlabeledRuntimeGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(unlabeledRuntimeInput)
		require.NoError(t, err)
		registerUnlabeledRuntimeRequest := fixtures.FixRegisterRuntimeRequest(unlabeledRuntimeGQL)

		unlabledRuntime := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, registerUnlabeledRuntimeRequest, &unlabledRuntime)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &unlabledRuntime)

		require.NoError(t, err)
		require.NotEmpty(t, unlabledRuntime.ID)

		deleteLabelRuntimeResp := graphql.Runtime{}
		deleteLabelRequest := fixtures.FixDeleteRuntimeLabelRequest(unlabledRuntime.ID, IsNormalizedLabel)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, deleteLabelRequest, &deleteLabelRuntimeResp)
		require.NoError(t, err)

		request := fixtures.FixApplicationForRuntimeRequest(unlabledRuntime.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, request, &applicationPage)
		saveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantNormalizedApplications))
		assert.ElementsMatch(t, tenantNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications For Runtime Query with normalization", func(t *testing.T) {
		//create runtime without normalization
		runtimeInputWithNormalization := fixtures.FixRuntimeRegisterInput("normalized-runtime")
		(runtimeInputWithNormalization.Labels)[ScenariosLabel] = scenarios
		(runtimeInputWithNormalization.Labels)[IsNormalizedLabel] = "true"
		runtimeInputWithNormalizationGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeInputWithNormalization)
		require.NoError(t, err)
		registerRuntimeWithNormalizationRequest := fixtures.FixRegisterRuntimeRequest(runtimeInputWithNormalizationGQL)

		runtimeWithNormalization := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, registerRuntimeWithNormalizationRequest, &runtimeWithNormalization)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtimeWithNormalization)

		require.NoError(t, err)
		require.NotEmpty(t, runtimeWithNormalization.ID)

		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, request, &applicationPage)
		saveExample(t, request.Query(), "query applications for runtime")

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

		err = testctx.Tc.NewOperation(ctx).WithTenant(tenantID).Run(request, oauthGraphQLClient, &applicationPage)

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

	defaultValue := conf.DefaultScenario
	scenarios := []string{defaultValue, "test-scenario"}

	jsonSchema := map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type": "string",
			"enum": scenarios,
		},
	}
	var schema interface{} = jsonSchema

	fixtures.CreateLabelDefinitionWithinTenant(t, ctx, certSecuredGraphQLClient, ScenariosLabel, schema, tenantID)

	applications := []struct {
		ApplicationName string
		Scenarios       []string
		Hidden          bool
	}{
		{
			ApplicationName: "second",
			Scenarios:       []string{"test-scenario"},
			Hidden:          false,
		},
		{
			ApplicationName: "third",
			Scenarios:       []string{"test-scenario"},
			Hidden:          true,
		},
	}

	if conf.DefaultScenarioEnabled {
		applications = append(applications, struct {
			ApplicationName string
			Scenarios       []string
			Hidden          bool
		}{
			ApplicationName: "first",
			Scenarios:       []string{defaultValue},
			Hidden:          false,
		})
	}

	applicationHideSelectorKey := "applicationHideSelectorKey"
	applicationHideSelectorValue := "applicationHideSelectorValue"

	for _, testApp := range applications {
		applicationInput := fixtures.FixSampleApplicationRegisterInputWithWebhooks(testApp.ApplicationName)
		applicationInput.Labels = graphql.Labels{ScenariosLabel: testApp.Scenarios}
		if testApp.Hidden {
			(applicationInput.Labels)[applicationHideSelectorKey] = applicationHideSelectorValue
		}
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.Application{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, createApplicationReq, &application)
		defer func(applicationID string) {
			fixtures.UnassignApplicationFromScenarios(t, ctx, certSecuredGraphQLClient, tenantID, applicationID, conf.DefaultScenarioEnabled)
			fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantID, &graphql.ApplicationExt{Application: application})
		}(application.ID)

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		if !testApp.Hidden {
			expectedApplications = append(expectedApplications, &application)

			normalizedApp := application
			normalizedApp.Name = conf.DefaultNormalizationPrefix + normalizedApp.Name
			expectedNormalizedApplications = append(expectedNormalizedApplications, &normalizedApp)
		}
	}

	//create runtime without normalization
	runtimeWithoutNormalizationInput := fixtures.FixRuntimeRegisterInput("unnormalized-runtime")
	(runtimeWithoutNormalizationInput.Labels)[ScenariosLabel] = scenarios
	(runtimeWithoutNormalizationInput.Labels)[IsNormalizedLabel] = "false"
	runtimeWithoutNormalizationInputGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeWithoutNormalizationInput)
	require.NoError(t, err)

	registerWithoutNormalizationRuntimeRequest := fixtures.FixRegisterRuntimeRequest(runtimeWithoutNormalizationInputGQL)
	runtimeWithoutNormalization := graphql.RuntimeExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, registerWithoutNormalizationRuntimeRequest, &runtimeWithoutNormalization)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtimeWithoutNormalization)

	require.NoError(t, err)
	require.NotEmpty(t, runtimeWithoutNormalization.ID)

	t.Run("Applications For Runtime Query without normalization", func(t *testing.T) {
		//WHEN
		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithoutNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, request, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedApplications))
		assert.ElementsMatch(t, expectedApplications, applicationPage.Data)
	})

	t.Run("Applications For Runtime Query with normalization", func(t *testing.T) {
		//create runtime with normalization
		runtimeWithNormalizationInput := fixtures.FixRuntimeRegisterInput("normalized-runtime")
		(runtimeWithNormalizationInput.Labels)[ScenariosLabel] = scenarios
		(runtimeWithNormalizationInput.Labels)[IsNormalizedLabel] = "true"
		runtimeWithNormalizationInputGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(runtimeWithNormalizationInput)
		require.NoError(t, err)

		registerWithNormalizationRuntimeRequest := fixtures.FixRegisterRuntimeRequest(runtimeWithNormalizationInputGQL)
		runtimeWithNormalization := graphql.RuntimeExt{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, registerWithNormalizationRuntimeRequest, &runtimeWithNormalization)
		defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantID, &runtimeWithNormalization)

		require.NoError(t, err)
		require.NotEmpty(t, runtimeWithNormalization.ID)

		//WHEN
		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantID, request, &applicationPage)

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

		err = testctx.Tc.NewOperation(ctx).WithTenant(tenantID).Run(request, oauthGraphQLClient, &applicationPage)

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
	if _, found := app.Labels["scenarios"]; found {
		fixtures.DeleteApplicationLabel(t, ctx, certSecuredGraphQLClient, actualApp.ID, "scenarios")
	}
}

func TestApplicationDeletionInScenario(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	scenarios := []string{conf.DefaultScenario, "test"}

	validSchema := map[string]interface{}{
		"type":        "array",
		"minItems":    1,
		"uniqueItems": true,
		"items": map[string]interface{}{
			"type": "string",
			"enum": scenarios,
		},
	}
	labelDefinitionInput := graphql.LabelDefinitionInput{
		Key:    "scenarios",
		Schema: json.MarshalJSONSchema(t, validSchema),
	}

	ldInputGql, err := testctx.Tc.Graphqlizer.LabelDefinitionInputToGQL(labelDefinitionInput)
	require.NoError(t, err)

	updateLabelDefinitionReq := fixtures.FixUpdateLabelDefinitionRequest(ldInputGql)

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, updateLabelDefinitionReq, nil)
	require.NoError(t, err)

	in := graphql.ApplicationRegisterInput{
		Name:           "wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: graphql.Labels{
			"scenarios": scenarios,
		},
	}

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, request, &actualApp)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &actualApp)
	require.NoError(t, err)

	inRuntime := graphql.RuntimeRegisterInput{
		Name: "test-runtime",
		Labels: graphql.Labels{
			"scenarios": scenarios,
		},
	}
	runtimeInputGQL, err := testctx.Tc.Graphqlizer.RuntimeRegisterInputToGQL(inRuntime)
	require.NoError(t, err)
	request = fixtures.FixRegisterRuntimeRequest(runtimeInputGQL)
	runtime := graphql.RuntimeExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, request, &runtime)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, tenantId, &runtime)
	require.NoError(t, err)

	request = fixtures.FixUnregisterApplicationRequest(actualApp.ID)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, request, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "The operation is not allowed [reason=System wordpress is still used and cannot be deleted. Unassign the system from the following formations first: test. Then, unassign the system from the following runtimes, too: test-runtime")

	request = fixtures.FixDeleteRuntimeLabel(runtime.ID, "scenarios")
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, request, nil)
	require.NoError(t, err)

	request = fixtures.FixUnregisterApplicationRequest(actualApp.ID)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, tenantId, request, nil)
	require.NoError(t, err)
}

func TestMergeApplications(t *testing.T) {
	// GIVEN

	ctx := context.Background()
	baseURL := ptr.String("http://base.com")
	healthURL := ptr.String("http://health.com")
	providerName := ptr.String("test-provider")
	description := ptr.String("app description")
	tenantId := tenant.TestTenants.GetDefaultTenantID()
	namePlaceholder := "name"
	managedLabelValue := "true"
	sccLabelValue := "cloud connector"
	expectedProductType := "MergeTemplate"
	newFormation := "formation-merge-applications-e2e"

	appTmplInput := fixtures.FixApplicationTemplate(expectedProductType)
	appTmplInput.ApplicationInput.Name = "{{name}}"
	appTmplInput.ApplicationInput.BaseURL = baseURL
	appTmplInput.ApplicationInput.ProviderName = nil
	appTmplInput.ApplicationInput.Description = nil
	appTmplInput.ApplicationInput.HealthCheckURL = nil
	appTmplInput.Placeholders = []*graphql.PlaceholderDefinitionInput{
		{
			Name:        namePlaceholder,
			Description: ptr.String("description"),
		},
	}

	// Create Application Template
	appTmpl, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, certSecuredGraphQLClient, tenantId, appTmplInput)
	defer fixtures.CleanupApplicationTemplate(t, ctx, certSecuredGraphQLClient, tenantId, &appTmpl)
	require.NoError(t, err)

	appFromTmplSrc := graphql.ApplicationFromTemplateInput{
		TemplateName: expectedProductType, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: namePlaceholder,
				Value:       "app1-e2e-merge",
			},
		},
	}

	appFromTmplDest := graphql.ApplicationFromTemplateInput{
		TemplateName: expectedProductType, Values: []*graphql.TemplateValueInput{
			{
				Placeholder: namePlaceholder,
				Value:       "app2-e2e-merge",
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
	updateInput.Description = description
	updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)

	updateRequest := fixtures.FixUpdateApplicationRequest(outputSrcApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateRequest, &updatedApp)
	require.NoError(t, err)

	fixtures.SetApplicationLabelWithTenant(t, ctx, certSecuredGraphQLClient, tenantId, outputSrcApp.ID, managedLabel, managedLabelValue)
	fixtures.SetApplicationLabelWithTenant(t, ctx, certSecuredGraphQLClient, tenantId, outputSrcApp.ID, sccLabel, sccLabelValue)

	t.Logf("Should create formation: %s", newFormation)
	var formation graphql.Formation
	createReq := fixtures.FixCreateFormationRequest(newFormation)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, createReq, &formation)
	require.NoError(t, err)
	require.Equal(t, newFormation, formation.Name)

	defer func() {
		t.Logf("Cleaning up formation: %s", newFormation)

		var response graphql.Formation
		deleteFormationReq := fixtures.FixDeleteFormationRequest(newFormation)
		err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, deleteFormationReq, &response)

		t.Logf("Deleted formation: %v", response)
	}()

	t.Logf("Assign application to formation %s", newFormation)
	assignReq := fixtures.FixAssignFormationRequest(outputSrcApp.ID, "APPLICATION", newFormation)
	var assignFormation graphql.Formation
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, assignReq, &assignFormation)
	require.NoError(t, err)
	require.Equal(t, newFormation, assignFormation.Name)

	// WHEN
	t.Logf("Should be able to merge application %s into %s", outputSrcApp.Name, outputDestApp.Name)
	destApp := graphql.ApplicationExt{}
	mergeRequest := fixtures.FixMergeApplicationsRequest(outputSrcApp.ID, outputDestApp.ID)
	saveExample(t, mergeRequest.Query(), "merge applications")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, mergeRequest, &destApp)

	// THEN
	require.NoError(t, err)

	assert.Equal(t, description, destApp.Description)
	assert.Equal(t, healthURL, destApp.HealthCheckURL)
	assert.Equal(t, providerName, destApp.ProviderName)
	assert.Equal(t, managedLabelValue, destApp.Labels[managedLabel])
	assert.Equal(t, sccLabelValue, destApp.Labels[sccLabel])
	assert.Contains(t, destApp.Labels[ScenariosLabel], newFormation)

	srcApp := graphql.ApplicationExt{}
	getSrcAppReq := fixtures.FixGetApplicationRequest(outputSrcApp.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, getSrcAppReq, &srcApp)
	require.NoError(t, err)

	// Source application is deleted
	assert.Empty(t, srcApp.BaseEntity)
}
