package formationtemplateconstraintreferences

// Entity represents the formation constraint entity
type Entity struct {
	ConstraintID        string `db:"formation_constraint_id"`
	FormationTemplateID string `db:"formation_template_id"`
}

// EntityCollection is a collection of formationTemplate-constraint references entities.
type EntityCollection []*Entity

// Len is implementation of a repo.Collection interface
func (s EntityCollection) Len() int {
	return len(s)
}
