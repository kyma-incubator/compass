package tombstone_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	var nilTSModel *model.Tombstone
	tombstoneModel := fixTombstoneModelForApp()
	tombstoneEntity := fixEntityTombstoneForApp()
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
				Args:        fixTombstoneRowForApp(),
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

func TestPgRepository_CreateGlobal(t *testing.T) {
	// GIVEN
	var nilTSModel *model.Tombstone
	tombstoneModel := fixTombstoneModelForAppTemplateVersion()
	tombstoneEntity := fixEntityTombstoneForAppTemplateVersion()
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Tombstone Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.tombstones \(.+\) VALUES \(.+\)$`,
				Args:        fixTombstoneRowForAppTemplateVersion(),
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
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		MethodName:                "CreateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	var nilTSModel *model.Tombstone
	entity := fixEntityTombstoneForApp()

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
		ModelEntity:               fixTombstoneModelForApp(),
		DBEntity:                  entity,
		NilModelEntity:            nilTSModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_UpdateGlobal(t *testing.T) {
	var nilTSModel *model.Tombstone
	entity := fixEntityTombstoneForAppTemplateVersion()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Tombstone Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.tombstones SET removal_date = ? WHERE id = ?`),
				Args:          append(fixTombstoneUpdateArgs(), entity.ID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       tombstone.NewRepository,
		ModelEntity:               fixTombstoneModelForAppTemplateVersion(),
		DBEntity:                  entity,
		NilModelEntity:            nilTSModel,
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		UpdateMethodName:          "UpdateGlobal",
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
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, app_template_version_id, removal_date, id FROM public.tombstones WHERE id = $1 AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{tombstoneID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns()).AddRow(fixTombstoneRowForApp()...)}
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
		ExpectedModelEntity: fixTombstoneModelForApp(),
		ExpectedDBEntity:    fixEntityTombstoneForApp(),
		MethodArgs:          []interface{}{tenantID, tombstoneID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Tombstone Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, app_template_version_id, removal_date, id FROM public.tombstones WHERE id = $1`),
				Args:     []driver.Value{tombstoneID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns()).AddRow(fixTombstoneRowForAppTemplateVersion()...)}
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
		ExpectedModelEntity: fixTombstoneModelForAppTemplateVersion(),
		ExpectedDBEntity:    fixEntityTombstoneForAppTemplateVersion(),
		MethodArgs:          []interface{}{tombstoneID},
		MethodName:          "GetByIDGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_ListByResourceID(t *testing.T) {
	suiteForApp := testdb.RepoListTestSuite{
		Name: "List Tombstones for Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, app_template_version_id, removal_date, id FROM public.tombstones WHERE app_id = $1 AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns()).AddRow(fixTombstoneRowWithIDForApp("id1")...).AddRow(fixTombstoneRowWithIDForApp("id2")...)}
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
		ExpectedModelEntities: []interface{}{fixTombstoneModelWithIDForApp("id1"), fixTombstoneModelWithIDForApp("id2")},
		ExpectedDBEntities:    []interface{}{fixEntityTombstoneWithIDForApp("id1"), fixEntityTombstoneWithIDForApp("id2")},
		MethodArgs:            []interface{}{tenantID, appID, resource.Application},
		MethodName:            "ListByResourceID",
	}

	suiteForAppTemplateVersion := testdb.RepoListTestSuite{
		Name: "List Tombstones for Application Template Version",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, app_template_version_id, removal_date, id FROM public.tombstones WHERE app_template_version_id = $1 FOR UPDATE`),
				Args:     []driver.Value{appTemplateVersionID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns()).AddRow(fixTombstoneRowWithIDForAppTemplateVersion("id1")...).AddRow(fixTombstoneRowWithIDForAppTemplateVersion("id2")...)}
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
		ExpectedModelEntities: []interface{}{fixTombstoneModelWithIDForAppTemplateVersion("id1"), fixTombstoneModelWithIDForAppTemplateVersion("id2")},
		ExpectedDBEntities:    []interface{}{fixEntityTombstoneWithIDForAppTemplateVersion("id1"), fixEntityTombstoneWithIDForAppTemplateVersion("id2")},
		MethodArgs:            []interface{}{tenantID, appTemplateVersionID, resource.ApplicationTemplateVersion},
		MethodName:            "ListByResourceID",
	}

	suiteForApp.Run(t)
	suiteForAppTemplateVersion.Run(t)
}
