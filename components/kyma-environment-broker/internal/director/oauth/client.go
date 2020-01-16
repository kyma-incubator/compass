package oauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"

	v1 "github.com/kubernetes/client-go/kubernetes/typed/core/v1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate mockery -name=Client
type Client interface {
	GetAuthorizationToken() (Token, error)
	WaitForCredentials() error
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

func (c *oauthClient) WaitForCredentials() error {
	err := wait.Poll(time.Second, time.Minute*3, func() (bool, error) {
		_, err := c.secretsClient.Get(c.secretName, metav1.GetOptions{})
		switch {
		case apierrors.IsNotFound(err):
			return false, nil
		case err != nil:
			return false, errors.Wrapf(err, "while waiting for secret %s", c.secretName)
		}
		return true, nil
	})

	return errors.Wrapf(err, "while waiting for secret %s", c.secretName)
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
	log.Infof("Getting authorization token for credentials to access Director from endpoint: %s", credentials.tokensEndpoint)

	form := url.Values{}
	form.Add(grantTypeFieldName, credentialsGrantType)
	form.Add(scopeFieldName, scopes)

	request, err := http.NewRequest(http.MethodPost, credentials.tokensEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		log.Errorf("Failed to create authorisation token request")
		return Token{}, err
	}

	request.SetBasicAuth(credentials.clientID, credentials.clientSecret)
	request.Header.Set(contentTypeHeader, contentTypeApplicationURLEncoded)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return Token{}, err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Warnf("Cannot close connection body")
		}
	}()

	if response.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("Get token call returned unexpected status code, %d, %s", response.StatusCode, response.Status)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return Token{}, fmt.Errorf("Failed to read token response body from '%s': %s", credentials.tokensEndpoint, err.Error())
	}

	tokenResponse := Token{}
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		return Token{}, fmt.Errorf("failed to unmarshal token response body: %s", err.Error())
	}
	log.Errorf("Successfully unmarshal response oauth token for accessing Director")
	tokenResponse.Expiration += time.Now().Unix()

	return tokenResponse, nil
}
