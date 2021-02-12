package model

import (
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Application struct {
	ProviderName        *string
	Tenant              string
	Name                string
	Description         *string
	Status              *ApplicationStatus
	HealthCheckURL      *string
	IntegrationSystemID *string
	BaseURL             *string
	Labels              json.RawMessage
	*BaseEntity
}

func (_ *Application) GetType() resource.Type {
	return resource.Application
}

func (app *Application) SetFromUpdateInput(update ApplicationUpdateInput, timestamp time.Time) {
	if app.Status == nil {
		app.Status = &ApplicationStatus{}
	}
	if update.Description != nil {
		app.Description = update.Description
	}
	if update.HealthCheckURL != nil {
		app.HealthCheckURL = update.HealthCheckURL
	}
	if update.IntegrationSystemID != nil {
		app.IntegrationSystemID = update.IntegrationSystemID
	}
	if update.ProviderName != nil {
		app.ProviderName = update.ProviderName
	}
	app.Status.Condition = getApplicationStatusConditionOrDefault(update.StatusCondition)
	app.Status.Timestamp = timestamp
	if update.BaseURL != nil {
		app.BaseURL = update.BaseURL
	}
	if update.Labels != nil {
		app.Labels = update.Labels
	}
}

type ApplicationStatus struct {
	Condition ApplicationStatusCondition
	Timestamp time.Time
}

type ApplicationStatusCondition string

const (
	ApplicationStatusConditionInitial   ApplicationStatusCondition = "INITIAL"
	ApplicationStatusConditionConnected ApplicationStatusCondition = "CONNECTED"
	ApplicationStatusConditionFailed    ApplicationStatusCondition = "FAILED"
)

type ApplicationPage struct {
	Data       []*Application
	PageInfo   *pagination.Page
	TotalCount int
}

type ApplicationRegisterInput struct {
	Name                string
	ProviderName        *string
	Description         *string
	Labels              map[string]interface{}
	HealthCheckURL      *string
	Webhooks            []*WebhookInput
	Bundles             []*BundleCreateInput
	IntegrationSystemID *string
	StatusCondition     *ApplicationStatusCondition
	BaseURL             *string
	OrdLabels           json.RawMessage
}

func (i *ApplicationRegisterInput) ToApplication(timestamp time.Time, id, tenant string) *Application {
	if i == nil {
		return nil
	}

	return &Application{
		Name:                i.Name,
		Description:         i.Description,
		Tenant:              tenant,
		HealthCheckURL:      i.HealthCheckURL,
		IntegrationSystemID: i.IntegrationSystemID,
		ProviderName:        i.ProviderName,
		Status: &ApplicationStatus{
			Condition: getApplicationStatusConditionOrDefault(i.StatusCondition),
			Timestamp: timestamp,
		},
		BaseURL: i.BaseURL,
		Labels:  i.OrdLabels,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}
}

func getApplicationStatusConditionOrDefault(in *ApplicationStatusCondition) ApplicationStatusCondition {
	statusCondition := ApplicationStatusConditionInitial
	if in != nil {
		statusCondition = *in
	}

	return statusCondition
}

type ApplicationUpdateInput struct {
	ProviderName        *string
	Description         *string
	HealthCheckURL      *string
	IntegrationSystemID *string
	StatusCondition     *ApplicationStatusCondition
	BaseURL             *string
	Labels              json.RawMessage
}
