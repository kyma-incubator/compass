package scenarioassignement

type Entity struct {
	Scenario string `db:"scenario"`
	TenantID string `db:"tenant_id"`
	Key      string `db:"key"`
	Value    string `db:"value"`
}
