package persistence_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapSQLError(t *testing.T) {
	//GIVEN
	testCases := []struct {
		Name       string
		Error      error
		AssertFunc func(err error) bool
	}{
		{
			Name:       "SQL NoErrRows error",
			Error:      sql.ErrNoRows,
			AssertFunc: apperrors.IsNotFoundError,
		},
		{
			Name:       "Standard error",
			Error:      errors.New("test error"),
			AssertFunc: isInternalServerErr(t, "Internal Server Error: Unexpected error while executing SQL query"),
		},
		{
			Name:       "Timeout error",
			Error:      context.DeadlineExceeded,
			AssertFunc: isInternalServerErr(t, "Internal Server Error: Maximum processing timeout reached"),
		},
		{
			Name:       "Not null violation",
			Error:      &pq.Error{Code: persistence.NotNullViolation},
			AssertFunc: apperrors.IsNewNotNullViolationError,
		},
		{
			Name:       "Check violation",
			Error:      &pq.Error{Code: persistence.CheckViolation},
			AssertFunc: apperrors.IsNewCheckViolationError,
		},
		{
			Name:       "Unique violation error",
			Error:      &pq.Error{Code: persistence.UniqueViolation},
			AssertFunc: apperrors.IsNotUniqueError,
		},
		{
			Name:       "Foreign key violation error",
			Error:      &pq.Error{Code: persistence.ForeignKeyViolation},
			AssertFunc: apperrors.IsNewInvalidOperationError,
		},
		{
			Name:       "Not mapper sql error",
			Error:      &pq.Error{Code: "123", Message: "SQL fault"},
			AssertFunc: isInternalServerErr(t, "Internal Server Error: Unexpected error while executing SQL query"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			err := persistence.MapSQLError(context.TODO(), testCase.Error, resource.Application, resource.Create, "testErr")

			// THEN
			require.Error(t, err)
			assert.True(t, testCase.AssertFunc(err))
		})

		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			err := persistence.MapSQLError(context.TODO(), testCase.Error, resource.Application, resource.Delete, "testErr")

			// THEN
			require.Error(t, err)
			assert.True(t, testCase.AssertFunc(err))
		})
	}

	t.Run("Error is nil", func(t *testing.T) {
		// WHEN
		err := persistence.MapSQLError(context.TODO(), nil, resource.Application, resource.Create, "test: %s", "test")

		// THEN
		require.NoError(t, err)
	})
}

func isInternalServerErr(t *testing.T, expectedErrMsg string) func(err error) bool {
	return func(err error) bool {
		return assert.Equal(t, err.Error(), expectedErrMsg)
	}
}
