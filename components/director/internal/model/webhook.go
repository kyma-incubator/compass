package model

type Webhook struct {
	ID                    string
	TenantID              *string
	ApplicationID         *string
	ApplicationTemplateID *string
	RuntimeID             *string
	IntegrationSystemID   *string
	CorrelationIDKey      *string
	Type                  WebhookType
	URL                   *string
	Auth                  *Auth
	Mode                  *WebhookMode
	RetryInterval         *int
	Timeout               *int
	URLTemplate           *string
	InputTemplate         *string
	HeaderTemplate        *string
	OutputTemplate        *string
	StatusTemplate        *string
}

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

type WebhookType string

const (
	WebhookTypeConfigurationChanged  WebhookType = "CONFIGURATION_CHANGED"
	WebhookTypeRegisterApplication   WebhookType = "REGISTER_APPLICATION"
	WebhookTypeDeleteApplication     WebhookType = "UNREGISTER_APPLICATION"
	WebhookTypeOpenResourceDiscovery WebhookType = "OPEN_RESOURCE_DISCOVERY"
)

type WebhookMode string

const (
	WebhookModeSync  WebhookMode = "SYNC"
	WebhookModeAsync WebhookMode = "ASYNC"
)

type WebhookConverterFunc func(i *WebhookInput, id string, tenant *string, resourceID string) *Webhook

func (i *WebhookInput) ToApplicationWebhook(id string, tenant *string, applicationID string) *Webhook {
	if i == nil {
		return nil
	}

	webhook := i.toGenericWebhook(id)
	webhook.ApplicationID = &applicationID
	webhook.TenantID = tenant
	return webhook
}

func (i *WebhookInput) ToApplicationTemplateWebhook(id string, tenant *string, appTemplateID string) *Webhook {
	if i == nil {
		return nil
	}

	webhook := i.toGenericWebhook(id)
	webhook.ApplicationTemplateID = &appTemplateID
	return webhook
}

func (i *WebhookInput) toGenericWebhook(id string) *Webhook {
	return &Webhook{
		ID:               id,
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
