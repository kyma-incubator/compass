package model

import (
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"time"
)

// AssignmentOperationType describes possible assignment operation types - assign, unassign
type AssignmentOperationType string

const (
	// Assign denotes the operation is assigning object to a formation
	Assign AssignmentOperationType = "ASSIGN"
	// Unassign denotes the operation is unassigning object from a formation
	Unassign AssignmentOperationType = "UNASSIGN"
)

// OperationTrigger denotes what triggered the operation - assign, unassign, reset etc.
type OperationTrigger string

const (
	// AssignObject denotes the operation was triggered by assigning object a formation
	AssignObject OperationTrigger = "ASSIGN_OBJECT"
	// UnassignObject denotes the operation was triggered by unassigning object a formation
	UnassignObject OperationTrigger = "UNASSIGN_OBJECT"
	// ResetAssignment denotes the operation was triggered by resetting formation assignment
	ResetAssignment OperationTrigger = "RESET"
	// ResyncAssignment denotes the operation was triggered by resynchronizing formation assignment
	ResyncAssignment OperationTrigger = "RESYNC"
)

// AssignmentOperation represent structure for AssignmentOperation
type AssignmentOperation struct {
	ID                    string                  `json:"id"`
	Type                  AssignmentOperationType `json:"type"`
	FormationAssignmentID string                  `json:"formation_assignment_id"`
	FormationID           string                  `json:"formation_id"`
	TriggeredBy           OperationTrigger        `json:"triggered_by"`
	StartedAtTimestamp    *time.Time              `json:"started_at_timestamp"`
	FinishedAtTimestamp   *time.Time              `json:"finished_at_timestamp"`
}

// AssignmentOperationInput is an input for creating a new AssignmentOperation
type AssignmentOperationInput struct {
	Type                  AssignmentOperationType `json:"type"`
	FormationAssignmentID string                  `json:"formation_assignment_id"`
	FormationID           string                  `json:"formation_id"`
	TriggeredBy           OperationTrigger        `json:"triggered_by"`
}

// ToModel converts AssignmentOperationInput to AssignmentOperation model object with provided id
func (a AssignmentOperationInput) ToModel(id string) *AssignmentOperation {
	return &AssignmentOperation{
		ID:                    id,
		Type:                  a.Type,
		FormationAssignmentID: a.FormationAssignmentID,
		FormationID:           a.FormationID,
		TriggeredBy:           a.TriggeredBy,
	}
}

// AssignmentOperationPage missing godoc
type AssignmentOperationPage struct {
	Data       []*AssignmentOperation
	PageInfo   *pagination.Page
	TotalCount int
}

func FromFormationOperationType(op FormationOperation) AssignmentOperationType {
	switch op {
	case AssignFormation:
		return Assign
	case UnassignFormation:
		return Unassign
	}
	return ""
}
