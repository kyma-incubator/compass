package tests

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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
	t.Run("Full report - create system", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
			Query: str.Ptr("$[*] ? (@ == \"scc\")"),
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
		require.Equal(t, http.StatusNoContent, resp.Status)

		applicationPage = graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		app := applicationPage.Data[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)
		require.Equal(t, "nonSAPsys", app.Labels["applicationType"])
		require.Equal(t, "http", app.Labels["systemProtocol"])
		require.Equal(t, "{\"host\":\"127.0.0.1:3000\", \"locationId\":\"\"}", app.Labels["scc"])
	})

	t.Run("Full report - update system", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
			Query: str.Ptr("$[*] ? (@ == \"scc\")"),
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		// Register application

		sccLabel := struct {
			Host       string `json:"host"`
			LocationId string `json:"locationId"`
		}{
			"127.0.0.1:3000",
			"",
		}
		in := graphql.ApplicationRegisterInput{
			Name:         "",
			ProviderName: str.Ptr("SAP"),
			Description:  str.Ptr("initial description"),
			Labels:       map[string]interface{}{"scc": sccLabel, "applicationType": "nonSAPsys", "systemProtocol": "http"},
			//TODO expose Status through GQL
		}

		application, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", in)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", &application)
		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{{
				Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID: "",
				ExposedSystems: []System{{
					Protocol:     "mail",
					Host:         "127.0.0.1:3000",
					SystemType:   "otherSAPsys",
					Description:  "edited",
					Status:       "reachable",
					SystemNumber: "",
				}},
			}},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.Status)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		app := applicationPage.Data[0]
		require.Equal(t, "otherSAPsys", app.Labels["applicationType"])
		require.Equal(t, "mail", app.Labels["systemProtocol"])
		require.Equal(t, "{\"host\":\"127.0.0.1:3000\", \"locationId\":\"\"}", app.Labels["scc"])
		require.Equal(t, "edited", app.Description)
	})

	t.Run("Full report - delete system", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
			Query: str.Ptr("$[*] ? (@ == \"scc\")"),
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		// Register application

		sccLabel := struct {
			Host       string `json:"host"`
			LocationId string `json:"locationId"`
		}{
			"127.0.0.1:3000",
			"",
		}
		in := graphql.ApplicationRegisterInput{
			Name:         "",
			ProviderName: str.Ptr("SAP"),
			Description:  str.Ptr("initial description"),
			Labels:       map[string]interface{}{"scc": sccLabel, "applicationType": "nonSAPsys", "systemProtocol": "http"},
			//TODO expose Status through GQL
		}

		application, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", in)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", &application)
		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

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
		require.Equal(t, http.StatusNoContent, resp.Status)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		//TODO expose status - require.Equal(t, "unreachable", applicationPage.Data[0].SystemStatus)
	})

	t.Run("Full report - create system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
			Query: str.Ptr("$[*] ? (@ == \"scc\")"),
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
		require.Equal(t, http.StatusNoContent, resp.Status)

		applicationPage = graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		app := applicationPage.Data[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)
		require.Equal(t, "nonSAPsys", app.Labels["applicationType"])
		require.Equal(t, "http", app.Labels["systemProtocol"])
		require.Equal(t, "{\"host\":\"127.0.0.1:3000\", \"locationId\":\"\"}", app.Labels["scc"])
	})

	t.Run("Full report - update system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
			Query: str.Ptr("$[*] ? (@ == \"scc\")"),
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		// Register application

		sccLabel := struct {
			Host       string `json:"host"`
			LocationId string `json:"locationId"`
		}{
			"127.0.0.1:3000",
			"",
		}
		in := graphql.ApplicationRegisterInput{
			Name:         "",
			ProviderName: str.Ptr("SAP"),
			Description:  str.Ptr("initial description"),
			Labels:       map[string]interface{}{"scc": sccLabel, "applicationType": "nonSAPsys", "systemProtocol": "http"},
			//TODO expose Status through GQL
		}

		application, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", in)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", &application)
		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{{
				Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID: "",
				ExposedSystems: []System{{
					Protocol:     "mail",
					Host:         "127.0.0.1:3000",
					SystemType:   "otherSAPsys",
					Description:  "edited",
					Status:       "reachable",
					SystemNumber: "sysNumber",
				}},
			}},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.Status)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		app := applicationPage.Data[0]
		require.Equal(t, "otherSAPsys", app.Labels["applicationType"])
		require.Equal(t, "mail", app.Labels["systemProtocol"])
		require.Equal(t, "{\"host\":\"127.0.0.1:3000\", \"locationId\":\"\"}", app.Labels["scc"])
		require.Equal(t, "edited", app.Description)
	})

	t.Run("Full report - delete system for entire SCC", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
			Query: str.Ptr("$[*] ? (@ == \"scc\")"),
		}

		//WHEN
		labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
		require.NoError(t, err)

		// Register application

		sccLabel := struct {
			Host       string `json:"host"`
			LocationId string `json:"locationId"`
		}{
			"127.0.0.1:3000",
			"",
		}
		in := graphql.ApplicationRegisterInput{
			Name:         "",
			ProviderName: str.Ptr("SAP"),
			Description:  str.Ptr("initial description"),
			Labels:       map[string]interface{}{"scc": sccLabel, "applicationType": "nonSAPsys", "systemProtocol": "http"},
			//TODO expose Status through GQL
		}

		application, err := fixtures.RegisterApplicationFromInput(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", in)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", &application)
		require.NoError(t, err)
		require.NotEmpty(t, application.ID)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.Status)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		//TODO expose status - require.Equal(t, "unreachable", applicationPage.Data[0].SystemStatus)
	})

	t.Run("Full report - no systems", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key:   "scc",
			Query: str.Ptr("$[*] ? (@ == \"scc\")"),
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
		require.Equal(t, http.StatusNoContent, resp.Status)

		applicationPage = graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Empty(t, applicationPage.Data)
	})
}