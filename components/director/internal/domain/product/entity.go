package product

import (
	"database/sql"
)

// Entity missing godoc
type Entity struct {
	ID               string         `db:"id"`
	OrdID            string         `db:"ord_id"`
	ApplicationID    string         `db:"app_id"`
	Title            string         `db:"title"`
	ShortDescription string         `db:"short_description"`
	Vendor           string         `db:"vendor"`
	Parent           sql.NullString `db:"parent"`
	CorrelationIDs   sql.NullString `db:"correlation_ids"`
	Labels           sql.NullString `db:"labels"`
}

func (e *Entity) GetID() string {
	return e.ID
}

func (e *Entity) GetParentID() string {
	return e.ApplicationID
}

