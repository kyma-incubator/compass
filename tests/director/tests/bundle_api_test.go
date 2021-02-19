package tests

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestAddAPIToBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	input := pkg.FixAPIDefinitionInput()
	inStr, err := pkg.Tc.Graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := pkg.FixAddAPIToBundleRequest(bndl.ID, inStr)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, req, &actualApi)
	require.NoError(t, err)

	pack := pkg.GetBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndl.ID)
	require.Equal(t, bndl.ID, pack.ID)

	assertAPI(t, []*graphql.APIDefinitionInput{&input}, []*graphql.APIDefinitionExt{&actualApi})
	saveExample(t, req.Query(), "add api definition to bundle")
}

func TestManageAPIInBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	api := pkg.AddAPIToBundle(t, ctx, dexGraphQLClient, bndl.ID)

	apiUpdateInput := pkg.FixAPIDefinitionInputWithName("new-name")
	apiUpdateGQL, err := pkg.Tc.Graphqlizer.APIDefinitionInputToGQL(apiUpdateInput)
	require.NoError(t, err)

	req := pkg.FixUpdateAPIRequest(api.ID, apiUpdateGQL)

	var updatedAPI graphql.APIDefinitionExt
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, req, &updatedAPI)
	require.NoError(t, err)

	assert.Equal(t, updatedAPI.ID, api.ID)
	assert.Equal(t, updatedAPI.Name, "new-name")
	saveExample(t, req.Query(), "update api definition")

	var deletedAPI graphql.APIDefinitionExt
	req = pkg.FixDeleteAPIRequest(api.ID)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, req, &deletedAPI)
	require.NoError(t, err)

	assert.Equal(t, api.ID, deletedAPI.ID)
	saveExample(t, req.Query(), "delete api definition")
}

func TestAddEventDefinitionToBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	input := pkg.FixEventAPIDefinitionInput()
	inStr, err := pkg.Tc.Graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualEvent := graphql.EventAPIDefinitionExt{}
	req := pkg.FixAddEventAPIToBundleRequest(bndl.ID, inStr)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, req, &actualEvent)
	require.NoError(t, err)

	assertEventsAPI(t, []*graphql.EventDefinitionInput{&input}, []*graphql.EventAPIDefinitionExt{&actualEvent})
	saveExample(t, req.Query(), "add event definition to bundle")
}

func TestManageEventDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	event := pkg.AddEventToBundle(t, ctx, dexGraphQLClient, bndl.ID)

	eventUpdateInput := pkg.FixEventAPIDefinitionInputWithName("new-name")
	eventUpdateGQL, err := pkg.Tc.Graphqlizer.EventDefinitionInputToGQL(eventUpdateInput)
	require.NoError(t, err)

	req := pkg.FixUpdateEventAPIRequest(event.ID, eventUpdateGQL)

	var updatedEvent graphql.EventAPIDefinitionExt
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, req, &updatedEvent)
	require.NoError(t, err)

	assert.Equal(t, updatedEvent.ID, event.ID)
	assert.Equal(t, updatedEvent.Name, "new-name")
	saveExample(t, req.Query(), "update event definition")

	var deletedEvent graphql.EventAPIDefinitionExt
	req = pkg.FixDeleteEventAPIRequest(event.ID)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, req, &deletedEvent)
	require.NoError(t, err)

	assert.Equal(t, event.ID, deletedEvent.ID)
	saveExample(t, req.Query(), "delete event definition")
}

func TestAddDocumentToBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	input := pkg.FixDocumentInput(t)
	inStr, err := pkg.Tc.Graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDocument := graphql.DocumentExt{}
	req := pkg.FixAddDocumentToBundleRequest(bndl.ID, inStr)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, req, &actualDocument)
	require.NoError(t, err)

	assertDocuments(t, []*graphql.DocumentInput{&input}, []*graphql.DocumentExt{&actualDocument})
	saveExample(t, req.Query(), "add document to bundle")
}

func TestManageDocumentInBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	document := pkg.AddDocumentToBundle(t, ctx, dexGraphQLClient, bndl.ID)

	var deletedDocument graphql.DocumentExt
	req := pkg.FixDeleteDocumentRequest(document.ID)
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, req, &deletedDocument)
	require.NoError(t, err)

	assert.Equal(t, document.ID, deletedDocument.ID)
	saveExample(t, req.Query(), "delete document")
}

func TestAPIDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	api := pkg.AddAPIToBundle(t, ctx, dexGraphQLClient, bndl.ID)

	queryApiForBndl := pkg.FixAPIDefinitionInBundleRequest(application.ID, bndl.ID, api.ID)
	app := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, queryApiForBndl, &app)
	require.NoError(t, err)

	actualApi := app.Bundle.APIDefinition
	assert.Equal(t, api.ID, actualApi.ID)
	saveExample(t, queryApiForBndl.Query(), "query api definition")

}

func TestEventDefinitionInBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	event := pkg.AddEventToBundle(t, ctx, dexGraphQLClient, bndl.ID)

	queryEventForBndl := pkg.FixEventDefinitionInBundleRequest(application.ID, bndl.ID, event.ID)
	app := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, queryEventForBndl, &app)
	require.NoError(t, err)

	actualEvent := app.Bundle.EventDefinition
	assert.Equal(t, event.ID, actualEvent.ID)
	saveExample(t, queryEventForBndl.Query(), "query event definition")

}

func TestDocumentInBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	doc := pkg.AddDocumentToBundle(t, ctx, dexGraphQLClient, bndl.ID)

	queryDocForBndl := pkg.FixDocumentInBundleRequest(application.ID, bndl.ID, doc.ID)
	app := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, queryDocForBndl, &app)
	require.NoError(t, err)

	actualDoc := app.Bundle.Document
	assert.Equal(t, doc.ID, actualDoc.ID)
	saveExample(t, queryDocForBndl.Query(), "query document")
}

func TestAPIDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	inputA := pkg.FixAPIDefinitionInputWithName("foo")
	pkg.AddAPIToBundleWithInput(t, ctx, dexGraphQLClient, tenant, bndl.ID, inputA)

	inputB := pkg.FixAPIDefinitionInputWithName("bar")
	pkg.AddAPIToBundleWithInput(t, ctx, dexGraphQLClient, tenant, bndl.ID, inputB)

	queryApisForBndl := pkg.FixAPIDefinitionsInBundleRequest(application.ID, bndl.ID)
	app := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, queryApisForBndl, &app)
	require.NoError(t, err)

	apis := app.Bundle.APIDefinitions
	require.Equal(t, 2, apis.TotalCount)
	assertAPI(t, []*graphql.APIDefinitionInput{&inputA, &inputB}, apis.Data)
	saveExample(t, queryApisForBndl.Query(), "query api definitions")
}

func TestEventDefinitionsInBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	inputA := pkg.FixEventAPIDefinitionInputWithName("foo")
	pkg.AddEventToBundleWithInput(t, ctx, dexGraphQLClient, bndl.ID, inputA)

	inputB := pkg.FixEventAPIDefinitionInputWithName("bar")
	pkg.AddEventToBundleWithInput(t, ctx, dexGraphQLClient, bndl.ID, inputB)

	queryEventsForBndl := pkg.FixEventDefinitionsInBundleRequest(application.ID, bndl.ID)

	app := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, queryEventsForBndl, &app)
	require.NoError(t, err)

	events := app.Bundle.EventDefinitions
	require.Equal(t, 2, events.TotalCount)
	assertEventsAPI(t, []*graphql.EventDefinitionInput{&inputA, &inputB}, events.Data)
	saveExample(t, queryEventsForBndl.Query(), "query event definitions")
}

func TestDocumentsInBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	appName := "app-test-bundle"
	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, appName, tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlName := "test-bundle"
	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, bndlName)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	inputA := pkg.FixDocumentInputWithName(t, "foo")
	pkg.AddDocumentToBundleWithInput(t, ctx, dexGraphQLClient, bndl.ID, inputA)

	inputB := pkg.FixDocumentInputWithName(t, "bar")
	pkg.AddDocumentToBundleWithInput(t, ctx, dexGraphQLClient, bndl.ID, inputB)

	queryDocsForBndl := pkg.FixDocumentsInBundleRequest(application.ID, bndl.ID)

	app := graphql.ApplicationExt{}
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, queryDocsForBndl, &app)
	require.NoError(t, err)

	docs := app.Bundle.Documents
	require.Equal(t, 2, docs.TotalCount)
	assertDocuments(t, []*graphql.DocumentInput{&inputA, &inputB}, docs.Data)
	saveExample(t, queryDocsForBndl.Query(), "query documents")
}

func TestAddBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndlInput := pkg.FixBundleCreateInputWithRelatedObjects(t, "bndl-app-1")
	bndl, err := pkg.Tc.Graphqlizer.BundleCreateInputToGQL(bndlInput)
	require.NoError(t, err)

	addBndlRequest := pkg.FixAddBundleRequest(application.ID, bndl)
	output := graphql.BundleExt{}

	// WHEN
	t.Log("Create bundle")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, addBndlRequest, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	assertBundle(t, &bndlInput, &output)
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, output.ID)

	saveExample(t, addBndlRequest.Query(), "add bundle")

	bundleRequest := pkg.FixBundleRequest(application.ID, output.ID)
	bndlFromAPI := graphql.ApplicationExt{}

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, bundleRequest, &bndlFromAPI)
	require.NoError(t, err)

	assertBundle(t, &bndlInput, &output)
	saveExample(t, bundleRequest.Query(), "query bundle")
}

func TestQueryBundles(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl1 := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl1.ID)

	bndl2 := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-2")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl2.ID)

	bundlesRequest := pkg.FixGetBundlesRequest(application.ID)
	bndlsFromAPI := graphql.ApplicationExt{}

	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, bundlesRequest, &bndlsFromAPI)
	require.NoError(t, err)
	require.Equal(t, 2, len(bndlsFromAPI.Bundles.Data))

	saveExample(t, bundlesRequest.Query(), "query bundles")
}

func TestUpdateBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")
	defer pkg.DeleteBundle(t, ctx, dexGraphQLClient, tenant, bndl.ID)

	bndlUpdateInput := pkg.FixBundleUpdateInput("bndl-app-1-up")
	bndlUpdate, err := pkg.Tc.Graphqlizer.BundleUpdateInputToGQL(bndlUpdateInput)
	require.NoError(t, err)

	updateBndlReq := pkg.FixUpdateBundleRequest(bndl.ID, bndlUpdate)
	output := graphql.Bundle{}

	// WHEN
	t.Log("Update bundle")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, updateBndlReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, updateBndlReq.Query(), "update bundle")
}

func TestDeleteBundle(t *testing.T) {
	ctx := context.Background()

	t.Log("Get Dex id_token")
	dexToken, err := idtokenprovider.GetDexToken()
	require.NoError(t, err)

	dexGraphQLClient := gql.NewAuthorizedGraphQLClient(dexToken)

	tenant := pkg.TestTenants.GetDefaultTenantID()

	application := pkg.RegisterApplication(t, ctx, dexGraphQLClient, "app-test-bundle", tenant)
	defer pkg.UnregisterApplication(t, ctx, dexGraphQLClient, tenant, application.ID)

	bndl := pkg.CreateBundle(t, ctx, dexGraphQLClient, tenant, application.ID, "bndl-app-1")

	pkdDeleteReq := pkg.FixDeleteBundleRequest(bndl.ID)
	output := graphql.Bundle{}

	// WHEN
	t.Log("Delete bundle")
	err = pkg.Tc.RunOperation(ctx, dexGraphQLClient, pkdDeleteReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, pkdDeleteReq.Query(), "delete bundle")
}
