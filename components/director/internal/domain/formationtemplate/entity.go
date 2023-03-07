package formationtemplate

import "database/sql"

// Entity represents the formation template entity
type Entity struct {
	ID                     string         `db:"id"`
	Name                   string         `db:"name"`
	ApplicationTypes       string         `db:"application_types"`
	RuntimeTypes           sql.NullString `db:"runtime_types"`
	RuntimeTypeDisplayName sql.NullString `db:"runtime_type_display_name"`
	RuntimeArtifactKind    sql.NullString `db:"runtime_artifact_kind"`
	LeadingProductIDs      sql.NullString `db:"leading_product_ids"`
	TenantID               sql.NullString `db:"tenant_id"`
}

// EntityCollection is a collection of formation template entities.
type EntityCollection []*Entity

// Len returns the number of entities in the collection.
func (s EntityCollection) Len() int {
	return len(s)
}
