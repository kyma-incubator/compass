package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// FormationAssignmentType describes possible source and target types
type FormationAssignmentType string

const (
	// FormationAssignmentTypeApplication represent application type in formation assignment
	FormationAssignmentTypeApplication FormationAssignmentType = "APPLICATION"
	// FormationAssignmentTypeRuntime represent runtime type in formation assignment
	FormationAssignmentTypeRuntime FormationAssignmentType = "RUNTIME"
	// FormationAssignmentTypeRuntimeContext represent runtime context type in formation assignment
	FormationAssignmentTypeRuntimeContext FormationAssignmentType = "RUNTIME_CONTEXT"
)

// FormationAssignment represent structure for FormationAssignment
type FormationAssignment struct {
	ID                         string                  `json:"id"`
	FormationID                string                  `json:"formation_id"`
	TenantID                   string                  `json:"tenant_id"`
	Source                     string                  `json:"source"`
	SourceType                 FormationAssignmentType `json:"source_type"`
	Target                     string                  `json:"target"`
	TargetType                 FormationAssignmentType `json:"target_type"`
	LastOperation              FormationOperation      `json:"last_operation"`
	LastOperationInitiator     string                  `json:"last_operation_initiator"`
	LastOperationInitiatorType FormationAssignmentType `json:"last_operation_initiator_type"`
	State                      string                  `json:"state"`
	Value                      json.RawMessage         `json:"value"`
}

// FormationAssignmentInput is an input for creating a new FormationAssignment
type FormationAssignmentInput struct {
	FormationID                string                  `json:"formation_id"`
	Source                     string                  `json:"source"`
	SourceType                 FormationAssignmentType `json:"source_type"`
	Target                     string                  `json:"target"`
	TargetType                 FormationAssignmentType `json:"target_type"`
	LastOperation              FormationOperation      `json:"last_operation"`
	LastOperationInitiator     string                  `json:"last_operation_initiator"`
	LastOperationInitiatorType FormationAssignmentType `json:"last_operation_initiator_type"`
	State                      string                  `json:"state"`
	Value                      json.RawMessage         `json:"value"`
}

// FormationAssignmentPage missing godoc
type FormationAssignmentPage struct {
	Data       []*FormationAssignment
	PageInfo   *pagination.Page
	TotalCount int
}

// FormationAssignmentState represents the possible states a formation assignment can be in
type FormationAssignmentState string

const (
	// InitialAssignmentState indicates that nothing has been done with the formation assignment
	InitialAssignmentState FormationAssignmentState = "INITIAL"
	// ReadyAssignmentState indicates that the formation assignment is in a ready state
	ReadyAssignmentState FormationAssignmentState = "READY"
	// ConfigPendingAssignmentState indicates that the config is either missing or not finalized in the formation assignment
	ConfigPendingAssignmentState FormationAssignmentState = "CONFIG_PENDING"
	// CreateErrorAssignmentState indicates that an error occurred during the creation of the formation assignment
	CreateErrorAssignmentState FormationAssignmentState = "CREATE_ERROR"
	// DeletingAssignmentState indicates that async unassing notification is sent and status report is expected from the reciever
	DeletingAssignmentState FormationAssignmentState = "DELETING"
	// DeleteErrorAssignmentState indicates that an error occurred during the deletion of the formation assignment
	DeleteErrorAssignmentState FormationAssignmentState = "DELETE_ERROR"
	// NotificationRecursionDepthLimit is the maximum count of configuration exchanges during assigning an object to formation
	NotificationRecursionDepthLimit int = 10
)

// ToModel converts FormationAssignmentInput to FormationAssignment
func (i *FormationAssignmentInput) ToModel(id, tenantID string) *FormationAssignment {
	if i == nil {
		return nil
	}

	return &FormationAssignment{
		ID:                         id,
		FormationID:                i.FormationID,
		TenantID:                   tenantID,
		Source:                     i.Source,
		SourceType:                 i.SourceType,
		Target:                     i.Target,
		TargetType:                 i.TargetType,
		LastOperation:              i.LastOperation,
		LastOperationInitiator:     i.LastOperationInitiator,
		LastOperationInitiatorType: i.LastOperationInitiatorType,
		State:                      i.State,
		Value:                      i.Value,
	}
}

// Clone clones the formation assignment
func (f *FormationAssignment) Clone() *FormationAssignment {
	return &FormationAssignment{
		ID:                         f.ID,
		FormationID:                f.FormationID,
		TenantID:                   f.TenantID,
		Source:                     f.Source,
		SourceType:                 f.SourceType,
		Target:                     f.Target,
		TargetType:                 f.TargetType,
		LastOperation:              f.LastOperation,
		LastOperationInitiator:     f.LastOperationInitiator,
		LastOperationInitiatorType: f.LastOperationInitiatorType,
		State:                      f.State,
		Value:                      f.Value,
	}
}
