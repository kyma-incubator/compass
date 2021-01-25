package binding_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"

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
	testContext       *common.TestContext
	mockedDirectorURL string
}

func (suite *UnbindTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.mockedDirectorURL = suite.testContext.Servers[common.DirectorServer].URL()
}

func (suite *UnbindTestSuite) SetupTest() {
	_, err := http.DefaultClient.Post(suite.mockedDirectorURL+"/config/reset", "application/json", nil)
	assert.NoError(suite.T(), err)
}

func (suite *UnbindTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *UnbindTestSuite) TearDownTest() {
	resp, err := suite.testContext.HttpClient.Get(suite.mockedDirectorURL + "/verify")
	assert.NoError(suite.T(), err)

	if resp.StatusCode == http.StatusInternalServerError {
		errorMsg, err := ioutil.ReadAll(resp.Body)
		assert.NoError(suite.T(), err)
		suite.Fail(string(errorMsg))
	}
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *UnbindTestSuite) TestUnbindWithoutAcceptsIncompleteHeaderShouldReturnUnprocessableEntity() {
	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusUnprocessableEntity)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsUnauthorizedOnFindCredentialsShouldReturnUnauthorized() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth", `{"error": "insufficient scopes provided"}`)
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusUnauthorized)

	resp.JSON().Path("$.description").String().Contains("unauthorized: insufficient scopes")
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsNotFoundShouldReturnGone() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusGone)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsUnusedCredentialsShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
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
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "mutation", "requestPackageInstanceAuthDeletion", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsUnauthorizedOnCredentialsDeletionShouldReturnUnauthorized() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth", fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "mutation", "requestPackageInstanceAuthDeletion", `{"error": "insufficient scopes provided"}`)
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusUnauthorized)

	resp.JSON().Path("$.description").String().Contains("unauthorized: insufficient scopes")
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsNotFoundOnCredentialsDeletionShouldReturnGone() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "mutation", "requestPackageInstanceAuthDeletion", notFoundMutationResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.DELETE(unbindPath).
		WithQuery("service_id", serviceID).
		WithQuery("plan_id", planID).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusGone)
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorAcceptsCredentialsDeletionRequestShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "mutation", "requestPackageInstanceAuthDeletion",
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
