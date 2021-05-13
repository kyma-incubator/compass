package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	scopes = []string{"uaa.resource"}
)

type OAuth2Config struct {
	ClientID                  string `envconfig:"APP_OAUTH_CLIENT_ID"`
	ClientSecret              string `envconfig:"APP_OAUTH_CLIENT_SECRET"`
	OAuthTokenEndpointPattern string `envconfig:"APP_OAUTH_TOKEN_ENDPOINT_PATTERN"`
}

type APIConfig struct {
	Endpoint string `envconfig:"APP_SYSTEM_INFORMATION_ENDPOINT"`
	Path     string `envconfig:"APP_SYSTEM_INFORMATION_PATH"`
}

type Client struct {
	apiConfig    APIConfig
	oAuth2Config OAuth2Config
}

func NewClient(apiConfig APIConfig, oAuth2Config OAuth2Config) *Client {
	return &Client{
		apiConfig:    apiConfig,
		oAuth2Config: oAuth2Config,
	}
}

func (c *Client) FetchSystemsForTenant(ctx context.Context, tenant string) ([]ProductInstanceExtended, error) {
	//reqBody := url.Values{}
	//reqBody.Set("grant_type", grantType)
	//reqBody.Set("client_id", clientID)
	//reqBody.Set("scope", scope)
	//
	//req, err := http.NewRequest("POST", fmt.Sprintf(tokenURLPattern, tenant)+"?"+reqBody.Encode(), nil)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(clientID+":"+clientSecret)))
	//
	//resp, err := client.Do(req)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//respBody, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Printf("%+v\n", string(respBody))

	//TODO: See if the custom HTTP client_creds fetch above isn't a better option because this now makes new http clients on every call
	cfg := clientcredentials.Config{
		ClientID:     c.oAuth2Config.ClientID,
		ClientSecret: c.oAuth2Config.ClientSecret,
		TokenURL:     fmt.Sprintf(c.oAuth2Config.OAuthTokenEndpointPattern, tenant),
		Scopes:       scopes,
	}

	// TODO: Check token, err := cfg.Token(ctx) optimization
	httpClient := cfg.Client(ctx)

	url := c.apiConfig.Endpoint + c.apiConfig.Path
	systems, err := fetchSystemsForTenant(ctx, httpClient, url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch systems from %s", url)
	}

	if len(systems) > 0 && len(systems[0].CRMCustomerID) == 0 {
		//TODO: This filter can be made configurable
		filterQuery := fmt.Sprintf("?$filter=additionalAttributes/globalAccountId eq '%s'", tenant)
		systems, err := fetchSystemsForTenant(ctx, httpClient, url+filterQuery)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to fetch systems from %s", url)
		}

		return systems, nil
	}

	return systems, nil
}

func fetchSystemsForTenant(ctx context.Context, httpClient *http.Client, url string) ([]ProductInstanceExtended, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new HTTP request")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute HTTP request")
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.C(ctx).Println("Failed to close HTTP response body")
		}
	}()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse HTTP response body")
	}

	var systems []ProductInstanceExtended
	if err = json.Unmarshal(respBody, &systems); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal systems response")
	}

	return systems, nil
}
