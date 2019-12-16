package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Application struct {
	ID                  string
	ProviderDisplayName string
	Tenant              string
	Name                string
	Description         *string
	Status              *ApplicationStatus
	HealthCheckURL      *string
	IntegrationSystemID *string
}

func (app *Application) UpdateApplication(update ApplicationUpdateInput) {
	app.Name = update.Name
	app.Description = update.Description
	app.HealthCheckURL = update.HealthCheckURL
	app.IntegrationSystemID = update.IntegrationSystemID
	app.ProviderDisplayName = update.ProviderDisplayName
}

type ApplicationStatus struct {
	Condition ApplicationStatusCondition
	Timestamp time.Time
}

type ApplicationStatusCondition string

const (
	ApplicationStatusConditionInitial ApplicationStatusCondition = "INITIAL"
	ApplicationStatusConditionUnknown ApplicationStatusCondition = "UNKNOWN"
	ApplicationStatusConditionReady   ApplicationStatusCondition = "READY"
	ApplicationStatusConditionFailed  ApplicationStatusCondition = "FAILED"
)

type ApplicationPage struct {
	Data       []*Application
	PageInfo   *pagination.Page
	TotalCount int
}

type ApplicationRegisterInput struct {
	Name                string
	ProviderDisplayName string
	Description         *string
	Labels              map[string]interface{}
	HealthCheckURL      *string
	Webhooks            []*WebhookInput
	APIDefinitions      []*APIDefinitionInput
	EventDefinitions    []*EventDefinitionInput
	Documents           []*DocumentInput
	IntegrationSystemID *string
}

func (i *ApplicationRegisterInput) ToApplication(timestamp time.Time, condition ApplicationStatusCondition, id, tenant string) *Application {
	if i == nil {
		return nil
	}

	return &Application{
		ID:                  id,
		Name:                i.Name,
		Description:         i.Description,
		Tenant:              tenant,
		HealthCheckURL:      i.HealthCheckURL,
		IntegrationSystemID: i.IntegrationSystemID,
		ProviderDisplayName: i.ProviderDisplayName,
		Status: &ApplicationStatus{
			Condition: condition,
			Timestamp: timestamp,
		},
	}
}

type ApplicationUpdateInput struct {
	Name                string
	ProviderDisplayName string
	Description         *string
	HealthCheckURL      *string
	IntegrationSystemID *string
}
