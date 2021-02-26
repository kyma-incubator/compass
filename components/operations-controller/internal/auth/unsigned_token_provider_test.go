package auth_test

import (
	"context"
	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/auth"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestUnsignedTokenAuthorizationProviderTestSuite(t *testing.T) {
	suite.Run(t, new(UnsignedTokenAuthorizationProviderTestSuite))
}

type UnsignedTokenAuthorizationProviderTestSuite struct {
	suite.Suite
}

const (
	targetURL = "http://localhost"
	tenantID  = "b1f5081d-4c67-4eff-90eb-b8ffaf7b590a"
)

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenProvider_New() {
	provider, err := auth.NewUnsignedTokenAuthorizationProvider("%zzz")
	suite.Require().Error(err)
	suite.Require().Nil(provider)
}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenProvider_Name() {
	provider, err := auth.NewUnsignedTokenAuthorizationProvider(targetURL)
	suite.Require().NoError(err)

	name := provider.Name()

	suite.Require().Equal(name, "UnsignedTokenAuthorizationProvider")
}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenProvider_Matches() {
	provider, err := auth.NewUnsignedTokenAuthorizationProvider(targetURL)
	suite.Require().NoError(err)

	matches := provider.Matches(context.TODO())
	suite.Require().Equal(matches, true)
}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenProvider_URL() {
	provider, err := auth.NewUnsignedTokenAuthorizationProvider(targetURL)
	suite.Require().NoError(err)

	url := provider.TargetURL()

	suite.Require().Equal(url.String(), targetURL)
}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenProvider_GetAuthorizationToken() {
	provider, err := auth.NewUnsignedTokenAuthorizationProvider(targetURL)
	suite.Require().NoError(err)

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
	suite.Require().Equal("application:read", claims.Scopes)

}

func (suite *UnsignedTokenAuthorizationProviderTestSuite) TestUnsignedTokenProvider_GetAuthorizationTokenFailsWhenNoTenantInContext() {
	provider, err := auth.NewUnsignedTokenAuthorizationProvider(targetURL)
	suite.Require().NoError(err)

	authorization, err := provider.GetAuthorization(context.TODO())

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot read tenant from context")
	suite.Require().Empty(authorization)
}
