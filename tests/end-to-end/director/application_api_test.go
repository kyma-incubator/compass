package director

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/end-to-end/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	createApplicationCategory = "create application"
	queryApplicationsCategory = "query applications"
	queryApplicationCategory  = "query application"
	deleteWebhookCategory     = "delete webhook"
	addWebhookCategory        = "add webhook"
	updateWebhookCategory     = "update webhook"
	webhookURL                = "https://kyma-project.io"
)

var integrationSystemID = "69230297-3c81-4711-aac2-3afa8cb42e2d"

func TestCreateApplicationWithAllSimpleFieldsProvided(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationCreateInput{
		Name:           "wordpress",
		Description:    ptr.String("my first wordpress application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: &graphql.Labels{
			"group":     []interface{}{"production", "experimental"},
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)

	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {
					%s
				}
			}`,
			appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), createApplicationCategory, "create application")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationCreateInput{
		Name: "wordpress",
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

	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
				result: createApplication(in: %s) { 
						%s 
					} 
				}`,
			appInputGQL,
			tc.gqlFieldsProvider.ForApplication(),
		))
	saveExampleInCustomDir(t, request.Query(), createApplicationCategory, "create application with webhooks")
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithAPIs(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationCreateInput{
		Name: "wordpress",
		Apis: []*graphql.APIDefinitionInput{
			{
				Name:        "comments/v1",
				Description: ptr.String("api for adding comments"),
				TargetURL:   "http://mywordpress.com/comments",
				Group:       ptr.String("comments"),
				DefaultAuth: fixBasicAuth(),
				Version:     fixDepracatedVersion1(),
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOpenAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptr.CLOB(graphql.CLOB("openapi")),
				},
			},
			{
				Name:      "reviews/v1",
				TargetURL: "http://mywordpress.com/reviews",
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatJSON,
					FetchRequest: &graphql.FetchRequestInput{
						URL:    "http://mywordpress.com/apis",
						Mode:   ptr.FetchMode(graphql.FetchModePackage),
						Filter: ptr.String("odata.json"),
						Auth:   fixBasicAuth(),
					},
				},
				DefaultAuth: &graphql.AuthInput{
					Credential: fixBasicCredential(),
					RequestAuth: &graphql.CredentialRequestAuthInput{
						Csrf: &graphql.CSRFTokenCredentialRequestAuthInput{
							Credential:       fixOAuthCredential(),
							TokenEndpointURL: "token-URL",
						},
					},
				},
			},
			{
				Name:      "xml",
				TargetURL: "http://mywordpress.com/xml",
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatXML,
					Data:   ptr.CLOB(graphql.CLOB("odata")),
				},
			},
		},
		Labels: &graphql.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
 			 result: createApplication(in: %s) { 
					%s 
				}
			}`,
			appInputGQL,
			tc.gqlFieldsProvider.ForApplication(),
		))
	saveExampleInCustomDir(t, request.Query(), createApplicationCategory, "create application with APIs")

	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithEventAPIs(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationCreateInput{
		Name: "create-application-with-event-apis",
		EventAPIs: []*graphql.EventAPIDefinitionInput{
			{
				Name:        "comments/v1",
				Description: ptr.String("comments events"),
				Version:     fixDepracatedVersion1(),
				Group:       ptr.String("comments"),
				Spec: &graphql.EventAPISpecInput{
					EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
					Format:        graphql.SpecFormatYaml,
					Data:          ptr.CLOB(graphql.CLOB([]byte("asyncapi"))),
				},
			},
			{
				Name:        "reviews/v1",
				Description: ptr.String("review events"),
				Spec: &graphql.EventAPISpecInput{
					EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
					Format:        graphql.SpecFormatYaml,
					FetchRequest: &graphql.FetchRequestInput{
						URL:    "http://mywordpress.com/events",
						Mode:   ptr.FetchMode(graphql.FetchModePackage),
						Filter: ptr.String("async.json"),
						Auth:   fixOauthAuth(),
					},
				},
			},
		},
		Labels: &graphql.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}

	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)

	actualApp := graphql.ApplicationExt{}
	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  			result: createApplication(in: %s) { 
					%s 
				}
			}`,
			appInputGQL,
			tc.gqlFieldsProvider.ForApplication(),
		))

	saveExampleInCustomDir(t, request.Query(), createApplicationCategory, "create application with event APIs")
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithDocuments(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationCreateInput{
		Name: "create-application-with-documents",
		Documents: []*graphql.DocumentInput{
			{
				Title:       "Readme",
				Description: "Detailed description of project",
				Format:      graphql.DocumentFormatMarkdown,
				DisplayName: "display-name",
				FetchRequest: &graphql.FetchRequestInput{
					URL:    "kyma-project.io",
					Mode:   ptr.FetchMode(graphql.FetchModePackage),
					Filter: ptr.String("/docs/README.md"),
					Auth:   fixBasicAuth(),
				},
			},
			{
				Title:       "Troubleshooting",
				Description: "Troubleshooting description",
				Format:      graphql.DocumentFormatMarkdown,
				DisplayName: "display-name",
				Data:        ptr.CLOB(graphql.CLOB("No problems, everything works on my machine")),
			},
		},
		Labels: &graphql.Labels{
			"scenarios": []interface{}{"DEFAULT"},
		},
	}
	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
				result: createApplication(in: %s) { 
						%s 
					}
				}`,
			appInputGQL,
			tc.gqlFieldsProvider.ForApplication(),
		))

	saveExampleInCustomDir(t, request.Query(), createApplicationCategory, "create application with documents")
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithNonExistentIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	in := fixSampleApplicationCreateInputWithIntegrationSystem("placeholder")
	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	request := fixCreateApplicationRequest(appInputGQL)
	// WHEN
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.Error(t, err)
	require.NotNil(t, err.Error())
	require.Contains(t, err.Error(), "does not exist")
}

func TestAddDependentObjectsWhenAppDoesNotExist(t *testing.T) {
	applicationId := "cf889c38-490d-4896-96a7-c0721eca9932"

	t.Run("add Webhook", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		webhookInStr, err := tc.graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL:  "http://new.webhook",
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})
		require.NoError(t, err)

		//WHEN
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addWebhook(applicationID: "%s", in: %s) {
					%s
				}
			}`, applicationId, webhookInStr, tc.gqlFieldsProvider.ForWebhooks()))
		err = tc.RunOperation(ctx, addReq, nil)

		//THEN
		require.EqualError(t, err, "graphql: Cannot add Webhook to not existing Application")
	})

	t.Run("add API", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		apiInStr, err := tc.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{
			Name:      "new-api-name",
			TargetURL: "new-api-url",
		})
		require.NoError(t, err)

		// WHEN
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addAPI(applicationID: "%s", in: %s) {
					%s
				}
			}`, applicationId, apiInStr, tc.gqlFieldsProvider.ForAPIDefinition()))

		err = tc.RunOperation(ctx, addReq, nil)

		//THEN
		require.EqualError(t, err, "graphql: Cannot add API to not existing Application")
	})

	t.Run("add Event API", func(t *testing.T) {
		// GIVEN
		ctx := context.Background()
		eventApiInStr, err := tc.graphqlizer.EventAPIDefinitionInputToGQL(graphql.EventAPIDefinitionInput{
			Name: "new-event-api",
			Spec: &graphql.EventAPISpecInput{
				EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
				Format:        graphql.SpecFormatYaml,
			},
		})
		require.NoError(t, err)

		// WHEN
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: addEventAPI(applicationID: "%s", in: %s) {
						%s	
					}
				}`, applicationId, eventApiInStr, tc.gqlFieldsProvider.ForEventAPI()))
		err = tc.RunOperation(ctx, addReq, nil)

		// THEN
		require.EqualError(t, err, "graphql: Cannot add EventAPI to not existing Application")
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
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: addDocument(applicationID: "%s", in: %s) {
						%s
					}
			}`, applicationId, documentInStr, tc.gqlFieldsProvider.ForDocument()))
		err = tc.RunOperation(ctx, addReq, nil)

		//THEN
		require.EqualError(t, err, "graphql: Cannot add Document to not existing Application")
	})
}

func TestUpdateApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	actualApp := createApplication(t, ctx, "before")
	defer deleteApplication(t, actualApp.ID)

	expectedApp := actualApp
	expectedApp.Name = "after"
	expectedApp.Description = ptr.String("after")
	expectedApp.HealthCheckURL = ptr.String("https://kyma-project.io")

	updateInput := fixSampleApplicationUpdateInput("after")
	updateInputGQL, err := tc.graphqlizer.ApplicationUpdateInputToGQL(updateInput)
	require.NoError(t, err)
	request := fixUpdateApplicationRequest(actualApp.ID, updateInputGQL)
	updatedApp := graphql.ApplicationExt{}

	//WHEN
	err = tc.RunOperation(ctx, request, &updatedApp)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, expectedApp, updatedApp)
	saveExample(t, request.Query(), "update application")
}

func TestUpdateApplicationWithNonExistentIntegrationSystem(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	actualApp := createApplication(t, ctx, "before")
	defer deleteApplication(t, actualApp.ID)

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

func TestCreateUpdateApplicationWithDuplicatedNamesWithinTenant(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	appName := "samename"

	actualApp := createApplication(t, ctx, appName)
	defer deleteApplication(t, actualApp.ID)

	t.Run("Error when creating second Application with same name", func(t *testing.T) {
		in := fixSampleApplicationCreateInputWithName("first", appName)
		appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
		require.NoError(t, err)
		request := fixCreateApplicationRequest(appInputGQL)

		// WHEN
		err = tc.RunOperation(ctx, request, nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not unique")
	})

	t.Run("Error when updating Application with name that exists", func(t *testing.T) {
		actualApp := createApplication(t, ctx, "differentname")
		defer deleteApplication(t, actualApp.ID)

		updateInput := fixSampleApplicationUpdateInput(appName)
		updateInputGQL, err := tc.graphqlizer.ApplicationUpdateInputToGQL(updateInput)
		require.NoError(t, err)
		request := fixUpdateApplicationRequest(actualApp.ID, updateInputGQL)

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
	in := fixSampleApplicationCreateInput("app")

	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: createApplication(in: %s) {
    					id
					}
				}`, appInputGQL))
	actualApp := graphql.ApplicationExt{}
	err = tc.RunOperation(ctx, createReq, &actualApp)
	require.NoError(t, err)

	require.NotEmpty(t, actualApp.ID)

	// WHEN
	delReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplication(id: "%s") {
					id
				}
			}`, actualApp.ID))
	saveExample(t, delReq.Query(), "delete application")
	err = tc.RunOperation(ctx, delReq, &actualApp)

	//THEN
	require.NoError(t, err)
}

func TestUpdateApplicationParts(t *testing.T) {
	ctx := context.Background()
	placeholder := "app"
	in := fixSampleApplicationCreateInput(placeholder)

	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: createApplication(in: %s) {
    					id
					}
				}`, appInputGQL))
	actualApp := graphql.ApplicationExt{}
	err = tc.RunOperation(ctx, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)

	t.Run("labels manipulation", func(t *testing.T) {
		expectedLabel := graphql.Label{Key: "brand-new-label", Value: []interface{}{"aaa", "bbb"}}

		// add label
		createdLabel := &graphql.Label{}

		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: setApplicationLabel(applicationID: "%s", key: "%s", value: %s) {
					key 
					value
				}
			}`, actualApp.ID, expectedLabel.Key, "[\"aaa\",\"bbb\"]"))
		saveExample(t, addReq.Query(), "set application label")
		err := tc.RunOperation(ctx, addReq, &createdLabel)
		require.NoError(t, err)
		assert.Equal(t, &expectedLabel, createdLabel)
		actualApp := getApp(ctx, t, actualApp.ID)
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "aaa")
		assert.Contains(t, actualApp.Labels[expectedLabel.Key], "bbb")

		// delete label value
		deletedLabel := graphql.Label{}
		delReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: deleteApplicationLabel(applicationID: "%s", key: "%s") {
					key 
					value
				}
			}`, actualApp.ID, expectedLabel.Key))
		saveExample(t, delReq.Query(), "delete application label")
		err = tc.RunOperation(ctx, delReq, &deletedLabel)
		require.NoError(t, err)
		assert.Equal(t, expectedLabel, deletedLabel)
		actualApp = getApp(ctx, t, actualApp.ID)
		assert.Nil(t, actualApp.Labels[expectedLabel.Key])

	})

	t.Run("manage webhooks", func(t *testing.T) {
		// add
		webhookInStr, err := tc.graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL:  "http://new-webhook.url",
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addWebhook(applicationID: "%s", in: %s) {
					%s
				}
			}`, actualApp.ID, webhookInStr, tc.gqlFieldsProvider.ForWebhooks()))
		saveExampleInCustomDir(t, addReq.Query(), addWebhookCategory, "add application webhook")

		actualWebhook := graphql.Webhook{}
		err = tc.RunOperation(ctx, addReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "http://new-webhook.url", actualWebhook.URL)
		assert.Equal(t, graphql.ApplicationWebhookTypeConfigurationChanged, actualWebhook.Type)
		id := actualWebhook.ID
		require.NotNil(t, id)

		// get all webhooks
		updatedApp := getApp(ctx, t, actualApp.ID)
		assert.Len(t, updatedApp.Webhooks, 2)

		// update
		webhookInStr, err = tc.graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL: "http://updated-webhook.url", Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		updateReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: updateWebhook(webhookID: "%s", in: %s) {
					%s
				}
			}`, actualWebhook.ID, webhookInStr, tc.gqlFieldsProvider.ForWebhooks()))
		saveExampleInCustomDir(t, updateReq.Query(), updateWebhookCategory, "update application webhook")
		err = tc.RunOperation(ctx, updateReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "http://updated-webhook.url", actualWebhook.URL)

		// delete

		//GIVEN
		deleteReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: deleteWebhook(webhookID: "%s") {
					%s
				}
			}`, actualWebhook.ID, tc.gqlFieldsProvider.ForWebhooks()))
		saveExampleInCustomDir(t, deleteReq.Query(), deleteWebhookCategory, "delete application webhook")

		//WHEN
		err = tc.RunOperation(ctx, deleteReq, &actualWebhook)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, "http://updated-webhook.url", actualWebhook.URL)

	})

	t.Run("manage APIs", func(t *testing.T) {
		// add
		inStr, err := tc.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{
			Name:      "new-api-name",
			TargetURL: "new-api-url",
			Spec: &graphql.APISpecInput{
				Format: graphql.SpecFormatJSON,
				Type:   graphql.APISpecTypeOpenAPI,
				FetchRequest: &graphql.FetchRequestInput{
					URL: "foo.bar",
				},
			},
		})

		require.NoError(t, err)
		actualAPI := graphql.APIDefinition{}

		// WHEN
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addAPI(applicationID: "%s", in: %s) {
					%s
				}
			}`, actualApp.ID, inStr, tc.gqlFieldsProvider.ForAPIDefinition()))
		saveExample(t, addReq.Query(), "add API")
		err = tc.RunOperation(ctx, addReq, &actualAPI)

		//THEN
		require.NoError(t, err)
		id := actualAPI.ID
		require.NotNil(t, id)
		assert.Equal(t, "new-api-name", actualAPI.Name)
		assert.Equal(t, "new-api-url", actualAPI.TargetURL)

		updatedApp := getApp(ctx, t, actualApp.ID)
		assert.Len(t, updatedApp.Apis.Data, 2)
		actualAPINames := make(map[string]struct{})
		for _, api := range updatedApp.Apis.Data {
			actualAPINames[api.Name] = struct{}{}
		}
		assert.Contains(t, actualAPINames, "new-api-name")
		assert.Contains(t, actualAPINames, placeholder)

		// update

		//GIVEN
		updateStr, err := tc.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{Name: "updated-api-name", TargetURL: "updated-api-url"})
		require.NoError(t, err)
		updatedAPI := graphql.APIDefinition{}

		// WHEN
		updateReq := gcli.NewRequest(
			fmt.Sprintf(`mutation { 
				result: updateAPI(id: "%s", in: %s) {
						%s
					}
				}`, id, updateStr, tc.gqlFieldsProvider.ForAPIDefinition()))
		err = tc.RunOperation(ctx, updateReq, &updatedAPI)
		saveExample(t, updateReq.Query(), "update API")

		//THEN
		require.NoError(t, err)
		updatedApp = getApp(ctx, t, actualApp.ID)
		assert.Len(t, updatedApp.Apis.Data, 2)
		actualAPINamesAfterUpdate := make(map[string]struct{})
		for _, api := range updatedApp.Apis.Data {
			actualAPINamesAfterUpdate[api.Name] = struct{}{}
		}
		assert.Contains(t, actualAPINamesAfterUpdate, "updated-api-name")
		assert.Contains(t, actualAPINamesAfterUpdate, placeholder)
		// delete
		delAPI := graphql.APIDefinition{}

		// WHEN
		deleteReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: deleteAPI(id: "%s") {
						id
					}
				}`, id))
		err = tc.RunOperation(ctx, deleteReq, &delAPI)
		saveExample(t, deleteReq.Query(), "delete API")

		//THEN
		require.NoError(t, err)
		assert.Equal(t, id, delAPI.ID)

		app := getApp(ctx, t, actualApp.ID)
		require.Len(t, app.Apis.Data, 1)
		assert.Equal(t, placeholder, app.Apis.Data[0].Name)

	})

	t.Run("manage event api", func(t *testing.T) {
		// add

		// GIVEN
		inStr, err := tc.graphqlizer.EventAPIDefinitionInputToGQL(graphql.EventAPIDefinitionInput{
			Name: "new-event-api",
			Spec: &graphql.EventAPISpecInput{
				EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
				Format:        graphql.SpecFormatYaml,
				FetchRequest: &graphql.FetchRequestInput{
					URL: "foo.bar",
				},
			},
		})

		actualEventAPI := graphql.EventAPIDefinition{}
		require.NoError(t, err)

		// WHEN
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: addEventAPI(applicationID: "%s", in: %s) {
						%s	
					}
				}`, actualApp.ID, inStr, tc.gqlFieldsProvider.ForEventAPI()))
		err = tc.RunOperation(ctx, addReq, &actualEventAPI)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "new-event-api", actualEventAPI.Name)
		assert.NotEmpty(t, actualEventAPI.ID)
		updatedApp := getApp(ctx, t, actualApp.ID)
		assert.Len(t, updatedApp.EventAPIs.Data, 2)

		// update

		// GIVEN
		updateStr, err := tc.graphqlizer.EventAPIDefinitionInputToGQL(graphql.EventAPIDefinitionInput{
			Name: "updated-event-api",
			Spec: &graphql.EventAPISpecInput{
				EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
				Format:        graphql.SpecFormatYaml,
			}})
		require.NoError(t, err)

		// WHEN
		updateReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: updateEventAPI(id: "%s", in: %s) {
						%s
					}
				}`, actualEventAPI.ID, updateStr, tc.gqlFieldsProvider.ForEventAPI()))
		err = tc.RunOperation(ctx, updateReq, &actualEventAPI)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, "updated-event-api", actualEventAPI.Name)

		// delete
		// WHEN
		delReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: deleteEventAPI(id: "%s") {
					id
				}
			}`, actualEventAPI.ID))
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
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: addDocument(applicationID: "%s", in: %s) {
						%s
					}
			}`, actualApp.ID, inStr, tc.gqlFieldsProvider.ForDocument()))
		err = tc.RunOperation(ctx, addReq, &actualDoc)
		saveExample(t, addReq.Query(), "add Document")

		//THEN
		require.NoError(t, err)
		id := actualDoc.ID
		require.NotNil(t, id)
		assert.Equal(t, "new-document", actualDoc.Title)

		//delete

		updatedApp := getApp(ctx, t, actualApp.ID)
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
		deleteReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: deleteDocument(id: "%s") {
						id
					}
				}`, id))
		err = tc.RunOperation(ctx, deleteReq, &delDocument)
		saveExample(t, deleteReq.Query(), "delete Document")

		//THEN
		require.NoError(t, err)
		assert.Equal(t, id, delDocument.ID)

		app := getApp(ctx, t, actualApp.ID)
		require.Len(t, app.Documents.Data, 1)
		assert.Equal(t, placeholder, app.Documents.Data[0].Title)
	})

	t.Run("refetch API", func(t *testing.T) {
		// TODO later
	})

	t.Run("refetch Event API", func(t *testing.T) {
		// TODO later
	})
}

func TestQueryApplications(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		in := graphql.ApplicationCreateInput{
			Name: fmt.Sprintf("app-%d", i),
		}

		appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
		require.NoError(t, err)
		actualApp := graphql.Application{}
		request := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {
					%s
				}
			}`, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
		err = tc.RunOperation(ctx, request, &actualApp)
		require.NoError(t, err)
		defer deleteApplication(t, actualApp.ID)
	}
	actualAppPage := graphql.ApplicationPage{}

	// WHEN
	queryReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: applications {
					%s
				}
			}`, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
	err := tc.RunOperation(ctx, queryReq, &actualAppPage)
	saveExampleInCustomDir(t, queryReq.Query(), queryApplicationsCategory, "query applications")

	//THEN
	require.NoError(t, err)
	assert.Len(t, actualAppPage.Data, 3)
	assert.Equal(t, 3, actualAppPage.TotalCount)

}

func TestQuerySpecificApplication(t *testing.T) {
	// GIVEN
	in := graphql.ApplicationCreateInput{
		Name: fmt.Sprintf("app"),
	}

	appInputGQL, err := tc.graphqlizer.ApplicationCreateInputToGQL(in)
	require.NoError(t, err)

	actualApp := graphql.Application{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {
					%s
				}
			}`, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	err = tc.RunOperation(context.Background(), request, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	createdID := actualApp.ID
	defer deleteApplication(t, actualApp.ID)

	// WHEN
	queryAppReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					%s
				}
			}`, actualApp.ID, tc.gqlFieldsProvider.ForApplication()))
	err = tc.RunOperation(context.Background(), queryAppReq, &actualApp)
	saveExampleInCustomDir(t, queryAppReq.Query(), queryApplicationCategory, "query application")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdID, actualApp.ID)
}

func TestTenantSeparation(t *testing.T) {
	// GIVEN
	appIn := fixSampleApplicationCreateInput("adidas")
	inStr, err := tc.graphqlizer.ApplicationCreateInputToGQL(appIn)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createApplication(in: %s) {
						%s
					}
				}`,
			inStr, tc.gqlFieldsProvider.ForApplication()))
	actualApp := graphql.ApplicationExt{}
	ctx := context.Background()
	err = tc.RunOperation(ctx, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)

	// WHEN
	getAppReq := gcli.NewRequest(fmt.Sprintf(`query {
			result: applications {
				%s
			}
		}`,
		tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
	customTenant := "4204255f-1262-47d7-9108-fbdd7a8d1096"
	anotherTenantsApps := graphql.ApplicationPage{}
	// THEN
	err = tc.RunOperationWithCustomTenant(ctx, customTenant, getAppReq, &anotherTenantsApps)
	require.NoError(t, err)
	assert.Empty(t, anotherTenantsApps.Data)
}

func TestQueryAPIRuntimeAuths(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	defaultAuth := fixBasicAuth()
	customAuth := graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "custom",
				Password: "auth",
			}},
		AdditionalHeaders: &graphql.HttpHeaders{
			"customHeader": []string{"custom", "custom"},
		},
	}

	exampleSaved := false
	rtmsToCreate := 3

	testCases := []struct {
		Name string
		Apis []*graphql.APIDefinitionInput
	}{
		{
			Name: "API without default auth",
			Apis: []*graphql.APIDefinitionInput{
				{
					Name:      "without-default-auth",
					TargetURL: "http://mywordpress.com/comments",
				},
			},
		},
		{
			Name: "API with default auth",
			Apis: []*graphql.APIDefinitionInput{
				{
					Name:        "with-default-auth",
					TargetURL:   "http://mywordpress.com/comments",
					DefaultAuth: defaultAuth,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appInput := graphql.ApplicationCreateInput{
				Name: "test-app",
				Apis: testCase.Apis,
				Labels: &graphql.Labels{
					"scenarios": []interface{}{"DEFAULT"},
				},
			}

			app := createApplicationFromInputWithinTenant(t, ctx, appInput, defaultTenant)
			defer deleteApplication(t, app.ID)

			var rtmIDs []string
			for i := 0; i < rtmsToCreate; i++ {
				rtm := createRuntime(t, ctx, fmt.Sprintf("test-rtm-%d", i))
				rtmIDs = append(rtmIDs, rtm.ID)
				defer deleteRuntime(t, rtm.ID)
			}
			require.Len(t, rtmIDs, rtmsToCreate)

			rtmIDWithUnsetAPIRtmAuth := rtmIDs[0]
			rtmIDWithSetAPIRtmAuth := rtmIDs[1]

			setAPIAuth(t, ctx, app.Apis.Data[0].ID, rtmIDWithSetAPIRtmAuth, customAuth)
			defer deleteAPIAuth(t, ctx, app.Apis.Data[0].ID, rtmIDWithSetAPIRtmAuth)

			innerTestCases := []struct {
				Name         string
				QueriedRtmID string
			}{
				{
					Name:         "Query set API Runtime Auth",
					QueriedRtmID: rtmIDWithSetAPIRtmAuth,
				},
				{
					Name:         "Query unset API Runtime Auth",
					QueriedRtmID: rtmIDWithUnsetAPIRtmAuth,
				},
			}

			for _, innerTestCase := range innerTestCases {
				t.Run(innerTestCase.Name, func(t *testing.T) {
					result := graphql.ApplicationExt{}
					request := fixAPIRuntimeAuthRequest(app.ID, innerTestCase.QueriedRtmID)

					// WHEN
					err := tc.RunOperation(ctx, request, &result)

					// THEN
					require.NoError(t, err)
					require.NotEmpty(t, result.ID)
					assertApplication(t, appInput, result)
					require.Len(t, result.Apis.Data, 1)

					assert.Equal(t, len(rtmIDs), len(result.Apis.Data[0].Auths))
					assert.Equal(t, innerTestCase.QueriedRtmID, result.Apis.Data[0].Auth.RuntimeID)
					if innerTestCase.QueriedRtmID == rtmIDWithSetAPIRtmAuth {
						assertAuth(t, &customAuth, result.Apis.Data[0].Auth.Auth)
					} else {
						assert.Equal(t, result.Apis.Data[0].DefaultAuth, result.Apis.Data[0].Auth.Auth)
					}

					for _, auth := range result.Apis.Data[0].Auths {
						if auth.RuntimeID == rtmIDWithSetAPIRtmAuth {
							assertAuth(t, &customAuth, auth.Auth)
						} else {
							assert.Equal(t, result.Apis.Data[0].DefaultAuth, auth.Auth)
						}
					}

					if !exampleSaved {
						saveExampleInCustomDir(t, request.Query(), queryApplicationCategory, "query api runtime auths")
						exampleSaved = true
					}
				})
			}
		})
	}
}

func TestQuerySpecificAPIDefinition(t *testing.T) {
	// GIVEN
	in := graphql.APIDefinitionInput{
		Name:      "test",
		TargetURL: "test",
	}

	APIInputGQL, err := tc.graphqlizer.APIDefinitionInputToGQL(in)
	require.NoError(t, err)
	applicationID := createApplication(t, context.Background(), "test").ID
	defer deleteApplication(t, applicationID)
	actualAPI := graphql.APIDefinition{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addAPI(applicationID: "%s", in: %s) {
					%s
				}
			}`, applicationID, APIInputGQL, tc.gqlFieldsProvider.ForAPIDefinition()))
	err = tc.RunOperation(context.TODO(), request, &actualAPI)
	require.NoError(t, err)
	require.NotEmpty(t, actualAPI.ID)
	createdID := actualAPI.ID
	defer deleteAPI(t, createdID)

	// WHEN
	queryAppReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					api(id: "%s"){
						%s
					}
				}
			}`, applicationID, actualAPI.ID, tc.gqlFieldsProvider.ForAPIDefinition()))
	err = tc.RunOperation(context.Background(), queryAppReq, &actualAPI)
	saveExample(t, queryAppReq.Query(), "query api")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdID, actualAPI.ID)
}

func TestQuerySpecificEventAPIDefinition(t *testing.T) {
	// GIVEN
	in := graphql.EventAPIDefinitionInput{
		Name: "test",
		Spec: &graphql.EventAPISpecInput{
			EventSpecType: "ASYNC_API",
			Format:        "YAML",
		},
	}
	EventAPIInputGQL, err := tc.graphqlizer.EventAPIDefinitionInputToGQL(in)
	require.NoError(t, err)
	applicationID := createApplication(t, context.Background(), "test").ID
	defer deleteApplication(t, applicationID)
	actualEventAPI := graphql.EventAPIDefinition{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addEventAPI(applicationID: "%s", in: %s) {
					%s
				}
			}`, applicationID, EventAPIInputGQL, tc.gqlFieldsProvider.ForEventAPI()))
	err = tc.RunOperation(context.TODO(), request, &actualEventAPI)
	require.NoError(t, err)
	require.NotEmpty(t, actualEventAPI.ID)
	createdID := actualEventAPI.ID
	defer deleteEventAPI(t, createdID)

	// WHEN
	queryAppReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					eventAPI(id: "%s"){
						%s
					}
				}
			}`, applicationID, actualEventAPI.ID, tc.gqlFieldsProvider.ForEventAPI()))
	err = tc.RunOperation(context.Background(), queryAppReq, &actualEventAPI)
	saveExample(t, queryAppReq.Query(), "query event api")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdID, actualEventAPI.ID)
}

func getApp(ctx context.Context, t *testing.T, id string) graphql.ApplicationExt {
	q := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
			} 
		}`, id, tc.gqlFieldsProvider.ForApplication()))
	var app graphql.ApplicationExt
	require.NoError(t, tc.RunOperation(ctx, q, &app))
	return app

}

func fixSampleApplicationCreateInput(placeholder string) graphql.ApplicationCreateInput {
	return graphql.ApplicationCreateInput{
		Name: placeholder,
		Documents: []*graphql.DocumentInput{{
			Title:       placeholder,
			DisplayName: placeholder,
			Description: placeholder,
			Format:      graphql.DocumentFormatMarkdown}},
		Apis: []*graphql.APIDefinitionInput{{
			Name:      placeholder,
			TargetURL: placeholder}},
		EventAPIs: []*graphql.EventAPIDefinitionInput{{
			Name: placeholder,
			Spec: &graphql.EventAPISpecInput{
				EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
				Format:        graphql.SpecFormatYaml,
			}}},
		Webhooks: []*graphql.WebhookInput{{
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
			URL:  webhookURL},
		},
		Labels: &graphql.Labels{placeholder: []interface{}{placeholder}},
	}
}

func fixSampleApplicationCreateInputWithName(placeholder, name string) graphql.ApplicationCreateInput {
	sampleInput := fixSampleApplicationCreateInput(placeholder)
	sampleInput.Name = name
	return sampleInput
}

func fixSampleApplicationCreateInputWithIntegrationSystem(placeholder string) graphql.ApplicationCreateInput {
	sampleInput := fixSampleApplicationCreateInput(placeholder)
	sampleInput.IntegrationSystemID = &integrationSystemID
	return sampleInput
}

func fixSampleApplicationUpdateInput(placeholder string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Name:           placeholder,
		Description:    &placeholder,
		HealthCheckURL: ptr.String(webhookURL),
	}
}

func fixSampleApplicationUpdateInputWithIntegrationSystem(placeholder string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Name:                placeholder,
		Description:         &placeholder,
		HealthCheckURL:      ptr.String(webhookURL),
		IntegrationSystemID: &integrationSystemID,
	}
}

func deleteApplicationInTenant(t *testing.T, id string, tenant string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteApplication(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperationWithCustomTenant(context.Background(), tenant, req, nil))
}

func deleteApplication(t *testing.T, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteApplication(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}

func deleteAPI(t *testing.T, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteAPI(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}

func deleteEventAPI(t *testing.T, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteEventAPI(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}
