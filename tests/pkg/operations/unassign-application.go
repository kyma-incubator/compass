package operations

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type UnassignAppToFormationOperation struct {
	applicationID string
	tenantID      string
	asserters     []asserters.Asserter
}

func NewUnassignAppToFormationOperation(applicationID string, tenantID string) *UnassignAppToFormationOperation {
	return &UnassignAppToFormationOperation{applicationID: applicationID, tenantID: tenantID}
}

func (o *UnassignAppToFormationOperation) WithAsserters(asserters ...asserters.Asserter) *UnassignAppToFormationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *UnassignAppToFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationName := ctx.Value(FormationNameKey).(string)
	fixtures.UnassignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *UnassignAppToFormationOperation) Cleanup(_ *testing.T, _ context.Context, _ *gcli.Client) {
	//nothing to defer
}

func (o *UnassignAppToFormationOperation) Operation() Operation {
	return o
}
