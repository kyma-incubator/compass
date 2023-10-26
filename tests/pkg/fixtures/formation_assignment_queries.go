package fixtures

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type AssignmentState struct {
	Config *string
	Value  *string
	Error  *string
	State  string
}

func ListFormationAssignments(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, listFormationAssignmentsReq *gcli.Request) *graphql.FormationAssignmentPage {
	var formation graphql.FormationExt
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, listFormationAssignmentsReq, &formation)
	require.NoError(t, err)

	return &formation.FormationAssignments
}

func GetFormationAssignmentsBySourceAndTarget(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID, formationID, sourceID, targetID string) *graphql.FormationAssignment {
	listFormationAssignmentsRequest := FixListFormationAssignmentRequest(formationID, 200)
	assignmentsPage := ListFormationAssignments(t, ctx, gqlClient, tenantID, listFormationAssignmentsRequest)
	assignments := assignmentsPage.Data

	for _, assignment := range assignments {
		if assignment.Source == sourceID && assignment.Target == targetID {
			return assignment
		}
	}

	return nil
}
