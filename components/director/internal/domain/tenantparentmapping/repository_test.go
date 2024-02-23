package tenantparentmapping_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantparentmapping"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Upsert(t *testing.T) {
	testCases := []struct {
		Name                 string
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		InputTenantID        string
		InputParentID        string
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs([]driver.Value{tenantID, parentID}...).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return db, dbMock
			},
			InputTenantID:        tenantID,
			InputParentID:        parentID,
			ExpectedErrorMessage: "",
		},
		{
			Name: "Error while upserting",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs([]driver.Value{tenantID, parentID}...).
					WillReturnError(testErr)
				return db, dbMock
			},
			InputTenantID:        tenantID,
			InputParentID:        parentID,
			ExpectedErrorMessage: "Unexpected error while executing SQL query",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantParentMappingRepo := tenantparentmapping.NewRepository()

			err := tenantParentMappingRepo.Upsert(ctx, testCase.InputTenantID, testCase.InputParentID)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
			}

			dbMock.AssertExpectations(t)
		})
	}
}

func TestPgRepository_UpsertMultiple(t *testing.T) {
	parentIDs := []string{"parent-id-1", "parent-id-2"}

	testCases := []struct {
		Name                 string
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		InputTenantID        string
		InputParentIDs       []string
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs([]driver.Value{tenantID, parentIDs[0]}...).
					WillReturnResult(sqlmock.NewResult(1, 1))
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs([]driver.Value{tenantID, parentIDs[1]}...).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return db, dbMock
			},
			InputTenantID:        tenantID,
			InputParentIDs:       parentIDs,
			ExpectedErrorMessage: "",
		},
		{
			Name: "Error while upserting multiple",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_parents ( tenant_id, parent_id ) VALUES ( ?, ? ) ON CONFLICT ( tenant_id, parent_id ) DO NOTHING`)).
					WithArgs([]driver.Value{tenantID, parentIDs[0]}...).
					WillReturnError(testErr)
				return db, dbMock
			},
			InputTenantID:        tenantID,
			InputParentIDs:       parentIDs,
			ExpectedErrorMessage: "Unexpected error while executing SQL query",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantParentMappingRepo := tenantparentmapping.NewRepository()

			err := tenantParentMappingRepo.UpsertMultiple(ctx, testCase.InputTenantID, testCase.InputParentIDs)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
			}

			dbMock.AssertExpectations(t)
		})
	}
}

func TestPgRepository_Delete(t *testing.T) {
	testCases := []struct {
		Name                 string
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		InputTenantID        string
		InputParentID        string
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_parents WHERE tenant_id = $1 AND parent_id = $2`)).
					WithArgs([]driver.Value{tenantID, parentID}...).
					WillReturnResult(sqlmock.NewResult(1, 1))
				return db, dbMock
			},
			InputTenantID:        tenantID,
			InputParentID:        parentID,
			ExpectedErrorMessage: "",
		},
		{
			Name: "Error while deleting",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM tenant_parents WHERE tenant_id = $1 AND parent_id = $2`)).
					WithArgs([]driver.Value{tenantID, parentID}...).
					WillReturnError(testErr)
				return db, dbMock
			},
			InputTenantID:        tenantID,
			InputParentID:        parentID,
			ExpectedErrorMessage: "Unexpected error while executing SQL query",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantParentMappingRepo := tenantparentmapping.NewRepository()

			err := tenantParentMappingRepo.Delete(ctx, testCase.InputTenantID, testCase.InputParentID)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
			}

			dbMock.AssertExpectations(t)
		})
	}
}

func TestPgRepository_ListParents(t *testing.T) {
	testCases := []struct {
		Name                 string
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		InputTenantID        string
		ExpectedParentIDs    []string
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs([]driver.Value{tenantID}...).
					WillReturnRows(fixSQLTenantParentsRows([]sqlTenantParentsRow{
						{tenantID: tenantID, parentID: parentID},
					}))
				return db, dbMock
			},
			InputTenantID:        tenantID,
			ExpectedParentIDs:    []string{parentID},
			ExpectedErrorMessage: "",
		},
		{
			Name: "Error while listing parents",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE tenant_id = $1`)).
					WithArgs([]driver.Value{tenantID}...).
					WillReturnError(testErr)
				return db, dbMock
			},
			InputTenantID:        tenantID,
			ExpectedErrorMessage: "Unexpected error while executing SQL query",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantParentMappingRepo := tenantparentmapping.NewRepository()

			parentIDs, err := tenantParentMappingRepo.ListParents(ctx, testCase.InputTenantID)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedParentIDs, parentIDs)
			}

			dbMock.AssertExpectations(t)
		})
	}
}

func TestPgRepository_ListByParent(t *testing.T) {
	testCases := []struct {
		Name                 string
		DBFN                 func(t *testing.T) (*sqlx.DB, testdb.DBMock)
		InputParentTenantID  string
		ExpectedTenantIDs    []string
		ExpectedErrorMessage string
	}{
		{
			Name: "Success",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)

				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs([]driver.Value{parentID}...).
					WillReturnRows(fixSQLTenantParentsRows([]sqlTenantParentsRow{
						{tenantID: tenantID, parentID: parentID},
					}))
				return db, dbMock
			},
			InputParentTenantID:  parentID,
			ExpectedTenantIDs:    []string{tenantID},
			ExpectedErrorMessage: "",
		},
		{
			Name: "Error while listing by parents",
			DBFN: func(t *testing.T) (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT tenant_id, parent_id FROM tenant_parents WHERE parent_id = $1`)).
					WithArgs([]driver.Value{parentID}...).
					WillReturnError(testErr)
				return db, dbMock
			},
			InputParentTenantID:  parentID,
			ExpectedErrorMessage: "Unexpected error while executing SQL query",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db, dbMock := testdb.MockDatabase(t)
			if testCase.DBFN != nil {
				db, dbMock = testCase.DBFN(t)
			}

			ctx := persistence.SaveToContext(context.TODO(), db)
			tenantParentMappingRepo := tenantparentmapping.NewRepository()

			parentIDs, err := tenantParentMappingRepo.ListByParent(ctx, testCase.InputParentTenantID)

			if testCase.ExpectedErrorMessage != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMessage)
			} else {
				require.NoError(t, err)
				require.Equal(t, testCase.ExpectedTenantIDs, parentIDs)
			}

			dbMock.AssertExpectations(t)
		})
	}
}
