package fixtures

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

const (
	auditlogTokenEndpoint        = "audit-log/v2/oauth/token"
	auditlogSearchEndpoint       = "audit-log/v2/configuration-changes/search"
	auditlogDeleteEndpointFormat = "audit-log/v2/configuration-changes/%s"
)

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

func GetAuditlogMockToken(t *testing.T, client *http.Client, baseURL string) Token {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", baseURL, auditlogTokenEndpoint), nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "client_id", "client_secret"))))
	resp, err := client.Do(req)
	require.NoError(t, err)

	var auditlogToken Token
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	err = json.Unmarshal(body, &auditlogToken)
	require.NoError(t, err)

	return auditlogToken
}

func SearchForAuditlogByString(t *testing.T, client *http.Client, baseURL string, auditlogToken Token, search string) []model.ConfigurationChange {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", baseURL, auditlogSearchEndpoint), nil)
	require.NoError(t, err)

	req.URL.RawQuery = fmt.Sprintf("query=%s", search)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auditlogToken.AccessToken))
	resp, err := client.Do(req)
	require.NoError(t, err)

	var auditlogs []model.ConfigurationChange
	body, err := ioutil.ReadAll(resp.Body)

	require.NoError(t, err)
	err = json.Unmarshal(body, &auditlogs)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	return auditlogs
}

func DeleteAuditlogByID(t *testing.T, client *http.Client, baseURL string, auditlogToken Token, id string) {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s%s", baseURL, fmt.Sprintf(auditlogDeleteEndpointFormat, id)), nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auditlogToken.AccessToken))
	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)
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
				}`, testctx.Tc.GQLFieldsProvider.ForTenant()))
}
