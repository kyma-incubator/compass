package binding_test

import (
	"fmt"
	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
			"instanceAuth": 
				{
					"id": "16cfe0b3-9633-4441-a639-15126ff3z4b9",
					"auth": {
						%s
					},
					"status": {
						"condition": "%s",
						"timestamp": "2020-04-20T10:54:07Z",
						"message": "Credentials were provided.",
						"reason": "CredentialsProvided"
					},
					"context": "{\"instance_id\": \"%s\", \"binding_id\": \"%s\"}"
				}
			,	
			"apiDefinitions": {
				"data": [
					{
						"id": "1a23bd57-ef6e-46fd-a59s-15632cd3c410",
						"name": "API Cloud - Inbound Test Price",
						"targetURL": "https://api.cloud.com/api/InboundTestPrice"
					},
					{
						"id": "2e635cc3-fc9b-4a8d-zb4e-iaf73cbd8846",
						"name": "API Cloud - Inbound Test Stock",
						"targetURL": "https://api.cloud.com/InboundTestStock"
					}
				]
			}
		}
	}
}`

	unknownAuth           = `"credential": { "a": "aa", "bb": "bb"}`
	basicAuth             = `"credential": { "username": "asd", "password": "asd"}`
	oAuth                 = `"credential": { "clientId": "test-id", "clientSecret": "test-secret", "url": "https://api.test.com/oauth/token" }`
	additionalHeaders     = `"additionalHeaders": "\"key\": \"value\""`
	additionalQueryParams = `"additionalQueryParams": "\"key\": \"value\""`
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
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsContextWithMismatchedInstanceOnFindCredentialsShouldReturnNotFound() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth",
		fmt.Sprintf(packageInstanceAuthResponse, schema.PackageInstanceAuthStatusConditionPending, "test", bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusNotFound)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsFailedConditionCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth",
		fmt.Sprintf(packageWithAPIsResp, basicAuth, schema.PackageInstanceAuthStatusConditionFailed, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsUnknownCredentialTypeShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth",
		fmt.Sprintf(packageWithAPIsResp, basicAuth, schema.PackageInstanceAuthStatusConditionUnused, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusNotFound)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnedCredentialsWithMismatchedContextInstanceShouldReturnNotFound() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth",
		fmt.Sprintf(packageWithAPIsResp, basicAuth, schema.PackageInstanceAuthStatusConditionFailed, "test", bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsValidCredentialsShouldReturnCredentials() {
	suite.Run("Basic Authentication", func() {
		err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth",
			fmt.Sprintf(packageWithAPIsResp, basicAuth, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
		assert.NoError(suite.T(), err)

		resp := suite.testContext.SystemBroker.GET(bindingPath).
			WithHeader("X-Broker-API-Version", brokerAPIVersion).
			Expect().Status(http.StatusOK)

		resp.JSON().Path("$.credentials").NotNull()
		resp.JSON().Path("$.credentials.credentials_type").Equal("basic_auth")
		resp.JSON().Path("$.credentials.auth_details.auth.username").Equal("asd")
		resp.JSON().Path("$.credentials.auth_details.auth.password").Equal("asd")
		resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Price").Equal("https://api.cloud.com/api/InboundTestPrice")
		resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Stock").Equal("https://api.cloud.com/InboundTestStock")
	})

	suite.Run("OAuth", func() {
		err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth",
			fmt.Sprintf(packageWithAPIsResp, oAuth, schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
		assert.NoError(suite.T(), err)

		resp := suite.testContext.SystemBroker.GET(bindingPath).
			WithHeader("X-Broker-API-Version", brokerAPIVersion).
			Expect().Status(http.StatusOK)

		resp.JSON().Path("$.credentials").NotNull()
		resp.JSON().Path("$.credentials.credentials_type").Equal("oauth")
		resp.JSON().Path("$.credentials.auth_details.auth.clientId").Equal("test-id")
		resp.JSON().Path("$.credentials.auth_details.auth.clientSecret").Equal("test-secret")
		resp.JSON().Path("$.credentials.auth_details.auth.tokenUrl").Equal("https://api.test.com/oauth/token")
		resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Price").Equal("https://api.cloud.com/api/InboundTestPrice")
		resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Stock").Equal("https://api.cloud.com/InboundTestStock")
	})

	suite.Run("None", func() {
		err := suite.testContext.ConfigureResponse(suite.configURL, "query", "packageByInstanceAuth",
			fmt.Sprintf(packageWithAPIsResp, "", schema.PackageInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
		assert.NoError(suite.T(), err)

		resp := suite.testContext.SystemBroker.GET(bindingPath).
			WithHeader("X-Broker-API-Version", brokerAPIVersion).
			Expect().Status(http.StatusOK)

		resp.JSON().Path("$.credentials").NotNull()
		resp.JSON().Path("$.credentials.credentials_type").Equal("no_auth")
		resp.JSON().Path("$.credentials.auth_details").Object().Values().Length().Equal(0)
		resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Price").Equal("https://api.cloud.com/api/InboundTestPrice")
		resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Stock").Equal("https://api.cloud.com/InboundTestStock")
	})

	suite.Run("Additional Headers, Query Params and CSRFConfig are provided", func() {
		// TODO: ...
	})
}
