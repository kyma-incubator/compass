package model

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/strings"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Runtime struct {
	ID          string
	Name        string
	Description *string
	Tenant      string
	Labels      map[string][]string
	Annotations map[string]interface{}
	Status      *RuntimeStatus
	AgentAuth   *Auth
}

type RuntimeStatus struct {
	Condition RuntimeStatusCondition
	Timestamp time.Time
}

type RuntimeStatusCondition string

const (
	RuntimeStatusConditionInitial RuntimeStatusCondition = "INITIAL"
	RuntimeStatusConditionReady   RuntimeStatusCondition = "READY"
	RuntimeStatusConditionFailed  RuntimeStatusCondition = "FAILED"
)

func (r *Runtime) AddLabel(key string, values []string) {
	if r.Labels == nil {
		r.Labels = make(map[string][]string)
	}

	if _, exists := r.Labels[key]; !exists {
		r.Labels[key] = strings.Unique(values)
		return
	}

	r.Labels[key] = strings.Unique(append(r.Labels[key], values...))
}

func (r *Runtime) DeleteLabel(key string, valuesToDelete []string) error {
	currentValues, exists := r.Labels[key]

	if !exists {
		return fmt.Errorf("label %s doesn't exist", key)
	}

	if len(valuesToDelete) == 0 {
		delete(r.Labels, key)
		return nil
	}

	set := strings.SliceToMap(currentValues)
	for _, val := range valuesToDelete {
		delete(set, val)
	}

	filteredValues := strings.MapToSlice(set)
	if len(filteredValues) == 0 {
		delete(r.Labels, key)
		return nil
	}

	r.Labels[key] = filteredValues
	return nil
}

func (r *Runtime) AddAnnotation(key string, value interface{}) error {
	if r.Annotations == nil {
		r.Annotations = make(map[string]interface{})
	}

	if _, exists := r.Annotations[key]; exists {
		return fmt.Errorf("annotation %s does already exist", key)
	}

	r.Annotations[key] = value
	return nil
}

func (r *Runtime) DeleteAnnotation(key string) error {
	if _, exists := r.Annotations[key]; !exists {
		return fmt.Errorf("annotation %s doesn't exist", key)
	}

	delete(r.Annotations, key)
	return nil
}

type RuntimeInput struct {
	Name        string
	Description *string
	Labels      map[string][]string
	Annotations map[string]interface{}
}

func (i *RuntimeInput) ToRuntime(id string, tenant string) *Runtime {
	return &Runtime{
		ID:          id,
		Name:        i.Name,
		Description: i.Description,
		Tenant:      tenant,
		Labels:      i.Labels,
		Annotations: i.Annotations,
	}
}

type RuntimePage struct {
	Data       []*Runtime
	PageInfo   *pagination.Page
	TotalCount int
}

type RuntimeAuth struct {
	RuntimeID string
	Auth      *Auth
}

