package external_services_mock_integration

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/retrier"

	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

const (
	auditlogTokenEndpoint  = "audit-log/v2/oauth/token"
	auditlogSearchEndpoint = "audit-log/v2/configuration-changes/search"
)

type token struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Application
func fixRegisterApplicationRequest(applicationInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: registerApplication(in: %s) {
					%s
				}
			}`,
			applicationInGQL, tc.gqlFieldsProvider.ForApplication()))
}

func fixDeleteApplicationRequest(t *testing.T, id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		unregisterApplication(id: "%s") {
			%s
		}	
	}`, id, tc.gqlFieldsProvider.ForApplication()))
}

// External services mock
func getAuditlogMockToken(t *testing.T, client *http.Client, baseURL string) token {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", baseURL, auditlogTokenEndpoint), nil)
	require.NoError(t, err)

	req.Header.Add("Authorization", base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", "client_id", "client_secret"))))

	var resp *http.Response
	err = retrier.DoOnTemporaryConnectionProblems("GetAuditLogMockTokenIntegrationTest", func() error {
		var err error
		resp, err = client.Do(req)
		return err
	})
	require.NoError(t, err)
	defer closeBody(resp.Body)

	var auditlogToken token
	require.NoError(t, err)

	err = json.NewDecoder(resp.Body).Decode(&auditlogToken)
	require.NoError(t, err)

	return auditlogToken
}

func searchForAuditlogByString(t *testing.T, client *http.Client, baseURL string, auditlogToken token, search string) []model.ConfigurationChange {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", baseURL, auditlogSearchEndpoint), nil)
	require.NoError(t, err)

	req.URL.RawQuery = fmt.Sprintf("query=%s", search)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", auditlogToken.AccessToken))

	var resp *http.Response
	err = retrier.DoOnTemporaryConnectionProblems("SearchForAuditLogIntegrationTest", func() error {
		var err error
		resp, err = client.Do(req)
		return err
	})
	require.NoError(t, err)
	defer closeBody(resp.Body)

	var auditlogs []model.ConfigurationChange

	// TEMP - START
	data, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	log.Printf("statuscode: %d, data:\n%s\n---", resp.StatusCode, string(data))
	bodyReader := ioutil.NopCloser(bytes.NewReader(data))
	defer closeBody(bodyReader)
	// TEMP - STOP

	err = json.NewDecoder(bodyReader).Decode(&auditlogs)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode)

	return auditlogs
}

//Package
func fixAddPackageRequest(appID, pkgCreateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addPackage(applicationID: "%s", in: %s) {
				%s
			}}`, appID, pkgCreateInput, tc.gqlFieldsProvider.ForPackage()))
}

func fixDeletePackageRequest(packageID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deletePackage(id: "%s") {
				%s
			}
		}`, packageID, tc.gqlFieldsProvider.ForPackage()))
}

func fixRefetchAPISpecRequest(id string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
				result: refetchAPISpec(apiID: "%s") {
						%s
					}
				}`,
			id, tc.gqlFieldsProvider.ForApiSpec()))
}

func fixPackageRequest(applicationID string, packageID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, tc.gqlFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.package": fmt.Sprintf(`package(id: "%s") {%s}`, packageID, tc.gqlFieldsProvider.ForPackage()),
		})))
}

func closeBody(body io.ReadCloser) {
	if body == nil {
		return
	}
	_, err := io.Copy(ioutil.Discard, body)
	if err != nil {
		log.Error(err)
	}

	err = body.Close()
	if err != nil {
		log.Error(err)
	}
}
