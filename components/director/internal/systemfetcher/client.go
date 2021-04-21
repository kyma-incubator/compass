package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2/clientcredentials"
)

var (
	scopes = []string{"uaa.resource"}
)

type OAuth2Config struct {
	ClientID                  string `envconfig:"APP_CLIENT_ID"`
	ClientSecret              string `envconfig:"APP_CLIENT_SECRET"`
	OAuthTokenEndpointPattern string `envconfig:"APP_OAUTH_TOKEN_ENDPOINT"`
}

type APIConfig struct {
	Endpoint string `envconfig:"APP_LANDSCAPE_INFORMATION_ENDPOINT"`
	Path     string `envconfig:"APP_LANDSCAPE_INFORMATION_PATH"`
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

func (c *Client) FetchSystemsForTenant(tenant string) []ProductInstanceExtended {

	//
	//client := http.Client{}
	//
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

	httpClient := cfg.Client(context.Background())

	url := c.apiConfig.Endpoint + c.apiConfig.Path
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var systems []ProductInstanceExtended
	if err = json.Unmarshal(respBody, &systems); err != nil {
		log.Fatal(err)
	}

	return systems
}
