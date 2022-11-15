package formationassignment

import "database/sql"

// Entity represents the formation assignments entity
type Entity struct {
	ID                         string         `db:"id"`
	FormationID                string         `db:"formation_id"`
	TenantID                   string         `db:"tenant_id"`
	Source                     string         `db:"source"`
	SourceType                 string         `db:"source_type"`
	Target                     string         `db:"target"`
	TargetType                 string         `db:"target_type"`
	LastOperation              string         `db:"last_operation"`
	LastOperationInitiator     string         `db:"last_operation_initiator"`
	LastOperationInitiatorType string         `db:"last_operation_initiator_type"`
	State                      string         `db:"state"`
	Value                      sql.NullString `db:"value"`
}

// EntityCollection is a collection of formation assignments entities.
type EntityCollection []*Entity

// Len is implementation of a repo.Collection interface
func (s EntityCollection) Len() int {
	return len(s)
}
