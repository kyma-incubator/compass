package model

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/strings"
	"time"
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

type ApplicationWebhookType string

const (
	ApplicationWebhookTypeConfigurationChanged ApplicationWebhookType = "CONFIGURATION_CHANGED"
)


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
