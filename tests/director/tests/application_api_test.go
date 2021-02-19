package tests

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/jwtbuilder"

	"github.com/kyma-incubator/compass/tests/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
	webhookURL                  = "https://kyma-project.io"
	defaultScenario             = "DEFAULT"
	defaultNormalizationPrefix  = "mp-"
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
		Labels: &graphql.Labels{
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	// WHEN
	request := pkg.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)
	assertApplication(t, in, actualApp)
	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualApp.Status.Condition)
}

func TestRegisterApplicationNormalizarionValidation(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	firstAppName := "app@wordpress"

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	actualApp := pkg.RegisterApplication(t, ctx, dexGraphQLClient, firstAppName, tenant)

	//THEN
	require.NotEmpty(t, actualApp.ID)
	require.Equal(t, actualApp.Name, firstAppName)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, actualApp.ID)

	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualApp.Status.Condition)

	// SECOND APP WITH SAME APP NAME WHEN NORMALIZED
	inSecond := graphql.ApplicationRegisterInput{
		Name:           "app!wordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: &graphql.Labels{
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
		},
	}
	appSecondInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(inSecond)
	require.NoError(t, err)
	actualSecondApp := graphql.ApplicationExt{}

	// WHEN

	request := pkg.FixRegisterApplicationRequest(appSecondInputGQL)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualSecondApp)

	//THEN
	require.EqualError(t, err, "graphql: Object name is not unique [object=application]")
	require.Empty(t, actualSecondApp.ID)

	// THIRD APP WITH DIFFERENT APP NAME WHEN NORMALIZED
	actualThirdApp := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "appwordpress", tenant)

	//THEN
	require.NotEmpty(t, actualThirdApp.ID)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, actualThirdApp.ID)

	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualThirdApp.Status.Condition)

	// FOURTH APP WITH DIFFERENT ALREADY NORMALIZED NAME WHICH MATCHES EXISTING APP WHEN NORMALIZED
	inFourth := graphql.ApplicationRegisterInput{
		Name:           "mp-appwordpress",
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: &graphql.Labels{
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
		},
	}
	appFourthInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(inFourth)
	require.NoError(t, err)
	actualFourthApp := graphql.ApplicationExt{}
	// WHEN
	request = pkg.FixRegisterApplicationRequest(appFourthInputGQL)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualFourthApp)
	//THEN
	require.EqualError(t, err, "graphql: Object name is not unique [object=application]")
	require.Empty(t, actualFourthApp.ID)

	// FIFTH APP WITH DIFFERENT ALREADY NORMALIZED NAME WHICH DOES NOT MATCH ANY EXISTING APP WHEN NORMALIZED
	fifthAppName := "mp-application"
	actualFifthApp := pkg.RegisterApplication(t, ctx, dexGraphQLClient, fifthAppName, tenant)
	//THEN
	require.NotEmpty(t, actualFifthApp.ID)
	require.Equal(t, actualFifthApp.Name, fifthAppName)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, actualFifthApp.ID)

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
		Labels: &graphql.Labels{
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
		},
		StatusCondition: &statusCond,
	}

	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	request := pkg.FixRegisterApplicationRequest(appInputGQL)

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	// WHEN
	actualApp := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with status")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)
	assertApplication(t, in, actualApp)
	assert.Equal(t, statusCond, actualApp.Status.Condition)
}

func TestRegisterApplicationWithWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationRegisterInput{
		Name:         "wordpress",
		ProviderName: ptr.String("compass"),
		Webhooks: []*graphql.WebhookInput{
			{
				Type: graphql.WebhookTypeConfigurationChanged,
				Auth: pkg.FixBasicAuth(t),
				URL:  ptr.String("http://mywordpress.com/webhooks1"),
			},
		},
		Labels: &graphql.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	// WHEN
	request := pkg.FixRegisterApplicationRequest(appInputGQL)
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with webhooks")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestRegisterApplicationWithBundles(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := pkg.FixApplicationRegisterInputWithBundles(t)
	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	// WHEN
	request := pkg.FixRegisterApplicationRequest(appInputGQL)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with bundles")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)
	assertApplication(t, in, actualApp)
}

// TODO: Delete after bundles are adopted

func TestRegisterApplicationWithPackagesBackwardsCompatibility(t *testing.T) {
	ctx := context.Background()
	expectedAppName := "create-app-with-packages"

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	type ApplicationWithPackagesExt struct {
		graphql.Application
		Labels                graphql.Labels                           `json:"labels"`
		Webhooks              []graphql.Webhook                        `json:"webhooks"`
		Auths                 []*graphql.SystemAuth                    `json:"auths"`
		Package               graphql.BundleExt                        `json:"package"`
		Packages              graphql.BundlePageExt                    `json:"packages"`
		EventingConfiguration graphql.ApplicationEventingConfiguration `json:"eventingConfiguration"`
	}

	t.Run("Register Application with Packages when useBundles=false", func(t *testing.T) {
		var actualApp ApplicationWithPackagesExt
		request := pkg.FixRegisterApplicationWithPackagesRequest(expectedAppName)
		err := pkg.Tc.NewOperation(ctx).WithQueryParam("useBundles", "false").Run(request, dexGraphQLClient, &actualApp)

		appID := actualApp.ID
		packageID := actualApp.Packages.Data[0].ID

		require.NoError(t, err)
		require.NotEmpty(t, appID)

		defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), appID)

		require.NotEmpty(t, packageID)
		require.Equal(t, expectedAppName, actualApp.Name)

		t.Run("Get Application with Package when useBundles=false should succeed", func(t *testing.T) {
			var actualAppWithPackage ApplicationWithPackagesExt
			request := pkg.FixGetApplicationWithPackageRequest(appID, packageID)
			err := pkg.Tc.NewOperation(ctx).WithQueryParam("useBundles", "false").Run(request, dexGraphQLClient, &actualAppWithPackage)

			require.NoError(t, err)
			require.NotEmpty(t, actualAppWithPackage.ID)
			require.NotEmpty(t, actualAppWithPackage.Package.ID)
		})

		t.Run("Get Application with Package when useBundles=true should fail", func(t *testing.T) {
			var actualAppWithPackage ApplicationWithPackagesExt
			request := pkg.FixGetApplicationWithPackageRequest(appID, packageID)
			err := pkg.Tc.NewOperation(ctx).
				WithQueryParam("useBundles", "true").
				Run(request, dexGraphQLClient, &actualAppWithPackage)

			require.Error(t, err)
			require.Empty(t, actualAppWithPackage.ID)
		})

		runtimeInput := pkg.FixRuntimeInput("test-runtime")
		(*runtimeInput.Labels)[ScenariosLabel] = []string{"DEFAULT"}
		runtimeInputGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInput)
		require.NoError(t, err)
		registerRuntimeRequest := pkg.FixRegisterRuntimeRequest(runtimeInputGQL)

		runtime := graphql.Runtime{}
		err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, registerRuntimeRequest, &runtime)
		require.NoError(t, err)
		require.NotEmpty(t, runtime.ID)

		defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), runtime.ID)

		t.Run("Get ApplicationForRuntime with Package when useBundles=false should succeed", func(t *testing.T) {
			applicationPage := struct {
				Data []*ApplicationWithPackagesExt `json:"data"`
			}{}
			request := pkg.FixApplicationsForRuntimeWithPackagesRequest(runtime.ID)
			err := pkg.Tc.NewOperation(ctx).WithConsumer(&jwtbuilder.Consumer{
				ID:   runtime.ID,
				Type: jwtbuilder.RuntimeConsumer,
			}).WithQueryParam("useBundles", "false").Run(request, dexGraphQLClient, &applicationPage)

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

		t.Run("Get ApplicationForRuntime with Package when useBundles=true should fail", func(t *testing.T) {
			var actualAppWithPackage ApplicationWithPackagesExt
			request := pkg.FixApplicationsForRuntimeWithPackagesRequest(runtime.ID)
			err := pkg.Tc.NewOperation(ctx).WithConsumer(&jwtbuilder.Consumer{
				ID:   runtime.ID,
				Type: jwtbuilder.RuntimeConsumer,
			}).WithQueryParam("useBundles", "true").Run(request, dexGraphQLClient, &actualAppWithPackage)

			require.Error(t, err)
			require.Empty(t, actualAppWithPackage.ID)
		})
	})

	t.Run("Register Application with Packages when useBundles=true should fail", func(t *testing.T) {
		var actualApp ApplicationWithPackagesExt
		request := pkg.FixRegisterApplicationWithPackagesRequest("failed-app-with-packages")
		op := pkg.Tc.NewOperation(ctx).WithQueryParam("useBundles", "true")
		err := op.Run(request, dexGraphQLClient, &actualApp)

		require.Error(t, err)
		require.Empty(t, actualApp.ID)
		require.Empty(t, actualApp.Name)
	})
}

func TestCreateApplicationWithNonExistentIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	in := pkg.FixSampleApplicationCreateInputWithIntegrationSystem("placeholder")
	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	request := pkg.FixRegisterApplicationRequest(appInputGQL)

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	// WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)

	//THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "Object not found")
}

func TestUpdateApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	actualApp := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "before", pkg.TestTenants.GetDefaultTenantID())
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)

	updateStatusCond := graphql.ApplicationStatusConditionConnected

	expectedApp := actualApp
	expectedApp.Name = "before"
	expectedApp.ProviderName = ptr.String("after")
	expectedApp.Description = ptr.String("after")
	expectedApp.HealthCheckURL = ptr.String(webhookURL)
	expectedApp.Status.Condition = updateStatusCond
	expectedApp.Labels["name"] = "before"

	updateInput := pkg.FixSampleApplicationUpdateInput("after")
	updateInput.StatusCondition = &updateStatusCond
	updateInputGQL, err := pkg.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)
	request := pkg.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &updatedApp)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	actualApp := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "before", pkg.TestTenants.GetDefaultTenantID())
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)

	updateInput := pkg.FixSampleApplicationUpdateInputWithIntegrationSystem("after")
	updateInputGQL, err := pkg.Tc.Graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)
	request := pkg.FixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}

	//WHEN
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &updatedApp)

	//THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "Object not found")
}

func TestCreateApplicationWithDuplicatedNamesWithinTenant(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appName := "samename"

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	actualApp := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, pkg.TestTenants.GetDefaultTenantID())
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)

	t.Run("Error when creating second Application with same name", func(t *testing.T) {
		in := pkg.FixSampleApplicationRegisterInputWithNameAndWebhooks("first", appName)
		appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)
		request := pkg.FixRegisterApplicationRequest(appInputGQL)

		// WHEN
		err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not unique")
	})
}

func TestDeleteApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := pkg.FixSampleApplicationRegisterInputWithWebhooks("app")

	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createReq := pkg.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &actualApp)
	require.NoError(t, err)

	require.NotEmpty(t, actualApp.ID)

	// WHEN
	delReq := pkg.FixUnregisterApplicationRequest(actualApp.ID)
	saveExample(t, delReq.Query(), "unregister application")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, delReq, &actualApp)

	//THEN
	require.NoError(t, err)
}

func TestUpdateApplicationParts(t *testing.T) {
	ctx := context.Background()
	placeholder := "app"
	in := pkg.FixSampleApplicationRegisterInputWithWebhooks(placeholder)

	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createReq := pkg.FixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}

	tenant := pkg.TestTenants.GetDefaultTenantID()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)

	t.Run("labels manipulation", func(t *testing.T) {
		expectedLabel := graphql.Label{Key: "brand_new_label", Value: []interface{}{"aaa", "bbb"}}

		// add label
		createdLabel := &graphql.Label{}

		addReq := pkg.FixSetApplicationLabelRequest(actualApp.ID, expectedLabel.Key, []string{"aaa", "bbb"})
		saveExample(t, addReq.Query(), "set application label")

		err := pkg.Tc.RunOperation(ctx, dexGraphQLClient, addReq, &createdLabel)
		require.NoError(t, err)
		assert.Equal(t, &expectedLabel, createdLabel)

		actualApp := pkg.GetApplication(t, ctx, dexGraphQLClient, actualApp.ID, tenant)
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "aaa")
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "bbb")

		// delete label value
		deletedLabel := graphql.Label{}
		delReq := pkg.FixDeleteApplicationLabelRequest(actualApp.ID, expectedLabel.Key)
		saveExample(t, delReq.Query(), "delete application label")
		err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, delReq, &deletedLabel)
		require.NoError(t, err)
		assert.Equal(t, expectedLabel, deletedLabel)
		actualApp = pkg.GetApplication(t, ctx, dexGraphQLClient, actualApp.ID, pkg.TestTenants.GetDefaultTenantID())
		assert.Nil(t, actualApp.Labels[expectedLabel.Key])

	})

	t.Run("manage webhooks", func(t *testing.T) {
		// add
		webhookInStr, err := pkg.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL:  ptr.String("http://new-webhook.url"),
			Type: graphql.WebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		addReq := pkg.FixAddWebhookRequest(actualApp.ID, webhookInStr)
		saveExampleInCustomDir(t, addReq.Query(), addWebhookCategory, "add application webhook")

		actualWebhook := graphql.Webhook{}
		err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, addReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "http://new-webhook.url", actualWebhook.URL)
		assert.Equal(t, graphql.WebhookTypeConfigurationChanged, actualWebhook.Type)
		id := actualWebhook.ID
		require.NotNil(t, id)

		// get all webhooks
		updatedApp := pkg.GetApplication(t, ctx, dexGraphQLClient, actualApp.ID, tenant)
		assert.Len(t, updatedApp.Webhooks, 2)

		// update
		webhookInStr, err = pkg.Tc.Graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL: ptr.String("http://updated-webhook.url"), Type: graphql.WebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		updateReq := pkg.FixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
		saveExampleInCustomDir(t, updateReq.Query(), updateWebhookCategory, "update application webhook")
		err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "http://updated-webhook.url", actualWebhook.URL)

		// delete

		//GIVEN
		deleteReq := pkg.FixDeleteWebhookRequest(actualWebhook.ID)
		saveExampleInCustomDir(t, deleteReq.Query(), deleteWebhookCategory, "delete application webhook")

		//WHEN
		err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, deleteReq, &actualWebhook)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, "http://updated-webhook.url", actualWebhook.URL)

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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	for i := 0; i < 3; i++ {
		in := graphql.ApplicationRegisterInput{
			Name: fmt.Sprintf("app-%d", i),
		}

		appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)

		actualApp := graphql.Application{}
		request := pkg.FixRegisterApplicationRequest(appInputGQL)

		err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, request, &actualApp)
		require.NoError(t, err)
		defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)
	}
	actualAppPage := graphql.ApplicationPage{}

	// WHEN
	queryReq := pkg.FixGetApplicationsRequestWithPagination()
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, queryReq, &actualAppPage)
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	apps := make(map[string]*graphql.ApplicationExt)
	for i := 0; i < appAmount; i++ {
		app := pkg.RegisterApplication(t, ctx, dexGraphQLClient, fmt.Sprintf("app-%d", i), tenant)
		defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), app.ID)
		apps[app.ID] = &app
	}
	appsPage := graphql.ApplicationPageExt{}

	// WHEN
	queriesForFullPage := appAmount / after
	for i := 0; i < queriesForFullPage; i++ {
		appReq := pkg.FixApplicationsPageableRequest(after, cursor)
		err := pkg.Tc.RunOperation(ctx, dexGraphQLClient, appReq, &appsPage)
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

	appReq := pkg.FixApplicationsPageableRequest(after, cursor)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, appReq, &appsPage)
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

	appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	actualApp := graphql.Application{}
	request := pkg.FixRegisterApplicationRequest(appInputGQL)
	err = pkg.Tc.RunOperation(context.Background(), dexGraphQLClient, request, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	appID := actualApp.ID
	defer pkg.UnregisterApplication(t, context.Background(), dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), appID)

	t.Run("Query Application With Consumer User", func(t *testing.T) {
		actualApp := graphql.Application{}

		// WHEN
		queryAppReq := pkg.FixGetApplicationRequest(appID)
		err = pkg.Tc.RunOperation(context.Background(), dexGraphQLClient, queryAppReq, &actualApp)
		saveExampleInCustomDir(t, queryAppReq.Query(), queryApplicationCategory, "query application")

		//THE
		require.NoError(t, err)
		assert.Equal(t, appID, actualApp.ID)
	})

	ctx := context.Background()

	input := pkg.FixRuntimeInput("runtime-test")

	runtime := pkg.RegisterRuntimeFromInputWithinTenant(t, ctx, dexGraphQLClient, tenant, &input)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, tenant, runtime.ID)

	scenarios := []string{defaultScenario, "test-scenario"}

	// update label definitions
	pkg.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), scenarios)
	defer pkg.UpdateScenariosLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), scenarios[:1])

	runtimeConsumer := pkg.Tc.NewOperation(ctx).WithConsumer(&jwtbuilder.Consumer{
		ID:   runtime.ID,
		Type: jwtbuilder.RuntimeConsumer,
	})

	t.Run("Query Application With Consumer Runtime in same scenario", func(t *testing.T) {
		// set application scenarios label
		pkg.SetApplicationLabel(t, ctx, dexGraphQLClient, appID, ScenariosLabel, scenarios[1:])
		defer pkg.SetApplicationLabel(t, ctx, dexGraphQLClient, appID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

		actualApp := graphql.Application{}

		// WHEN
		queryAppReq := pkg.FixGetApplicationRequest(appID)
		err = runtimeConsumer.Run(queryAppReq, dexGraphQLClient, &actualApp)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, appID, actualApp.ID)
	})

	t.Run("Query Application With Consumer Runtime not in same scenario", func(t *testing.T) {
		// set application scenarios label
		pkg.SetApplicationLabel(t, ctx, dexGraphQLClient, appID, ScenariosLabel, scenarios[:1])

		// set runtime scenarios label
		pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[1:])
		defer pkg.SetRuntimeLabel(t, ctx, dexGraphQLClient, tenant, runtime.ID, ScenariosLabel, scenarios[:1])

		actualApp := graphql.Application{}

		// WHEN
		queryAppReq := pkg.FixGetApplicationRequest(appID)
		err = runtimeConsumer.Run(queryAppReq, dexGraphQLClient, &actualApp)

		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "The operation is not allowed")
	})
}

func TestTenantSeparation(t *testing.T) {
	// GIVEN
	appIn := pkg.FixSampleApplicationRegisterInputWithWebhooks("tenantseparation")
	inStr, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(appIn)
	require.NoError(t, err)

	createReq := pkg.FixRegisterApplicationRequest(inStr)
	actualApp := graphql.ApplicationExt{}
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, pkg.TestTenants.GetDefaultTenantID(), actualApp.ID)

	// WHEN
	getAppReq := pkg.FixGetApplicationsRequestWithPagination()
	customTenant := pkg.TestTenants.GetIDByName(t, "Test1")
	anotherTenantsApps := graphql.ApplicationPage{}
	// THEN
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, customTenant, getAppReq, &anotherTenantsApps)
	require.NoError(t, err)
	assert.Empty(t, anotherTenantsApps.Data)
}

func TestApplicationsForRuntime(t *testing.T) {
	//GIVEN
	ctx := context.Background()
	tenantID := pkg.TestTenants.GetIDByName(t, "Test1")
	otherTenant := pkg.TestTenants.GetIDByName(t, "Test2")
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	pkg.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, ScenariosLabel, schema, tenantID)
	pkg.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, ScenariosLabel, schema, otherTenant)

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
		applicationInput := pkg.FixSampleApplicationRegisterInputWithWebhooks(testApp.ApplicationName)
		applicationInput.Labels = &graphql.Labels{ScenariosLabel: testApp.Scenarios}
		appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := pkg.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.Application{}

		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, testApp.Tenant, createApplicationReq, &application)

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, testApp.Tenant, application.ID)
		if testApp.WithinTenant {
			tenantUnnormalizedApplications = append(tenantUnnormalizedApplications, &application)

			normalizedApp := application
			normalizedApp.Name = defaultNormalizationPrefix + normalizedApp.Name
			tenantNormalizedApplications = append(tenantNormalizedApplications, &normalizedApp)
		}
	}

	//create runtime without normalization
	runtimeInputWithoutNormalization := pkg.FixRuntimeInput("unnormalized-runtime")
	(*runtimeInputWithoutNormalization.Labels)[ScenariosLabel] = scenarios
	(*runtimeInputWithoutNormalization.Labels)[IsNormalizedLabel] = "false"
	runtimeInputWithoutNormalizationGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInputWithoutNormalization)
	require.NoError(t, err)
	registerRuntimeWithNormalizationRequest := pkg.FixRegisterRuntimeRequest(runtimeInputWithoutNormalizationGQL)

	runtimeWithoutNormalization := graphql.Runtime{}
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerRuntimeWithNormalizationRequest, &runtimeWithoutNormalization)
	require.NoError(t, err)
	require.NotEmpty(t, runtimeWithoutNormalization.ID)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, runtimeWithoutNormalization.ID, tenantID)

	t.Run("Applications For Runtime Query without normalization", func(t *testing.T) {
		request := pkg.FixApplicationForRuntimeRequest(runtimeWithoutNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)
		saveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantUnnormalizedApplications))
		assert.ElementsMatch(t, tenantUnnormalizedApplications, applicationPage.Data)

	})

	t.Run("Applications For Runtime Query without normalization due to missing label", func(t *testing.T) {
		//create runtime without normalization
		unlabeledRuntimeInput := pkg.FixRuntimeInput("unlabeled-runtime")
		(*unlabeledRuntimeInput.Labels)[ScenariosLabel] = scenarios
		(*unlabeledRuntimeInput.Labels)[IsNormalizedLabel] = "false"
		unlabeledRuntimeGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(unlabeledRuntimeInput)
		require.NoError(t, err)
		registerUnlabeledRuntimeRequest := pkg.FixRegisterRuntimeRequest(unlabeledRuntimeGQL)

		unlabledRuntime := graphql.Runtime{}
		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerUnlabeledRuntimeRequest, &unlabledRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, unlabledRuntime.ID)
		defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, unlabledRuntime.ID, tenantID)

		deleteLabelRuntimeResp := graphql.Runtime{}
		deleteLabelRequest := pkg.FixDeleteRuntimeLabelRequest(unlabledRuntime.ID, IsNormalizedLabel)
		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, deleteLabelRequest, &deleteLabelRuntimeResp)
		require.NoError(t, err)

		request := pkg.FixApplicationForRuntimeRequest(unlabledRuntime.ID)
		applicationPage := graphql.ApplicationPage{}

		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)
		saveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantNormalizedApplications))
		assert.ElementsMatch(t, tenantNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications For Runtime Query with normalization", func(t *testing.T) {
		//create runtime without normalization
		runtimeInputWithNormalization := pkg.FixRuntimeInput("normalized-runtime")
		(*runtimeInputWithNormalization.Labels)[ScenariosLabel] = scenarios
		(*runtimeInputWithNormalization.Labels)[IsNormalizedLabel] = "true"
		runtimeInputWithNormalizationGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(runtimeInputWithNormalization)
		require.NoError(t, err)
		registerRuntimeWithNormalizationRequest := pkg.FixRegisterRuntimeRequest(runtimeInputWithNormalizationGQL)

		runtimeWithNormalization := graphql.Runtime{}
		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerRuntimeWithNormalizationRequest, &runtimeWithNormalization)
		require.NoError(t, err)
		require.NotEmpty(t, runtimeWithNormalization.ID)
		defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, runtimeWithNormalization.ID, tenantID)

		request := pkg.FixApplicationForRuntimeRequest(runtimeWithNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)
		saveExample(t, request.Query(), "query applications for runtime")

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantNormalizedApplications))
		assert.ElementsMatch(t, tenantNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications Query With Consumer Runtime", func(t *testing.T) {
		request := pkg.FixGetApplicationsRequestWithPagination()
		applicationPage := graphql.ApplicationPage{}

		err = pkg.Tc.NewOperation(ctx).WithTenant(tenantID).WithConsumer(&jwtbuilder.Consumer{
			ID:   runtimeWithoutNormalization.ID,
			Type: jwtbuilder.RuntimeConsumer,
		}).Run(request, dexGraphQLClient, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(tenantUnnormalizedApplications))
		assert.ElementsMatch(t, tenantUnnormalizedApplications, applicationPage.Data)
	})
}

func TestApplicationsForRuntimeWithHiddenApps(t *testing.T) {
	//GIVEN
	ctx := context.Background()
	tenantID := pkg.TestTenants.GetIDByName(t, "TestApplicationsForRuntimeWithHiddenApps")
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

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	pkg.CreateLabelDefinitionWithinTenant(t, ctx, dexGraphQLClient, ScenariosLabel, schema, tenantID)

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
		applicationInput := pkg.FixSampleApplicationRegisterInputWithWebhooks(testApp.ApplicationName)
		applicationInput.Labels = &graphql.Labels{ScenariosLabel: testApp.Scenarios}
		if testApp.Hidden {
			(*applicationInput.Labels)[applicationHideSelectorKey] = applicationHideSelectorValue
		}
		appInputGQL, err := pkg.Tc.Graphqlizer.ApplicationRegisterInputToGQL(applicationInput)
		require.NoError(t, err)

		createApplicationReq := pkg.FixRegisterApplicationRequest(appInputGQL)
		application := graphql.Application{}

		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, createApplicationReq, &application)

		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenantID, application.ID)
		if !testApp.Hidden {
			expectedApplications = append(expectedApplications, &application)

			normalizedApp := application
			normalizedApp.Name = defaultNormalizationPrefix + normalizedApp.Name
			expectedNormalizedApplications = append(expectedNormalizedApplications, &normalizedApp)
		}
	}

	//create runtime without normalization
	runtimeWithoutNormalizationInput := pkg.FixRuntimeInput("unnormalized-runtime")
	(*runtimeWithoutNormalizationInput.Labels)[ScenariosLabel] = scenarios
	(*runtimeWithoutNormalizationInput.Labels)[IsNormalizedLabel] = "false"
	runtimeWithoutNormalizationInputGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(runtimeWithoutNormalizationInput)
	require.NoError(t, err)
	registerWithoutNormalizationRuntimeRequest := pkg.FixRegisterRuntimeRequest(runtimeWithoutNormalizationInputGQL)
	runtimeWithoutNormalization := graphql.Runtime{}
	err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerWithoutNormalizationRuntimeRequest, &runtimeWithoutNormalization)
	require.NoError(t, err)
	require.NotEmpty(t, runtimeWithoutNormalization.ID)
	defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, runtimeWithoutNormalization.ID, tenantID)

	t.Run("Applications For Runtime Query without normalization", func(t *testing.T) {
		//WHEN
		request := pkg.FixApplicationForRuntimeRequest(runtimeWithoutNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedApplications))
		assert.ElementsMatch(t, expectedApplications, applicationPage.Data)
	})

	t.Run("Applications For Runtime Query with normalization", func(t *testing.T) {
		//create runtime with normalization
		runtimeWithNormalizationInput := pkg.FixRuntimeInput("normalized-runtime")
		(*runtimeWithNormalizationInput.Labels)[ScenariosLabel] = scenarios
		(*runtimeWithNormalizationInput.Labels)[IsNormalizedLabel] = "true"
		runtimeWithNormalizationInputGQL, err := pkg.Tc.Graphqlizer.RuntimeInputToGQL(runtimeWithNormalizationInput)
		require.NoError(t, err)
		registerWithNormalizationRuntimeRequest := pkg.FixRegisterRuntimeRequest(runtimeWithNormalizationInputGQL)
		runtimeWithNormalization := graphql.Runtime{}
		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, registerWithNormalizationRuntimeRequest, &runtimeWithNormalization)
		require.NoError(t, err)
		require.NotEmpty(t, runtimeWithNormalization.ID)
		defer pkg.UnregisterRuntime(t, ctx, dexGraphQLClient, runtimeWithNormalization.ID, tenantID)

		//WHEN
		request := pkg.FixApplicationForRuntimeRequest(runtimeWithNormalization.ID)
		applicationPage := graphql.ApplicationPage{}

		err = pkg.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, tenantID, request, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedNormalizedApplications))
		assert.ElementsMatch(t, expectedNormalizedApplications, applicationPage.Data)
	})

	t.Run("Applications Query With Consumer Runtime", func(t *testing.T) {
		//WHEN
		request := pkg.FixGetApplicationsRequestWithPagination()
		applicationPage := graphql.ApplicationPage{}

		err = pkg.Tc.NewOperation(ctx).WithTenant(tenantID).WithConsumer(&jwtbuilder.Consumer{
			ID:   runtimeWithoutNormalization.ID,
			Type: jwtbuilder.RuntimeConsumer,
		}).Run(request, dexGraphQLClient, &applicationPage)

		//THEN
		require.NoError(t, err)
		require.Len(t, applicationPage.Data, len(expectedApplications))
		assert.ElementsMatch(t, expectedApplications, applicationPage.Data)
	})
}
