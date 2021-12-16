package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
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

	t.Run("Delta report - create system", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key: "scc",
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

		resp := sendRequest(t, body, "delta")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)
		fmt.Printf("Response: %+v", resp)

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

	t.Run("Delta report - update system", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key: "scc",
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

		resp := sendRequest(t, body, "delta")
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

	t.Run("Delta report - delete system", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key: "scc",
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

		resp := sendRequest(t, body, "delta")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		query := fixtures.FixApplicationsFilteredPageableRequest(labelFilterGQL, 100, "")
		applicationPage := graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Equal(t, 1, len(applicationPage.Data))
		//TODO expose status - require.Equal(t, "unreachable", applicationPage.Data[0].SystemStatus)
	})

	t.Run("Delta report - create system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key: "scc",
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

		resp := sendRequest(t, body, "delta")
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

	t.Run("Delta report - update system with systemNumber", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key: "scc",
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

		resp = sendRequest(t, body, "delta")
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

	t.Run("Delta report - no systems", func(t *testing.T) {
		ctx := context.Background()

		// Query for application with LabelFilter "scc"
		labelFilter := graphql.LabelFilter{
			Key: "scc",
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

		resp := sendRequest(t, body, "delta")
		require.Equal(t, http.StatusNoContent, resp.StatusCode)

		applicationPage = graphql.ApplicationPageExt{}
		err = testctx.Tc.RunOperation(ctx, dexGraphQLClient, query, &applicationPage)
		require.NoError(t, err)
		require.Empty(t, applicationPage.Data)
	})
}

func sendRequest(t *testing.T, body []byte, reportType string) *http.Response {
	buffer := bytes.NewBuffer(body)
	req, err := http.NewRequest(http.MethodPost, "https://compass-gateway-sap-mtls.kyma.local/nsadapter/api/v1/notifications", buffer)
	if err != nil {
		panic(err)
	}
	q := req.URL.Query()
	q.Add("reportType", reportType)
	req.URL.RawQuery = q.Encode()

	clientKey, rawCertChain := certs.IssueExternalIssuerCertificate(t, testConfig.CA.Certificate, testConfig.CA.Key, "08b6da37-e911-48fb-a0cb-fa635a6c5678")
	client := gql.NewCertAuthorizedHTTPClient(clientKey, rawCertChain)

	fmt.Printf("req >>>>>>>>>>>>>>>>>>> %+v", req)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	return resp
}
