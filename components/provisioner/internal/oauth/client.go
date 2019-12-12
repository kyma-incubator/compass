package oauth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/kubernetes/client-go/kubernetes/typed/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"mime/multipart"
	"net/http"
)

//go:generate mockery -name=Client
type Client interface {
	GetAuthorizationToken() (Token, error)
}

type oauthClient struct {
	tokensEndpoint string
	httpClient     *http.Client
	secretsClient  v1.SecretInterface
	secretName     string
}

func NewOauthClient(hydraTokensEndpoint string, client *http.Client, secrets v1.SecretInterface, secretName string) Client {
	return &oauthClient{
		tokensEndpoint: hydraTokensEndpoint,
		httpClient:     client,
		secretsClient:  secrets,
		secretName:     secretName,
	}
}

func (c *oauthClient) GetAuthorizationToken() (Token, error) {
	credentials, err := c.getCredentials()

	if err != nil {
		return Token{}, err
	}

	return c.getAuthorizationToken(credentials)
}

func (c *oauthClient) getCredentials() (credentials, error) {
	secret, err := c.secretsClient.Get(c.secretName, metav1.GetOptions{})

	if err != nil {
		return credentials{}, err
	}

	clientID, err := decodeSecret(secret.Data[clientIDKey])
	if err != nil {
		return credentials{}, err
	}
	clientSecret, err := decodeSecret(secret.Data[clientSecretKey])
	if err != nil {
		return credentials{}, err
	}

	return credentials{
		clientID:     clientID,
		clientSecret: clientSecret,
	}, nil
}

func (c *oauthClient) getAuthorizationToken(credentials credentials) (Token, error) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)

	err := setRequiredFields(writer)

	if err != nil {
		return Token{}, err
	}

	request, err := http.NewRequest(http.MethodPost, c.tokensEndpoint, buffer)

	if err != nil {
		return Token{}, err
	}

	request.SetBasicAuth(credentials.clientID, credentials.clientSecret)

	request.Header.Set(contentTypeHeader, writer.FormDataContentType())

	response, err := c.httpClient.Do(request)

	if err != nil {
		return Token{}, err
	}

	if response.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("get token call returned unexpected status code, %d", response.StatusCode)
	}

	defer response.Body.Close()

	var tokenResponse Token

	err = json.NewDecoder(response.Body).Decode(&tokenResponse)

	if err != nil {
		return Token{}, err
	}

	return tokenResponse, nil
}

func setRequiredFields(w *multipart.Writer) error {
	defer w.Close()

	err := w.WriteField(grantTypeFieldName, credentialsGrantType)
	if err != nil {
		return err
	}
	err = w.WriteField(scopeFieldName, scopes)
	if err != nil {
		return err
	}
	return nil
}

func decodeSecret(encoded []byte) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(encoded))

	if err != nil {
		return "", err
	}

	return string(decoded), nil
}