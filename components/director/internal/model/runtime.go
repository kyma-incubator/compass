package model

import (
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
