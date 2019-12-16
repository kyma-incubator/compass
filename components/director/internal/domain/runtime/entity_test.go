package runtime_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntity_EntityFromRuntimeModel_RuntimeWithDescription(t *testing.T) {
	// given
	description := "Description for Runtime XYZ"
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	modelRuntime := model.Runtime{
		ID:          uuid.New().String(),
		Tenant:      uuid.New().String(),
		Name:        "Runtime XYZ",
		Description: &description,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusConditionInitial,
			Timestamp: time,
		},
		CreationTimestamp: time,
	}

	// when
	entityRuntime, err := runtime.EntityFromRuntimeModel(&modelRuntime)

	// then
	require.NoError(t, err)
	assert.Equal(t, modelRuntime.ID, entityRuntime.ID)
	assert.Equal(t, modelRuntime.Tenant, entityRuntime.TenantID)
	assert.Equal(t, modelRuntime.Name, entityRuntime.Name)
	assert.True(t, entityRuntime.Description.Valid)
	assert.Equal(t, *modelRuntime.Description, entityRuntime.Description.String)
	assert.Equal(t, "INITIAL", entityRuntime.StatusCondition)
	assert.Equal(t, modelRuntime.Status.Timestamp, entityRuntime.StatusTimestamp)
	assert.Equal(t, modelRuntime.CreationTimestamp, entityRuntime.CreationTimestamp)
}

func TestEntity_EntityFromRuntimeModel_RuntimeWithoutDescription(t *testing.T) {
	// given
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	modelRuntime := model.Runtime{
		ID:          uuid.New().String(),
		Tenant:      uuid.New().String(),
		Name:        "Runtime ABC",
		Description: nil,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusConditionInitial,
			Timestamp: time,
		},
		CreationTimestamp: time,
	}

	// when
	entityRuntime, err := runtime.EntityFromRuntimeModel(&modelRuntime)

	// then
	require.NoError(t, err)
	assert.Equal(t, modelRuntime.ID, entityRuntime.ID)
	assert.Equal(t, modelRuntime.Tenant, entityRuntime.TenantID)
	assert.Equal(t, modelRuntime.Name, entityRuntime.Name)
	assert.False(t, entityRuntime.Description.Valid)
	assert.Equal(t, "INITIAL", entityRuntime.StatusCondition)
	assert.Equal(t, modelRuntime.Status.Timestamp, entityRuntime.StatusTimestamp)
	assert.Equal(t, modelRuntime.CreationTimestamp, entityRuntime.CreationTimestamp)
}

func TestEntity_RuntimeToModel_RuntimeWithDescription(t *testing.T) {
	// given
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	description := sql.NullString{
		Valid:  true,
		String: "Description for runtime QWE",
	}

	entityRuntime := runtime.Runtime{
		ID:                uuid.New().String(),
		TenantID:          uuid.New().String(),
		Name:              "Runtime QWE",
		Description:       description,
		StatusCondition:   "INITIAL",
		StatusTimestamp:   time,
		CreationTimestamp: time,
	}

	// when
	modelRuntime, err := entityRuntime.ToModel()

	// then
	require.NoError(t, err)
	assert.Equal(t, entityRuntime.ID, modelRuntime.ID)
	assert.Equal(t, entityRuntime.TenantID, modelRuntime.Tenant)
	assert.Equal(t, entityRuntime.Name, modelRuntime.Name)
	assert.Equal(t, "Description for runtime QWE", *modelRuntime.Description)
	assert.Equal(t, entityRuntime.StatusCondition, string(modelRuntime.Status.Condition))
	assert.Equal(t, entityRuntime.StatusTimestamp, modelRuntime.Status.Timestamp)
	assert.Equal(t, entityRuntime.CreationTimestamp, modelRuntime.CreationTimestamp)
}

func TestEntity_RuntimeToModel_RuntimeWithoutDescription(t *testing.T) {
	// given
	time, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	description := sql.NullString{
		Valid:  false,
		String: "",
	}

	entityRuntime := runtime.Runtime{
		ID:                uuid.New().String(),
		TenantID:          uuid.New().String(),
		Name:              "Runtime AZE",
		Description:       description,
		StatusCondition:   "INITIAL",
		StatusTimestamp:   time,
		CreationTimestamp: time,
	}

	// when
	modelRuntime, err := entityRuntime.ToModel()

	// then
	require.NoError(t, err)
	assert.Equal(t, entityRuntime.ID, modelRuntime.ID)
	assert.Equal(t, entityRuntime.TenantID, modelRuntime.Tenant)
	assert.Equal(t, entityRuntime.Name, modelRuntime.Name)
	assert.Nil(t, modelRuntime.Description)
	assert.Equal(t, entityRuntime.StatusCondition, string(modelRuntime.Status.Condition))
	assert.Equal(t, entityRuntime.StatusTimestamp, modelRuntime.Status.Timestamp)
	assert.Equal(t, entityRuntime.CreationTimestamp, modelRuntime.CreationTimestamp)
}
