package persistence_test

import (
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
			Name:       "Unique violation error",
			Error:      &pq.Error{Code: persistence.UniqueViolation},
			AssertFunc: apperrors.IsNotUniqueError,
		},
		{
			Name:       "Unique violation error",
			Error:      &pq.Error{Code: persistence.ForeignKeyViolation},
			AssertFunc: apperrors.IsNewInvalidDataError,
		},
		{
			Name:       "Not mapper sql error",
			Error:      &pq.Error{Code: "123", Message: "SQL fault"},
			AssertFunc: isInternalServerErr(t, "Internal Server Error: SQL Error occurred"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//WHEN
			err := persistence.MapSQLError(testCase.Error, resource.Application, "testErr")

			//THEN
			require.Error(t, err)
			assert.True(t, testCase.AssertFunc(err))
		})
	}

	t.Run("Error is nil", func(t *testing.T) {
		//WHEN
		err := persistence.MapSQLError(nil, resource.Application, "test: %s", "test")

		//THEN
		require.NoError(t, err)
	})
}

func isInternalServerErr(t *testing.T, expectedErrMsg string) func(err error) bool {
	return func(err error) bool {
		return assert.Equal(t, err.Error(), expectedErrMsg)
	}
}
