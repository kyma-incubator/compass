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

const (
	serviceID        = "02d6080f-8c06-4d05-a7e0-cb15149261f8"
	planID           = "acf316ac-c129-4440-8052-5fc69a3b1486"
	brokerAPIVersion = "2.15"

	instanceID = "1728835a-8e74-4fae-93aa-3e58a022fb85"
	bindingID  = "6f21aca9-4506-4086-b9ba-4aa4c169d018"
)

var (
	bindingPath = fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)

	packageInstanceAuthResponse = `{
	  "data": {
		"result": {
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
	testContext *common.TestContext
	configURL   string
}

func (suite *BindCreateTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.configURL = suite.testContext.Servers[common.DirectorServer].URL() + "/config"
}

func (suite *BindCreateTestSuite) SetupTest() {
	http.DefaultClient.Post(suite.configURL+"/reset", "application/json", nil)
}

func (suite *BindCreateTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *BindCreateTestSuite) TestBindWithoutAcceptsIncompleteHeaderShouldReturnUnprocessableEntity() {
	suite.testContext.SystemBroker.PUT(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusUnprocessableEntity)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorOnFindCredentialsReturnsCredentialsWithMismatchedContextShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionSucceeded, "mismatched-id", bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsErrorOnPackageInstanceCreationShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "mutation", "requestPackageInstanceAuthCreation", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsAuthWithFailedConditionOnPackageInstanceCreationShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "mutation", "requestPackageInstanceAuthCreation",
		fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionFailed, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundWithFailedAuthShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionFailed, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusAccepted)

	resp.JSON().Path("$.operation").String().Equal(string(osb.BindOp))
}

func (suite *BindCreateTestSuite) TestBindWhenNewCredentialsAreCreatedShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "mutation", "requestPackageInstanceAuthCreation",
		fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionPending, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.PUT(bindingPath).
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusAccepted)

	resp.JSON().Path("$.operation").String().Equal(string(osb.BindOp))
}
