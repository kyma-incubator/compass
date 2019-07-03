package end_to_end

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tc = testContext{graphqlizer: graphqlizer{}, gqlFieldsProvider: gqlFieldsProvider{}, cli: gcli.NewClient(getDirectorURL())}

func (tc *testContext) RunQuery(ctx context.Context, req *gcli.Request, resp interface{}) error {
	m := resultMapperFor(&resp)
	return tc.cli.Run(ctx, req, &m)
}

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
		Annotations: &graphql.Annotations{
			"createdBy": "admin",
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

	assert.Equal(t, in.Name, actualApp.Name)
	assert.Equal(t, in.Description, actualApp.Description)
	assert.Equal(t, *in.Annotations, actualApp.Annotations)
	assert.Equal(t, *in.Labels, actualApp.Labels)
	assert.Equal(t, in.HealthCheckURL, actualApp.HealthCheckURL)
}

func TestCreateApplicationWithWebhooks(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationInput{
		Name: "wordpress",
		Webhooks: []*graphql.ApplicationWebhookInput{
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

	assert.Len(t, actualApp.Webhooks, 1)
	actWh := actualApp.Webhooks[0]
	assert.NotEmpty(t, actWh.ID)
	assert.Equal(t, in.Webhooks[0].URL, actWh.URL)
	assert.Equal(t, in.Webhooks[0].Type, actWh.Type)
	assert.Equal(t, in.Webhooks[0].Auth.AdditionalQueryParams, actWh.Auth.AdditionalQueryParams)
	assert.Equal(t, in.Webhooks[0].Auth.AdditionalHeaders, actWh.Auth.AdditionalHeaders)

	actBasic, ok := (actWh.Auth.Credential).(*graphql.BasicCredentialData)
	require.True(t, ok)
	assert.Equal(t, in.Webhooks[0].Auth.Credential.Basic.Username, actBasic.Username)
	assert.Equal(t, in.Webhooks[0].Auth.Credential.Basic.Password, actBasic.Password)
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

	require.Len(t, actualApp.Apis.Data, 2)
	var actCommentsApi, actReviewsApi *graphql.APIDefinition
	if actualApp.Apis.Data[0].Name == "comments/v1" {
		actCommentsApi = actualApp.Apis.Data[0]
		actReviewsApi = actualApp.Apis.Data[1]
	} else {
		actCommentsApi = actualApp.Apis.Data[1]
		actReviewsApi = actualApp.Apis.Data[0]

	}
	assert.NotNil(t, actCommentsApi.ID)
	assert.Equal(t, in.Apis[0].Name, actCommentsApi.Name)
	assert.Equal(t, in.Apis[0].Description, actCommentsApi.Description)
	assert.Equal(t, in.Apis[0].TargetURL, actCommentsApi.TargetURL)
	assert.Equal(t, in.Apis[0].Group, actCommentsApi.Group)
	assert.NotNil(t, actCommentsApi.DefaultAuth)
	assert.NotNil(t, actCommentsApi.Version)
	assert.NotNil(t, actCommentsApi.Spec)

	assert.Equal(t, in.Apis[0].Spec.Type, actCommentsApi.Spec.Type)
	assert.Equal(t, in.Apis[0].Spec.Format, *actCommentsApi.Spec.Format)
	assert.Equal(t, *in.Apis[0].Spec.Data, *actCommentsApi.Spec.Data)

	require.NotNil(t, actReviewsApi.Spec.FetchRequest)
	assert.Equal(t, in.Apis[1].Spec.FetchRequest.URL, actReviewsApi.Spec.FetchRequest.URL)
	assert.Equal(t, *in.Apis[1].Spec.FetchRequest.Mode, actReviewsApi.Spec.FetchRequest.Mode)
	assert.Equal(t, in.Apis[1].Spec.FetchRequest.Filter, actReviewsApi.Spec.FetchRequest.Filter)
	assert.NotNil(t, actReviewsApi.Spec.FetchRequest.Auth)

}

func TestCreateApplicationWithEventAPIs(t *testing.T) {
	t.SkipNow()
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
					Data:          ptrCLOB(graphql.CLOB([]byte(`asyncapi: \"1.0.0\"`))),
				},
			},
			{
				Name:        "reviews/v1",
				Description: ptrString("review events"),
				Spec: &graphql.EventAPISpecInput{
					EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
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

	assert.Len(t, actualApp.EventAPIs.Data, 2)
	actCommentsApi := actualApp.EventAPIs.Data[0]
	assert.NotNil(t, actCommentsApi.ID)
	assert.Equal(t, in.EventAPIs[0].Name, actCommentsApi.Name)
	assert.Equal(t, in.EventAPIs[0].Description, actCommentsApi.Description)
	assert.Equal(t, in.EventAPIs[0].Group, actCommentsApi.Group)
	assert.NotNil(t, actCommentsApi.Version)
	assert.NotNil(t, actCommentsApi.Spec)

	assert.Equal(t, in.EventAPIs[0].Spec.EventSpecType, actCommentsApi.Spec.Type)
	assert.Equal(t, in.EventAPIs[0].Spec.Data, actCommentsApi.Spec.Data)

	actReviewsApi := actualApp.EventAPIs.Data[1]
	require.NotNil(t, actReviewsApi.Spec.FetchRequest)
	assert.Equal(t, in.EventAPIs[1].Spec.FetchRequest.URL, actReviewsApi.Spec.FetchRequest.URL)
	assert.Equal(t, in.EventAPIs[1].Spec.FetchRequest.Mode, actReviewsApi.Spec.FetchRequest.Mode)
	assert.Equal(t, in.EventAPIs[1].Spec.FetchRequest.Filter, actReviewsApi.Spec.FetchRequest.Filter)
	assert.NotNil(t, actReviewsApi.Spec.FetchRequest.Auth)

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

	assert.Len(t, actualApp.Documents.Data, 2)
	var actReadme, actTrouble *graphql.Document
	if actualApp.Documents.Data[0].Title == "Readme" {
		actReadme = actualApp.Documents.Data[0]
		actTrouble = actualApp.Documents.Data[1]
	} else {
		actReadme = actualApp.Documents.Data[1]
		actTrouble = actualApp.Documents.Data[0]
	}
	assert.Equal(t, in.Documents[0].Title, actReadme.Title)
	require.NotNil(t, actReadme.FetchRequest)
	assert.Equal(t, in.Documents[0].FetchRequest.URL, actReadme.FetchRequest.URL)

	assert.Equal(t, in.Documents[1].Title, actTrouble.Title)
	assert.Equal(t, in.Documents[1].Data, actTrouble.Data)

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
	assert.Equal(t, "after", updatedApp.Name)
	require.Len(t, updatedApp.Documents.Data, 1)
	assert.Equal(t, "after", updatedApp.Documents.Data[0].Title)
	require.Len(t, updatedApp.Apis.Data, 1)
	assert.Equal(t, "after", updatedApp.Apis.Data[0].Name)
	assert.Equal(t, "after", updatedApp.Apis.Data[0].TargetURL)
	// TODO
	// require.Len(t, updatedApp.EventAPIs.Data, 1)
	// assert.Equal(t, "after", updatedApp.EventAPIs.Data[0].Name)
	require.Len(t, updatedApp.Webhooks, 1)

	assert.Equal(t, "after", updatedApp.Webhooks[0].URL)
	assert.Equal(t, graphql.Labels{"after": []string{"after"}}, updatedApp.Labels)
	assert.Equal(t, graphql.Annotations{"after": "after"}, updatedApp.Annotations)
	assert.Equal(t, id, updatedApp.ID)    // id was not changed
	assert.Nil(t, updatedApp.Description) // all fields are overridden
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

	t.Run("annotations manipulation", func(t *testing.T) {
		// create label
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addApplicationAnnotation(applicationID: "%s", key: "%s", value: "%s") {
					key 
					value
				}
			}`, actualApp.ID, "brand-new-annotation", "ccc"))
		saveQueryInExamples(t, addReq.Query(), "add application annotation")
		actualAnnotation := graphql.Annotation{}
		err := tc.RunQuery(ctx, addReq, &actualAnnotation)
		require.NoError(t, err)
		assert.Equal(t, graphql.Annotation{Key: "brand-new-annotation", Value: "ccc"}, actualAnnotation)
		actualApp := getApp(ctx, t, actualApp.ID)
		assert.Equal(t, "ccc", actualApp.Annotations["brand-new-annotation"])

		// delete label
		delReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: deleteApplicationAnnotation(applicationID: "%s", key: "%s") {
					key 
					value
				}
			}`, actualApp.ID, "brand-new-annotation"))
		saveQueryInExamples(t, delReq.Query(), "delete application annotation")
		err = tc.RunQuery(ctx, delReq, &actualAnnotation)
		require.NoError(t, err)
		assert.Equal(t, graphql.Annotation{Key: "brand-new-annotation", Value: "ccc"}, actualAnnotation)
		// TODO here we have inconsistency with labels
		actualApp = getApp(ctx, t, actualApp.ID)
		assert.Nil(t, actualApp.Annotations["brand-new-annotation"])
	})

	t.Run("manage webhooks", func(t *testing.T) {
		// add
		webhookInStr, err := tc.graphqlizer.ApplicationWebhookInputToGQL(&graphql.ApplicationWebhookInput{
			URL:  "new-webhook",
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		addReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: addApplicationWebhook(applicationID: "%s", in: %s) {
					%s
				}
			}`, actualApp.ID, webhookInStr, tc.gqlFieldsProvider.ForWebhooks()))
		saveQueryInExamples(t, addReq.Query(), "add aplication webhook")

		actualWebhook := graphql.ApplicationWebhook{}
		err = tc.RunQuery(ctx, addReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "new-webhook", actualWebhook.URL)
		assert.Equal(t, graphql.ApplicationWebhookTypeConfigurationChanged, actualWebhook.Type)
		id := actualWebhook.ID
		require.NotNil(t, id)

		// get all webhooks
		updatedApp := getApp(ctx, t, actualApp.ID)
		assert.Len(t, updatedApp.Webhooks, 2)

		// update
		webhookInStr, err = tc.graphqlizer.ApplicationWebhookInputToGQL(&graphql.ApplicationWebhookInput{
			URL: "updated-webhook", Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		updateReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: updateApplicationWebhook(webhookID: "%s", in: %s) {
					%s
				}
			}`, actualWebhook.ID, webhookInStr, tc.gqlFieldsProvider.ForWebhooks()))
		saveQueryInExamples(t, updateReq.Query(), "update application webhook")
		err = tc.RunQuery(ctx, updateReq, &actualWebhook)
		require.NoError(t, err)
		assert.Equal(t, "updated-webhook", actualWebhook.URL)

		// delete
		deleteReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: deleteApplicationWebhook(webhookID: "%s") {
					%s
				}
			}`, actualWebhook.ID, tc.gqlFieldsProvider.ForWebhooks()))
		saveQueryInExamples(t, deleteReq.Query(), "delete application webhook")
		err = tc.RunQuery(ctx, deleteReq, &actualWebhook)
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

		// update
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
		// TODO
	})

	t.Run("manage documents", func(t *testing.T) {
		t.SkipNow()
		// add
		inStr, err := tc.graphqlizer.DocumentInputToGQL(&graphql.DocumentInput{
			Title: "new-document",
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
		//
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
		// TODO
	})
}

func TestRuntimeCreateUpdateAndDelete(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	givenInput := graphql.RuntimeInput{
		Name:        "runtime-1",
		Description: ptrString("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []string{"hhh"}},
		Annotations: &graphql.Annotations{"kkk": "lll"},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	actualRuntime := graphql.Runtime{}

	// WHEN
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createRuntime(in: %s) {
					%s
				}
			}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	saveQueryInExamples(t, createReq.Query(), "create runtime")
	err = tc.RunQuery(ctx, createReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assert.Equal(t, givenInput.Name, actualRuntime.Name)
	assert.Equal(t, *givenInput.Description, *actualRuntime.Description)
	assert.Equal(t, *givenInput.Labels, actualRuntime.Labels)
	assert.Equal(t, *givenInput.Annotations, actualRuntime.Annotations)
	assert.NotNil(t, actualRuntime.AgentAuth)

	// update runtime
	givenInput.Description = ptrString("modified-runtime-1-description")
	runtimeInGQL, err = tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)

	// WHEN
	updateReq := gcli.NewRequest(
		fmt.Sprintf(`mutation{ 
			result: updateRuntime(id: "%s", in: %s) {
					%s
				}
			}`, actualRuntime.ID, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	saveQueryInExamples(t, updateReq.Query(), "update runtime")
	err = tc.RunQuery(ctx, updateReq, &actualRuntime)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, *givenInput.Description, *actualRuntime.Description)

	// add Label
	actualLabel := graphql.Label{}

	// WHEN
	addLabelReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addRuntimeLabel(runtimeID: "%s", key: "%s", values: %s) {
					%s
				}
			}`, actualRuntime.ID, "new-label", "[\"bbb\"]", tc.gqlFieldsProvider.ForLabel()))
	err = tc.RunQuery(ctx, addLabelReq, &actualLabel)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, "new-label", actualLabel.Key)
	assert.Len(t, actualLabel.Values, 1)
	assert.Contains(t, actualLabel.Values, "bbb")

	// add Annotation
	actualAnnotation := graphql.Annotation{}

	// WHEN
	addAnnotationReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addRuntimeAnnotation(runtimeID: "%s", key: "%s", value: %s) {
					%s
				}
			}`, actualRuntime.ID, "new-anno", `["zzz"]`, tc.gqlFieldsProvider.ForAnnotation()))
	err = tc.RunQuery(ctx, addAnnotationReq, &actualAnnotation)

	//THEN
	require.NoError(t, err)
	assert.Equal(t, "new-anno", actualAnnotation.Key)
	assert.Equal(t, []interface{}{"zzz"}, actualAnnotation.Value)

	// get runtime and validate runtimes and annotations
	getRuntimeReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}
			}`, actualRuntime.ID, tc.gqlFieldsProvider.ForRuntime()))
	err = tc.RunQuery(ctx, getRuntimeReq, &actualRuntime)
	require.NoError(t, err)
	assert.Len(t, actualRuntime.Labels, 2)
	assert.Len(t, actualRuntime.Annotations, 2)

	// delete label

	// WHEN
	delLabelReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: deleteRuntimeLabel(runtimeID: "%s", key: "%s") {
						%s
					}
				}
		`, actualRuntime.ID, "new-label", tc.gqlFieldsProvider.ForLabel()))
	err = tc.RunQuery(ctx, delLabelReq, nil)

	//THEN
	require.NoError(t, err)
	// delete annotation

	// WHEN
	delAnnoReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: deleteRuntimeAnnotation(runtimeID: "%s", key: "%s") {
						%s
				}
			}`, actualRuntime.ID, "new-anno", tc.gqlFieldsProvider.ForAnnotation()))
	err = tc.RunQuery(ctx, delAnnoReq, nil)

	//THEN
	require.NoError(t, err)

	// delete runtime

	// WHEN
	delReq := gcli.NewRequest(fmt.Sprintf(`mutation{result: deleteRuntime(id: "%s") {%s}}`, actualRuntime.ID, tc.gqlFieldsProvider.ForRuntime()))
	saveQueryInExamples(t, delReq.Query(), "delete runtime")
	err = tc.RunQuery(ctx, delReq, nil)

	//THEN
	require.NoError(t, err)
}

func TestSetAndDeleteAPIAuth(t *testing.T) {
	// GIVEN
	// create application
	ctx := context.Background()
	placeholder := "app"
	in := generateSampleApplicationInput(placeholder)

	appInputGQL, err := tc.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
  				result: createApplication(in: %s) {
    					%s
					}
				}`, appInputGQL, tc.gqlFieldsProvider.ForApplication()))
	actualApp := ApplicationExt{}
	err = tc.RunQuery(ctx, createReq, &actualApp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, actualApp.ID)

	// create runtime
	runtimeInput := graphql.RuntimeInput{
		Name:        "runtime-1",
		Description: ptrString("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []string{"hhh"}},
		Annotations: &graphql.Annotations{"kkk": "lll"},
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(runtimeInput)
	require.NoError(t, err)
	actualRuntime := graphql.Runtime{}
	createRuntimeReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: createRuntime(in: %s) {
						%s
					}
				}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	err = tc.RunQuery(ctx, createRuntimeReq, &actualRuntime)
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)

	defer deleteRuntime(t, actualRuntime.ID)

	actualRuntimeAuth := graphql.RuntimeAuth{}

	// WHEN
	// set Auth
	authIn := graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "x-men",
				Password: "secret",
			}}}

	authInStr, err := tc.graphqlizer.AuthInputToGQL(&authIn)
	require.NoError(t, err)
	setAuthReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setAPIAuth(apiID: "%s", runtimeID: "%s", in: %s) {
					%s
				}
			}`, actualApp.Apis.Data[0].ID, actualRuntime.ID, authInStr, tc.gqlFieldsProvider.ForRuntimeAuth()))
	err = tc.RunQuery(ctx, setAuthReq, &actualRuntimeAuth)

	//THEN
	require.NoError(t, err)
	require.NotNil(t, actualRuntimeAuth.Auth)
	assert.Equal(t, actualRuntime.ID, actualRuntimeAuth.RuntimeID)
	actualBasic, ok := actualRuntimeAuth.Auth.Credential.(*graphql.BasicCredentialData)
	require.True(t, ok)
	assert.Equal(t, "x-men", actualBasic.Username)
	assert.Equal(t, "secret", actualBasic.Password)
	// delete Auth
	delAuthReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteAPIAuth(apiID: "%s",runtimeID: "%s") {
					%s
				} 
			}`, actualApp.Apis.Data[0].ID, actualRuntime.ID, tc.gqlFieldsProvider.ForRuntimeAuth()))
	err = tc.RunQuery(ctx, delAuthReq, nil)
	require.NoError(t, err)

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

func TestQueryRuntimes(t *testing.T) {
	// GIVEN
	ctx := context.Background()

	idsToRemove := make([]string, 3)
	defer func() {
		for _, id := range idsToRemove {
			if id != "" {
				deleteRuntime(t, id)
			}
		}
	}()

	for i := 0; i < 3; i++ {
		givenInput := graphql.RuntimeInput{
			Name: fmt.Sprintf("runtime-%d", i),
		}
		runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
		require.NoError(t, err)
		createReq := gcli.NewRequest(
			fmt.Sprintf(`mutation {
				result: createRuntime(in: %s) {
						%s
					} 
				}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
		actualRuntime := graphql.Runtime{}
		err = tc.RunQuery(ctx, createReq, &actualRuntime)
		require.NoError(t, err)
		require.NotEmpty(t, actualRuntime.ID)
		idsToRemove[i] = actualRuntime.ID
	}
	actualPage := graphql.RuntimePage{}

	// WHEN
	queryReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtimes {
					%s
				}
			}`, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForRuntime())))
	err := tc.RunQuery(ctx, queryReq, &actualPage)
	saveQueryInExamples(t, queryReq.Query(), "query runtimes")

	//THEN
	require.NoError(t, err)
	assert.Len(t, actualPage.Data, 3)
	assert.Equal(t, 3, actualPage.TotalCount)

}

func TestQuerySpecificRuntime(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	givenInput := graphql.RuntimeInput{
		Name: "runtime-1",
	}
	runtimeInGQL, err := tc.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createRuntime(in: %s) {
					%s
				}
			}`, runtimeInGQL, tc.gqlFieldsProvider.ForRuntime()))
	createdRuntime := graphql.Runtime{}
	err = tc.RunQuery(ctx, createReq, &createdRuntime)
	require.NoError(t, err)
	require.NotEmpty(t, createdRuntime.ID)

	defer deleteRuntime(t, createdRuntime.ID)
	queriedRuntime := graphql.Runtime{}

	// WHEN
	queryReq := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: runtime(id: "%s") {
					%s
				}
			}`, createdRuntime.ID, tc.gqlFieldsProvider.ForRuntime()))
	err = tc.RunQuery(ctx, queryReq, &queriedRuntime)
	saveQueryInExamples(t, queryReq.Query(), "query specific runtime")

	//THEN
	require.NoError(t, err)
	assert.Equal(t, createdRuntime.ID, queriedRuntime.ID)
	assert.Equal(t, createdRuntime.Name, queriedRuntime.Name)
}

func TestTenantSeparation(t *testing.T) {
	//TODO
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
			}}},
		Webhooks: []*graphql.ApplicationWebhookInput{{
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
			URL:  placeholder},
		},
		Labels:      &graphql.Labels{placeholder: []string{placeholder}},
		Annotations: &graphql.Annotations{placeholder: placeholder},
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

func deleteRuntime(t *testing.T, id string) {
	delReq := gcli.NewRequest(
		fmt.Sprintf(`mutation{deleteRuntime(id: "%s") {
				id
			}
		}`, id))
	err := tc.RunQuery(context.Background(), delReq, nil)
	require.NoError(t, err)
}

func fixBasicAuth() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Basic: &graphql.BasicCredentialDataInput{
				Username: "admin",
				Password: "secret",
			},
		},
		AdditionalHeaders: &graphql.HttpHeaders{
			"headerA": []string{"ha1", "ha2"},
			"headerB": []string{"hb1", "hb2"},
		},
		AdditionalQueryParams: &graphql.QueryParams{
			"qA": []string{"qa1", "qa2"},
			"qB": []string{"qb1", "qb2"},
		},
	}
}

func fixOauthAuth() *graphql.AuthInput {
	return &graphql.AuthInput{
		Credential: &graphql.CredentialDataInput{
			Oauth: &graphql.OAuthCredentialDataInput{
				URL:          "http://oauth/token",
				ClientID:     "clientID",
				ClientSecret: "clientSecret",
			},
		},
	}
}

func fixDepracatedVersion1() *graphql.VersionInput {
	return &graphql.VersionInput{
		Value:           "v1",
		Deprecated:      ptrBool(true),
		ForRemoval:      ptrBool(true),
		DeprecatedSince: ptrString("v5"),
	}
}

// testContext contains dependencies that help executing tests
type testContext struct {
	graphqlizer       graphqlizer
	gqlFieldsProvider gqlFieldsProvider
	cli               *gcli.Client
}

func getDirectorURL() string {
	url := os.Getenv("DIRECTOR_GRAPHQL_API")
	if url == "" {
		url = "http://127.0.0.1:3000/graphql"
	}
	return url
}

// resultMapperFor returns generic object that can be passed to Run method for storing response.
// In GraphQL, set `result` alias for your query
func resultMapperFor(target interface{}) genericGQLResponse {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		panic("target has to be a pointer")
	}
	return genericGQLResponse{
		Result: target,
	}
}

type genericGQLResponse struct {
	Result interface{} `json:"result"`
}

func saveQueryInExamples(t *testing.T, query string, exampleName string) {
	// TODO uncomment this in https://github.com/kyma-incubator/compass/issues/77
	//t.Helper()
	//sanitizedName := strings.Replace(exampleName, " ", "-", -1)
	//// replace uuids with constant value
	//r, err := regexp.Compile("[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}")
	//require.NoError(t, err)
	//query = r.ReplaceAllString(query, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	//content := fmt.Sprintf("# This example is generated from test compass_e2e_test.go\n%s", query)
	//err = ioutil.WriteFile(fmt.Sprintf("%s/src/github.com/kyma-incubator/compass/components/director/examples/%s.graphql", os.Getenv("GOPATH"), sanitizedName), []byte(content), 0660)
	//require.NoError(t, err)
}
