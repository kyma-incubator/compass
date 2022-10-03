package fixtures

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func ListFormationAssignments(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, listFormationAssignmentsReq *gcli.Request, expectedCount int) *graphql.FormationAssignmentPage {
	var formationAssignmentPage graphql.FormationAssignmentPage
	err := testctx.Tc.RunOperation(ctx, gqlClient, listFormationAssignmentsReq, &formationAssignmentPage)
	require.NoError(t, err)
	require.NotEmpty(t, formationAssignmentPage)
	require.Equal(t, expectedCount, formationAssignmentPage.TotalCount)

	return &formationAssignmentPage
}
