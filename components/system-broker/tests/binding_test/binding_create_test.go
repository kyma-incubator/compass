package binding_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	serviceID        = "02d6080f-8c06-4d05-a7e0-cb15149261f8"
	planID           = "acf316ac-c129-4440-8052-5fc69a3b1486"
	brokerAPIVersion = "2.15"

	instanceID = "1728835a-8e74-4fae-93aa-3e58a022fb85"
	bindingID  = "6f21aca9-4506-4086-b9ba-4aa4c169d018"
)

var (
	bindingPath = fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)

	bundleInstanceAuthResponse = `{
	  "data": {
		"result": {
			"id": "%s",
			"status": {
			  "condition": "%s",
			  "timestamp": "2020-11-04T16:21:20Z",
			  "message": "Credentials user-facing message",
			  "reason": "CredentialsReason"
			},
			"context": "{\"instance_id\": \"%s\", \"binding_id\": \"%s\"}"
		}
	  }
	}`

	notFoundResponse = fmt.Sprint(`{
	  "data": {
		"res": null
	  }
	}`)
)

func TestBindCreate(t *testing.T) {
	suite.Run(t, new(BindCreateTestSuite))
}

type BindCreateTestSuite struct {
	suite.Suite
	testContext       *common.TestContext
	mockedDirectorURL string
}

func (suite *BindCreateTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.mockedDirectorURL = suite.testContext.Servers[common.DirectorServer].URL()
}

func (suite *BindCreateTestSuite) SetupTest() {
	_, err := http.DefaultClient.Post(suite.mockedDirectorURL+"/config/reset", "application/json", nil)
	assert.NoError(suite.T(), err)
}

func (suite *BindCreateTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *BindCreateTestSuite) TearDownTest() {
	resp, err := suite.testContext.HttpClient.Get(suite.mockedDirectorURL + "/verify")
	assert.NoError(suite.T(), err)

	if resp.StatusCode == http.StatusInternalServerError {
		errorMsg, err := ioutil.ReadAll(resp.Body)
		assert.NoError(suite.T(), err)
		suite.Fail(string(errorMsg))
	}
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *BindCreateTestSuite) TestBindWithoutAcceptsIncompleteHeaderShouldReturnUnprocessableEntity() {
	suite.testContext.SystemBroker.PUT(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusUnprocessableEntity)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleInstanceAuth",
		`{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsInsufficientScopesOnFindCredentialsShouldReturnUnauthorized() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleInstanceAuth",
		`{"error": "insufficient scopes provided"}`)
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusUnauthorized)

	resp.JSON().Path("$.description").String().Contains("unauthorized: insufficient scopes")
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorOnFindCredentialsReturnsCredentialsWithMismatchedContextShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleInstanceAuth",
		fmt.Sprintf(bundleInstanceAuthResponse, bindingID, schema.BundleInstanceAuthStatusConditionSucceeded, "mismatched-id", bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsErrorOnBundleInstanceCreationShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleInstanceAuth",
		notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "mutation", "requestBundleInstanceAuthCreation", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsUnauthorizedOnBundleInstanceCreationShouldReturnUnauthorized() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleInstanceAuth",
		notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "mutation", "requestBundleInstanceAuthCreation", `{"error": "insufficient scopes provided"}`)
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusUnauthorized)

	resp.JSON().Path("$.description").String().Contains("unauthorized: insufficient scopes")
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsAuthWithFailedConditionOnBundleInstanceCreationShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "mutation", "requestBundleInstanceAuthCreation",
		fmt.Sprintf(bundleInstanceAuthResponse, bindingID, schema.BundleInstanceAuthStatusConditionFailed, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundWithFailedAuthShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleInstanceAuth",
		fmt.Sprintf(bundleInstanceAuthResponse, bindingID, schema.BundleInstanceAuthStatusConditionFailed, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleInstanceAuth",
		fmt.Sprintf(bundleInstanceAuthResponse, bindingID, schema.BundleInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusAccepted)

	resp.JSON().Path("$.operation").String().Equal(string(osb.BindOp))
}

func (suite *BindCreateTestSuite) TestBindWhenNewCredentialsAreCreatedShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "mutation", "requestBundleInstanceAuthCreation",
		fmt.Sprintf(bundleInstanceAuthResponse, bindingID, schema.BundleInstanceAuthStatusConditionPending, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusAccepted)

	resp.JSON().Path("$.operation").String().Equal(string(osb.BindOp))
}
