package asserters

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/tests/example"

	gql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type FormationExpectation struct {
	FormationName    string
	AssignmentsCount int
	StatusCondition  gql.FormationStatusCondition
}

type FormationVisibilityAsserter struct {
	certSecuredGraphQLClient *graphql.Client
	formationExpectations    []*FormationExpectation
	participantID            string
	formations               gql.FormationExt
}

func NewFormationVisibilityAsserter(certSecuredGraphQLClient *graphql.Client) *FormationVisibilityAsserter {
	return &FormationVisibilityAsserter{
		certSecuredGraphQLClient: certSecuredGraphQLClient,
	}
}

func (a *FormationVisibilityAsserter) WithFormationExpectations(expectations []*FormationExpectation) *FormationVisibilityAsserter {
	a.formationExpectations = expectations

	return a
}

func (a *FormationVisibilityAsserter) WithParticipantID(participantID string) *FormationVisibilityAsserter {
	a.participantID = participantID

	return a
}

func (a *FormationVisibilityAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	// Get the formations for participant globally
	var gotFormations []*gql.FormationExt
	getFormationsForObjectGlobalReq := fixtures.FixGetFormationsForObjectRequest(a.participantID)
	err := testctx.Tc.RunOperationWithoutTenant(ctx, a.certSecuredGraphQLClient, getFormationsForObjectGlobalReq, &gotFormations)
	require.NoError(t, err)
	example.SaveExample(t, getFormationsForObjectGlobalReq.Query(), "list formations for object global")

	require.Len(t, gotFormations, len(a.formationExpectations))

	formationsByName := make(map[string]*gql.FormationExt, len(gotFormations))
	for _, formation := range gotFormations {
		formationsByName[formation.Name] = formation
	}

	for _, currentFormationExpectations := range a.formationExpectations {
		actualFormation, ok := formationsByName[currentFormationExpectations.FormationName]
		require.True(t, ok)
		require.Equal(t, currentFormationExpectations.StatusCondition, actualFormation.Status.Condition)
		require.Len(t, actualFormation.FormationAssignments.Data, currentFormationExpectations.AssignmentsCount)
	}
}
