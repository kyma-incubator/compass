package destination_test

import (
	"database/sql"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/destination"
	"github.com/stretchr/testify/require"
)

var converter = destination.NewConverter()

func TestEntityConverter_ToEntity(t *testing.T) {
	destinationModelWithoutAssignmentID := fixDestinationModel(destinationName)
	destinationModelWithoutAssignmentID.FormationAssignmentID = nil

	destinationEntityWithoutAssignmentID := fixDestinationEntity(destinationName)
	destinationEntityWithoutAssignmentID.FormationAssignmentID = sql.NullString{}

	testCases := []struct {
		name           string
		input          *model.Destination
		expectedEntity *destination.Entity
	}{
		{
			name:           "success when all nullable properties are filled",
			input:          destinationModel,
			expectedEntity: destinationEntity,
		},
		{
			name:           "success when all nullable properties are empty",
			input:          destinationModelWithoutAssignmentID,
			expectedEntity: destinationEntityWithoutAssignmentID,
		},
		{
			name:           "returns nil when input is nil",
			input:          nil,
			expectedEntity: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// WHEN
			entity := converter.ToEntity(testCase.input)

			// THEN
			require.Equal(t, testCase.expectedEntity, entity)
		})
	}
}

func TestEntityConverter_FromEntity(t *testing.T) {
	destinationEntityWithoutAssignmentID := fixDestinationEntity(destinationName)
	destinationEntityWithoutAssignmentID.FormationAssignmentID = sql.NullString{}

	destinationModelWithoutAssignmentID := fixDestinationModel(destinationName)
	destinationModelWithoutAssignmentID.FormationAssignmentID = nil

	testCases := []struct {
		name          string
		input         *destination.Entity
		expectedModel *model.Destination
	}{
		{
			name:          "success when all nullable properties are filled",
			input:         destinationEntity,
			expectedModel: destinationModel,
		},
		{
			name:          "success when all nullable properties are empty",
			input:         destinationEntityWithoutAssignmentID,
			expectedModel: destinationModelWithoutAssignmentID,
		},
		{
			name:          "returns nil when input is nil",
			input:         nil,
			expectedModel: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// WHEN
			destModel := converter.FromEntity(testCase.input)

			// THEN
			require.Equal(t, testCase.expectedModel, destModel)
		})
	}
}
