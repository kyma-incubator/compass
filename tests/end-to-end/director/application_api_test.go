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
		ProviderName:   "provider name",
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
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`,
			appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application")
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestRegisterApplicationWithWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationRegisterInput{
		Name:         "wordpress",
		ProviderName: "compass",
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
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
				result: registerApplication(in: %s) { 
						%s 
					} 
				}`,
			appInputGQL,
			tc.gqlFieldsProvider.ForApplication(),
		))
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with webhooks")
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestRegisterApplicationWithAPIs(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationRegisterInput{
		Name:         "wordpress",
		ProviderName: "compass",
		APIDefinitions: []*graphql.APIDefinitionInput{
			{
				Name:        "comments-v1",
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
				Name:      "reviews-v1",
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
							TokenEndpointURL: "http://token.URL",
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

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)

	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
 			 result: registerApplication(in: %s) { 
					%s 
				}
			}`,
			appInputGQL,
			tc.gqlFieldsProvider.ForApplication(),
		))
	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with API definitions")

	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestRegisterApplicationWithEventDefinitions(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationRegisterInput{
		Name:         "create-application-with-event-apis",
		ProviderName: "compass",
		EventDefinitions: []*graphql.EventDefinitionInput{
			{
				Name:        "comments-v1",
				Description: ptr.String("comments events"),
				Version:     fixDepracatedVersion1(),
				Group:       ptr.String("comments"),
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptr.CLOB(graphql.CLOB([]byte("asyncapi"))),
				},
			},
			{
				Name:        "reviews-v1",
				Description: ptr.String("review events"),
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatYaml,
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

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	actualApp := graphql.ApplicationExt{}
	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  			result: registerApplication(in: %s) { 
					%s 
				}
			}`,
			appInputGQL,
			tc.gqlFieldsProvider.ForApplication(),
		))

	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with event definitions")
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)
	assertApplication(t, in, actualApp)
}

func TestRegisterApplicationWithDocuments(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationRegisterInput{
		Name:         "create-application-with-documents",
		ProviderName: "compass",
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
	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	actualApp := graphql.ApplicationExt{}

	// WHEN
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
				result: registerApplication(in: %s) { 
						%s 
					}
				}`,
			appInputGQL,
			tc.gqlFieldsProvider.ForApplication(),
		))

	saveExampleInCustomDir(t, request.Query(), registerApplicationCategory, "register application with documents")
	err = tc.RunOperation(ctx, request, &actualApp)

	//THEN
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
	applicationId := "cf889c38-490d-4896-96a7-c0721eca9932"

	t.Run("add Webhook", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		webhookInStr, err := tc.graphqlizer.WebhookInputToGQL(&graphql.WebhookInput{
			URL:  webhookURL,
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

	t.Run("add API Definition", func(t *testing.T) {
		//GIVEN
		ctx := context.Background()
		apiInStr, err := tc.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{
			Name:      "new-api-name",
			TargetURL: "https://target.url",
		})
		require.NoError(t, err)

		// WHEN
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addAPIDefinition(applicationID: "%s", in: %s) {
					%s
				}
			}`, applicationId, apiInStr, tc.gqlFieldsProvider.ForAPIDefinition()))

		err = tc.RunOperation(ctx, addReq, nil)

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
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: addEventDefinition(applicationID: "%s", in: %s) {
						%s	
					}
				}`, applicationId, eventApiInStr, tc.gqlFieldsProvider.ForEventDefinition()))
		err = tc.RunOperation(ctx, addReq, nil)

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

	actualApp := registerApplication(t, ctx, "before")
	defer unregisterApplication(t, actualApp.ID)

	expectedApp := actualApp
	expectedApp.Name = "after"
	expectedApp.ProviderName = "after"
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

func TestCreateUpdateApplicationWithDuplicatedNamesWithinTenant(t *testing.T) {
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

	t.Run("Error when updating Application with name that exists", func(t *testing.T) {
		actualApp := registerApplication(t, ctx, "differentname")
		defer unregisterApplication(t, actualApp.ID)

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
	in := fixSampleApplicationRegisterInput("app")

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: registerApplication(in: %s) {
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
			result: unregisterApplication(id: "%s") {
					id
				}
			}`, actualApp.ID))
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
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: registerApplication(in: %s) {
    					id
					}
				}`, appInputGQL))
	actualApp := graphql.ApplicationExt{}
	err = tc.RunOperation(ctx, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)

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
		actualApp := getApplication(t, ctx, actualApp.ID)
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
		updatedApp := getApplication(t, ctx, actualApp.ID)
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
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addAPIDefinition(applicationID: "%s", in: %s) {
					%s
				}
			}`, actualApp.ID, inStr, tc.gqlFieldsProvider.ForAPIDefinition()))
		saveExample(t, addReq.Query(), "add API Definition")
		err = tc.RunOperation(ctx, addReq, &actualAPI)

		//THEN
		require.NoError(t, err)
		id := actualAPI.ID
		require.NotNil(t, id)
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
		updateReq := gcli.NewRequest(
			fmt.Sprintf(`mutation { 
				result: updateAPIDefinition(id: "%s", in: %s) {
						%s
					}
				}`, id, updateStr, tc.gqlFieldsProvider.ForAPIDefinition()))
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
		deleteReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: deleteAPIDefinition(id: "%s") {
						id
					}
				}`, id))
		err = tc.RunOperation(ctx, deleteReq, &delAPI)
		saveExample(t, deleteReq.Query(), "delete API Definition")

		//THEN
		require.NoError(t, err)
		assert.Equal(t, id, delAPI.ID)

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
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: addEventDefinition(applicationID: "%s", in: %s) {
						%s	
					}
				}`, actualApp.ID, inStr, tc.gqlFieldsProvider.ForEventDefinition()))
		saveExample(t, addReq.Query(), "add Event Definition")
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
		updateReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: updateEventDefinition(id: "%s", in: %s) {
						%s
					}
				}`, actualEventAPI.ID, updateStr, tc.gqlFieldsProvider.ForEventDefinition()))
		saveExample(t, updateReq.Query(), "update Event Definition")
		err = tc.RunOperation(ctx, updateReq, &actualEventAPI)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, "updated-event-api", actualEventAPI.Name)

		// delete
		// WHEN
		delReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: deleteEventDefinition(id: "%s") {
					id
				}
			}`, actualEventAPI.ID))
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
			Name:         fmt.Sprintf("app-%d", i),
			ProviderName: "compass",
		}

		appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
		require.NoError(t, err)
		actualApp := graphql.Application{}
		request := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
		err = tc.RunOperation(ctx, request, &actualApp)
		require.NoError(t, err)
		defer unregisterApplication(t, actualApp.ID)
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
	in := graphql.ApplicationRegisterInput{
		Name:         fmt.Sprintf("app"),
		ProviderName: "Compass",
	}

	appInputGQL, err := tc.graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	actualApp := graphql.Application{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	err = tc.RunOperation(context.Background(), request, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	createdID := actualApp.ID
	defer unregisterApplication(t, actualApp.ID)

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
	appIn := fixSampleApplicationRegisterInput("adidas")
	inStr, err := tc.graphqlizer.ApplicationRegisterInputToGQL(appIn)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: registerApplication(in: %s) {
						%s
					}
				}`,
			inStr, tc.gqlFieldsProvider.ForApplication()))
	actualApp := graphql.ApplicationExt{}
	ctx := context.Background()
	err = tc.RunOperation(ctx, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer unregisterApplication(t, actualApp.ID)

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
			appInput := graphql.ApplicationRegisterInput{
				Name:           "test-app",
				ProviderName:   "compass",
				APIDefinitions: testCase.Apis,
				Labels: &graphql.Labels{
					"scenarios": []interface{}{"DEFAULT"},
				},
			}

			app := registerApplicationFromInputWithinTenant(t, ctx, appInput, defaultTenant)
			defer unregisterApplication(t, app.ID)

			var rtmIDs []string
			for i := 0; i < rtmsToCreate; i++ {
				rtm := registerRuntime(t, ctx, fmt.Sprintf("test-rtm-%d", i))
				rtmIDs = append(rtmIDs, rtm.ID)
				defer unregisterRuntime(t, rtm.ID)
			}
			require.Len(t, rtmIDs, rtmsToCreate)

			rtmIDWithUnsetAPIRtmAuth := rtmIDs[0]
			rtmIDWithSetAPIRtmAuth := rtmIDs[1]

			setAPIAuth(t, ctx, app.APIDefinitions.Data[0].ID, rtmIDWithSetAPIRtmAuth, customAuth)
			defer deleteAPIAuth(t, ctx, app.APIDefinitions.Data[0].ID, rtmIDWithSetAPIRtmAuth)

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
					require.Len(t, result.APIDefinitions.Data, 1)

					assert.Equal(t, len(rtmIDs), len(result.APIDefinitions.Data[0].Auths))
					assert.Equal(t, innerTestCase.QueriedRtmID, result.APIDefinitions.Data[0].Auth.RuntimeID)
					if innerTestCase.QueriedRtmID == rtmIDWithSetAPIRtmAuth {
						assertAuth(t, &customAuth, result.APIDefinitions.Data[0].Auth.Auth)
					} else {
						assert.Equal(t, result.APIDefinitions.Data[0].DefaultAuth, result.APIDefinitions.Data[0].Auth.Auth)
					}

					for _, auth := range result.APIDefinitions.Data[0].Auths {
						if auth.RuntimeID == rtmIDWithSetAPIRtmAuth {
							assertAuth(t, &customAuth, auth.Auth)
						} else {
							assert.Equal(t, result.APIDefinitions.Data[0].DefaultAuth, auth.Auth)
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
		TargetURL: "http://target.url",
	}

	APIInputGQL, err := tc.graphqlizer.APIDefinitionInputToGQL(in)
	require.NoError(t, err)
	applicationID := registerApplication(t, context.Background(), "test").ID
	defer unregisterApplication(t, applicationID)
	actualAPI := graphql.APIDefinition{}
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addAPIDefinition(applicationID: "%s", in: %s) {
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
					apiDefinition(id: "%s"){
						%s
					}
				}
			}`, applicationID, actualAPI.ID, tc.gqlFieldsProvider.ForAPIDefinition()))
	err = tc.RunOperation(context.Background(), queryAppReq, &actualAPI)
	saveExample(t, queryAppReq.Query(), "query api definition")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdID, actualAPI.ID)
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
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addEventDefinition(applicationID: "%s", in: %s) {
					%s
				}
			}`, applicationID, EventAPIInputGQL, tc.gqlFieldsProvider.ForEventDefinition()))
	err = tc.RunOperation(context.TODO(), request, &actualEventAPI)
	require.NoError(t, err)
	require.NotEmpty(t, actualEventAPI.ID)
	createdID := actualEventAPI.ID
	defer deleteEventAPI(t, createdID)

	// WHEN
	queryAppReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
					eventDefinition(id: "%s"){
						%s
					}
				}
			}`, applicationID, actualEventAPI.ID, tc.gqlFieldsProvider.ForEventDefinition()))
	err = tc.RunOperation(context.Background(), queryAppReq, &actualEventAPI)
	saveExample(t, queryAppReq.Query(), "query event definition")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdID, actualEventAPI.ID)
}

func fixSampleApplicationRegisterInput(placeholder string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:         placeholder,
		ProviderName: "compass",
		Documents: []*graphql.DocumentInput{{
			Title:       placeholder,
			DisplayName: placeholder,
			Description: placeholder,
			Format:      graphql.DocumentFormatMarkdown}},
		APIDefinitions: []*graphql.APIDefinitionInput{{
			Name:      placeholder,
			TargetURL: "http://kyma-project.io"}},
		EventDefinitions: []*graphql.EventDefinitionInput{{
			Name: placeholder,
			Spec: &graphql.EventSpecInput{
				Type:   graphql.EventSpecTypeAsyncAPI,
				Format: graphql.SpecFormatYaml,
				FetchRequest: &graphql.FetchRequestInput{
					URL: "https://kyma-project.io",
				},
			}}},
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
		Name:           placeholder,
		Description:    &placeholder,
		HealthCheckURL: ptr.String(webhookURL),
		ProviderName:   placeholder,
	}
}

func fixSampleApplicationUpdateInputWithIntegrationSystem(placeholder string) graphql.ApplicationUpdateInput {
	return graphql.ApplicationUpdateInput{
		Name:                placeholder,
		Description:         &placeholder,
		HealthCheckURL:      ptr.String(webhookURL),
		IntegrationSystemID: &integrationSystemID,
		ProviderName:        placeholder,
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
