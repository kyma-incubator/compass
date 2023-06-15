package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func AddEventToApplicationWithInput(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant, appID string, in graphql.EventDefinitionInput) graphql.EventAPIDefinitionExt {
	inputGQL, err := testctx.Tc.Graphqlizer.EventDefinitionInputToGQL(in)
	require.NoError(t, err)

	request := FixAddEventToApplicationRequest(appID, inputGQL)
	eventDef := graphql.EventAPIDefinitionExt{}

	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, request, &eventDef)
	require.NoError(t, err)
	return eventDef
}

func AddEventToApplication(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, appID string) graphql.EventAPIDefinitionExt {
	return AddEventToApplicationWithInput(t, ctx, gqlClient, tenant.TestTenants.GetDefaultTenantID(), appID, FixEventAPIDefinitionInput())
}
