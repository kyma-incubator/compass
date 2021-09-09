package testdb

import "github.com/DATA-DOG/go-sqlmock"

// RowWhenObjectExist missing godoc
func RowWhenObjectExist() *sqlmock.Rows {
	return sqlmock.NewRows([]string{""}).AddRow("1")
}

// RowCount missing godoc
func RowCount(totalCount int) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
}

// RowWhenObjectDoesNotExist missing godoc
func RowWhenObjectDoesNotExist() *sqlmock.Rows {
	return sqlmock.NewRows([]string{""})
}
