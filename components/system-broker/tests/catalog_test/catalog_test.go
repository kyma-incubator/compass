package catalog_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/kyma-incubator/compass/components/system-broker/tests/common"
	"github.com/pivotal-cf/brokerapi/v7/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestOSBCatalog(t *testing.T) {
	suite.Run(t, new(OSBCatalogTestSuite))
}

type OSBCatalogTestSuite struct {
	suite.Suite
	testContext       *common.TestContext
	mockedDirectorURL string
}

func (suite *OSBCatalogTestSuite) SetupSuite() {
	suite.testContext = common.NewTestContextBuilder().Build(suite.T())
	suite.mockedDirectorURL = suite.testContext.Servers[common.DirectorServer].URL()
}

func (suite *OSBCatalogTestSuite) SetupTest() {
	_, err := suite.testContext.HttpClient.Post(suite.mockedDirectorURL+"/config/reset", "application/json", nil)
	assert.NoError(suite.T(), err)
}

func (suite *OSBCatalogTestSuite) TearDownSuite() {
	suite.testContext.CleanUp()
}

func (suite *OSBCatalogTestSuite) TearDownTest() {
	resp, err := suite.testContext.HttpClient.Get(suite.mockedDirectorURL + "/verify")
	assert.NoError(suite.T(), err)

	if resp.StatusCode == http.StatusInternalServerError {
		errorMsg, err := ioutil.ReadAll(resp.Body)
		assert.NoError(suite.T(), err)
		suite.Fail(string(errorMsg))
	}
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *OSBCatalogTestSuite) TestEmptyCatalog() {
	testCatalog(suite, 1, 0, 0, 0, 0)
}

func (suite *OSBCatalogTestSuite) TestAppWithEverything() {
	testCatalog(suite, 1, 1, 1, 1, 1)
}

func (suite *OSBCatalogTestSuite) TestWithManyApps() {
	testCatalog(suite, 3, 3, 3, 3, 3)
}

func (suite *OSBCatalogTestSuite) TestWhenDirectorReturnsError() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "applications", `{"error": "Test-error"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET("/v2/catalog").WithHeader("X-Broker-API-Version", "2.15").Expect().
		Status(http.StatusInternalServerError).
		JSON().Object().Value("description").String().Contains("could not build catalog")
}

func (suite *OSBCatalogTestSuite) TestWhenDirectorReturnsUnauthorizedShouldReturnUnauthorized() {
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "applications", `{"error": "insufficient scopes provided"}`)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET("/v2/catalog").WithHeader("X-Broker-API-Version", "2.15").Expect().
		Status(http.StatusUnauthorized).
		JSON().Object().Value("description").String().Contains("unauthorized: insufficient scopes")
}

func testCatalog(suite *OSBCatalogTestSuite, apps, bundles, apiDefs, eventDefs, docs int) {
	mockedApps := genMockApps(apps, bundles, apiDefs, eventDefs, docs)
	mockedAppsResponse := toDirectorResponse(suite.T(), mockedApps)
	expectedCatalog := toCatalog(suite.T(), mockedApps)
	err := suite.testContext.ConfigureResponse(suite.mockedDirectorURL+"/config", "query", "applications", mockedAppsResponse)
	assert.NoError(suite.T(), err)

	suite.testContext.SystemBroker.GET("/v2/catalog").WithHeader("X-Broker-API-Version", "2.15").Expect().
		Status(http.StatusOK).
		JSON().Equal(expectedCatalog)
}

func toDirectorResponse(t *testing.T, mockApp interface{}) string {
	fixture := map[string]interface{}{
		"data": mockApp,
	}

	appsEmptyResponseBytes, err := json.Marshal(fixture)
	assert.NoError(t, err)
	apps := string(appsEmptyResponseBytes)
	return apps
}

func toCatalog(t *testing.T, mockApp director.ApplicationsOutput) interface{} {
	converter := osb.CatalogConverter{}
	svcs := make([]domain.Service, 0)
	for _, app := range mockApp.Result.Data {
		s, err := converter.Convert(app)
		assert.NoError(t, err)

		if len(s.Plans) > 0 {
			svcs = append(svcs, *s)
		}
	}

	catalogObj := map[string]interface{}{
		"services": svcs,
	}

	return catalogObj
}

func genMockApps(n, bundles, apiDefs, eventDefs, docs int) director.ApplicationsOutput {
	result := director.ApplicationsOutput{
		Result: &graphql.ApplicationPageExt{
			ApplicationPage: graphql.ApplicationPage{
				Data:     []*graphql.Application{},
				PageInfo: &graphql.PageInfo{},
			},
		},
	}

	for i := 0; i < n; i++ {
		result.Result.Data = append(result.Result.Data, genMockApp(bundles, apiDefs, eventDefs, docs))
	}
	return result
}

func genMockApp(bundles, apiDefs, eventDefs, docs int) *graphql.ApplicationExt {
	id := uuid.New().String()
	result := &graphql.ApplicationExt{
		Application: graphql.Application{
			BaseEntity: &graphql.BaseEntity{
				ID: id,
			},
			Name: "name-" + id,
		},
		Bundles: graphql.BundlePageExt{
			Data: []*graphql.BundleExt{},
		},
	}
	for i := 0; i < bundles; i++ {
		result.Bundles.Data = append(result.Bundles.Data, genMockBundle(apiDefs, eventDefs, docs))
	}
	return result
}

func genMockBundle(apiDefs, eventDefs, docs int) *graphql.BundleExt {
	id := uuid.New().String()
	result := &graphql.BundleExt{
		Bundle: graphql.Bundle{
			BaseEntity: &graphql.BaseEntity{
				ID: id,
			},
			Name: "name-" + id,
		},
		APIDefinitions: graphql.APIDefinitionPageExt{
			Data: []*graphql.APIDefinitionExt{},
		},
		EventDefinitions: graphql.EventAPIDefinitionPageExt{
			Data: []*graphql.EventAPIDefinitionExt{},
		},
		Documents: graphql.DocumentPageExt{
			Data: []*graphql.DocumentExt{},
		},
	}
	for i := 0; i < apiDefs; i++ {
		result.APIDefinitions.Data = append(result.APIDefinitions.Data, genApiDef())
	}
	for i := 0; i < eventDefs; i++ {
		result.EventDefinitions.Data = append(result.EventDefinitions.Data, genEventDef())
	}
	for i := 0; i < docs; i++ {
		result.Documents.Data = append(result.Documents.Data, genDoc())
	}
	return result
}

func genApiDef() *graphql.APIDefinitionExt {
	id := uuid.New().String()
	return &graphql.APIDefinitionExt{
		APIDefinition: graphql.APIDefinition{
			BaseEntity: &graphql.BaseEntity{
				ID: id,
			},
			Name: "name-" + id,
		},
	}
}

func genEventDef() *graphql.EventAPIDefinitionExt {
	id := uuid.New().String()
	return &graphql.EventAPIDefinitionExt{
		EventDefinition: graphql.EventDefinition{
			BaseEntity: &graphql.BaseEntity{
				ID: id,
			},
			Name: "name-" + id,
		},
	}
}

func genDoc() *graphql.DocumentExt {
	id := uuid.New().String()
	return &graphql.DocumentExt{
		Document: graphql.Document{
			BaseEntity: &graphql.BaseEntity{
				ID: id,
			},
			DisplayName: "display-name-" + id,
		},
	}
}
