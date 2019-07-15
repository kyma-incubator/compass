package director

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateApplicationWithAllSimpleFieldsProvided(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationInput{
		Name:           "wordpress",
		Description:    ptrString("my first wordpress application"),
		HealthCheckURL: ptrString("http://mywordpress.com/health"),
		Labels: &graphql.Labels{
			"group": []string{"production", "experimental"},
		},
	}

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	actualApp := ApplicationExt{}

	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {
					%s
				}
			}`,
			appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	err = tc.RunQuery(ctx, request, &actualApp)

	//THEN
	saveQueryInExamples(t, request.Query(), "create application")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationInput{
		Name: "wordpress",
		Webhooks: []*graphql.WebhookInput{
			{
				Type: graphql.ApplicationWebhookTypeConfigurationChanged,
				Auth: fixBasicAuth(),
				URL:  "http://mywordpress.com/webhooks1",
			},
		},
	}

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	actualApp := ApplicationExt{}

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
	saveQueryInExamples(t, request.Query(), "create application with webhooks")
	err = tc.RunQuery(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithAPIs(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationInput{
		Name: "wordpress",
		Apis: []*graphql.APIDefinitionInput{
			{
				Name:        "comments/v1",
				Description: ptrString("api for adding comments"),
				TargetURL:   "http://mywordpress.com/comments",
				Group:       ptrString("comments"),
				DefaultAuth: fixBasicAuth(),
				Version:     fixDepracatedVersion1(),
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOpenAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptrCLOB(graphql.CLOB("openapi")),
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
						Mode:   ptrFetchMode(graphql.FetchModePackage),
						Filter: ptrString("odata.json"),
						Auth:   fixBasicAuth(),
					},
				},
			},
		}}

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	actualApp := ApplicationExt{}

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
	saveQueryInExamples(t, request.Query(), "create application with APIs")

	err = tc.RunQuery(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithEventAPIs(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationInput{
		Name: "wordpress",
		EventAPIs: []*graphql.EventAPIDefinitionInput{
			{
				Name:        "comments/v1",
				Description: ptrString("comments events"),
				Version:     fixDepracatedVersion1(),
				Group:       ptrString("comments"),
				Spec: &graphql.EventAPISpecInput{
					EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
					Format:        graphql.SpecFormatYaml,
					Data:          ptrCLOB(graphql.CLOB([]byte("asyncapi"))),
				},
			},
			{
				Name:        "reviews/v1",
				Description: ptrString("review events"),
				Spec: &graphql.EventAPISpecInput{
					EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
					Format:        graphql.SpecFormatYaml,
					FetchRequest: &graphql.FetchRequestInput{
						URL:    "http://mywordpress.com/events",
						Mode:   ptrFetchMode(graphql.FetchModePackage),
						Filter: ptrString("async.json"),
						Auth:   fixOauthAuth(),
					},
				},
			},
		},
	}

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)

	actualApp := ApplicationExt{}
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

	saveQueryInExamples(t, request.Query(), "create application with event APIs")
	err = tc.RunQuery(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestCreateApplicationWithDocuments(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationInput{
		Name: "wordpress",
		Documents: []*graphql.DocumentInput{
			{
				Title:       "Readme",
				Description: "Detailed description of project",
				Format:      graphql.DocumentFormatMarkdown,

				FetchRequest: &graphql.FetchRequestInput{
					URL:    "kyma-project.io",
					Mode:   ptrFetchMode(graphql.FetchModePackage),
					Filter: ptrString("/docs/README.md"),
					Auth:   fixBasicAuth(),
				},
			},
			{
				Title:       "Troubleshooting",
				Description: "Troubleshooting description",
				Format:      graphql.DocumentFormatMarkdown,
				Data:        ptrCLOB(graphql.CLOB("No problems, everything works on my machine")),
			},
		},
	}
	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	actualApp := ApplicationExt{}

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

	saveQueryInExamples(t, request.Query(), "create application with documents")
	err = tc.RunQuery(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestUpdateApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := generateSampleApplicationInput("before")
	in.Description = ptrString("before")

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)

	actualApp := ApplicationExt{}

	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createApplication(in: %s) {
    					id
					}
				}`, appInputGQL))
	err = tc.RunQuery(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	id := actualApp.ID
	require.NotEmpty(t, id)
	defer deleteApplication(t, id)
	in = generateSampleApplicationInput("after")

	appInputGQL, err = tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	request = gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: updateApplication(id: "%s", in: %s) {
    					%s
					}
				}`, id, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	saveQueryInExamples(t, request.Query(), "update application")

	updatedApp := ApplicationExt{}
	err = tc.RunQuery(ctx, request, &updatedApp)
	require.NoError(t, err)

	assertApplication(t, in, updatedApp)
}

func TestDeleteApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := generateSampleApplicationInput("app")

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: createApplication(in: %s) {
    					id
					}
				}`, appInputGQL))
	actualApp := ApplicationExt{}
	err = tc.RunQuery(ctx, createReq, &actualApp)
	require.NoError(t, err)

	require.NotEmpty(t, actualApp.ID)

	// WHEN
	delReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteApplication(id: "%s") {
					id
				}
			}`, actualApp.ID))
	saveQueryInExamples(t, delReq.Query(), "delete application")
	err = tc.RunQuery(ctx, delReq, &actualApp)

	//THEN
	require.NoError(t, err)
}

func TestUpdateApplicationParts(t *testing.T) {
	ctx := context.Background()
	placeholder := "app"
	in := generateSampleApplicationInput(placeholder)

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: createApplication(in: %s) {
    					id
					}
				}`, appInputGQL))
	actualApp := ApplicationExt{}
	err = tc.RunQuery(ctx, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)

	t.Run("labels manipulation", func(t *testing.T) {
		// add label
		createdLabel := &graphql.Label{}

		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addApplicationLabel(applicationID: "%s", key: "%s", values: %s) {
					key 
					values
				}
			}`, actualApp.ID, "brand-new-label", "[\"aaa\",\"bbb\"]"))
		saveQueryInExamples(t, addReq.Query(), "add application label")
		err := tc.RunQuery(ctx, addReq, &createdLabel)
		require.NoError(t, err)
		assert.Equal(t, &graphql.Label{Key: "brand-new-label", Values: []string{"aaa", "bbb"}}, createdLabel)
		actualApp := getApp(ctx, t, actualApp.ID)
		assert.Contains(t, actualApp.Labels["brand-new-label"], "aaa")
		assert.Contains(t, actualApp.Labels["brand-new-label"], "bbb")

		// delete first label value
		delReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: deleteApplicationLabel(applicationID: "%s", key: "%s", values: %s) {
						key 
						values
					}
				}`, actualApp.ID, "brand-new-label", "[\"aaa\"]"))
		saveQueryInExamples(t, delReq.Query(), "delete application label")
		deletedLabel := &graphql.Label{}
		err = tc.RunQuery(ctx, delReq, &deletedLabel)
		require.NoError(t, err)
		assert.Equal(t, &graphql.Label{Key: "brand-new-label", Values: []string{"bbb"}}, deletedLabel)
		actualApp = getApp(ctx, t, actualApp.ID)

		// delete second label value
		delReq = gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: deleteApplicationLabel(applicationID: "%s", key: "%s", values: %s) {
					key 
					values
				}
			}`, actualApp.ID, "brand-new-label", "[\"bbb\"]"))
		err = tc.RunQuery(ctx, delReq, &deletedLabel)
		require.NoError(t, err)
		assert.Equal(t, &graphql.Label{Key: "brand-new-label", Values: []string{}}, deletedLabel)
		actualApp = getApp(ctx, t, actualApp.ID)
		assert.Nil(t, actualApp.Labels["brand-new-label"])

	})

	t.Run("manage webhooks", func(t *testing.T) {
		// add
		webhookInStr, err := tc.graphqlizer.ApplicationWebhookInputToGQL(&graphql.WebhookInput{
			URL:  "new-webhook",
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addWebhook(applicationID: "%s", in: %s) {
					%s
				}
			}`, actualApp.ID, webhookInStr, tc.gqlFieldsProvider.ForWebhooks()))
		saveQueryInExamples(t, addReq.Query(), "add aplication webhook")

		actualWebhook := graphql.Webhook{}
		err = tc.RunQuery(ctx, addReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "new-webhook", actualWebhook.URL)
		assert.Equal(t, graphql.ApplicationWebhookTypeConfigurationChanged, actualWebhook.Type)
		id := actualWebhook.ID
		require.NotNil(t, id)

		// add to non exist application

		//GIVEN
		webhookInStr, err = tc.graphqlizer.ApplicationWebhookInputToGQL(&graphql.WebhookInput{
			URL:  "new-webhook",
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})
		require.NoError(t, err)

		//WHEN
		addReq = gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addWebhook(applicationID: "%s", in: %s) {
					%s
				}
			}`, "foo", webhookInStr, tc.gqlFieldsProvider.ForWebhooks()))
		saveQueryInExamples(t, addReq.Query(), "add aplication webhook")

		//THEN
		tmpWebhook := graphql.Webhook{}
		err = tc.RunQuery(ctx, addReq, &tmpWebhook)
		require.Error(t, err)

		// get all webhooks
		updatedApp := getApp(ctx, t, actualApp.ID)
		assert.Len(t, updatedApp.Webhooks, 2)

		// update
		webhookInStr, err = tc.graphqlizer.ApplicationWebhookInputToGQL(&graphql.WebhookInput{
			URL: "updated-webhook", Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		updateReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: updateWebhook(webhookID: "%s", in: %s) {
					%s
				}
			}`, actualWebhook.ID, webhookInStr, tc.gqlFieldsProvider.ForWebhooks()))
		saveQueryInExamples(t, updateReq.Query(), "update application webhook")
		err = tc.RunQuery(ctx, updateReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "updated-webhook", actualWebhook.URL)

		// delete

		//GIVEN
		deleteReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: deleteWebhook(webhookID: "%s") {
					%s
				}
			}`, actualWebhook.ID, tc.gqlFieldsProvider.ForWebhooks()))
		saveQueryInExamples(t, deleteReq.Query(), "delete application webhook")

		//WHEN
		err = tc.RunQuery(ctx, deleteReq, &actualWebhook)

		//THEN
		require.NoError(t, err)
		assert.Equal(t, "updated-webhook", actualWebhook.URL)

	})

	t.Run("manage APIs", func(t *testing.T) {
		// add
		inStr, err := tc.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{
			Name:      "new-api-name",
			TargetURL: "new-api-url",
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
		saveQueryInExamples(t, addReq.Query(), "add API")
		err = tc.RunQuery(ctx, addReq, &actualAPI)

		//THEN
		require.NoError(t, err)
		id := actualAPI.ID
		require.NotNil(t, id)
		assert.Equal(t, "new-api-name", actualAPI.Name)
		assert.Equal(t, "new-api-url", actualAPI.TargetURL)
		//
		updatedApp := getApp(ctx, t, actualApp.ID)
		assert.Len(t, updatedApp.Apis.Data, 2)
		actualAPINames := make(map[string]struct{})
		for _, api := range updatedApp.Apis.Data {
			actualAPINames[api.Name] = struct{}{}
		}
		assert.Contains(t, actualAPINames, "new-api-name")
		assert.Contains(t, actualAPINames, placeholder)

		// add to non exist application

		//GIVEN
		inStr, err = tc.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{
			Name:      "new-api-name",
			TargetURL: "new-api-url",
		})
		tmpAPI := graphql.APIDefinition{}
		require.NoError(t, err)

		// WHEN
		addReq = gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addAPI(applicationID: "%s", in: %s) {
					%s
				}
			}`, "foo", inStr, tc.gqlFieldsProvider.ForAPIDefinition()))
		saveQueryInExamples(t, addReq.Query(), "add API")

		err = tc.RunQuery(ctx, addReq, &tmpAPI)

		//THEN
		require.Error(t, err)

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
		err = tc.RunQuery(ctx, updateReq, &updatedAPI)
		saveQueryInExamples(t, updateReq.Query(), "update API")

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
		err = tc.RunQuery(ctx, deleteReq, &delAPI)
		saveQueryInExamples(t, deleteReq.Query(), "delete API")

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
		err = tc.RunQuery(ctx, addReq, &actualEventAPI)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "new-event-api", actualEventAPI.Name)
		assert.NotEmpty(t, actualEventAPI.ID)
		updatedApp := getApp(ctx, t, actualApp.ID)
		assert.Len(t, updatedApp.EventAPIs.Data, 2)

		// add to non exist application

		// GIVEN
		inStr, err = tc.graphqlizer.EventAPIDefinitionInputToGQL(graphql.EventAPIDefinitionInput{
			Name: "new-event-api",
			Spec: &graphql.EventAPISpecInput{
				EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
				Format:        graphql.SpecFormatYaml,
			},
		})
		tmpEventAPI := graphql.EventAPIDefinition{}
		require.NoError(t, err)

		// WHEN
		addReq = gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: addEventAPI(applicationID: "%s", in: %s) {
						%s	
					}
				}`, "foo", inStr, tc.gqlFieldsProvider.ForEventAPI()))
		err = tc.RunQuery(ctx, addReq, &tmpEventAPI)

		// THEN
		require.Error(t, err)

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
		err = tc.RunQuery(ctx, updateReq, &actualEventAPI)

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
		err = tc.RunQuery(ctx, delReq, nil)
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
		err = tc.RunQuery(ctx, addReq, &actualDoc)
		saveQueryInExamples(t, addReq.Query(), "add Document")

		//THEN
		require.NoError(t, err)
		id := actualDoc.ID
		require.NotNil(t, id)
		assert.Equal(t, "new-document", actualDoc.Title)

		// add to non exists application

		//GIVEN
		inStr, err = tc.graphqlizer.DocumentInputToGQL(&graphql.DocumentInput{
			Title:       "new-document",
			Format:      graphql.DocumentFormatMarkdown,
			DisplayName: "new-document-display-name",
			Description: "new-description",
		})

		require.NoError(t, err)
		tmpDoc := graphql.Document{}

		// WHEN
		addReq = gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: addDocument(applicationID: "%s", in: %s) {
						%s
					}
			}`, "foo", inStr, tc.gqlFieldsProvider.ForDocument()))
		err = tc.RunQuery(ctx, addReq, &tmpDoc)
		saveQueryInExamples(t, addReq.Query(), "add Document")

		//THEN
		require.Error(t, err)

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
		err = tc.RunQuery(ctx, deleteReq, &delDocument)
		saveQueryInExamples(t, deleteReq.Query(), "delete Document")

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
	idsToRemove := make([]string, 3)
	defer func() {
		for _, id := range idsToRemove {
			if id != "" {
				deleteApplication(t, id)
			}
		}
	}()
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		in := graphql.ApplicationInput{
			Name: fmt.Sprintf("app-%d", i),
		}

		appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
		require.NoError(t, err)
		actualApp := graphql.Application{}
		request := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {
					%s
				}
			}`, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
		err = tc.RunQuery(ctx, request, &actualApp)
		require.NoError(t, err)
		idsToRemove[i] = actualApp.ID
	}
	actualAppPage := graphql.ApplicationPage{}

	// WHEN
	queryReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: applications {
					%s
				}
			}`, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForApplication())))
	err := tc.RunQuery(ctx, queryReq, &actualAppPage)
	saveQueryInExamples(t, queryReq.Query(), "query applications")

	//THEN
	require.NoError(t, err)
	assert.Len(t, actualAppPage.Data, 3)
	assert.Equal(t, 3, actualAppPage.TotalCount)

}

func TestQuerySpecificApplication(t *testing.T) {
	// GIVEN
	in := graphql.ApplicationInput{
		Name: fmt.Sprintf("app"),
	}

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)

	actualApp := graphql.Application{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {
					%s
				}
			}`, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	err = tc.RunQuery(context.Background(), request, &actualApp)
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
	err = tc.RunQuery(context.Background(), queryAppReq, &actualApp)
	saveQueryInExamples(t, queryAppReq.Query(), "query specific application")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdID, actualApp.ID)
}

func TestTenantSeparation(t *testing.T) {
	// GIVEN
	appIn := generateSampleApplicationInput("adidas")
	inStr, err := tc.graphqlizer.ApplicationInputToGQL(appIn)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createApplication(in: %s) {
						%s
					}
				}`,
			inStr, tc.gqlFieldsProvider.ForApplication()))
	actualApp := ApplicationExt{}
	ctx := context.Background()
	err = tc.RunQuery(ctx, createReq, &actualApp)
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
	getAppReq.Header["Tenant"] = []string{"completely-another-tenant"}
	anotherTenantsApps := graphql.ApplicationPage{}
	// THEN
	err = tc.RunQuery(ctx, getAppReq, &anotherTenantsApps)
	require.NoError(t, err)
	assert.Empty(t, anotherTenantsApps.Data)

}

func getApp(ctx context.Context, t *testing.T, id string) ApplicationExt {
	q := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
			} 
		}`, id, tc.gqlFieldsProvider.ForApplication()))
	var app ApplicationExt
	require.NoError(t, tc.RunQuery(ctx, q, &app))
	return app

}

func generateSampleApplicationInput(placeholder string) graphql.ApplicationInput {
	return graphql.ApplicationInput{
		Name: placeholder,
		Documents: []*graphql.DocumentInput{{
			Title:  placeholder,
			Format: graphql.DocumentFormatMarkdown}},
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
			URL:  placeholder},
		},
		Labels: &graphql.Labels{placeholder: []string{placeholder}},
	}
}

func deleteApplication(t *testing.T, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
		deleteApplication(id: "%s") {
			id
		}	
	}`, id))
	require.NoError(t, tc.RunQuery(context.Background(), req, nil))
}
