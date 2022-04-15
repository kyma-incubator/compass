package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestAddAPIToBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	input := fixtures.FixAPIDefinitionInput()
	inStr, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := fixtures.FixAddAPIToBundleRequest(bndl.ID, inStr)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &actualApi)
	require.NoError(t, err)

	pack := fixtures.GetBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndl.ID)
	require.Equal(t, bndl.ID, pack.ID)

	appWithBaseURL := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)
	assert.NotNil(t, appWithBaseURL.BaseURL)
	assert.Equal(t, input.TargetURL, *appWithBaseURL.BaseURL)

	assertions.AssertAPI(t, []*graphql.APIDefinitionInput{&input}, []*graphql.APIDefinitionExt{&actualApi})
	saveExample(t, req.Query(), "add api definition to bundle")
}

func TestAddAPIToBundleForApplicationWithAlreadySetBaseURL(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	baseURL := "https://compass.kyma.local/api"
	application, err := fixtures.RegisterApplicationWithBaseURL(t, ctx, certSecuredGraphQLClient, baseURL, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	input := fixtures.FixAPIDefinitionInput()
	inStr, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := fixtures.FixAddAPIToBundleRequest(bndl.ID, inStr)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &actualApi)
	require.NoError(t, err)

	pack := fixtures.GetBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndl.ID)
	require.Equal(t, bndl.ID, pack.ID)

	appWithBaseURL := fixtures.GetApplication(t, ctx, certSecuredGraphQLClient, tenantId, application.ID)
	assert.NotNil(t, appWithBaseURL.BaseURL)
	assert.NotEqual(t, input.TargetURL, *appWithBaseURL.BaseURL)

	assertions.AssertAPI(t, []*graphql.APIDefinitionInput{&input}, []*graphql.APIDefinitionExt{&actualApi})
	saveExample(t, req.Query(), "add api definition to bundle")
}

func TestManageAPIInBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	api := fixtures.AddAPIToBundle(t, ctx, certSecuredGraphQLClient, bndl.ID)

	apiUpdateInput := fixtures.FixAPIDefinitionInputWithName("new-name")
	apiUpdateGQL, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(apiUpdateInput)
	require.NoError(t, err)

	req := fixtures.FixUpdateAPIRequest(api.ID, apiUpdateGQL)

	var updatedAPI graphql.APIDefinitionExt
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &updatedAPI)
	require.NoError(t, err)

	assert.Equal(t, updatedAPI.ID, api.ID)
	assert.Equal(t, updatedAPI.Name, "new-name")
	saveExample(t, req.Query(), "update api definition")

	var deletedAPI graphql.APIDefinitionExt
	req = fixtures.FixDeleteAPIRequest(api.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &deletedAPI)
	require.NoError(t, err)

	assert.Equal(t, api.ID, deletedAPI.ID)
	saveExample(t, req.Query(), "delete api definition")
}

func TestAddEventDefinitionToBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	input := fixtures.FixEventAPIDefinitionInput()
	inStr, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualEvent := graphql.EventAPIDefinitionExt{}
	req := fixtures.FixAddEventAPIToBundleRequest(bndl.ID, inStr)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &actualEvent)
	require.NoError(t, err)

	assertions.AssertEventsAPI(t, []*graphql.EventDefinitionInput{&input}, []*graphql.EventAPIDefinitionExt{&actualEvent})
	saveExample(t, req.Query(), "add event definition to bundle")
}

func TestManageEventDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	event := fixtures.AddEventToBundle(t, ctx, certSecuredGraphQLClient, bndl.ID)

	eventUpdateInput := fixtures.FixEventAPIDefinitionInputWithName("new-name")
	eventUpdateGQL, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(eventUpdateInput)
	require.NoError(t, err)

	req := fixtures.FixUpdateEventAPIRequest(event.ID, eventUpdateGQL)

	var updatedEvent graphql.EventAPIDefinitionExt
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &updatedEvent)
	require.NoError(t, err)

	assert.Equal(t, updatedEvent.ID, event.ID)
	assert.Equal(t, updatedEvent.Name, "new-name")
	saveExample(t, req.Query(), "update event definition")

	var deletedEvent graphql.EventAPIDefinitionExt
	req = fixtures.FixDeleteEventAPIRequest(event.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &deletedEvent)
	require.NoError(t, err)

	assert.Equal(t, event.ID, deletedEvent.ID)
	saveExample(t, req.Query(), "delete event definition")
}

func TestAddDocumentToBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	input := fixtures.FixDocumentInput(t)
	inStr, err := testctx.Tc.Graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDocument := graphql.DocumentExt{}
	req := fixtures.FixAddDocumentToBundleRequest(bndl.ID, inStr)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &actualDocument)
	require.NoError(t, err)

	assertions.AssertDocuments(t, []*graphql.DocumentInput{&input}, []*graphql.DocumentExt{&actualDocument})
	saveExample(t, req.Query(), "add document to bundle")
}

func TestManageDocumentInBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	document := fixtures.AddDocumentToBundle(t, ctx, certSecuredGraphQLClient, bndl.ID)

	var deletedDocument graphql.DocumentExt
	req := fixtures.FixDeleteDocumentRequest(document.ID)
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, req, &deletedDocument)
	require.NoError(t, err)

	assert.Equal(t, document.ID, deletedDocument.ID)
	saveExample(t, req.Query(), "delete document")
}

func TestAPIDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	api := fixtures.AddAPIToBundle(t, ctx, certSecuredGraphQLClient, bndl.ID)

	queryApiForBndl := fixtures.FixAPIDefinitionInBundleRequest(application.ID, bndl.ID, api.ID)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryApiForBndl, &app)
	require.NoError(t, err)

	actualApi := app.Bundle.APIDefinition
	assert.Equal(t, api.ID, actualApi.ID)
	saveExample(t, queryApiForBndl.Query(), "query api definition")

}

func TestEventDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	event := fixtures.AddEventToBundle(t, ctx, certSecuredGraphQLClient, bndl.ID)

	queryEventForBndl := fixtures.FixEventDefinitionInBundleRequest(application.ID, bndl.ID, event.ID)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryEventForBndl, &app)
	require.NoError(t, err)

	actualEvent := app.Bundle.EventDefinition
	assert.Equal(t, event.ID, actualEvent.ID)
	saveExample(t, queryEventForBndl.Query(), "query event definition")

}

func TestDocumentInBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	doc := fixtures.AddDocumentToBundle(t, ctx, certSecuredGraphQLClient, bndl.ID)

	queryDocForBndl := fixtures.FixDocumentInBundleRequest(application.ID, bndl.ID, doc.ID)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryDocForBndl, &app)
	require.NoError(t, err)

	actualDoc := app.Bundle.Document
	assert.Equal(t, doc.ID, actualDoc.ID)
	saveExample(t, queryDocForBndl.Query(), "query document")
}

func TestAPIDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	inputA := fixtures.FixAPIDefinitionInputWithName("foo")
	fixtures.AddAPIToBundleWithInput(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID, inputA)

	inputB := fixtures.FixAPIDefinitionInputWithName("bar")
	fixtures.AddAPIToBundleWithInput(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID, inputB)

	queryApisForBndl := fixtures.FixAPIDefinitionsInBundleRequest(application.ID, bndl.ID)
	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryApisForBndl, &app)
	require.NoError(t, err)

	apis := app.Bundle.APIDefinitions
	require.Equal(t, 2, apis.TotalCount)
	assertions.AssertAPI(t, []*graphql.APIDefinitionInput{&inputA, &inputB}, apis.Data)
	saveExample(t, queryApisForBndl.Query(), "query api definitions")
}

func TestEventDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	inputA := fixtures.FixEventAPIDefinitionInputWithName("foo")
	fixtures.AddEventToBundleWithInput(t, ctx, certSecuredGraphQLClient, bndl.ID, inputA)

	inputB := fixtures.FixEventAPIDefinitionInputWithName("bar")
	fixtures.AddEventToBundleWithInput(t, ctx, certSecuredGraphQLClient, bndl.ID, inputB)

	queryEventsForBndl := fixtures.FixEventDefinitionsInBundleRequest(application.ID, bndl.ID)

	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryEventsForBndl, &app)
	require.NoError(t, err)

	events := app.Bundle.EventDefinitions
	require.Equal(t, 2, events.TotalCount)
	assertions.AssertEventsAPI(t, []*graphql.EventDefinitionInput{&inputA, &inputB}, events.Data)
	saveExample(t, queryEventsForBndl.Query(), "query event definitions")
}

func TestDocumentsInBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, appName, tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlName := "test-bundle"
	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, bndlName)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	inputA := fixtures.FixDocumentInputWithName(t, "foo")
	fixtures.AddDocumentToBundleWithInput(t, ctx, certSecuredGraphQLClient, bndl.ID, inputA)

	inputB := fixtures.FixDocumentInputWithName(t, "bar")
	fixtures.AddDocumentToBundleWithInput(t, ctx, certSecuredGraphQLClient, bndl.ID, inputB)

	queryDocsForBndl := fixtures.FixDocumentsInBundleRequest(application.ID, bndl.ID)

	app := graphql.ApplicationExt{}
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, queryDocsForBndl, &app)
	require.NoError(t, err)

	docs := app.Bundle.Documents
	require.Equal(t, 2, docs.TotalCount)
	assertions.AssertDocuments(t, []*graphql.DocumentInput{&inputA, &inputB}, docs.Data)
	saveExample(t, queryDocsForBndl.Query(), "query documents")
}

func TestAddBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndlInput := fixtures.FixBundleCreateInputWithRelatedObjects(t, "bndl-app-1")
	bndl, err := testctx.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixtures.FixAddBundleRequest(application.ID, bndl)
	output := graphql.BundleExt{}

	// WHEN
	t.Log("Create bundle")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, addBndlRequest, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	assertions.AssertBundle(t, &bndlInput, &output)
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, output.ID)

	saveExample(t, addBndlRequest.Query(), "add bundle")

	bundleRequest := fixtures.FixBundleRequest(application.ID, output.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bundleRequest, &bndlFromAPI)
	require.NoError(t, err)

	assertions.AssertBundle(t, &bndlInput, &output)
	saveExample(t, bundleRequest.Query(), "query bundle")
}

func TestQueryBundles(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)
	require.NoError(t, err)
	require.NotEmpty(t, application.ID)

	bndl1 := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl1.ID)

	bndl2 := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, "bndl-app-2")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl2.ID)

	bundlesRequest := fixtures.FixGetBundlesRequest(application.ID)
	bndlsFromAPI := graphql.ApplicationExt{}

	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, bundlesRequest, &bndlsFromAPI)
	require.NoError(t, err)
	require.Equal(t, 2, len(bndlsFromAPI.Bundles.Data))

	saveExample(t, bundlesRequest.Query(), "query bundles")
}

func TestUpdateBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, "bndl-app-1")
	defer fixtures.DeleteBundle(t, ctx, certSecuredGraphQLClient, tenantId, bndl.ID)

	bndlUpdateInput := fixtures.FixBundleUpdateInput("bndl-app-1-up")
	bndlUpdate, err := testctx.Tc.Graphqlizer.BundleUpdateInputToGQL(bndlUpdateInput)
	require.NoError(t, err)

	updateBndlReq := fixtures.FixUpdateBundleRequest(bndl.ID, bndlUpdate)
	output := graphql.Bundle{}

	// WHEN
	t.Log("Update bundle")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, updateBndlReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, updateBndlReq.Query(), "update bundle")
}

func TestDeleteBundle(t *testing.T) {
	ctx := context.Background()

	tenantId := tenant.TestTenants.GetDefaultTenantID()

	application, err := fixtures.RegisterApplication(t, ctx, certSecuredGraphQLClient, "app-test-bundle", tenantId)
	defer fixtures.CleanupApplication(t, ctx, certSecuredGraphQLClient, tenantId, &application)

	bndl := fixtures.CreateBundle(t, ctx, certSecuredGraphQLClient, tenantId, application.ID, "bndl-app-1")

	pkdDeleteReq := fixtures.FixDeleteBundleRequest(bndl.ID)
	output := graphql.Bundle{}

	// WHEN
	t.Log("Delete bundle")
	err = testctx.Tc.RunOperation(ctx, certSecuredGraphQLClient, pkdDeleteReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, pkdDeleteReq.Query(), "delete bundle")
}
