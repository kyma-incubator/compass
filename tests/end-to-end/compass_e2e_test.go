package end_to_end

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-project/kyma/components/helm-broker/platform/ptr"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var ts = testSuite{graphqlizer: graphqlizer{}, fieldsProvider: fieldsProvider{}, cli: gcli.NewClient(getDirectorURL())}

//TODO cleanup objects created in the test

func TestCreateApplicationWithAllSimpleFieldsProvided(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationInput{
		Name:           "wordpress",
		Description:    ptrString("my first wordperss application"),
		HealthCheckURL: ptrString("http://mywordpress.com/health"),
		Labels: &graphql.Labels{
			"group": []string{"production", "experimental"},
		},
		Annotations: &graphql.Annotations{
			"createdBy": "admin",
		},
	}

	// WHEN
	appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	request := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {%s}
			}`, appInputGQL, ts.fieldsProvider.ForApplication()))

	storeExampleQuery(t, request.Query(), "create application")
	actualApp := graphql.Application{}
	resp := resultMapperFor(&actualApp)
	err = ts.cli.Run(ctx, request, &resp)
	// THEN
	require.NoError(t, err)
	assert.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, ts.cli, actualApp.ID)

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

	// WHEN

	appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
	result: createApplication(in: %s) {
	%s
}
}`,
			appInputGQL,
			ts.fieldsProvider.ForApplication(),
		))
	storeExampleQuery(t, request.Query(), "create application with webhooks")
	actualApp := graphql.Application{}
	createResp := resultMapperFor(&actualApp)
	err = ts.cli.Run(ctx, request, &createResp)
	// THEN
	require.NoError(t, err)

	assert.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, ts.cli, actualApp.ID)

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
	t.SkipNow() //TODO
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationInput{
		Name: "wordpress",
		Apis: []*graphql.APIDefinitionInput{
			{
				Name:        "comments/v1",
				Description: ptrString("api for adding comments"),
				TargetURL:   "http://mywordpress.com/comments",
				Group:       ptr.String("comments"),
				DefaultAuth: fixBasicAuth(),
				Version:     fixDepracatedVersion1(),
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOpenAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptrCLOB(graphql.CLOB(`openapi: \"3.0.0\"`)),
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

	// WHEN

	appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  result: createApplication(in: %s) {
    %s
  }
}`,
			appInputGQL,
			ts.fieldsProvider.ForApplication(),
		))
	storeExampleQuery(t, request.Query(), "create application with APIs")
	actualApp := graphql.Application{}
	createResp := resultMapperFor(&actualApp)
	err = ts.cli.Run(ctx, request, &createResp)
	// THEN
	require.NoError(t, err)
	assert.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, ts.cli, actualApp.ID)

	require.Len(t, actualApp.Apis.Data, 2)
	actCommentsApi := actualApp.Apis.Data[0]
	assert.NotNil(t, actCommentsApi.ID)
	assert.Equal(t, in.Apis[0].Name, actCommentsApi.Name)
	assert.Equal(t, in.Apis[0].Description, actCommentsApi.Description)
	assert.Equal(t, in.Apis[0].TargetURL, actCommentsApi.TargetURL)
	assert.Equal(t, in.Apis[0].Group, actCommentsApi.Group)
	assert.NotNil(t, actCommentsApi.DefaultAuth)
	assert.NotNil(t, actCommentsApi.Version)
	assert.NotNil(t, actCommentsApi.Spec)

	assert.Equal(t, in.Apis[0].Spec.Type, actCommentsApi.Spec.Type)
	assert.Equal(t, in.Apis[0].Spec.Format, actCommentsApi.Spec.Format)
	assert.Equal(t, in.Apis[0].Spec.Data, actCommentsApi.Spec.Data)

	actReviewsApi := actualApp.Apis.Data[1]
	require.NotNil(t, actReviewsApi.Spec.FetchRequest)
	assert.Equal(t, in.Apis[1].Spec.FetchRequest.URL, actReviewsApi.Spec.FetchRequest.URL)
	assert.Equal(t, in.Apis[1].Spec.FetchRequest.Mode, actReviewsApi.Spec.FetchRequest.Mode)
	assert.Equal(t, in.Apis[1].Spec.FetchRequest.Filter, actReviewsApi.Spec.FetchRequest.Filter)
	assert.NotNil(t, actReviewsApi.Spec.FetchRequest.Auth)

}

func TestCreateApplicationWithEventAPIs(t *testing.T) {
	t.SkipNow() //TODO
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

	// WHEN

	appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  result: createApplication(in: %s) {
    %s
  }
}`,
			appInputGQL,
			ts.fieldsProvider.ForApplication(),
		))
	storeExampleQuery(t, request.Query(), "create application with event APIs")
	actualApp := graphql.Application{}
	createResp := resultMapperFor(&actualApp)
	err = ts.cli.Run(ctx, request, &createResp)
	// THEN
	require.NoError(t, err)
	assert.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, ts.cli, actualApp.ID)

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
	// WHEN

	appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  result: createApplication(in: %s) {
    %s
  }
}`,
			appInputGQL,
			ts.fieldsProvider.ForApplication(),
		))
	storeExampleQuery(t, request.Query(), "create application with documents")
	actualApp := graphql.Application{}
	createResp := resultMapperFor(actualApp)
	err = ts.cli.Run(ctx, request, &createResp)
	// THEN
	require.NoError(t, err)
	assert.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, ts.cli, actualApp.ID)

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

func TestCreateApplicationWithAllDependencies(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := graphql.ApplicationInput{
		Name:        "wordpress",
		Description: ptrString("my first wordperss application"),

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
		Apis: []*graphql.APIDefinitionInput{
			{
				Name:        "comments/v1",
				Description: ptrString("api for adding comments"),
				TargetURL:   "http://mywordpress.com/comments",
				Group:       ptr.String("comments"),
				DefaultAuth: fixBasicAuth(),
				Version:     fixDepracatedVersion1(),
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOpenAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptrCLOB(graphql.CLOB(`openapi: \"3.0.0\"`)),
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
		},
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
		Webhooks: []*graphql.ApplicationWebhookInput{
			{
				Type: graphql.ApplicationWebhookTypeConfigurationChanged,
				Auth: fixBasicAuth(),
				URL:  "http://mywordpress.com/webhooks1",
			},
			{
				Type: graphql.ApplicationWebhookTypeConfigurationChanged,
				Auth: fixBasicAuth(),
				URL:  "http://mywordpress.com/webhooks2",
			},
		},
		HealthCheckURL: ptrString("http://mywordpress.com/health"),
		Labels: &graphql.Labels{
			"group": []string{"production", "experimental"},
		},
		Annotations: &graphql.Annotations{
			"createdBy": "admin",
		},
	}
	// WHEN

	appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  result: createApplication(in: %s) {
    %s
  }
}`,
			appInputGQL,
			ts.fieldsProvider.ForApplication(),
		))
	storeExampleQuery(t, request.Query(), "create application full")
	var actualApp graphql.Application
	resp := resultMapperFor(&actualApp)

	err = ts.cli.Run(ctx, request, &resp)
	// THEN
	require.NoError(t, err)
	assertApplication(t, in, actualApp)

}

func TestUpdateApplication(t *testing.T) {
	t.SkipNow() // TODO

	// GIVEN
	ctx := context.Background()
	in := getApplicationInput("before")
	in.Description = ptrString("before")
	// WHEN

	appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	request := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  				result: createApplication(in: %s) {
    				id}}`, appInputGQL))
	actualApp := graphql.Application{}
	createResp := resultMapperFor(&actualApp)
	err = ts.cli.Run(ctx, request, &createResp)
	// THEN
	require.NoError(t, err)
	id := actualApp.ID
	require.NotEmpty(t, id)
	in = getApplicationInput("after")

	appInputGQL, err = ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	request = gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  				result: updateApplication(id: "%s", in: %s) {
    				%s}}`, id, appInputGQL, ts.fieldsProvider.ForApplication()))
	storeExampleQuery(t, request.Query(), "update application")

	updateAppResp := resultMapperFor(&actualApp)
	err = ts.cli.Run(ctx, request, &updateAppResp)
	require.NoError(t, err)
	assert.Equal(t, "after", actualApp.Name)
	require.Len(t, actualApp.Documents.Data, 1)
	assert.Equal(t, "after", actualApp.Documents.Data[0].Title)
	require.Len(t, actualApp.Apis.Data, 1)
	assert.Equal(t, "after", actualApp.Apis.Data[0].Name)
	assert.Equal(t, "after", actualApp.Apis.Data[0].TargetURL)
	require.Len(t, actualApp.EventAPIs.Data, 1)
	assert.Equal(t, "after", actualApp.EventAPIs.Data[0].Name)
	require.Len(t, actualApp.Webhooks, 1)

	assert.Equal(t, "after", actualApp.Webhooks[0].URL)
	assert.Equal(t, &graphql.Labels{"after": []string{"after"}}, actualApp.Labels)
	assert.Equal(t, &graphql.Annotations{"after": "after"}, actualApp.Annotations)
	assert.Equal(t, id, actualApp.ID)    // id was not changed
	assert.Nil(t, actualApp.Description) // all fields are overridden
	deleteApplication(t, ts.cli, actualApp.ID)

}

func TestDeleteApplication(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	in := getApplicationInput("app")

	appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  				result: createApplication(in: %s) {
    				id}}`, appInputGQL))
	actualApp := graphql.Application{}
	createResp := resultMapperFor(actualApp)
	err = ts.cli.Run(ctx, createReq, &createResp)
	require.NoError(t, err)

	require.NotEmpty(t, actualApp.ID)
	// WHEN
	delReq := gcli.NewRequest(fmt.Sprintf(`mutation{ressult: deleteApplication(id: "%s") {id}}`, actualApp.ID))
	storeExampleQuery(t, delReq.Query(), "delete application")
	err = ts.cli.Run(ctx, delReq, &actualApp)
	require.NoError(t, err)
}

func TestUpdateApplicationParts(t *testing.T) {
	ctx := context.Background()
	in := getApplicationInput("app")

	appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
	require.NoError(t, err)
	createReq := gcli.NewRequest(
		fmt.Sprintf(
			`mutation {
  				result: createApplication(in: %s) {
    				id}}`, appInputGQL))
	actualApp := graphql.Application{}
	createAppResp := resultMapperFor(&actualApp)
	err = ts.cli.Run(ctx, createReq, &createAppResp)
	require.NoError(t, err)
	require.NotEmpty(t, actualApp.ID)
	defer deleteApplication(t, ts.cli, actualApp.ID)

	t.Run("labels manipulation", func(t *testing.T) {
		addReq := gcli.NewRequest(fmt.Sprintf(`mutation {
			result: addApplicationLabel(applicationID: "%s", key: "%s", values: %s) {key values}
		}`, actualApp.ID, "brand-new-label", "[\"aaa\",\"bbb\"]"))
		storeExampleQuery(t, addReq.Query(), "add application label")
		createdLabel := &graphql.Label{}
		addResp := resultMapperFor(createdLabel)
		err := ts.cli.Run(ctx, addReq, &addResp)
		require.NoError(t, err)
		assert.Equal(t, &graphql.Label{Key: "brand-new-label", Values: []string{"aaa", "bbb"}}, createdLabel)
		actualApp := getApp(ctx, t, actualApp.ID, ts.cli)
		assert.Equal(t, []string{"aaa", "bbb"}, actualApp.Labels["brand-new-label"])

		delReq := gcli.NewRequest(fmt.Sprintf(`mutation {
			result: deleteApplicationLabel(applicationID: "%s", key: "%s", values: %s) {key values}
		}`, actualApp.ID, "brand-new-label", "[\"aaa\"]"))
		storeExampleQuery(t, delReq.Query(), "delete application label")
		deletedLabel := &graphql.Label{}
		delResp := resultMapperFor(deletedLabel)
		err = ts.cli.Run(ctx, delReq, &delResp)
		require.NoError(t, err)
		assert.Equal(t, &graphql.Label{Key: "brand-new-label", Values: []string{"bbb"}}, deletedLabel)
		actualApp = getApp(ctx, t, actualApp.ID, ts.cli)

		delReq = gcli.NewRequest(fmt.Sprintf(`mutation {
			result: deleteApplicationLabel(applicationID: "%s", key: "%s", values: %s) {key values}
		}`, actualApp.ID, "brand-new-label", "[\"bbb\"]"))
		err = ts.cli.Run(ctx, delReq, &delResp)
		require.NoError(t, err)
		assert.Equal(t, &graphql.Label{Key: "brand-new-label", Values: []string{}}, deletedLabel)
		actualApp = getApp(ctx, t, actualApp.ID, ts.cli)
		assert.Nil(t, actualApp.Labels["brand-new-label"])

	})

	t.Run("annotations manipulation", func(t *testing.T) {
		addReq := gcli.NewRequest(fmt.Sprintf(`mutation {
			result: addApplicationAnnotation(applicationID: "%s", key: "%s", value: "%s")  {key value}
		}`, actualApp.ID, "brand-new-annotation", "ccc"))
		storeExampleQuery(t, addReq.Query(), "add application annotation")
		actualAnnotation := graphql.Annotation{}
		addResp := resultMapperFor(&actualAnnotation)
		err := ts.cli.Run(ctx, addReq, &addResp)
		require.NoError(t, err)
		assert.Equal(t, graphql.Annotation{Key: "brand-new-annotation", Value: "ccc"}, actualAnnotation)
		actualApp := getApp(ctx, t, actualApp.ID, ts.cli)
		assert.Equal(t, "ccc", actualApp.Annotations["brand-new-annotation"])

		delReq := gcli.NewRequest(fmt.Sprintf(`mutation {
			result: deleteApplicationAnnotation(applicationID: "%s", key: "%s") {key value}
		}`, actualApp.ID, "brand-new-annotation"))
		storeExampleQuery(t, delReq.Query(), "delete application label")
		remResp := resultMapperFor(&actualAnnotation)
		err = ts.cli.Run(ctx, delReq, &remResp)
		require.NoError(t, err)
		assert.Equal(t, graphql.Annotation{Key: "brand-new-annotation", Value: "ccc"}, actualAnnotation)
		// TODO inconsistency
		actualApp = getApp(ctx, t, actualApp.ID, ts.cli)
		assert.Nil(t, actualApp.Annotations["brand-new-annotation"])
	})

	t.Run("manage webhooks", func(t *testing.T) {
		// add
		webhookInStr, err := ts.graphqlizer.ApplicationWebhookInputToGQL(&graphql.ApplicationWebhookInput{
			URL:  "new-webhook",
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		addReq := gcli.NewRequest(fmt.Sprintf(`mutation {
			result: addApplicationWebhook(applicationID: "%s", in: %s)  {%s}
		}`, actualApp.ID, webhookInStr, ts.fieldsProvider.ForWebhooks()))
		storeExampleQuery(t, addReq.Query(), "add aplication webhook")

		actualWebhook := graphql.ApplicationWebhook{}
		addResp := resultMapperFor(&actualWebhook)
		err = ts.cli.Run(ctx, addReq, &addResp)
		require.NoError(t, err)
		assert.Equal(t, "new-webhook", actualWebhook.URL)
		assert.Equal(t, graphql.ApplicationWebhookTypeConfigurationChanged, actualWebhook.Type)
		id := actualWebhook.ID
		require.NotNil(t, id)

		// get all webhooks
		updatedApp := getApp(ctx, t, actualApp.ID, ts.cli)
		assert.Len(t, updatedApp.Webhooks, 2)

		// update
		webhookInStr, err = ts.graphqlizer.ApplicationWebhookInputToGQL(&graphql.ApplicationWebhookInput{
			URL: "updated-webhook", Type: graphql.ApplicationWebhookTypeConfigurationChanged,
		})

		require.NoError(t, err)
		updateReq := gcli.NewRequest(fmt.Sprintf(`mutation {
			result: updateApplicationWebhook(webhookID: "%s", in: %s)  {%s}
		}`, actualWebhook.ID, webhookInStr, ts.fieldsProvider.ForWebhooks()))
		storeExampleQuery(t, updateReq.Query(), "update application webhook")
		updateResp := resultMapperFor(&actualWebhook)
		err = ts.cli.Run(ctx, updateReq, &updateResp)
		require.NoError(t, err)
		assert.Equal(t, "updated-webhook", actualWebhook.URL)

		// delete
		deleteReq := gcli.NewRequest(fmt.Sprintf(`mutation {
			result: deleteApplicationWebhook(webhookID: "%s")  {%s}
		}`, actualWebhook.ID, ts.fieldsProvider.ForWebhooks()))
		storeExampleQuery(t, deleteReq.Query(), "delete application webhook")
		deleteResp := resultMapperFor(&actualWebhook)
		err = ts.cli.Run(ctx, deleteReq, &deleteResp)
		require.NoError(t, err)
		assert.Equal(t, "updated-webhook", actualWebhook.URL)

	})

	t.Run("manage APIs", func(t *testing.T) {
		t.SkipNow() //TODO
		// add
		inStr, err := ts.graphqlizer.APIDefinitionInputToGQL(graphql.APIDefinitionInput{
			Name:      "new-api-name",
			TargetURL: "new-api-url",
		})

		require.NoError(t, err)
		addReq := gcli.NewRequest(fmt.Sprintf(`mutation {
			result: addAPI(applicationID: "%s", in: %s)  {%s}
		}`, actualApp.ID, inStr, ts.fieldsProvider.ForAPIDefinition()))
		actualAPI := graphql.APIDefinition{}
		addResp := resultMapperFor(&actualAPI)
		err = ts.cli.Run(ctx, addReq, &addResp)
		require.NoError(t, err)
		id := actualAPI.ID
		require.NotNil(t, id)
		assert.Equal(t, "new-api-name", actualAPI.Name)
		assert.Equal(t, "new-api-url", actualAPI.TargetURL)
		//
		updatedApp := getApp(ctx, t, actualApp.ID, ts.cli)
		assert.Len(t, updatedApp.Apis.Data, 2) //TODO

		// update
		//webhookInStr, err = ApplicationWebhookInputToGQL(&graphql.ApplicationWebhookInput{
		//	URL:  "updated-webhook", Type:graphql.ApplicationWebhookTypeConfigurationChanged,
		//})
		//
		//require.NoError(t, err)
		//updateReq := gcli.NewRequest(fmt.Sprintf(`mutation {
		//	result: updateApplicationWebhook(webhookID: "%s", in: %s)  {%s}
		//}`, actualWebhook.ID, webhookInStr, ForWebhooks()))
		//storeExampleQuery(t, updateReq.Query())
		//updateResp := resultMapperFor(&actualWebhook)
		//err = cli.Run(ctx, updateReq, &updateResp)
		//require.NoError(t, err)
		//assert.Equal(t, "updated-webhook", actualWebhook.URL)
		//
		//// delete
		//deleteReq := gcli.NewRequest(fmt.Sprintf(`mutation {
		//	result: deleteApplicationWebhook(webhookID: "%s")  {%s}
		//}`, actualWebhook.ID, ForWebhooks()))
		//storeExampleQuery(t, updateReq.Query())
		//deleteResp := resultMapperFor(&actualWebhook)
		//err = cli.Run(ctx, deleteReq, &deleteResp)
		//require.NoError(t, err)
		//assert.Equal(t, "updated-webhook", actualWebhook.URL)

	})

	t.Run("manage event api", func(t *testing.T) {
		// TODO
	})

	t.Run("manage documents", func(t *testing.T) {

	})
	//TODO set auth for runtime

}

func TestRuntimeCreateUpdateAndDelete(t *testing.T) {
	// TODO
	// GIVEN
	ctx := context.Background()
	givenInput := graphql.RuntimeInput{
		Name:        "runtime-1",
		Description: ptrString("runtime-1-description"),
		Labels:      &graphql.Labels{"ggg": []string{"hhh"}}, // TODO label-1 does not work
		Annotations: &graphql.Annotations{"kkk": "lll"},
	}
	runtimeInGQL, err := ts.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)
	createReq := gcli.NewRequest(fmt.Sprintf(`mutation {result: createRuntime(in: %s) {%s} }`, runtimeInGQL, ts.fieldsProvider.ForRuntime()))
	storeExampleQuery(t, createReq.Query(), "create runtime")
	actualRuntime := graphql.Runtime{}
	resp := resultMapperFor(&actualRuntime)
	// WHEN
	err = ts.cli.Run(ctx, createReq, &resp)
	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, actualRuntime.ID)
	assert.Equal(t, givenInput.Name, actualRuntime.Name)
	assert.Equal(t, *givenInput.Description, *actualRuntime.Description)
	assert.Equal(t, *givenInput.Labels, actualRuntime.Labels)
	assert.Equal(t, *givenInput.Annotations, actualRuntime.Annotations)

	// update runtime
	givenInput.Description = ptrString("modified-runtime-1-description")
	runtimeInGQL, err = ts.graphqlizer.RuntimeInputToGQL(givenInput)
	require.NoError(t, err)

	updateReq := gcli.NewRequest(fmt.Sprintf(`mutation{result: updateRuntime(id: "%s", in: %s) {%s} }`, actualRuntime.ID, runtimeInGQL, ts.fieldsProvider.ForRuntime()))
	storeExampleQuery(t, updateReq.Query(), "update runtime")
	err = ts.cli.Run(ctx, updateReq, &resp)
	require.NoError(t, err)
	assert.Equal(t, *givenInput.Description, *actualRuntime.Description)

	// delete runtime
	delReq := gcli.NewRequest(fmt.Sprintf(`mutation{result: deleteRuntime(id: "%s") {%s}}`, actualRuntime.ID, ts.fieldsProvider.ForRuntime()))
	storeExampleQuery(t, delReq.Query(), "delete runtime")
	err = ts.cli.Run(ctx, delReq, &resp)
	require.NoError(t, err)
}

func TestQueryApplications(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		in := graphql.ApplicationInput{
			Name: fmt.Sprintf("app-%d", i),
		}

		// WHEN
		appInputGQL, err := ts.graphqlizer.ApplicationInputToGQL(in)
		require.NoError(t, err)
		request := gcli.NewRequest(
			fmt.Sprintf(`mutation {
			result: createApplication(in: %s) {%s}
			}`, appInputGQL, ts.fieldsProvider.ForApplication()))

		err = ts.cli.Run(ctx, request, nil)
		require.NoError(t, err)
	}

	queryReq := gcli.NewRequest(fmt.Sprintf(`query {result: applications {%s}}`, ts.fieldsProvider.Page(ts.fieldsProvider.ForApplication())))
	actualAppPage := graphql.ApplicationPage{}
	result := resultMapperFor(&actualAppPage)
	err := ts.cli.Run(ctx, queryReq, &result)
	require.NoError(t,err)
	fmt.Println(actualAppPage)


}

func TestQuerySpecificApplication(t *testing.T) {
	// TODO
}

func TestQueryRuntimes(t *testing.T) {
	// TODO
}

func TestQuerySpecificRuntime(t *testing.T) {
	// TODO
}

func TestTenantSeparation(t *testing.T) {
	//TODO
}

func getApp(ctx context.Context, t *testing.T, id string, cli *gcli.Client) graphql.Application {
	q := gcli.NewRequest(fmt.Sprintf(`query {result: application(id: "%s") {%s} }`, id, ts.fieldsProvider.ForApplication()))
	var app graphql.Application
	resp := resultMapperFor(&app)
	require.NoError(t, cli.Run(ctx, q, &resp))
	return app

}

func getApplicationInput(placeholder string) graphql.ApplicationInput {
	return graphql.ApplicationInput{
		Name: placeholder,
		Documents: []*graphql.DocumentInput{{
			Title:  placeholder,
			Format: graphql.DocumentFormatMarkdown,}},
		Apis: []*graphql.APIDefinitionInput{{
			Name:      placeholder,
			TargetURL: placeholder,},},
		EventAPIs: []*graphql.EventAPIDefinitionInput{{
			Name: placeholder,
			Spec: &graphql.EventAPISpecInput{
				EventSpecType: graphql.EventAPISpecTypeAsyncAPI,
			}},},
		Webhooks: []*graphql.ApplicationWebhookInput{{
			Type: graphql.ApplicationWebhookTypeConfigurationChanged,
			URL:  placeholder,},
		},
		Labels:      &graphql.Labels{placeholder: []string{placeholder},},
		Annotations: &graphql.Annotations{placeholder: placeholder,},
	}
}

func deleteApplication(t *testing.T, cli *gcli.Client, id string) {
	req := gcli.NewRequest(fmt.Sprintf(`mutation {
		deleteApplication(id: "%s") {
			id
		}	
	}`, id))
	fmt.Println("DEL", req.Query())
	require.NoError(t, cli.Run(context.Background(), req, nil))
}

func assertApplication(t *testing.T, in graphql.ApplicationInput, actualApp graphql.Application) {
	assert.NotEmpty(t, actualApp.ID)

	assert.Equal(t, in.Name, actualApp.Name)
	assert.Equal(t, *in.Description, *actualApp.Description)
	assert.Equal(t, in.Annotations, actualApp.Annotations)
	assert.Equal(t, in.Labels, actualApp.Labels)
	assert.Equal(t, in.HealthCheckURL, actualApp.HealthCheckURL)
	assert.Len(t, actualApp.Apis.Data, len(in.Apis))
	assert.Equal(t, len(in.Apis), actualApp.Apis.TotalCount)
	assert.False(t, actualApp.Apis.PageInfo.HasNextPage)

	assert.Len(t, actualApp.EventAPIs.Data, len(in.EventAPIs))
	assert.Equal(t, len(in.EventAPIs), actualApp.EventAPIs.TotalCount)
	assert.False(t, actualApp.EventAPIs.PageInfo.HasNextPage)

	assert.Len(t, actualApp.Documents.Data, len(in.Documents))
	assert.Equal(t, len(in.Documents), actualApp.Documents.TotalCount)
	assert.False(t, actualApp.Documents.PageInfo.HasNextPage)

	assert.Len(t, actualApp.Webhooks, len(in.Webhooks))
	// TODO what about: actualApp.Status

	actApiMap := make(map[string]graphql.APIDefinition)
	expApiMap := make(map[string]graphql.APIDefinitionInput)

	for _, actApi := range actualApp.Apis.Data {
		// TODO what distinguish API??? IMO we can use `name` but we need to add such validator to our service
		actApiMap[actApi.Name] = *actApi
	}

	for _, expApi := range in.Apis {
		expApiMap[expApi.Name] = *expApi
	}

	for k, act := range actApiMap {
		exp := expApiMap[k]
		assertApiDefinition(t, exp, act)

	}
}

func assertApiDefinition(t *testing.T, exp graphql.APIDefinitionInput, act graphql.APIDefinition) {
	assert.NotEmpty(t, act.ID)

	assert.Equal(t, exp.Name, act.Name)
	assert.Equal(t, exp.Description, act.Description)
	assert.Equal(t, exp.Group, act.Group)
	assert.Equal(t, exp.TargetURL, act.TargetURL)

}

func storeExampleQuery(t *testing.T, query string, exampleName string) {
	t.Helper()
	//t.Log(query)
	sanitizedName := strings.Replace(exampleName, " ", "-", -1)
	err := ioutil.WriteFile(fmt.Sprintf("%s/src/github.com/kyma-incubator/compass/examples/%s.graphql", os.Getenv("GOPATH"), sanitizedName), []byte(query), 0660)
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

// TODO: to test:
// - cannot specify basic and auth at the same time
// specify label created-by

type testSuite struct {
	graphqlizer    graphqlizer
	fieldsProvider fieldsProvider
	cli            *gcli.Client
}

func getDirectorURL() string {
	return "http://127.0.0.1:3000/graphql"
}

func resultMapperFor(target interface{}) genericGQLResponse {
	return genericGQLResponse{
		Result: target,
	}
}

type genericGQLResponse struct {
	Result interface{} `json:"result"`
}
