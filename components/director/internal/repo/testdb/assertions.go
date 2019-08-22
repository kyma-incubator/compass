package testdb

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertSqlNullString(t *testing.T, in sql.NullString, text *string) {
	if text != nil && len(*text) > 1 {
		sqlStr := sql.NullString{}
		err := sqlStr.Scan(*text)
		require.NoError(t, err)
		assert.Equal(t, in, sqlStr)
	} else {
		require.False(t, in.Valid)
	}
}

func AssertSqlNullBool(t *testing.T, in sql.NullBool, exptected *bool) {
	if exptected != nil {
		require.True(t, in.Valid)
		assert.Equal(t, *exptected, in.Bool)
	} else {
		assert.False(t, in.Valid)
	}
}
