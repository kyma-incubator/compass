package formationtemplateconstraintreferences

// Entity represents the formation constraint entity
type Entity struct {
	Constraint        string `db:"constraint"`
	FormationTemplate string `db:"formation_template"`
}

// EntityCollection is a collection of formationTemplate-constraint references entities.
type EntityCollection []*Entity

// Len is implementation of a repo.Collection interface
func (s EntityCollection) Len() int {
	return len(s)
}
