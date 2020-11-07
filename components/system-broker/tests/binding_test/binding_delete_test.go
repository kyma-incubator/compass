package binding_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
)

var (
	unbindPath               = fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)
	notFoundMutationResponse = `{
  "errors": [
	{
	  "message": "Object not found [object=PackageInstanceAuth]",
	  "path": [
		"result"
	  ],
	  "extensions": {
		"error": "NotFound",
		"error_code": 20
	  }
	}
  ],
  "data": null
}`
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
	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusUnprocessableEntity)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsNotFoundShouldReturnGone() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusGone)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsUnusedCredentialsShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionUnused, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect()
	resp.Status(http.StatusAccepted).JSON().Path("$.operation").String().Equal(string(osb.UnbindOp))
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsErrorOnCredentialsDeletionShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "requestPackageInstanceAuthDeletion", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsNotFoundOnCredentialsDeletionShouldReturnGone() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "mutation", "requestPackageInstanceAuthDeletion", notFoundMutationResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusGone)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorAcceptsCredentialsDeletionRequestShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "mutation", "requestPackageInstanceAuthDeletion",
		fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionUnused, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusAccepted).JSON().Path("$.operation").String().Equal(string(osb.UnbindOp))
}
