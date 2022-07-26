package tombstone_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	var nilTSModel *model.Tombstone
	tombstoneModel := fixTombstoneModel()
	tombstoneEntity := fixEntityTombstone()
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Tombstone",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, appID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.tombstones \(.+\) VALUES \(.+\)$`,
				Args:        fixTombstoneRow(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       tombstone.NewRepository,
		ModelEntity:               tombstoneModel,
		DBEntity:                  tombstoneEntity,
		NilModelEntity:            nilTSModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	var nilTSModel *model.Tombstone
	entity := fixEntityTombstone()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Tombstone",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.tombstones SET removal_date = ? WHERE id = ? AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          append(fixTombstoneUpdateArgs(), entity.ID, tenantID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       tombstone.NewRepository,
		ModelEntity:               fixTombstoneModel(),
		DBEntity:                  entity,
		NilModelEntity:            nilTSModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Tombstone Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.tombstones WHERE id = $1 AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{tombstoneID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: tombstone.NewRepository,
		MethodArgs:          []interface{}{tenantID, tombstoneID},
	}

	suite.Run(t)
}

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Tombstone Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.tombstones WHERE id = $1 AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{tombstoneID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: tombstone.NewRepository,
		TargetID:            tombstoneID,
		TenantID:            tenantID,
		MethodName:          "Exists",
		MethodArgs:          []interface{}{tenantID, tombstoneID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Tombstone",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, removal_date, id FROM public.tombstones WHERE id = $1 AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{tombstoneID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns()).AddRow(fixTombstoneRow()...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: tombstone.NewRepository,
		ExpectedModelEntity: fixTombstoneModel(),
		ExpectedDBEntity:    fixEntityTombstone(),
		MethodArgs:          []interface{}{tenantID, tombstoneID},
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "List Tombstones",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, removal_date, id FROM public.tombstones WHERE app_id = $1 AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns()).AddRow(fixTombstoneRowWithID("id1")...).AddRow(fixTombstoneRowWithID("id2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   tombstone.NewRepository,
		ExpectedModelEntities: []interface{}{fixTombstoneModelWithID("id1"), fixTombstoneModelWithID("id2")},
		ExpectedDBEntities:    []interface{}{fixEntityTombstoneWithID("id1"), fixEntityTombstoneWithID("id2")},
		MethodArgs:            []interface{}{tenantID, appID},
		MethodName:            "ListByApplicationID",
	}

	suite.Run(t)
}
