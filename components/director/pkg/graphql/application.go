package graphql

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

type Application struct {
	Name                string             `json:"name"`
	ProviderName        *string            `json:"providerName"`
	IntegrationSystemID *string            `json:"integrationSystemID"`
	Description         *string            `json:"description"`
	Status              *ApplicationStatus `json:"status"`
	HealthCheckURL      *string            `json:"healthCheckURL"`
	*BaseEntity
}

func (e *Application) GetType() resource.Type {
	return resource.Application
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
	Bundle                BundleExt                        `json:"bundle"`
	Bundles               BundlePageExt                    `json:"bundles"`
	EventingConfiguration ApplicationEventingConfiguration `json:"eventingConfiguration"`
}
