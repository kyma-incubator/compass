package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func AssignFormationWithTenantObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID, parent string) *graphql.Formation {
	createRequest := FixAssignFormationRequest(tenantID, string(graphql.FormationObjectTypeTenant), in.Name)

	formation := graphql.Formation{}

	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, parent, createRequest, &formation))
	require.NotEmpty(t, formation.Name)
	return &formation
}

func UnassignFormationWithTenantObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID, parent string) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(tenantID, string(graphql.FormationObjectTypeTenant), in.Name)

	formation := graphql.Formation{}

	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, parent, unassignRequest, &formation))
	require.NotEmpty(t, formation.Name)
	return &formation
}

func CleanupUnassignFormationWithTenantObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID, parent string) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(tenantID, string(graphql.FormationObjectTypeTenant), in.Name)

	formation := graphql.Formation{}

	assertions.AssertNoErrorForOtherThanNotFound(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, parent, unassignRequest, &formation))
	return &formation
}
