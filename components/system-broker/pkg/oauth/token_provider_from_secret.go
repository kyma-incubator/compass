package oauth

import (
	"context"
	"encoding/json"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/pkg/errors"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"net/http"
	"net/url"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	k8scfg "sigs.k8s.io/controller-runtime/pkg/client/config"
	"strings"
	"time"
)

type OAuthTokenProvider struct {
	httpClient      *http.Client
	requestProvider RequestProvider
	k8sClient       client.Client
	secretName      string
	secretNamespace string
}

type credentials struct {
	clientID       string
	clientSecret   string
	tokensEndpoint string
}

func NewTokenProviderFromSecret(config *Config, httpClient *http.Client, requestProvider RequestProvider) (*OAuthTokenProvider, error) {
	k8sClient, err := prepareK8sClient()
	if err != nil {
		return nil, err
	}
	return &OAuthTokenProvider{
		httpClient:      httpClient,
		requestProvider: requestProvider,
		k8sClient:       k8sClient,
		secretName:      config.SecretName,
		secretNamespace: config.SecretNamespace,
	}, nil
}

func (c *OAuthTokenProvider) GetAuthorizationToken(ctx context.Context) (httputils.Token, error) {
	credentials, err := c.extractOAuthClientFromSecret(ctx)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "while get credentials from secret")
	}

	return c.getAuthorizationToken(ctx, credentials)
}

func (c *OAuthTokenProvider) WaitForCredentials(ctx context.Context) error {
	err := wait.Poll(time.Second, time.Minute*3, func() (bool, error) {
		secret := &v1.Secret{}
		err := c.k8sClient.Get(ctx, client.ObjectKey{
			Namespace: c.secretNamespace,
			Name:      c.secretName,
		}, secret)
		// it fails on connection-refused error on first call and it restarts our application.
		if err != nil {
			log.C(ctx).Warnf("secret %s not found", c.secretName)
			return false, nil
		}
		return true, nil
	})

	return errors.Wrapf(err, "while waiting for secret %s", c.secretName)
}

func (c *OAuthTokenProvider) extractOAuthClientFromSecret(ctx context.Context) (credentials, error) {
	secret := &v1.Secret{}
	err := c.k8sClient.Get(ctx, client.ObjectKey{
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

func (c *OAuthTokenProvider) getAuthorizationToken(ctx context.Context, credentials credentials) (httputils.Token, error) {
	log.C(ctx).Infof("Getting authorization token from endpoint: %s", credentials.tokensEndpoint)

	form := url.Values{}
	form.Add(grantTypeFieldName, credentialsGrantType)
	form.Add(scopeFieldName, scopes)
	body := strings.NewReader(form.Encode())
	request, err := http.NewRequest(http.MethodPost, credentials.tokensEndpoint, body)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "Failed to create authorisation token request")
	}

	//we can use a request provider or maybe its an overkill (reason for making it was correlation ids but then i moved them to a transport)
	//input := httputils.RequestInput{
	//	Method:  http.MethodPost,
	//	URL:     credentials.tokensEndpoint,
	//	Body:    body,
	//	Headers: headers,
	//}
	//
	//log.C(ctx).Errorf("%+v", input)
	//request, err := c.requestProvider.Provide(ctx, input)
	//if err != nil {
	//	return httputils.Token{}, errors.Wrap(err, "while creating authorisation token request")
	//}

	request.SetBasicAuth(credentials.clientID, credentials.clientSecret)
	request.Header.Set(contentTypeHeader, contentTypeApplicationURLEncoded)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "while send request to token endpoint")
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.C(ctx).Warn("Cannot close connection body inside oauth client")
		}
	}()

	respBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return httputils.Token{}, errors.Wrapf(err, "while reading token response body from %q", credentials.tokensEndpoint)
	}

	tokenResponse := httputils.Token{}
	err = json.Unmarshal(respBody, &tokenResponse)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "while unmarshalling token response body")
	}

	if tokenResponse.AccessToken == "" {
		return httputils.Token{}, errors.New("while fetching token: access token from oauth client is empty")
	}

	log.C(ctx).Info("Successfully unmarshal response oauth token for accessing Director")
	tokenResponse.Expiration += time.Now().Unix()

	return tokenResponse, nil
}

func prepareK8sClient() (client.Client, error) {
	k8sCfg, err := k8scfg.GetConfig()
	if err != nil {
		return nil, err
	}

	mapper, err := apiutil.NewDiscoveryRESTMapper(k8sCfg)
	if err != nil {
		err = wait.Poll(time.Second, time.Minute, func() (bool, error) {
			mapper, err = apiutil.NewDiscoveryRESTMapper(k8sCfg)
			if err != nil {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			return nil, errors.Wrap(err, "while waiting for client mapper")
		}
	}

	cli, err := client.New(k8sCfg, client.Options{Mapper: mapper})
	if err != nil {
		return nil, errors.Wrap(err, "while creating a client")
	}

	return cli, nil
}
