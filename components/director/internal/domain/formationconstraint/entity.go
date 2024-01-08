package formationconstraint

import "time"

// Entity represents the formation constraint entity
type Entity struct {
	ID              string     `db:"id"`
	Name            string     `db:"name"`
	Description     string     `db:"description"`
	ConstraintType  string     `db:"constraint_type"`
	TargetOperation string     `db:"target_operation"`
	Operator        string     `db:"operator"`
	ResourceType    string     `db:"resource_type"`
	ResourceSubtype string     `db:"resource_subtype"`
	InputTemplate   string     `db:"input_template"`
	ConstraintScope string     `db:"constraint_scope"`
	Priority        int        `db:"priority"`
	CreatedAt       *time.Time `db:"created_at"`
}

// EntityCollection is a collection of formation constraint entities.
type EntityCollection []Entity

// Len is implementation of a repo.Collection interface
func (s EntityCollection) Len() int {
	return len(s)
}
