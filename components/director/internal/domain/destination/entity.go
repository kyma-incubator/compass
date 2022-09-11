package destination

// Entity is a representation of a destination entity in the database.
type Entity struct {
	ID             string `db:"id"`
	Name           string `db:"name"`
	Type           string `db:"type"`
	URL            string `db:"url"`
	Authentication string `db:"authentication"`
	TenantID       string `db:"tenant_id"`
	BundleID       string `db:"bundle_id"`
	Revision       string `db:"revision"`
}
