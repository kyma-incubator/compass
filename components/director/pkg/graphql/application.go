package graphql

type Application struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	ProviderName        *string            `json:"providerName"`
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
	Auths                 []*SystemAuth                    `json:"auths"`
	Package               PackageExt                       `json:"package"`
	Packages              PackagePageExt                   `json:"packages"`
	EventingConfiguration ApplicationEventingConfiguration `json:"eventingConfiguration"`
}
