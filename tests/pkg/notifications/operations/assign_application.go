package operations

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type AssignAppToFormationOperation struct {
	applicationID string
	tenantID      string
	asserters     []asserters.Asserter
}

func NewAssignAppToFormationOperation(applicationID string, tenantID string) *AssignAppToFormationOperation {
	return &AssignAppToFormationOperation{applicationID: applicationID, tenantID: tenantID}
}

func (o *AssignAppToFormationOperation) WithAsserters(asserters ...asserters.Asserter) *AssignAppToFormationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *AssignAppToFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationName := ctx.Value(context_keys.FormationNameKey).(string)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AssignAppToFormationOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationName := ctx.Value(context_keys.FormationNameKey).(string)
	fixtures.UnassignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
}

func (o *AssignAppToFormationOperation) Operation() Operation {
	return o
}
