package tenantparentmapping_test

import (
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
)

var (
	testErr  = errors.New("testError")
	tenantID = "tenant-id"
	parentID = "parent-id"
)

type sqlTenantParentsRow struct {
	tenantID string
	parentID string
}

func fixSQLTenantParentsRows(rows []sqlTenantParentsRow) *sqlmock.Rows {
	out := sqlmock.NewRows([]string{"tenant_id", "parent_id"})
	for _, row := range rows {
		out.AddRow(row.tenantID, row.parentID)
	}
	return out
}
