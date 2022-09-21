/*
* Copyright 2020 The Compass Authors
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"

	httpdirector "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/pkg/errors"
)

// CertificateCache missing godoc
//go:generate mockery --name=CertificateCache --output=automock --outpkg=automock --case=underscore --disable-version-string
type CertificateCache interface {
	Get() []*tls.Certificate
}

// MtlsClientCreator is a constructor function for http.Clients
type MtlsClientCreator func(cache CertificateCache, skipSSLValidation bool, timeout time.Duration) *http.Client

// DefaultMtlsClientCreator is the default http client creator
func DefaultMtlsClientCreator(cc CertificateCache, skipSSLValidation bool, timeout time.Duration) *http.Client {
	httpTransport := httpdirector.NewCorrelationIDTransport(httpdirector.NewHTTPTransportWrapper(&http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
			GetClientCertificate: func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
				return cc.Get()[0], nil
			},
		},
	}))

	return &http.Client{
		Transport: httpTransport,
		Timeout:   timeout,
	}
}

// mtlsTokenAuthorizationProvider presents a AuthorizationProvider implementation which crafts OAuth Bearer token values for the Authorization header using mtls http client
type mtlsTokenAuthorizationProvider struct {
	httpClient *http.Client
}

// NewMtlsTokenAuthorizationProvider constructs an TokenAuthorizationProvider
func NewMtlsTokenAuthorizationProvider(oauthCfg oauth.Config, cache CertificateCache, creator MtlsClientCreator) *mtlsTokenAuthorizationProvider {
	return &mtlsTokenAuthorizationProvider{
		httpClient: creator(cache, oauthCfg.SkipSSLValidation, oauthCfg.TokenRequestTimeout),
	}
}

// NewMtlsTokenAuthorizationProviderWithClient constructs an TokenAuthorizationProvider using the provided mtls client
func NewMtlsTokenAuthorizationProviderWithClient(client *http.Client) *mtlsTokenAuthorizationProvider {
	return &mtlsTokenAuthorizationProvider{
		httpClient: client,
	}
}

// Name specifies the name of the AuthorizationProvider
func (p *mtlsTokenAuthorizationProvider) Name() string {
	return "MtlsTokenAuthorizationProvider"
}

// Matches contains the logic for matching the AuthorizationProvider
func (p *mtlsTokenAuthorizationProvider) Matches(ctx context.Context) bool {
	credentials, err := LoadFromContext(ctx)
	if err != nil {
		return false
	}

	return credentials.Type() == OAuthMtlsCredentialType
}

// GetAuthorization crafts an OAuth Bearer token to inject as part of the executing request
func (p *mtlsTokenAuthorizationProvider) GetAuthorization(ctx context.Context) (string, error) {
	log.C(ctx).Debug("Getting new token...")

	credentials, err := LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	mtlsCredentials, ok := credentials.Get().(*OAuthMtlsCredentials)
	if !ok {
		return "", errors.New("failed to cast credentials to mtls oauth credentials type")
	}

	token, err := p.getToken(ctx, mtlsCredentials)
	if err != nil {
		return "", err
	}

	return "Bearer " + token.AccessToken, nil
}

func (p *mtlsTokenAuthorizationProvider) getToken(ctx context.Context, credentials *OAuthMtlsCredentials) (httputils.Token, error) {
	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("client_id", credentials.ClientID)
	form.Add("scope", credentials.Scopes)

	body := strings.NewReader(form.Encode())
	request, err := http.NewRequest(http.MethodPost, credentials.TokenURL, body)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "Failed to create authorisation token request")
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if credentials.AdditionalHeaders != nil {
		for headerName, headerValue := range credentials.AdditionalHeaders {
			request.Header.Set(headerName, headerValue)
		}
	}

	response, err := p.httpClient.Do(request)
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
		return httputils.Token{}, errors.Wrapf(err, "while reading token response body from %q", credentials.TokenURL)
	}

	if response.StatusCode != http.StatusOK {
		return httputils.Token{}, errors.Wrapf(err, "oauth server returned unexpected status code %d and body: %s", response.StatusCode, respBody)
	}

	tokenResponse := httputils.Token{}
	err = json.Unmarshal(respBody, &tokenResponse)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "while unmarshalling token response body")
	}

	if tokenResponse.AccessToken == "" {
		return httputils.Token{}, errors.New("while fetching token: access token from oauth client is empty")
	}

	log.C(ctx).Debug("Successfully unmarshal response oauth token")
	return tokenResponse, nil
}
