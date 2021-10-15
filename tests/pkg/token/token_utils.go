package token

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

type HydraToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

const (
	RuntimeScopes           = "runtime:read runtime:write application:read runtime.auths:read bundle.instance_auths:read"
	ApplicationScopes       = "application:read application:write application.auths:read application.webhooks:read bundle.instance_auths:read document.fetch_request:read event_spec.fetch_request:read api_spec.fetch_request:read fetch-request.auth:read"
	IntegrationSystemScopes = "application:read application:write application_template:read application_template:write runtime:read runtime:write integration_system:read label_definition:read label_definition:write automatic_scenario_assignment:read automatic_scenario_assignment:write integration_system.auths:read application_template.webhooks:read"
)

func FetchHydraAccessToken(t *testing.T, encodedCredentials string, tokenURL string, scopes string) (*HydraToken, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", scopes)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	require.NoError(t, err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encodedCredentials))

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer HttpRequestBodyCloser(t, resp)

	token, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	hydraToken := HydraToken{}
	err = json.Unmarshal(token, &hydraToken)
	require.NoError(t, err)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("response status code is %d", resp.StatusCode))
	}
	return &hydraToken, nil
}

func GetAccessToken(t *testing.T, oauthCredentialData *graphql.OAuthCredentialData, scopes string) string {
	oauthCredentials := fmt.Sprintf("%s:%s", oauthCredentialData.ClientID, oauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	hydraToken, err := FetchHydraAccessToken(t, encodedCredentials, oauthCredentialData.URL, scopes)
	require.NoError(t, err)
	return hydraToken.AccessToken
}

func HttpRequestBodyCloser(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}

func FromExternalServicesMock(t *testing.T, externalServicesMockURL string, claims map[string]interface{}) string {
	data, err := json.Marshal(claims)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, externalServicesMockURL+"/oauth/token", bytes.NewBuffer(data))
	require.NoError(t, err)

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()

	token := gjson.GetBytes(body, "access_token")
	require.True(t, token.Exists())

	return token.String()
}
