package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
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

func (o *AddWebhookToApplicationOperation) WithWebhookMode(webhookMode graphql.WebhookMode) *AddWebhookToApplicationOperation {
	o.webhookMode = webhookMode
	return o
}

func (o *AddWebhookToApplicationOperation) WithWebhookType(webhookType graphql.WebhookType) *AddWebhookToApplicationOperation {
	o.webhookType = webhookType
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

func (o *AddWebhookToApplicationOperation) WithAsserters(asserters ...asserters.Asserter) *AddWebhookToApplicationOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *AddWebhookToApplicationOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	applicationWebhookInput := fixtures.FixFormationNotificationWebhookInput(o.webhookType, o.webhookMode, o.urlTemplate, o.inputTemplate, o.outputTemplate)
	wh := fixtures.AddWebhookToApplication(t, ctx, gqlClient, applicationWebhookInput, o.tenantID, o.applicationID)
	o.webhookID = wh.ID

	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AddWebhookToApplicationOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupWebhook(t, ctx, gqlClient, o.tenantID, o.webhookID)
}

func (o *AddWebhookToApplicationOperation) Operation() Operation {
	return o
}
