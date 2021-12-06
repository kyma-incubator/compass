package testdb

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertSQLNullStringEqualTo is a helper to check if the sql.NullString is equal to the given string.
func AssertSQLNullStringEqualTo(t *testing.T, in sql.NullString, text *string) {
	if text != nil {
		sqlStr := sql.NullString{}
		err := sqlStr.Scan(*text)
		require.NoError(t, err)
		assert.Equal(t, in, sqlStr)
	} else {
		require.False(t, in.Valid)
	}
}

// AssertSQLNullBool is a helper to check if the sql.NullBool is equal to the given bool.
func AssertSQLNullBool(t *testing.T, in sql.NullBool, boolean *bool) {
	if boolean != nil {
		require.True(t, in.Valid)
		assert.Equal(t, *boolean, in.Bool)
	} else {
		assert.False(t, in.Valid)
	}
}
