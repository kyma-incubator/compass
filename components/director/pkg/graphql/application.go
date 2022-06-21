package graphql

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// Application missing godoc
type Application struct {
	Name                  string             `json:"name"`
	ProviderName          *string            `json:"providerName"`
	IntegrationSystemID   *string            `json:"integrationSystemID"`
	ApplicationTemplateID *string            `json:"applicationTemplateID"`
	Description           *string            `json:"description"`
	Status                *ApplicationStatus `json:"status"`
	HealthCheckURL        *string            `json:"healthCheckURL"`
	SystemNumber          *string            `json:"systemNumber"`
	LocalTenantID         *string            `json:"localTenantID"`
	SystemStatus          *string            `json:"systemStatus"`
	BaseURL               *string            `json:"baseUrl"`
	*BaseEntity
}

// GetType missing godoc
func (e *Application) GetType() resource.Type {
	return resource.Application
}

// Sentinel missing godoc
func (e *Application) Sentinel() {}

// ApplicationPageExt is an extended type used by external API
type ApplicationPageExt struct {
	ApplicationPage
	Data []*ApplicationExt `json:"data"`
}

// ApplicationExt missing godoc
type ApplicationExt struct {
	Application
	Labels                Labels                           `json:"labels"`
	Webhooks              []Webhook                        `json:"webhooks"`
	Auths                 []*AppSystemAuth                 `json:"auths"`
	Bundle                BundleExt                        `json:"bundle"`
	Bundles               BundlePageExt                    `json:"bundles"`
	EventingConfiguration ApplicationEventingConfiguration `json:"eventingConfiguration"`
}
