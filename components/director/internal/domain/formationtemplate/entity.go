package formationtemplate

import "database/sql"

// Entity represents the formation template entity
type Entity struct {
	ID                     string `db:"id"`
	Name                   string `db:"name"`
	ApplicationTypes       string `db:"application_types"`
	RuntimeTypes           string `db:"runtime_types"`
	RuntimeTypeDisplayName string `db:"runtime_type_display_name"`
	RuntimeArtifactKind    string `db:"runtime_artifact_kind"`
	FormationConstraints   string `db:"formation_constraints"`
	TenantID               sql.NullString `db:"tenant_id"`
}

// EntityCollection is a collection of formation template entities.
type EntityCollection []*Entity

// Len returns the number of entities in the collection.
func (s EntityCollection) Len() int {
	return len(s)
}
