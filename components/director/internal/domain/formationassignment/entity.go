package formationassignment

import (
	"database/sql"
	"time"
)

// Entity represents the formation assignments entity
type Entity struct {
	ID                            string         `db:"id"`
	FormationID                   string         `db:"formation_id"`
	TenantID                      string         `db:"tenant_id"`
	Source                        string         `db:"source"`
	SourceType                    string         `db:"source_type"`
	Target                        string         `db:"target"`
	TargetType                    string         `db:"target_type"`
	State                         string         `db:"state"`
	Value                         sql.NullString `db:"value"`
	Error                         sql.NullString `db:"error"`
	LastStateChangeTimestamp      *time.Time     `db:"last_state_change_timestamp"`
	LastNotificationSentTimestamp *time.Time     `db:"last_notification_sent_timestamp"`
}

// EntityCollection is a collection of formation assignments entities.
type EntityCollection []*Entity

// Len is implementation of a repo.Collection interface
func (s EntityCollection) Len() int {
	return len(s)
}
