package model

import (
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// Application missing godoc
type Application struct {
	ProviderName          *string
	Name                  string
	Description           *string
	Status                *ApplicationStatus
	HealthCheckURL        *string
	IntegrationSystemID   *string
	ApplicationTemplateID *string
	SystemNumber          *string
	LocalTenantID         *string
	BaseURL               *string         `json:"baseUrl"`
	OrdLabels             json.RawMessage `json:"labels"`
	CorrelationIDs        json.RawMessage `json:"correlationIds,omitempty"`
	Type                  string          `json:"-"`
	// SystemStatus shows whether the on-premise system is reachable or unreachable
	SystemStatus        *string
	DocumentationLabels json.RawMessage `json:"documentationLabels"`

	*BaseEntity
}

// GetType returns Type application
func (*Application) GetType() resource.Type {
	return resource.Application
}

// SetFromUpdateInput missing godoc
func (app *Application) SetFromUpdateInput(update ApplicationUpdateInput, timestamp time.Time) {
	if app.Status == nil {
		app.Status = &ApplicationStatus{}
	}
	app.Status.Condition = getApplicationStatusConditionOrDefault(update.StatusCondition)
	app.Status.Timestamp = timestamp

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
	if update.BaseURL != nil {
		app.BaseURL = update.BaseURL
	}
	if update.OrdLabels != nil {
		app.OrdLabels = update.OrdLabels
	}
	if update.CorrelationIDs != nil {
		app.CorrelationIDs = update.CorrelationIDs
	}
	if update.SystemStatus != nil {
		app.SystemStatus = update.SystemStatus
	}
	if update.LocalTenantID != nil {
		app.LocalTenantID = update.LocalTenantID
	}
	if update.DocumentationLabels != nil {
		app.DocumentationLabels = update.DocumentationLabels
	}
}

// ApplicationStatus missing godoc
type ApplicationStatus struct {
	Condition ApplicationStatusCondition
	Timestamp time.Time
}

// ApplicationStatusCondition missing godoc
type ApplicationStatusCondition string

const (
	// ApplicationStatusConditionInitial missing godoc
	ApplicationStatusConditionInitial ApplicationStatusCondition = "INITIAL"
	// ApplicationStatusConditionConnected missing godoc
	ApplicationStatusConditionConnected ApplicationStatusCondition = "CONNECTED"
	// ApplicationStatusConditionFailed missing godoc
	ApplicationStatusConditionFailed ApplicationStatusCondition = "FAILED"
	// ApplicationStatusConditionCreating missing godoc
	ApplicationStatusConditionCreating ApplicationStatusCondition = "CREATING"
	// ApplicationStatusConditionCreateFailed missing godoc
	ApplicationStatusConditionCreateFailed ApplicationStatusCondition = "CREATE_FAILED"
	// ApplicationStatusConditionCreateSucceeded missing godoc
	ApplicationStatusConditionCreateSucceeded ApplicationStatusCondition = "CREATE_SUCCEEDED"
	// ApplicationStatusConditionUpdating missing godoc
	ApplicationStatusConditionUpdating ApplicationStatusCondition = "UPDATING"
	// ApplicationStatusConditionUpdateFailed missing godoc
	ApplicationStatusConditionUpdateFailed ApplicationStatusCondition = "UPDATE_FAILED"
	// ApplicationStatusConditionUpdateSucceeded missing godoc
	ApplicationStatusConditionUpdateSucceeded ApplicationStatusCondition = "UPDATE_SUCCEEDED"
	// ApplicationStatusConditionDeleting missing godoc
	ApplicationStatusConditionDeleting ApplicationStatusCondition = "DELETING"
	// ApplicationStatusConditionDeleteFailed missing godoc
	ApplicationStatusConditionDeleteFailed ApplicationStatusCondition = "DELETE_FAILED"
	// ApplicationStatusConditionDeleteSucceeded missing godoc
	ApplicationStatusConditionDeleteSucceeded ApplicationStatusCondition = "DELETE_SUCCEEDED"
	// ApplicationStatusConditionUnpairing Status condition when an application is being unpairing
	ApplicationStatusConditionUnpairing ApplicationStatusCondition = "UNPAIRING"
	// ApplicationStatusConditionUnpairFailed Status condition when an application has failed to unpair
	ApplicationStatusConditionUnpairFailed ApplicationStatusCondition = "UNPAIR_FAILED"
)

// ApplicationPage missing godoc
type ApplicationPage struct {
	Data       []*Application
	PageInfo   *pagination.Page
	TotalCount int
}

// ApplicationRegisterInput missing godoc
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
	SystemNumber        *string
	OrdLabels           json.RawMessage
	CorrelationIDs      json.RawMessage
	SystemStatus        *string
	DocumentationLabels json.RawMessage
	LocalTenantID       *string
}

// ApplicationRegisterInputWithTemplate missing godoc
type ApplicationRegisterInputWithTemplate struct {
	ApplicationRegisterInput
	TemplateID string
}

// ToApplication missing godoc
func (i *ApplicationRegisterInput) ToApplication(timestamp time.Time, id string) *Application {
	if i == nil {
		return nil
	}

	return &Application{
		Name:                i.Name,
		Description:         i.Description,
		HealthCheckURL:      i.HealthCheckURL,
		IntegrationSystemID: i.IntegrationSystemID,
		ProviderName:        i.ProviderName,
		Status: &ApplicationStatus{
			Condition: getApplicationStatusConditionOrDefault(i.StatusCondition),
			Timestamp: timestamp,
		},
		BaseURL:             i.BaseURL,
		OrdLabels:           i.OrdLabels,
		CorrelationIDs:      i.CorrelationIDs,
		SystemNumber:        i.SystemNumber,
		LocalTenantID:       i.LocalTenantID,
		SystemStatus:        i.SystemStatus,
		DocumentationLabels: i.DocumentationLabels,
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

// ApplicationUpdateInput missing godoc
type ApplicationUpdateInput struct {
	ProviderName        *string
	Description         *string
	HealthCheckURL      *string
	IntegrationSystemID *string
	StatusCondition     *ApplicationStatusCondition
	BaseURL             *string
	OrdLabels           json.RawMessage
	CorrelationIDs      json.RawMessage
	SystemStatus        *string
	DocumentationLabels json.RawMessage
	LocalTenantID       *string
}

// ApplicationWithLabel missing godoc
type ApplicationWithLabel struct {
	App      *Application
	SccLabel *Label
}
