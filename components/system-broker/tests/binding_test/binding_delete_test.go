package binding_test

import (
	"fmt"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

func TestBindDelete(t *testing.T) {
	suite.Run(t, new(UnbindTestSuite))
}

type UnbindTestSuite struct {
	suite.Suite
	testContext *common.TestContext
	configURL   string
}

func (suite *UnbindTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.configURL = suite.testContext.Servers[common.DirectorServer].URL() + "/config"
}

func (suite *UnbindTestSuite) SetupTest() {
	http.DefaultClient.Post(suite.configURL+"/reset", "application/json", nil)
}

func (suite *UnbindTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *UnbindTestSuite) TestUnbindWithoutAcceptsIncompleteHeaderShouldReturnUnprocessableEntity() {
	suite.testContext.SystemBroker.DELETE("/v2/service_instances/123/service_bindings/456").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusUnprocessableEntity)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", "{}")
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsNotFoundShouldReturnGone() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusGone)
}

/* TODO: Probably unnecessary test
func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsCredentialsWithMultipleAuthsShouldReturnError() {

}
*/

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsUnusedCredentialsShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionUnused))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusAccepted).JSON().Path("$.operation_data").String().Equal(string(osb.UnbindOp))
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsErrorOnCredentialsDeletionShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionSucceeded))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "requestPackageInstanceAuthDeletion", "{}")
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsNotFoundOnCredentialsDeletionShouldReturnGone() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionSucceeded))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "requestPackageInstanceAuthDeletion", notFoundResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusGone)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorAcceptsCredentialsDeletionRequestShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionSucceeded))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "requestPackageInstanceAuthDeletion", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionUnused))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusAccepted).JSON().Path("$.operation_data").String().Equal(string(osb.UnbindOp))
}
