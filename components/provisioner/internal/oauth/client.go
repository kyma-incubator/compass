package oauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
)

type Client interface {
	GetAuthorizationToken(credentials Credentials) (TokenResponse, error)
}

type oauthClient struct {
	tokensEndpoint  string
	httpClient      *http.Client
}

func NewOauthClient(hydraPublicURL string, client *http.Client) Client {
	return &oauthClient{
		tokensEndpoint:  fmt.Sprintf("%s/oauth2/token", hydraPublicURL),
		httpClient:      client,
	}
}

func (c *oauthClient) GetAuthorizationToken(credentials Credentials) (TokenResponse, error) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)

	err := setRequiredFields(writer)

	if err != nil {
		return TokenResponse{}, err
	}

	request, err := http.NewRequest(http.MethodPost, c.tokensEndpoint, buffer)

	if err != nil {
		return TokenResponse{}, err
	}

	request.SetBasicAuth(credentials.ClientID, credentials.ClientSecret)

	request.Header.Set(ContentTypeHeader, writer.FormDataContentType())

	response, err := c.httpClient.Do(request)

	if err != nil {
		return TokenResponse{}, err
	}

	if response.StatusCode != http.StatusOK {
		return TokenResponse{}, fmt.Errorf("get token call returned unexpected status code, %d", response.StatusCode)
	}

	defer response.Body.Close()

	var tokenResponse TokenResponse

	err = json.NewDecoder(response.Body).Decode(&tokenResponse)

	if err != nil {
		return TokenResponse{}, err
	}

	return tokenResponse, nil
}

func setRequiredFields(w *multipart.Writer) error {
	defer w.Close()

	err := w.WriteField(GrantTypeFieldName, CredentialsGrantType)
	if err != nil {
		return err
	}
	err = w.WriteField(ScopeFieldName, Scopes)
	if err != nil {
		return err
	}
	return nil
}


