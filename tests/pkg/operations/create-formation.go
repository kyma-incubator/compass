package operations

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type CreateFormationOperation struct {
	formationName         string
	formationTemplateName *string
	tenantID              string
	formationID           string
	asserters             []asserters.Asserter
}

func NewCreateFormationOperation(formationName string, tenantID string, formationTemplateName *string) *CreateFormationOperation {
	return &CreateFormationOperation{formationName: formationName, tenantID: tenantID, formationTemplateName: formationTemplateName}
}

func (o *CreateFormationOperation) WithAsserter(asserter asserters.Asserter) *CreateFormationOperation {
	o.asserters = append(o.asserters, asserter)
	return o
}

func (o *CreateFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, gqlClient, o.tenantID, o.formationName, o.formationTemplateName)
	o.formationID = formation.ID
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *CreateFormationOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.DeleteFormationWithinTenant(t, ctx, gqlClient, o.tenantID, o.formationName)
}

func (o *CreateFormationOperation) GetFormationID() string {
	return o.formationID
}

func (o *CreateFormationOperation) Operation() Operation {
	return o
}
