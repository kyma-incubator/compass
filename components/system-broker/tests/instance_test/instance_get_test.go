package instance_test

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
)

func TestInstanceGet(t *testing.T) {
	suite.Run(t, new(InstanceGetTestSuite))
}

type InstanceGetTestSuite struct {
	suite.Suite
	testContext *common.TestContext
}

func (suite *InstanceGetTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
}

func (suite *InstanceGetTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *InstanceGetTestSuite) TestGet() {
	suite.testContext.SystemBroker.GET("/v2/service_instances/123").WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError).Body().Equal("{\"description\":\"not supported\"}\n")
}
