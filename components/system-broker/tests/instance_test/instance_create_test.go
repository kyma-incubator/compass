package instance_test

import (
	"fmt"
	"io/ioutil"
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
		  "bundles": {
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
)

func TestInstanceProvision(t *testing.T) {
	suite.Run(t, new(InstanceProvisionTestSuite))
}

type InstanceProvisionTestSuite struct {
	suite.Suite
	testContext       *common.TestContext
	mockedDirectorURL string
}

func (suite *InstanceProvisionTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.mockedDirectorURL = suite.testContext.Servers[common.DirectorServer].URL()
}

func (suite *InstanceProvisionTestSuite) SetupTest() {
	_, err := http.DefaultClient.Post(suite.mockedDirectorURL+"/config/reset", "application/json", nil)
	assert.NoError(suite.T(), err)
}

func (suite *InstanceProvisionTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *InstanceProvisionTestSuite) TearDownTest() {
	resp, err := suite.testContext.HttpClient.Get(suite.mockedDirectorURL + "/verify")
	assert.NoError(suite.T(), err)

	if resp.StatusCode == http.StatusInternalServerError {
		errorMsg, err := ioutil.ReadAll(resp.Body)
		assert.NoError(suite.T(), err)
		suite.Fail(string(errorMsg))
	}
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *InstanceProvisionTestSuite) TestProvision() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "applications", appsMockResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT("/v2/service_instances/123").WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusCreated).Body().Equal("{}\n")
}
