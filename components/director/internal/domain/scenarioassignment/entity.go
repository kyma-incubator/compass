package scenarioassignment

// Entity missing godoc
type Entity struct {
	Scenario      string `db:"scenario"`
	TenantID      string `db:"tenant_id"`
	SelectorKey   string `db:"selector_key"`
	SelectorValue string `db:"selector_value"`
}

// EntityCollection missing godoc
type EntityCollection []Entity

// Len missing godoc
func (s EntityCollection) Len() int {
	return len(s)
}
