package formationtemplate

// Entity missing godoc
type Entity struct {
	ID                            string `db:"id"`
	Name                          string `db:"name"`
	ApplicationTypes              string `db:"application_types"`
	RuntimeTypes                  string `db:"runtime_types"`
	MissingArtifactInfoMessage    string `db:"missing_artifact_info_message"`
	MissingArtifactWarningMessage string `db:"missing_artifact_warning_message"`
}

// EntityCollection missing godoc
type EntityCollection []Entity

// Len missing godoc
func (s EntityCollection) Len() int {
	return len(s)
}
