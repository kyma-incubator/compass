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

type FormationAssignmentState string

const (
	ReadyAssignmentState         FormationAssignmentState = "READY"
	ConfigPendingAssignmentState FormationAssignmentState = "CONFIG_PENDING"
	CreateErrorAssignmentState   FormationAssignmentState = "CREATE_ERROR"
	DeleteErrorAssignmentState   FormationAssignmentState = "DELETE_ERROR"
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
