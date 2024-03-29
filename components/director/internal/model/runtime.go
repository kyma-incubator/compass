package model

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// Runtime missing godoc
type Runtime struct {
	ID                   string
	Name                 string
	Description          *string
	Status               *RuntimeStatus
	CreationTimestamp    time.Time
	ApplicationNamespace *string
}

// GetID missing godoc
func (runtime *Runtime) GetID() string {
	return runtime.ID
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
	Name                 string
	Description          *string
	Labels               map[string]interface{}
	Webhooks             []*WebhookInput
	StatusCondition      *RuntimeStatusCondition
	ApplicationNamespace *string
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
		CreationTimestamp:    creationTimestamp,
		ApplicationNamespace: i.ApplicationNamespace,
	}
}

// RuntimeUpdateInput missing godoc
type RuntimeUpdateInput struct {
	Name                 string
	Description          *string
	Labels               map[string]interface{}
	StatusCondition      *RuntimeStatusCondition
	ApplicationNamespace *string
}

// SetFromUpdateInput sets fields to model Runtime from RuntimeUpdateInput
func (runtime *Runtime) SetFromUpdateInput(update RuntimeUpdateInput, id string, creationTimestamp, conditionTimestamp time.Time) {
	if runtime.Status == nil {
		runtime.Status = &RuntimeStatus{}
	}

	runtime.ID = id
	runtime.Name = update.Name

	runtime.Status.Condition = getRuntimeStatusConditionOrDefault(update.StatusCondition)
	runtime.Status.Timestamp = conditionTimestamp
	runtime.CreationTimestamp = creationTimestamp

	if update.Description != nil {
		runtime.Description = update.Description
	}

	if update.ApplicationNamespace != nil {
		runtime.ApplicationNamespace = update.ApplicationNamespace
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
