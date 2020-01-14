package tenant

type Entity struct {
	ID             string       `db:"id"`
	Name           string       `db:"name"`
	ExternalTenant string       `db:"external_tenant"`
	InternalTenant string       `db:"internal_tenant"`
	ProviderName   string       `db:"provider_name"`
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
