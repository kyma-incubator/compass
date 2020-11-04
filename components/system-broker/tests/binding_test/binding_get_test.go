package binding_test

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

var (
	packagePendingAuthResp = `{
	"data": {
		"result": {
			"id": "76ades8b-7a33-49db-b4h2-da061d50b922",
			"instanceAuths": [
				{
					"id": "46cfe0b3-9633-4441-a639-15926df364b9",
					"status": {
						"condition": "PENDING",
						"timestamp": "2020-04-20T10:54:07Z",
						"message": "Credentials are pending.",
						"reason": "CredentialsPending"
					}
				}
			]	
		}
	  }
	}`

	packageWithAPIsResp = `{
	"data": {
		"result": {
			"id": "76ad2c2b-7a33-49fb-b412-daz67d50b922",
			"instanceAuths": [
				{
					"id": "16cfe0b3-9633-4441-a639-15126ff3z4b9",
					"auth": {
						%s
					},
					"context": %s,
					"status": {
						"condition": "SUCCEEDED",
						"timestamp": "2020-04-20T10:54:07Z",
						"message": "Credentials were provided.",
						"reason": "CredentialsProvided"
					}
				}
			],	
			"apiDefinitions": {
				"data": [
					{
						"id": "1a23bd57-ef6e-46fd-a59s-15632cd3c410",
						"name": "API Cloud - Inbound Test Price",
						"targetURL": "https://api.cloud.com/api/InboundTestPrice",
					},
					{
						"id": "2e635cc3-fc9b-4a8d-zb4e-iaf73cbd8846",
						"name": "API Cloud - Inbound Test Stock",
						"targetURL": "https://api.cloud.com/InboundTestStock",
					}
				]
			}
		}
	  }`

	unknownAuth           = `"credential": { "a": "aa", "bb": "bb"},`
	basicAuth             = `"credential": { "username": "asd", "password": "asd"},`
	oAuth                 = `"credential": { "clientId": "test-id", "clientSecret": "test-secret", "url": "https://api.test.com/oauth/token" }`
	additionalHeaders     = `"additionalHeaders": "\"key\": \"value\"",`
	additionalQueryParams = `"additionalQueryParams": "\"key\": \"value\"",`
)

func TestBindGet(t *testing.T) {
	suite.Run(t, new(BindGetTestSuite))
}

type BindGetTestSuite struct {
	suite.Suite
	testContext *common.TestContext
	configURL   string
}

func (suite *BindGetTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.configURL = suite.testContext.Servers[common.DirectorServer].URL() + "/config"
}

func (suite *BindGetTestSuite) SetupTest() {
	http.DefaultClient.Post(suite.configURL+"/reset", "application/json", nil)
}

func (suite *BindGetTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth", "{}")
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET("/v2/service_instances/123/service_bindings/456").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

/* TODO: Maybe unnecessary
func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsCredentialsWithMultipleAuthsShouldReturnError() {
}
*/

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsUnsuccessfulCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth", packagePendingAuthResp)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET("/v2/service_instances/123/service_bindings/456").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsUnknownCredentialTypeShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth", fmt.Sprintf(packageWithAPIsResp, unknownAuth, "{}"))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET("/v2/service_instances/123/service_bindings/456").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnedCredentialsButNoneMatchInstanceAndBindingShouldReturnNotFound() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth", fmt.Sprintf(packageWithAPIsResp, basicAuth, "{}"))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET("/v2/service_instances/123/service_bindings/456").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusNotFound)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsValidCredentialsShouldReturnCredentials() {
	instanceID, bindingID := "123", "456"
	getBindingPath := fmt.Sprintf("/v2/service_instances/%s/service_bindings/%s", instanceID, bindingID)
	context := fmt.Sprintf(`{"instance_id": "%s", "binding_id": "%s"}`, instanceID, bindingID)

	suite.Run("Basic Authentication", func() {
		err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth", fmt.Sprintf(packageWithAPIsResp, basicAuth, context))
		assert.NoError(suite.T(), err)

		resp := suite.testContext.SystemBroker.GET(getBindingPath).
			WithHeader("X-Broker-API-Version", brokerAPIVersion).
			Expect().Status(http.StatusOK)

		resp.JSON().Path("$.credentials").NotNull()
		// TODO: Add more concrete asserts - check if basic creds are available & target urls
	})

	suite.Run("OAuth", func() {
		err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth", fmt.Sprintf(packageWithAPIsResp, oAuth, context))
		assert.NoError(suite.T(), err)

		resp := suite.testContext.SystemBroker.GET(getBindingPath).
			WithHeader("X-Broker-API-Version", brokerAPIVersion).
			Expect().Status(http.StatusOK)

		resp.JSON().Path("$.credentials").NotNull()
		// TODO: Add more concrete asserts - check if oauth creds are available & target url
	})

	suite.Run("None", func() {
		err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth", fmt.Sprintf(packageWithAPIsResp, "", context))
		assert.NoError(suite.T(), err)

		resp := suite.testContext.SystemBroker.GET(getBindingPath).
			WithHeader("X-Broker-API-Version", brokerAPIVersion).
			Expect().Status(http.StatusOK)

		resp.JSON().Path("$.credentials").Null()
		// TODO: Add more concrete asserts - check that creds are null & target url exists
	})

	suite.Run("Additional Headers, Query Params and CSRFConfig are provided", func() {
		// TODO: ...
	})
}
