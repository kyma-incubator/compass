package destination_test

import (
	"database/sql"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/destination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("success when all nullable properties are filled", func(t *testing.T) {
		// GIVEN
		destinationName := "test-dest-name"
		destinationModel := fixDestinationModel(destinationName)
		require.NotNil(t, destinationModel)
		conv := destination.NewConverter()

		// WHEN
		entity := conv.ToEntity(destinationModel)

		// THEN
		expectedDestinationEntity := fixDestinationEntity(destinationName)
		assert.Equal(t, expectedDestinationEntity, entity)
	})
	t.Run("success when all nullable properties are empty", func(t *testing.T) {
		// GIVEN
		destinationName := "test-dest-name"
		destinationModel := fixDestinationModel(destinationName)
		destinationModel.FormationAssignmentID = nil
		require.NotNil(t, destinationModel)
		conv := destination.NewConverter()

		// WHEN
		entity := conv.ToEntity(destinationModel)

		// THEN
		expectedDestinationEntity := fixDestinationEntity(destinationName)
		expectedDestinationEntity.FormationAssignmentID = sql.NullString{}
		assert.Equal(t, expectedDestinationEntity, entity)
	})
	t.Run("returns nil when input is nil", func(t *testing.T) {
		// GIVEN
		conv := destination.NewConverter()

		// WHEN
		entity := conv.ToEntity(nil)

		// THEN
		assert.Equal(t, nilEntity, entity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("success when all nullable properties are filled", func(t *testing.T) {
		// GIVEN
		destinationName := "test-dest-name"
		destinationEntity := fixDestinationEntity(destinationName)
		require.NotNil(t, destinationEntity)
		conv := destination.NewConverter()

		// WHEN
		model := conv.FromEntity(destinationEntity)

		// THEN
		expectedDestinationModel := fixDestinationModel(destinationName)
		assert.Equal(t, expectedDestinationModel, model)
	})
	t.Run("success when all nullable properties are empty", func(t *testing.T) {
		// GIVEN
		destinationName := "test-dest-name"
		destinationEntity := fixDestinationEntity(destinationName)
		destinationEntity.FormationAssignmentID = sql.NullString{}
		require.NotNil(t, destinationEntity)
		conv := destination.NewConverter()

		// WHEN
		model := conv.FromEntity(destinationEntity)

		// THEN
		expectedDestinationModel := fixDestinationModel(destinationName)
		expectedDestinationModel.FormationAssignmentID = nil
		assert.Equal(t, expectedDestinationModel, model)
	})
	t.Run("returns nil when input is nil", func(t *testing.T) {
		// GIVEN
		conv := destination.NewConverter()

		// WHEN
		model := conv.FromEntity(nil)

		// THEN
		assert.Equal(t, nilModel, model)
	})
}
