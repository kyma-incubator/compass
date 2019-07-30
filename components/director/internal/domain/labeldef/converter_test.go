package labeldef_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTenant = "e9ed370d-4056-45ee-8257-49a64f99771b"

func TestFromGraphQL(t *testing.T) {
	// GIVEN
	sut := labeldef.NewConverter()
	anyString := "anyString"
	var schema interface{}
	schema = anyString
	// WHEN
	actual := sut.FromGraphQL(graphql.LabelDefinitionInput{
		Key:    "some-key",
		Schema: &schema,
	}, testTenant)
	// THEN
	assert.Empty(t, actual.ID)
	assert.Equal(t, testTenant, actual.Tenant)
	require.NotNil(t, actual.Schema)
	assert.Equal(t, anyString, *actual.Schema)
	assert.Equal(t, "some-key", actual.Key)
}

func TestToGraphQL(t *testing.T) {
	// GIVEN
	sut := labeldef.NewConverter()
	// WHEN
	anyString := "anyString"
	var schema interface{}
	schema = anyString
	actual := sut.ToGraphQL(model.LabelDefinition{
		Key:    "some-key",
		Schema: &schema,
	})
	// THEN
	assert.Equal(t, "some-key", actual.Key)
	require.NotNil(t, actual.Schema)
	assert.Equal(t, anyString, *actual.Schema)
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
	assert.Equal(t, `{"$id":"id","title":"title"}`, actual.SchemaJSON)
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

type ExampleSchema struct {
	ID    string `json:"$id"`
	Title string `json:"title"`
}
