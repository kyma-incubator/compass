package runtime_context_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntity_EntityFromRuntimeModel(t *testing.T) {
	// given
	modelRuntime := model.RuntimeContext{
		ID:        "id",
		RuntimeID: "runtime_id",
		Tenant:    "tenant_id",
		Key:       "key",
		Value:     "value",
	}

	// when
	entityRuntime := runtime_context.EntityFromRuntimeContextModel(&modelRuntime)

	// then
	assert.Equal(t, modelRuntime.ID, entityRuntime.ID)
	assert.Equal(t, modelRuntime.RuntimeID, entityRuntime.RuntimeID)
	assert.Equal(t, modelRuntime.Tenant, entityRuntime.TenantID)
	assert.Equal(t, modelRuntime.Key, entityRuntime.Key)
	assert.Equal(t, modelRuntime.Value, entityRuntime.Value)
}

func TestEntity_RuntimeContextToModel(t *testing.T) {
	// given
	entityRuntime := runtime_context.RuntimeContext{
		ID:        "id",
		RuntimeID: "runtime_id",
		TenantID:  "tenant_id",
		Key:       "key",
		Value:     "value",
	}

	// when
	modelRuntime := entityRuntime.ToModel()

	// then
	assert.Equal(t, entityRuntime.ID, modelRuntime.ID)
	assert.Equal(t, entityRuntime.RuntimeID, modelRuntime.RuntimeID)
	assert.Equal(t, entityRuntime.TenantID, modelRuntime.Tenant)
	assert.Equal(t, entityRuntime.Key, modelRuntime.Key)
	assert.Equal(t, entityRuntime.Value, modelRuntime.Value)
}
