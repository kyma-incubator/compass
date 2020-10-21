package healthcheck_test

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
)

const dummyResponse = `{
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

func TestOSBCatalog(t *testing.T) {
	suite.Run(t, new(OSBCatalogTestSuite))
}

type OSBCatalogTestSuite struct {
	suite.Suite

	testContext *common.TestContext
}

func (suite *OSBCatalogTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
}

func (suite *OSBCatalogTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *OSBCatalogTestSuite) TestCatalogIsOK() {
	url := suite.testContext.Servers[common.DirectorServer].URL() + "/config"

	body := common.ConfigRequestBody{
		GraphqlQueryKey: common.GraphqlQueryKey{
			Type: "query",
			Name: "applications",
		},
		Response: dummyResponse,
	}
	jsonBody, err := json.Marshal(body)

	assert.NoError(suite.T(), err)
	http.DefaultClient.Post(url, "application/json", bytes.NewReader(jsonBody))

	suite.testContext.SystemBroker.GET("/v2/catalog").WithHeader("X-Broker-API-Version", "2.15").Expect().
		Status(http.StatusOK).
		Body().Equal(`{"status": "success"}`)
}
