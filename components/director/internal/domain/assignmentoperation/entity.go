package assignmentOperation

import (
	"time"
)

// Entity represents the assignment operation entity
type Entity struct {
	ID                    string     `db:"id"`
	Type                  string     `db:"type"`
	FormationAssignmentID string     `db:"formation_assignment_id"`
	FormationID           string     `db:"formation_id"`
	TriggeredBy           string     `db:"triggered_by"`
	StartedAtTimestamp    *time.Time `db:"started_at_timestamp"`
	FinishedAtTimestamp   *time.Time `db:"finished_at_timestamp"`
}

// EntityCollection is a collection of assignment operation entities.
type EntityCollection []*Entity

// Len is implementation of a repo.Collection interface
func (s EntityCollection) Len() int {
	return len(s)
}
