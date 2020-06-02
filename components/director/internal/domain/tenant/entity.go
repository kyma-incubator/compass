package tenant

type Entity struct {
	ID             string       `db:"id"`
	Name           string       `db:"external_name"`
	ExternalTenant string       `db:"external_tenant"`
	ProviderName   string       `db:"provider_name"`
	Initialized    *bool        `db:"initialized"` // computed value
	Status         TenantStatus `db:"status"`
}

type TenantStatus string

const (
	Active   TenantStatus = "Active"
	Inactive TenantStatus = "Inactive"
)

type EntityCollection []Entity

func (a EntityCollection) Len() int {
	return len(a)
}

func (e Entity) WithStatus(status TenantStatus) Entity {
	e.Status = status
	return e
}
