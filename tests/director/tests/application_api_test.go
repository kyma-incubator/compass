package tests

import (
	"context"
	"fmt"
	"testing"

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
	eventingCategory            = "eventing"
	registerApplicationCategory = "register application"
	queryApplicationsCategory   = "query applications"
	queryApplicationCategory    = "query application"
	deleteWebhookCategory       = "delete webhook"
	addWebhookCategory          = "add webhook"
	updateWebhookCategory       = "update webhook"
)

var integrationSystemID = "69230297-3c81-4711-aac2-3afa8cb42e2d"

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

	t.Log("DIRECTOR URL: ", gql.GetDirectorGraphQLURL())

	// WHEN
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), request, &actualApp)
	t.Log("SCOPES: ", testctx.Tc.CurrentScopes)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)
	assertions.AssertApplication(t, in, actualApp)
	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualApp.Status.Condition)
}

func TestRegisterApplicationNormalizationValidation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	firstAppName := "app@wordpress"

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	actualApp := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, firstAppName, tenantId)

	//THEN
	require.NotEmpty(t, actualApp.ID)
	require.Equal(t, actualApp.Name, firstAppName)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, actualApp.ID)

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
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualSecondApp)

	//THEN
	require.EqualError(t, err, "graphql: Object name is not unique [object=application]")
	require.Empty(t, actualSecondApp.BaseEntity)

	// THIRD APP WITH DIFFERENT APP NAME WHEN NORMALIZED
	actualThirdApp := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "appwordpress", tenantId)

	//THEN
	require.NotEmpty(t, actualThirdApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, actualThirdApp.ID)

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
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualFourthApp)
	//THEN
	require.EqualError(t, err, "graphql: Object name is not unique [object=application]")
	require.Empty(t, actualFourthApp.BaseEntity)

	// FIFTH APP WITH DIFFERENT ALREADY NORMALIZED NAME WHICH DOES NOT MATCH ANY EXISTING APP WHEN NORMALIZED
	fifthAppName := "mp-application"
	actualFifthApp := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, fifthAppName, tenantId)
	//THEN
	require.NotEmpty(t, actualFifthApp.ID)
	require.Equal(t, actualFifthApp.Name, fifthAppName)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantId, actualFifthApp.ID)

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

	// WHEN
	actualApp := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with status")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)
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
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)
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
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with bundles")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)
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
		err := testctx.Tc.NewOperation(ctx).Run(request, dexGraphQLClient, &actualApp)

		appID := actualApp.ID
		packageID := actualApp.Packages.Data[0].ID

		require.NoError(t, err)
		require.NotEmpty(t, appID)

		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appID)

		require.NotEmpty(t, packageID)
		require.Equal(t, expectedAppName, actualApp.Name)

		t.Run("Get Application with Package should succeed", func(t *testing.T) {
			var actualAppWithPackage ApplicationWithPackagesExt

			request := fixtures.FixGetApplicationWithPackageRequest(appID, packageID)
			err := testctx.Tc.NewOperation(ctx).Run(request, dexGraphQLClient, &actualAppWithPackage)

			require.NoError(t, err)
			require.NotEmpty(t, actualAppWithPackage.ID)
			require.NotEmpty(t, actualAppWithPackage.Package.ID)
		})

		runtimeInput := fixtures.FixRuntimeInput("test-runtime")
		(runtimeInput.Labels)[ScenariosLabel] = []string{"DEFAULT"}
		runtimeInputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInput)

		require.NoError(t, err)
		registerRuntimeRequest := fixtures.FixRegisterRuntimeRequest(runtimeInputGQL)

		runtime := graphql.Runtime{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, registerRuntimeRequest, &runtime)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), runtime.ID)

		t.Run("Get ApplicationForRuntime with Package should succeed", func(t *testing.T) {
			applicationPage := struct {
				Data []*ApplicationWithPackagesExt `json:"data"`
			}{}
			request := fixtures.FixApplicationsForRuntimeWithPackagesRequest(runtime.ID)

			rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), runtime.ID)
			rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
			require.True(t, ok)
			require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
			require.NotEmpty(t, rtmOauthCredentialData.ClientID)

			t.Log("Issue a Hydra token with Client Credentials")
			accessToken := token.GetAccessToken(t, rtmOauthCredentialData, "runtime:read application:read")
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
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)

	//THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "Object not found")
}

func TestUpdateApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	actualApp := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "before", tenant.TestTenants.GetDefaultTenantID())
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)

	updateStatusCond := graphql.ApplicationStatusConditionConnected

	expectedApp := actualApp
	expectedApp.Name = "before"
	expectedApp.ProviderName = ptr.String("after")
	expectedApp.Description = ptr.String("after")
	expectedApp.HealthCheckURL = ptr.String(conf.WebhookUrl)
	expectedApp.Status.Condition = updateStatusCond
	expectedApp.Labels["name"] = "before"

	updateInput := fixtures.FixSampleApplicationUpdateInput("after")
	updateInput.StatusCondition = &updateStatusCond
	updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)
	request := fixtures.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &updatedApp)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, expectedApp.ID, updatedApp.ID)
	assert.Equal(t, expectedApp.Name, updatedApp.Name)
	assert.Equal(t, expectedApp.ProviderName, updatedApp.ProviderName)
	assert.Equal(t, expectedApp.Description, updatedApp.Description)
	assert.Equal(t, expectedApp.HealthCheckURL, updatedApp.HealthCheckURL)
	assert.Equal(t, expectedApp.Status.Condition, updatedApp.Status.Condition)

	saveExample(t, request.Query(), "update application")
}

func TestUpdateApplicationWithNonExistentIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	actualApp := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, "before", tenant.TestTenants.GetDefaultTenantID())
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)

	updateInput := fixtures.FixSampleApplicationUpdateInputWithIntegrationSystem("after")
	updateInputGQL, err := testctx.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)
	request := fixtures.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}

	//WHEN
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &updatedApp)

	//THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "Object not found")
}

func TestCreateApplicationWithDuplicatedNamesWithinTenant(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appName := "samename"

	actualApp := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant.TestTenants.GetDefaultTenantID())
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)

	t.Run("Error when creating second Application with same name", func(t *testing.T) {
		in := fixtures.FixSampleApplicationRegisterInputWithNameAndWebhooks("first", appName)
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)
		request := fixtures.FixRegisterApplicationRequest(appInputGQL)

		// WHEN
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not unique")
	})
}

func TestDeleteApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := fixtures.FixSampleApplicationRegisterInputWithWebhooks("app")

	appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &actualApp)
	require.NoError(t, err)

	require.NotEmpty(t, actualApp.ID)

	// WHEN
	delReq := fixtures.FixUnregisterApplicationRequest(actualApp.ID)
	saveExample(t, delReq.Query(), "unregister application")
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, delReq, &actualApp)

	//THEN
	require.NoError(t, err)
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

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)

	t.Run("labels manipulation", func(t *testing.T) {
		expectedLabel := graphql.Label{Key: "brand_new_label", Value: []interface{}{"aaa", "bbb"}}

		// add label
		createdLabel := &graphql.Label{}

		addReq := fixtures.FixSetApplicationLabelRequest(actualApp.ID, expectedLabel.Key, []string{"aaa", "bbb"})
		saveExample(t, addReq.Query(), "set application label")

		err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, addReq, &createdLabel)
		require.NoError(t, err)
		assert.Equal(t, &expectedLabel, createdLabel)

		actualApp := fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, actualApp.ID)
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "aaa")
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "bbb")

		// delete label value
		deletedLabel := graphql.Label{}
		delReq := fixtures.FixDeleteApplicationLabelRequest(actualApp.ID, expectedLabel.Key)
		saveExample(t, delReq.Query(), "delete application label")
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, delReq, &deletedLabel)
		require.NoError(t, err)
		assert.Equal(t, expectedLabel, deletedLabel)
		actualApp = fixtures.GetApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)
		assert.Nil(t, actualApp.Labels[expectedLabel.Key])

	})

	t.Run("manage webhooks", func(t *testing.T) {
		// add

		url := "http://new-webhook.url"
		urlUpdated := "http://updated-webhook.url"
		webhookInStr, err := testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL:  &url,
			Type: graphql.WebhookTypeUnregisterApplication,
		})

		require.NoError(t, err)
		addReq := fixtures.FixAddWebhookRequest(actualApp.ID, webhookInStr)
		saveExampleInCustomDir(t, addReq.Query(), addWebhookCategory, "add application webhook")

		actualWebhook := graphql.Webhook{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, addReq, &actualWebhook)
		require.NoError(t, err)

		assert.NotNil(t, actualWebhook.URL)
		assert.Equal(t, "http://new-webhook.url", *actualWebhook.URL)
		assert.Equal(t, graphql.WebhookTypeUnregisterApplication, actualWebhook.Type)
		id := actualWebhook.ID
		require.NotNil(t, id)

		// get all webhooks
		updatedApp := fixtures.GetApplication(t, ctx, dexGraphQLClient, tenantId, actualApp.ID)
		assert.Len(t, updatedApp.Webhooks, 2)

		// update
		webhookInStr, err = testctx.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL: &urlUpdated, Type: graphql.WebhookTypeUnregisterApplication})

		require.NoError(t, err)
		updateReq := fixtures.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
		saveExampleInCustomDir(t, updateReq.Query(), updateWebhookCategory, "update application webhook")
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, updateReq, &actualWebhook)
		require.NoError(t, err)
		assert.NotNil(t, actualWebhook.URL)
		assert.Equal(t, urlUpdated, *actualWebhook.URL)

		// delete

		//GIVEN
		deleteReq := fixtures.FixDeleteWebhookRequest(actualWebhook.ID)
		saveExampleInCustomDir(t, deleteReq.Query(), deleteWebhookCategory, "delete application webhook")

		//WHEN
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, deleteReq, &actualWebhook)

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

		actualApp := graphql.Application{}
		request := fixtures.FixRegisterApplicationRequest(appInputGQL)

		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)
		require.NoError(t, err)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)
	}
	actualAppPage := graphql.ApplicationPage{}

	// WHEN
	queryReq := fixtures.FixGetApplicationsRequestWithPagination()
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, queryReq, &actualAppPage)
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
		app := fixtures.RegisterApplication(t, ctx, dexGraphQLClient, fmt.Sprintf("app-%d", i), tenantId)
		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), app.ID)
		apps[app.ID] = &app
	}
	appsPage := graphql.ApplicationPageExt{}

	// WHEN
	queriesForFullPage := appAmount / after
	for i := 0; i < queriesForFullPage; i++ {
		appReq := fixtures.FixApplicationsPageableRequest(after, cursor)
		err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, appReq, &appsPage)
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
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, appReq, &appsPage)
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

	actualApp := graphql.Application{}
	request := fixtures.FixRegisterApplicationRequest(appInputGQL)
	err = testctx.Tc.RunOperation(context.Background(), dexGraphQLClient, request, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	appID := actualApp.ID
	defer fixtures.UnregisterApplication(t, context.Background(), dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), appID)

	t.Run("Query Application With Consumer User", func(t *testing.T) {
		actualApp := graphql.Application{}

		// WHEN
		queryAppReq := fixtures.FixGetApplicationRequest(appID)
		err = testctx.Tc.RunOperation(context.Background(), dexGraphQLClient, queryAppReq, &actualApp)
		saveExampleInCustomDir(t, queryAppReq.Query(), queryApplicationCategory, "query application")

		//THE
		require.NoError(t, err)
		assert.Equal(t, appID, actualApp.ID)
	})

	ctx := context.Background()

	input := fixtures.FixRuntimeInput("runtime-test")

	runtime := fixtures.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenantId, &input)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantId, runtime.ID)

	rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantId, runtime.ID)
	rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
	require.True(t, ok)
	require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
	require.NotEmpty(t, rtmOauthCredentialData.ClientID)

	t.Log("Issue a Hydra token with Client Credentials")
	accessToken := token.GetAccessToken(t, rtmOauthCredentialData, "runtime:read application:read")
	oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

	scenarios := []string{conf.DefaultScenario, "test-scenario"}

	// update label definitions
	fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), scenarios)
	defer fixtures.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), scenarios[:1])

	runtimeConsumer := testctx.Tc.NewOperation(ctx)

	t.Run("Query Application With Consumer Runtime in same scenario", func(t *testing.T) {
		// set application scenarios label
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, appID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, appID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[:1])

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
		fixtures.SetApplicationLabel(t, ctx, dexGraphQLClient, appID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[1:])
		defer fixtures.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenantId, runtime.ID, ScenariosLabel, scenarios[:1])

		actualApp := graphql.Application{}

		// WHEN
		queryAppReq := fixtures.FixGetApplicationRequest(appID)
		err = runtimeConsumer.Run(queryAppReq, oauthGraphQLClient, &actualApp)
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "The operation is not allowed")
	})
}

func TestTenantSeparation(t *testing.T) {
	// GIVEN
	appIn := fixtures.FixSampleApplicationRegisterInputWithWebhooks("tenantseparation")
	inStr, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(appIn)
	require.NoError(t, err)

	createReq := fixtures.FixRegisterApplicationRequest(inStr)
	actualApp := graphql.ApplicationExt{}
	ctx := context.Background()

	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenant.TestTenants.GetDefaultTenantID(), actualApp.ID)

	// WHEN
	getAppReq := fixtures.FixGetApplicationsRequestWithPagination()
	customTenant := tenant.TestTenants.GetIDByName(t, tenant.TenantSeparationTenantName)
	anotherTenantsApps := graphql.ApplicationPage{}
	// THEN
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, customTenant, getAppReq, &anotherTenantsApps)
	require.NoError(t, err)
	assert.Empty(t, anotherTenantsApps.Data)
}

func TestApplicationsForRuntime(t *testing.T) {
	//GIVEN
	ctx := context.Background()
	tenantID := tenant.TestTenants.GetIDByName(t, tenant.TenantSeparationTenantName)
	otherTenant := tenant.TestTenants.GetIDByName(t, tenant.ApplicationsForRuntimeTenantName)
	tenantUnnormalizedApplications := []*graphql.Application{}
	tenantNormalizedApplications := []*graphql.Application{}
	defaultValue := "DEFAULT"
	scenarios := []string{defaultValue, "black-friday-campaign", "christmas-campaign", "summer-campaign"}

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

	fixtures.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, ScenariosLabel, schema, tenantID)
	fixtures.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, ScenariosLabel, schema, otherTenant)

	applications := []struct {
		ApplicationName string
		Tenant          string
		WithinTenant    bool
		Scenarios       []string
	}{
		{
			Tenant:          tenantID,
			ApplicationName: "first",
			WithinTenant:    true,
			Scenarios:       []string{defaultValue},
		},
		{
			Tenant:          tenantID,
			ApplicationName: "second",
			WithinTenant:    true,
			Scenarios:       []string{defaultValue, "black-friday-campaign"},
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
			Scenarios:       []string{defaultValue, "black-friday-campaign", "christmas-campaign", "summer-campaign"},
		},
		{
			Tenant:          otherTenant,
			ApplicationName: "test",
			WithinTenant:    false,
			Scenarios:       []string{defaultValue, "black-friday-campaign"},
		},
	}

	for _, testApp := range applications {
		applicationInput := fixtures.FixSampleApplicationRegisterInputWithWebhooks(testApp.ApplicationName)
		applicationInput.Labels = graphql.Labels{ScenariosLabel: testApp.Scenarios}
		appInputGQL, err := testctx.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := fixtures.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.Application{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, testApp.Tenant, createApplicationReq, &application)

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, testApp.Tenant, application.ID)
		if testApp.WithinTenant {
			tenantUnnormalizedApplications = append(tenantUnnormalizedApplications, &application)

			normalizedApp := application
			normalizedApp.Name = conf.DefaultNormalizationPrefix + normalizedApp.Name
			tenantNormalizedApplications = append(tenantNormalizedApplications, &normalizedApp)
		}
	}

	//create runtime without normalization
	runtimeInputWithoutNormalization := fixtures.FixRuntimeInput("unnormalized-runtime")
	(runtimeInputWithoutNormalization.Labels)[ScenariosLabel] = scenarios
	(runtimeInputWithoutNormalization.Labels)[IsNormalizedLabel] = "false"
	runtimeInputWithoutNormalizationGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInputWithoutNormalization)
	require.NoError(t, err)
	registerRuntimeWithNormalizationRequest := fixtures.FixRegisterRuntimeRequest(runtimeInputWithoutNormalizationGQL)

	runtimeWithoutNormalization := graphql.Runtime{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerRuntimeWithNormalizationRequest, &runtimeWithoutNormalization)
	require.NoError(t, err)
	require.NotEmpty(t, runtimeWithoutNormalization.ID)
	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantID, runtimeWithoutNormalization.ID)

	t.Run("Applications For Runtime Query without normalization", func(t *testing.T) {
		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithoutNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)
		saveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantUnnormalizedApplications))
		assert.ElementsMatch(t, tenantUnnormalizedApplications, applicationPage.Data)

	})

	t.Run("Applications For Runtime Query without normalization due to missing label", func(t *testing.T) {
		//create runtime without normalization
		unlabeledRuntimeInput := fixtures.FixRuntimeInput("unlabeled-runtime")
		(unlabeledRuntimeInput.Labels)[ScenariosLabel] = scenarios
		(unlabeledRuntimeInput.Labels)[IsNormalizedLabel] = "false"
		unlabeledRuntimeGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(unlabeledRuntimeInput)
		require.NoError(t, err)
		registerUnlabeledRuntimeRequest := fixtures.FixRegisterRuntimeRequest(unlabeledRuntimeGQL)

		unlabledRuntime := graphql.Runtime{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerUnlabeledRuntimeRequest, &unlabledRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, unlabledRuntime.ID)
		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantID, unlabledRuntime.ID)

		deleteLabelRuntimeResp := graphql.Runtime{}
		deleteLabelRequest := fixtures.FixDeleteRuntimeLabelRequest(unlabledRuntime.ID, IsNormalizedLabel)
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, deleteLabelRequest, &deleteLabelRuntimeResp)
		require.NoError(t, err)

		request := fixtures.FixApplicationForRuntimeRequest(unlabledRuntime.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)
		saveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantNormalizedApplications))
		assert.ElementsMatch(t, tenantNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications For Runtime Query with normalization", func(t *testing.T) {
		//create runtime without normalization
		runtimeInputWithNormalization := fixtures.FixRuntimeInput("normalized-runtime")
		(runtimeInputWithNormalization.Labels)[ScenariosLabel] = scenarios
		(runtimeInputWithNormalization.Labels)[IsNormalizedLabel] = "true"
		runtimeInputWithNormalizationGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInputWithNormalization)
		require.NoError(t, err)
		registerRuntimeWithNormalizationRequest := fixtures.FixRegisterRuntimeRequest(runtimeInputWithNormalizationGQL)

		runtimeWithNormalization := graphql.Runtime{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerRuntimeWithNormalizationRequest, &runtimeWithNormalization)
		require.NoError(t, err)
		require.NotEmpty(t, runtimeWithNormalization.ID)
		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantID, runtimeWithNormalization.ID)

		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)
		saveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantNormalizedApplications))
		assert.ElementsMatch(t, tenantNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications Query With Consumer Runtime", func(t *testing.T) {
		request := fixtures.FixGetApplicationsRequestWithPagination()
		applicationPage := graphql.ApplicationPage{}

		rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantID, runtimeWithoutNormalization.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, rtmOauthCredentialData, "runtime:read application:read")
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

	fixtures.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, ScenariosLabel, schema, tenantID)

	applications := []struct {
		ApplicationName string
		Scenarios       []string
		Hidden          bool
	}{
		{
			ApplicationName: "first",
			Scenarios:       []string{defaultValue},
			Hidden:          false,
		},
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

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, createApplicationReq, &application)

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		defer fixtures.UnregisterApplication(t, ctx, dexGraphQLClient, tenantID, application.ID)
		if !testApp.Hidden {
			expectedApplications = append(expectedApplications, &application)

			normalizedApp := application
			normalizedApp.Name = conf.DefaultNormalizationPrefix + normalizedApp.Name
			expectedNormalizedApplications = append(expectedNormalizedApplications, &normalizedApp)
		}
	}

	//create runtime without normalization
	runtimeWithoutNormalizationInput := fixtures.FixRuntimeInput("unnormalized-runtime")
	(runtimeWithoutNormalizationInput.Labels)[ScenariosLabel] = scenarios
	(runtimeWithoutNormalizationInput.Labels)[IsNormalizedLabel] = "false"
	runtimeWithoutNormalizationInputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeWithoutNormalizationInput)
	require.NoError(t, err)

	registerWithoutNormalizationRuntimeRequest := fixtures.FixRegisterRuntimeRequest(runtimeWithoutNormalizationInputGQL)
	runtimeWithoutNormalization := graphql.Runtime{}
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerWithoutNormalizationRuntimeRequest, &runtimeWithoutNormalization)
	require.NoError(t, err)
	require.NotEmpty(t, runtimeWithoutNormalization.ID)

	defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantID, runtimeWithoutNormalization.ID)

	t.Run("Applications For Runtime Query without normalization", func(t *testing.T) {
		//WHEN
		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithoutNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedApplications))
		assert.ElementsMatch(t, expectedApplications, applicationPage.Data)
	})

	t.Run("Applications For Runtime Query with normalization", func(t *testing.T) {
		//create runtime with normalization
		runtimeWithNormalizationInput := fixtures.FixRuntimeInput("normalized-runtime")
		(runtimeWithNormalizationInput.Labels)[ScenariosLabel] = scenarios
		(runtimeWithNormalizationInput.Labels)[IsNormalizedLabel] = "true"
		runtimeWithNormalizationInputGQL, err := testctx.Tc.Graphqlizer.RuntimeInputToGQL(runtimeWithNormalizationInput)
		require.NoError(t, err)

		registerWithNormalizationRuntimeRequest := fixtures.FixRegisterRuntimeRequest(runtimeWithNormalizationInputGQL)
		runtimeWithNormalization := graphql.Runtime{}
		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerWithNormalizationRuntimeRequest, &runtimeWithNormalization)
		require.NoError(t, err)
		require.NotEmpty(t, runtimeWithNormalization.ID)

		defer fixtures.UnregisterRuntime(t, ctx, dexGraphQLClient, tenantID, runtimeWithNormalization.ID)

		//WHEN
		request := fixtures.FixApplicationForRuntimeRequest(runtimeWithNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedNormalizedApplications))
		assert.ElementsMatch(t, expectedNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications Query With Consumer Runtime", func(t *testing.T) {
		//WHEN
		request := fixtures.FixGetApplicationsRequestWithPagination()
		applicationPage := graphql.ApplicationPage{}

		rtmAuth := fixtures.RequestClientCredentialsForRuntime(t, context.Background(), dexGraphQLClient, tenantID, runtimeWithoutNormalization.ID)
		rtmOauthCredentialData, ok := rtmAuth.Auth.Credential.(*graphql.OAuthCredentialData)
		require.True(t, ok)
		require.NotEmpty(t, rtmOauthCredentialData.ClientSecret)
		require.NotEmpty(t, rtmOauthCredentialData.ClientID)

		t.Log("Issue a Hydra token with Client Credentials")
		accessToken := token.GetAccessToken(t, rtmOauthCredentialData, "runtime:read application:read")
		oauthGraphQLClient := gql.NewAuthorizedGraphQLClientWithCustomURL(accessToken, conf.GatewayOauth)

		err = testctx.Tc.NewOperation(ctx).WithTenant(tenantID).Run(request, oauthGraphQLClient, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedApplications))
		assert.ElementsMatch(t, expectedApplications, applicationPage.Data)
	})
}
