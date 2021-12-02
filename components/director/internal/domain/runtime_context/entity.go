package runtimectx

import "github.com/kyma-incubator/compass/components/director/pkg/resource"

// RuntimeContext struct represents database entity for RuntimeContext
type RuntimeContext struct {
	ID        string `db:"id"`
	RuntimeID string `db:"runtime_id"`
	Key       string `db:"key"`
	Value     string `db:"value"`
}

// GetID returns ID of RuntimeContext
func (e *RuntimeContext) GetID() string {
	return e.ID
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *RuntimeContext) GetParent() (resource.Type, string) {
	return resource.Runtime, e.RuntimeID
}

// DecorateWithTenantID decorates the entity with the given tenant ID.
func (e *RuntimeContext) DecorateWithTenantID(tenant string) interface{} {
	return struct {
		*RuntimeContext
		TenantID string `db:"tenant_id"`
	}{
		RuntimeContext: e,
		TenantID:       tenant,
	}
}
