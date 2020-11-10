package runtime_context_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestEntity_EntityFromRuntimeModel(t *testing.T) {
	// given
	modelRuntimeCtx := model.RuntimeContext{
		ID:        "id",
		RuntimeID: "runtime_id",
		Tenant:    "tenant_id",
		Key:       "key",
		Value:     "value",
	}

	// when
	entityRuntimeCtx := runtime_context.EntityFromRuntimeContextModel(&modelRuntimeCtx)

	// then
	assert.Equal(t, modelRuntimeCtx.ID, entityRuntimeCtx.ID)
	assert.Equal(t, modelRuntimeCtx.RuntimeID, entityRuntimeCtx.RuntimeID)
	assert.Equal(t, modelRuntimeCtx.Tenant, entityRuntimeCtx.TenantID)
	assert.Equal(t, modelRuntimeCtx.Key, entityRuntimeCtx.Key)
	assert.Equal(t, modelRuntimeCtx.Value, entityRuntimeCtx.Value)
}

func TestEntity_RuntimeContextToModel(t *testing.T) {
	// given
	entityRuntimeCtx := runtime_context.RuntimeContext{
		ID:        "id",
		RuntimeID: "runtime_id",
		TenantID:  "tenant_id",
		Key:       "key",
		Value:     "value",
	}

	// when
	modelRuntimeCtx := entityRuntimeCtx.ToModel()

	// then
	assert.Equal(t, entityRuntimeCtx.ID, modelRuntimeCtx.ID)
	assert.Equal(t, entityRuntimeCtx.RuntimeID, modelRuntimeCtx.RuntimeID)
	assert.Equal(t, entityRuntimeCtx.TenantID, modelRuntimeCtx.Tenant)
	assert.Equal(t, entityRuntimeCtx.Key, modelRuntimeCtx.Key)
	assert.Equal(t, entityRuntimeCtx.Value, modelRuntimeCtx.Value)
}
