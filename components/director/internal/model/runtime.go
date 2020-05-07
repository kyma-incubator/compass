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
	RuntimeStatusConditionInitial      RuntimeStatusCondition = "INITIAL"
	RuntimeStatusConditionProvisioning RuntimeStatusCondition = "PROVISIONING"
	RuntimeStatusConditionConnected    RuntimeStatusCondition = "CONNECTED"
	RuntimeStatusConditionFailed       RuntimeStatusCondition = "FAILED"
)

type RuntimeInput struct {
	Name            string
	Description     *string
	Labels          map[string]interface{}
	StatusCondition *RuntimeStatusCondition
}

func (i *RuntimeInput) ToRuntime(id string, tenant string, creationTimestamp, conditionTimestamp time.Time) *Runtime {
	if i == nil {
		return nil
	}

	return &Runtime{
		ID:          id,
		Name:        i.Name,
		Description: i.Description,
		Tenant:      tenant,
		Status: &RuntimeStatus{
			Condition: getRuntimeStatusConditionOrDefault(i.StatusCondition),
			Timestamp: conditionTimestamp,
		},
		CreationTimestamp: creationTimestamp,
	}
}

func getRuntimeStatusConditionOrDefault(in *RuntimeStatusCondition) RuntimeStatusCondition {
	statusCondition := RuntimeStatusConditionInitial
	if in != nil {
		statusCondition = *in
	}

	return statusCondition
}

type RuntimePage struct {
	Data       []*Runtime
	PageInfo   *pagination.Page
	TotalCount int
}
