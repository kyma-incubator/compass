package external_caller

import (
	"github.com/kyma-incubator/compass/components/cim-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/pkg/errors"
	"net/http"
)

// ExternalSvcCaller is used to call external services with given authentication
//
//go:generate mockery --name=ExternalSvcCaller --output=automock --outpkg=automock --case=underscore --disable-version-string
type ExternalSvcCaller interface {
	Call(*http.Request) (*http.Response, error)
}

type Caller struct {
	httpClient  *http.Client
	credentials auth.Credentials
}

func NewCaller(httpClient *http.Client, instanceCfg config.InstanceConfig) (*Caller, error) {

	credentials := &auth.OAuthCredentials{
		ClientID:     instanceCfg.ClientID,
		ClientSecret: instanceCfg.ClientSecret,
		TokenURL:     instanceCfg.OAuthURL + instanceCfg.OAuthTokenPath,
	}

	return &Caller{
		httpClient:  httpClient,
		credentials: credentials,
	}, nil
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
