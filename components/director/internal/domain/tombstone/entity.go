package tombstone

// Entity missing godoc
type Entity struct {
	ID            string `db:"id"`
	OrdID         string `db:"ord_id"`
	ApplicationID string `db:"app_id"`
	RemovalDate   string `db:"removal_date"`
}

func (e *Entity) GetID() string {
	return e.ID
}

func (e *Entity) GetParentID() string {
	return e.ApplicationID
}
