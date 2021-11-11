package model

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// Webhook missing godoc
type Webhook struct {
	ID               string
	ObjectID         string
	ObjectType       WebhookReferenceObjectType
	CorrelationIDKey *string
	Type             WebhookType
	URL              *string
	Auth             *Auth
	Mode             *WebhookMode
	RetryInterval    *int
	Timeout          *int
	URLTemplate      *string
	InputTemplate    *string
	HeaderTemplate   *string
	OutputTemplate   *string
	StatusTemplate   *string
}

// WebhookInput missing godoc
type WebhookInput struct {
	CorrelationIDKey *string
	Type             WebhookType
	URL              *string
	Auth             *AuthInput
	Mode             *WebhookMode
	RetryInterval    *int
	Timeout          *int
	URLTemplate      *string
	InputTemplate    *string
	HeaderTemplate   *string
	OutputTemplate   *string
	StatusTemplate   *string
}

// WebhookType missing godoc
type WebhookType string

const (
	// WebhookTypeConfigurationChanged missing godoc
	WebhookTypeConfigurationChanged WebhookType = "CONFIGURATION_CHANGED"
	// WebhookTypeRegisterApplication missing godoc
	WebhookTypeRegisterApplication WebhookType = "REGISTER_APPLICATION"
	// WebhookTypeDeleteApplication missing godoc
	WebhookTypeDeleteApplication WebhookType = "UNREGISTER_APPLICATION"
	// WebhookTypeOpenResourceDiscovery missing godoc
	WebhookTypeOpenResourceDiscovery WebhookType = "OPEN_RESOURCE_DISCOVERY"
	// WebhookTypeUnpairApplication WebhookType to describe unpairing application webhook
	WebhookTypeUnpairApplication WebhookType = "UNPAIR_APPLICATION"
)

// WebhookMode missing godoc
type WebhookMode string

const (
	// WebhookModeSync missing godoc
	WebhookModeSync WebhookMode = "SYNC"
	// WebhookModeAsync missing godoc
	WebhookModeAsync WebhookMode = "ASYNC"
)

// WebhookReferenceObjectType missing godoc
type WebhookReferenceObjectType string

const (
	// UnknownWebhookReference is used when the webhook's reference entity cannot be determined.
	// For example in case of update we have only the target's webhook ID and the input, we cannot determine the reference entity.
	// In those cases an aggregated view with all the webhook ref entity tenant access views unioned together is used for tenant isolation.
	UnknownWebhookReference WebhookReferenceObjectType = "Unknown"
	// ApplicationWebhookReference missing godoc
	ApplicationWebhookReference WebhookReferenceObjectType = "ApplicationWebhook"
	// RuntimeWebhookReference missing godoc
	RuntimeWebhookReference WebhookReferenceObjectType = "RuntimeWebhook"
	// ApplicationTemplateWebhookReference missing godoc
	ApplicationTemplateWebhookReference WebhookReferenceObjectType = "ApplicationTemplateWebhook"
	// IntegrationSystemWebhookReference missing godoc
	IntegrationSystemWebhookReference WebhookReferenceObjectType = "IntegrationSystemWebhook"
)

func (obj WebhookReferenceObjectType) GetResourceType() resource.Type {
	switch obj {
	case UnknownWebhookReference:
		return resource.Webhook
	case ApplicationWebhookReference:
		return resource.AppWebhook
	case RuntimeWebhookReference:
		return resource.RuntimeWebhook
	case ApplicationTemplateWebhookReference:
		return resource.Webhook
	case IntegrationSystemWebhookReference:
		return resource.Webhook
	}
	return ""
}

func (i *WebhookInput) ToWebhook(id, objID string, objectType WebhookReferenceObjectType) *Webhook {
	return &Webhook{
		ID:               id,
		ObjectID:         objID,
		ObjectType:       objectType,
		CorrelationIDKey: i.CorrelationIDKey,
		Type:             i.Type,
		URL:              i.URL,
		Auth:             i.Auth.ToAuth(),
		Mode:             i.Mode,
		RetryInterval:    i.RetryInterval,
		Timeout:          i.Timeout,
		URLTemplate:      i.URLTemplate,
		InputTemplate:    i.InputTemplate,
		HeaderTemplate:   i.HeaderTemplate,
		OutputTemplate:   i.OutputTemplate,
		StatusTemplate:   i.StatusTemplate,
	}
}
