package document

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity is a representation of a document entity in the database.
type Entity struct {
	BndlID      string         `db:"bundle_id"`
	AppID       string         `db:"app_id"`
	Title       string         `db:"title"`
	DisplayName string         `db:"display_name"`
	Description string         `db:"description"`
	Format      string         `db:"format"`
	Kind        sql.NullString `db:"kind"`
	Data        sql.NullString `db:"data"`
	*repo.BaseEntity
}

// GetParentID returns the parent ID of the entity.
func (e *Entity) GetParentID() string {
	return e.BndlID
}

// Collection is a collection of entities.
type Collection []Entity

// Len returns the length of the collection.
func (r Collection) Len() int {
	return len(r)
}
