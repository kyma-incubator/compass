package model

import (
	"encoding/json"
	"strings"
	"unsafe"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
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
	ID          string                  `json:"id"`
	FormationID string                  `json:"formation_id"`
	TenantID    string                  `json:"tenant_id"`
	Source      string                  `json:"source"`
	SourceType  FormationAssignmentType `json:"source_type"`
	Target      string                  `json:"target"`
	TargetType  FormationAssignmentType `json:"target_type"`
	State       string                  `json:"state"`
	Value       json.RawMessage         `json:"value"`
	Error       json.RawMessage         `json:"error"`
}

// FormationAssignmentInput is an input for creating a new FormationAssignment
type FormationAssignmentInput struct {
	FormationID string                  `json:"formation_id"`
	Source      string                  `json:"source"`
	SourceType  FormationAssignmentType `json:"source_type"`
	Target      string                  `json:"target"`
	TargetType  FormationAssignmentType `json:"target_type"`
	State       string                  `json:"state"`
	Value       json.RawMessage         `json:"value"`
	Error       json.RawMessage         `json:"error"`
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
	// DeletingAssignmentState indicates that async unassign notification is sent and status report is expected from the receiver
	DeletingAssignmentState FormationAssignmentState = "DELETING"
	// InstanceCreatorDeletingAssignmentState indicates that async unassign notification is sent and status report is expected from the receiver
	InstanceCreatorDeletingAssignmentState FormationAssignmentState = "INSTANCE_CREATOR_DELETING"
	// DeleteErrorAssignmentState indicates that an error occurred during the deletion of the formation assignment
	DeleteErrorAssignmentState FormationAssignmentState = "DELETE_ERROR"
	// InstanceCreatorDeleteErrorAssignmentState indicates that an error occurred during the deletion of the formation assignment by the instance creator operator
	InstanceCreatorDeleteErrorAssignmentState FormationAssignmentState = "INSTANCE_CREATOR_DELETE_ERROR"
	// NotificationRecursionDepthLimit is the maximum count of configuration exchanges during assigning an object to formation
	NotificationRecursionDepthLimit int = 10
)

// SupportedFormationAssignmentStates contains the supported formation assignment states
var SupportedFormationAssignmentStates = map[string]bool{
	string(InitialAssignmentState):       true,
	string(ReadyAssignmentState):         true,
	string(ConfigPendingAssignmentState): true,
	string(CreateErrorAssignmentState):   true,
	string(DeletingAssignmentState):      true,
	string(DeleteErrorAssignmentState):   true,
}

// ResynchronizableFormationAssignmentStates is an array of supported assignment states for resynchronization
var ResynchronizableFormationAssignmentStates = []string{string(InitialAssignmentState),
	string(DeletingAssignmentState),
	string(InstanceCreatorDeletingAssignmentState),
	string(CreateErrorAssignmentState),
	string(DeleteErrorAssignmentState),
	string(InstanceCreatorDeleteErrorAssignmentState)}

// ToModel converts FormationAssignmentInput to FormationAssignment
func (i *FormationAssignmentInput) ToModel(id, tenantID string) *FormationAssignment {
	if i == nil {
		return nil
	}

	return &FormationAssignment{
		ID:          id,
		FormationID: i.FormationID,
		TenantID:    tenantID,
		Source:      i.Source,
		SourceType:  i.SourceType,
		Target:      i.Target,
		TargetType:  i.TargetType,
		State:       i.State,
		Value:       i.Value,
		Error:       i.Error,
	}
}

// Clone clones the formation assignment
func (f *FormationAssignment) Clone() *FormationAssignment {
	return &FormationAssignment{
		ID:          f.ID,
		FormationID: f.FormationID,
		TenantID:    f.TenantID,
		Source:      f.Source,
		SourceType:  f.SourceType,
		Target:      f.Target,
		TargetType:  f.TargetType,
		State:       f.State,
		Value:       f.Value,
		Error:       f.Error,
	}
}

// IsInErrorState returns if the formation assignment is in error state
func (f *FormationAssignment) IsInErrorState() bool {
	state := f.State
	return state == string(CreateErrorAssignmentState) ||
		strings.HasSuffix(state, string(DeleteErrorAssignmentState))
}

// IsInProgressState returns if the formation assignment is in progress state
func (f *FormationAssignment) IsInProgressState() bool {
	return f.isInProgressAssignState() || f.isInProgressUnassignState()
}

// IsInRegularUnassignState returns if the formation assignment is in regular unassign stage
func (f *FormationAssignment) IsInRegularUnassignState() bool {
	unassignOperationStates := []string{string(DeletingAssignmentState),
		string(DeleteErrorAssignmentState)}
	return str.ValueIn(f.State, unassignOperationStates)
}

// SetStateToDeleting sets the state to deleting and returns if the formation assignment is updated
func (f *FormationAssignment) SetStateToDeleting() bool {
	if f.isInProgressUnassignState() {
		return false
	}
	if strings.HasSuffix(f.State, string(DeleteErrorAssignmentState)) {
		f.State = strings.Replace(f.State, string(DeleteErrorAssignmentState), string(DeletingAssignmentState), 1)
		return true
	}
	f.State = string(DeletingAssignmentState)
	return true
}

// GetOperation returns the formation operation that is determined based on the state of the assignment
func (f *FormationAssignment) GetOperation() FormationOperation {
	operation := AssignFormation
	if f.State == string(DeleteErrorAssignmentState) || f.State == string(DeletingAssignmentState) ||
		f.State == string(InstanceCreatorDeleteErrorAssignmentState) || f.State == string(InstanceCreatorDeletingAssignmentState) {
		operation = UnassignFormation
	}
	return operation
}

// GetNotificationState returns the assignment state that is to be sent for notifications
// and is exposed to the GraphQL layer
func (f *FormationAssignment) GetNotificationState() string {
	state := f.State
	if strings.HasSuffix(state, string(DeleteErrorFormationState)) {
		state = string(DeleteErrorFormationState)
	} else if f.isInProgressUnassignState() {
		state = string(DeletingAssignmentState)
	}
	return state
}

func (f *FormationAssignment) isInProgressAssignState() bool {
	return f.State == string(InitialAssignmentState) ||
		f.State == string(ConfigPendingAssignmentState)
}

func (f *FormationAssignment) isInProgressUnassignState() bool {
	return strings.HasSuffix(f.State, string(DeletingAssignmentState))
}

// GetAddress returns the memory address of the FormationAssignment in form of an uninterpreted type(integer number)
// Currently, it's used in some formation constraints input templates, so we could propagate the memory address to the formation constraints operators and later on to modify/update it.
func (f *FormationAssignment) GetAddress() uintptr {
	return uintptr(unsafe.Pointer(f))
}
