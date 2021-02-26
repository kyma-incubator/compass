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

const targetURL = "http://localhost"

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_New() {
	provider, err := oauth.NewTokenAuthorizationProviderFromHeader("%zzz")
	suite.Require().Error(err)
	suite.Require().Nil(provider)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_Name() {
	provider, err := oauth.NewTokenAuthorizationProviderFromHeader(targetURL)
	suite.Require().NoError(err)

	name := provider.Name()

	suite.Require().Equal(name, "TokenAuthorizationProviderFromHeader")
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_Matches() {
	provider, err := oauth.NewTokenAuthorizationProviderFromHeader(targetURL)
	suite.Require().NoError(err)

	ctx := context.TODO()
	matches := provider.Matches(ctx)
	suite.Require().Equal(matches, false)

	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "Bearer token")
	matches = provider.Matches(ctx)
	suite.Require().Equal(matches, true)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_URL() {
	provider, err := oauth.NewTokenAuthorizationProviderFromHeader(targetURL)
	suite.Require().NoError(err)

	url := provider.TargetURL()

	suite.Require().Equal(url.String(), targetURL)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorization() {
	const tokenVal = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0"
	provider, err := oauth.NewTokenAuthorizationProviderFromHeader(targetURL)
	suite.Require().NoError(err)

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "Bearer "+tokenVal)
	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(authorization, "Bearer "+tokenVal)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorizationFailsWhenNoHeadersInContext() {
	provider, err := oauth.NewTokenAuthorizationProviderFromHeader(targetURL)
	suite.Require().NoError(err)

	ctx := context.TODO()
	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot read headers from context")
	suite.Require().Empty(authorization)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorizationFailsWhenNoAuthHeaderInContext() {
	provider, err := oauth.NewTokenAuthorizationProviderFromHeader(targetURL)
	suite.Require().NoError(err)

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, "key", "val")

	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot read header Authorization from context")
	suite.Require().Empty(authorization)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorizationFailsWhenAuthHeaderIsEmpty() {
	provider, err := oauth.NewTokenAuthorizationProviderFromHeader(targetURL)
	suite.Require().NoError(err)

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "")

	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "missing bearer token")
	suite.Require().Empty(authorization)
}

func (suite *TokenAuthorizationProviderFromHeaderTestSuite) TestTokenAuthorizationProviderFromHeader_GetAuthorizationFailsWhenAuthHeaderIsInvalid() {
	provider, err := oauth.NewTokenAuthorizationProviderFromHeader(targetURL)
	suite.Require().NoError(err)

	ctx := context.TODO()
	ctx = httputils.SaveToContext(ctx, oauth.AuthzHeader, "NotBearer ")

	authorization, err := provider.GetAuthorization(ctx)
	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "invalid bearer token prefix")
	suite.Require().Empty(authorization)
}
