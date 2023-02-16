package formation

import "database/sql"

// Entity represents the formation entity
type Entity struct {
	ID                  string         `db:"id"`
	TenantID            string         `db:"tenant_id"`
	FormationTemplateID string         `db:"formation_template_id"`
	Name                string         `db:"name"`
	State               string         `db:"state"`
	Error               sql.NullString `db:"error"`
}

// EntityCollection is a collection of formation entities.
type EntityCollection []*Entity

// Len returns the number of entities in the collection.
func (s EntityCollection) Len() int {
	return len(s)
}
