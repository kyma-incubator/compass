package labeldef_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTenant = "e9ed370d-4056-45ee-8257-49a64f99771b"

func TestFromGraphQL(t *testing.T) {
	t.Run("Correct schema", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewConverter()
		schema := graphql.JSONSchema(`{"schema":"schema"}`)
		expectedSchema := map[string]interface{}{
			"schema": interface{}("schema"),
		}

		// WHEN
		actual, err := sut.FromGraphQL(graphql.LabelDefinitionInput{
			Key:    "some-key",
			Schema: &schema,
		}, testTenant)
		// THEN
		require.NoError(t, err)
		assert.Empty(t, actual.ID)
		assert.Equal(t, testTenant, actual.Tenant)
		require.NotNil(t, actual.Schema)
		assert.Equal(t, expectedSchema, *actual.Schema)
		assert.Equal(t, "some-key", actual.Key)
	})

	t.Run("Error - invalid schema", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewConverter()
		invalidSchema := graphql.JSONSchema(`"schema":`)
		// WHEN
		_, err := sut.FromGraphQL(graphql.LabelDefinitionInput{
			Key:    "some-key",
			Schema: &invalidSchema,
		}, testTenant)
		// THEN
		require.Error(t, err)
	})
}

func TestToGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		sut := labeldef.NewConverter()
		// WHEN
		expectedSchema := graphql.JSONSchema(`{"schema":"schema"}`)
		anySchema := map[string]interface{}{
			"schema": interface{}("schema"),
		}
		var schema interface{}
		schema = anySchema
		actual, err := sut.ToGraphQL(model.LabelDefinition{
			Key:    "some-key",
			Schema: &schema,
		})
		// THEN
		require.NoError(t, err)
		assert.Equal(t, "some-key", actual.Key)
		require.NotNil(t, actual.Schema)
		assert.Equal(t, expectedSchema, *actual.Schema)
	})
}

func TestToEntity(t *testing.T) {
	// GIVEN
	var schema interface{} = ExampleSchema{
		ID:    "id",
		Title: "title",
	}

	sut := labeldef.NewConverter()
	// WHEN
	actual, err := sut.ToEntity(model.LabelDefinition{
		Key:    "some-key",
		Tenant: "tenant",
		ID:     "id",
		Schema: &schema,
	})
	// THEN
	require.NoError(t, err)
	assert.Equal(t, "some-key", actual.Key)
	assert.Equal(t, "tenant", actual.TenantID)
	assert.Equal(t, "id", actual.ID)
	assert.True(t, actual.SchemaJSON.Valid)
	assert.Equal(t, `{"$id":"id","title":"title"}`, actual.SchemaJSON.String)
}

func TestToEntityWhenNoSchema(t *testing.T) {
	// GIVEN
	sut := labeldef.NewConverter()
	// WHEN
	actual, err := sut.ToEntity(model.LabelDefinition{
		Key:    "key",
		Tenant: "tenant",
		ID:     "id",
	})
	// THENr
	require.NoError(t, err)
	assert.Empty(t, actual.SchemaJSON)
}

func TestFromEntityWhenNoSchema(t *testing.T) {
	// GIVEN
	in := labeldef.Entity{
		ID:       "id",
		Key:      "key",
		TenantID: "tenant",
	}
	sut := labeldef.NewConverter()
	// WHEN
	actual, err := sut.FromEntity(in)
	// THEN
	require.NoError(t, err)
	assert.Equal(t, "id", actual.ID)
	assert.Equal(t, "key", actual.Key)
	assert.Equal(t, "tenant", actual.Tenant)
	assert.Nil(t, actual.Schema)
}

func TestFromEntityWhenSchemaProvided(t *testing.T) {
	// GIVEN
	in := labeldef.Entity{
		ID:       "id",
		Key:      "key",
		TenantID: "tenant",
		SchemaJSON: sql.NullString{
			Valid:  true,
			String: `{"$id":"xxx","title":"title"}`,
		},
	}
	sut := labeldef.NewConverter()
	// WHEN
	actual, err := sut.FromEntity(in)
	// THEN
	require.NoError(t, err)
	assert.Equal(t, "id", actual.ID)
	assert.Equal(t, "key", actual.Key)
	assert.Equal(t, "tenant", actual.Tenant)
	assert.NotNil(t, actual.Schema)
	// converting to specific type
	b, err := json.Marshal(actual.Schema)
	require.NoError(t, err)
	var exSchema ExampleSchema
	err = json.Unmarshal(b, &exSchema)
	require.NoError(t, err)
	assert.Equal(t, ExampleSchema{ID: "xxx", Title: "title"}, exSchema)

}

type ExampleSchema struct {
	ID    string `json:"$id"`
	Title string `json:"title"`
}
