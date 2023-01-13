package http

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const tenantHeader = "Tenant"

// GetRequestWithCredentials executes a GET http request to the given url with the provided auth credentials
func GetRequestWithCredentials(ctx context.Context, client *http.Client, url, tnt string, auth *model.Auth) (*http.Response, error) {
	if auth == nil || (auth.Credential.Basic == nil && auth.Credential.Oauth == nil) {
		return nil, apperrors.NewInvalidDataError("Credentials not provided")
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if len(tnt) > 0 {
		req.Header.Set(tenantHeader, tnt)
	}

	var resp *http.Response
	if auth.Credential.Basic != nil {
		req.SetBasicAuth(auth.Credential.Basic.Username, auth.Credential.Basic.Password)

		resp, err = client.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {
			return resp, nil
		}
	}

	if auth.Credential.Oauth != nil {
		resp, err = secureClient(ctx, client, auth).Do(req)
	}

	return resp, err
}

func secureClient(ctx context.Context, client *http.Client, auth *model.Auth) *http.Client {
	conf := &clientcredentials.Config{
		ClientID:     auth.Credential.Oauth.ClientID,
		ClientSecret: auth.Credential.Oauth.ClientSecret,
		TokenURL:     auth.Credential.Oauth.URL,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	securedClient := conf.Client(ctx)
	securedClient.Timeout = client.Timeout
	return securedClient
}

// GetRequestWithoutCredentials executes a GET http request to the given url
func GetRequestWithoutCredentials(client *http.Client, url, tnt string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if len(tnt) > 0 {
		req.Header.Set(tenantHeader, tnt)
	}

	return client.Do(req)
}
