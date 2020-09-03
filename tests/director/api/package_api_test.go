package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestAddAPIToBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	input := fixAPIDefinitionInput()
	inStr, err := tc.graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := fixAddAPIToBundleRequest(pkg.ID, inStr)
	err = tc.RunOperation(ctx, req, &actualApi)
	require.NoError(t, err)

	pack := getBundle(t, ctx, application.ID, pkg.ID)
	require.Equal(t, pkg.ID, pack.ID)

	assertAPI(t, []*graphql.APIDefinitionInput{&input}, []*graphql.APIDefinitionExt{&actualApi})
	saveExample(t, req.Query(), "add api definition to bundle")
}

func TestManageAPIInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	api := addAPIToBundle(t, ctx, pkg.ID)

	apiUpdateInput := fixAPIDefinitionInputWithName("new-name")
	apiUpdateGQL, err := tc.graphqlizer.APIDefinitionInputToGQL(apiUpdateInput)
	require.NoError(t, err)

	req := fixUpdateAPIRequest(api.ID, apiUpdateGQL)

	var updatedAPI graphql.APIDefinitionExt
	err = tc.RunOperation(ctx, req, &updatedAPI)
	require.NoError(t, err)

	assert.Equal(t, updatedAPI.ID, api.ID)
	assert.Equal(t, updatedAPI.Name, "new-name")
	saveExample(t, req.Query(), "update api definition")

	var deletedAPI graphql.APIDefinitionExt
	req = fixDeleteAPIRequest(api.ID)
	err = tc.RunOperation(ctx, req, &deletedAPI)
	require.NoError(t, err)

	assert.Equal(t, api.ID, deletedAPI.ID)
	saveExample(t, req.Query(), "delete api definition")
}

func TestAddEventDefinitionToBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	input := fixEventAPIDefinitionInput()
	inStr, err := tc.graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualEvent := graphql.EventAPIDefinitionExt{}
	req := fixAddEventAPIToBundleRequest(pkg.ID, inStr)
	err = tc.RunOperation(ctx, req, &actualEvent)
	require.NoError(t, err)

	assertEventsAPI(t, []*graphql.EventDefinitionInput{&input}, []*graphql.EventAPIDefinitionExt{&actualEvent})
	saveExample(t, req.Query(), "add event definition to bundle")
}

func TestManageEventDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	event := addEventToBundle(t, ctx, pkg.ID)

	eventUpdateInput := fixEventAPIDefinitionInputWithName("new-name")
	eventUpdateGQL, err := tc.graphqlizer.EventDefinitionInputToGQL(eventUpdateInput)
	require.NoError(t, err)

	req := fixUpdateEventAPIRequest(event.ID, eventUpdateGQL)

	var updatedEvent graphql.EventAPIDefinitionExt
	err = tc.RunOperation(ctx, req, &updatedEvent)
	require.NoError(t, err)

	assert.Equal(t, updatedEvent.ID, event.ID)
	assert.Equal(t, updatedEvent.Name, "new-name")
	saveExample(t, req.Query(), "update event definition")

	var deletedEvent graphql.EventAPIDefinitionExt
	req = fixDeleteEventAPIRequest(event.ID)
	err = tc.RunOperation(ctx, req, &deletedEvent)
	require.NoError(t, err)

	assert.Equal(t, event.ID, deletedEvent.ID)
	saveExample(t, req.Query(), "delete event definition")
}

func TestAddDocumentToBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	input := fixDocumentInput(t)
	inStr, err := tc.graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDocument := graphql.DocumentExt{}
	req := fixAddDocumentToBundleRequest(pkg.ID, inStr)
	err = tc.RunOperation(ctx, req, &actualDocument)
	require.NoError(t, err)

	assertDocuments(t, []*graphql.DocumentInput{&input}, []*graphql.DocumentExt{&actualDocument})
	saveExample(t, req.Query(), "add document to bundle")
}

func TestManageDocumentInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	document := addDocumentToBundle(t, ctx, pkg.ID)

	var deletedDocument graphql.DocumentExt
	req := fixDeleteDocumentRequest(document.ID)
	err := tc.RunOperation(ctx, req, &deletedDocument)
	require.NoError(t, err)

	assert.Equal(t, document.ID, deletedDocument.ID)
	saveExample(t, req.Query(), "delete document")
}

func TestAPIDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	api := addAPIToBundle(t, ctx, pkg.ID)

	queryApiForPkg := fixAPIDefinitionInBundleRequest(application.ID, pkg.ID, api.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApiForPkg, &app)
	require.NoError(t, err)

	actualApi := app.Bundle.APIDefinition
	assert.Equal(t, api.ID, actualApi.ID)
	saveExample(t, queryApiForPkg.Query(), "query api definition")

}

func TestEventDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	event := addEventToBundle(t, ctx, pkg.ID)

	queryEventForPkg := fixEventDefinitionInBundleRequest(application.ID, pkg.ID, event.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventForPkg, &app)
	require.NoError(t, err)

	actualEvent := app.Bundle.EventDefinition
	assert.Equal(t, event.ID, actualEvent.ID)
	saveExample(t, queryEventForPkg.Query(), "query event definition")

}

func TestDocumentInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	doc := addDocumentToBundle(t, ctx, pkg.ID)

	queryDocForPkg := fixDocumentInBundleRequest(application.ID, pkg.ID, doc.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocForPkg, &app)
	require.NoError(t, err)

	actualDoc := app.Bundle.Document
	assert.Equal(t, doc.ID, actualDoc.ID)
	saveExample(t, queryDocForPkg.Query(), "query document")
}

func TestAPIDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	inputA := fixAPIDefinitionInputWithName("foo")
	addAPIToBundleWithInput(t, ctx, pkg.ID, inputA)

	inputB := fixAPIDefinitionInputWithName("bar")
	addAPIToBundleWithInput(t, ctx, pkg.ID, inputB)

	queryApisForPkg := fixAPIDefinitionsInBundleRequest(application.ID, pkg.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApisForPkg, &app)
	require.NoError(t, err)

	apis := app.Bundle.APIDefinitions
	require.Equal(t, 2, apis.TotalCount)
	assertAPI(t, []*graphql.APIDefinitionInput{&inputA, &inputB}, apis.Data)
	saveExample(t, queryApisForPkg.Query(), "query api definitions")
}

func TestEventDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	inputA := fixEventAPIDefinitionInputWithName("foo")
	addEventToBundleWithInput(t, ctx, pkg.ID, inputA)

	inputB := fixEventAPIDefinitionInputWithName("bar")
	addEventToBundleWithInput(t, ctx, pkg.ID, inputB)

	queryEventsForPkg := fixEventDefinitionsInBundleRequest(application.ID, pkg.ID)

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventsForPkg, &app)
	require.NoError(t, err)

	events := app.Bundle.EventDefinitions
	require.Equal(t, 2, events.TotalCount)
	assertEventsAPI(t, []*graphql.EventDefinitionInput{&inputA, &inputB}, events.Data)
	saveExample(t, queryEventsForPkg.Query(), "query event definitions")
}

func TestDocumentsInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-bundle"
	pkg := createBundle(t, ctx, application.ID, pkgName)
	defer deleteBundle(t, ctx, pkg.ID)

	inputA := fixDocumentInputWithName(t, "foo")
	addDocumentToBundleWithInput(t, ctx, pkg.ID, inputA)

	inputB := fixDocumentInputWithName(t, "bar")
	addDocumentToBundleWithInput(t, ctx, pkg.ID, inputB)

	queryDocsForPkg := fixDocumentsInBundleRequest(application.ID, pkg.ID)

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocsForPkg, &app)
	require.NoError(t, err)

	docs := app.Bundle.Documents
	require.Equal(t, 2, docs.TotalCount)
	assertDocuments(t, []*graphql.DocumentInput{&inputA, &inputB}, docs.Data)
	saveExample(t, queryDocsForPkg.Query(), "query documents")
}

func TestAddBundle(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	pkgInput := fixBundleCreateInputWithRelatedObjects(t, "pkg-app-1")
	pkg, err := tc.graphqlizer.BundleCreateInputToGQL(pkgInput)
	require.NoError(t, err)

	addPkgRequest := fixAddBundleRequest(application.ID, pkg)
	output := graphql.BundleExt{}

	// WHEN
	t.Log("Create bundle")
	err = tc.RunOperation(ctx, addPkgRequest, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	assertBundle(t, &pkgInput, &output)
	defer deleteBundle(t, ctx, output.ID)

	saveExample(t, addPkgRequest.Query(), "add bundle")

	bundleRequest := fixBundleRequest(application.ID, output.ID)
	pkgFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, bundleRequest, &pkgFromAPI)
	require.NoError(t, err)

	assertBundle(t, &pkgInput, &output)
	saveExample(t, bundleRequest.Query(), "query bundle")
}

func TestQueryBundles(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	pkg1 := createBundle(t, ctx, application.ID, "pkg-app-1")
	defer deleteBundle(t, ctx, pkg1.ID)

	pkg2 := createBundle(t, ctx, application.ID, "pkg-app-2")
	defer deleteBundle(t, ctx, pkg2.ID)

	bundlesRequest := fixBundlesRequest(application.ID)
	pkgsFromAPI := graphql.ApplicationExt{}

	err := tc.RunOperation(ctx, bundlesRequest, &pkgsFromAPI)
	require.NoError(t, err)
	require.Equal(t, 2, len(pkgsFromAPI.Bundles.Data))

	saveExample(t, bundlesRequest.Query(), "query bundles")
}

func TestUpdateBundle(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	pkg := createBundle(t, ctx, application.ID, "pkg-app-1")
	defer deleteBundle(t, ctx, pkg.ID)

	pkgUpdateInput := fixBundleUpdateInput("pkg-app-1-up")
	pkgUpdate, err := tc.graphqlizer.BundleUpdateInputToGQL(pkgUpdateInput)
	require.NoError(t, err)

	updatePkgReq := fixUpdateBundleRequest(pkg.ID, pkgUpdate)
	output := graphql.Bundle{}

	// WHEN
	t.Log("Update bundle")
	err = tc.RunOperation(ctx, updatePkgReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, updatePkgReq.Query(), "update bundle")
}

func TestDeleteBundle(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	pkg := createBundle(t, ctx, application.ID, "pkg-app-1")

	pkdDeleteReq := fixDeleteBundleRequest(pkg.ID)
	output := graphql.Bundle{}

	// WHEN
	t.Log("Delete bundle")
	err := tc.RunOperation(ctx, pkdDeleteReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, pkdDeleteReq.Query(), "delete bundle")
}
