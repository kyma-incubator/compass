package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

func TestAddAPIToPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	input := fixAPIDefinitionInput()
	inStr, err := tc.graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := fixAddAPIToPackageRequest(pkg.ID, inStr)
	err = tc.RunOperation(ctx, req, &actualApi)
	require.NoError(t, err)

	pack := getPackage(t, ctx, application.ID, pkg.ID)
	require.Equal(t, pkg.ID, pack.ID)

	assertAPI(t, []*graphql.APIDefinitionInput{&input}, []*graphql.APIDefinitionExt{&actualApi})
	saveExample(t, req.Query(), "add api definition to package")
}

func TestAddEventDefinitionToPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	input := fixEventAPIDefinitionInput()
	inStr, err := tc.graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualEvent := graphql.EventAPIDefinitionExt{}
	req := fixAddEventAPIToPackageRequest(pkg.ID, inStr)
	err = tc.RunOperation(ctx, req, &actualEvent)
	require.NoError(t, err)

	assertEventsAPI(t, []*graphql.EventDefinitionInput{&input}, []*graphql.EventAPIDefinitionExt{&actualEvent})
	saveExample(t, req.Query(), "add event definition to package")
}

func TestAddDocumentToPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	input := fixDocumentInput()
	inStr, err := tc.graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDocument := graphql.DocumentExt{}
	req := fixAddDocumentToPackageRequest(pkg.ID, inStr)
	err = tc.RunOperation(ctx, req, &actualDocument)
	require.NoError(t, err)

	assertDocuments(t, []*graphql.DocumentInput{&input}, []*graphql.DocumentExt{&actualDocument})
	saveExample(t, req.Query(), "add document to package")
}

func TestAPIDefinitionInPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	api := addAPIToPackage(t, ctx, pkg.ID)

	queryApiForPkg := fixAPIDefinitionInPackageRequest(application.ID, pkg.ID, api.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApiForPkg, &app)
	require.NoError(t, err)

	actualApi := app.Package.APIDefinition
	assert.Equal(t, api.ID, actualApi.ID)
	saveExample(t, queryApiForPkg.Query(), "query api definition")

}

func TestEventDefinitionInPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	event := addEventToPackage(t, ctx, pkg.ID)

	queryEventForPkg := fixEventDefinitionInPackageRequest(application.ID, pkg.ID, event.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventForPkg, &app)
	require.NoError(t, err)

	actualEvent := app.Package.EventDefinition
	assert.Equal(t, event.ID, actualEvent.ID)
	saveExample(t, queryEventForPkg.Query(), "query event definition")

}

func TestDocumentInPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	doc := addDocumentToPackage(t, ctx, pkg.ID)

	queryDocForPkg := fixDocumentInPackageRequest(application.ID, pkg.ID, doc.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocForPkg, &app)
	require.NoError(t, err)

	actualDoc := app.Package.Document
	assert.Equal(t, doc.ID, actualDoc.ID)
	saveExample(t, queryDocForPkg.Query(), "query document")
}

func TestAPIDefinitionsInPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	inputA := fixAPIDefinitionInputWithName("foo")
	addAPIToPackageWithInput(t, ctx, pkg.ID, inputA)

	inputB := fixAPIDefinitionInputWithName("bar")
	addAPIToPackageWithInput(t, ctx, pkg.ID, inputB)

	queryApisForPkg := fixAPIDefinitionsInPackageRequest(application.ID, pkg.ID)
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApisForPkg, &app)
	require.NoError(t, err)

	apis := app.Package.APIDefinitions
	require.Equal(t, 2, apis.TotalCount)
	assertAPI(t, []*graphql.APIDefinitionInput{&inputA, &inputB}, apis.Data)
	saveExample(t, queryApisForPkg.Query(), "query api definitions")
}

func TestEventDefinitionsInPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	inputA := fixEventAPIDefinitionInputWithName("foo")
	addEventToPackageWithInput(t, ctx, pkg.ID, inputA)

	inputB := fixEventAPIDefinitionInputWithName("bar")
	addEventToPackageWithInput(t, ctx, pkg.ID, inputB)

	queryEventsForPkg := fixEventDefinitionsInPackageRequest(application.ID, pkg.ID)

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventsForPkg, &app)
	require.NoError(t, err)

	events := app.Package.EventDefinitions
	require.Equal(t, 2, events.TotalCount)
	assertEventsAPI(t, []*graphql.EventDefinitionInput{&inputA, &inputB}, events.Data)
	saveExample(t, queryEventsForPkg.Query(), "query event definitions")
}

func TestDocumentsInPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	inputA := fixDocumentInputWithName("foo")
	addDocumentToPackageWithInput(t, ctx, pkg.ID, inputA)

	inputB := fixDocumentInputWithName("bar")
	addDocumentToPackageWithInput(t, ctx, pkg.ID, inputB)

	queryDocsForPkg := fixDocumentsInPackageRequest(application.ID, pkg.ID)

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocsForPkg, &app)
	require.NoError(t, err)

	docs := app.Package.Documents
	require.Equal(t, 2, docs.TotalCount)
	assertDocuments(t, []*graphql.DocumentInput{&inputA, &inputB}, docs.Data)
	saveExample(t, queryDocsForPkg.Query(), "query documents")
}

func TestAddPackage(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkgInput := fixPackageCreateInputWithRelatedObjects("pkg-app-1")
	pkg, err := tc.graphqlizer.PackageCreateInputToGQL(pkgInput)
	require.NoError(t, err)

	addPkgRequest := fixAddPackageRequest(application.ID, pkg)
	output := graphql.PackageExt{}

	// WHEN
	t.Log("Create package")
	err = tc.RunOperation(ctx, addPkgRequest, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	assertPackage(t, &pkgInput, &output)
	defer deletePackage(t, ctx, output.ID)

	saveExample(t, addPkgRequest.Query(), "add package")

	packageRequest := fixPackageRequest(application.ID, output.ID)
	pkgFromAPI := graphql.ApplicationExt{}

	err = tc.RunOperation(ctx, packageRequest, &pkgFromAPI)
	require.NoError(t, err)

	assertPackage(t, &pkgInput, &output)
	saveExample(t, packageRequest.Query(), "query package")
}

func TestQueryPackages(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkg1 := createPackage(t, ctx, application.ID, "pkg-app-1")
	defer deletePackage(t, ctx, pkg1.ID)

	pkg2 := createPackage(t, ctx, application.ID, "pkg-app-2")
	defer deletePackage(t, ctx, pkg2.ID)

	packagesRequest := fixPackagesRequest(application.ID)
	pkgsFromAPI := graphql.ApplicationExt{}

	err := tc.RunOperation(ctx, packagesRequest, &pkgsFromAPI)
	require.NoError(t, err)
	require.Equal(t, 2, len(pkgsFromAPI.Packages.Data))

	saveExample(t, packagesRequest.Query(), "query packages")
}

func TestUpdatePackage(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkg := createPackage(t, ctx, application.ID, "pkg-app-1")
	defer deletePackage(t, ctx, pkg.ID)

	pkgUpdateInput := fixPackageUpdateInput("pkg-app-1-up")
	pkgUpdate, err := tc.graphqlizer.PackageUpdateInputToGQL(pkgUpdateInput)
	require.NoError(t, err)

	updatePkgReq := fixUpdatePackageRequest(pkg.ID, pkgUpdate)
	output := graphql.Package{}

	// WHEN
	t.Log("Update package")
	err = tc.RunOperation(ctx, updatePkgReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, updatePkgReq.Query(), "update package")
}

func TestDeletePackage(t *testing.T) {
	ctx := context.Background()

	application := registerApplication(t, ctx, "app-test-package")
	defer unregisterApplication(t, application.ID)

	pkg := createPackage(t, ctx, application.ID, "pkg-app-1")

	pkdDeleteReq := fixDeletePackageRequest(pkg.ID)
	output := graphql.Package{}

	// WHEN
	t.Log("Delete package")
	err := tc.RunOperation(ctx, pkdDeleteReq, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, pkdDeleteReq.Query(), "delete package")
}
