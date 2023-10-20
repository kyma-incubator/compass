package resource_providers

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
)

type FormationProvider struct {
	formationName         string
	formationTemplateName *string
	tenantID              string
}

func NewFormationProvider(formationName string, tenantID string, formationTemplateName *string) *FormationProvider {
	return &FormationProvider{formationName: formationName, tenantID: tenantID, formationTemplateName: formationTemplateName}
}

func (p *FormationProvider) Provide(t *testing.T, ctx context.Context, gqlClient *gcli.Client) string {
	formation := fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, gqlClient, p.tenantID, p.formationName, p.formationTemplateName)
	return formation.ID
}

func (p *FormationProvider) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.DeleteFormationWithinTenant(t, ctx, gqlClient, p.tenantID, p.formationName)
}
