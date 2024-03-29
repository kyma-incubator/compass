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

type UnassignAppFromFormationOperation struct {
	applicationID string
	tenantID      string
	formationName string // used when the test operates with formation different from the one provided in pre  setup
	asserters     []asserters.Asserter
}

func NewUnassignAppFromFormationOperation(applicationID string, tenantID string) *UnassignAppFromFormationOperation {
	return &UnassignAppFromFormationOperation{applicationID: applicationID, tenantID: tenantID}
}

func (o *UnassignAppFromFormationOperation) WithFormationName(formationName string) *UnassignAppFromFormationOperation {
	o.formationName = formationName
	return o
}

func (o *UnassignAppFromFormationOperation) WithAsserters(asserters ...asserters.Asserter) *UnassignAppFromFormationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *UnassignAppFromFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	var formationName string
	if o.formationName != "" {
		formationName = o.formationName
	} else {
		formationName = ctx.Value(context_keys.FormationNameKey).(string)
	}
	fixtures.UnassignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *UnassignAppFromFormationOperation) Cleanup(_ *testing.T, _ context.Context, _ *gcli.Client) {
	//nothing to defer
}

func (o *UnassignAppFromFormationOperation) Operation() Operation {
	return o
}
