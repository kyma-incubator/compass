package formation

// Entity represents the formation entity
type Entity struct {
	ID                  string `db:"id"`
	TenantID            string `db:"tenant_id"`
	FormationTemplateID string `db:"formation_template_id"`
	Name                string `db:"name"`
}
