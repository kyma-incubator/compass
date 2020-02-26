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
	saveExample(t, req.Query(), "add api to package")
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
	assert.Equal(t, api.ID, app.Package.APIDefinition.ID)
}

func TestEventDefinitionInPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	api := addEventToPackage(t, ctx, pkg.ID)
	queryApiForPkg := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							eventDefinition(id: "%s"){
						%s
						}					
					}
				}
			}`, application.ID, pkg.ID, api.ID, tc.gqlFieldsProvider.ForEventDefinition()))
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApiForPkg, &app)
	require.NoError(t, err)
	assert.Equal(t, api.ID, app.Package.EventDefinition.ID)
}

func TestDocumentInPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	api := addDocumentoPackage(t, ctx, pkg.ID)
	queryApiForPkg := gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						package(id: "%s"){
							document(id: "%s"){
						%s
						}					
					}
				}
			}`, application.ID, pkg.ID, api.ID, tc.gqlFieldsProvider.ForDocument()))
	app := graphql.ApplicationExt{}
	err := tc.RunOperation(ctx, queryApiForPkg, &app)
	require.NoError(t, err)
	assert.Equal(t, api.ID, app.Package.Document.ID)
}

func createPackage(t *testing.T, ctx context.Context, appID, pkgName string) graphql.Package {
	in, err := tc.graphqlizer.PackageCreateInputToGQL(fixPackageCreateInput(pkgName))
	require.NoError(t, err)

	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addPackage(applicationID: "%s", in: %s) {
				id
			}}`, appID, in))

	var resp graphql.Package

	err = tc.RunOperation(ctx, req, &resp)
	require.NoError(t, err)

	return resp
}

func getPackage(t *testing.T, ctx context.Context, appID, pkgID string) graphql.PackageExt {
	req := fixPackageRequest(appID, pkgID)
	pkg := graphql.ApplicationExt{}
	require.NoError(t, tc.RunOperation(ctx, req, &pkg))
	return pkg.Package
}

func deletePackage(t *testing.T, ctx context.Context, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			deletePackage(id: "%s") {
				id
			}}`, id))

	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}

func addAPIToPackage(t *testing.T, ctx context.Context, pkgID string) graphql.APIDefinitionExt {
	input := fixAPIDefinitionInput()
	inStr, err := tc.graphqlizer.APIDefinitionInputToGQL(input)
	require.NoError(t, err)

	actualApi := graphql.APIDefinitionExt{}
	req := fixAddAPIToPackageRequest(pkgID, inStr)
	err = tc.RunOperation(ctx, req, &actualApi)
	require.NoError(t, err)
	return actualApi
}

func addEventToPackage(t *testing.T, ctx context.Context, pkgID string) graphql.EventDefinition {
	input := fixEventAPIDefinitionInput()
	inStr, err := tc.graphqlizer.EventDefinitionInputToGQL(input)
	require.NoError(t, err)

	event := graphql.EventDefinition{}
	req := fixAddEventAPIToPackageRequest(pkgID, inStr)
	err = tc.RunOperation(ctx, req, &event)
	require.NoError(t, err)
	return event
}

func addDocumentoPackage(t *testing.T, ctx context.Context, pkgID string) graphql.DocumentExt {
	input := fixDocumentInput()
	inStr, err := tc.graphqlizer.DocumentInputToGQL(&input)
	require.NoError(t, err)

	actualDoc := graphql.DocumentExt{}
	req := fixAddDocumentToPackageRequest(pkgID, inStr)
	err = tc.RunOperation(ctx, req, &actualDoc)
	require.NoError(t, err)
	return actualDoc
}
