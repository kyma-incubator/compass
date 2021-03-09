package healthcheck_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
)

func TestTokenAuthorizationProviderFromSecretTestSuite(t *testing.T) {
	suite.Run(t, new(HealthcheckTestSuite))
}

type HealthcheckTestSuite struct {
	suite.Suite

	testContext *common.TestContext
}

func (suite *HealthcheckTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
}

func (suite *HealthcheckTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}
