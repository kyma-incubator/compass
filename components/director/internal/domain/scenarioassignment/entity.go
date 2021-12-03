package scenarioassignment

// Entity missing godoc
type Entity struct {
	Scenario       string `db:"scenario"`
	TenantID       string `db:"tenant_id"`
	TargetTenantID string `db:"target_tenant_id"`
}

// EntityCollection missing godoc
type EntityCollection []Entity

// Len missing godoc
func (s EntityCollection) Len() int {
	return len(s)
}
