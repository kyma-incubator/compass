package apptemplate

import (
	"database/sql"
)

// Entity missing godoc
type Entity struct {
	ID                   string         `db:"id"`
	Name                 string         `db:"name"`
	Description          sql.NullString `db:"description"`
	ApplicationNamespace sql.NullString `db:"application_namespace"`
	ApplicationInputJSON string         `db:"application_input"`
	PlaceholdersJSON     sql.NullString `db:"placeholders"`
	AccessLevel          string         `db:"access_level"`
}

// EntityCollection missing godoc
type EntityCollection []Entity

// Len missing godoc
func (a EntityCollection) Len() int {
	return len(a)
}
