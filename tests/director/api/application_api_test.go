package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
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

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := fixRegisterApplicationRequest(appInputGQL)
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
	assert.Equal(t, graphql.ApplicationStatusConditionInitial, actualApp.Status.Condition)
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

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	request := fixRegisterApplicationRequest(appInputGQL)

	// WHEN
	actualApp := graphql.ApplicationExt{}
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with status")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)
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
				Type: graphql.ApplicationWebhookTypeConfigurationChanged,
				Auth: fixBasicAuth(),
				URL:  "http://mywordpress.com/webhooks1",
			},
		},
		Labels: &graphql.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := fixRegisterApplicationRequest(appInputGQL)
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with webhooks")
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestRegisterApplicationWithPackages(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := fixApplicationRegisterInputWithPackages()
	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := fixRegisterApplicationRequest(appInputGQL)
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with packages")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithNonExistentIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	in := fixSampleApplicationCreateInputWithIntegrationSystem("placeholder")
	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	request := fixRegisterApplicationRequest(appInputGQL)
	// WHEN
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "does not exist")
}

func TestAddDependentObjectsWhenAppDoesNotExist(t *testing.T) {
	applicationID := "cf889c38-490d-4896-96a7-c0721eca9932"

	t.Run("add Webhook", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		webhookInStr, err := tc.graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL:  webhookURL,
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})
		require.NoError(t, err)

		//WHEN
		addWebhookReq := fixAddWebhookRequest(applicationID, webhookInStr)
		err = tc.RunOperation(ctx, addWebhookReq, nil)

		//THEN
		require.EqualError(t, err, "graphql: Cannot add Webhook to not existing Application")
	})

	t.Run("add API Definition", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		apiInStr, err := tc.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{
			Name:      "new-api-name",
			TargetURL: "https://target.url",
		})
		require.NoError(t, err)

		// WHEN
		addAPIDefReq := fixAddAPIRequest(applicationID, apiInStr)
		err = tc.RunOperation(ctx, addAPIDefReq, nil)

		//THEN
		require.EqualError(t, err, "graphql: Cannot add API to not existing Application")
	})

	t.Run("add Event Definition", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		eventApiInStr, err := tc.graphqlizer.EventDefinitionInputToGQL(graphql.EventDefinitionInput{
			Name: "new-event-api",
			Spec: &graphql.EventSpecInput{
				Type:   graphql.EventSpecTypeAsyncAPI,
				Format: graphql.SpecFormatYaml,
				FetchRequest: &graphql.FetchRequestInput{
					URL: "https://kyma-project.io",
				},
			},
		})
		require.NoError(t, err)

		// WHEN
		addEventDefReq := fixAddEventAPIRequest(applicationID, eventApiInStr)
		err = tc.RunOperation(ctx, addEventDefReq, nil)

		// THEN
		require.EqualError(t, err, "graphql: Cannot add Event Definition to not existing Application")
	})

	t.Run("add Document", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		documentInStr, err := tc.graphqlizer.DocumentInputToGQL(&graphql.DocumentInput{
			Title:       "new-document",
			Format:      graphql.DocumentFormatMarkdown,
			DisplayName: "new-document-display-name",
			Description: "new-description",
		})
		require.NoError(t, err)

		// WHEN
		addDocReq := fixAddDocumentRequest(applicationID, documentInStr)
		err = tc.RunOperation(ctx, addDocReq, nil)

		//THEN
		require.EqualError(t, err, "graphql: Cannot add Document to not existing Application")
	})
}

func TestUpdateApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	actualApp := registerApplication(t, ctx, "before")
	defer unregisterApplication(t, actualApp.ID)

	updateStatusCond := graphql.ApplicationStatusConditionConnected

	expectedApp := actualApp
	expectedApp.Name = "before"
	expectedApp.ProviderName = ptr.String("after")
	expectedApp.Description = ptr.String("after")
	expectedApp.HealthCheckURL = ptr.String(webhookURL)
	expectedApp.Status.Condition = updateStatusCond
	expectedApp.Labels["name"] = "before"

	updateInput := fixSampleApplicationUpdateInput("after")
	updateInput.StatusCondition = &updateStatusCond
	updateInputGQL, err := tc.graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)
	request := fixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}

	//WHEN
	err = tc.RunOperation(ctx, request, &updatedApp)

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

	actualApp := registerApplication(t, ctx, "before")
	defer unregisterApplication(t, actualApp.ID)

	updateInput := fixSampleApplicationUpdateInputWithIntegrationSystem("after")
	updateInputGQL, err := tc.graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)
	request := fixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}

	//WHEN
	err = tc.RunOperation(ctx, request, &updatedApp)

	//THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "does not exist")
}

func TestCreateApplicationWithDuplicatedNamesWithinTenant(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appName := "samename"

	actualApp := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, actualApp.ID)

	t.Run("Error when creating second Application with same name", func(t *testing.T) {
		in := fixSampleApplicationRegisterInputWithName("first", appName)
		appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)
		request := fixRegisterApplicationRequest(appInputGQL)

		// WHEN
		err = tc.RunOperation(ctx, request, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not unique")
	})
}

func TestDeleteApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := fixSampleApplicationRegisterInput("app")

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	createReq := fixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}
	err = tc.RunOperation(ctx, createReq, &actualApp)
	require.NoError(t, err)

	require.NotEmpty(t, actualApp.ID)

	// WHEN
	delReq := fixUnregisterApplicationRequest(actualApp.ID)
	saveExample(t, delReq.Query(), "unregister application")
	err = tc.RunOperation(ctx, delReq, &actualApp)

	//THEN
	require.NoError(t, err)
}

func TestUpdateApplicationParts(t *testing.T) {
	ctx := context.Background()
	placeholder := "app"
	in := fixSampleApplicationRegisterInput(placeholder)

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	createReq := fixRegisterApplicationRequest(appInputGQL)
	actualApp := graphql.ApplicationExt{}
	err = tc.RunOperation(ctx, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)

	t.Run("labels manipulation", func(t *testing.T) {
		expectedLabel := graphql.Label{Key: "brand_new_label", Value: []interface{}{"aaa", "bbb"}}

		// add label
		createdLabel := &graphql.Label{}

		addReq := fixSetApplicationLabelRequest(actualApp.ID, expectedLabel.Key, []string{"aaa", "bbb"})
		saveExample(t, addReq.Query(), "set application label")
		err := tc.RunOperation(ctx, addReq, &createdLabel)
		require.NoError(t, err)
		assert.Equal(t, &expectedLabel, createdLabel)
		actualApp := getApplication(t, ctx, actualApp.ID)
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "aaa")
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "bbb")

		// delete label value
		deletedLabel := graphql.Label{}
		delReq := fixDeleteApplicationLabelRequest(actualApp.ID, expectedLabel.Key)
		saveExample(t, delReq.Query(), "delete application label")
		err = tc.RunOperation(ctx, delReq, &deletedLabel)
		require.NoError(t, err)
		assert.Equal(t, expectedLabel, deletedLabel)
		actualApp = getApplication(t, ctx, actualApp.ID)
		assert.Nil(t, actualApp.Labels[expectedLabel.Key])

	})

	t.Run("manage webhooks", func(t *testing.T) {
		// add
		webhookInStr, err := tc.graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL:  "http://new-webhook.url",
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		addReq := fixAddWebhookRequest(actualApp.ID, webhookInStr)
		saveExampleInCustomDir(t, addReq.Query(), addWebhookCategory, "add application webhook")

		actualWebhook := graphql.Webhook{}
		err = tc.RunOperation(ctx, addReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "http://new-webhook.url", actualWebhook.URL)
		assert.Equal(t, graphql.ApplicationWebhookTypeConfigurationChanged, actualWebhook.Type)
		id := actualWebhook.ID
		require.NotNil(t, id)

		// get all webhooks
		updatedApp := getApplication(t, ctx, actualApp.ID)
		assert.Len(t, updatedApp.Webhooks, 2)

		// update
		webhookInStr, err = tc.graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL: "http://updated-webhook.url", Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		updateReq := fixUpdateWebhookRequest(actualWebhook.ID, webhookInStr)
		saveExampleInCustomDir(t, updateReq.Query(), updateWebhookCategory, "update application webhook")
		err = tc.RunOperation(ctx, updateReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "http://updated-webhook.url", actualWebhook.URL)

		// delete

		//GIVEN
		deleteReq := fixDeleteWebhookRequest(actualWebhook.ID)
		saveExampleInCustomDir(t, deleteReq.Query(), deleteWebhookCategory, "delete application webhook")

		//WHEN
		err = tc.RunOperation(ctx, deleteReq, &actualWebhook)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, "http://updated-webhook.url", actualWebhook.URL)

	})

	t.Run("manage API Definitions", func(t *testing.T) {
		// add
		inStr, err := tc.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{
			Name:      "new-api-name",
			TargetURL: "https://target.url",
			Spec: &graphql.APISpecInput{
				Format: graphql.SpecFormatJSON,
				Type:   graphql.APISpecTypeOpenAPI,
				FetchRequest: &graphql.FetchRequestInput{
					URL: "https://foo.bar",
				},
			},
		})

		require.NoError(t, err)
		actualAPI := graphql.APIDefinition{}

		// WHEN
		addReq := fixAddAPIRequest(actualApp.ID, inStr)
		err = tc.RunOperation(ctx, addReq, &actualAPI)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, actualAPI.ID)
		assert.Equal(t, "new-api-name", actualAPI.Name)
		assert.Equal(t, "https://target.url", actualAPI.TargetURL)

		updatedApp := getApplication(t, ctx, actualApp.ID)
		assert.Len(t, updatedApp.APIDefinitions.Data, 2)
		actualAPINames := make(map[string]struct{})
		for _, api := range updatedApp.APIDefinitions.Data {
			actualAPINames[api.Name] = struct{}{}
		}
		assert.Contains(t, actualAPINames, "new-api-name")
		assert.Contains(t, actualAPINames, placeholder)

		// update

		//GIVEN
		updateStr, err := tc.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{Name: "updated-api-name", TargetURL: "http://updated-target.url"})
		require.NoError(t, err)
		updatedAPI := graphql.APIDefinition{}

		// WHEN
		updateReq := fixUpdateAPIRequest(actualAPI.ID, updateStr)
		err = tc.RunOperation(ctx, updateReq, &updatedAPI)
		saveExample(t, updateReq.Query(), "update API Definition")

		//THEN
		require.NoError(t, err)
		updatedApp = getApplication(t, ctx, actualApp.ID)
		assert.Len(t, updatedApp.APIDefinitions.Data, 2)
		actualAPINamesAfterUpdate := make(map[string]struct{})
		for _, api := range updatedApp.APIDefinitions.Data {
			actualAPINamesAfterUpdate[api.Name] = struct{}{}
		}
		assert.Contains(t, actualAPINamesAfterUpdate, "updated-api-name")
		assert.Contains(t, actualAPINamesAfterUpdate, placeholder)
		// delete
		delAPI := graphql.APIDefinition{}

		// WHEN
		deleteReq := fixDeleteAPIRequest(actualAPI.ID)
		err = tc.RunOperation(ctx, deleteReq, &delAPI)
		saveExample(t, deleteReq.Query(), "delete API Definition")

		//THEN
		require.NoError(t, err)
		assert.Equal(t, actualAPI.ID, delAPI.ID)

		app := getApplication(t, ctx, actualApp.ID)
		require.Len(t, app.APIDefinitions.Data, 1)
		assert.Equal(t, placeholder, app.APIDefinitions.Data[0].Name)

	})

	t.Run("manage event definition", func(t *testing.T) {
		// add

		// GIVEN
		inStr, err := tc.graphqlizer.EventDefinitionInputToGQL(graphql.EventDefinitionInput{
			Name: "new-event-api",
			Spec: &graphql.EventSpecInput{
				Type:   graphql.EventSpecTypeAsyncAPI,
				Format: graphql.SpecFormatYaml,
				FetchRequest: &graphql.FetchRequestInput{
					URL: "foo.bar",
				},
			},
		})

		actualEventAPI := graphql.EventDefinition{}
		require.NoError(t, err)

		// WHEN
		addReq := fixAddEventAPIRequest(actualApp.ID, inStr)
		err = tc.RunOperation(ctx, addReq, &actualEventAPI)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "new-event-api", actualEventAPI.Name)
		assert.NotEmpty(t, actualEventAPI.ID)
		updatedApp := getApplication(t, ctx, actualApp.ID)
		assert.Len(t, updatedApp.EventDefinitions.Data, 2)

		// update

		// GIVEN
		updateStr, err := tc.graphqlizer.EventDefinitionInputToGQL(graphql.EventDefinitionInput{
			Name: "updated-event-api",
			Spec: &graphql.EventSpecInput{
				Type:   graphql.EventSpecTypeAsyncAPI,
				Format: graphql.SpecFormatYaml,
				FetchRequest: &graphql.FetchRequestInput{
					URL: "https://kyma-project.io",
				},
			}})
		require.NoError(t, err)

		// WHEN
		updateReq := fixUpdateEventAPIRequest(actualEventAPI.ID, updateStr)
		saveExample(t, updateReq.Query(), "update Event Definition")
		err = tc.RunOperation(ctx, updateReq, &actualEventAPI)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, "updated-event-api", actualEventAPI.Name)

		// delete
		// WHEN
		delReq := fixDeleteEventAPIRequest(actualEventAPI.ID)
		saveExample(t, delReq.Query(), "delete Event Definition")
		err = tc.RunOperation(ctx, delReq, nil)
		// THEN
		require.NoError(t, err)
	})

	t.Run("manage documents", func(t *testing.T) {
		// add

		//GIVEN
		inStr, err := tc.graphqlizer.DocumentInputToGQL(&graphql.DocumentInput{
			Title:       "new-document",
			Format:      graphql.DocumentFormatMarkdown,
			DisplayName: "new-document-display-name",
			Description: "new-description",
		})

		require.NoError(t, err)
		actualDoc := graphql.Document{}

		// WHEN
		addReq := fixAddDocumentRequest(actualApp.ID, inStr)
		err = tc.RunOperation(ctx, addReq, &actualDoc)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, actualDoc.ID)
		assert.Equal(t, "new-document", actualDoc.Title)

		updatedApp := getApplication(t, ctx, actualApp.ID)
		assert.Len(t, updatedApp.Documents.Data, 2)
		actualDocuTitles := make(map[string]struct{})
		for _, docu := range updatedApp.Documents.Data {
			actualDocuTitles[docu.Title] = struct{}{}
		}
		assert.Contains(t, actualDocuTitles, "new-document")
		assert.Contains(t, actualDocuTitles, placeholder)

		// delete
		delDocument := graphql.Document{}

		// WHEN
		deleteReq := fixDeleteDocumentRequest(actualDoc.ID)
		err = tc.RunOperation(ctx, deleteReq, &delDocument)
		saveExample(t, deleteReq.Query(), "delete Document")

		//THEN
		require.NoError(t, err)
		assert.Equal(t, actualDoc.ID, delDocument.ID)

		app := getApplication(t, ctx, actualApp.ID)
		require.Len(t, app.Documents.Data, 1)
		assert.Equal(t, placeholder, app.Documents.Data[0].Title)
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

		appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)
		actualApp := graphql.Application{}
		request := fixRegisterApplicationRequest(appInputGQL)
		err = tc.RunOperation(ctx, request, &actualApp)
		require.NoError(t, err)
		defer unregisterApplication(t, actualApp.ID)
	}
	actualAppPage := graphql.ApplicationPage{}

	// WHEN
	queryReq := fixApplicationsRequest()
	err := tc.RunOperation(ctx, queryReq, &actualAppPage)
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

	apps := make(map[string]*graphql.ApplicationExt)
	for i := 0; i < appAmount; i++ {
		app := registerApplication(t, ctx, fmt.Sprintf("app-%d", i))
		defer unregisterApplication(t, app.ID)
		apps[app.ID] = &app
	}
	appsPage := graphql.ApplicationPageExt{}

	// WHEN
	queriesForFullPage := appAmount / after
	for i := 0; i < queriesForFullPage; i++ {
		appReq := fixApplicationsPageableRequest(after, cursor)
		err := tc.RunOperation(ctx, appReq, &appsPage)
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

	appReq := fixApplicationsPageableRequest(after, cursor)
	err := tc.RunOperation(ctx, appReq, &appsPage)
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

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	actualApp := graphql.Application{}
	request := fixRegisterApplicationRequest(appInputGQL)
	err = tc.RunOperation(context.Background(), request, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	createdID := actualApp.ID
	defer unregisterApplication(t, actualApp.ID)

	// WHEN
	queryAppReq := fixApplicationRequest(actualApp.ID)
	err = tc.RunOperation(context.Background(), queryAppReq, &actualApp)
	saveExampleInCustomDir(t, queryAppReq.Query(), queryApplicationCategory, "query application")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdID, actualApp.ID)
}

func TestTenantSeparation(t *testing.T) {
	// GIVEN
	appIn := fixSampleApplicationRegisterInput("adidas")
	inStr, err := tc.graphqlizer.ApplicationRegisterInputToGQL(appIn)
	require.NoError(t, err)
	createReq := fixRegisterApplicationRequest(inStr)
	actualApp := graphql.ApplicationExt{}
	ctx := context.Background()
	err = tc.RunOperation(ctx, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)

	// WHEN
	getAppReq := fixApplicationsRequest()
	customTenant := testTenants.GetIDByName(t, "Test1")
	anotherTenantsApps := graphql.ApplicationPage{}
	// THEN
	err = tc.RunOperationWithCustomTenant(ctx, customTenant, getAppReq, &anotherTenantsApps)
	require.NoError(t, err)
	assert.Empty(t, anotherTenantsApps.Data)
}

func TestQuerySpecificAPIDefinition(t *testing.T) {
	// GIVEN
	in := graphql.APIDefinitionInput{
		Name:      "test",
		TargetURL: "http://target.url",
	}

	APIInputGQL, err := tc.graphqlizer.APIDefinitionInputToGQL(in)
	require.NoError(t, err)
	applicationID := registerApplication(t, context.Background(), "test").ID
	defer unregisterApplication(t, applicationID)
	actualAPI := graphql.APIDefinition{}
	request := fixAddAPIRequest(applicationID, APIInputGQL)
	err = tc.RunOperation(context.TODO(), request, &actualAPI)
	require.NoError(t, err)
	require.NotEmpty(t, actualAPI.ID)
	defer deleteAPI(t, actualAPI.ID)

	// WHEN
	queryAppReq := fixAPIDefinitionRequest(applicationID, actualAPI.ID)
	err = tc.RunOperation(context.Background(), queryAppReq, &actualAPI)

	//THEN
	require.NoError(t, err)
}

func TestQuerySpecificEventAPIDefinition(t *testing.T) {
	// GIVEN
	in := graphql.EventDefinitionInput{
		Name: "test",
		Spec: &graphql.EventSpecInput{
			Type:   graphql.EventSpecTypeAsyncAPI,
			Format: graphql.SpecFormatYaml,
			FetchRequest: &graphql.FetchRequestInput{
				URL: "https://kyma-project.io",
			},
		},
	}
	EventAPIInputGQL, err := tc.graphqlizer.EventDefinitionInputToGQL(in)
	require.NoError(t, err)
	applicationID := registerApplication(t, context.Background(), "test").ID
	defer unregisterApplication(t, applicationID)

	actualEventAPI := graphql.EventDefinition{}
	request := fixAddEventAPIRequest(applicationID, EventAPIInputGQL)
	err = tc.RunOperation(context.TODO(), request, &actualEventAPI)
	require.NoError(t, err)
	require.NotEmpty(t, actualEventAPI.ID)
	defer deleteEventAPI(t, actualEventAPI.ID)

	// WHEN
	queryAppReq := fixEventDefinitionRequest(applicationID, actualEventAPI.ID)
	err = tc.RunOperation(context.Background(), queryAppReq, &actualEventAPI)

	//THEN
	require.NoError(t, err)
}

func fixSampleApplicationRegisterInput(placeholder string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:         placeholder,
		ProviderName: ptr.String("compass"),
		Webhooks: []*graphql.WebhookInput{{
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
			URL:  webhookURL},
		},
		Labels: &graphql.Labels{placeholder: []interface{}{placeholder}},
	}
}

func fixSampleApplicationRegisterInputWithName(placeholder, name string) graphql.ApplicationRegisterInput {
	sampleInput := fixSampleApplicationRegisterInput(placeholder)
	sampleInput.Name = name
	return sampleInput
}

func fixSampleApplicationCreateInputWithIntegrationSystem(placeholder string) graphql.ApplicationRegisterInput {
	sampleInput := fixSampleApplicationRegisterInput(placeholder)
	sampleInput.IntegrationSystemID = &integrationSystemID
	return sampleInput
}

func fixSampleApplicationUpdateInput(placeholder string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Description:    &placeholder,
		HealthCheckURL: ptr.String(webhookURL),
		ProviderName:   &placeholder,
	}
}

func fixSampleApplicationUpdateInputWithIntegrationSystem(placeholder string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Description:         &placeholder,
		HealthCheckURL:      ptr.String(webhookURL),
		IntegrationSystemID: &integrationSystemID,
		ProviderName:        ptr.String(placeholder),
	}
}

func unregisterApplicationInTenant(t *testing.T, id string, tenant string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperationWithCustomTenant(context.Background(), tenant, req, nil))
}

func unregisterApplication(t *testing.T, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}

func deleteAPI(t *testing.T, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteAPIDefinition(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}

func deleteEventAPI(t *testing.T, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteEventDefinition(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}
