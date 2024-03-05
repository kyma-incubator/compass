package operations

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type AssignAppToFormationOperation struct {
	applicationID           string
	tenantID                string
	formationNameContextKey string
	asserters               []asserters.Asserter
}

func NewAssignAppToFormationOperation(applicationID string, tenantID string) *AssignAppToFormationOperation {
	return &AssignAppToFormationOperation{applicationID: applicationID, tenantID: tenantID, formationNameContextKey: context_keys.FormationNameKey}
}

func (o *AssignAppToFormationOperation) WithFormationNameContextKey(formationNAmeContextKey string) *AssignAppToFormationOperation {
	o.formationNameContextKey = formationNAmeContextKey
	return o
}

func (o *AssignAppToFormationOperation) WithAsserters(asserters ...asserters.Asserter) *AssignAppToFormationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *AssignAppToFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationName := ctx.Value(o.formationNameContextKey).(string)
	fixtures.AssignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AssignAppToFormationOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationName := ctx.Value(o.formationNameContextKey).(string)
	fixtures.UnassignFormationWithApplicationObjectType(t, ctx, gqlClient, graphql.FormationInput{Name: formationName}, o.applicationID, o.tenantID)
}

func (o *AssignAppToFormationOperation) Operation() Operation {
	return o
}
