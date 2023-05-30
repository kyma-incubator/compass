package destination

import "database/sql"

// Entity is a representation of a destination entity in the database.
type Entity struct {
	ID                    string         `db:"id"`
	Name                  string         `db:"name"`
	Type                  string         `db:"type"`
	URL                   string         `db:"url"`
	Authentication        string         `db:"authentication"`
	TenantID              string         `db:"tenant_id"`
	BundleID              sql.NullString `db:"bundle_id"`
	Revision              sql.NullString `db:"revision"`
	FormationAssignmentID sql.NullString `db:"formation_assignment_id"`
}

// EntityCollection missing godoc
type EntityCollection []Entity

// Len missing godoc
func (e EntityCollection) Len() int {
	return len(e)
}
