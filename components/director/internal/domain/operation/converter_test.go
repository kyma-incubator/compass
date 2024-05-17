package operation_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		// GIVEN
		opModel := fixOperationModel(testOpType, model.OperationStatusScheduled)
		require.NotNil(t, opModel)

		conv := operation.NewConverter()

		// WHEN
		entity := conv.ToEntity(opModel)

		// THEN
		expectedOperation := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled)

		assert.Equal(t, expectedOperation, entity)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		// GIVEN
		opModel := &model.Operation{
			ID:            operationID,
			OpType:        testOpType,
			Status:        model.OperationStatusScheduled,
			Data:          nil,
			Error:         nil,
			ErrorSeverity: model.OperationErrorSeverityNone,
			Priority:      1,
			CreatedAt:     nil,
			UpdatedAt:     nil,
		}

		expectedEntity := &operation.Entity{
			ID:            operationID,
			Type:          string(testOpType),
			Status:        string(model.OperationStatusScheduled),
			Data:          sql.NullString{},
			Error:         sql.NullString{},
			ErrorSeverity: sql.NullString{},
			Priority:      1,
			CreatedAt:     nil,
			UpdatedAt:     nil,
		}
		conv := operation.NewConverter()

		// WHEN
		entity := conv.ToEntity(opModel)

		// THEN
		assert.Equal(t, expectedEntity, entity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		// GIVEN
		entity := fixEntityOperation(operationID, testOpType, model.OperationStatusScheduled)
		conv := operation.NewConverter()

		// WHEN
		opModel := conv.FromEntity(entity)

		// THEN
		expectedOperation := fixOperationModel(testOpType, model.OperationStatusScheduled)
		assert.Equal(t, expectedOperation, opModel)
	})

	t.Run("success all nullable properties empty", func(t *testing.T) {
		// GIVEN
		entity := &operation.Entity{
			ID:            operationID,
			Type:          string(testOpType),
			Status:        string(model.OperationStatusScheduled),
			Data:          sql.NullString{},
			Error:         sql.NullString{},
			ErrorSeverity: sql.NullString{},
			Priority:      1,
			CreatedAt:     nil,
			UpdatedAt:     nil,
		}
		expectedModel := &model.Operation{
			ID:            operationID,
			OpType:        testOpType,
			Status:        model.OperationStatusScheduled,
			Data:          nil,
			Error:         nil,
			ErrorSeverity: model.OperationErrorSeverityNone,
			Priority:      1,
			CreatedAt:     nil,
			UpdatedAt:     nil,
		}
		conv := operation.NewConverter()

		// WHEN
		opModel := conv.FromEntity(entity)

		// THEN
		assert.Equal(t, expectedModel, opModel)
	})
}

func TestConverter_ToGraphQL(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		Name                 string
		Input                *model.Operation
		Expected             *graphql.Operation
		ExpectedErrorMessage string
	}{
		{
			Name:     "Success",
			Input:    fixOperationModelWithIDAndTimestamp("operation-id", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, errorMsg, 1, &now),
			Expected: fixOperationGraphqlWithIDAndTimestamp("operation-id", graphql.ScheduledOperationTypeOrdAggregation, graphql.OperationStatusScheduled, errorMsg, &now),
		},
		{
			Name:                 "Error - invalid operation type",
			Input:                fixOperationModelWithIDAndTimestamp("operation-id", "invalid-op-type", model.OperationStatusScheduled, errorMsg, 1, &now),
			Expected:             &graphql.Operation{},
			ExpectedErrorMessage: "unknown operation type invalid-op-type",
		},
		{
			Name:                 "Error - invalid operation status",
			Input:                fixOperationModelWithIDAndTimestamp("operation-id", model.OperationTypeOrdAggregation, "invalid-op-status", errorMsg, 1, &now),
			Expected:             &graphql.Operation{},
			ExpectedErrorMessage: "unknown operation status invalid-op-status",
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := operation.NewConverter()
			res, err := converter.ToGraphQL(testCase.Input)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.Expected, res)
			}
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	now := time.Now()

	// GIVEN
	input := []*model.Operation{
		fixOperationModelWithIDAndTimestamp("operation-id-1", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, errorMsg, 1, &now),
		fixOperationModelWithIDAndTimestamp("operation-id-2", model.OperationTypeOrdAggregation, model.OperationStatusScheduled, errorMsg, 1, &now),
		nil,
	}
	expected := []*graphql.Operation{
		fixOperationGraphqlWithIDAndTimestamp("operation-id-1", graphql.ScheduledOperationTypeOrdAggregation, graphql.OperationStatusScheduled, errorMsg, &now),
		fixOperationGraphqlWithIDAndTimestamp("operation-id-2", graphql.ScheduledOperationTypeOrdAggregation, graphql.OperationStatusScheduled, errorMsg, &now),
	}

	// WHEN
	converter := operation.NewConverter()
	res, _ := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}
