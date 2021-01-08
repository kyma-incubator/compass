package runtime_context

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// RuntimeContext struct represents database entity for RuntimeContext
type RuntimeContext struct {
	ID        string `db:"id"`
	RuntimeID string `db:"runtime_id"`
	TenantID  string `db:"tenant_id"`
	Key       string `db:"key"`
	Value     string `db:"value"`
}

// EntityFromRuntimeModel converts RuntimeContext model to RuntimeContext entity
func EntityFromRuntimeContextModel(model *model.RuntimeContext) *RuntimeContext {
	return &RuntimeContext{
		ID:        model.ID,
		RuntimeID: model.RuntimeID,
		TenantID:  model.Tenant,
		Key:       model.Key,
		Value:     model.Value,
	}
}

// GraphQLToModel converts RuntimeContext entity to RuntimeContext model
func (e RuntimeContext) ToModel() *model.RuntimeContext {
	return &model.RuntimeContext{
		ID:        e.ID,
		RuntimeID: e.RuntimeID,
		Tenant:    e.TenantID,
		Key:       e.Key,
		Value:     e.Value,
	}
}
