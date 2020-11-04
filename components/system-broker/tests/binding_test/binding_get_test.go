package binding

import (
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
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
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsCredentialsWithMultipleAuthsShouldReturnError() {
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsUnsuccessfulCredentialsShouldReturnError() {
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsUnknownCredentialTypeShouldReturnError() {
}

func (suite *BindGetTestSuite) TestBindGetWhenDirectorReturnsValidCredentialsShouldReturnCredentials() {
}
