package operations

import (
	"context"
	"testing"

	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
)

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
	deletedFormation := fixtures.DeleteFormationWithinTenant(t, ctx, gqlClient, o.tenantID, o.name)

	// For validating formation deletion a formation provisioned in the test case is used.
	// As the formation is created from Operation and not from Provider its ID can not be stored in the context
	// as the operations do not return result. For most operations and asserters providing the formation name as
	// configuration is sufficient as the formation is fetched using its name. However, when asserting the
	// formation deletion by the time the asserter is executed the formation may already be gone and an error will
	// be returned when trying to fetch it. Thus, overriding the context with the required data for the deleted
	// formation is needed.
	//
	// NOTE!!!
	//
	// Do NOT provide formationName as configuration for the asserters for DeleteFormationOperation as it has
	// higher precedence than the contents of the context!!!
	ctx = context.WithValue(ctx, context_keys.FormationIDKey, deletedFormation.ID)
	ctx = context.WithValue(ctx, context_keys.FormationNameKey, deletedFormation.Name)

	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *DeleteFormationOperation) Cleanup(_ *testing.T, _ context.Context, _ *gcli.Client) {
}

func (o *DeleteFormationOperation) Operation() Operation {
	return o
}
