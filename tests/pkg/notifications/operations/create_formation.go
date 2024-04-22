package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
)

type CreateFormationOperation struct {
	name                  string
	state                 string
	formationTemplateName string
	tenantID              string
	id                    string
	asserters             []asserters.Asserter
}

func NewCreateFormationOperation(tenantID string) *CreateFormationOperation {
	return &CreateFormationOperation{tenantID: tenantID}
}

func (o *CreateFormationOperation) WithName(name string) *CreateFormationOperation {
	o.name = name
	return o
}

func (o *CreateFormationOperation) WithState(state string) *CreateFormationOperation {
	o.state = state
	return o
}

func (o *CreateFormationOperation) WithFormationTemplateName(formationTemplateName string) *CreateFormationOperation {
	o.formationTemplateName = formationTemplateName
	return o
}

func (o *CreateFormationOperation) WithAsserters(asserters ...asserters.Asserter) *CreateFormationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *CreateFormationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	var formation graphql.FormationExt
	if o.state != "" {
		formation = fixtures.CreateFormationFromTemplateWithStateWithinTenant(t, ctx, gqlClient, o.tenantID, o.name, &o.formationTemplateName, &o.state)
	} else {
		formation = fixtures.CreateFormationFromTemplateWithinTenant(t, ctx, gqlClient, o.tenantID, o.name, &o.formationTemplateName)
	}
	o.id = formation.ID

	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *CreateFormationOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	formationTemplateID := ctx.Value(context_keys.FormationTemplateIDKey).(string)
	formationTemplate := fixtures.QueryFormationTemplate(t, ctx, gqlClient, formationTemplateID)
	for _, webhook := range formationTemplate.Webhooks {
		fixtures.DeleteWebhook(t, ctx, gqlClient, o.tenantID, webhook.ID)
	}
	fixtures.DeleteFormationWithinTenant(t, ctx, gqlClient, o.tenantID, o.name)
}

func (o *CreateFormationOperation) Operation() Operation {
	return o
}
