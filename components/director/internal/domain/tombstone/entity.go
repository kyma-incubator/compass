package tombstone

type Entity struct {
	OrdID         string `db:"ord_id"`
	TenantID      string `db:"tenant_id"`
	ApplicationID string `db:"app_id"`
	RemovalDate   string `db:"removal_date"`
}
