package fixtures

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/config"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

const timeFormat = "%d-%02d-%02dT%02d:%02d:%02d"

type Token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func FixEventDefinitionInBundleRequest(appID, bndlID, eventID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							eventDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, bndlID, eventID, testctx.Tc.GQLFieldsProvider.ForEventDefinition()))
}

func GetAuditlogToken(t require.TestingT, client *http.Client, auditlogConfig config.AuditlogConfig) Token {
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	reqBody := strings.NewReader(form.Encode())
	req, err := http.NewRequest(http.MethodPost, auditlogConfig.TokenURL+"/oauth/token", reqBody)
	require.NoError(t, err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Basic "+base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", auditlogConfig.ClientID, auditlogConfig.ClientSecret))))
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK, fmt.Sprintf("unexpected status code: expected: %d, actual: %d", http.StatusOK, resp.StatusCode))

	var auditlogToken Token
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &auditlogToken)
	require.NoError(t, err)

	return auditlogToken
}

func SearchForAuditlogByTimestampAndString(t require.TestingT, client *http.Client, auditlogConfig config.AuditlogConfig, auditlogToken Token, search string, timeFrom, timeTo time.Time) []model.ConfigurationChange {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", auditlogConfig.ManagementURL, auditlogConfig.ManagementAPIPath), nil)
	require.NoError(t, err)

	timeFromStr := fmt.Sprintf(timeFormat,
		timeFrom.Year(), timeFrom.Month(), timeFrom.Day(),
		timeFrom.Hour(), timeFrom.Minute(), timeFrom.Second())

	timeToStr := fmt.Sprintf(timeFormat,
		timeTo.Year(), timeTo.Month(), timeTo.Day(),
		timeTo.Hour(), timeTo.Minute(), timeTo.Second())

	req.URL.RawQuery = fmt.Sprintf("time_from=%s&time_to=%s", timeFromStr, timeToStr)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auditlogToken.AccessToken))
	fmt.Println("Requesting: ", fmt.Sprintf("%s%s", auditlogConfig.ManagementURL, auditlogConfig.ManagementAPIPath))
	fmt.Println(req)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, resp.StatusCode, http.StatusOK, fmt.Sprintf("unexpected status code: expected: %d, actual: %d", http.StatusOK, resp.StatusCode))

	type configurationChange struct {
		model.ConfigurationChange
		Message *string `json:"message"`
	}

	var auditlogs []configurationChange
	body, err := ioutil.ReadAll(resp.Body)

	require.NoError(t, err)
	err = json.Unmarshal(body, &auditlogs)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var matchingAuditlogs []model.ConfigurationChange
	for i := range auditlogs {

		// Our productive & mocked auditlog logic is all based on the the model.ConfigurationChange struct, which doesn't contain
		// the new Message attribute which is part of the payload when using the real Auditlog Management Read API.
		// This is why when running the e2e tests on real env we need to adapt and populate the existing model.ConfigurationChange struct
		// properties Attributes & Object with the ones contained in the new Message attribute.
		if auditlogs[i].Message != nil {
			message := struct {
				Attributes []model.Attribute `json:"attributes"`
				Object     model.Object      `json:"object"`
			}{}

			err := json.Unmarshal([]byte(*auditlogs[i].Message), &message)
			require.NoError(t, err)

			auditlogs[i].Attributes = message.Attributes
			auditlogs[i].Object = message.Object

			require.NoError(t, err)
		}
		for _, attribute := range auditlogs[i].Attributes {
			if strings.Contains(attribute.New, search) {
				matchingAuditlogs = append(matchingAuditlogs, auditlogs[i].ConfigurationChange)
			}
		}
	}

	return matchingAuditlogs
}

func FixGetViewerRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: viewer {
					%s
				}
			}`,
			testctx.Tc.GQLFieldsProvider.ForViewer()))
}

func FixDeleteDocumentRequest(docID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteDocument(id: "%s") {
					id
				}
			}`, docID))
}

func FixTenantsRequest() *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: tenants {
						%s
					}
				}`, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.OmitForTenant([]string{"labels", "initialized"}))))
}

func FixTenantsPageRequest(first int) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: tenants(first: %d) {
						%s
					}
				}`, first, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.OmitForTenant([]string{"labels", "initialized"}))))
}

func FixTenantsSearchRequest(searchTerm string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: tenants(searchTerm: "%s") {
						%s
					}
				}`, searchTerm, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.OmitForTenant([]string{"labels", "initialized"}))))
}

func FixTenantsPageSearchRequest(searchTerm string, first int) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: tenants(searchTerm: "%s", first: %d) {
						%s
					}
				}`, searchTerm, first, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.OmitForTenant([]string{"labels", "initialized"}))))
}

func FixTenantRequest(externalID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
				result: tenantByExternalID(id: "%s") {
						%s
					}
				}`, externalID, testctx.Tc.GQLFieldsProvider.ForTenant()))
}

func FixWriteTenantsRequest(t require.TestingT, tenants []graphql.BusinessTenantMappingInput) *gcli.Request {
	gqlizer := graphqlizer.Graphqlizer{}
	in, err := gqlizer.WriteTenantsInputToGQL(tenants)
	require.NoError(t, err)

	tenantsQuery := fmt.Sprintf("mutation { writeTenants(in:[%s])}", in)
	return gcli.NewRequest(tenantsQuery)
}

func FixDeleteTenantsRequest(t require.TestingT, tenants []graphql.BusinessTenantMappingInput) *gcli.Request {
	gqlizer := graphqlizer.Graphqlizer{}
	in, err := gqlizer.DeleteTenantsInputToGQL(tenants)
	require.NoError(t, err)

	tenantsQuery := fmt.Sprintf("mutation { deleteTenants(in:[%s])}", in)
	return gcli.NewRequest(tenantsQuery)
}
