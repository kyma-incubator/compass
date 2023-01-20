package certsubjectmapping

// Entity is a representation of a certificate subject mapping in the DB
type Entity struct {
	ID                 string `db:"id"`
	Subject            string `db:"subject"`
	ConsumerType       string `db:"consumer_type"`
	InternalConsumerID *string `db:"internal_consumer_id"`
	TenantAccessLevels string `db:"tenant_access_levels"`
}

// EntityCollection is a collection of certificate subject mapping entities.
type EntityCollection []*Entity

// Len returns the number of entities in the collection.
func (s EntityCollection) Len() int {
	return len(s)
}
