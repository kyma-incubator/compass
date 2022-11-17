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
	"fmt"
	"io"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	httputilsfakes "github.com/kyma-incubator/compass/components/system-broker/pkg/http/httpfakes"
	"github.com/pkg/errors"

	"testing"

	"github.com/stretchr/testify/suite"
)

func TestTokenAuthorizationProviderTestSuite(t *testing.T) {
	suite.Run(t, new(TokenAuthorizationProviderTestSuite))
}

type TokenAuthorizationProviderTestSuite struct {
	suite.Suite
}

func (suite *TokenAuthorizationProviderTestSuite) TestTokenAuthorizationProvider_New() {
	provider := auth.NewTokenAuthorizationProvider(nil)
	suite.Require().NotNil(provider)
}

func (suite *TokenAuthorizationProviderTestSuite) TestTokenAuthorizationProvider_Name() {
	provider := auth.NewTokenAuthorizationProvider(nil)

	name := provider.Name()

	suite.Require().Equal(name, "TokenAuthorizationProvider")
}

func (suite *TokenAuthorizationProviderTestSuite) TestTokenAuthorizationProvider_Matches() {
	provider := auth.NewTokenAuthorizationProvider(nil)

	matches := provider.Matches(auth.SaveToContext(context.Background(), &auth.OAuthCredentials{}))
	suite.Require().Equal(matches, true)
}

func (suite *TokenAuthorizationProviderTestSuite) TestTokenAuthorizationProvider_DoesNotMatchWhenBasicCredentialsInContext() {
	provider := auth.NewTokenAuthorizationProvider(nil)

	matches := provider.Matches(auth.SaveToContext(context.Background(), &auth.BasicCredentials{}))
	suite.Require().Equal(matches, false)
}

func (suite *TokenAuthorizationProviderTestSuite) TestTokenAuthorizationProvider_DoesNotMatchNoCredentialsInContext() {
	provider := auth.NewTokenAuthorizationProvider(nil)

	matches := provider.Matches(context.TODO())
	suite.Require().Equal(matches, false)
}

func (suite *TokenAuthorizationProviderTestSuite) TestTokenAuthorizationProvider_GetAuthorization() {
	fakeTkn := "fake-token"
	fakeClient := &httputilsfakes.FakeClient{}
	fakeClient.DoReturns(&http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(fmt.Sprintf(`{"access_token": "%s"}`, fakeTkn)))),
	}, nil)

	provider := auth.NewTokenAuthorizationProvider(fakeClient)

	clientID, clientSecret, tokenURL, scopes := "client-id", "client-secret", "https://test-domain.com/oauth/token", "scopes"
	ctx := auth.SaveToContext(context.Background(), &auth.OAuthCredentials{
		ClientID:          clientID,
		ClientSecret:      clientSecret,
		TokenURL:          tokenURL,
		Scopes:            scopes,
		AdditionalHeaders: map[string]string{"h1": "v1"},
	})
	authorization, err := provider.GetAuthorization(ctx)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(authorization)

	suite.Require().Equal("Bearer "+fakeTkn, authorization)
}

func (suite *TokenAuthorizationProviderTestSuite) TestTokenAuthorizationProvider_GetAuthorizationFailsWhenRequestFails() {
	mockedErr := errors.New("test error")
	fakeClient := &httputilsfakes.FakeClient{}
	fakeClient.DoReturns(nil, mockedErr)

	provider := auth.NewTokenAuthorizationProvider(fakeClient)

	ctx := auth.SaveToContext(context.Background(), &auth.OAuthCredentials{})
	authorization, err := provider.GetAuthorization(ctx)

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), mockedErr.Error())
	suite.Require().Empty(authorization)
}

func (suite *TokenAuthorizationProviderTestSuite) TestTokenAuthorizationProvider_GetAuthorizationFailsWhenNoCredentialsInContext() {
	provider := auth.NewTokenAuthorizationProvider(nil)

	authorization, err := provider.GetAuthorization(context.TODO())

	suite.Require().Error(err)
	suite.Require().True(apperrors.IsNotFoundError(err))
	suite.Require().Empty(authorization)
}

func (suite *TokenAuthorizationProviderTestSuite) TestTokenAuthorizationProvider_GetAuthorizationFailsWhenBasicCredentialsAreInContext() {
	provider := auth.NewTokenAuthorizationProvider(nil)

	authorization, err := provider.GetAuthorization(auth.SaveToContext(context.Background(), &auth.BasicCredentials{}))

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "failed to cast credentials to oauth credentials type")
	suite.Require().Empty(authorization)
}
