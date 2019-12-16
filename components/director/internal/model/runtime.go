package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

type Runtime struct {
	ID                string
	Name              string
	Description       *string
	Tenant            string
	Status            *RuntimeStatus
	CreationTimestamp time.Time
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

func (i *RuntimeInput) ToRuntime(id string, tenant string, creationTimestamp time.Time) *Runtime {
	if i == nil {
		return nil
	}

	return &Runtime{
		ID:                id,
		Name:              i.Name,
		Description:       i.Description,
		Tenant:            tenant,
		Status:            &RuntimeStatus{},
		CreationTimestamp: creationTimestamp,
	}
}

type RuntimePage struct {
	Data       []*Runtime
	PageInfo   *pagination.Page
	TotalCount int
}
