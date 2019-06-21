package model

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/strings"
)

type Application struct {
	ID             string
	Tenant         string
	Name           string
	Description    *string
	Labels         map[string][]string
	Annotations    map[string]interface{}
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

func (a *Application) AddLabel(key string, values []string) {
	if a.Labels == nil {
		a.Labels = make(map[string][]string)
	}

	if _, exists := a.Labels[key]; !exists {
		a.Labels[key] = strings.Unique(values)
		return
	}

	a.Labels[key] = strings.Unique(append(a.Labels[key], values...))
}

func (a *Application) DeleteLabel(key string, valuesToDelete []string) error {
	currentValues, exists := a.Labels[key]

	if !exists {
		return fmt.Errorf("label %s doesn't exist", key)
	}

	if len(valuesToDelete) == 0 {
		delete(a.Labels, key)
		return nil
	}

	set := strings.SliceToMap(currentValues)
	for _, val := range valuesToDelete {
		delete(set, val)
	}

	filteredValues := strings.MapToSlice(set)
	if len(filteredValues) == 0 {
		delete(a.Labels, key)
		return nil
	}

	a.Labels[key] = filteredValues
	return nil
}

func (a *Application) AddAnnotation(key string, value interface{}) error {
	if a.Annotations == nil {
		a.Annotations = make(map[string]interface{})
	}

	if _, exists := a.Annotations[key]; exists {
		return fmt.Errorf("annotation %s does already exist", key)
	}

	a.Annotations[key] = value
	return nil
}

func (a *Application) DeleteAnnotation(key string) error {
	if _, exists := a.Annotations[key]; !exists {
		return fmt.Errorf("annotation %s doesn't exist", key)
	}

	delete(a.Annotations, key)
	return nil
}

type ApplicationInput struct {
	Name           string
	Description    *string
	Labels         map[string][]string
	Annotations    map[string]interface{}
	HealthCheckURL *string
	Webhooks       []*ApplicationWebhookInput
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
		Labels:         i.Labels,
		Annotations:    i.Annotations,
		HealthCheckURL: i.HealthCheckURL,
	}
}
