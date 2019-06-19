package model

type ApplicationWebhookInput struct {
	Type ApplicationWebhookType
	URL  string
	Auth *AuthInput
}


func (i *ApplicationWebhookInput) ToWebhook() *ApplicationWebhook {
	// TODO: Replace with actual model
	return &ApplicationWebhook{

	}
}


type ApplicationWebhook struct {
	ID   string                 `json:"id"`
	Type ApplicationWebhookType `json:"type"`
	URL  string                 `json:"url"`
	Auth *Auth                  `json:"auth"`
}

type ApplicationWebhookType string

const (
	ApplicationWebhookTypeConfigurationChanged ApplicationWebhookType = "CONFIGURATION_CHANGED"
)
