package healthcheck_test

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
)

func TestTokenProviderFromSecretTestSuite(t *testing.T) {
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

func (suite *HealthcheckTestSuite) TestHealthcheck() {
	suite.testContext.SystemBroker.GET("/healthz").Expect().
		Status(http.StatusOK).
		Body().Equal(`{"status": "success"}`)
}
