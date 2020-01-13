package testdb

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertSqlNullStringEqualTo(t *testing.T, in sql.NullString, text *string) {
	if text != nil {
		sqlStr := sql.NullString{}
		err := sqlStr.Scan(*text)
		require.NoError(t, err)
		assert.Equal(t, in, sqlStr)
	} else {
		require.False(t, in.Valid)
	}
}

func AssertSqlNullBool(t *testing.T, in sql.NullBool, boolean *bool) {
	if boolean != nil {
		require.True(t, in.Valid)
		assert.Equal(t, *boolean, in.Bool)
	} else {
		assert.False(t, in.Valid)
	}
}
