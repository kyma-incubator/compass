package model

import (
	"encoding/json"
	"strings"
	"time"
	"unsafe"

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

// InitialConfigurations holds the Initial configuration provided for the AssignFormation mutation.
// It is a map with key SourceID from the InitialConfiguration object and values - map with key TargetID
// from the InitialConfiguration Object and value - the configuration
type InitialConfigurations map[string]map[string]json.RawMessage

// FormationAssignment represent structure for FormationAssignment
type FormationAssignment struct {
	ID                            string                  `json:"id"`
	FormationID                   string                  `json:"formation_id"`
	TenantID                      string                  `json:"tenant_id"`
	Source                        string                  `json:"source"`
	SourceType                    FormationAssignmentType `json:"source_type"`
	Target                        string                  `json:"target"`
	TargetType                    FormationAssignmentType `json:"target_type"`
	State                         string                  `json:"state"`
	Value                         json.RawMessage         `json:"value"`
	Error                         json.RawMessage         `json:"error"`
	LastStateChangeTimestamp      *time.Time              `json:"last_state_change_timestamp"`
	LastNotificationSentTimestamp *time.Time              `json:"last_notification_sent_timestamp"`
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
	// DeleteErrorAssignmentState indicates that an error occurred during the deletion of the formation assignment
	DeleteErrorAssignmentState FormationAssignmentState = "DELETE_ERROR"
	// CreateReadyFormationAssignmentState indicates that the formation assignment is in a ready state and the response is for an assign notification
	CreateReadyFormationAssignmentState FormationAssignmentState = "CREATE_READY"
	// DeleteReadyFormationAssignmentState indicates that the formation assignment is in a ready state and the response is for an unassign notification
	DeleteReadyFormationAssignmentState FormationAssignmentState = "DELETE_READY"
	// NotificationRecursionDepthLimit is the maximum count of configuration exchanges during assigning an object to formation
	NotificationRecursionDepthLimit int = 10
)

// SupportedFormationAssignmentStates contains the supported formation assignment states
var SupportedFormationAssignmentStates = map[string]bool{
	string(InitialAssignmentState):              true,
	string(ReadyAssignmentState):                true,
	string(CreateReadyFormationAssignmentState): true,
	string(DeleteReadyFormationAssignmentState): true,
	string(ConfigPendingAssignmentState):        true,
	string(CreateErrorAssignmentState):          true,
	string(DeletingAssignmentState):             true,
	string(DeleteErrorAssignmentState):          true,
}

// ResynchronizableFormationAssignmentStates is an array of supported assignment states for resynchronization
var ResynchronizableFormationAssignmentStates = []string{string(InitialAssignmentState),
	string(DeletingAssignmentState),
	string(CreateErrorAssignmentState),
	string(DeleteErrorAssignmentState)}

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
func (fa *FormationAssignment) Clone() *FormationAssignment {
	return &FormationAssignment{
		ID:                            fa.ID,
		FormationID:                   fa.FormationID,
		TenantID:                      fa.TenantID,
		Source:                        fa.Source,
		SourceType:                    fa.SourceType,
		Target:                        fa.Target,
		TargetType:                    fa.TargetType,
		State:                         fa.State,
		Value:                         fa.Value,
		Error:                         fa.Error,
		LastStateChangeTimestamp:      fa.LastStateChangeTimestamp,
		LastNotificationSentTimestamp: fa.LastNotificationSentTimestamp,
	}
}

// IsInErrorState returns if the formation assignment is in error state
func (fa *FormationAssignment) IsInErrorState() bool {
	state := fa.State
	return state == string(CreateErrorAssignmentState) ||
		strings.HasSuffix(state, string(DeleteErrorAssignmentState))
}

// IsInProgressState returns if the formation assignment is in progress state
func (fa *FormationAssignment) IsInProgressState() bool {
	return fa.isInProgressAssignState() || fa.isInProgressUnassignState()
}

// SetStateToDeleting sets the state to deleting and returns if the formation assignment is updated
func (fa *FormationAssignment) SetStateToDeleting() bool {
	if fa.isInProgressUnassignState() {
		return false
	}
	if strings.HasSuffix(fa.State, string(DeleteErrorAssignmentState)) {
		fa.State = strings.Replace(fa.State, string(DeleteErrorAssignmentState), string(DeletingAssignmentState), 1)
		return true
	}
	fa.State = string(DeletingAssignmentState)
	return true
}

func (fa *FormationAssignment) isInProgressAssignState() bool {
	return fa.State == string(InitialAssignmentState) ||
		fa.State == string(ConfigPendingAssignmentState)
}

func (fa *FormationAssignment) isInProgressUnassignState() bool {
	return strings.HasSuffix(fa.State, string(DeletingAssignmentState))
}

// GetAddress returns the memory address of the FormationAssignment in form of an uninterpreted type(integer number)
// Currently, it's used in some formation constraints input templates, so we could propagate the memory address to the formation constraints operators and later on to modify/update it.
func (fa *FormationAssignment) GetAddress() uintptr {
	return uintptr(unsafe.Pointer(fa))
}
