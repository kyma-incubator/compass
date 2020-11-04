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
)

var (
	packageInstanceAuthResponse = `{
	  "data": {
		"result": {
			"status": {
			  "condition": "%s",
			  "timestamp": "2020-11-04T16:21:20Z",
			  "message": "Credentials user-facing message",
			  "reason": "CredentialsReason"
			}
		}
	  }
	}`

	notFoundResponse = fmt.Sprint(`{
	  "error": {
		"error": {
		  "code": 404,
		  "status": "Not Found",
		  "request": "04497842-0836-4b1a-8595-a1b5ba0be38f",
		  "message": "The requested resource could not be found"
		}
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
	suite.testContext.SystemBroker.PUT("/v2/service_instances/123/service_bindings/456").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusUnprocessableEntity)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", "{}")
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsErrorOnPackageInstanceCreationShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "requestPackageInstanceAuthCreation", "{}")
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

/* TODO: Probably unnecessary
func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsMultipleAuthsOnPackageInstanceCreationShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "requestPackageInstanceAuthCreation", "TODO: ...")
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}
*/

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsFailedAuthOnPackageInstanceCreationShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", notFoundResponse)
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "requestPackageInstanceAuthCreation", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionFailed))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

/* TODO: Probably unnecessary
func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundWithMultipleAuthsShouldReturnError() {
}
*/

func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundWithFailedAuthShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionFailed))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionSucceeded))
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.PUT("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusAccepted)

	resp.JSON().Path("$.operation_data").String().Equal(string(osb.BindOp))
	resp.JSON().Path("$.already_exists").String().Equal("true")
}

func (suite *BindCreateTestSuite) TestBindWhenNewCredentialsAreCreatedShouldReturnAccepted() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageInstanceAuth", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionSucceeded))
	assert.NoError(suite.T(), err)

	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "requestPackageInstanceAuthCreation", fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionPending))
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.PUT("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusAccepted)

	resp.JSON().Path("$.operation_data").String().Equal(string(osb.BindOp))
	resp.JSON().Path("$.already_exists").String().Equal("false")
}
