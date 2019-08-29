package testdb

import "github.com/DATA-DOG/go-sqlmock"

func RowWhenObjectExist() *sqlmock.Rows {
	return sqlmock.NewRows([]string{""}).AddRow("1")

}

func RowCount(totalCount int) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"count"}).AddRow(totalCount)
}

func RowWhenObjectDoesNotExist() *sqlmock.Rows {
	return sqlmock.NewRows([]string{""})
}
