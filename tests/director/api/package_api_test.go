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

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	input := fixAPIDefinitionInput()
	inStr, err := tc.graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := fixAddAPIToBundleRequest(bundle.ID, inStr)
	err = tc.RunOperation(ctx, req, &actualApi)
	require.NoError(t, err)

	pack := getBundle(t, ctx, application.ID, bundle.ID)
	require.Equal(t, bundle.ID, pack.ID)

	assertAPI(t, []*graphql.APIDefinitionInput{&input}, []*graphql.APIDefinitionExt{&actualApi})
	saveExample(t, req.Query(), "add api definition to bundle")
}

func TestManageAPIInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	api := addAPIToBundle(t, ctx, bundle.ID)

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

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	input := fixEventAPIDefinitionInput()
	inStr, err := tc.graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualEvent := graphql.EventAPIDefinitionExt{}
	req := fixAddEventAPIToBundleRequest(bundle.ID, inStr)
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

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	event := addEventToBundle(t, ctx, bundle.ID)

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

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	input := fixDocumentInput(t)
	inStr, err := tc.graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDocument := graphql.DocumentExt{}
	req := fixAddDocumentToBundleRequest(bundle.ID, inStr)
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

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	document := addDocumentToBundle(t, ctx, bundle.ID)

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

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	api := addAPIToBundle(t, ctx, bundle.ID)

	queryApiForBundle := fixAPIDefinitionInBundleRequest(application.ID, bundle.ID, api.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApiForBundle, &app)
	require.NoError(t, err)

	actualApi := app.Bundle.APIDefinition
	assert.Equal(t, api.ID, actualApi.ID)
	saveExample(t, queryApiForBundle.Query(), "query api definition")

}

func TestEventDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	event := addEventToBundle(t, ctx, bundle.ID)

	queryEventForBundle := fixEventDefinitionInBundleRequest(application.ID, bundle.ID, event.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventForBundle, &app)
	require.NoError(t, err)

	actualEvent := app.Bundle.EventDefinition
	assert.Equal(t, event.ID, actualEvent.ID)
	saveExample(t, queryEventForBundle.Query(), "query event definition")

}

func TestDocumentInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	doc := addDocumentToBundle(t, ctx, bundle.ID)

	queryDocForBundle := fixDocumentInBundleRequest(application.ID, bundle.ID, doc.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocForBundle, &app)
	require.NoError(t, err)

	actualDoc := app.Bundle.Document
	assert.Equal(t, doc.ID, actualDoc.ID)
	saveExample(t, queryDocForBundle.Query(), "query document")
}

func TestAPIDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	inputA := fixAPIDefinitionInputWithName("foo")
	addAPIToBundleWithInput(t, ctx, bundle.ID, inputA)

	inputB := fixAPIDefinitionInputWithName("bar")
	addAPIToBundleWithInput(t, ctx, bundle.ID, inputB)

	queryApisForBundle := fixAPIDefinitionsInBundleRequest(application.ID, bundle.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApisForBundle, &app)
	require.NoError(t, err)

	apis := app.Bundle.APIDefinitions
	require.Equal(t, 2, apis.TotalCount)
	assertAPI(t, []*graphql.APIDefinitionInput{&inputA, &inputB}, apis.Data)
	saveExample(t, queryApisForBundle.Query(), "query api definitions")
}

func TestEventDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	inputA := fixEventAPIDefinitionInputWithName("foo")
	addEventToBundleWithInput(t, ctx, bundle.ID, inputA)

	inputB := fixEventAPIDefinitionInputWithName("bar")
	addEventToBundleWithInput(t, ctx, bundle.ID, inputB)

	queryEventsForBundle := fixEventDefinitionsInBundleRequest(application.ID, bundle.ID)

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventsForBundle, &app)
	require.NoError(t, err)

	events := app.Bundle.EventDefinitions
	require.Equal(t, 2, events.TotalCount)
	assertEventsAPI(t, []*graphql.EventDefinitionInput{&inputA, &inputB}, events.Data)
	saveExample(t, queryEventsForBundle.Query(), "query event definitions")
}

func TestDocumentsInBundle(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-bundle"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	bundleName := "test-bundle"
	bundle := createBundle(t, ctx, application.ID, bundleName)
	defer deleteBundle(t, ctx, bundle.ID)

	inputA := fixDocumentInputWithName(t, "foo")
	addDocumentToBundleWithInput(t, ctx, bundle.ID, inputA)

	inputB := fixDocumentInputWithName(t, "bar")
	addDocumentToBundleWithInput(t, ctx, bundle.ID, inputB)

	queryDocsForBundle := fixDocumentsInBundleRequest(application.ID, bundle.ID)

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocsForBundle, &app)
	require.NoError(t, err)

	docs := app.Bundle.Documents
	require.Equal(t, 2, docs.TotalCount)
	assertDocuments(t, []*graphql.DocumentInput{&inputA, &inputB}, docs.Data)
	saveExample(t, queryDocsForBundle.Query(), "query documents")
}

func TestAddBundle(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bundleInput := fixBundleCreateInputWithRelatedObjects(t, "bundle-app-1")
	bundle, err := tc.graphqlizer.BundleCreateInputToGQL(bundleInput)
	require.NoError(t, err)

	addBundleRequest := fixAddBundleRequest(application.ID, bundle)
	output := graphql.BundleExt{}

	// WHEN
	t.Log("Create bundle")
	err = tc.RunOperation(ctx, addBundleRequest, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	assertBundle(t, &bundleInput, &output)
	defer deleteBundle(t, ctx, output.ID)

	saveExample(t, addBundleRequest.Query(), "add bundle")

	bundleRequest := fixBundleRequest(application.ID, output.ID)
	bundleFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, bundleRequest, &bundleFromAPI)
	require.NoError(t, err)

	assertBundle(t, &bundleInput, &output)
	saveExample(t, bundleRequest.Query(), "query bundle")
}

func TestQueryBundles(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bundle1 := createBundle(t, ctx, application.ID, "bundle-app-1")
	defer deleteBundle(t, ctx, bundle1.ID)

	bundle2 := createBundle(t, ctx, application.ID, "bundle-app-2")
	defer deleteBundle(t, ctx, bundle2.ID)

	bundlesRequest := fixBundlesRequest(application.ID)
	bundlesFromAPI := graphql.ApplicationExt{}

	err := tc.RunOperation(ctx, bundlesRequest, &bundlesFromAPI)
	require.NoError(t, err)
	require.Equal(t, 2, len(bundlesFromAPI.Bundles.Data))

	saveExample(t, bundlesRequest.Query(), "query bundles")
}

func TestUpdateBundle(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bundle := createBundle(t, ctx, application.ID, "bundle-app-1")
	defer deleteBundle(t, ctx, bundle.ID)

	bundleUpdateInput := fixBundleUpdateInput("bundle-app-1-up")
	bundleUpdate, err := tc.graphqlizer.BundleUpdateInputToGQL(bundleUpdateInput)
	require.NoError(t, err)

	updateBundleReq := fixUpdateBundleRequest(bundle.ID, bundleUpdate)
	output := graphql.Bundle{}

	// WHEN
	t.Log("Update bundle")
	err = tc.RunOperation(ctx, updateBundleReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, updateBundleReq.Query(), "update bundle")
}

func TestDeleteBundle(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-bundle")
	defer unregisterApplication(t, application.ID)

	bundle := createBundle(t, ctx, application.ID, "bundle-app-1")

	pkdDeleteReq := fixDeleteBundleRequest(bundle.ID)
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
