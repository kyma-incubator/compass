package testdb

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// MockDatabase returns a new *sqlx.DB mock alongside with a wrapper providing easier interface for asserting expectations.
func MockDatabase(t *testing.T) (*sqlx.DB, DBMock) {
	sqlDB, sqlMock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(sqlDB, "sqlmock")

	return sqlxDB, &sqlMockWithAssertions{sqlMock}
}

// DBMock represents a wrapper providing easier interface for asserting expectations.
type DBMock interface {
	sqlmock.Sqlmock
	AssertExpectations(t *testing.T)
}

type sqlMockWithAssertions struct {
	sqlmock.Sqlmock
}

// AssertExpectations asserts that all the expectations to the mock were met.
func (s *sqlMockWithAssertions) AssertExpectations(t *testing.T) {
	err := s.ExpectationsWereMet()
	require.NoError(t, err)
}
