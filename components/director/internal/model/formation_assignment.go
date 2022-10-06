package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// FormationAssignment represent structure for FormationAssignment
type FormationAssignment struct {
	ID          string          `json:"id"`
	FormationID string          `json:"formation_id"`
	TenantID    string          `json:"tenant_id"`
	Source      string          `json:"source"`
	SourceType  string          `json:"source_type"`
	Target      string          `json:"target"`
	TargetType  string          `json:"target_type"`
	State       string          `json:"state"`
	Value       json.RawMessage `json:"value"`
}

// FormationAssignmentInput is an input for creating a new FormationAssignment
type FormationAssignmentInput struct {
	FormationID string          `json:"formation_id"`
	Source      string          `json:"source"`
	SourceType  string          `json:"source_type"`
	Target      string          `json:"target"`
	TargetType  string          `json:"target_type"`
	State       string          `json:"state"`
	Value       json.RawMessage `json:"value"`
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
	// DeleteErrorAssignmentState indicates that an error occurred during the deletion of the formation assignment
	DeleteErrorAssignmentState FormationAssignmentState = "DELETE_ERROR"
)

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
	}
}
