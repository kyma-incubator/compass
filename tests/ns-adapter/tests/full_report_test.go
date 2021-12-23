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

	expectedLabelWithLocId := map[string]interface{}{
		"Host":       "127.0.0.1:3000",
		"Subaccount": "08b6da37-e911-48fb-a0cb-fa635a6c4321",
		"LocationId": "loc-id",
	}

	sccLabelFilter := graphql.LabelFilter{
		Key: "scc",
	}

	t.Run("Full report - create system", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)

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

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)

		validateApplication(t, app, "nonSAPsys", "http", "", expectedLabel)
	})

	t.Run("Full report - create system from two sccs connected to one subaccount", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := Report{
			ReportType: "notification service",
			Value: []SCC{
				{
					Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
					LocationID: "",
					ExposedSystems: []System{{
						Protocol:     "http",
						Host:         "127.0.0.1:3000",
						SystemType:   "nonSAPsys",
						Description:  "system_one",
						Status:       "reachable",
						SystemNumber: "",
					}},
				},
				{
					Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
					LocationID: "loc-id",
					ExposedSystems: []System{{
						Protocol:     "http",
						Host:         "127.0.0.1:3000",
						SystemType:   "nonSAPsys",
						Description:  "system_two",
						Status:       "reachable",
						SystemNumber: "",
					}},
				}},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 2, len(apps))

		appOne := apps[0]
		appTwo := apps[1]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", appOne)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", appTwo)

		validateApplication(t, appOne, "nonSAPsys", "http", "system_one", expectedLabel)
		validateApplication(t, appTwo, "nonSAPsys", "http", "system_two", expectedLabelWithLocId)
	}) //TODO add more tests if this one pass

	t.Run("Full report - update system", func(t *testing.T) {
		ctx := context.Background()

		// Register application
		appFromTmpl := createApplicationFromTemplateInput(
			"S4HANA", "description of the system", "08b6da37-e911-48fb-a0cb-fa635a6c4321", "",
			"nonSAPsys", "127.0.0.1:3000", "mail", "", "reachable")

		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := graphql.ApplicationExt{}
		//WHEN

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", createAppFromTmplRequest, &outputApp)
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

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		validateApplication(t, app, "nonSAPsys", "mail", "edited", expectedLabel)
	})

	t.Run("Full report - delete system", func(t *testing.T) {
		ctx := context.Background()

		// Register application
		appFromTmpl := createApplicationFromTemplateInput(
			"S4HANA", "description of the system", "08b6da37-e911-48fb-a0cb-fa635a6c4321", "",
			"nonSAPsys", "127.0.0.1:3000", "mail", "", "reachable")

		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := graphql.ApplicationExt{}
		//WHEN

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", createAppFromTmplRequest, &outputApp)
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

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
	})

	t.Run("Full report - create system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)

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

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)

		validateApplication(t, app, "nonSAPsys", "http", "", expectedLabel)
	})

	t.Run("Full report - update system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

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

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
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

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app = apps[0]
		validateApplication(t, app, "nonSAPsys", "mail", "edited", expectedLabel)
	})

	t.Run("Full report - delete system for entire SCC", func(t *testing.T) {
		ctx := context.Background()

		// Register application
		appFromTmpl := createApplicationFromTemplateInput(
			"S4HANA", "description of the system", "08b6da37-e911-48fb-a0cb-fa635a6c4321", "",
			"nonSAPsys", "127.0.0.1:3000", "mail", "", "reachable")

		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := graphql.ApplicationExt{}
		//WHEN

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", createAppFromTmplRequest, &outputApp)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", &outputApp)
		require.NoError(t, err)
		require.NotEmpty(t, outputApp.ID)

		report := Report{
			ReportType: "notification service",
			Value:      []SCC{},
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
	})

	t.Run("Full report - no systems", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)

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

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)
	})
}
