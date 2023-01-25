package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

type SccKey struct {
	Subaccount string
	LocationId string
}

const testTenant = "08b6da37-e911-48fb-a0cb-fa635a6c4321"

func TestFullReport(stdT *testing.T) {
	t := testingx.NewT(stdT)

	ctx := context.Background()
	applications := make([]*graphql.ApplicationExt, 0, 200)
	after := ""

	applicationRequest := fixtures.FixApplicationsFilteredPageableRequest(" { key: \"scc\" }", 200, after)
	applicationPage := graphql.ApplicationPageExt{}
	err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, applicationRequest, &applicationPage)
	require.NoError(t, err)
	for _, app := range applicationPage.Data {
		applications = append(applications, app)
	}

	for applicationPage.PageInfo.HasNextPage {
		err = applicationPage.PageInfo.EndCursor.UnmarshalGQL(&after)
		require.NoError(stdT, err)
		fixtures.FixApplicationsFilteredPageableRequest(" { key: \"scc\" }", 200, after)
		err := testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, applicationRequest, &applicationPage)
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
		systemType, ok := app.Labels["systemType"].(string)
		require.True(stdT, ok)

		system := System{
			Protocol:     protocol,
			Host:         sccLabel["Host"].(string),
			SystemType:   systemType,
			Description:  *app.Description,
			Status:       *app.SystemStatus,
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
		"Subaccount": testTenant,
		"LocationID": "",
	}

	expectedLabelWithLocId := map[string]interface{}{
		"Host":       "127.0.0.1:3000",
		"Subaccount": testTenant,
		"LocationID": "loc-id",
	}

	//&LabelFilter{key, &query}
	filterQueryWithLocationID := fmt.Sprintf("{\"LocationID\":\"%s\", \"Subaccount\":\"%s\"}", "loc-id", testTenant)
	sccLabelFilterWithLocationID := graphql.LabelFilter{
		Key:   "scc",
		Query: &filterQueryWithLocationID,
	}

	filterQueryWithoutLocationID := fmt.Sprintf("{\"LocationID\":\"%s\", \"Subaccount\":\"%s\"}", "", testTenant)
	sccLabelFilterWithoutLocationID := graphql.LabelFilter{
		Key:   "scc",
		Query: &filterQueryWithoutLocationID,
	}

	var token string
	if testConfig.UseClone {
		instanceName := getInstanceName(stdT)
		defer deleteClone(stdT, instanceName)
		token = getTokenFromClone(stdT, instanceName)
		token = strings.TrimPrefix(token, "Bearer ")
	} else {
		token = getTokenFromExternalSVCMock(stdT)
	}

	t.Run("Full report - create system", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount: testTenant,
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

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, app)

		validateApplication(t, app, "nonSAPsys", "http", "", expectedLabel, "reachable")
	})

	t.Run("Full report - create systems for two sccs connected to one subaccount", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Empty(t, apps)

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithLocationID)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount: testTenant,
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
				Subaccount: testTenant,
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

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		appOne := apps[0]
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, appOne)

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		appTwo := apps[0]
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, appTwo)

		validateApplication(t, appOne, "nonSAPsys", "http", "system_one", expectedLabel, "reachable")
		validateApplication(t, appTwo, "nonSAPsys", "http", "system_two", expectedLabelWithLocId, "reachable")
	})

	t.Run("Full report - delete system when there are two sccs connected to one subaccount", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Empty(t, apps)

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithLocationID)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount: testTenant,
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
				Subaccount: testTenant,
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

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		appOne := apps[0]
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, appOne)

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		appTwo := apps[0]
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, appTwo)

		validateApplication(t, appOne, "nonSAPsys", "http", "system_one", expectedLabel, "reachable")
		validateApplication(t, appTwo, "nonSAPsys", "http", "system_two", expectedLabelWithLocId, "reachable")

		report = baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount: testTenant,
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
				Subaccount:     testTenant,
				LocationID:     "loc-id",
				ExposedSystems: []System{},
			})

		body, err = json.Marshal(report)
		require.NoError(t, err)

		resp = sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		appOne = apps[0]

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		appTwo = apps[0]

		validateApplication(t, appOne, "nonSAPsys", "http", "system_updated", expectedLabel, "reachable")
		validateApplication(t, appTwo, "nonSAPsys", "http", "system_two", expectedLabelWithLocId, "unreachable")
	})

	t.Run("Full report - update system", func(t *testing.T) {
		ctx := context.Background()

		// Register application
		appFromTmpl := createApplicationFromTemplateInput(
			"on-promise-system-1", string(util.ApplicationTypeS4HANAOnPremise), "description of the system", testTenant, "",
			"nonSAPsys", "127.0.0.1:3000", "mail", "", "reachable")

		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := graphql.ApplicationExt{}
		//WHEN

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testTenant, createAppFromTmplRequest, &outputApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, &outputApp)
		require.NoError(t, err)
		require.NotEmpty(t, outputApp.ID)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount: testTenant,
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

		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		validateApplication(t, app, "nonSAPsys", "mail", "edited", expectedLabel, "reachable")
	})

	t.Run("Full report - delete system", func(t *testing.T) {
		ctx := context.Background()

		// Register application
		appFromTmpl := createApplicationFromTemplateInput(
			"on-promise-system-1", string(util.ApplicationTypeS4HANAOnPremise), "description of the system", testTenant, "",
			"nonSAPsys", "127.0.0.1:3000", "mail", "", "reachable")

		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)
		outputApp := graphql.ApplicationExt{}
		//WHEN

		err = testctx.Tc.RunOperationWithCustomTenant(ctx, certSecuredGraphQLClient, testTenant, createAppFromTmplRequest, &outputApp)
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, &outputApp)
		require.NoError(t, err)
		require.NotEmpty(t, outputApp.ID)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount:     testTenant,
				LocationID:     "",
				ExposedSystems: []System{},
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		validateApplication(t, apps[0], "nonSAPsys", "mail", "description of the system", expectedLabel, "unreachable")
	})

	t.Run("Full report - create system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount: testTenant,
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

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, app)

		validateApplication(t, app, "nonSAPsys", "http", "", expectedLabel, "reachable")
	})

	t.Run("Full report - update system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		// Register application
		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount: testTenant,
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

		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, app)

		report = baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount: testTenant,
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

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
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
				Subaccount: testTenant,
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

		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, app)
		validateApplication(t, apps[0], "nonSAPsys", "mail", "initial description", expectedLabel, "reachable")

		report = baseReport
		body, err = json.Marshal(report)
		require.NoError(t, err)

		resp = sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		validateApplication(t, apps[0], "nonSAPsys", "mail", "initial description", expectedLabel, "unreachable")
	})

	t.Run("Full report - no systems", func(t *testing.T) {
		ctx := context.Background()

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount:     testTenant,
				LocationID:     "",
				ExposedSystems: []System{},
			})

		body, err := json.Marshal(report)
		require.NoError(t, err)

		resp := sendRequest(t, body, "full", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Empty(t, apps)
	})

	t.Run("Full report with large request body", func(t *testing.T) {
		ctx := context.Background()
		appsToCreate := 110

		// Create a fake description in order to make the request body
		// larger. The whole request body must be larger than the auditlog
		// max-body-size limit, which is currently 1MB
		dummyDescription := strings.Repeat("#", 10_000)

		//WHEN
		apps, err := retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
		require.NoError(t, err)
		require.Empty(t, apps)

		report := baseReport
		report.Value = append(report.Value,
			SCC{
				Subaccount:     testTenant,
				LocationID:     "",
				ExposedSystems: []System{},
			})

		for i := 0; i < appsToCreate; i++ {
			report.Value[0].ExposedSystems = append(report.Value[0].ExposedSystems, System{
				Protocol:     "http",
				Host:         fmt.Sprintf("virtual-host-%d:3000", i),
				SystemType:   "nonSAPsys",
				Description:  dummyDescription,
				Status:       "reachable",
				SystemNumber: "",
			})
		}

		body, err := json.Marshal(report)
		require.NoError(t, err)
		log.C(ctx).Infof("Sending request with body length of %d bytes", len(body))

		defer func() {
			for {
				apps, err = retrieveApps(t, ctx, sccLabelFilterWithoutLocationID)
				if err != nil {
					log.C(ctx).Errorf("failed to clean-up after test: %v", err)
				}

				if len(apps) == 0 {
					break
				}

				log.C(ctx).Infof("Cleaning up %d applications", len(apps))
				for i := 0; i < len(apps); i++ {
					fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, testTenant, apps[i])
				}
			}

			log.C(ctx).Infof("Successfully cleaned up after test")
		}()

		resp, err := sendRequestWithTimeout(body, "full", token, 5*time.Minute)
		require.NoError(t, err, "failed to send request with a large request body")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
	})
}
