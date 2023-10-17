package operations

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
	"testing"
)

type UpdateWebhookOperation struct {
	webhookType    graphql.WebhookType
	webhookMode    graphql.WebhookMode
	inputTemplate  string
	urlTemplate    string
	outputTemplate string
	applicationID  string
	tenantID       string
	asserters      []asserters.Asserter
}

func NewUpdateWebhookOperation() *UpdateWebhookOperation {
	return &UpdateWebhookOperation{}
}

func (o *UpdateWebhookOperation) WithWebhookType(webhookType graphql.WebhookType) *UpdateWebhookOperation {
	o.webhookType = webhookType
	return o
}

func (o *UpdateWebhookOperation) WithWebhookMode(webhookMode graphql.WebhookMode) *UpdateWebhookOperation {
	o.webhookMode = webhookMode
	return o
}

func (o *UpdateWebhookOperation) WithInputTemplate(inputTemplate string) *UpdateWebhookOperation {
	o.inputTemplate = inputTemplate
	return o
}

func (o *UpdateWebhookOperation) WithURLTemplate(urlTemplate string) *UpdateWebhookOperation {
	o.urlTemplate = urlTemplate
	return o
}

func (o *UpdateWebhookOperation) WithOutputTemplate(outputTemplate string) *UpdateWebhookOperation {
	o.outputTemplate = outputTemplate
	return o
}

func (o *UpdateWebhookOperation) WithApplicationID(applicationID string) *UpdateWebhookOperation {
	o.applicationID = applicationID
	return o
}

func (o *UpdateWebhookOperation) WithTenantID(tenantID string) *UpdateWebhookOperation {
	o.tenantID = tenantID
	return o
}

func (o *UpdateWebhookOperation) WithAsserters(asserters ...asserters.Asserter) *UpdateWebhookOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *UpdateWebhookOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	applicationExt := fixtures.GetApplication(t, ctx, gqlClient, o.tenantID, o.applicationID)
	var webhook graphql.Webhook
	for _, wh := range applicationExt.Webhooks {
		if wh.Type == o.webhookType {
			webhook = wh
			break
		}
	}
	require.NotEmpty(t, webhook)

	input := fixtures.FixFormationNotificationWebhookInput(o.webhookType, o.webhookMode, o.urlTemplate, o.inputTemplate, o.outputTemplate)
	updatedWebhook := fixtures.UpdateWebhook(t, ctx, gqlClient, o.tenantID, webhook.ID, input)
	require.Equal(t, updatedWebhook.ID, webhook.ID)
	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *UpdateWebhookOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {

}

func (o *UpdateWebhookOperation) Operation() Operation {
	return o
}
