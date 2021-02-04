package instance_test

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
)

func TestInstanceDeprovision(t *testing.T) {
	suite.Run(t, new(InstanceDeprovisionTestSuite))
}

type InstanceDeprovisionTestSuite struct {
	suite.Suite
	testContext *common.TestContext
}

func (suite *InstanceDeprovisionTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
}

func (suite *InstanceDeprovisionTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *InstanceDeprovisionTestSuite) TestDeprovision() {
	suite.testContext.SystemBroker.DELETE("/v2/service_instances/123").WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		Expect().Status(http.StatusOK).Body().Equal("{}\n")
}
