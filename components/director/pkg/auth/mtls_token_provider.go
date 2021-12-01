package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	oauth "github.com/kyma-incubator/compass/components/director/pkg/oauth"

	httpdirector "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/pkg/errors"
)

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

// CertificateCache missing godoc
//go:generate mockery --name=CertificateCache --output=automock --outpkg=automock --case=underscore
type CertificateCache interface {
	Get() *tls.Certificate
}

type mtlsClientCreator func(cache CertificateCache, timeout time.Duration) *http.Client

func DefaultCreator(cc CertificateCache, timeout time.Duration) *http.Client {
	httpTransport := httpdirector.NewCorrelationIDTransport(&http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			GetClientCertificate: func(_ *tls.CertificateRequestInfo) (*tls.Certificate, error) {
				return cc.Get(), nil
			},
		},
	})

	return &http.Client{
		Transport: httpTransport,
		Timeout:   timeout,
	}
}

// mtlsTokenAuthorizationProvider presents a AuthorizationProvider implementation which crafts OAuth Bearer token values for the Authorization header using mtls http client
type mtlsTokenAuthorizationProvider struct {
	clientID     string
	tokenURL     string
	scopesClaim  string
	token        httputils.Token
	tokenTimeout time.Duration
	httpClient   *http.Client
	lock         sync.RWMutex
}

// NewTokenAuthorizationProvider constructs an TokenAuthorizationProvider
func NewMtlsTokenAuthorizationProvider(oauthCfg oauth.Config, cache CertificateCache, creator mtlsClientCreator) *mtlsTokenAuthorizationProvider {
	return &mtlsTokenAuthorizationProvider{
		clientID:     oauthCfg.ClientID,
		tokenURL:     oauthCfg.TokenEndpointProtocol + "://" + oauthCfg.TokenBaseURL + oauthCfg.TokenPath,
		scopesClaim:  strings.Join(oauthCfg.ScopesClaim, " "),
		tokenTimeout: oauthCfg.TokenExpirationTimeout,
		httpClient:   creator(cache, oauthCfg.TokenRequestTimeout),
		lock:         sync.RWMutex{},
	}
}

// Name specifies the name of the AuthorizationProvider
func (p *mtlsTokenAuthorizationProvider) Name() string {
	return "MtlsTokenAuthorizationProvider"
}

// Matches contains the logic for matching the AuthorizationProvider
func (p *mtlsTokenAuthorizationProvider) Matches(_ context.Context) bool {
	return true
}

// GetAuthorization crafts an OAuth Bearer token to inject as part of the executing request
func (p *mtlsTokenAuthorizationProvider) GetAuthorization(ctx context.Context) (string, error) {
	p.lock.RLock()
	isValidToken := !p.token.EmptyOrExpired(p.tokenTimeout)
	p.lock.RUnlock()
	if isValidToken {
		return "Bearer " + p.token.AccessToken, nil
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	if !p.token.EmptyOrExpired(p.tokenTimeout) {
		return "Bearer " + p.token.AccessToken, nil
	}

	log.C(ctx).Debug("Token is invalid, getting a new one...")
	token, err := p.getToken(ctx)
	if err != nil {
		return "", err
	}

	p.token = token
	return "Bearer " + token.AccessToken, nil
}

func (p *mtlsTokenAuthorizationProvider) getToken(ctx context.Context) (httputils.Token, error) {
	log.C(ctx).Info("Getting authorization token")

	form := url.Values{}
	form.Add("grant_type", "client_credentials")
	form.Add("client_id", p.clientID)
	form.Add("scopes", p.scopesClaim)

	body := strings.NewReader(form.Encode())
	request, err := http.NewRequest(http.MethodPost, p.tokenURL, body)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "Failed to create authorisation token request")
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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
		return httputils.Token{}, errors.Wrapf(err, "while reading token response body from %q", p.tokenURL)
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
	tokenResponse.Expiration += time.Now().Unix()

	return tokenResponse, nil
}
