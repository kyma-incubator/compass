package tests

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/kyma-incubator/compass/tests/pkg/token"
	"github.com/stretchr/testify/require"
	"net/http"
	"sort"
	"testing"
	"time"
)

type System struct {
	Protocol     string `json:"protocol"`
	Host         string `json:"host"`
	SystemType   string `json:"type"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	SystemNumber string `json:"systemNumber"`
}

type SCC struct {
	Subaccount     string   `json:"subaccount"`
	LocationID     string   `json:"locationID"`
	ExposedSystems []System `json:"exposedSystems"`
}

type Report struct {
	ReportType string `json:"type"`
	Value      []SCC  `json:"value"`
}

func TestDeltaReport(stdT *testing.T) {
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

	t.Run("Delta report - create system", func(t *testing.T) {
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

		resp := sendRequest(t, body, "delta", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)

		validateApplication(t, app, "nonSAPsys", "http", "", expectedLabel, "reachable")
	})

	t.Run("Delta report - create systems from two sccs connected to one subaccount", func(t *testing.T) {
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

		resp := sendRequest(t, body, "delta", token)
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

	t.Run("Delta report - delete system when there are two sccs connected to one subaccount", func(t *testing.T) {
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

		resp := sendRequest(t, body, "delta", token)
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

		report = Report{
			ReportType: "notification service",
			Value: []SCC{
				{
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
				{
					Subaccount:     "08b6da37-e911-48fb-a0cb-fa635a6c4321",
					LocationID:     "loc-id",
					ExposedSystems: []System{},
				}},
		}

		body, err = json.Marshal(report)
		require.NoError(t, err)

		resp = sendRequest(t, body, "delta", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 2, len(apps))
		sort.Slice(apps, func(i, j int) bool {
			return *apps[i].Description < *apps[j].Description
		})
		validateApplication(t, apps[1], "nonSAPsys", "http", "system_updated", expectedLabel, "reachable")
		validateApplication(t, apps[0], "nonSAPsys", "http", "system_two", expectedLabelWithLocId, "unreachable")
	})

	t.Run("Delta report - update system", func(t *testing.T) {
		ctx := context.Background()

		// Register application
		appFromTmpl := createApplicationFromTemplateInput(
			"S4HANA", "description of the system", "08b6da37-e911-48fb-a0cb-fa635a6c4321", "",
			"nonSAPsys", "127.0.0.1:3000", "mail", "", "reachable")
		appFromTmplGQL, err := testctx.Tc.Graphqlizer.ApplicationFromTemplateInputToGQL(appFromTmpl)
		require.NoError(t, err)
		createAppFromTmplRequest := fixtures.FixRegisterApplicationFromTemplate(appFromTmplGQL)

		//WHEN

		outputApp := graphql.ApplicationExt{}
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

		resp := sendRequest(t, body, "delta", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		app := apps[0]

		validateApplication(t, app, "nonSAPsys", "mail", "edited", expectedLabel, "reachable")
	})

	t.Run("Delta report - delete system", func(t *testing.T) {
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

		resp := sendRequest(t, body, "delta", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err := retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		validateApplication(t, apps[0], "nonSAPsys", "mail", "description of the system", expectedLabel, "unreachable")
	})

	t.Run("Delta report - create system with systemNumber", func(t *testing.T) {
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

		resp := sendRequest(t, body, "delta", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))
		app := apps[0]
		defer fixtures.CleanupApplication(t, ctx, dexGraphQLClient, "08b6da37-e911-48fb-a0cb-fa635a6c4321", app)

		validateApplication(t, app, "nonSAPsys", "http", "", expectedLabel, "reachable")
	})

	t.Run("Delta report - update system with systemNumber", func(t *testing.T) {
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

		resp := sendRequest(t, body, "delta", token)
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

		resp = sendRequest(t, body, "delta", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Equal(t, 1, len(apps))

		app = apps[0]
		validateApplication(t, app, "nonSAPsys", "mail", "edited", expectedLabel, "reachable")
	})

	t.Run("Delta report - no systems", func(t *testing.T) {
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

		resp := sendRequest(t, body, "delta", token)
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		apps, err = retrieveApps(t, ctx, sccLabelFilter)
		require.NoError(t, err)
		require.Empty(t, apps)
	})
}

func sendRequest(t *testing.T, body []byte, reportType string, token string) *http.Response {
	buffer := bytes.NewBuffer(body)
	req, err := http.NewRequest(http.MethodPost, "https://compass-gateway-xsuaa.kyma.local/nsadapter/api/v1/notifications", buffer)
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	q.Add("reportType", reportType)
	req.URL.RawQuery = q.Encode()
	header := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", header)

	//client := gql.NewAuthorizedHTTPClient(token)
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

func createApplicationFromTemplateInput(templateName, description, subaccount, locId, systemType, host, protocol, systemNumber, systemStatus string) graphql.ApplicationFromTemplateInput {
	return graphql.ApplicationFromTemplateInput{TemplateName: templateName, Values: []*graphql.TemplateValueInput{
		{
			Placeholder: "description",
			Value:       description,
		},
		{
			Placeholder: "subaccount",
			Value:       subaccount,
		},
		{
			Placeholder: "location-id",
			Value:       locId,
		},
		{
			Placeholder: "system-type",
			Value:       systemType,
		},
		{
			Placeholder: "host",
			Value:       host,
		},
		{
			Placeholder: "protocol",
			Value:       protocol,
		},
		{
			Placeholder: "system-number",
			Value:       systemNumber,
		},
		{
			Placeholder: "system-status",
			Value:       systemStatus,
		},
	}}
}

func validateApplication(t *testing.T, app *graphql.ApplicationExt, appType, systemProtocol, description string, label map[string]interface{}, systemStatus string) {
	require.Equal(t, appType, app.Labels["applicationType"])
	require.Equal(t, systemProtocol, app.Labels["systemProtocol"])
	require.Equal(t, description, *app.Description)
	require.Equal(t, label, app.Labels["scc"])
	require.Equal(t, systemStatus, *app.SystemStatus)
}

func retrieveApps(t *testing.T, ctx context.Context, labelFilter graphql.LabelFilter) ([]*graphql.ApplicationExt, error) {
	labelFilterGQL, err := testctx.Tc.Graphqlizer.LabelFilterToGQL(labelFilter)
	require.NoError(t, err)

	query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
	applicationPage := graphql.ApplicationPageExt{}
	err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
	return applicationPage.Data, err
}
