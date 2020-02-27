package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	queryApiForPkg := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							apiDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, application.ID, pkg.ID, api.ID, tc.gqlFieldsProvider.ForAPIDefinition()))
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApiForPkg, &app)
	require.NoError(t, err)

	actualApi := app.Package.APIDefinition
	assert.Equal(t, api.ID, actualApi.ID)
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

	queryEventForPkg := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							eventDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, application.ID, pkg.ID, event.ID, tc.gqlFieldsProvider.ForEventDefinition()))
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventForPkg, &app)
	require.NoError(t, err)

	actualEvent := app.Package.EventDefinition
	assert.Equal(t, event.ID, actualEvent.ID)
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

	queryDocForPkg := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							document(id: "%s"){
						%s
						}					
					}
				}
			}`, application.ID, pkg.ID, doc.ID, tc.gqlFieldsProvider.ForDocument()))
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocForPkg, &app)
	require.NoError(t, err)

	actualDoc := app.Package.Document
	assert.Equal(t, doc.ID, actualDoc.ID)
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

	queryApisForPkg := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							apiDefinitions{
						%s
						}					
					}
				}
			}`, application.ID, pkg.ID, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForAPIDefinition())))

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApisForPkg, &app)
	require.NoError(t, err)

	apis := app.Package.APIDefinitions
	require.Equal(t, 2, apis.TotalCount)
	assertAPI(t, []*graphql.APIDefinitionInput{&inputA, &inputB}, apis.Data)

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

	queryEventsForPkg := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							eventDefinitions{
						%s
						}					
					}
				}
			}`, application.ID, pkg.ID, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForEventDefinition())))

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryEventsForPkg, &app)
	require.NoError(t, err)

	events := app.Package.EventDefinitions
	require.Equal(t, 2, events.TotalCount)
	assertEventsAPI(t, []*graphql.EventDefinitionInput{&inputA, &inputB}, events.Data)

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

	queryDocsForPkg := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							documents{
						%s
						}					
					}
				}
			}`, application.ID, pkg.ID, tc.gqlFieldsProvider.Page(tc.gqlFieldsProvider.ForDocument())))

	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryDocsForPkg, &app)
	require.NoError(t, err)

	docs := app.Package.Documents
	require.Equal(t, 2, docs.TotalCount)
	assertDocuments(t, []*graphql.DocumentInput{&inputA, &inputB}, docs.Data)

}
