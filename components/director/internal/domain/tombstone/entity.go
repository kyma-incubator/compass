package tombstone

// Entity missing godoc
type Entity struct {
	ID            string `db:"id"`
	OrdID         string `db:"ord_id"`
	TenantID      string `db:"tenant_id"`
	ApplicationID string `db:"app_id"`
	RemovalDate   string `db:"removal_date"`
}
