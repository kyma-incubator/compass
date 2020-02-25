package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/gql"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAddPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgInput := fixPackageCreateInput("pkg-app-1")
	pkg, err := tc.graphqlizer.PackageCreateInputToGQL(pkgInput)
	require.NoError(t, err)

	addPkgRequest := fixAddPackageRequest(application.ID, pkg)
	output := graphql.Package{}

	// WHEN
	t.Log("Create package")
	err = tc.RunOperation(ctx, addPkgRequest, &output)

	// THEN
	require.NoError(t, err)
	require.NotEmpty(t, output.ID)
	defer deletePackage(t, ctx, output.ID)

	require.NotEmpty(t, output.Name)
	saveExample(t, addPkgRequest.Query(), "create package")

	t.Log("Check if package was created")

	getPackageRequest := fixPackageRequest(application.ID, output.ID)
	pkgOutput := graphql.Package{}
	logrus.Info(fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, "applicationID", tc.gqlFieldsProvider.ForApplication(gql.FieldCtx{
		"Package.package": fmt.Sprintf(`package(id: "%s") {id}`, "packageID"),
	})))
	err = tc.RunOperation(ctx, getPackageRequest, &pkgOutput)

	require.NoError(t, err)
	//require.NotEmpty(t, pkgOutput)
	//assertPackage(t, pkgInput, pkgOutput)
	//assert.Equal(t, pkgInput.Name, pkgOutput.Name)
	//saveExample(t, getPackageRequest.Query(), "query package")
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
