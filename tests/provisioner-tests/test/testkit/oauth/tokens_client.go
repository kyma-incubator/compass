package oauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
)

type Client struct {
	tokensEndpoint string
	credentials    Credentials
	httpClient     http.Client
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
	Expiration  int    `json:"expires_in"`
}

func NewOauthTokensClient(hydraPublicURL string, credentials Credentials) *Client {
	return &Client{
		tokensEndpoint: fmt.Sprintf("%s/oauth2/token", hydraPublicURL),
		credentials:    credentials,
		httpClient:     http.Client{},
	}
}

func (c *Client) GetAccessToken() (AccessToken, error) {
	token, err := c.getOAuthToken()
	if err != nil {
		return AccessToken{}, errors.Wrap(err, "Failed to get Access token")
	}

	return token, nil
}

func (c *Client) getOAuthToken() (AccessToken, error) {
	formValues := url.Values{
		GrantTypeFieldName: []string{CredentialsGrantType},
		ScopeFieldName:     []string{Scopes},
	}

	request, err := http.NewRequest(http.MethodPost, c.tokensEndpoint, strings.NewReader(formValues.Encode()))
	if err != nil {
		return AccessToken{}, err
	}
	request.SetBasicAuth(c.credentials.ClientID, c.credentials.ClientSecret)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return AccessToken{}, err
	}
	defer testkit.CloseReader(response.Body)

	if response.StatusCode != http.StatusOK {
		return AccessToken{}, errors.Errorf("get token call returned unexpected status code, %s. Response: %s", response.Status, testkit.DumpErrorResponse(response))
	}

	var tokenResponse AccessToken
	err = json.NewDecoder(response.Body).Decode(&tokenResponse)
	if err != nil {
		return AccessToken{}, err
	}

	return tokenResponse, nil
}
