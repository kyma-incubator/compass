package binding

import (
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

func TestBindLastOp(t *testing.T) {
	suite.Run(t, new(BindLastOpTestSuite))
}

type BindLastOpTestSuite struct {
	suite.Suite
	testContext *common.TestContext
	configURL   string
}

func (suite *BindLastOpTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.configURL = suite.testContext.Servers[common.DirectorServer].URL() + "/config"
}

func (suite *BindLastOpTestSuite) SetupTest() {
	http.DefaultClient.Post(suite.configURL+"/reset", "application/json", nil)
}

func (suite *BindLastOpTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *BindLastOpTestSuite) TestLastOpWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
}

func (suite *BindLastOpTestSuite) TestLastOpWhenDirectorReturnsNotFound() {
	suite.Run("BindOpShouldReturnGone", func() {

	})

	suite.Run("UnbindOpShouldReturnSucceeded", func() {

	})
}

func (suite *BindLastOpTestSuite) TestLastOpWhenDirectorReturnsCredentialsWithMissingContextShouldReturnError() {

}

func (suite *BindLastOpTestSuite) TestLastOpWhenDirectorReturnsCredentialsWithUnprocessableContextShouldReturnError() {

}

func (suite *BindLastOpTestSuite) TestLastOpWhenDirectorReturnsCredentialsWithDifferentInstanceAndBindingIDsShouldReturnGone() {

}

func (suite *BindLastOpTestSuite) TestLastOpWithStatus() {
	suite.Run("BindOp", func() {
		suite.Run("Credentials succeeded condition should return succeeded state", func() {

		})

		suite.Run("Credentials pending condition should return in progress state", func() {

		})

		suite.Run("Credentials failed condition should return failed state", func() {

		})

		suite.Run("Credentials unused condition should return error", func() {

		})

		suite.Run("Credentials unknown condition should return error", func() {

		})
	})

	suite.Run("UnbindOp", func() {
		suite.Run("Credentials succeeded condition should return in progress state", func() {

		})

		suite.Run("Credentials pending condition should return in progress state", func() {

		})

		suite.Run("Credentials failed condition should return in progress state", func() {

		})

		suite.Run("Credentials unused condition should return in progress state", func() {

		})

		suite.Run("Credentials unknown condition should return in progress state", func() {

		})
	})
}
