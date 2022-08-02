package fixtures

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func ListFormations(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, listFormationsReq *gcli.Request, expectedCount int) *graphql.FormationPage {
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

func CreateFormationWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID, formationName string, formationTemplateName *string) graphql.Formation {
	t.Logf("Creating formation with name: %q from template with name: %q", formationName, *formationTemplateName)
	formationInput := FixFormationInput(formationName, formationTemplateName)
	formationInputGQL, err := testctx.Tc.Graphqlizer.FormationInputToGQL(formationInput)
	require.NoError(t, err)

	var formation graphql.Formation
	createFormationReq := FixCreateFormationWithTemplateRequest(formationInputGQL)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, createFormationReq, &formation)
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

func DeleteFormationWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID, formationName string) *graphql.Formation {
	t.Logf("Deleting formation with name: %q", formationName)
	deleteRequest := FixDeleteFormationRequest(formationName)
	var deleteFormation graphql.Formation
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, deleteRequest, &deleteFormation)
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

func CleanupFormation(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, objectID string, objectType graphql.FormationObjectType, parent string) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(objectID, string(objectType), in.Name)

	formation := graphql.Formation{}

	assertions.AssertNoErrorForOtherThanNotFound(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, parent, unassignRequest, &formation))
	return &formation
}

func AssignFormation(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID string, objectType graphql.FormationObjectType) *graphql.Formation {
	createRequest := FixAssignFormationRequest(tenantID, string(objectType), in.Name)

	formation := graphql.Formation{}

	require.NoError(t, testctx.Tc.RunOperation(ctx, gqlClient, createRequest, &formation))
	require.NotEmpty(t, formation.Name)
	return &formation
}

func UnassignFormation(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID string, objectType graphql.FormationObjectType) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(tenantID, string(objectType), in.Name)

	formation := graphql.Formation{}

	require.NoError(t, testctx.Tc.RunOperation(ctx, gqlClient, unassignRequest, &formation))
	require.NotEmpty(t, formation.Name)
	return &formation
}
