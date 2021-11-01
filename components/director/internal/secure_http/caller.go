package secure_http

import (
	"errors"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	director_http "github.com/kyma-incubator/compass/components/director/pkg/http"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const OauthTokenPath = "/oauth/token"

type Caller struct {
	credentials graphql.CredentialData

	provider director_http.AuthorizationProvider
	client   *http.Client
}

func NewCaller(credentials graphql.CredentialData, clientTimeout time.Duration) (*Caller, error) {
	c := &Caller{
		credentials: credentials,
		client:      &http.Client{Timeout: clientTimeout},
	}

	switch credentials.(type) {
	case *graphql.BasicCredentialData:
		c.provider = auth.NewBasicAuthorizationProvider()
	case *graphql.OAuthCredentialData:
		c.provider = auth.NewTokenAuthorizationProvider(http.DefaultClient)
	default:
		return nil, errors.New("unsupported credentials type")
	}
	c.client.Transport = director_http.NewSecuredTransport(http.DefaultTransport, c.provider)

	return c, nil
}

func (c *Caller) Call(req *http.Request) (*http.Response, error) {
	req = c.addCredentialsToContext(req)
	return c.client.Do(req)
}

func (c *Caller) addCredentialsToContext(req *http.Request) *http.Request {
	authCtx := req.Context()
	authCtx = auth.SaveToContext(authCtx, c.credentials)
	return req.WithContext(authCtx)
}
