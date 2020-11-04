package binding

import (
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

func TestBindDelete(t *testing.T) {
	suite.Run(t, new(UnbindTestSuite))
}

type UnbindTestSuite struct {
	suite.Suite
	testContext *common.TestContext
	configURL   string
}

func (suite *UnbindTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.configURL = suite.testContext.Servers[common.DirectorServer].URL() + "/config"
}

func (suite *UnbindTestSuite) SetupTest() {
	http.DefaultClient.Post(suite.configURL+"/reset", "application/json", nil)
}

func (suite *UnbindTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *UnbindTestSuite) TestUnbindWithoutAcceptsIncompleteHeaderShouldReturnUnprocessableEntity() {
}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsErrorOnFindCredentialsShouldReturnError() {

}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsNotFoundShouldReturnGone() {

}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsCredentialsWithMultipleAuthsShouldReturnError() {

}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsUnusedCredentialsShouldReturnAccepted() {

}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsErrorOnCredentialsDeletionShouldReturnError() {

}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorReturnsNotFoundOnCredentialsDeletionShouldReturnGone() {

}

func (suite *UnbindTestSuite) TestUnbindWhenDirectorAcceptsCredentialsDeletionRequestShouldReturnAccepted() {

}
