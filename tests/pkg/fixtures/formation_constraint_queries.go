package fixtures

import (
	"context"

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
	formationConstraint := graphql.FormationConstraint{}
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, createRequest, &formationConstraint))
	require.NotEmpty(t, formationConstraint.ID)

	return &formationConstraint
}

func CleanupFormationConstraint(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, id string) *graphql.FormationConstraint {
	deleteRequest := FixDeleteFormationConstraintRequest(id)

	formationConstraint := graphql.FormationConstraint{}
	err := testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, deleteRequest, &formationConstraint)

	assertions.AssertNoErrorForOtherThanNotFound(t, err)

	return &formationConstraint
}

func ListFormationConstraintsForFormationTemplate(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, formationTemplateID string) []*graphql.FormationConstraint {
	queryRequest := FixQueryFormationConstraintsForFormationTemplateRequest(formationTemplateID)

	var actualFormationConstraints []*graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, queryRequest, &actualFormationConstraints))
	return actualFormationConstraints
}

func UpdateFormationConstraint(t require.TestingT, ctx context.Context, constraintID string, updateInput graphql.FormationConstraintUpdateInput, gqlClient *gcli.Client) *graphql.FormationConstraint {
	formationConstraintGQL, err := testctx.Tc.Graphqlizer.FormationConstraintUpdateInputToGQL(updateInput)
	require.NoError(t, err)

	updateRequest := FixUpdateFormationConstraintRequest(constraintID, formationConstraintGQL)

	var actualFormationConstraint *graphql.FormationConstraint
	require.NoError(t, testctx.Tc.RunOperationWithoutTenant(ctx, gqlClient, updateRequest, &actualFormationConstraint))
	return actualFormationConstraint
}
