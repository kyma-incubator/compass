package securehttp

import (
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	director_http "github.com/kyma-incubator/compass/components/director/pkg/http"
)

// CallerConfig holds the configuration for Caller
type CallerConfig struct {
	Credentials   auth.Credentials
	ClientTimeout time.Duration

	SkipSSLValidation bool
	Cache             certloader.Cache
}

// Caller can be used to call secured http endpoints with given credentials
type Caller struct {
	Credentials auth.Credentials

	Provider director_http.AuthorizationProvider
	client   *http.Client
}

// NewCaller creates a new Caller
func NewCaller(config CallerConfig) (*Caller, error) {
	c := &Caller{
		Credentials: config.Credentials,
		client:      &http.Client{Timeout: config.ClientTimeout},
	}

	switch config.Credentials.Type() {
	case auth.BasicCredentialType:
		c.Provider = auth.NewBasicAuthorizationProvider()
	case auth.OAuthCredentialType:
		c.Provider = auth.NewTokenAuthorizationProvider(&http.Client{Timeout: config.ClientTimeout})
	case auth.OAuthMtlsCredentialType:
		oauthCfg := oauth.Config{
			TokenRequestTimeout: config.ClientTimeout,
			SkipSSLValidation:   config.SkipSSLValidation,
		}
		credentials, ok := config.Credentials.Get().(*auth.OAuthMtlsCredentials)
		if !ok {
			return nil, errors.New("failed to cast credentials to mtls oauth credentials type")
		}
		c.Provider = auth.NewMtlsTokenAuthorizationProvider(oauthCfg, credentials.CertCache, auth.DefaultMtlsClientCreator)
	}
	c.client.Transport = director_http.NewCorrelationIDTransport(director_http.NewSecuredTransport(http.DefaultTransport, c.Provider))
	return c, nil
}

// Call executes a http call with the configured credentials
func (c *Caller) Call(req *http.Request) (*http.Response, error) {
	req = c.addCredentialsToContext(req)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "while executing call to %s: ", req.URL)
	}
	return resp, nil
}

func (c *Caller) addCredentialsToContext(req *http.Request) *http.Request {
	authCtx := req.Context()
	authCtx = auth.SaveToContext(authCtx, c.Credentials)
	return req.WithContext(authCtx)
}
