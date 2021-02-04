package instance_test

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
)

func TestInstanceLastOp(t *testing.T) {
	suite.Run(t, new(InstanceLastOpTestSuite))
}

type InstanceLastOpTestSuite struct {
	suite.Suite
	testContext *common.TestContext
}

func (suite *InstanceLastOpTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
}

func (suite *InstanceLastOpTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *InstanceLastOpTestSuite) TestLastOp() {
	suite.testContext.SystemBroker.GET("/v2/service_instances/123/last_operation").WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		Expect().Status(http.StatusOK).Body().Equal("{\"state\":\"\"}\n")
}
