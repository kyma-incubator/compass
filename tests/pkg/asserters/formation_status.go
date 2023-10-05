package asserters

import (
	"context"
	gql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"testing"
)

type FormationStatusAsserter struct {
	formationID              string
	tenant                   string
	certSecuredGraphQLClient *graphql.Client
	condition                gql.FormationStatusCondition
	errors                   []*gql.FormationStatusError
}

func NewFormationStatusAsserter(formationID string, tenant string, certSecuredGraphQLClient *graphql.Client) *FormationStatusAsserter {
	return &FormationStatusAsserter{formationID: formationID, tenant: tenant, certSecuredGraphQLClient: certSecuredGraphQLClient, condition: gql.FormationStatusConditionReady, errors: nil}
}

func (a *FormationStatusAsserter) WithCondition(condition gql.FormationStatusCondition) *FormationStatusAsserter {
	a.condition = condition
	return a
}

func (a *FormationStatusAsserter) WithErrors(errors []*gql.FormationStatusError) *FormationStatusAsserter {
	a.errors = errors
	return a
}

func (a *FormationStatusAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	a.assertFormationStatus(t, ctx, a.tenant, a.formationID, gql.FormationStatus{
		Condition: a.condition,
		Errors:    a.errors,
	})
}

func (a *FormationStatusAsserter) assertFormationStatus(t *testing.T, ctx context.Context, tenant, formationID string, expectedFormationStatus gql.FormationStatus) {
	// Get the formation with its status
	t.Logf("Getting formation with ID: %q", formationID)
	var gotFormation gql.FormationExt
	getFormationReq := fixtures.FixGetFormationRequest(formationID)
	err := testctx.Tc.RunOperationWithCustomTenant(ctx, a.certSecuredGraphQLClient, tenant, getFormationReq, &gotFormation)
	require.NoError(t, err)

	// Assert the status
	require.Equal(t, expectedFormationStatus.Condition, gotFormation.Status.Condition, "Formation with ID %q is with status %q, but %q was expected", formationID, gotFormation.Status.Condition, expectedFormationStatus.Condition)

	if expectedFormationStatus.Errors == nil {
		require.Nil(t, gotFormation.Status.Errors)
	} else { // assert only the Message and ErrorCode
		require.Len(t, gotFormation.Status.Errors, len(expectedFormationStatus.Errors))
		for _, expectedError := range expectedFormationStatus.Errors {
			found := false
			for _, gotError := range gotFormation.Status.Errors {
				if gotError.ErrorCode == expectedError.ErrorCode && gotError.Message == expectedError.Message {
					found = true
					break
				}
			}
			require.Truef(t, found, "Error %q with error code %d was not found", expectedError.Message, expectedError.ErrorCode)
		}
	}
}
