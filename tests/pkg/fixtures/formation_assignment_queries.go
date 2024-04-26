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

type Operation struct {
	SourceID    string
	TargetID    string
	Type        graphql.AssignmentOperationType
	TriggeredBy graphql.OperationTrigger
	IsFinished  bool
}

func NewOperation(sourceID, targetID string, operationType graphql.AssignmentOperationType, triggeredBy graphql.OperationTrigger, isFinished bool) *Operation {
	return &Operation{
		SourceID:    sourceID,
		TargetID:    targetID,
		Type:        operationType,
		TriggeredBy: triggeredBy,
		IsFinished:  isFinished,
	}
}

type Assignment struct {
	AssignmentStatus AssignmentState
	Operations       []*Operation
}

func ListFormationAssignments(t require.TestingT, ctx context.Context, gqlClient *gcli.Client, tenant string, listFormationAssignmentsReq *gcli.Request) *graphql.FormationAssignmentPageExt {
	var formation graphql.FormationExt
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, gqlClient, tenant, listFormationAssignmentsReq, &formation)
	require.NoError(t, err)

	return &formation.FormationAssignments
}

func GetFormationAssignmentsBySourceAndTarget(t *testing.T, ctx context.Context, gqlClient *gcli.Client, tenantID, formationID, sourceID, targetID string) *graphql.FormationAssignmentExt {
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
