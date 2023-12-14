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

func CreateFormationConstraintAndAttach(t *testing.T, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationConstraintInput, formationTemplateID, formationTemplateName string) *graphql.FormationConstraint {
	formationConstraint := CreateFormationConstraint(t, ctx, gqlClient, in)
	require.NotEmpty(t, formationConstraint.ID)
	AttachConstraintToFormationTemplate(t, ctx, gqlClient, formationConstraint.ID, formationConstraint.Name, formationTemplateID, formationTemplateName)
	return formationConstraint
}

func CleanupFormationConstraintAndDetach(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, constraintID, formationTemplateID string) *graphql.FormationConstraint {
	DetachConstraintFromFormationTemplateNoCheckError(ctx, gqlClient, constraintID, formationTemplateID)
	formationConstraint := CleanupFormationConstraint(t, ctx, gqlClient, constraintID)
	require.NotEmpty(t, formationConstraint.ID)
	return formationConstraint
}

func CreateFormationConstraint(t *testing.T, ctx context.Context, gqlClient *gcli.Client, in graphql.FormationConstraintInput) *graphql.FormationConstraint {
	t.Logf("Creating formation constraint with name: %s", in.Name)
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
