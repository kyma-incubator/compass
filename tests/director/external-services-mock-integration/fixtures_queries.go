package external_services_mock_integration

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/pkg/ptr"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	gcli "github.com/machinebox/graphql"

	"github.com/stretchr/testify/require"
)

//Application
func unregisterApplicationInTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, id string, tenant string) {
	req := fixDeleteApplicationRequest(t, id)
	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil))
}

func registerApplication(t *testing.T, ctx context.Context, gqlClient *gcli.Client, name, tenant string) graphql.ApplicationExt {
	in := fixSampleApplicationRegisterInputWithName("first", name)
	return registerApplicationFromInputWithinTenant(t, ctx, gqlClient, in, tenant)
}

func registerApplicationFromInputWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, in graphql.ApplicationRegisterInput, tenantID string) graphql.ApplicationExt {
	appInputGQL, err := tc.Graphqlizer.ApplicationRegisterInputToGQL(in)
	require.NoError(t, err)

	createRequest := fixRegisterApplicationRequest(appInputGQL)

	app := graphql.ApplicationExt{}

	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, createRequest, &app))
	require.NotEmpty(t, app.ID)
	return app
}

func fixSampleApplicationRegisterInputWithName(placeholder, name string) graphql.ApplicationRegisterInput {
	sampleInput := fixSampleApplicationRegisterInput(placeholder)
	sampleInput.Name = name
	return sampleInput
}

func fixSampleApplicationRegisterInput(placeholder string) graphql.ApplicationRegisterInput {
	return graphql.ApplicationRegisterInput{
		Name:         placeholder,
		ProviderName: ptr.String("compass"),
		Labels:       &graphql.Labels{placeholder: []interface{}{placeholder}},
	}
}

func unregisterApplication(t *testing.T, gqlClient *gcli.Client, id, tenant string) {
	req := fixDeleteApplicationRequest(t, id)
	require.NoError(t, tc.RunOperationWithCustomTenant(context.Background(), gqlClient, tenant, req, nil))
}

func createPackageWithInput(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, appID string, input graphql.PackageCreateInput) graphql.PackageExt {
	in, err := tc.Graphqlizer.PackageCreateInputToGQL(input)
	require.NoError(t, err)

	req := fixAddPackageRequest(appID, in)
	var resp graphql.PackageExt

	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &resp)
	require.NoError(t, err)

	return resp
}

func deletePackage(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	req := fixDeletePackageRequest(id)

	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil))
}

func getPackage(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, appID, pkgID string) graphql.PackageExt {
	req := fixPackageRequest(appID, pkgID)
	pkg := graphql.ApplicationExt{}
	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &pkg))
	return pkg.Package
}
