package model

type Webhook struct {
	ID                  string
	TenantID            string
	ApplicationID       *string
	RuntimeID           *string
	IntegrationSystemID *string
	CorrelationIDKey    *string
	Type                WebhookType
	URL                 *string
	Auth                *Auth
	Mode                *WebhookMode
	RetryInterval       *int
	Timeout             *int
	URLTemplate         *string
	InputTemplate       *string
	HeaderTemplate      *string
	OutputTemplate      *string
	StatusTemplate      *string
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
	WebhookTypeConfigurationChanged WebhookType = "CONFIGURATION_CHANGED"
	WebhookTypeRegisterApplication  WebhookType = "REGISTER_APPLICATION"
	WebhookTypeDeleteApplication    WebhookType = "UNREGISTER_APPLICATION"
)

type WebhookMode string

const (
	WebhookModeSync  WebhookMode = "SYNC"
	WebhookModeAsync WebhookMode = "ASYNC"
)

func (i *WebhookInput) ToWebhook(id, tenant, applicationID string) *Webhook {
	if i == nil {
		return nil
	}

	return &Webhook{
		ID:             id,
		TenantID:       tenant,
		ApplicationID:  &applicationID,
		Type:           i.Type,
		URL:            i.URL,
		Auth:           i.Auth.ToAuth(),
		Mode:           i.Mode,
		RetryInterval:  i.RetryInterval,
		Timeout:        i.Timeout,
		URLTemplate:    i.URLTemplate,
		InputTemplate:  i.InputTemplate,
		HeaderTemplate: i.HeaderTemplate,
		OutputTemplate: i.OutputTemplate,
		StatusTemplate: i.StatusTemplate,
	}
}
