package model

import (
	"time"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/validation"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Application struct {
	ID             string
	Tenant         string
	Name           string
	Description    *string
	Status         *ApplicationStatus
	HealthCheckURL *string
	IntegrationSystemID *string
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

const applicationNameMaxLength = 36

type ApplicationPage struct {
	Data       []*Application
	PageInfo   *pagination.Page
	TotalCount int
}

type ApplicationCreateInput struct {
	Name           string
	Description    *string
	Labels         map[string]interface{}
	HealthCheckURL *string
	Webhooks       []*WebhookInput
	Apis           []*APIDefinitionInput
	EventAPIs      []*EventAPIDefinitionInput
	Documents      []*DocumentInput
	IntegrationSystemID *string
}

func (i *ApplicationCreateInput) ToApplication(timestamp time.Time, condition ApplicationStatusCondition, id, tenant string) *Application {
	if i == nil {
		return nil
	}

	return &Application{
		ID:             id,
		Name:           i.Name,
		Description:    i.Description,
		Tenant:         tenant,
		HealthCheckURL: i.HealthCheckURL,
		IntegrationSystemID:  i.IntegrationSystemID,
		Status: &ApplicationStatus{
			Condition: condition,
			Timestamp: timestamp,
		},
	}
}

func (i *ApplicationCreateInput) Validate() error {
	return validateApplicationName(i.Name)
}

type ApplicationUpdateInput struct {
	Name           string
	Description    *string
	HealthCheckURL *string
	IntegrationSystemID *string
}

func (i *ApplicationUpdateInput) Validate() error {
	return validateApplicationName(i.Name)
}

func validateApplicationName(name string) error {
	if errorMsg := validation.NameIsDNSSubdomain(name, false); errorMsg != nil {
		return errors.Errorf("%v", errorMsg)
	}
	if len(name) > applicationNameMaxLength {
		return errors.Errorf("application name is too long, must be maximum %d characters long", applicationNameMaxLength)
	}
	return nil
}
