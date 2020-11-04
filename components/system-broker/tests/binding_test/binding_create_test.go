package binding_test

import (
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

const (
	serviceID        = "02d6080f-8c06-4d05-a7e0-cb15149261f8"
	planID           = "acf316ac-c129-4440-8052-5fc69a3b1486"
	brokerAPIVersion = "2.15"
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
	//err := suite.testContext.ConfigureResponse(suite.configURL, "query", "applications", appsMockResponse)
	//assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.PUT("/v2/service_instances/123/service_bindings/456").
		WithQuery("accepts_incomplete", "true").
		WithHeader("X-Broker-API-Version", brokerAPIVersion).
		WithJSON(map[string]string{"service_id": serviceID, "plan_id": planID}).
		Expect().Status(http.StatusInternalServerError)
}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsErrorOnPackageInstanceCreationShouldReturnError() {

}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsMultipleAuthsOnPackageInstanceCreationShouldReturnError() {

}

func (suite *BindCreateTestSuite) TestBindWhenDirectorReturnsFailedAuthOnPackageInstanceCreationShouldReturnError() {

}

func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundWithMultipleAuthsShouldReturnError() {

}

func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundWithFailedAuthShouldReturnError() {

}

func (suite *BindCreateTestSuite) TestBindWhenExistingCredentialIsFoundShouldReturnAccepted() {

}

func (suite *BindCreateTestSuite) TestBindWhenNewCredentialsAreCreatedShouldReturnAccepted() {

}
