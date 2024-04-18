package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	gcli "github.com/machinebox/graphql"
)

type ResynchronizeFormationOperation struct {
	tenantID      string
	formationName string // used when the test operates with formation different from the one provided in pre  setup
	asserters     []asserters.Asserter
}

func NewResynchronizeFormationOperation() *ResynchronizeFormationOperation {
	return &ResynchronizeFormationOperation{}
}

func (o *ResynchronizeFormationOperation) WithTenantID(tenantID string) *ResynchronizeFormationOperation {
	o.tenantID = tenantID
	return o
}

func (o *ResynchronizeFormationOperation) WithFormationName(formationName string) *ResynchronizeFormationOperation {
	o.formationName = formationName
	return o
}

func (o *ResynchronizeFormationOperation) WithAsserters(asserters ...asserters.Asserter) *ResynchronizeFormationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *ResynchronizeFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
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
	fixtures.ResynchronizeFormation(t, ctx, gqlClient, o.tenantID, formationID, formationName)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *ResynchronizeFormationOperation) Cleanup(_ *testing.T, _ context.Context, _ *gcli.Client) {
}

func (o *ResynchronizeFormationOperation) Operation() Operation {
	return o
}
