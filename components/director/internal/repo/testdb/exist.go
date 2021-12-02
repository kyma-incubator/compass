package testdb

import "github.com/DATA-DOG/go-sqlmock"

// RowWhenObjectExist represents a sql row when object exist.
func RowWhenObjectExist() *sqlmock.Rows {
	return sqlmock.NewRows([]string{""}).AddRow("1")
}

// RowCount represents a sql row when count query is executed.
func RowCount(totalCount int) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
}

// RowWhenObjectDoesNotExist represents a sql row when object does not exist.
func RowWhenObjectDoesNotExist() *sqlmock.Rows {
	return sqlmock.NewRows([]string{""})
}
