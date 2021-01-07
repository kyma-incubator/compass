package oauth_test

import (
	"context"
	"testing"

	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"

	"github.com/stretchr/testify/suite"
)

func TestTokenProviderFromHeaderTestSuite(t *testing.T) {
	suite.Run(t, new(TokenProviderFromHeaderTestSuite))
}

type TokenProviderFromHeaderTestSuite struct {
	suite.Suite
}

func (suite *TokenProviderFromHeaderTestSuite) TestOAuthTokenProvider_Name() {
	provider := oauth.NewTokenProviderFromHeader()
	name := provider.Name()

	suite.Require().Equal(name, "TokenProviderFromHeader")
}

func (suite *TokenProviderFromHeaderTestSuite) TestOAuthTokenProvider_Matches() {
	provider := oauth.NewTokenProviderFromHeader()

	ctx := context.TODO()
	matches := provider.Matches(ctx)
	suite.Require().Equal(matches, false)

	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "Bearer token")
	matches = provider.Matches(ctx)
	suite.Require().Equal(matches, true)
}

func (suite *TokenProviderFromHeaderTestSuite) TestOAuthTokenProvider_GetAuthorizationToken() {
	const tokenVal = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0"
	provider := oauth.NewTokenProviderFromHeader()

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "Bearer "+tokenVal)
	token, err := provider.GetAuthorizationToken(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(token.AccessToken, tokenVal)
	suite.Require().Equal(token.Expiration, int64(0))
}

func (suite *TokenProviderFromHeaderTestSuite) TestOAuthTokenProvider_GetAuthorizationTokenFailsWhenNoHeadersInContext() {
	provider := oauth.NewTokenProviderFromHeader()

	ctx := context.TODO()
	token, err := provider.GetAuthorizationToken(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot read headers from context")
	suite.Require().Equal(token, httputils.Token{})
}

func (suite *TokenProviderFromHeaderTestSuite) TestOAuthTokenProvider_GetAuthorizationTokenFailsWhenNoAuthHeaderInContext() {
	provider := oauth.NewTokenProviderFromHeader()

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, "key", "val")

	token, err := provider.GetAuthorizationToken(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot read header Authorization from context")
	suite.Require().Equal(token, httputils.Token{})
}

func (suite *TokenProviderFromHeaderTestSuite) TestOAuthTokenProvider_GetAuthorizationTokenFailsWhenAuthHeaderIsEmpty() {
	provider := oauth.NewTokenProviderFromHeader()

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "")

	token, err := provider.GetAuthorizationToken(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "missing bearer token")
	suite.Require().Equal(token, httputils.Token{})
}

func (suite *TokenProviderFromHeaderTestSuite) TestOAuthTokenProvider_GetAuthorizationTokenFailsWhenAuthHeaderIsInvalid() {
	provider := oauth.NewTokenProviderFromHeader()

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "NotBearer ")

	token, err := provider.GetAuthorizationToken(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid bearer token prefix")
	suite.Require().Equal(token, httputils.Token{})
}
