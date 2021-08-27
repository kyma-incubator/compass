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
	"strings"
	"testing"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/stretchr/testify/suite"
)

var scopes = "entity.operation"

func TestUnsignedTokenAuthorizationProviderTestSuite(t *testing.T) {
	suite.Run(t, new(UnsignedTokenAuthorizationProviderTestSuite))
}

type UnsignedTokenAuthorizationProviderTestSuite struct {
	suite.Suite
}

const tenantID = "b1f5081d-4c67-4eff-90eb-b8ffaf7b590a"

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenAuthorizationProvider_New() {
	provider := auth.NewUnsignedTokenAuthorizationProvider(scopes)
	suite.Require().NotNil(provider)
}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenAuthorizationProvider_Name() {
	provider := auth.NewUnsignedTokenAuthorizationProvider(scopes)

	name := provider.Name()

	suite.Require().Equal(name, "UnsignedTokenAuthorizationProvider")
}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenAuthorizationProvider_Matches() {
	provider := auth.NewUnsignedTokenAuthorizationProvider(scopes)

	matches := provider.Matches(context.TODO())
	suite.Require().Equal(matches, true)
}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenAuthorizationProvider_DoesNotMatchWhenBasicCredentialsInContext() {
	provider := auth.NewUnsignedTokenAuthorizationProvider(scopes)

	matches := provider.Matches(auth.SaveToContext(context.Background(), &graphql.BasicCredentialData{}))
	suite.Require().Equal(matches, false)
}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenAuthorizationProvider_DoesNotMatchWhenOAuthCredentialsInContext() {
	provider := auth.NewUnsignedTokenAuthorizationProvider(scopes)

	matches := provider.Matches(auth.SaveToContext(context.Background(), &graphql.OAuthCredentialData{}))
	suite.Require().Equal(matches, false)
}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenAuthorizationProvider_GetAuthorization() {
	provider := auth.NewUnsignedTokenAuthorizationProvider(scopes)

	ctx := tenant.SaveToContext(context.TODO(), tenantID)
	authorization, err := provider.GetAuthorization(ctx)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(authorization)

	claims := auth.Claims{}
	_, err = jwt.ParseWithClaims(strings.TrimPrefix(authorization, "Bearer "), &claims, func(token *jwt.Token) (interface{}, error) {
		return jwt.UnsafeAllowNoneSignatureType, nil
	})

	suite.Require().NoError(err)
	suite.Require().Equal(tenantID, claims.Tenant)
	suite.Require().Equal(scopes, claims.Scopes)

}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenAuthorizationProvider_GetAuthorizationFailsWhenNoTenantInContext() {
	provider := auth.NewUnsignedTokenAuthorizationProvider(scopes)

	authorization, err := provider.GetAuthorization(context.TODO())

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot read tenant from context")
	suite.Require().Empty(authorization)
}
