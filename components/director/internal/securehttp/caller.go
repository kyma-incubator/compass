package securehttp

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	director_http "github.com/kyma-incubator/compass/components/director/pkg/http"
)

// Caller can be used to call secured http endpoints with given credentials
type Caller struct {
	credentials auth.Credentials

	provider director_http.AuthorizationProvider
	client   *http.Client
}

// NewCaller creates a new Caller
func NewCaller(credentials auth.Credentials, clientTimeout time.Duration) *Caller {
	c := &Caller{
		credentials: credentials,
		client:      &http.Client{Timeout: clientTimeout},
	}

	switch credentials.Type() {
	case auth.BasicCredentialType:
		c.provider = auth.NewBasicAuthorizationProvider()
	case auth.OAuthCredentialType:
		c.provider = auth.NewTokenAuthorizationProvider(&http.Client{Timeout: clientTimeout})
	}
	c.client.Transport = director_http.NewCorrelationIDTransport(director_http.NewSecuredTransport(http.DefaultTransport, c.provider))
	return c
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
	authCtx = auth.SaveToContext(authCtx, c.credentials)
	return req.WithContext(authCtx)
}
