package fixtures

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func CreateFormationConstraint(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationConstraintInput) *graphql.FormationConstraint {
	formationConstraintInputGQLString, err := testctx.Tc.Graphqlizer.FormationConstraintInputToGQL(in)
	require.NoError(t, err)
	createRequest := FixCreateFormationConstraintRequest(formationConstraintInputGQLString)
	fmt.Println(createRequest)
	formationConstraint := graphql.FormationConstraint{}
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, createRequest, &formationConstraint))
	require.NotEmpty(t, formationConstraint.ID)

	return &formationConstraint
}

func QueryFormationConstraint(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.FormationConstraint {
	queryRequest := FixQueryFormationConstraintRequest(id)

	formationConstraint := graphql.FormationConstraint{}
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, queryRequest, &formationConstraint))
	require.NotEmpty(t, formationConstraint.ID)

	return &formationConstraint
}

func QueryFormationConstraintsWithPageSize(t require.TestingT, ctx context.Context, gqlClient *gcli.Client) []*graphql.FormationConstraint {
	queryRequest := FixQueryFormationConstraintsRequest()

	var formationConstraints []*graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, queryRequest, &formationConstraints))
	require.NotEmpty(t, formationConstraints)

	return formationConstraints
}

func CleanupFormationConstraint(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.FormationConstraint {
	deleteRequest := FixDeleteFormationConstraintRequest(id)

	formationConstraint := graphql.FormationConstraint{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, deleteRequest, &formationConstraint)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	return &formationConstraint
}
