package operations

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/asserters"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	gcli "github.com/machinebox/graphql"
	"testing"
)

type AddWebhookToApplicationOperation struct {
	webhookType    graphql.WebhookType
	webhookMode    graphql.WebhookMode
	urlTemplate    string
	inputTemplate  string
	outputTemplate string
	applicationID  string
	tenantID       string
	webhookID      string
	asserters      []asserters.Asserter
}

func NewAddWebhookToApplicationOperation(webhookType graphql.WebhookType, applicationID string, tenantID string) *AddWebhookToApplicationOperation {
	return &AddWebhookToApplicationOperation{webhookType: webhookType, applicationID: applicationID, tenantID: tenantID}
}

func (o *AddWebhookToApplicationOperation) WithMode(mode graphql.WebhookMode) *AddWebhookToApplicationOperation {
	o.webhookMode = mode
	return o
}

func (o *AddWebhookToApplicationOperation) WithURLTemplate(urlTemplate string) *AddWebhookToApplicationOperation {
	o.urlTemplate = urlTemplate
	return o
}

func (o *AddWebhookToApplicationOperation) WithInputTemplate(inputTemplate string) *AddWebhookToApplicationOperation {
	o.inputTemplate = inputTemplate
	return o
}

func (o *AddWebhookToApplicationOperation) WithOutputTemplate(outputTemplate string) *AddWebhookToApplicationOperation {
	o.outputTemplate = outputTemplate
	return o
}

func (o *AddWebhookToApplicationOperation) WithAsserter(asserter asserters.Asserter) *AddWebhookToApplicationOperation {
	o.asserters = append(o.asserters, asserter)
	return o
}

func (o *AddWebhookToApplicationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(o.webhookType, o.webhookMode, o.urlTemplate, o.inputTemplate, o.outputTemplate)
	wh := fixtures.AddWebhookToApplication(t, ctx, gqlClient, applicationWebhookInput, o.tenantID, o.applicationID)
	o.webhookID = wh.ID
}

func (o *AddWebhookToApplicationOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupWebhook(t, ctx, gqlClient, o.tenantID, o.webhookID)
}

func (o *AddWebhookToApplicationOperation) Operation() Operation {
	return o
}
