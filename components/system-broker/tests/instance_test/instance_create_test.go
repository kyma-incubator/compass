package instance_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	serviceID        = "02d6080f-8c06-4d05-a7e0-cb15149261f8"
	planID           = "acf316ac-c129-4440-8052-5fc69a3b1486"
	brokerAPIVersion = "2.15"
)

var (
	appsMockResponse = fmt.Sprintf(`{
 "data": {
   "result": {
     "data": [
       {
          "id": "%s",
		  "name": "varkes",
     	  "providerName": "",
    	  "description": "",
		  "packages": {
			"data": [
			  {
				"id": "%s",
				"name": "ac",
				"description": ""
			  }
			]
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
}`, serviceID, planID)

	appMockResponse = fmt.Sprintf(`{
  "data": {
    "result": {
      "id": "%s",
      "name": "varkes",
      "providerName": "",
      "description": "",
      "packages": {
        "data": [
          {
            "id": "%s",
            "name": "ac",
            "description": ""
          }
        ]
      }
    }
  }
}`, serviceID, planID)
)

func TestInstanceCreate(t *testing.T) {
	suite.Run(t, new(InstanceCreateTestSuite))
}

type InstanceCreateTestSuite struct {
	suite.Suite
	testContext *common.TestContext
	configURL   string
}

func (suite *InstanceCreateTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.configURL = suite.testContext.Servers[common.DirectorServer].URL() + "/config"
}

func (suite *InstanceCreateTestSuite) SetupTest() {
	http.DefaultClient.Post(suite.configURL+"/reset", "application/json", nil)
}

func (suite *InstanceCreateTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *InstanceCreateTestSuite) TestProvision() {
	err := suite.testContext.ConfigureResponse(suite.configURL, "query", "applications", appsMockResponse)
	assert.NoError(suite.T(), err)
	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "application", appMockResponse)
	assert.NoError(suite.T(), err)
	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "application", appMockResponse)
	assert.NoError(suite.T(), err)
	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "application", appMockResponse)
	assert.NoError(suite.T(), err)
	err = suite.testContext.ConfigureResponse(suite.configURL, "query", "application", appMockResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT("/v2/service_instances/123").WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusCreated).Body().Equal("{}\n")
}
