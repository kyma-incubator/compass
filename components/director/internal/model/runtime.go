package model

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/validation"

	"github.com/kyma-incubator/compass/components/director/pkg/strings"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Runtime struct {
	ID          string
	Name        string
	Description *string
	Tenant      string
	Labels      map[string][]string
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

type RuntimeInput struct {
	Name        string
	Description *string
	Labels      map[string][]string
}

func (i *RuntimeInput) ToRuntime(id string, tenant string) *Runtime {
	if i == nil {
		return nil
	}

	return &Runtime{
		ID:          id,
		Name:        i.Name,
		Description: i.Description,
		Tenant:      tenant,
		Labels:      i.Labels,
		AgentAuth:   &Auth{},
		Status:      &RuntimeStatus{},
	}
}

func (i *RuntimeInput) ValidateInput() []string {
	return validation.NameIsDNSSubdomain(i.Name, false)
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
