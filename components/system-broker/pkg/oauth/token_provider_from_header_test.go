package oauth_test

import (
	"context"
	"testing"

	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"

	"github.com/stretchr/testify/suite"
)

func TestTokenAuthorizationProviderFromHeaderTestSuite(t *testing.T) {
	suite.Run(t, new(TokenAuthorizationProviderFromHeaderTestSuite))
}

type TokenAuthorizationProviderFromHeaderTestSuite struct {
	suite.Suite
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_Name() {
	provider := oauth.NewTokenAuthorizationProviderFromHeader()

	name := provider.Name()

	suite.Require().Equal(name, "TokenAuthorizationProviderFromHeader")
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_Matches() {
	provider := oauth.NewTokenAuthorizationProviderFromHeader()

	ctx := context.TODO()
	matches := provider.Matches(ctx)
	suite.Require().Equal(matches, false)

	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "Bearer token")
	matches = provider.Matches(ctx)
	suite.Require().Equal(matches, true)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorization() {
	const tokenVal = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0"
	provider := oauth.NewTokenAuthorizationProviderFromHeader()

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "Bearer "+tokenVal)
	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(authorization, "Bearer "+tokenVal)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorizationFailsWhenNoHeadersInContext() {
	provider := oauth.NewTokenAuthorizationProviderFromHeader()

	ctx := context.TODO()
	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot read headers from context")
	suite.Require().Empty(authorization)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorizationFailsWhenNoAuthHeaderInContext() {
	provider := oauth.NewTokenAuthorizationProviderFromHeader()

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, "key", "val")

	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot read header Authorization from context")
	suite.Require().Empty(authorization)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorizationFailsWhenAuthHeaderIsEmpty() {
	provider := oauth.NewTokenAuthorizationProviderFromHeader()

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "")

	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "missing bearer token")
	suite.Require().Empty(authorization)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorizationFailsWhenAuthHeaderIsInvalid() {
	provider := oauth.NewTokenAuthorizationProviderFromHeader()

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "NotBearer ")

	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid bearer token prefix")
	suite.Require().Empty(authorization)
}
