package model

type ApplicationWebhook struct {
	ApplicationID string
	ID            string
	Type          ApplicationWebhookType
	URL           string
	Auth          *Auth
}

type ApplicationWebhookInput struct {
	Type ApplicationWebhookType
	URL  string
	Auth *AuthInput
}

type ApplicationWebhookType string

const (
	ApplicationWebhookTypeConfigurationChanged ApplicationWebhookType = "CONFIGURATION_CHANGED"
)

func (i *ApplicationWebhookInput) ToWebhook(id, applicationID string) *ApplicationWebhook {
	if i == nil {
		return nil
	}

	return &ApplicationWebhook{
		ApplicationID: applicationID,
		ID:            id,
		Type:          i.Type,
		URL:           i.URL,
		Auth:          i.Auth.ToAuth(),
	}
}
