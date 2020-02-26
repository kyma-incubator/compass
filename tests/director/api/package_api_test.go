package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
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

func deletePackage(t *testing.T, ctx context.Context, id string) {
	req := gcli.NewRequest(
		fmt.Sprintf(`mutation {
			deletePackage(id: "%s") {
				id
			}}`, id))

	require.NoError(t, tc.RunOperation(context.Background(), req, nil))
}
