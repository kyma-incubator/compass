package runtimectx_test

import (
	"context"
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"testing"

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
		SqlQueryDetails: []testdb.SqlQueryDetails{
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
		SqlQueryDetails: []testdb.SqlQueryDetails{
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

	//then
	require.NoError(t, err)
	require.Equal(t, runtimeCtxModel, modelRuntimeCtx)
	mockConverter.AssertExpectations(t)
}

/*
func TestPgRepository_List(t *testing.T) {
	//GIVEN
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()
	runtimeCtx1ID := uuid.New().String()
	runtimeCtx2ID := uuid.New().String()

	limit := 2
	offset := 3

	pageableQuery := `^SELECT (.+) FROM public.runtime_contexts WHERE %s AND runtime_id = \$2 ORDER BY id LIMIT %d OFFSET %d$`
	countQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT COUNT(*) FROM public.runtime_contexts WHERE %s AND runtime_id = $2`, fixUnescapedTenantIsolationSubquery()))

	testCases := []struct {
		Name           string
		InputCursor    string
		InputPageSize  int
		ExpectedOffset int
		ExpectedLimit  int
		Rows           *sqlmock.Rows
		TotalCount     int
	}{
		{
			Name:           "Success getting first page",
			InputPageSize:  2,
			InputCursor:    "",
			ExpectedOffset: 0,
			ExpectedLimit:  limit,
			Rows: sqlmock.NewRows([]string{"id", "runtime_id", "tenant_id", "key", "value"}).
				AddRow(runtimeCtx1ID, runtimeID, tenantID, "key", "val").
				AddRow(runtimeCtx2ID, runtimeID, tenantID, "key", "val"),
			TotalCount: 2,
		},
		{
			Name:           "Success getting next page",
			InputPageSize:  2,
			InputCursor:    convertIntToBase64String(offset),
			ExpectedOffset: offset,
			ExpectedLimit:  limit,
			Rows: sqlmock.NewRows([]string{"id", "runtime_id", "tenant_id", "key", "value"}).
				AddRow(runtimeCtx1ID, runtimeID, tenantID, "key", "val").
				AddRow(runtimeCtx2ID, runtimeID, tenantID, "key", "val"),
			TotalCount: 2,
		}}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sqlxDB, sqlMock := testdb.MockDatabase(t)
			defer sqlMock.AssertExpectations(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
			pgRepository := runtimectx.NewRepository()
			expectedQuery := fmt.Sprintf(pageableQuery, fixTenantIsolationSubquery(), testCase.ExpectedLimit, testCase.ExpectedOffset)

			sqlMock.ExpectQuery(expectedQuery).
				WithArgs(tenantID, runtimeID).
				WillReturnRows(testCase.Rows)
			countRow := sqlMock.NewRows([]string{"count"}).AddRow(testCase.TotalCount)

			sqlMock.ExpectQuery(countQuery).
				WithArgs(tenantID, runtimeID).
				WillReturnRows(countRow)

			//THEN
			modelRuntimePage, err := pgRepository.List(ctx, runtimeID, tenantID, nil, testCase.InputPageSize, testCase.InputCursor)

			//THEN
			require.NoError(t, err)
			assert.Equal(t, testCase.ExpectedLimit, modelRuntimePage.TotalCount)
			require.NoError(t, sqlMock.ExpectationsWereMet())

			assert.Equal(t, runtimeCtx1ID, modelRuntimePage.Data[0].ID)
			assert.Equal(t, runtimeID, modelRuntimePage.Data[0].RuntimeID)
			assert.Equal(t, tenantID, modelRuntimePage.Data[0].Tenant)

			assert.Equal(t, runtimeCtx2ID, modelRuntimePage.Data[1].ID)
			assert.Equal(t, runtimeID, modelRuntimePage.Data[1].RuntimeID)
			assert.Equal(t, tenantID, modelRuntimePage.Data[1].Tenant)
		})
	}

	t.Run("Returns error when decoded cursor is non-positive number", func(t *testing.T) {
		//GIVEN
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("MultipleFromEntities", mock.Anything)

		pgRepository := runtime.NewRepository(mockConverter)
		//THEN
		_, err := pgRepository.List(ctx, tenantID, nil, 2, convertIntToBase64String(-3))

		//THEN
		require.EqualError(t, err, "while decoding page cursor: Invalid data [reason=cursor is not correct]")
	})
}

func TestPgRepository_List_WithFiltersShouldReturnRuntimeModelsForRuntimeEntities(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()
	runtimeCtx1ID := uuid.New().String()
	runtimeCtx2ID := uuid.New().String()

	rowSize := 2

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "runtime_id", "tenant_id", "key", "value"}).
		AddRow(runtimeCtx1ID, runtimeID, tenantID, "key", "val").
		AddRow(runtimeCtx2ID, runtimeID, tenantID, "key", "val")

	filterQuery := fmt.Sprintf(`  AND id IN
						\(SELECT "runtime_context_id" FROM public.labels
							WHERE "runtime_context_id" IS NOT NULL
							AND %s
							AND "key" = \$4\)`, fixTenantIsolationSubqueryWithArg(3))
	sqlQuery := fmt.Sprintf(`^SELECT (.+) FROM public.runtime_contexts
								WHERE %s AND runtime_id = \$2 %s ORDER BY id LIMIT %d OFFSET 0`, fixTenantIsolationSubqueryWithArg(1), filterQuery, rowSize)

	sqlMock.ExpectQuery(sqlQuery).
		WithArgs(tenantID, runtimeID, tenantID, "foo").
		WillReturnRows(rows)

	countRows := sqlMock.NewRows([]string{"count"}).AddRow(rowSize)

	countQuery := fmt.Sprintf(`^SELECT COUNT\(\*\) FROM public.runtime_contexts WHERE %s AND runtime_id = \$2 %s`, fixTenantIsolationSubquery(), filterQuery)
	sqlMock.ExpectQuery(countQuery).
		WithArgs(tenantID, runtimeID, tenantID, "foo").
		WillReturnRows(countRows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	labelFilterFoo := labelfilter.LabelFilter{
		Key: "foo",
	}
	filter := []*labelfilter.LabelFilter{&labelFilterFoo}

	pgRepository := runtimectx.NewRepository()

	// when
	modelRuntimePage, err := pgRepository.List(ctx, runtimeID, tenantID, filter, rowSize, "")

	//then
	assert.NoError(t, err)
	require.NotNil(t, modelRuntimePage)
	assert.Equal(t, rowSize, modelRuntimePage.TotalCount)
	require.NoError(t, sqlMock.ExpectationsWereMet())

	assert.Equal(t, runtimeCtx1ID, modelRuntimePage.Data[0].ID)
	assert.Equal(t, runtimeID, modelRuntimePage.Data[0].RuntimeID)
	assert.Equal(t, tenantID, modelRuntimePage.Data[0].Tenant)

	assert.Equal(t, runtimeCtx2ID, modelRuntimePage.Data[1].ID)
	assert.Equal(t, runtimeID, modelRuntimePage.Data[1].RuntimeID)
	assert.Equal(t, tenantID, modelRuntimePage.Data[1].Tenant)
}
*/
func TestPgRepository_Create(t *testing.T) {
	var nilRuntimeCtxModel *model.RuntimeContext
	runtimeCtxModel := fixModelRuntimeCtx()
	runtimeCtxEntity := fixEntityRuntimeCtx()

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Runtime Context",
		SqlQueryDetails: []testdb.SqlQueryDetails{
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
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.runtime_contexts SET key = ?, value = ? WHERE id = ? AND (id IN (SELECT id FROM runtime_contexts_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:          []driver.Value{runtimeCtxModel.Key, runtimeCtxModel.Value, runtimeCtxModel.ID},
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
		SqlQueryDetails: []testdb.SqlQueryDetails{
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
		SqlQueryDetails: []testdb.SqlQueryDetails{
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

func convertIntToBase64String(number int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(number)))
}
