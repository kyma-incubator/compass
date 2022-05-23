package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// Runtime missing godoc
type Runtime struct {
	ID                string
	Name              string
	Description       *string
	Status            *RuntimeStatus
	CreationTimestamp time.Time
}

// RuntimeStatus missing godoc
type RuntimeStatus struct {
	Condition RuntimeStatusCondition
	Timestamp time.Time
}

// RuntimeStatusCondition missing godoc
type RuntimeStatusCondition string

const (
	// RuntimeStatusConditionInitial missing godoc
	RuntimeStatusConditionInitial RuntimeStatusCondition = "INITIAL"
	// RuntimeStatusConditionProvisioning missing godoc
	RuntimeStatusConditionProvisioning RuntimeStatusCondition = "PROVISIONING"
	// RuntimeStatusConditionConnected missing godoc
	RuntimeStatusConditionConnected RuntimeStatusCondition = "CONNECTED"
	// RuntimeStatusConditionFailed missing godoc
	RuntimeStatusConditionFailed RuntimeStatusCondition = "FAILED"
)

// RuntimeRegisterInput missing godoc
type RuntimeRegisterInput struct {
	Name            string
	Description     *string
	Labels          map[string]interface{}
	Webhooks        []*WebhookInput
	StatusCondition *RuntimeStatusCondition
}

// ToRuntime missing godoc
func (i *RuntimeRegisterInput) ToRuntime(id string, creationTimestamp, conditionTimestamp time.Time) *Runtime {
	if i == nil {
		return nil
	}

	return &Runtime{
		ID:          id,
		Name:        i.Name,
		Description: i.Description,
		Status: &RuntimeStatus{
			Condition: getRuntimeStatusConditionOrDefault(i.StatusCondition),
			Timestamp: conditionTimestamp,
		},
		CreationTimestamp: creationTimestamp,
	}
}

// RuntimeUpdateInput missing godoc
type RuntimeUpdateInput struct {
	Name            string
	Description     *string
	Labels          map[string]interface{}
	StatusCondition *RuntimeStatusCondition
}

// ToRuntime missing godoc
func (i *RuntimeUpdateInput) ToRuntime(id string, creationTimestamp, conditionTimestamp time.Time) *Runtime {
	if i == nil {
		return nil
	}

	return &Runtime{
		ID:          id,
		Name:        i.Name,
		Description: i.Description,
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

// RuntimePage missing godoc
type RuntimePage struct {
	Data       []*Runtime
	PageInfo   *pagination.Page
	TotalCount int
}
