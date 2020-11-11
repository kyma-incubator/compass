package binding_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	packageInstanceAuthWithoutContextResponse = `{
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

	lastOperationBindingPath = bindingPath + "/last_operation"
)

func TestBindLastOp(t *testing.T) {
	suite.Run(t, new(BindLastOpTestSuite))
}

type BindLastOpTestSuite struct {
	suite.Suite
	testContext       *common.TestContext
	mockedDirectorURL string
}

func (suite *BindLastOpTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.mockedDirectorURL = suite.testContext.Servers[common.DirectorServer].URL()
}

func (suite *BindLastOpTestSuite) SetupTest() {
	http.DefaultClient.Post(suite.mockedDirectorURL+"/config/reset", "application/json", nil)
}

func (suite *BindLastOpTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *BindLastOpTestSuite) TearDownTest() {
	resp, err := suite.testContext.HttpClient.Get(suite.mockedDirectorURL + "/verify")
	assert.NoError(suite.T(), err)

	if resp.StatusCode == http.StatusInternalServerError {
		errorMsg, err := ioutil.ReadAll(resp.Body)
		assert.NoError(suite.T(), err)
		suite.Fail(string(errorMsg))
	}
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *BindLastOpTestSuite) TestLastOpWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(lastOperationBindingPath).
		WithQuery("operation", osb.BindOp).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindLastOpTestSuite) TestLastOpWhenDirectorReturnsNotFound() {
	suite.Run("BindOpShouldReturnGone", func() {
		err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth", notFoundResponse)
		assert.NoError(suite.T(), err)

		suite.testContext.SystemBroker.GET(lastOperationBindingPath).
			WithQuery("operation", osb.BindOp).
			WithHeader("X-Broker-API-Version", brokerAPIVersion).
			Expect().Status(http.StatusNotFound)
	})

	suite.Run("UnbindOpShouldReturnSucceeded", func() {
		err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth", notFoundResponse)
		assert.NoError(suite.T(), err)

		suite.testContext.SystemBroker.GET(lastOperationBindingPath).
			WithQuery("operation", osb.UnbindOp).
			WithHeader("X-Broker-API-Version", brokerAPIVersion).
			Expect().Status(http.StatusOK).JSON().Path("$.state").String().Equal(string(domain.Succeeded))
	})
}

func (suite *BindLastOpTestSuite) TestLastOpWhenDirectorReturnsCredentialsWithMissingContextShouldReturnNotFound() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthWithoutContextResponse, schema.PackageInstanceAuthStatusConditionPending))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(lastOperationBindingPath).
		WithQuery("operation", osb.BindOp).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusNotFound)
}

func (suite *BindLastOpTestSuite) TestLastOpWhenDirectorReturnsCredentialsWithDifferentInstanceAndBindingIDsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionPending, "111", bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(lastOperationBindingPath).
		WithQuery("operation", osb.BindOp).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindLastOpTestSuite) TestLastOpWithStatus() {
	const UnknownCondition = "UNKNOWN_CONDITION"

	suite.Run("BindOp", func() {
		suite.Run("Credentials succeeded condition should return succeeded state", func() {
			err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
				fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
			assert.NoError(suite.T(), err)

			suite.testContext.SystemBroker.GET(lastOperationBindingPath).
				WithQuery("operation", osb.BindOp).
				WithHeader("X-Broker-API-Version", brokerAPIVersion).
				Expect().Status(http.StatusOK).JSON().Path("$.state").String().Equal(string(domain.Succeeded))
		})

		suite.Run("Credentials pending condition should return in progress state", func() {
			err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
				fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionPending, instanceID, bindingID))
			assert.NoError(suite.T(), err)

			suite.testContext.SystemBroker.GET(lastOperationBindingPath).
				WithQuery("operation", osb.BindOp).
				WithHeader("X-Broker-API-Version", brokerAPIVersion).
				Expect().Status(http.StatusOK).JSON().Path("$.state").String().Equal(string(domain.InProgress))
		})

		suite.Run("Credentials failed condition should return failed state", func() {
			err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
				fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionFailed, instanceID, bindingID))
			assert.NoError(suite.T(), err)

			suite.testContext.SystemBroker.GET(lastOperationBindingPath).
				WithQuery("operation", osb.BindOp).
				WithHeader("X-Broker-API-Version", brokerAPIVersion).
				Expect().Status(http.StatusOK).JSON().Path("$.state").String().Equal(string(domain.Failed))
		})

		suite.Run("Credentials unused condition should return error", func() {
			err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
				fmt.Sprintf(packageInstanceAuthResponse, bindingID, schema.PackageInstanceAuthStatusConditionUnused, instanceID, bindingID))
			assert.NoError(suite.T(), err)

			suite.testContext.SystemBroker.GET(lastOperationBindingPath).
				WithQuery("operation", osb.BindOp).
				WithHeader("X-Broker-API-Version", brokerAPIVersion).
				Expect().Status(http.StatusInternalServerError)
		})

		suite.Run("Credentials unknown condition should return error", func() {
			err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
				fmt.Sprintf(packageInstanceAuthResponse, bindingID, UnknownCondition, instanceID, bindingID))
			assert.NoError(suite.T(), err)

			suite.testContext.SystemBroker.GET(lastOperationBindingPath).
				WithQuery("operation", osb.BindOp).
				WithHeader("X-Broker-API-Version", brokerAPIVersion).
				Expect().Status(http.StatusInternalServerError)
		})
	})

	suite.Run("UnbindOp", func() {
		suite.Run("Any Credentials condition should return in progress state", func() {
			conditions := []string{
				string(schema.PackageInstanceAuthStatusConditionSucceeded),
				string(schema.PackageInstanceAuthStatusConditionPending),
				string(schema.PackageInstanceAuthStatusConditionFailed),
				string(schema.PackageInstanceAuthStatusConditionUnused),
				UnknownCondition,
			}

			for _, condition := range conditions {
				err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "packageInstanceAuth",
					fmt.Sprintf(packageInstanceAuthResponse, bindingID, condition, instanceID, bindingID))
				assert.NoError(suite.T(), err)

				suite.testContext.SystemBroker.GET(lastOperationBindingPath).
					WithQuery("operation", osb.UnbindOp).
					WithHeader("X-Broker-API-Version", brokerAPIVersion).
					Expect().Status(http.StatusOK).JSON().Path("$.state").String().Equal(string(domain.InProgress))
			}
		})
	})
}
