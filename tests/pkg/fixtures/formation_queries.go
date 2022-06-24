package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func ListFormations(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, first, expectedCount int) *graphql.FormationPage {
	listFormationsReq := FixListFormationsRequestWithPageSize(first)
	var formationPage graphql.FormationPage
	err := testctx.Tc.RunOperation(ctx, gqlClient, listFormationsReq, &formationPage)
	require.NoError(t, err)
	require.NotEmpty(t, formationPage)
	require.Equal(t, expectedCount, formationPage.TotalCount)

	return &formationPage
}

func CreateFormation(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, formationName string) graphql.Formation {
	var formation graphql.Formation
	createFirstFormationReq := FixCreateFormationRequest(formationName)
	err := testctx.Tc.RunOperation(ctx, gqlClient, createFirstFormationReq, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)

	return formation
}

func DeleteFormation(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, formationName string) *graphql.Formation {
	deleteRequest := FixDeleteFormationRequest(formationName)
	var deleteFormation graphql.Formation
	err := testctx.Tc.RunOperation(ctx, gqlClient, deleteRequest, &deleteFormation)
	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	return &deleteFormation
}

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

func CleanupFormationWithTenantObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID, parent string) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(tenantID, string(graphql.FormationObjectTypeTenant), in.Name)

	formation := graphql.Formation{}

	assertions.AssertNoErrorForOtherThanNotFound(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, parent, unassignRequest, &formation))
	return &formation
}
