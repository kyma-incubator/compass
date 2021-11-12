package runtimectx_test

import (
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"regexp"
	"strconv"
	"testing"
)
/*
func TestPgRepository_GetByID_ShouldReturnRuntimeContextModelForRuntimeContextEntity(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()
	runtimeContextID := uuid.New().String()

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "runtime_id", "tenant_id", "key", "value"}).
		AddRow(runtimeContextID, runtimeID, tenantID, "key", "val")

	sqlMock.ExpectQuery(fmt.Sprintf(`^SELECT (.+) FROM public.runtime_contexts WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())).
		WithArgs(tenantID, runtimeContextID).
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtimectx.NewRepository()

	// when
	modelRuntimeCtx, err := pgRepository.GetByID(ctx, tenantID, runtimeContextID)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeContextID, modelRuntimeCtx.ID)
	assert.Equal(t, runtimeID, modelRuntimeCtx.RuntimeID)
	assert.Equal(t, tenantID, modelRuntimeCtx.Tenant)
}

func TestPgRepository_GetByFiltersAndID_WithoutAdditionalFiltersShouldReturnRuntimeContextModelForRuntimeContextEntity(t *testing.T) {
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()
	runtimeContextID := uuid.New().String()
	// given
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "runtime_id", "tenant_id", "key", "value"}).
		AddRow(runtimeContextID, runtimeID, tenantID, "key", "val")

	sqlMock.ExpectQuery(fmt.Sprintf(`^SELECT (.+) FROM public.runtime_contexts WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())).
		WithArgs(tenantID, runtimeContextID).
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtimectx.NewRepository()

	// when
	modelRuntimeCtx, err := pgRepository.GetByFiltersAndID(ctx, tenantID, runtimeContextID, nil)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeContextID, modelRuntimeCtx.ID)
	assert.Equal(t, runtimeID, modelRuntimeCtx.RuntimeID)
	assert.Equal(t, tenantID, modelRuntimeCtx.Tenant)
}

func TestPgRepository_GetByFiltersAndID_WithAdditionalFiltersShouldReturnRuntimeContextModelForRuntimeContextEntity(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()
	runtimeContextID := uuid.New().String()

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "runtime_id", "tenant_id", "key", "value"}).
		AddRow(runtimeContextID, runtimeID, tenantID, "key", "val")

	sqlMock.ExpectQuery(fmt.Sprintf(`^SELECT (.+) FROM public.runtime_contexts WHERE %s AND id = \$2 AND id IN \(SELECT "runtime_context_id" FROM public\.labels WHERE "runtime_context_id" IS NOT NULL AND %s AND "key" = \$4\)$`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(3))).
		WithArgs(tenantID, runtimeContextID, tenantID, "someKey").
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtimectx.NewRepository()

	// when
	filters := []*labelfilter.LabelFilter{labelfilter.NewForKey("someKey")}
	modelRuntimeCtx, err := pgRepository.GetByFiltersAndID(ctx, tenantID, runtimeContextID, filters)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeContextID, modelRuntimeCtx.ID)
	assert.Equal(t, runtimeID, modelRuntimeCtx.RuntimeID)
	assert.Equal(t, tenantID, modelRuntimeCtx.Tenant)
}

func TestPgRepository_GetByFiltersGlobal_ShouldReturnRuntimeContextModelForRuntimeContextEntity(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()
	runtimeContextID := uuid.New().String()

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "runtime_id", "tenant_id", "key", "value"}).
		AddRow(runtimeContextID, runtimeID, tenantID, "key", "val")

	sqlMock.ExpectQuery(`^SELECT (.+) FROM public.runtime_contexts WHERE id IN \(SELECT "runtime_context_id" FROM public\.labels WHERE "runtime_context_id" IS NOT NULL AND "key" = \$1\)$`).
		WithArgs("someKey").
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtimectx.NewRepository()

	// when
	filters := []*labelfilter.LabelFilter{labelfilter.NewForKey("someKey")}
	modelRuntimeCtx, err := pgRepository.GetByFiltersGlobal(ctx, filters)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeContextID, modelRuntimeCtx.ID)
	assert.Equal(t, runtimeID, modelRuntimeCtx.RuntimeID)
}

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
				Query:       regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.runtime_contexts SET key = ?, value = ? WHERE id = ? AND (id IN (SELECT id FROM runtime_contexts_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:        []driver.Value{runtimeCtxModel.Key, runtimeCtxModel.Value, runtimeCtxModel.ID},
				ValidResult: sqlmock.NewResult(-1, 1),
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

/*
func TestPgRepository_Delete_ShouldDeleteRuntimeEntityUsingValidModel(t *testing.T) {
	// given
	runtimeCtxID := uuid.New().String()
	tenantID := uuid.New().String()

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	sqlMock.ExpectExec(fmt.Sprintf(`^DELETE FROM public.runtime_contexts WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())).
		WithArgs(tenantID, runtimeCtxID).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtimectx.NewRepository()

	// when
	err := pgRepository.Delete(ctx, tenantID, runtimeCtxID)

	// then
	assert.NoError(t, err)
}
*/

func TestPgRepository_Exist(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name:                "Runtime Context Exists",
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
