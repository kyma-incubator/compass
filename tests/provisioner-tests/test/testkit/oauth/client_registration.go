package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
	"github.com/pkg/errors"
)

type Credentials struct {
	ClientID     string
	ClientSecret string
}

type oauthClient struct {
	ClientID      string   `json:"client_id,omitempty"`
	Secret        string   `json:"client_secret,omitempty"`
	GrantTypes    []string `json:"grant_types"`
	ResponseTypes []string `json:"response_types,omitempty"`
	Scope         string   `json:"scope"`
	Owner         string   `json:"owner"`
}

type ClientManager struct {
	hydraClientsURL string
	httpClient      *http.Client
}

func NewClientManager(hydraAdminURL string) ClientManager {
	return ClientManager{
		hydraClientsURL: hydraAdminURL + "/clients",
		httpClient:      &http.Client{},
	}
}

func (c ClientManager) RegisterClient() (Credentials, error) {
	clientCredentials, err := c.registerOAuth2Client()
	if err != nil {
		return Credentials{}, errors.Wrap(err, "Failed to register Oauth2 client")
	}

	return clientCredentials, nil
}

func (c ClientManager) registerOAuth2Client() (Credentials, error) {
	oauthClient := oauthClient{
		GrantTypes: []string{CredentialsGrantType},
		Scope:      Scopes,
	}

	registerClientBody, err := json.Marshal(oauthClient)
	if err != nil {
		return Credentials{}, err
	}

	request, err := http.NewRequest(http.MethodPost, c.hydraClientsURL, bytes.NewBuffer(registerClientBody))
	if err != nil {
		return Credentials{}, err
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Credentials{}, err
	}
	defer testkit.CloseReader(response.Body)

	if response.StatusCode != http.StatusCreated {
		return Credentials{}, errors.Errorf("create OAuth2 client call returned unexpected status code, %s. Response: %s", response.Status, testkit.DumpErrorResponse(response))
	}

	err = json.NewDecoder(response.Body).Decode(&oauthClient)
	if err != nil {
		return Credentials{}, err
	}

	return Credentials{
		ClientID:     oauthClient.ClientID,
		ClientSecret: oauthClient.Secret,
	}, nil
}

func (c ClientManager) RemoveClient(clientId string) error {
	url := fmt.Sprintf("%s/%s", c.hydraClientsURL, clientId)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return errors.WithMessagef(err, "Failed to delete %s OAuth client", clientId)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return errors.WithMessagef(err, "Failed to delete %s OAuth client", clientId)
	}
	defer testkit.CloseReader(response.Body)

	if response.StatusCode != http.StatusNoContent {
		return errors.Errorf("Failed to delete %s OAuth client, service responded with unexpected status %s", clientId, response.Status)
	}

	return nil
}
