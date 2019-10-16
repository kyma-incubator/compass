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

func RegisterClient(oauthAdminURL string) (Credentials, error) {
	clientCredentials, err := registerOAuth2Client(fmt.Sprintf("%s/clients", oauthAdminURL))
	if err != nil {
		return Credentials{}, errors.Wrap(err, "Failed to register Oauth2 client")
	}

	return clientCredentials, nil
}

func registerOAuth2Client(oauthClientsURL string) (Credentials, error) {
	oauthClient := oauthClient{
		GrantTypes: []string{CredentialsGrantType},
		Scope:      Scopes,
	}

	registerClientBody, err := json.Marshal(oauthClient)
	if err != nil {
		return Credentials{}, err
	}

	request, err := http.NewRequest(http.MethodPost, oauthClientsURL, bytes.NewBuffer(registerClientBody))
	if err != nil {
		return Credentials{}, err
	}

	response, err := http.DefaultClient.Do(request)
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
