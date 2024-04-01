package operations

import (
	"context"
	context_keys "github.com/kyma-incubator/compass/tests/pkg/notifications/context-keys"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

type UpdateWebhookOperation struct {
	webhookType    graphql.WebhookType
	webhookMode    graphql.WebhookMode
	inputTemplate  string
	urlTemplate    string
	outputTemplate string
	headerTemplate *string
	objectID       string
	objectType     WebhookReferenceObjectType
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

func (o *UpdateWebhookOperation) WithHeaderTemplate(headerTemplate *string) *UpdateWebhookOperation {
	o.headerTemplate = headerTemplate
	return o
}

func (o *UpdateWebhookOperation) WithObjectID(objectID string) *UpdateWebhookOperation {
	o.objectID = objectID
	return o
}

func (o *UpdateWebhookOperation) WithObjectType(objectType WebhookReferenceObjectType) *UpdateWebhookOperation {
	o.objectType = objectType
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
	var webhooks []graphql.Webhook
	switch o.objectType {
	case WebhookReferenceObjectTypeApplication:
		webhooks = fixtures.GetApplication(t, ctx, gqlClient, o.tenantID, o.objectID).Webhooks
	case WebhookReferenceObjectTypeApplicationTemplate:
		webhooks = fixtures.GetApplicationTemplate(t, ctx, gqlClient, o.tenantID, o.objectID).Webhooks
	case WebhookReferenceObjectTypeFormationTemplate:
		formationTemplateID := ctx.Value(context_keys.FormationTemplateIDKey).(string)
		webhooksForTemplate := fixtures.QueryFormationTemplate(t, ctx, gqlClient, formationTemplateID).Webhooks
		for i, _ := range webhooksForTemplate {
			webhooks = append(webhooks, *webhooksForTemplate[i])
		}
	}

	var webhook graphql.Webhook
	for _, wh := range webhooks {
		if wh.Type == o.webhookType {
			webhook = wh
			break
		}
	}
	require.NotEmpty(t, webhook)

	input := fixtures.FixFormationNotificationWebhookInput(o.webhookType, o.webhookMode, o.urlTemplate, o.inputTemplate, o.outputTemplate, o.headerTemplate)
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
