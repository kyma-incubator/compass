package runtimectx

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// RuntimeContext struct represents database entity for RuntimeContext
type RuntimeContext struct {
	ID        string `db:"id"`
	RuntimeID string `db:"runtime_id"`
	Key       string `db:"key"`
	Value     string `db:"value"`
}

func (e *RuntimeContext) GetID() string {
	return e.ID
}

func (e *RuntimeContext) GetParentID() string {
	return e.RuntimeID
}

// EntityFromRuntimeContextModel converts RuntimeContext model to RuntimeContext entity
func EntityFromRuntimeContextModel(model *model.RuntimeContext) *RuntimeContext {
	return &RuntimeContext{
		ID:        model.ID,
		RuntimeID: model.RuntimeID,
		Key:       model.Key,
		Value:     model.Value,
	}
}

// ToModel converts RuntimeContext entity to RuntimeContext model
func (e RuntimeContext) ToModel() *model.RuntimeContext {
	return &model.RuntimeContext{
		ID:        e.ID,
		RuntimeID: e.RuntimeID,
		Key:       e.Key,
		Value:     e.Value,
	}
}
