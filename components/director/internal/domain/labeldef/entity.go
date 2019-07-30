package labeldef

type Entity struct {
	ID         string `db:"id"`
	TenantID   string `db:"tenantID"`
	Key        string `db:"key"`
	SchemaJSON string `db:"schemaJSON"`
}
