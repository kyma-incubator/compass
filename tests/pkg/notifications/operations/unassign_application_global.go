package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/director/tests/example"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	gcli "github.com/machinebox/graphql"
)

type UnassignAppFromFormationOperationGlobal struct {
	applicationID         string
	formationIDContextKey string
	asserters             []asserters.Asserter
}

func NewUnassignAppToFormationOperationGlobal(applicationID string) *UnassignAppFromFormationOperationGlobal {
	return &UnassignAppFromFormationOperationGlobal{applicationID: applicationID, formationIDContextKey: context_keys.FormationIDKey}
}

func (o *UnassignAppFromFormationOperationGlobal) WithAsserters(asserters ...asserters.Asserter) *UnassignAppFromFormationOperationGlobal {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *UnassignAppFromFormationOperationGlobal) WithFormationIDContextKey(formationIDContextKey string) *UnassignAppFromFormationOperationGlobal {
	o.formationIDContextKey = formationIDContextKey
	return o
}

func (o *UnassignAppFromFormationOperationGlobal) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationID := ctx.Value(o.formationIDContextKey).(string)
	query, _ := fixtures.UnassignFormationApplicationGlobal(t, ctx, gqlClient, o.applicationID, formationID)
	example.SaveExampleInCustomDir(t, query, example.UnassignFormationGlobalCategory, "unassign application from formation global")
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *UnassignAppFromFormationOperationGlobal) Cleanup(_ *testing.T, _ context.Context, _ *gcli.Client) {
	//nothing to defer
}

func (o *UnassignAppFromFormationOperationGlobal) Operation() Operation {
	return o
}
