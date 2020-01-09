package graphql

type Application struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	ProviderName        string             `json:"ProviderName"`
	IntegrationSystemID *string            `json:"integrationSystemID"`
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
	Labels                Labels                           `json:"labels"`
	Webhooks              []Webhook                        `json:"webhooks"`
	APIDefinitions        APIDefinitionPageExt             `json:"apiDefinitions"`
	EventDefinitions      EventAPIDefinitionPageExt        `json:"eventDefinitions"`
	APIDefinition         APIDefinition                    `json:"apiDefinition"`
	EventDefinition       EventDefinition                  `json:"eventDefinition"`
	Documents             DocumentPageExt                  `json:"documents"`
	Auths                 []*SystemAuth                    `json:"auths"`
	EventingConfiguration ApplicationEventingConfiguration `json:"eventingConfiguration"`
}
