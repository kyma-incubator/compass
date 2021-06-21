package tenant

type Entity struct {
	ID             string  `db:"id"`
	Name           string  `db:"external_name"`
	ExternalTenant string  `db:"external_tenant"`
	Parent         *string `db:"parent"`
	Type           Type    `db:"type"`
	ProviderName   string  `db:"provider_name"`
	Initialized    *bool   `db:"initialized"` // computed value
	Status         Status  `db:"status"`
}

type Type string

const (
	Account  Type = "account"
	Customer Type = "customer"
)

type Status string

const (
	Active   Status = "Active"
	Inactive Status = "Inactive"
)

type EntityCollection []Entity

func (a EntityCollection) Len() int {
	return len(a)
}

func (e Entity) WithStatus(status Status) Entity {
	e.Status = status
	return e
}
