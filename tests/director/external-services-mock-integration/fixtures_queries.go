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

func createBundleWithInput(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, appID string, input graphql.BundleCreateInput) graphql.BundleExt {
	in, err := tc.Graphqlizer.BundleCreateInputToGQL(input)
	require.NoError(t, err)

	req := fixAddBundleRequest(appID, in)
	var resp graphql.BundleExt

	err = tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &resp)
	require.NoError(t, err)

	return resp
}

func deleteBundle(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, id string) {
	req := fixDeleteBundleRequest(id)

	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, nil))
}

func getBundle(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenant, appID, bndlID string) graphql.BundleExt {
	req := fixBundleRequest(appID, bndlID)
	bndl := graphql.ApplicationExt{}
	require.NoError(t, tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, req, &bndl))
	return bndl.Bundle
}
