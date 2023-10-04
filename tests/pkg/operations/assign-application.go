package operations

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type AssignAppToFormationOperation struct {
	formationName string
	applicationID string
	tenantID      string
	asserters     []asserters.Asserter
}

func NewAssignAppToFormationOperation(formationName string, applicationID string, tenantID string) *AssignAppToFormationOperation {
	return &AssignAppToFormationOperation{formationName: formationName, applicationID: applicationID, tenantID: tenantID}
}

func (o *AssignAppToFormationOperation) WithAsserter(asserter asserters.Asserter) *AssignAppToFormationOperation {
	o.asserters = append(o.asserters, asserter)
	return o
}

func (o *AssignAppToFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: o.formationName}, o.applicationID, o.tenantID)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AssignAppToFormationOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.UnassignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: o.formationName}, o.applicationID, o.tenantID)
}

func (o *AssignAppToFormationOperation) Operation() Operation {
	return o
}
