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

type ApplicationInput struct {
	Name           string
	Description    *string
	Labels         map[string]interface{}
	HealthCheckURL *string
	Webhooks       []*WebhookInput
	Apis           []*APIDefinitionInput
	EventAPIs      []*EventAPIDefinitionInput
	Documents      []*DocumentInput
}

func (i *ApplicationInput) ToApplication(id, tenant string) *Application {
	if i == nil {
		return nil
	}

	return &Application{
		ID:             id,
		Name:           i.Name,
		Description:    i.Description,
		Tenant:         tenant,
		HealthCheckURL: i.HealthCheckURL,
	}
}

func (i *ApplicationInput) Validate() error {
	if errorMgs := validation.NameIsDNSSubdomain(i.Name, false); errorMgs != nil {
		return errors.Errorf("%v", errorMgs)
	}
	return nil
}
