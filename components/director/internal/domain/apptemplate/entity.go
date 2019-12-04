package apptemplate

import (
	"database/sql"
)

type Entity struct {
	ID                   string         `db:"id"`
	Name                 string         `db:"name"`
	Description          sql.NullString `db:"description"`
	ApplicationInputJSON string         `db:"application_input"`
	PlaceholdersJSON     sql.NullString `db:"placeholders"`
	AccessLevel          string         `db:"access_level"`
}

type EntityCollection []Entity

func (a EntityCollection) Len() int {
	return len(a)
}
