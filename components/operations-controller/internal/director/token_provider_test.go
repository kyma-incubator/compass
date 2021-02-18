package director_test

import (
	"context"
	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	"testing"

	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/stretchr/testify/suite"
)

func TestUnsignedTokenProviderTestSuite(t *testing.T) {
	suite.Run(t, new(UnsignedTokenProviderTestSuite))
}

type UnsignedTokenProviderTestSuite struct {
	suite.Suite
}

const (
	targetURL = "http://localhost"
	tenantID  = "b1f5081d-4c67-4eff-90eb-b8ffaf7b590a"
)

func (suite *UnsignedTokenProviderTestSuite) TestUnsignedTokenProvider_New() {
	provider, err := director.NewUnsignedTokenProvider("%zzz")
	suite.Require().Error(err)
	suite.Require().Nil(provider)
}

func (suite *UnsignedTokenProviderTestSuite) TestUnsignedTokenProvider_Name() {
	provider, err := director.NewUnsignedTokenProvider(targetURL)
	suite.Require().NoError(err)

	name := provider.Name()

	suite.Require().Equal(name, "UnsignedTokenProvider")
}

func (suite *UnsignedTokenProviderTestSuite) TestUnsignedTokenProvider_Matches() {
	provider, err := director.NewUnsignedTokenProvider(targetURL)
	suite.Require().NoError(err)

	matches := provider.Matches(context.TODO())
	suite.Require().Equal(matches, true)
}

func (suite *UnsignedTokenProviderTestSuite) TestUnsignedTokenProvider_URL() {
	provider, err := director.NewUnsignedTokenProvider(targetURL)
	suite.Require().NoError(err)

	url := provider.TargetURL()

	suite.Require().Equal(url.String(), targetURL)
}

func (suite *UnsignedTokenProviderTestSuite) TestUnsignedTokenProvider_GetAuthorizationToken() {
	provider, err := director.NewUnsignedTokenProvider(targetURL)
	suite.Require().NoError(err)

	ctx := tenant.SaveToContext(context.TODO(), tenantID)
	token, err := provider.GetAuthorizationToken(ctx)

	suite.Require().NoError(err)
	suite.Require().NotEmpty(token.AccessToken)
	suite.Require().Equal(token.Expiration, int64(0))

	claims := director.Claims{}
	_, err = jwt.ParseWithClaims(token.AccessToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwt.UnsafeAllowNoneSignatureType, nil
	})

	suite.Require().NoError(err)
	suite.Require().Equal(tenantID, claims.Tenant)
	suite.Require().Equal("application:read", claims.Scopes)

}

func (suite *UnsignedTokenProviderTestSuite) TestUnsignedTokenProvider_GetAuthorizationTokenFailsWhenNoTenantInContext() {
	provider, err := director.NewUnsignedTokenProvider(targetURL)
	suite.Require().NoError(err)

	token, err := provider.GetAuthorizationToken(context.TODO())

	suite.Require().Error(err)
	suite.Require().Contains(err.Error(), "cannot read tenant from context")
	suite.Require().Equal(token, httputils.Token{})
}
