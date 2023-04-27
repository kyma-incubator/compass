package external_caller

import (
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/tm-adapter/internal/config"
	"github.com/pkg/errors"
	"net/http"
)

// ExternalSvcCaller is used to call external services with given authentication
//
//go:generate mockery --name=ExternalSvcCaller --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCaller interface { // todo::: improvement: the code in director could be adapted and reuse it here
	Call(*http.Request) (*http.Response, error)
}

type Caller struct {
	httpClient  *http.Client
	credentials auth.Credentials
}

func NewCaller(httpClient *http.Client, oauthConfig config.OAuthConfig) *Caller {
	credentials := &auth.OAuthCredentials{
		ClientID:     oauthConfig.ClientID,
		ClientSecret: oauthConfig.ClientSecret,
		TokenURL:     oauthConfig.OAuthURL + oauthConfig.OAuthTokenPath,
	}

	return &Caller{
		httpClient:  httpClient,
		credentials: credentials,
	}
}

func (c *Caller) Call(req *http.Request) (*http.Response, error) {
	req = c.addCredentialsToContext(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "An error occurred while executing call to %q", req.URL)
	}

	return resp, nil
}

func (c *Caller) addCredentialsToContext(req *http.Request) *http.Request {
	authCtx := req.Context()
	authCtx = auth.SaveToContext(authCtx, c.credentials)
	return req.WithContext(authCtx)
}
