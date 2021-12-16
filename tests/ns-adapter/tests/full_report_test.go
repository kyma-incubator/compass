package tests

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"net/http"
	"testing"
)

func TestFullReport(stdT *testing.T) {
	t := testingx.NewT(stdT)

	expectedLabel := map[string]interface{}{
		"Host":       "127.0.0.1:3000",
		"Subaccount": "08b6da37-e911-48fb-a0cb-fa635a6c4321",
		"LocationId": "",
	}

	t.Run("Full report - create system", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Empty(t, applicationPage.Data)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{{
				Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID: "",
				ExposedSystems: []System{{
					Protocol:     "http",
					Host:         "127.0.0.1:3000",
					SystemType:   "nonSAPsys",
					Description:  "",
					Status:       "reachable",
					SystemNumber: "",
				}},
			}},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		applicationPage = graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		app := applicationPage.Data[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)
		require.Equal(t, "nonSAPsys", app.Labels["applicationType"])
		require.Equal(t, "http", app.Labels["systemProtocol"])
		require.Equal(t, expectedLabel, app.Labels["scc"])
	})

	t.Run("Full report - update system", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		// Register application

		appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: "S4HANA", Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "description",
				Value:       "description of the system",
			},
			{
				Placeholder: "subaccount",
				Value:       "08b6da37-e911-48fb-a0cb-fa635a6c4321",
			},
			{
				Placeholder: "location-id",
				Value:       "",
			},
			{
				Placeholder: "system-type",
				Value:       "nonSAPsys",
			},
			{
				Placeholder: "host",
				Value:       "127.0.0.1:3000",
			},
			{
				Placeholder: "protocol",
				Value:       "mail",
			},
			{
				Placeholder: "system-number",
				Value:       "",
			},
			{
				Placeholder: "system-status",
				Value:       "reachable",
			},
		}}
		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := graphql.ApplicationExt{}
		//WHEN

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient,"08b6da37-e911-48fb-a0cb-fa635a6c4321", createAppFromTmplRequest, &outputApp)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", &outputApp)
		require.NoError(t, err)
		require.NotEmpty(t, outputApp.ID)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{{
				Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID: "",
				ExposedSystems: []System{{
					Protocol:     "mail",
					Host:         "127.0.0.1:3000",
					SystemType:   "nonSAPsys",
					Description:  "edited",
					Status:       "reachable",
					SystemNumber: "",
				}},
			}},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		app := applicationPage.Data[0]
		require.Equal(t, "nonSAPsys", app.Labels["applicationType"])
		require.Equal(t, "mail", app.Labels["systemProtocol"])
		require.Equal(t, expectedLabel, app.Labels["scc"])
		require.Equal(t, "edited", *app.Description)
	})

	t.Run("Full report - delete system", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		// Register application

		appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: "S4HANA", Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "description",
				Value:       "description of the system",
			},
			{
				Placeholder: "subaccount",
				Value:       "08b6da37-e911-48fb-a0cb-fa635a6c4321",
			},
			{
				Placeholder: "location-id",
				Value:       "",
			},
			{
				Placeholder: "system-type",
				Value:       "nonSAPsys",
			},
			{
				Placeholder: "host",
				Value:       "127.0.0.1:3000",
			},
			{
				Placeholder: "protocol",
				Value:       "mail",
			},
			{
				Placeholder: "system-number",
				Value:       "",
			},
			{
				Placeholder: "system-status",
				Value:       "reachable",
			},
		}}
		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := graphql.ApplicationExt{}
		//WHEN

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient,"08b6da37-e911-48fb-a0cb-fa635a6c4321", createAppFromTmplRequest, &outputApp)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", &outputApp)
		require.NoError(t, err)
		require.NotEmpty(t, outputApp.ID)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{{
				Subaccount:     "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID:     "",
				ExposedSystems: []System{},
			}},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
	})

	t.Run("Full report - create system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Empty(t, applicationPage.Data)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{{
				Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID: "",
				ExposedSystems: []System{{
					Protocol:     "http",
					Host:         "127.0.0.1:3000",
					SystemType:   "nonSAPsys",
					Description:  "",
					Status:       "reachable",
					SystemNumber: "sysNumber",
				}},
			}},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		applicationPage = graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		app := applicationPage.Data[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)
		require.Equal(t, "nonSAPsys", app.Labels["applicationType"])
		require.Equal(t, "http", app.Labels["systemProtocol"])
		require.Equal(t, expectedLabel, app.Labels["scc"])
	})

	t.Run("Full report - update system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		// Register application
		report := Report{
			ReportType: "notification service",
			Value: []SCC{{
				Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID: "",
				ExposedSystems: []System{{
					Protocol:     "mail",
					Host:         "127.0.0.1:3000",
					SystemType:   "nonSAPsys",
					Description:  "initial description",
					Status:       "reachable",
					SystemNumber: "sysNumber",
				}},
			}},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "delta")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		app := applicationPage.Data[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)

		report = Report{
			ReportType: "notification service",
			Value: []SCC{{
				Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID: "",
				ExposedSystems: []System{{
					Protocol:     "mail",
					Host:         "127.0.0.1:3000",
					SystemType:   "nonSAPsys",
					Description:  "edited",
					Status:       "reachable",
					SystemNumber: "sysNumber",
				}},
			}},
		}

		body, err = json.Marshal(report)
		require.NoError(t, err)

		resp = sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		query = fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage = graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		app = applicationPage.Data[0]
		require.Equal(t, "nonSAPsys", app.Labels["applicationType"])
		require.Equal(t, "mail", app.Labels["systemProtocol"])
		require.Equal(t, expectedLabel, app.Labels["scc"])
		require.Equal(t, "edited", *app.Description)
	})

	t.Run("Full report - delete system for entire SCC", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		// Register application


		appFromTmpl := graphql.ApplicationFromTemplateInput{TemplateName: "S4HANA", Values: []*graphql.TemplateValueInput{
			{
				Placeholder: "description",
				Value:       "description of the system",
			},
			{
				Placeholder: "subaccount",
				Value:       "08b6da37-e911-48fb-a0cb-fa635a6c4321",
			},
			{
				Placeholder: "location-id",
				Value:       "",
			},
			{
				Placeholder: "system-type",
				Value:       "nonSAPsys",
			},
			{
				Placeholder: "host",
				Value:       "127.0.0.1:3000",
			},
			{
				Placeholder: "protocol",
				Value:       "mail",
			},
			{
				Placeholder: "system-number",
				Value:       "",
			},
			{
				Placeholder: "system-status",
				Value:       "reachable",
			},
		}}
		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := graphql.ApplicationExt{}
		//WHEN

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient,"08b6da37-e911-48fb-a0cb-fa635a6c4321", createAppFromTmplRequest, &outputApp)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", &outputApp)
		require.NoError(t, err)
		require.NotEmpty(t, outputApp.ID)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
	})

	t.Run("Full report - no systems", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Empty(t, applicationPage.Data)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{{
				Subaccount:     "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID:     "",
				ExposedSystems: []System{},
			}},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		applicationPage = graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Empty(t, applicationPage.Data)
	})
}