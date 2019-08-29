package model

type Webhook struct {
	ApplicationID string
	Tenant        string
	ID            string
	Type          WebhookType
	URL           string
	Auth          *Auth
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
