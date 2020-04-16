package scenarioassignment

type Entity struct {
	Scenario      string `db:"scenario"`
	TenantID      string `db:"tenant_id"`
	SelectorKey   string `db:"selector_key"`
	SelectorValue string `db:"selector_value"`
}

type EntityCollection []Entity

func (s EntityCollection) Len() int {
	return len(s)
}
