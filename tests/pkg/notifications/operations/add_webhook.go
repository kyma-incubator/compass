package operations

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/notifications/asserters"
	gcli "github.com/machinebox/graphql"
)

type WebhookReferenceObjectType string

const (
	WebhookReferenceObjectTypeApplication         = "APPLICATION"
	WebhookReferenceObjectTypeApplicationTemplate = "APPLICATION_TEMPLATE"
	WebhookReferenceObjectTypeFormationTemplate   = "FORMATION_TEMPLATE"
)

type AddWebhookToObjectOperation struct {
	webhookType    graphql.WebhookType
	webhookMode    graphql.WebhookMode
	urlTemplate    string
	inputTemplate  string
	outputTemplate string
	objectID       string
	objectType     WebhookReferenceObjectType
	tenantID       string
	webhookID      string
	asserters      []asserters.Asserter
}

func NewAddWebhookToObjectOperation(webhookType graphql.WebhookType, webhookObjectType WebhookReferenceObjectType, objectID string, tenantID string) *AddWebhookToObjectOperation {
	return &AddWebhookToObjectOperation{webhookType: webhookType, objectID: objectID, objectType: webhookObjectType, tenantID: tenantID}
}

func (o *AddWebhookToObjectOperation) WithWebhookMode(webhookMode graphql.WebhookMode) *AddWebhookToObjectOperation {
	o.webhookMode = webhookMode
	return o
}

func (o *AddWebhookToObjectOperation) WithWebhookType(webhookType graphql.WebhookType) *AddWebhookToObjectOperation {
	o.webhookType = webhookType
	return o
}

func (o *AddWebhookToObjectOperation) WithURLTemplate(urlTemplate string) *AddWebhookToObjectOperation {
	o.urlTemplate = urlTemplate
	return o
}

func (o *AddWebhookToObjectOperation) WithInputTemplate(inputTemplate string) *AddWebhookToObjectOperation {
	o.inputTemplate = inputTemplate
	return o
}

func (o *AddWebhookToObjectOperation) WithOutputTemplate(outputTemplate string) *AddWebhookToObjectOperation {
	o.outputTemplate = outputTemplate
	return o
}

func (o *AddWebhookToObjectOperation) WithAsserters(asserters ...asserters.Asserter) *AddWebhookToObjectOperation {
	for i, _ := range asserters {
		o.asserters = append(o.asserters, asserters[i])
	}
	return o
}

func (o *AddWebhookToObjectOperation) Execute(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	webhookInput := fixtures.FixFormationNotificationWebhookInput(o.webhookType, o.webhookMode, o.urlTemplate, o.inputTemplate, o.outputTemplate)
	var wh *graphql.Webhook
	switch o.objectType {
	case WebhookReferenceObjectTypeApplication:
		wh = fixtures.AddWebhookToApplication(t, ctx, gqlClient, webhookInput, o.tenantID, o.objectID)
	case WebhookReferenceObjectTypeApplicationTemplate:
		wh = fixtures.AddWebhookToApplicationTemplate(t, ctx, gqlClient, webhookInput, o.tenantID, o.objectID)
	case WebhookReferenceObjectTypeFormationTemplate:
		wh = fixtures.AddWebhookToFormationTemplate(t, ctx, gqlClient, webhookInput, o.tenantID, o.objectID)
	}
	o.webhookID = wh.ID

	for _, asserter := range o.asserters {
		asserter.AssertExpectations(t, ctx)
	}
}

func (o *AddWebhookToObjectOperation) Cleanup(t *testing.T, ctx context.Context, gqlClient *gcli.Client) {
	fixtures.CleanupWebhook(t, ctx, gqlClient, o.tenantID, o.webhookID)
}

func (o *AddWebhookToObjectOperation) Operation() Operation {
	return o
}
