package oauth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"
	"github.com/stretchr/testify/suite"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestTokenProviderFromSecretTestSuite(t *testing.T) {
	suite.Run(t, new(TokenProviderFromSecretTestSuite))
}

type TokenProviderFromSecretTestSuite struct {
	suite.Suite

	clientID      string
	clientSecret  string
	tokenEndpoint string
	accessToken   httputils.Token

	config        *oauth.Config
	k8sClientFunc func(time.Duration) (client.Client, error)
}

func (suite *TokenProviderFromSecretTestSuite) SetupTest() {
	suite.clientID = "admin_client_id"
	suite.clientSecret = "admin_client_secret"
	suite.tokenEndpoint = "http://localhost:8080/oauth/token"
	suite.accessToken = httputils.Token{AccessToken: "access_token", Expiration: 9999}

	suite.config = oauth.DefaultConfig()
	suite.k8sClientFunc = func(duration time.Duration) (client.Client, error) {
		return fake.NewFakeClientWithScheme(scheme.Scheme, &v1.Secret{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      suite.config.SecretName,
				Namespace: suite.config.SecretNamespace,
			},
			Data: map[string][]byte{
				"client_id":       []byte(suite.clientID),
				"client_secret":   []byte(suite.clientSecret),
				"tokens_endpoint": []byte(suite.tokenEndpoint),
			},
			Type: "Opaque",
		}), nil
	}
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_New_InvalidKubernetesClient() {
	httpClient := &MockClient{}

	suite.k8sClientFunc = func(duration time.Duration) (client.Client, error) {
		return nil, errors.New("error")
	}
	provider, err := oauth.NewTokenProviderFromSecret(suite.config, targetURL, httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().EqualError(err, "error")
	suite.Require().Nil(provider)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_New_InvalidURL() {
	httpClient := &MockClient{}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, "%zzz", httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().Error(err)
	suite.Require().Nil(provider)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_Name() {
	httpClient := &MockClient{}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, targetURL, httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().NoError(err)

	name := provider.Name()

	suite.Require().Equal(name, "TokenProviderFromSecret")
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_Matches() {
	httpClient := &MockClient{}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, targetURL, httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().NoError(err)

	ctx := context.TODO()
	matches := provider.Matches(ctx)
	suite.Require().Equal(matches, true)

	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "Bearer token")
	matches = provider.Matches(ctx)
	suite.Require().Equal(matches, false)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_URL() {
	httpClient := &MockClient{}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, targetURL, httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().NoError(err)

	url := provider.TargetURL()

	suite.Require().Equal(url.String(), targetURL)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_GetAuthorizationToken() {
	httpClient := &MockClient{
		DoFunc: suite.validResponseDoFunc(),
	}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, targetURL, httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().NoError(err)

	token, err := provider.GetAuthorizationToken(context.TODO())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.accessToken.AccessToken, token.AccessToken)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_GetAuthorizationTokenBasedOnTokenValidity() {
	httpClient := &MockClient{
		DoFunc: suite.validResponseDoFunc(),
	}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, targetURL, httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().NoError(err)

	// expired token
	suite.accessToken.Expiration = -1

	token, err := provider.GetAuthorizationToken(context.TODO())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.accessToken.AccessToken, token.AccessToken)
	suite.Require().Equal(token.EmptyOrExpired(time.Second), true)

	// unexpired token
	suite.accessToken.AccessToken = "new-token"
	suite.accessToken.Expiration = 9999

	token, err = provider.GetAuthorizationToken(context.TODO())
	suite.Require().NoError(err)
	suite.Require().Equal(suite.accessToken.AccessToken, token.AccessToken)
	suite.Require().Equal(token.EmptyOrExpired(time.Second), false)

	// reuse the same token
	suite.accessToken.AccessToken = "new-fake-token"

	token, err = provider.GetAuthorizationToken(context.TODO())
	suite.Require().NoError(err)
	suite.Require().Equal("new-token", token.AccessToken)
	suite.Require().Equal(token.EmptyOrExpired(time.Second), false)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_GetAuthorizationToken_ShouldReturnErrorIfEmptyBody() {
	httpClient := &MockClient{}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, targetURL, httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().NoError(err)

	token, err := provider.GetAuthorizationToken(context.TODO())
	suite.Require().Error(err)
	suite.Require().Empty(token)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_GetAuthorizationToken_ShouldReturnErrorIfReturnedTokenIsEmpty() {
	suite.accessToken.AccessToken = ""
	httpClient := &MockClient{
		DoFunc: suite.validResponseDoFunc(),
	}

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, targetURL, httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().NoError(err)

	token, err := provider.GetAuthorizationToken(context.TODO())
	suite.Require().Error(err)
	suite.Require().Empty(token)
}

func (suite *TokenProviderFromSecretTestSuite) TestOAuthTokenProvider_GetAuthorizationToken_ShouldReturnErrorWhenSecretNotFound() {
	suite.k8sClientFunc = func(time.Duration) (client.Client, error) {
		return fake.NewFakeClientWithScheme(scheme.Scheme), nil
	}
	httpClient := &MockClient{}
	suite.config.WaitSecretTimeout = 5 * time.Second

	provider, err := oauth.NewTokenProviderFromSecret(suite.config, targetURL, httpClient, time.Second, suite.k8sClientFunc)
	suite.Require().NoError(err)

	token, err := provider.GetAuthorizationToken(context.TODO())
	suite.Require().Error(err)
	suite.Require().Empty(token)
}

func (suite *TokenProviderFromSecretTestSuite) validResponseDoFunc() func(*http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		suite.Require().Equal(suite.tokenEndpoint, req.URL.String())

		username, password, ok := req.BasicAuth()
		suite.Require().True(ok, "Expected client credentials from secret to be provided as basic header")
		suite.Require().Equal(suite.clientID, username)
		suite.Require().Equal(suite.clientSecret, password)

		token := httputils.Token{
			AccessToken: suite.accessToken.AccessToken,
			Expiration:  suite.accessToken.Expiration,
		}
		jsonBytes, err := json.Marshal(token)
		suite.Require().NoError(err)

		return &http.Response{
			Body: ioutil.NopCloser(bytes.NewBuffer(jsonBytes)),
		}, nil
	}
}

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	return &http.Response{
		Body: ioutil.NopCloser(&bytes.Buffer{}),
	}, nil
}
