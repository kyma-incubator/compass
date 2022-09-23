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

package auth_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/auth/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	fakeTkn = "fake-token"
	tenant  = "tenant42"
)

var oauthCfg = oauth.Config{
	ClientID:              "client-id",
	TokenEndpointProtocol: "https",
	TokenBaseURL:          "test.mtls.domain.com",
	TokenPath:             "/cert/token",
	ScopesClaim:           []string{"my-scope"},
	TenantHeaderName:      "x-tenant",
	ExternalClientCertSecretName: "resource-name",
}

func TestMtlsTokenAuthorizationProviderTestSuite(t *testing.T) {
	suite.Run(t, new(MtlsTokenAuthorizationProviderTestSuite))
}

type MtlsTokenAuthorizationProviderTestSuite struct {
	suite.Suite
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_DefaultMtlsClientCreator() {
	cache := &automock.CertificateCache{}
	cache.On("Get").Return(map[string]*tls.Certificate{"resource-name": &tls.Certificate{}}, nil).Once()
	defer cache.AssertExpectations(suite.T())

	client := auth.DefaultMtlsClientCreator(cache, true, time.Second, "resource-name")

	ts := httptest.NewUnstartedServer(testServerHandlerFunc(suite.T()))
	ts.TLS = &tls.Config{
		ClientAuth: tls.RequestClientCert,
	}

	ts.StartTLS()
	defer ts.Close()

	resp, err := client.Get(ts.URL)
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_New() {
	provider := auth.NewMtlsTokenAuthorizationProvider(oauth.Config{}, &automock.CertificateCache{}, auth.DefaultMtlsClientCreator)
	suite.Require().NotNil(provider)
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_Name() {
	provider := auth.NewMtlsTokenAuthorizationProvider(oauth.Config{}, &automock.CertificateCache{}, auth.DefaultMtlsClientCreator)

	name := provider.Name()

	suite.Require().Equal(name, "MtlsTokenAuthorizationProvider")
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_Matches() {
	provider := auth.NewMtlsTokenAuthorizationProvider(oauth.Config{}, &automock.CertificateCache{}, auth.DefaultMtlsClientCreator)

	matches := provider.Matches(auth.SaveToContext(context.Background(), &auth.OAuthMtlsCredentials{}))
	suite.Require().Equal(matches, true)
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_DoesNotMatchWhenBasicCredentialsInContext() {
	provider := auth.NewMtlsTokenAuthorizationProvider(oauth.Config{}, &automock.CertificateCache{}, auth.DefaultMtlsClientCreator)

	matches := provider.Matches(auth.SaveToContext(context.Background(), &auth.BasicCredentials{}))
	suite.Require().Equal(matches, false)
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_DoesNotMatchNoCredentialsInContext() {
	provider := auth.NewMtlsTokenAuthorizationProvider(oauth.Config{}, &automock.CertificateCache{}, auth.DefaultMtlsClientCreator)

	matches := provider.Matches(context.TODO())
	suite.Require().Equal(matches, false)
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_GetAuthorization() {
	provider := auth.NewMtlsTokenAuthorizationProvider(oauthCfg, nil, getFakeCreator(oauthCfg, suite.Suite, false))

	ctx := auth.SaveToContext(context.Background(), &auth.OAuthMtlsCredentials{
		ClientID:          oauthCfg.ClientID,
		TokenURL:          oauthCfg.TokenEndpointProtocol + "://" + oauthCfg.TokenBaseURL + oauthCfg.TokenPath,
		Scopes:            strings.Join(oauthCfg.ScopesClaim, " "),
		AdditionalHeaders: map[string]string{oauthCfg.TenantHeaderName: tenant},
	})
	authorization, err := provider.GetAuthorization(ctx)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(authorization)

	suite.Require().Equal("Bearer "+fakeTkn, authorization)
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_GetAuthorizationFailsWhenRequestFails() {
	provider := auth.NewMtlsTokenAuthorizationProvider(oauthCfg, nil, getFakeCreator(oauthCfg, suite.Suite, true))

	ctx := auth.SaveToContext(context.Background(), &auth.OAuthMtlsCredentials{
		ClientID:          oauthCfg.ClientID,
		TokenURL:          oauthCfg.TokenEndpointProtocol + "://" + oauthCfg.TokenBaseURL + oauthCfg.TokenPath,
		Scopes:            strings.Join(oauthCfg.ScopesClaim, " "),
		AdditionalHeaders: map[string]string{oauthCfg.TenantHeaderName: tenant},
	})
	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "error")
	suite.Require().Empty(authorization)
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_GetAuthorizationFailsWhenNoCredentialsInContext() {
	provider := auth.NewMtlsTokenAuthorizationProvider(oauthCfg, nil, getFakeCreator(oauthCfg, suite.Suite, true))

	authorization, err := provider.GetAuthorization(context.TODO())

	suite.Require().Error(err)
	suite.Require().True(apperrors.IsNotFoundError(err))
	suite.Require().Empty(authorization)
}

func (suite *MtlsTokenAuthorizationProviderTestSuite) TestMtlsTokenAuthorizationProvider_GetAuthorizationFailsWhenBasicCredentialsAreInContext() {
	provider := auth.NewMtlsTokenAuthorizationProvider(oauthCfg, nil, getFakeCreator(oauthCfg, suite.Suite, true))

	authorization, err := provider.GetAuthorization(auth.SaveToContext(context.Background(), &auth.BasicCredentials{}))

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "failed to cast credentials to mtls oauth credentials type")
	suite.Require().Empty(authorization)
}

func getFakeCreator(oauthCfg oauth.Config, suite suite.Suite, shouldFail bool) auth.MtlsClientCreator {
	return func(_ auth.CertificateCache, skipSSLValidation bool, timeout time.Duration, secretName string) *http.Client {
		return &http.Client{
			Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
				suite.Require().Equal(req.URL.Host, oauthCfg.TokenBaseURL)
				suite.Require().Equal(req.URL.Scheme, oauthCfg.TokenEndpointProtocol)
				suite.Require().Equal(req.URL.Path, oauthCfg.TokenPath)
				suite.Require().Equal(req.Header.Get(oauthCfg.TenantHeaderName), tenant)
				suite.Require().Equal(req.Header.Get("Content-Type"), "application/x-www-form-urlencoded")

				body, err := ioutil.ReadAll(req.Body)
				suite.Require().NoError(err)
				suite.Require().NotNil(body)

				form, err := url.ParseQuery(string(body))
				suite.Require().NoError(err)
				suite.Require().NotNil(form)

				suite.Require().Equal(form.Get("grant_type"), "client_credentials")
				suite.Require().Equal(form.Get("client_id"), oauthCfg.ClientID)
				suite.Require().Equal(form.Get("scope"), strings.Join(oauthCfg.ScopesClaim, " "))

				if shouldFail {
					return nil, errors.New("error")
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader([]byte(fmt.Sprintf(`{"access_token": "%s"}`, fakeTkn)))),
				}, nil
			}),
		}
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testServerHandlerFunc(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "Hello, client")
		require.NoError(t, err)
	}
}
