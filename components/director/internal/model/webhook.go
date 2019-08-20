package model

import "fmt"

type Webhook struct {
	ApplicationID string
	Tenant        string
	ID            string
	Type          WebhookType
	URL           string
	Auth          *Auth
}

func (w Webhook) PrettyString() string {
	return fmt.Sprintf("Webhook [URL: %s, Type: %s]", w.URL, w.Type)
}

type WebhookInput struct {
	Type WebhookType
	URL  string
	Auth *AuthInput
}

type WebhookType string

const (
	WebhookTypeConfigurationChanged WebhookType = "CONFIGURATION_CHANGED"
)

func (i *WebhookInput) ToWebhook(id, tenant, applicationID string) *Webhook {
	if i == nil {
		return nil
	}

	return &Webhook{
		ApplicationID: applicationID,
		ID:            id,
		Tenant:        tenant,
		Type:          i.Type,
		URL:           i.URL,
		Auth:          i.Auth.ToAuth(),
	}
}
