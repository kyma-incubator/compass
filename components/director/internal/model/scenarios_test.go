package model_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/stretchr/testify/require"
)

func Test_SchemaForSchema(t *testing.T) {
	// GIVEN
	validator, err := jsonschema.NewValidatorFromRawSchema(model.SchemaForScenariosSchema)
	require.NoError(t, err)
	// WHEN
	result, err := validator.ValidateRaw(model.NewScenariosSchema([]string{"test-scenario"}))
	// THEN
	require.NoError(t, err)
	require.True(t, result.Valid)
}
