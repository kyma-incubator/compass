package formationtemplate

// Entity represents the formation template entity
type Entity struct {
	ID                            string `db:"id"`
	Name                          string `db:"name"`
	ApplicationTypes              string `db:"application_types"`
	RuntimeTypes                  string `db:"runtime_types"`
	MissingArtifactInfoMessage    string `db:"missing_artifact_info_message"`
	MissingArtifactWarningMessage string `db:"missing_artifact_warning_message"`
}

// EntityCollection is a collection of formation template entities.
type EntityCollection []*Entity

// Len returns the number of entities in the collection.
func (s EntityCollection) Len() int {
	return len(s)
}
