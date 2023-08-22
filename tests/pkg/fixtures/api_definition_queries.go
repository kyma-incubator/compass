package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func AddAPIToApplicationWithInput(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, appID string, in graphql.APIDefinitionInput) graphql.APIDefinitionExt {
	inputGQL, err := testctx.Tc.Graphqlizer.APIDefinitionInputToGQL(in)
	require.NoError(t, err)

	request := FixAddAPIToApplicationRequest(appID, inputGQL)
	apiDef := graphql.APIDefinitionExt{}

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, request, &apiDef)
	require.NoError(t, err)
	return apiDef
}

func AddAPIToApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, appID string) graphql.APIDefinitionExt {
	return AddAPIToApplicationWithInput(t, ctx, gqlClient, tenant.TestTenants.GetDefaultTenantID(), appID, FixAPIDefinitionInput())
}
