package tests

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

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	input := fixAPIDefinitionInput()
	inStr, err := tc.graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := fixAddAPIToBundleRequest(bndl.ID, inStr)
	err = tc.RunOperation(ctx, req, &actualApi)
	require.NoError(t, err)

	pack := getBundle(t, ctx, application.ID, bndl.ID)
	require.Equal(t, bndl.ID, pack.ID)

	assertAPI(t, []*graphql.APIDefinitionInput{&input}, []*graphql.APIDefinitionExt{&actualApi})
	saveExample(t, req.Query(), "add api definition to bundle")
}

func TestManageAPIInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	api := addAPIToBundle(t, ctx, bndl.ID)

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

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	input := fixEventAPIDefinitionInput()
	inStr, err := tc.graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualEvent := graphql.EventAPIDefinitionExt{}
	req := fixAddEventAPIToBundleRequest(bndl.ID, inStr)
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

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	event := addEventToBundle(t, ctx, bndl.ID)

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

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	input := fixDocumentInput(t)
	inStr, err := tc.graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDocument := graphql.DocumentExt{}
	req := fixAddDocumentToBundleRequest(bndl.ID, inStr)
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

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	document := addDocumentToBundle(t, ctx, bndl.ID)

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

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	api := addAPIToBundle(t, ctx, bndl.ID)

	queryApiForBndl := fixAPIDefinitionInBundleRequest(application.ID, bndl.ID, api.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApiForBndl, &app)
	require.NoError(t, err)

	actualApi := app.Bundle.APIDefinition
	assert.Equal(t, api.ID, actualApi.ID)
	saveExample(t, queryApiForBndl.Query(), "query api definition")

}

func TestEventDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	event := addEventToBundle(t, ctx, bndl.ID)

	queryEventForBndl := fixEventDefinitionInBundleRequest(application.ID, bndl.ID, event.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventForBndl, &app)
	require.NoError(t, err)

	actualEvent := app.Bundle.EventDefinition
	assert.Equal(t, event.ID, actualEvent.ID)
	saveExample(t, queryEventForBndl.Query(), "query event definition")

}

func TestDocumentInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	doc := addDocumentToBundle(t, ctx, bndl.ID)

	queryDocForBndl := fixDocumentInBundleRequest(application.ID, bndl.ID, doc.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocForBndl, &app)
	require.NoError(t, err)

	actualDoc := app.Bundle.Document
	assert.Equal(t, doc.ID, actualDoc.ID)
	saveExample(t, queryDocForBndl.Query(), "query document")
}

func TestAPIDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	inputA := fixAPIDefinitionInputWithName("foo")
	addAPIToBundleWithInput(t, ctx, bndl.ID, inputA)

	inputB := fixAPIDefinitionInputWithName("bar")
	addAPIToBundleWithInput(t, ctx, bndl.ID, inputB)

	queryApisForBndl := fixAPIDefinitionsInBundleRequest(application.ID, bndl.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApisForBndl, &app)
	require.NoError(t, err)

	apis := app.Bundle.APIDefinitions
	require.Equal(t, 2, apis.TotalCount)
	assertAPI(t, []*graphql.APIDefinitionInput{&inputA, &inputB}, apis.Data)
	saveExample(t, queryApisForBndl.Query(), "query api definitions")
}

func TestEventDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	inputA := fixEventAPIDefinitionInputWithName("foo")
	addEventToBundleWithInput(t, ctx, bndl.ID, inputA)

	inputB := fixEventAPIDefinitionInputWithName("bar")
	addEventToBundleWithInput(t, ctx, bndl.ID, inputB)

	queryEventsForBndl := fixEventDefinitionsInBundleRequest(application.ID, bndl.ID)

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventsForBndl, &app)
	require.NoError(t, err)

	events := app.Bundle.EventDefinitions
	require.Equal(t, 2, events.TotalCount)
	assertEventsAPI(t, []*graphql.EventDefinitionInput{&inputA, &inputB}, events.Data)
	saveExample(t, queryEventsForBndl.Query(), "query event definitions")
}

func TestDocumentsInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bndlName := "test-bundle"
	bndl := createBundle(t, ctx, application.ID, bndlName)
	defer deleteBundle(t, ctx, bndl.ID)

	inputA := fixDocumentInputWithName(t, "foo")
	addDocumentToBundleWithInput(t, ctx, bndl.ID, inputA)

	inputB := fixDocumentInputWithName(t, "bar")
	addDocumentToBundleWithInput(t, ctx, bndl.ID, inputB)

	queryDocsForBndl := fixDocumentsInBundleRequest(application.ID, bndl.ID)

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocsForBndl, &app)
	require.NoError(t, err)

	docs := app.Bundle.Documents
	require.Equal(t, 2, docs.TotalCount)
	assertDocuments(t, []*graphql.DocumentInput{&inputA, &inputB}, docs.Data)
	saveExample(t, queryDocsForBndl.Query(), "query documents")
}

func TestAddBundle(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndlInput := fixBundleCreateInputWithRelatedObjects(t, "bndl-app-1")
	bndl, err := tc.graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := fixAddBundleRequest(application.ID, bndl)
	output := graphql.BundleExt{}

	// WHEN
	t.Log("Create bundle")
	err = tc.RunOperation(ctx, addBndlRequest, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	assertBundle(t, &bndlInput, &output)
	defer deleteBundle(t, ctx, output.ID)

	saveExample(t, addBndlRequest.Query(), "add bundle")

	bundleRequest := fixBundleRequest(application.ID, output.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, bundleRequest, &bndlFromAPI)
	require.NoError(t, err)

	assertBundle(t, &bndlInput, &output)
	saveExample(t, bundleRequest.Query(), "query bundle")
}

func TestQueryBundles(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndl1 := createBundle(t, ctx, application.ID, "bndl-app-1")
	defer deleteBundle(t, ctx, bndl1.ID)

	bndl2 := createBundle(t, ctx, application.ID, "bndl-app-2")
	defer deleteBundle(t, ctx, bndl2.ID)

	bundlesRequest := fixBundlesRequest(application.ID)
	bndlsFromAPI := graphql.ApplicationExt{}

	err := tc.RunOperation(ctx, bundlesRequest, &bndlsFromAPI)
	require.NoError(t, err)
	require.Equal(t, 2, len(bndlsFromAPI.Bundles.Data))

	saveExample(t, bundlesRequest.Query(), "query bundles")
}

func TestUpdateBundle(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndl := createBundle(t, ctx, application.ID, "bndl-app-1")
	defer deleteBundle(t, ctx, bndl.ID)

	bndlUpdateInput := fixBundleUpdateInput("bndl-app-1-up")
	bndlUpdate, err := tc.graphqlizer.BundleUpdateInputToGQL(bndlUpdateInput)
	require.NoError(t, err)

	updateBndlReq := fixUpdateBundleRequest(bndl.ID, bndlUpdate)
	output := graphql.Bundle{}

	// WHEN
	t.Log("Update bundle")
	err = tc.RunOperation(ctx, updateBndlReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, updateBndlReq.Query(), "update bundle")
}

func TestDeleteBundle(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bndl := createBundle(t, ctx, application.ID, "bndl-app-1")

	pkdDeleteReq := fixDeleteBundleRequest(bndl.ID)
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
