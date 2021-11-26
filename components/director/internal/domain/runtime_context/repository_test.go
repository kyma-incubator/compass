package runtimectx_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

	"github.com/DATA-DOG/go-sqlmock"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context/automock"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_GetByID(t *testing.T) {
	runtimeCtxModel := fixModelRuntimeCtx()
	runtimeCtxEntity := fixEntityRuntimeCtx()

	suite := testdb.RepoGetTestSuite{
		Name: "Get Runtime Context By ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, runtime_id, key, value FROM public.runtime_contexts WHERE id = $1 AND (id IN (SELECT id FROM runtime_contexts_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{runtimeCtxID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(runtimeCtxModel.ID, runtimeCtxModel.RuntimeID, runtimeCtxModel.Key, runtimeCtxModel.Value)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       runtimectx.NewRepository,
		ExpectedModelEntity:       runtimeCtxModel,
		ExpectedDBEntity:          runtimeCtxEntity,
		MethodArgs:                []interface{}{tenantID, runtimeCtxID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByFiltersAndID(t *testing.T) {
	runtimeCtxModel := fixModelRuntimeCtx()
	runtimeCtxEntity := fixEntityRuntimeCtx()

	suite := testdb.RepoGetTestSuite{
		Name: "Get Runtime Context By Filters and ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, runtime_id, key, value FROM public.runtime_contexts WHERE id = $1
												AND id IN (SELECT "runtime_context_id" FROM public.labels WHERE "runtime_context_id" IS NOT NULL AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $2)) AND "key" = $3 AND "value" ?| array[$4])
												AND (id IN (SELECT id FROM runtime_contexts_tenants WHERE tenant_id = $5))`),
				Args:     []driver.Value{runtimeCtxID, tenantID, model.ScenariosKey, "scenario", tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(runtimeCtxModel.ID, runtimeCtxModel.RuntimeID, runtimeCtxModel.Key, runtimeCtxModel.Value)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       runtimectx.NewRepository,
		ExpectedModelEntity:       runtimeCtxModel,
		ExpectedDBEntity:          runtimeCtxEntity,
		MethodName:                "GetByFiltersAndID",
		MethodArgs:                []interface{}{tenantID, runtimeCtxID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(model.ScenariosKey, `$[*] ? ( @ == "scenario" )`)}},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByFiltersGlobal_ShouldReturnRuntimeContextModelForRuntimeContextEntity(t *testing.T) {
	// given
	runtimeCtxModel := fixModelRuntimeCtx()
	runtimeCtxEntity := fixEntityRuntimeCtx()

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows(fixColumns).AddRow(runtimeCtxModel.ID, runtimeCtxModel.RuntimeID, runtimeCtxModel.Key, runtimeCtxModel.Value)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM public.runtime_contexts WHERE id IN \(SELECT "runtime_context_id" FROM public\.labels WHERE "runtime_context_id" IS NOT NULL AND "key" = \$1\)$`).
		WithArgs("someKey").
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	mockConverter := &automock.EntityConverter{}
	mockConverter.On("FromEntity", runtimeCtxEntity).Return(runtimeCtxModel, nil).Once()
	pgRepository := runtimectx.NewRepository(mockConverter)

	// when
	filters := []*labelfilter.LabelFilter{labelfilter.NewForKey("someKey")}
	modelRuntimeCtx, err := pgRepository.GetByFiltersGlobal(ctx, filters)

	// THEN
	require.NoError(t, err)
	require.Equal(t, runtimeCtxModel, modelRuntimeCtx)
	mockConverter.AssertExpectations(t)
}

func TestPgRepository_List(t *testing.T) {
	runtimeCtx1ID := "aec0e9c5-06da-4625-9f8a-bda17ab8c3b9"
	runtimeCtx2ID := "ccdbef8f-b97a-490c-86e2-2bab2862a6e4"
	runtimeCtxEntity1 := fixEntityRuntimeCtxWithID(runtimeCtx1ID)
	runtimeCtxEntity2 := fixEntityRuntimeCtxWithID(runtimeCtx2ID)

	runtimeCtxModel1 := fixModelRuntimeCtxWithID(runtimeCtx1ID)
	runtimeCtxModel2 := fixModelRuntimeCtxWithID(runtimeCtx2ID)

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Runtime Contexts",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, runtime_id, key, value FROM public.runtime_contexts WHERE runtime_id = $1
												AND id IN (SELECT "runtime_context_id" FROM public.labels WHERE "runtime_context_id" IS NOT NULL AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $2)) AND "key" = $3 AND "value" ?| array[$4])
												AND (id IN (SELECT id FROM runtime_contexts_tenants WHERE tenant_id = $5)) ORDER BY id LIMIT 2 OFFSET 0`),
				Args:     []driver.Value{runtimeID, tenantID, model.ScenariosKey, "scenario", tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(runtimeCtxEntity1.ID, runtimeCtxEntity1.RuntimeID, runtimeCtxEntity1.Key, runtimeCtxEntity1.Value).
						AddRow(runtimeCtxEntity2.ID, runtimeCtxEntity2.RuntimeID, runtimeCtxEntity2.Key, runtimeCtxEntity2.Value),
					}
				},
			},
			{
				Query: regexp.QuoteMeta(`SELECT COUNT(*) FROM public.runtime_contexts WHERE runtime_id = $1
												AND id IN (SELECT "runtime_context_id" FROM public.labels WHERE "runtime_context_id" IS NOT NULL AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $2)) AND "key" = $3 AND "value" ?| array[$4])
												AND (id IN (SELECT id FROM runtime_contexts_tenants WHERE tenant_id = $5))`),
				Args:     []driver.Value{runtimeID, tenantID, model.ScenariosKey, "scenario", tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(2)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{runtimeCtxModel1, runtimeCtxModel2},
				ExpectedDBEntities:    []interface{}{runtimeCtxEntity1, runtimeCtxEntity2},
				ExpectedPage: &model.RuntimeContextPage{
					Data: []*model.RuntimeContext{runtimeCtxModel1, runtimeCtxModel2},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 2,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       runtimectx.NewRepository,
		MethodArgs:                []interface{}{runtimeID, tenantID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(model.ScenariosKey, `$[*] ? ( @ == "scenario" )`)}, 2, ""},
		MethodName:                "List",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Create(t *testing.T) {
	var nilRuntimeCtxModel *model.RuntimeContext
	runtimeCtxModel := fixModelRuntimeCtx()
	runtimeCtxEntity := fixEntityRuntimeCtx()

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Runtime Context",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_runtimes WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, runtimeID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.runtime_contexts \(.+\) VALUES \(.+\)$`,
				Args:        []driver.Value{runtimeCtxModel.ID, runtimeCtxModel.RuntimeID, runtimeCtxModel.Key, runtimeCtxModel.Value},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       runtimectx.NewRepository,
		ModelEntity:               runtimeCtxModel,
		DBEntity:                  runtimeCtxEntity,
		NilModelEntity:            nilRuntimeCtxModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	var nilRuntimeCtxModel *model.RuntimeContext
	runtimeCtxModel := fixModelRuntimeCtx()
	runtimeCtxEntity := fixEntityRuntimeCtx()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Runtime Context",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.runtime_contexts SET key = ?, value = ? WHERE id = ? AND (id IN (SELECT id FROM runtime_contexts_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          []driver.Value{runtimeCtxModel.Key, runtimeCtxModel.Value, runtimeCtxModel.ID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       runtimectx.NewRepository,
		ModelEntity:               runtimeCtxModel,
		DBEntity:                  runtimeCtxEntity,
		NilModelEntity:            nilRuntimeCtxModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Context Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.runtime_contexts WHERE id = $1 AND (id IN (SELECT id FROM runtime_contexts_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{runtimeCtxID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: runtimectx.NewRepository,
		MethodArgs:          []interface{}{tenantID, runtimeCtxID},
	}

	suite.Run(t)
}

func TestPgRepository_Exist(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Runtime Context Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.runtime_contexts WHERE id = $1 AND (id IN (SELECT id FROM runtime_contexts_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{runtimeCtxID, tenantID},
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
		RepoConstructorFunc: runtimectx.NewRepository,
		TargetID:            runtimeCtxID,
		TenantID:            tenantID,
	}

	suite.Run(t)
}
