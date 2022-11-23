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
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/stretchr/testify/suite"
)

func TestServiceAccountTokenAuthorizationProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceAccountTokenAuthorizationProviderTestSuite))
}

type ServiceAccountTokenAuthorizationProviderTestSuite struct {
	suite.Suite
}

func (suite *ServiceAccountTokenAuthorizationProviderTestSuite) TestServiceAccountTokenAuthorizationProvider_New() {
	provider := auth.NewServiceAccountTokenAuthorizationProvider()
	suite.Require().NotNil(provider)
}

func (suite *ServiceAccountTokenAuthorizationProviderTestSuite) TestServiceAccountTokenAuthorizationProvider_Name() {
	provider := auth.NewServiceAccountTokenAuthorizationProvider()

	name := provider.Name()

	suite.Require().Equal(name, "ServiceAccountTokenAuthorizationProvider")
}

func (suite *ServiceAccountTokenAuthorizationProviderTestSuite) TestServiceAccountTokenAuthorizationProvider_Matches() {
	provider := auth.NewServiceAccountTokenAuthorizationProvider()

	matches := provider.Matches(context.TODO())
	suite.Require().Equal(matches, true)
}

func (suite *ServiceAccountTokenAuthorizationProviderTestSuite) TestServiceAccountTokenAuthorizationProvider_DoesNotMatchWhenBasicCredentialsInContext() {
	provider := auth.NewServiceAccountTokenAuthorizationProvider()

	matches := provider.Matches(auth.SaveToContext(context.Background(), &auth.BasicCredentials{}))
	suite.Require().Equal(matches, false)
}

func (suite *ServiceAccountTokenAuthorizationProviderTestSuite) TestServiceAccountTokenAuthorizationProvider_DoesNotMatchWhenOAuthCredentialsInContext() {
	provider := auth.NewServiceAccountTokenAuthorizationProvider()

	matches := provider.Matches(auth.SaveToContext(context.Background(), &auth.OAuthCredentials{}))
	suite.Require().Equal(matches, false)
}

func (suite *ServiceAccountTokenAuthorizationProviderTestSuite) TestServiceAccountTokenAuthorizationProvider_GetAuthorization() {
	tokenContent := "test-token"
	tokenFileName := "token"
	err := os.WriteFile(tokenFileName, []byte(tokenContent), os.ModePerm)
	suite.Require().NoError(err)

	defer func() {
		err := os.Remove(tokenFileName)
		suite.Require().NoError(err)
	}()

	provider := auth.NewServiceAccountTokenAuthorizationProviderWithPath(tokenFileName)

	authorization, err := provider.GetAuthorization(context.TODO())

	suite.Require().NoError(err)
	suite.Require().NotEmpty(authorization)
	suite.Require().Equal("Bearer "+tokenContent, authorization)
}
