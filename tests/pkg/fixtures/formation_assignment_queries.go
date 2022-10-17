package fixtures

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type AssignmentState struct {
	Config *string
	State  string
}

func ListFormationAssignments(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, listFormationAssignmentsReq *gcli.Request) *graphql.FormationAssignmentPage {
	var formation graphql.FormationExt
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, listFormationAssignmentsReq, &formation)
	require.NoError(t, err)

	return &formation.FormationAssignments
}
