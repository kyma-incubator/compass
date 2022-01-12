package tests

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"net/http"
	"sort"
	"testing"
	"time"
)

type SccKey struct {
	Subaccount string
	LocationId string
}

func TestFullReport(stdT *testing.T) {
	t := testingx.NewT(stdT)

	ctx := context.Background()
	applications := make([]*graphql.ApplicationExt, 0, 200)
	after := ""

	applicationRequest := fixtures.FixApplicationsFilteredPageableRequest(" { key: \"scc\" }", 200, after)
	applicationPage := graphql.ApplicationPageExt{}
	err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, applicationRequest, &applicationPage)
	require.NoError(t, err)
	for _, app := range applicationPage.Data {
		applications = append(applications, app)
	}

	for applicationPage.PageInfo.HasNextPage {
		err = applicationPage.PageInfo.EndCursor.UnmarshalGQL(&after)
		require.NoError(stdT, err)
		fixtures.FixApplicationsFilteredPageableRequest(" { key: \"scc\" }", 200, after)
		err := testctx.Tc.RunOperation(ctx, dexGraphQLClient, applicationRequest, &applicationPage)
		require.NoError(t, err)
		for _, app := range applicationPage.Data {
			applications = append(applications, app)
		}
	}

	keyToScc := make(map[SccKey]SCC, 100)
	for _, app := range applications {
		sccLabel, ok := app.Labels["scc"].(map[string]interface{})
		require.True(stdT, ok)
		key := SccKey{
			Subaccount: sccLabel["subaccount"].(string),
			LocationId: sccLabel["locationId"].(string),
		}

		scc, ok := keyToScc[key]
		if !ok {
			scc = SCC{
				Subaccount:     key.Subaccount,
				LocationID:     key.LocationId,
				ExposedSystems: nil,
			}
			keyToScc[key] = scc
		}

		protocol, ok := app.Labels["systemProtocol"].(string)
		require.True(stdT, ok)
		systemType, ok := app.Labels["applicationType"].(string)
		require.True(stdT, ok)

		system := System{
			Protocol:     protocol,
			Host:         sccLabel["Host"].(string),
			SystemType:   systemType,
			Description:  *app.Description,
			Status:       "", //*app.SystemStatus,
			SystemNumber: *app.SystemNumber,
		}

		scc.ExposedSystems = append(scc.ExposedSystems, system)
	}

	sccs := make([]SCC, len(keyToScc), 0)
	for _, scc := range keyToScc {
		sccs = append(sccs, scc)
	}

	baseReport := Report{
		ReportType: "notification service",
		Value:      sccs,
	}

	expectedLabel := map[string]interface{}{
		"Host":       "127.0.0.1:3000",
		"Subaccount": "08b6da37-e911-48fb-a0cb-fa635a6c4321",
		"LocationID": "",
	}

	expectedLabelWithLocId := map[string]interface{}{
		"Host":       "127.0.0.1:3000",
		"Subaccount": "08b6da37-e911-48fb-a0cb-fa635a6c4321",
		"LocationID": "loc-id",
	}

	sccLabelFilter := graphql.LabelFilter{
		Key: "scc",
	}

	claims := map[string]interface{}{
		"ns-adapter-test": "ns-adapter-flow",
		"ext_attr" : map[string]interface{}{
			"subaccountid": "08b6da37-e911-48fb-a0cb-fa635a6c4321",
		},
		"scope":           []string{},
		"tenant":          testConfig.DefaultTestTenant,
		"identity":        "nsadapter-flow-identity",
		"iss":             testConfig.ExternalServicesMockURL,
		"exp":             time.Now().Unix() + int64(time.Minute.Seconds()*10),
	}
	token := token.FromExternalServicesMock(stdT, testConfig.ExternalServicesMockURL, testConfig.ClientID, testConfig.ClientSecret, claims)

	t.Run("Full report - create system", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
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
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)

		validateApplication(t, app, "nonSAPsys", "http", "", expectedLabel, "reachable")
	})

	t.Run("Full report - create system from two sccs connected to one subaccount", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
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
			SCC{
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
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 2, len(apps))

		sort.Slice(apps, func(i, j int) bool {
			return *apps[i].Description < *apps[j].Description
		})
		appOne := apps[0]
		appTwo := apps[1]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", appOne)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", appTwo)

		validateApplication(t, appOne, "nonSAPsys", "http", "system_one", expectedLabel, "reachable")
		validateApplication(t, appTwo, "nonSAPsys", "http", "system_two", expectedLabelWithLocId, "reachable")
	})

	t.Run("Full report - delete system when there are two sccs connected to one subaccount", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
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
			SCC{
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
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 2, len(apps))

		sort.Slice(apps, func(i, j int) bool {
			return *apps[i].Description < *apps[j].Description
		})
		appOne := apps[0]
		appTwo := apps[1]

		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", appOne)
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", appTwo)

		validateApplication(t, appOne, "nonSAPsys", "http", "system_one", expectedLabel, "reachable")
		validateApplication(t, appTwo, "nonSAPsys", "http", "system_two", expectedLabelWithLocId, "reachable")

		report = baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount: "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID: "",
				ExposedSystems: []System{{
					Protocol:     "http",
					Host:         "127.0.0.1:3000",
					SystemType:   "nonSAPsys",
					Description:  "system_updated",
					Status:       "reachable",
					SystemNumber: "",
				}},
			},
			SCC{
				Subaccount:     "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID:     "loc-id",
				ExposedSystems: []System{},
			})

		body, err = json.Marshal(report)
		require.NoError(t, err)

		resp = sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 2, len(apps))
		sort.Slice(apps, func(i, j int) bool {
			return *apps[i].Description < *apps[j].Description
		})
		validateApplication(t, apps[0], "nonSAPsys", "http", "system_two", expectedLabelWithLocId, "unreachable")
		validateApplication(t, apps[1], "nonSAPsys", "http", "system_updated", expectedLabel, "reachable")
	})

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

		report := baseReport
		report.Value = append(report.Value,
			SCC{
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
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		validateApplication(t, app, "nonSAPsys", "mail", "edited", expectedLabel, "reachable")
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

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount:     "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID:     "",
				ExposedSystems: []System{},
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		validateApplication(t, apps[0], "nonSAPsys", "mail", "description of the system", expectedLabel, "unreachable")
	})

	t.Run("Full report - create system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
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
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)

		validateApplication(t, app, "nonSAPsys", "http", "", expectedLabel, "reachable")
	})

	t.Run("Full report - update system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		// Register application
		report := baseReport
		report.Value = append(report.Value,
			SCC{
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
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)

		report = baseReport
		report.Value = append(report.Value,
			SCC{
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
			})

		body, err = json.Marshal(report)
		require.NoError(t, err)

		resp = sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app = apps[0]
		validateApplication(t, app, "nonSAPsys", "mail", "edited", expectedLabel, "reachable")
	})

	t.Run("Full report - delete system for entire SCC", func(t *testing.T) {
		ctx := context.Background()

		report := baseReport
		report.Value = append(report.Value,
			SCC{
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
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)
		validateApplication(t, apps[0], "nonSAPsys", "mail", "initial description", expectedLabel, "reachable")

		report = baseReport
		body, err = json.Marshal(report)
		require.NoError(t, err)

		resp = sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		validateApplication(t, apps[0], "nonSAPsys", "mail", "initial description", expectedLabel, "unreachable")
	})

	t.Run("Full report - no systems", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount:     "08b6da37-e911-48fb-a0cb-fa635a6c4321",
				LocationID:     "",
				ExposedSystems: []System{},
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)
	})
}
