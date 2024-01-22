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

func ListFormations(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, listFormationsReq *gcli.Request) *graphql.FormationPage {
	var formationPage graphql.FormationPage
	err := testctx.Tc.RunOperation(ctx, gqlClient, listFormationsReq, &formationPage)
	require.NoError(t, err)
	require.NotEmpty(t, formationPage)

	return &formationPage
}

func ListFormationsWithinTenant(t require.TestingT, ctx context.Context, tenantID string, gqlClient *gcli.Client) *graphql.FormationPage {
	first := 100
	listFormationsReq := FixListFormationsRequestWithPageSize(first)

	var formationPage graphql.FormationPage
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, listFormationsReq, &formationPage)
	require.NoError(t, err)
	require.NotEmpty(t, formationPage)

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

func CreateFormationWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantId, formationName string) graphql.Formation {
	t.Logf("Creating formation with name: %q", formationName)
	var formation graphql.Formation
	createFormationReq := FixCreateFormationRequest(formationName)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantId, createFormationReq, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)
	t.Logf("Formation with name: %q is successfully created", formationName)

	return formation
}

func CreateFormationFromTemplateWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID, formationName string, formationTemplateName *string) graphql.FormationExt {
	t.Logf("Creating formation with name: %q from template with name: %q", formationName, *formationTemplateName)
	formationInput := FixFormationInput(formationName, formationTemplateName)
	formationInputGQL, err := testctx.Tc.Graphqlizer.FormationInputToGQL(formationInput)
	require.NoError(t, err)

	var formation graphql.FormationExt
	createFormationReq := FixCreateFormationWithTemplateRequest(formationInputGQL)
	err = testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, createFormationReq, &formation)
	require.NoError(t, err)
	require.Equal(t, formationName, formation.Name)

	return formation
}

func DeleteFormation(t *testing.T, ctx context.Context, gqlClient *gcli.Client, formationName string) *graphql.Formation {
	t.Logf("Deleting formation with name: %q", formationName)
	deleteRequest := FixDeleteFormationRequest(formationName)
	var deleteFormation graphql.Formation
	err := testctx.Tc.RunOperation(ctx, gqlClient, deleteRequest, &deleteFormation)
	assertions.AssertNoErrorForOtherThanNotFound(t, err)
	t.Logf("Deleted formation with name: %s", deleteFormation.Name)

	return &deleteFormation
}

func DeleteFormationWithinTenant(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID, formationName string) *graphql.FormationExt {
	t.Logf("Deleting formation with name: %q", formationName)
	deleteRequest := FixDeleteFormationRequest(formationName)
	var deleteFormation graphql.FormationExt
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, deleteRequest, &deleteFormation)
	assertions.AssertNoErrorForOtherThanNotFound(t, err)
	t.Logf("Deleted formation with name: %q", formationName)

	return &deleteFormation
}

func DeleteFormationWithinTenantExpectError(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID, formationName string) {
	t.Logf("Expect error while deleting formation with name: %q", formationName)
	deleteRequest := FixDeleteFormationRequest(formationName)
	var deleteFormation graphql.Formation
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, deleteRequest, &deleteFormation)
	require.Error(t, err)
}

func AssignFormationWithTenantObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID, parent string) *graphql.Formation {
	return assignFormationWithCustomObjectType(t, ctx, gqlClient, in, tenantID, string(graphql.FormationObjectTypeTenant), parent)
}

func AssignFormationWithTenantObjectTypeExpectError(t *testing.T, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID, parent string) *graphql.Formation {
	return assignFormationWithCustomObjectTypeExpectError(t, ctx, gqlClient, in, tenantID, string(graphql.FormationObjectTypeTenant), parent)
}

func UnassignFormationWithTenantObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID, parent string) *graphql.Formation {
	return unassignFormationWithCustomObjectType(t, ctx, gqlClient, in, tenantID, string(graphql.FormationObjectTypeTenant), parent)
}

func CleanupFormationWithTenantObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, name string, tenantID, parent string) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(tenantID, string(graphql.FormationObjectTypeTenant), name)

	formation := graphql.Formation{}

	assertions.AssertNoErrorForOtherThanNotFound(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, parent, unassignRequest, &formation))
	return &formation
}

func DeleteFormationWithTenantObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, name string, tenantID, parent string) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(tenantID, string(graphql.FormationObjectTypeTenant), name)

	formation := graphql.Formation{}

	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, parent, unassignRequest, &formation))
	return &formation
}

func UnassignFormationApplicationGlobal(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, applicationID, formationID string) *graphql.Formation {
	unassignRequest := FixUnassignFormationApplicationGlobalRequest(applicationID, formationID)

	formation := graphql.Formation{}

	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, unassignRequest, &formation)
	require.NoError(t, err)
	return &formation
}

func AssignFormationWithApplicationObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, appID, tenantID string) *graphql.Formation {
	return assignFormationWithCustomObjectType(t, ctx, gqlClient, in, appID, string(graphql.FormationObjectTypeApplication), tenantID)
}

func AssignFormationWithApplicationObjectTypeExpectError(t *testing.T, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, appID, tenantID string) *graphql.Formation {
	return assignFormationWithCustomObjectTypeExpectError(t, ctx, gqlClient, in, appID, string(graphql.FormationObjectTypeApplication), tenantID)
}

func UnassignFormationWithApplicationObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, appID, tenantID string) *graphql.Formation {
	return unassignFormationWithCustomObjectType(t, ctx, gqlClient, in, appID, string(graphql.FormationObjectTypeApplication), tenantID)
}

func AssignFormationWithRuntimeObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, runtimeID, tenantID string) *graphql.Formation {
	return assignFormationWithCustomObjectType(t, ctx, gqlClient, in, runtimeID, string(graphql.FormationObjectTypeRuntime), tenantID)
}

func AssignFormationWithRuntimeObjectTypeExpectError(t *testing.T, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, runtimeID, tenantID string) *graphql.Formation {
	return assignFormationWithCustomObjectTypeExpectError(t, ctx, gqlClient, in, runtimeID, string(graphql.FormationObjectTypeRuntime), tenantID)
}

func UnassignFormationWithRuntimeObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, runtimeID, tenantID string) *graphql.Formation {
	return unassignFormationWithCustomObjectType(t, ctx, gqlClient, in, runtimeID, string(graphql.FormationObjectTypeRuntime), tenantID)
}

func AssignFormationWithRuntimeContextObjectTypeExpectError(t *testing.T, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, runtimeContextID, tenantID string) *graphql.Formation {
	return assignFormationWithCustomObjectTypeExpectError(t, ctx, gqlClient, in, runtimeContextID, string(graphql.FormationObjectTypeRuntimeContext), tenantID)
}

func UnassignFormationWithRuntimeContextObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, runtimeContextID, tenantID string) *graphql.Formation {
	return unassignFormationWithCustomObjectType(t, ctx, gqlClient, in, runtimeContextID, string(graphql.FormationObjectTypeRuntimeContext), tenantID)
}

func assignFormationWithCustomObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, objectID, objectType, tenantID string) *graphql.Formation {
	createRequest := FixAssignFormationRequest(objectID, objectType, in.Name)

	formation := graphql.Formation{}

	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, createRequest, &formation))
	require.NotEmpty(t, formation.Name)
	return &formation
}

func assignFormationWithCustomObjectTypeExpectError(t *testing.T, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, objectID, objectType, tenantID string) *graphql.Formation {
	createRequest := FixAssignFormationRequest(objectID, objectType, in.Name)

	formation := graphql.Formation{}

	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, createRequest, &formation)
	require.Error(t, err)
	t.Logf("Error: %s", err.Error())
	return &formation
}

func unassignFormationWithCustomObjectType(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, objectID, objectType, tenantID string) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(objectID, objectType, in.Name)

	formation := graphql.Formation{}

	assertions.AssertNoErrorForOtherThanNotFound(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, unassignRequest, &formation))
	return &formation
}

func CleanupFormation(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, objectID string, objectType graphql.FormationObjectType, parent string) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(objectID, string(objectType), in.Name)

	formation := graphql.Formation{}

	assertions.AssertNoErrorForOtherThanNotFound(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, parent, unassignRequest, &formation))
	return &formation
}

func UnassignFormation(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationInput, tenantID, objectID string, objectType graphql.FormationObjectType) *graphql.Formation {
	unassignRequest := FixUnassignFormationRequest(objectID, string(objectType), in.Name)

	formation := graphql.Formation{}

	require.NoError(t, testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, unassignRequest, &formation))
	require.NotEmpty(t, formation.Name)
	return &formation
}

func ResynchronizeFormation(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenantID, formationID, formationName string) *graphql.Formation {
	resynchronizeReq := FixResynchronizeFormationNotificationsRequest(formationID)
	assignedFormation := &graphql.Formation{}
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenantID, resynchronizeReq, &assignedFormation)
	require.NoError(t, err)
	require.Equal(t, formationName, assignedFormation.Name)
	return assignedFormation
}
