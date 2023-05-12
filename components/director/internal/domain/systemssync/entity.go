package systemssync

import (
	"time"
)

// Entity represents the systems last sync timestamps as an entity
type Entity struct {
	ID                string    `db:"id"`
	TenantID          string    `db:"tenant_id"`
	ProductID         string    `db:"product_id"`
	LastSyncTimestamp time.Time `db:"last_sync_timestamp"`
}

// EntityCollection is a collection of systems last sync timestamp entities
type EntityCollection []*Entity

// Len is implementation of a repo.Collection interface
func (a EntityCollection) Len() int {
	return len(a)
}
