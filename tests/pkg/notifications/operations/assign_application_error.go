package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	gcli "github.com/machinebox/graphql"
)

type AssignAppToFormationErrorOperation struct {
	applicationID string
	tenantID      string
	asserters     []asserters.Asserter
}

func NewAssignAppToFormationErrorOperation(applicationID string, tenantID string) *AssignAppToFormationErrorOperation {
	return &AssignAppToFormationErrorOperation{applicationID: applicationID, tenantID: tenantID}
}

func (o *AssignAppToFormationErrorOperation) WithAsserters(asserters ...asserters.Asserter) *AssignAppToFormationErrorOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *AssignAppToFormationErrorOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationName := ctx.Value(context_keys.FormationNameKey).(string)
	fixtures.AssignFormationWithApplicationObjectTypeExpectError(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AssignAppToFormationErrorOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationName := ctx.Value(context_keys.FormationNameKey).(string)
	fixtures.UnassignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
}

func (o *AssignAppToFormationErrorOperation) Operation() Operation {
	return o
}
