package oauth

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/sirupsen/logrus"
)

//go:generate mockery -name=Client
type Client interface {
	GetAuthorizationToken() (Token, error)
	WaitForCredentials() error
}

type oauthClient struct {
	httpClient      *http.Client
	k8sClient       client.Client
	secretName      string
	secretNamespace string
}

func NewOauthClient(httpClient *http.Client, k8sClient client.Client, secretName, secretNamespace string) Client {
	return &oauthClient{
		httpClient:      httpClient,
		k8sClient:       k8sClient,
		secretName:      secretName,
		secretNamespace: secretNamespace,
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
		secret := &v1.Secret{}
		err := c.k8sClient.Get(context.Background(), client.ObjectKey{
			Namespace: c.secretNamespace,
			Name:      c.secretName,
		}, secret)
		switch {
		case apierrors.IsNotFound(err):
			log.Warnf("secret %s not found", c.secretName)
			return false, nil
		case err != nil:
			return false, errors.Wrapf(err, "while waiting for secret %s", c.secretName)
		}
		return true, nil
	})

	return errors.Wrapf(err, "while waiting for secret %s", c.secretName)
}

func (c *oauthClient) getCredentials() (credentials, error) {
	secret := &v1.Secret{}
	err := c.k8sClient.Get(context.Background(), client.ObjectKey{
		Namespace: c.secretNamespace,
		Name:      c.secretName,
	}, secret)
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
