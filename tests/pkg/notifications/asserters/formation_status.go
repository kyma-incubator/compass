package asserters

import (
	"context"
	"testing"

	gql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type FormationStatusAsserter struct {
	tenant                   string
	formationName            string
	formationState           string
	certSecuredGraphQLClient *graphql.Client
	condition                gql.FormationStatusCondition
	errors                   []*gql.FormationStatusError
}

func NewFormationStatusAsserter(tenant string, certSecuredGraphQLClient *graphql.Client) *FormationStatusAsserter {
	return &FormationStatusAsserter{
		tenant:                   tenant,
		certSecuredGraphQLClient: certSecuredGraphQLClient,
		condition:                gql.FormationStatusConditionReady,
		errors:                   nil,
	}
}

func (a *FormationStatusAsserter) WithCondition(condition gql.FormationStatusCondition) *FormationStatusAsserter {
	a.condition = condition
	return a
}

func (a *FormationStatusAsserter) WithErrors(errors []*gql.FormationStatusError) *FormationStatusAsserter {
	a.errors = errors
	return a
}

func (a *FormationStatusAsserter) WithFormationName(formationName string) *FormationStatusAsserter {
	a.formationName = formationName
	return a
}

func (a *FormationStatusAsserter) WithState(state string) *FormationStatusAsserter {
	a.formationState = state
	return a
}

func (a *FormationStatusAsserter) AssertExpectations(t *testing.T, ctx context.Context) {
	formationID := ctx.Value(context_keys.FormationIDKey).(string)
	a.assertFormationStatus(t, ctx, a.tenant, formationID, a.formationName, gql.FormationStatus{
		Condition: a.condition,
		Errors:    a.errors,
	})
}

func (a *FormationStatusAsserter) assertFormationStatus(t *testing.T, ctx context.Context, tenant, formationID, formationName string, expectedFormationStatus gql.FormationStatus) {
	// Get the formation with its status
	var gotFormation *gql.FormationExt

	if formationName != "" {
		gotFormation = fixtures.GetFormationByName(t, ctx, a.certSecuredGraphQLClient, formationName, tenant)
	} else {
		gotFormation = fixtures.GetFormationByID(t, ctx, a.certSecuredGraphQLClient, formationID, tenant)
	}

	// Assert the status
	require.Equal(t, expectedFormationStatus.Condition, gotFormation.Status.Condition, "Formation with ID %q is with status %q, but %q was expected", formationID, gotFormation.Status.Condition, expectedFormationStatus.Condition)

	if a.formationState != "" {
		require.Equal(t, a.formationState, gotFormation.State)
	}

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
	t.Logf("Formation status was successfully asserted for ID: %q", formationID)
}
