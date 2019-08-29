package graphql

type Application struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Description    *string            `json:"description"`
	Status         *ApplicationStatus `json:"status"`
	HealthCheckURL *string            `json:"healthCheckURL"`
}

// Extended types used by external API

type ApplicationPageExt struct {
	Data       []*ApplicationExt `json:"data"`
	PageInfo   *PageInfo         `json:"pageInfo"`
	TotalCount int               `json:"totalCount"`
}

type ApplicationExt struct {
	Application
	Labels    Labels                 `json:"labels"`
	Webhooks  []Webhook              `json:"webhooks"`
	Apis      APIDefinitionPage      `json:"apis"`
	EventAPIs EventAPIDefinitionPage `json:"eventAPIs"`
	Documents DocumentPage           `json:"documents"`
}
