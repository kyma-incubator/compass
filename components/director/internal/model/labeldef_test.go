package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/require"
)

func TestValidateLabelDef(t *testing.T) {
	t.Run("valid input when schema not provided", func(t *testing.T) {
		// GIVEN
		in := model.LabelDefinition{ID: "id", Key: "key", Tenant: "tenant"}
		// WHEN
		err := in.Validate()
		// THEN
		require.NoError(t, err)
	})
	t.Run("valid input when correct schema provided", func(t *testing.T) {
		// TODO
		t.SkipNow()
	})

	t.Run("id is required", func(t *testing.T) {
		// GIVEN
		in := model.LabelDefinition{}
		// WHEN
		err := in.Validate()
		// THEN
		require.EqualError(t, err, "missing ID field")
	})

	t.Run("key is required", func(t *testing.T) {
		// GIVEN
		in := model.LabelDefinition{ID: "id", Tenant: "tenant"}
		// WHEN
		err := in.Validate()
		// THEN
		require.EqualError(t, err, "missing Key field")
	})

	t.Run("tenant is required", func(t *testing.T) {
		// GIVEN
		in := model.LabelDefinition{ID: "id", Key: "key"}
		// WHEN
		err := in.Validate()
		// THEN
		require.EqualError(t, err, "missing Tenant field")
	})

	t.Run("invalid schema", func(t *testing.T) {
		// TODO
		t.SkipNow()
		// GIVEN
		var sch interface{} = "anything"
		in := model.LabelDefinition{ID: "id", Key: "key", Tenant: "tenant", Schema: &sch}
		// WHEN
		err := in.Validate()
		// THEN
		require.EqualError(t, err, "xxx")
	})

	t.Run("valid scenarios definition update", func(t *testing.T) {
		// GIVEN
		var schema interface{} = model.ScenariosSchema
		in := model.LabelDefinition{ID: "id", Key: model.ScenariosKey, Tenant: "tenant", Schema: &schema}
		// WHEN
		err := in.Validate()
		// THEN
		require.NoError(t, err)
	})

	t.Run("invalid scenarios definition update", func(t *testing.T) {
		// GIVEN
		var sch interface{} = map[string]interface{}{"test": "test"}
		in := model.LabelDefinition{ID: "id", Key: model.ScenariosKey, Tenant: "tenant", Schema: &sch}
		// WHEN
		err := in.Validate()
		// THEN
		require.Error(t, err)
	})

	t.Run("scenarios definition update when schema is nil", func(t *testing.T) {
		// GIVEN
		in := model.LabelDefinition{ID: "id", Key: model.ScenariosKey, Tenant: "tenant", Schema: nil}
		// WHEN
		err := in.Validate()
		// THEN
		require.Error(t, err)
	})

}

func TestValidateForUpdateLabelDef(t *testing.T) {
	t.Run("valid input when schema not provided", func(t *testing.T) {
		// GIVEN
		in := model.LabelDefinition{Key: "key", Tenant: "tenant"}
		// WHEN
		err := in.ValidateForUpdate()
		// THEN
		require.NoError(t, err)

	})
	t.Run("valid input when correct schema provided", func(t *testing.T) {
		// TODO
		t.SkipNow()

	})

	t.Run("key is required", func(t *testing.T) {
		// GIVEN
		in := model.LabelDefinition{Tenant: "tenant"}
		// WHEN
		err := in.ValidateForUpdate()
		// THEN
		require.EqualError(t, err, "missing Key field")
	})

	t.Run("tenant is required", func(t *testing.T) {
		// GIVEN
		in := model.LabelDefinition{Key: "key"}
		// WHEN
		err := in.ValidateForUpdate()
		// THEN
		require.EqualError(t, err, "missing Tenant field")
	})

	t.Run("invalid schema", func(t *testing.T) {
		// TODO
		t.SkipNow()
		// GIVEN
		var sch interface{} = "anything"
		in := model.LabelDefinition{Key: "key", Tenant: "tenant", Schema: &sch}
		// WHEN
		err := in.ValidateForUpdate()
		// THEN
		require.EqualError(t, err, "xxx")
	})

	t.Run("valid scenarios definition update", func(t *testing.T) {
		// GIVEN
		var schema interface{} = model.ScenariosSchema
		in := model.LabelDefinition{ID: "id", Key: model.ScenariosKey, Tenant: "tenant", Schema: &schema}
		// WHEN
		err := in.ValidateForUpdate()
		// THEN
		require.NoError(t, err)
	})

	t.Run("invalid scenarios definition update", func(t *testing.T) {
		// GIVEN
		var sch interface{} = map[string]interface{}{"test": "test"}
		in := model.LabelDefinition{ID: "id", Key: model.ScenariosKey, Tenant: "tenant", Schema: &sch}
		// WHEN
		err := in.ValidateForUpdate()
		// THEN
		require.Error(t, err)
	})

	t.Run("scenarios definition update when schema is nil", func(t *testing.T) {
		// GIVEN
		in := model.LabelDefinition{ID: "id", Key: model.ScenariosKey, Tenant: "tenant", Schema: nil}
		// WHEN
		err := in.ValidateForUpdate()
		// THEN
		require.Error(t, err)
	})

}
