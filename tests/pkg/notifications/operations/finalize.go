package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	gcli "github.com/machinebox/graphql"
)

type FinalizeFormationOperation struct {
	tenantID      string
	formationName string // used when the test operates with formation different from the one provided in pre  setup
	asserters     []asserters.Asserter
}

func NewFinalizeFormationOperation() *FinalizeFormationOperation {
	return &FinalizeFormationOperation{}
}

func (o *FinalizeFormationOperation) WithTenantID(tenantID string) *FinalizeFormationOperation {
	o.tenantID = tenantID
	return o
}

func (o *FinalizeFormationOperation) WithFormationName(formationName string) *FinalizeFormationOperation {
	o.formationName = formationName
	return o
}

func (o *FinalizeFormationOperation) WithAsserters(asserters ...asserters.Asserter) *FinalizeFormationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *FinalizeFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	var formationID string
	var formationName string

	if o.formationName != "" {
		formation := fixtures.GetFormationByName(t, ctx, gqlClient, o.formationName, o.tenantID)
		formationID = formation.ID
		formationName = formation.Name
	} else {
		formationID = ctx.Value(context_keys.FormationIDKey).(string)
		formationName = ctx.Value(context_keys.FormationNameKey).(string)
	}
	fixtures.FinalizeFormation(t, ctx, gqlClient, o.tenantID, formationID, formationName)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *FinalizeFormationOperation) Cleanup(_ *testing.T, _ context.Context, _ *gcli.Client) {
}

func (o *FinalizeFormationOperation) Operation() Operation {
	return o
}
