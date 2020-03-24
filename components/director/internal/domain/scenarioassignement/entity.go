package scenarioassignement

type Entity struct {
	Scenario string `db:"scenario"`
	TenantID string `db:"tenant_id"`
	SelectorKey      string `db:"selector_key"`
	SelectorValue    string `db:"selector_value"`
}
