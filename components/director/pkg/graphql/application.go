package graphql

type Application struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	IntegrationSystemID *string             `json:"integrationSystemID"`
	Description         *string            `json:"description"`
	Status              *ApplicationStatus `json:"status"`
	HealthCheckURL      *string            `json:"healthCheckURL"`
}

// Extended types used by external API

type ApplicationPageExt struct {
	ApplicationPage
	Data []*ApplicationExt `json:"data"`
}

type ApplicationExt struct {
	Application
	Labels    Labels                    `json:"labels"`
	Webhooks  []Webhook                 `json:"webhooks"`
	Apis      APIDefinitionPageExt      `json:"apis"`
	EventAPIs EventAPIDefinitionPageExt `json:"eventAPIs"`
	Api       APIDefinition             `json:"api"`
	EventAPI  EventAPIDefinition        `json:"eventAPI"`
	Documents DocumentPageExt           `json:"documents"`
	Auths     []*SystemAuth             `json:"auths"`
}
