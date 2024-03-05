package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
)
// todo~~ in case of async lifecycle the returned state is deleting - can it be validated - we store only the ID and asserting things should be in asserter
type DeleteFormationOperation struct {
	name                  string
	formationTemplateName string
	tenantID              string
	id                    string
	asserters             []asserters.Asserter
}

func NewDeleteFormationOperation(tenantID string) *DeleteFormationOperation {
	return &DeleteFormationOperation{tenantID: tenantID}
}

func (o *DeleteFormationOperation) WithName(name string) *DeleteFormationOperation {
	o.name = name
	return o
}

func (o *DeleteFormationOperation) WithFormationTemplateName(formationTemplateName string) *DeleteFormationOperation {
	o.formationTemplateName = formationTemplateName
	return o
}

func (o *DeleteFormationOperation) WithAsserters(asserters ...asserters.Asserter) *DeleteFormationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *DeleteFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.DeleteFormationWithinTenant(t, ctx, gqlClient, o.tenantID, o.name)

	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *DeleteFormationOperation) Cleanup(_ *testing.T, _ context.Context, _ *gcli.Client) {
}

func (o *DeleteFormationOperation) Operation() Operation {
	return o
}
