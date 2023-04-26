package tenant_business_type

// Entity represents the tenant business types entity
type Entity struct {
	ID   string `db:"id"`
	Code string `db:"code"`
	Name string `db:"name"`
}

// EntityCollection is a collection of tenant business types entities.
type EntityCollection []*Entity

// Len is implementation of a repo.Collection interface
func (s EntityCollection) Len() int {
	return len(s)
}
