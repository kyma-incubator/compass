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
	"context"
	"encoding/base64"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/suite"
)

func TestBasicAuthorizationProviderTestSuite(t *testing.T) {
	suite.Run(t, new(BasicAuthorizationProviderTestSuite))
}

type BasicAuthorizationProviderTestSuite struct {
	suite.Suite
}

func (suite *BasicAuthorizationProviderTestSuite) TestBasicAuthorizationProvider_New() {
	provider := auth.NewBasicAuthorizationProvider()
	suite.Require().NotNil(provider)
}

func (suite *BasicAuthorizationProviderTestSuite) TestBasicAuthorizationProvider_Name() {
	provider := auth.NewBasicAuthorizationProvider()

	name := provider.Name()

	suite.Require().Equal(name, "BasicAuthorizationProvider")
}

func (suite *BasicAuthorizationProviderTestSuite) TestBasicAuthorizationProvider_Matches() {
	provider := auth.NewBasicAuthorizationProvider()

	matches := provider.Matches(auth.SaveToContext(context.Background(), &graphql.BasicCredentialData{}))
	suite.Require().Equal(matches, true)
}

func (suite *BasicAuthorizationProviderTestSuite) TestBasicAuthorizationProvider_DoesNotMatchWhenOAuthCredentialsInContext() {
	provider := auth.NewBasicAuthorizationProvider()

	matches := provider.Matches(auth.SaveToContext(context.Background(), &graphql.OAuthCredentialData{}))
	suite.Require().Equal(matches, false)
}

func (suite *BasicAuthorizationProviderTestSuite) TestBasicAuthorizationProvider_DoesNotMatchNoCredentialsInContext() {
	provider := auth.NewBasicAuthorizationProvider()

	matches := provider.Matches(context.TODO())
	suite.Require().Equal(matches, false)
}

func (suite *BasicAuthorizationProviderTestSuite) TestBasicAuthorizationProvider_GetAuthorization() {
	provider := auth.NewBasicAuthorizationProvider()

	username, password := "user", "pass"
	ctx := auth.SaveToContext(context.Background(), &graphql.BasicCredentialData{
		Username: username,
		Password: password,
	})
	authorization, err := provider.GetAuthorization(ctx)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(authorization)

	auth := username + ":" + password
	base64Creds := base64.StdEncoding.EncodeToString([]byte(auth))
	suite.Require().Equal("Basic "+base64Creds, authorization)
}

func (suite *BasicAuthorizationProviderTestSuite) TestBasicAuthorizationProvider_GetAuthorizationFailsWhenNoCredentialsInContext() {
	provider := auth.NewBasicAuthorizationProvider()

	authorization, err := provider.GetAuthorization(context.TODO())

	suite.Require().Error(err)
	suite.Require().True(apperrors.IsNotFoundError(err))
	suite.Require().Empty(authorization)
}

func (suite *BasicAuthorizationProviderTestSuite) TestBasicAuthorizationProvider_GetAuthorizationFailsWhenOAuthCredentialsAreInContext() {
	provider := auth.NewBasicAuthorizationProvider()

	authorization, err := provider.GetAuthorization(auth.SaveToContext(context.Background(), &graphql.OAuthCredentialData{}))

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "failed to cast credentials to basic credentials type")
	suite.Require().Empty(authorization)
}
