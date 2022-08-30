package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Webhook represents a webhook that is called by Compass.
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

// WebhookInput represents a webhook input for creating/updating webhooks.
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

// WebhookType represents the type of the webhook.
type WebhookType string

const (
	// WebhookTypeConfigurationChanged represents a webhook that is called when a configuration is changed.
	WebhookTypeConfigurationChanged WebhookType = "CONFIGURATION_CHANGED"
	// WebhookTypeApplicationTenantMapping represents a webhook that is called for app to app notifications for formation changes.
	WebhookTypeApplicationTenantMapping WebhookType = "APPLICATION_TENANT_MAPPING"
	// WebhookTypeRegisterApplication represents a webhook that is called when an application is registered.
	WebhookTypeRegisterApplication WebhookType = "REGISTER_APPLICATION"
	// WebhookTypeDeleteApplication represents a webhook that is called when an application is deleted.
	WebhookTypeDeleteApplication WebhookType = "UNREGISTER_APPLICATION"
	// WebhookTypeOpenResourceDiscovery represents a webhook that is called to aggregate ORD information of a system.
	WebhookTypeOpenResourceDiscovery WebhookType = "OPEN_RESOURCE_DISCOVERY"
	// WebhookTypeUnpairApplication represents a webhook that is called when an application is unpaired.
	WebhookTypeUnpairApplication WebhookType = "UNPAIR_APPLICATION"
)

// WebhookMode represents the mode of the webhook.
type WebhookMode string

const (
	// WebhookModeSync represents a webhook that is called synchronously.
	WebhookModeSync WebhookMode = "SYNC"
	// WebhookModeAsync represents a webhook that is called asynchronously.
	WebhookModeAsync WebhookMode = "ASYNC"
)

// WebhookReferenceObjectType represents the type of the object that is referenced by the webhook.
type WebhookReferenceObjectType string

const (
	// UnknownWebhookReference is used when the webhook's reference entity cannot be determined.
	// For example in case of update we have only the target's webhook ID and the input, we cannot determine the reference entity.
	// In those cases an aggregated view with all the webhook ref entity tenant access views unioned together is used for tenant isolation.
	UnknownWebhookReference WebhookReferenceObjectType = "Unknown"
	// ApplicationWebhookReference is used when the webhook's reference entity is an application.
	ApplicationWebhookReference WebhookReferenceObjectType = "ApplicationWebhook"
	// RuntimeWebhookReference is used when the webhook's reference entity is a runtime.
	RuntimeWebhookReference WebhookReferenceObjectType = "RuntimeWebhook"
	// ApplicationTemplateWebhookReference is used when the webhook's reference entity is an application template.
	ApplicationTemplateWebhookReference WebhookReferenceObjectType = "ApplicationTemplateWebhook"
	// IntegrationSystemWebhookReference is used when the webhook's reference entity is an integration system.
	IntegrationSystemWebhookReference WebhookReferenceObjectType = "IntegrationSystemWebhook"
)

// GetResourceType returns the resource type of the webhook based on the referenced entity.
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

// ToWebhook converts the given input to a webhook.
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
