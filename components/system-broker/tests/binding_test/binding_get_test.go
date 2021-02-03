package binding_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	schema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
)

var (
	bundleWithAPIsResp = `{
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

	unknownAuth                     = `"credential": { "a": "aa", "bb": "bb"}`
	basicAuth                       = `"credential": { "username": "asd", "password": "asd"}`
	oAuth                           = `"credential": { "clientId": "test-id", "clientSecret": "test-secret", "url": "https://api.test.com/oauth/token" }`
	additionalHeadersSerialized     = `"additionalHeadersSerialized": "{\"header-A\": [\"ha1\", \"ha2\"], \"header-B\": [\"hb1\", \"hb2\"]}"`
	additionalQueryParamsSerialized = `"additionalQueryParamsSerialized": "{\"qA\": [\"qa1\", \"qa2\"], \"qB\": [\"qb1\", \"qb2\"]}"`
	requestAuth                     = `"requestAuth": { "csrf": { "tokenEndpointURL": "https://some-url/token"}}`
)

func TestBindGet(t *testing.T) {
	suite.Run(t, new(BindGetTestSuite))
}

type BindGetTestSuite struct {
	suite.Suite
	testContext       *common.TestContext
	mockedDirectorURL string
}

func (suite *BindGetTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.mockedDirectorURL = suite.testContext.Servers[common.DirectorServer].URL()
}

func (suite *BindGetTestSuite) SetupTest() {
	_, err := http.DefaultClient.Post(suite.mockedDirectorURL+"/config/reset", "application/json", nil)
	assert.NoError(suite.T(), err)
}

func (suite *BindGetTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *BindGetTestSuite) TearDownTest() {
	resp, err := suite.testContext.HttpClient.Get(suite.mockedDirectorURL + "/verify")
	assert.NoError(suite.T(), err)

	if resp.StatusCode == http.StatusInternalServerError {
		errorMsg, err := ioutil.ReadAll(resp.Body)
		assert.NoError(suite.T(), err)
		suite.Fail(string(errorMsg))
	}
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsUnauthorizedOnFindCredentialsShouldReturnUnauthorized() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth", `{"error": "insufficient scopes provided"}`)
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusUnauthorized)

	resp.JSON().Path("$.description").String().Contains("unauthorized: insufficient scopes")
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsContextWithMismatchedInstanceOnFindCredentialsShouldReturnNotFound() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth",
		fmt.Sprintf(bundleInstanceAuthResponse, bindingID, schema.BundleInstanceAuthStatusConditionPending, "test", bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusNotFound)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsFailedConditionCredentialsShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth",
		fmt.Sprintf(bundleWithAPIsResp, basicAuth, schema.BundleInstanceAuthStatusConditionFailed, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsUnknownCredentialTypeShouldReturnCredentialsWithNoAuth() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth",
		fmt.Sprintf(bundleWithAPIsResp, unknownAuth, schema.BundleInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	resp := suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusOK)

	resp.JSON().Path("$.credentials").NotNull()
	resp.JSON().Path("$.credentials.credentials_type").Equal("no_auth")
	resp.JSON().Path("$.credentials.auth_details").Object().Values().Length().Equal(0)
	resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Price").Equal("https://api.cloud.com/api/InboundTestPrice")
	resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Stock").Equal("https://api.cloud.com/InboundTestStock")
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsUnusedCredentialShouldReturnError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth",
		fmt.Sprintf(bundleWithAPIsResp, basicAuth, schema.BundleInstanceAuthStatusConditionUnused, instanceID, bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusNotFound)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnedCredentialsWithMismatchedContextInstanceShouldReturnNotFound() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth",
		fmt.Sprintf(bundleWithAPIsResp, basicAuth, schema.BundleInstanceAuthStatusConditionFailed, "test", bindingID))
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET(bindingPath).
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsValidCredentialsShouldReturnCredentials() {
	suite.Run("Basic Authentication", func() {
		err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth",
			fmt.Sprintf(bundleWithAPIsResp, basicAuth, schema.BundleInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
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
		err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth",
			fmt.Sprintf(bundleWithAPIsResp, oAuth, schema.BundleInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
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
		err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth",
			fmt.Sprintf(bundleWithAPIsResp, "", schema.BundleInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
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
		err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "bundleByInstanceAuth",
			fmt.Sprintf(bundleWithAPIsResp, fmt.Sprintf(`%s, %s, %s, %s`, basicAuth, additionalHeadersSerialized, additionalQueryParamsSerialized, requestAuth), schema.BundleInstanceAuthStatusConditionSucceeded, instanceID, bindingID))
		assert.NoError(suite.T(), err)

		resp := suite.testContext.SystemBroker.GET(bindingPath).
			WithHeader("X-Broker-API-Version", brokerAPIVersion).
			Expect().Status(http.StatusOK)

		resp.JSON().Path("$.credentials").NotNull()
		resp.JSON().Path("$.credentials.credentials_type").Equal("basic_auth")
		resp.JSON().Path("$.credentials.auth_details.auth.username").Equal("asd")
		resp.JSON().Path("$.credentials.auth_details.auth.password").Equal("asd")

		resp.JSON().Path("$.credentials.auth_details.csrf_config.token_url").Equal("https://some-url/token")
		resp.JSON().Path("$.credentials.auth_details.request_parameters.headers").Object().Value("header-A").Array().Elements("ha1", "ha2")
		resp.JSON().Path("$.credentials.auth_details.request_parameters.headers").Object().Value("header-B").Array().Elements("hb1", "hb2")
		resp.JSON().Path("$.credentials.auth_details.request_parameters.query_parameters").Object().Value("qA").Array().Elements("qa1", "qa2")
		resp.JSON().Path("$.credentials.auth_details.request_parameters.query_parameters").Object().Value("qB").Array().Elements("qb1", "qb2")

		resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Price").Equal("https://api.cloud.com/api/InboundTestPrice")
		resp.JSON().Path("$.credentials.target_urls").Object().Value("API Cloud - Inbound Test Stock").Equal("https://api.cloud.com/InboundTestStock")
	})
}
