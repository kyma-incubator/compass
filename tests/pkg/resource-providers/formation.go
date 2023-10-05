package resource_providers

import (
	"context"
	"github.com/kyma-incubator/compass/tests/pkg/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type FormationProvider struct {
	formationName         string
	formationTemplateName *string
	tenantID              string
	asserters             []asserters.Asserter
}

func NewFormationProvider(formationName string, tenantID string, formationTemplateName *string) *FormationProvider {
	return &FormationProvider{formationName: formationName, tenantID: tenantID, formationTemplateName: formationTemplateName}
}

func (o *FormationProvider) Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string {
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, gqlClient, o.tenantID, o.formationName, o.formationTemplateName)
	return formation.ID
}

func (o *FormationProvider) TearDown(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.DeleteFormationWithinTenant(t, ctx, gqlClient, o.tenantID, o.formationName)
}
