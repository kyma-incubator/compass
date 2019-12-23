package oauth

import (
	"encoding/json"
	"fmt"
	"github.com/kubernetes/client-go/kubernetes/typed/core/v1"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"net/url"
	"strings"
	"time"
)

//go:generate mockery -name=Client
type Client interface {
	GetAuthorizationToken() (Token, error)
}

type oauthClient struct {
	httpClient    *http.Client
	secretsClient v1.SecretInterface
	secretName    string
}

func NewOauthClient(client *http.Client, secrets v1.SecretInterface, secretName string) Client {
	return &oauthClient{
		httpClient:    client,
		secretsClient: secrets,
		secretName:    secretName,
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

	return credentials{
		clientID:       string(secret.Data[clientIDKey]),
		clientSecret:   string(secret.Data[clientSecretKey]),
		tokensEndpoint: string(secret.Data[tokensEndpointKey]),
	}, nil
}

func (c *oauthClient) getAuthorizationToken(credentials credentials) (Token, error) {

	log.Infof("Getting authorisation token for credentials clientID %s, secret %s, from endpoint: %s", credentials.clientID, credentials.clientSecret, credentials.tokensEndpoint)

	form := url.Values{}
	form.Add(grantTypeFieldName, credentialsGrantType)
	form.Add(scopeFieldName, scopes)

	log.Infof("Generated request:%s", form.Encode())

	request, err := http.NewRequest(http.MethodPost, credentials.tokensEndpoint, strings.NewReader(form.Encode()))

	if err != nil {
		log.Errorf("Failed to create token request")
		return Token{}, err
	}

	now := time.Now().Unix()

	request.SetBasicAuth(credentials.clientID, credentials.clientSecret)
	request.Header.Set(contentTypeHeader, contentTypeApplicationURLEncoded)
	response, err := c.httpClient.Do(request)

	log.Errorf("Sent request!")

	if err != nil {
		return Token{}, err
	}

	if response.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("Get token call returned unexpected status code, %d, %s", response.StatusCode, response.Status)
	}

	body, err := ioutil.ReadAll(response.Body)

	log.Errorf("Received response body %s", body)
	defer response.Body.Close()
	if err != nil {
		return Token{}, fmt.Errorf("failed to read token response body from '%s': %s", credentials.tokensEndpoint, err.Error())
	}

	tokenResponse := Token{}

	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return Token{}, fmt.Errorf("failed to unmarshal token response body: %s", err.Error())
	}

	log.Errorf("Sucessfully unmarshal response tokens")
	log.Errorf("Access token: %s", tokenResponse.AccessToken)
	log.Errorf("Expiration: %d", tokenResponse.Expiration)

	tokenResponse.Expiration += now

	return tokenResponse, nil
}
