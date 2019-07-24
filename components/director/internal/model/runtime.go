package model

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/api/validation"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Runtime struct {
	ID          string
	Name        string
	Description *string
	Tenant      string
	Labels      map[string]interface{}
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

func (r *Runtime) SetLabel(key string, value interface{}) {
	if r.Labels == nil {
		r.Labels = make(map[string]interface{})
	}

	r.Labels[key] = value
}

func (r *Runtime) DeleteLabel(key string) error {
	_, exists := r.Labels[key]

	if !exists {
		return fmt.Errorf("label %s doesn't exist", key)
	}

	delete(r.Labels, key)
	return nil
}

type RuntimeInput struct {
	Name        string
	Description *string
	Labels      map[string]interface{}
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

func (i *RuntimeInput) Validate() error {
	if errorMgs := validation.NameIsDNSSubdomain(i.Name, false); errorMgs != nil {
		return errors.Errorf("%v", errorMgs)
	}
	return nil
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
