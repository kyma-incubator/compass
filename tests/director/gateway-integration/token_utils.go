package gateway_integration

import (
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

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

type hydraToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

const applicationScopes = "application:read application:write"
const integrationSystemScopes = "application:read application:write application_template:read application_template:write runtime:read runtime:write"

func fetchHydraAccessToken(t *testing.T, encodedCredentials string, tokenURL string, scopes string) (*hydraToken, error) {
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
	defer httpRequestBodyCloser(t, resp)

	token, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	hydraToken := hydraToken{}
	err = json.Unmarshal(token, &hydraToken)
	require.NoError(t, err)

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("response status code is %d", resp.StatusCode))
	}
	return &hydraToken, nil
}

func getAccessToken(t *testing.T, oauthCredentialData *graphql.OAuthCredentialData, scopes string) string {
	oauthCredentials := fmt.Sprintf("%s:%s", oauthCredentialData.ClientID, oauthCredentialData.ClientSecret)
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(oauthCredentials))
	hydraToken, err := fetchHydraAccessToken(t, encodedCredentials, oauthCredentialData.URL, scopes)
	require.NoError(t, err)
	return hydraToken.AccessToken
}

func httpRequestBodyCloser(t *testing.T, resp *http.Response) {
	err := resp.Body.Close()
	require.NoError(t, err)
}
