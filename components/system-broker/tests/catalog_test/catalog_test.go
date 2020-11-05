package catalog_test

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	appsEmptyResponse = `{
  "data": {
    "result": {
      "data": [],
      "pageInfo": {
        "startCursor": "",
        "endCursor": "",
        "hasNextPage": false
      },
      "totalCount": 0
    }
  }
}`
	appsEmptyPackagesMockResponse = `{
  "data": {
    "result": {
      "data": [
        {
          "id": "3e3cecce-74b3-4881-854e-58791021b522",
          "name": "test-app",
          "providerName": "test provider",
          "description": "a test application",
          "integrationSystemID": null,
          "labels": {
            "group": [
              "production",
              "experimental"
            ],
            "integrationSystemID": "",
            "name": "test-app",
            "scenarios": [
              "DEFAULT"
            ]
          },
          "status": {
            "condition": "INITIAL",
            "timestamp": "2020-10-21T18:23:59Z"
          },
          "webhooks": null,
          "healthCheckURL": "http://test-app.com/health",
          "packages": {
            "data": [],
            "pageInfo": {
              "startCursor": "",
              "endCursor": "",
              "hasNextPage": false
            },
            "totalCount": 0
          },
          "auths": null,
          "eventingConfiguration": {
            "defaultURL": ""
          }
        }
      ],
      "pageInfo": {
        "startCursor": "",
        "endCursor": "",
        "hasNextPage": false
      },
      "totalCount": 1
    }
  }
}`
	appsExpectedEmptyCatalog = `{"services":[]}` + "\n"
	appsPageResponse1        = `{
  "data": {
    "result": {
      "data": [
        {
          "id": "4310851a-04a0-4217-a2b8-1766c6a3f0fe",
          "name": "test-app2",
          "providerName": "test provider",
          "description": "a test application",
          "integrationSystemID": null,
          "labels": {
            "group": [
              "production",
              "experimental"
            ],
            "integrationSystemID": "",
            "name": "test-app2",
            "scenarios": [
              "DEFAULT"
            ]
          },
          "status": {
            "condition": "INITIAL",
            "timestamp": "2020-10-22T12:17:57Z"
          },
          "webhooks": null,
          "healthCheckURL": "http://test-app.com/health",
          "packages": {
            "data": [],
            "pageInfo": {
              "startCursor": "",
              "endCursor": "",
              "hasNextPage": false
            },
            "totalCount": 0
          },
          "auths": null,
          "eventingConfiguration": {
            "defaultURL": ""
          }
        },
        {
          "id": "5cf79030-0433-4b32-8618-9844085ca7a6",
          "name": "test-app",
          "providerName": "test provider",
          "description": "a test application",
          "integrationSystemID": null,
          "labels": {
            "group": [
              "production",
              "experimental"
            ],
            "integrationSystemID": "",
            "name": "test-app",
            "scenarios": [
              "DEFAULT"
            ]
          },
          "status": {
            "condition": "INITIAL",
            "timestamp": "2020-10-22T12:17:50Z"
          },
          "webhooks": null,
          "healthCheckURL": "http://test-app.com/health",
          "packages": {
            "data": [],
            "pageInfo": {
              "startCursor": "",
              "endCursor": "",
              "hasNextPage": false
            },
            "totalCount": 0
          },
          "auths": null,
          "eventingConfiguration": {
            "defaultURL": ""
          }
        }
      ],
      "pageInfo": {
        "startCursor": "",
        "endCursor": "RHBLdEo0ajlqRHEy",
        "hasNextPage": true
      },
      "totalCount": 4
    }
  }
}`
	appsPageResponse2 = `{
  "data": {
    "result": {
      "data": [
        {
          "id": "75ab9628-24d1-4e39-bdae-9c5042e908f2",
          "name": "test-app3",
          "providerName": "test provider",
          "description": "a test application",
          "integrationSystemID": null,
          "labels": {
            "group": [
              "production",
              "experimental"
            ],
            "integrationSystemID": "",
            "name": "test-app3",
            "scenarios": [
              "DEFAULT"
            ]
          },
          "status": {
            "condition": "INITIAL",
            "timestamp": "2020-10-22T12:18:00Z"
          },
          "webhooks": null,
          "healthCheckURL": "http://test-app.com/health",
          "packages": {
            "data": [],
            "pageInfo": {
              "startCursor": "",
              "endCursor": "",
              "hasNextPage": false
            },
            "totalCount": 0
          },
          "auths": null,
          "eventingConfiguration": {
            "defaultURL": ""
          }
        },
        {
          "id": "f945951c-bcaf-46af-a017-b3e2b575bdbd",
          "name": "test-app1",
          "providerName": "test provider",
          "description": "a test application",
          "integrationSystemID": null,
          "labels": {
            "group": [
              "production",
              "experimental"
            ],
            "integrationSystemID": "",
            "name": "test-app1",
            "scenarios": [
              "DEFAULT"
            ]
          },
          "status": {
            "condition": "INITIAL",
            "timestamp": "2020-10-22T12:17:53Z"
          },
          "webhooks": null,
          "healthCheckURL": "http://test-app.com/health",
          "packages": {
            "data": [],
            "pageInfo": {
              "startCursor": "",
              "endCursor": "",
              "hasNextPage": false
            },
            "totalCount": 0
          },
          "auths": null,
          "eventingConfiguration": {
            "defaultURL": ""
          }
        }
      ],
      "pageInfo": {
        "startCursor": "RHBLdEo0ajlqRHEy",
        "endCursor": "",
        "hasNextPage": false
      },
      "totalCount": 4
    }
  }
}`
	appsExpectedCatalogPaging = `{"services":[{"id":"4310851a-04a0-4217-a2b8-1766c6a3f0fe","name":"test-app2","description":"a test application","bindable":true,"plan_updateable":false,"plans":null,"metadata":{"displayName":"test-app2","group":["production","experimental"],"integrationSystemID":"","name":"test-app2","providerDisplayName":"test provider","scenarios":["DEFAULT"]}},{"id":"5cf79030-0433-4b32-8618-9844085ca7a6","name":"test-app","description":"a test application","bindable":true,"plan_updateable":false,"plans":null,"metadata":{"displayName":"test-app","group":["production","experimental"],"integrationSystemID":"","name":"test-app","providerDisplayName":"test provider","scenarios":["DEFAULT"]}},{"id":"75ab9628-24d1-4e39-bdae-9c5042e908f2","name":"test-app3","description":"a test application","bindable":true,"plan_updateable":false,"plans":null,"metadata":{"displayName":"test-app3","group":["production","experimental"],"integrationSystemID":"","name":"test-app3","providerDisplayName":"test provider","scenarios":["DEFAULT"]}},{"id":"f945951c-bcaf-46af-a017-b3e2b575bdbd","name":"test-app1","description":"a test application","bindable":true,"plan_updateable":false,"plans":null,"metadata":{"displayName":"test-app1","group":["production","experimental"],"integrationSystemID":"","name":"test-app1","providerDisplayName":"test provider","scenarios":["DEFAULT"]}}]}` + "\n"
	appsErrorResponse         = `{
  "errors": [
    {
      "message": "Internal Server Error",
      "path": [
        "result"
      ],
      "extensions": {
        "error": "InternalError",
        "error_code": 10
      }
    }
  ],
  "data": null
}`
)

func TestOSBCatalog(t *testing.T) {
	suite.Run(t, new(OSBCatalogTestSuite))
}

type OSBCatalogTestSuite struct {
	suite.Suite
	testContext *common.TestContext
	configURL   string
}

func (suite *OSBCatalogTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().
		WithEnvExtensions(func(e env.Environment, servers map[string]common.FakeServer) {
			e.Set("director_gql.page_size", 3)
		}).Build(suite.T())
	suite.configURL = suite.testContext.Servers[common.DirectorServer].URL() + "/config"
}

func (suite *OSBCatalogTestSuite) SetupTest() {
	suite.testContext.HttpClient.Post(suite.configURL+"/reset", "application/json", nil)
}

func (suite *OSBCatalogTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *OSBCatalogTestSuite) TeardownTest() {
	resp, err := suite.testContext.HttpClient.Get(suite.configURL + "/verify")
	assert.NoError(suite.T(), err)

	if resp.StatusCode == http.StatusInternalServerError {
		errorMsg, err := ioutil.ReadAll(resp.Body)
		assert.NoError(suite.T(), err)
		suite.Fail(string(errorMsg))
	}
	assert.Equal(suite.T(), resp.StatusCode, http.StatusOK)
}

func (suite *OSBCatalogTestSuite) TestEmptyResponse() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "applications", appsEmptyResponse)
	assert.NoError(suite.T(), err)
	suite.testContext.SystemBroker.GET("/v2/catalog").WithHeader("X-Broker-API-Version", "2.15").Expect().
		Status(http.StatusOK).
		Body().Equal(appsExpectedEmptyCatalog)
}

func (suite *OSBCatalogTestSuite) TestResponseWithOnePage() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "applications", appsEmptyPackagesMockResponse)
	assert.NoError(suite.T(), err)
	suite.testContext.SystemBroker.GET("/v2/catalog").WithHeader("X-Broker-API-Version", "2.15").Expect().
		Status(http.StatusOK).
		Body().Equal(appsExpectedEmptyCatalog)
}

func (suite *OSBCatalogTestSuite) TestResponseWithSeveralPages() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "applications", appsPageResponse1)
	assert.NoError(suite.T(), err)
	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "applications", appsPageResponse2)
	assert.NoError(suite.T(), err)
	suite.testContext.SystemBroker.GET("/v2/catalog").WithHeader("X-Broker-API-Version", "2.15").Expect().
		Status(http.StatusOK).
		Body().Equal(appsExpectedCatalogPaging)
}

func (suite *OSBCatalogTestSuite) TestErrorWhileFetchingApplicaitons() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "applications", appsPageResponse1)
	assert.NoError(suite.T(), err)
	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "applications", appsErrorResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET("/v2/catalog").WithHeader("X-Broker-API-Version", "2.15").Expect().
		Status(http.StatusInternalServerError).
		JSON().Object().Value("description").Equal("could not build catalog")
}
