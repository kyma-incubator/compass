package runtime_test

import (
	"context"
	"database/sql/driver"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_GetByID(t *testing.T) {
	rtModel := fixDetailedModelRuntime(t, "foo", "Foo", "Lorem ipsum")
	rtEntity := fixDetailedEntityRuntime(t, "foo", "Foo", "Lorem ipsum")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Runtime By ID",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, name, description, status_condition, status_timestamp, creation_timestamp FROM public.runtimes WHERE id = $1 AND (id IN (SELECT id FROM tenant_runtimes WHERE tenant_id = $2))`),
				Args:     []driver.Value{runtimeID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(rtModel.ID, rtModel.Name, rtModel.Description, rtModel.Status.Condition, rtModel.Status.Timestamp, rtModel.CreationTimestamp)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       runtime.NewRepository,
		ExpectedModelEntity:       rtModel,
		ExpectedDBEntity:          rtEntity,
		MethodArgs:                []interface{}{tenantID, runtimeID},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByFiltersAndID(t *testing.T) {
	rtModel := fixDetailedModelRuntime(t, "foo", "Foo", "Lorem ipsum")
	rtEntity := fixDetailedEntityRuntime(t, "foo", "Foo", "Lorem ipsum")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Runtime By Filters and ID",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, name, description, status_condition, status_timestamp, creation_timestamp FROM public.runtimes WHERE id = $1 
												AND id IN (SELECT "runtime_id" FROM public.labels WHERE "runtime_id" IS NOT NULL AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $2)) AND "key" = $3 AND "value" ?| array[$4]) 
												AND (id IN (SELECT id FROM tenant_runtimes WHERE tenant_id = $5))`),
				Args:     []driver.Value{runtimeID, tenantID, model.ScenariosKey, "scenario", tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(rtModel.ID, rtModel.Name, rtModel.Description, rtModel.Status.Condition, rtModel.Status.Timestamp, rtModel.CreationTimestamp)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       runtime.NewRepository,
		ExpectedModelEntity:       rtModel,
		ExpectedDBEntity:          rtEntity,
		MethodName:                "GetByFiltersAndID",
		MethodArgs:                []interface{}{tenantID, runtimeID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(model.ScenariosKey, `$[*] ? ( @ == "scenario" )`)}},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByFiltersGlobal_ShouldReturnRuntimeModelForRuntimeEntity(t *testing.T) {
	// given
	rtModel := fixDetailedModelRuntime(t, "foo", "Foo", "Lorem ipsum")
	rtEntity := fixDetailedEntityRuntime(t, "foo", "Foo", "Lorem ipsum")

	mockConverter := &automock.EntityConverter{}
	mockConverter.On("FromEntity", rtEntity).Return(rtModel, nil).Once()

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
		AddRow(rtModel.ID, rtModel.Name, rtModel.Description, rtModel.Status.Condition, rtModel.Status.Timestamp, rtModel.CreationTimestamp)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM public.runtimes WHERE id IN \(SELECT "runtime_id" FROM public\.labels WHERE "runtime_id" IS NOT NULL AND "key" = \$1\)$`).
		WithArgs("someKey").
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	filters := []*labelfilter.LabelFilter{labelfilter.NewForKey("someKey")}
	modelRuntime, err := pgRepository.GetByFiltersGlobal(ctx, filters)

	//then
	require.NoError(t, err)
	require.Equal(t, rtModel, modelRuntime)
	mockConverter.AssertExpectations(t)
}

func TestPgRepository_GetOldestForFilters(t *testing.T) {
	rtModel := fixDetailedModelRuntime(t, "foo", "Foo", "Lorem ipsum")
	rtEntity := fixDetailedEntityRuntime(t, "foo", "Foo", "Lorem ipsum")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Oldest Runtime By Filters",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, name, description, status_condition, status_timestamp, creation_timestamp FROM public.runtimes WHERE  
												id IN (SELECT "runtime_id" FROM public.labels WHERE "runtime_id" IS NOT NULL AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $1)) AND "key" = $2 AND "value" ?| array[$3]) 
												AND (id IN (SELECT id FROM tenant_runtimes WHERE tenant_id = $4)) ORDER BY creation_timestamp ASC`),
				Args:     []driver.Value{tenantID, model.ScenariosKey, "scenario", tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(rtModel.ID, rtModel.Name, rtModel.Description, rtModel.Status.Condition, rtModel.Status.Timestamp, rtModel.CreationTimestamp)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       runtime.NewRepository,
		ExpectedModelEntity:       rtModel,
		ExpectedDBEntity:          rtEntity,
		MethodName:                "GetOldestForFilters",
		MethodArgs:                []interface{}{tenantID, []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(model.ScenariosKey, `$[*] ? ( @ == "scenario" )`)}},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

/*
func TestPgRepository_ListByFiltersGlobal(t *testing.T) {
	// GIVEN
	runtime1ID := uuid.New().String()
	runtime2ID := uuid.New().String()
	tenantID := uuid.New().String()

	runtimes := []*model.Runtime{
		fixModelRuntime(t, runtime1ID, tenantID, "Runtime ABC", "Description for runtime ABC"),
		fixModelRuntime(t, runtime2ID, tenantID, "Runtime XYZ", "Description for runtime XYZ"),
	}

	mockConverter := &automock.EntityConverter{}
	mockConverter.On("MultipleFromEntities", mock.Anything).Return(runtimes)

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
		AddRow(runtime1ID, tenantID, runtimes[0].Name, runtimes[0].Description, runtimes[0].Status.Condition, runtimes[0].CreationTimestamp, runtimes[0].CreationTimestamp).
		AddRow(runtime2ID, tenantID, runtimes[1].Name, runtimes[1].Description, runtimes[1].Status.Condition, runtimes[1].CreationTimestamp, runtimes[1].CreationTimestamp)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM public.runtimes WHERE id IN \(SELECT "runtime_id" FROM public\.labels WHERE "runtime_id" IS NOT NULL AND "key" = \$1 AND "value" \@\> \$2\ INTERSECT SELECT "runtime_id" FROM public\.labels WHERE "runtime_id" IS NOT NULL AND "key" = \$3 AND "value" \@\> \$4\)$`).
		WithArgs("someKey", "someValue", "someKey2", "someValue2").
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	filters := []*labelfilter.LabelFilter{
		{
			Key:   "someKey",
			Query: str.Ptr(`someValue`),
		},
		{
			Key:   "someKey2",
			Query: str.Ptr(`someValue2`),
		},
	}
	// WHEN
	modelRuntimes, err := pgRepository.ListByFiltersGlobal(ctx, filters)

	// THEN
	assert.NoError(t, err)
	require.NotNil(t, modelRuntimes)
	require.NoError(t, sqlMock.ExpectationsWereMet())

	assert.Equal(t, runtimes, modelRuntimes)
}

func TestPgRepository_List(t *testing.T) {
	//GIVEN
	tenantID := uuid.New().String()
	runtime1ID := uuid.New().String()
	runtime2ID := uuid.New().String()

	runtimes := []*model.Runtime{
		fixModelRuntime(t, runtime1ID, tenantID, "Runtime ABC", "Description for runtime ABC"),
		fixModelRuntime(t, runtime2ID, tenantID, "Runtime XYZ", "Description for runtime XYZ"),
	}

	limit := 2
	offset := 3

	pageableQuery := `^SELECT (.+) FROM public.runtimes WHERE %s ORDER BY name LIMIT %d OFFSET %d$`
	countQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT COUNT(*) FROM public.runtimes WHERE %s`, fixUnescapedTenantIsolationSubquery()))

	testCases := []struct {
		Name           string
		InputCursor    string
		InputPageSize  int
		ExpectedOffset int
		ExpectedLimit  int
		Rows           *sqlmock.Rows
		TotalCount     int
		Runtimes       []*model.Runtime
	}{
		{
			Name:           "Success getting first page",
			InputPageSize:  2,
			InputCursor:    "",
			ExpectedOffset: 0,
			ExpectedLimit:  limit,
			Rows: sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
				AddRow(runtime1ID, tenantID, runtimes[0].Name, runtimes[0].Description, runtimes[0].Status.Condition, runtimes[0].CreationTimestamp, runtimes[0].CreationTimestamp).
				AddRow(runtime2ID, tenantID, runtimes[1].Name, runtimes[1].Description, runtimes[1].Status.Condition, runtimes[1].CreationTimestamp, runtimes[1].CreationTimestamp),
			TotalCount: 2,
			Runtimes:   runtimes,
		},
		{
			Name:           "Success getting next page",
			InputPageSize:  2,
			InputCursor:    convertIntToBase64String(offset),
			ExpectedOffset: offset,
			ExpectedLimit:  limit,
			Rows: sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
				AddRow(runtime1ID, tenantID, runtimes[0].Name, runtimes[0].Description, runtimes[0].Status.Condition, runtimes[0].CreationTimestamp, runtimes[0].CreationTimestamp).
				AddRow(runtime2ID, tenantID, runtimes[1].Name, runtimes[1].Description, runtimes[1].Status.Condition, runtimes[1].CreationTimestamp, runtimes[1].CreationTimestamp),
			TotalCount: 2,
			Runtimes:   runtimes,
		}}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			mockConverter := &automock.EntityConverter{}
			mockConverter.On("MultipleFromEntities", mock.Anything).Return(testCase.Runtimes)

			sqlxDB, sqlMock := testdb.MockDatabase(t)
			defer sqlMock.AssertExpectations(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
			pgRepository := runtime.NewRepository(mockConverter)
			expectedQuery := fmt.Sprintf(pageableQuery, fixTenantIsolationSubquery(), testCase.ExpectedLimit, testCase.ExpectedOffset)

			sqlMock.ExpectQuery(expectedQuery).
				WithArgs(tenantID).
				WillReturnRows(testCase.Rows)
			countRow := sqlMock.NewRows([]string{"count"}).AddRow(testCase.TotalCount)

			sqlMock.ExpectQuery(countQuery).
				WithArgs(tenantID).
				WillReturnRows(countRow)

			//THEN
			modelRuntimePage, err := pgRepository.List(ctx, tenantID, nil, testCase.InputPageSize, testCase.InputCursor)

			//THEN
			require.NoError(t, err)
			assert.Equal(t, testCase.ExpectedLimit, modelRuntimePage.TotalCount)
			require.NoError(t, sqlMock.ExpectationsWereMet())

			assert.Equal(t, runtime1ID, modelRuntimePage.Data[0].ID)
			assert.Equal(t, tenantID, modelRuntimePage.Data[0].Tenant)

			assert.Equal(t, runtime2ID, modelRuntimePage.Data[1].ID)
			assert.Equal(t, tenantID, modelRuntimePage.Data[1].Tenant)
		})
	}

	t.Run("Returns error when decoded cursor is non-positive number", func(t *testing.T) {
		//GIVEN
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("multipleFromEntities", mock.Anything).Return(runtimes)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		pgRepository := runtime.NewRepository(mockConverter)
		//THEN
		_, err := pgRepository.List(ctx, tenantID, nil, 2, convertIntToBase64String(-3))

		//THEN
		require.EqualError(t, err, "while decoding page cursor: Invalid data [reason=cursor is not correct]")
	})
}

func TestPgRepository_ListAll(t *testing.T) {
	//GIVEN
	tenantID := uuid.New().String()
	runtime1ID := uuid.New().String()
	runtime2ID := uuid.New().String()

	runtimes := []*model.Runtime{
		fixModelRuntime(t, runtime1ID, tenantID, "Runtime ABC", "Description for runtime ABC"),
		fixModelRuntime(t, runtime2ID, tenantID, "Runtime XYZ", "Description for runtime XYZ"),
	}

	pageableQuery := fmt.Sprintf(`^SELECT (.+) FROM public.runtimes WHERE %s$`, fixTenantIsolationSubquery())

	testCases := []struct {
		Name       string
		Rows       *sqlmock.Rows
		TotalCount int
		Runtimes   []*model.Runtime
	}{
		{
			Name: "Success listing",
			Rows: sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
				AddRow(runtime1ID, tenantID, runtimes[0].Name, runtimes[0].Description, runtimes[0].Status.Condition, runtimes[0].CreationTimestamp, runtimes[0].CreationTimestamp).
				AddRow(runtime2ID, tenantID, runtimes[1].Name, runtimes[1].Description, runtimes[1].Status.Condition, runtimes[1].CreationTimestamp, runtimes[1].CreationTimestamp),
			TotalCount: 2,
			Runtimes:   runtimes,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			mockConverter := &automock.EntityConverter{}
			mockConverter.On("MultipleFromEntities", mock.Anything).Return(testCase.Runtimes)
			sqlxDB, sqlMock := testdb.MockDatabase(t)
			defer sqlMock.AssertExpectations(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
			pgRepository := runtime.NewRepository(mockConverter)
			expectedQuery := pageableQuery

			sqlMock.ExpectQuery(expectedQuery).
				WithArgs(tenantID).
				WillReturnRows(testCase.Rows)

			//WHEN
			modelRuntime, err := pgRepository.ListAll(ctx, tenantID, nil)

			//THEN
			require.NoError(t, err)
			require.NoError(t, sqlMock.ExpectationsWereMet())

			assert.Equal(t, runtime1ID, modelRuntime[0].ID)
			assert.Equal(t, tenantID, modelRuntime[0].Tenant)

			assert.Equal(t, runtime2ID, modelRuntime[1].ID)
			assert.Equal(t, tenantID, modelRuntime[1].Tenant)

			assert.Len(t, modelRuntime, testCase.TotalCount)
		})
	}
}

func TestPgRepository_List_WithFiltersShouldReturnRuntimeModelsForRuntimeEntities(t *testing.T) {
	// given
	runtime1ID := uuid.New().String()
	runtime2ID := uuid.New().String()
	tenantID := uuid.New().String()
	rowSize := 2

	runtimes := []*model.Runtime{
		fixModelRuntime(t, runtime1ID, tenantID, "Runtime ABC", "Description for runtime ABC"),
		fixModelRuntime(t, runtime2ID, tenantID, "Runtime XYZ", "Description for runtime XYZ"),
	}

	mockConverter := &automock.EntityConverter{}
	mockConverter.On("MultipleFromEntities", mock.Anything).Return(runtimes)

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
		AddRow(runtime1ID, tenantID, runtimes[0].Name, runtimes[0].Description, runtimes[0].Status.Condition, runtimes[0].CreationTimestamp, runtimes[0].CreationTimestamp).
		AddRow(runtime2ID, tenantID, runtimes[1].Name, runtimes[1].Description, runtimes[1].Status.Condition, runtimes[1].CreationTimestamp, runtimes[1].CreationTimestamp)

	filterQuery := fmt.Sprintf(`  AND id IN
						\(SELECT "runtime_id" FROM public.labels
							WHERE "runtime_id" IS NOT NULL
							AND %s
							AND "key" = \$3\)`, fixTenantIsolationSubqueryWithArg(2))
	sqlQuery := fmt.Sprintf(`^SELECT (.+) FROM public.runtimes
								WHERE %s %s ORDER BY name LIMIT %d OFFSET 0`, fixTenantIsolationSubqueryWithArg(1), filterQuery, rowSize)

	sqlMock.ExpectQuery(sqlQuery).
		WithArgs(tenantID, tenantID, "foo").
		WillReturnRows(rows)

	countRows := sqlMock.NewRows([]string{"count"}).AddRow(rowSize)

	countQuery := fmt.Sprintf(`^SELECT COUNT\(\*\) FROM public.runtimes WHERE %s %s`, fixTenantIsolationSubquery(), filterQuery)
	sqlMock.ExpectQuery(countQuery).
		WithArgs(tenantID, tenantID, "foo").
		WillReturnRows(countRows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	labelFilterFoo := labelfilter.LabelFilter{
		Key: "foo",
	}
	filter := []*labelfilter.LabelFilter{&labelFilterFoo}

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	modelRuntimePage, err := pgRepository.List(ctx, tenantID, filter, rowSize, "")

	//then
	assert.NoError(t, err)
	require.NotNil(t, modelRuntimePage)
	assert.Equal(t, rowSize, modelRuntimePage.TotalCount)
	require.NoError(t, sqlMock.ExpectationsWereMet())

	assert.Equal(t, runtime1ID, modelRuntimePage.Data[0].ID)
	assert.Equal(t, tenantID, modelRuntimePage.Data[0].Tenant)

	assert.Equal(t, runtime2ID, modelRuntimePage.Data[1].ID)
	assert.Equal(t, tenantID, modelRuntimePage.Data[1].Tenant)
}
*/
func TestPgRepository_Create(t *testing.T) {
	var nilRtModel *model.Runtime
	rtModel := fixDetailedModelRuntime(t, "foo", "Foo", "Lorem ipsum")
	rtEntity := fixDetailedEntityRuntime(t, "foo", "Foo", "Lorem ipsum")

	suite := testdb.RepoCreateTestSuite{
		Name: "Generic Create Runtime",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:       regexp.QuoteMeta(`INSERT INTO public.runtimes ( id, name, description, status_condition, status_timestamp, creation_timestamp ) VALUES ( ?, ?, ?, ?, ?, ? )`),
				Args:        []driver.Value{rtModel.ID, rtModel.Name, rtModel.Description, rtModel.Status.Condition, rtModel.Status.Timestamp, rtModel.CreationTimestamp},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
			{
				Query:       regexp.QuoteMeta(`INSERT INTO tenant_runtimes ( tenant_id, id, owner ) VALUES ( ?, ?, ? )`),
				Args:        []driver.Value{tenantID, rtModel.ID, true},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: runtime.NewRepository,
		ModelEntity:         rtModel,
		DBEntity:            rtEntity,
		NilModelEntity:      nilRtModel,
		TenantID:            tenantID,
		IsTopLevelEntity:    true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	var nilRtModel *model.Runtime
	rtModel := fixDetailedModelRuntime(t, "foo", "Foo", "Lorem ipsum")
	rtEntity := fixDetailedEntityRuntime(t, "foo", "Foo", "Lorem ipsum")

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Runtime",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.runtimes SET name = ?, description = ?, status_condition = ?, status_timestamp = ? WHERE id = ? AND (id IN (SELECT id FROM tenant_runtimes WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:          []driver.Value{rtModel.Name, rtModel.Description, rtModel.Status.Condition, rtModel.Status.Timestamp, rtModel.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: runtime.NewRepository,
		ModelEntity:         rtModel,
		DBEntity:            rtEntity,
		NilModelEntity:      nilRtModel,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

/*
func TestPgRepository_Delete_ShouldDeleteRuntimeEntityUsingValidModel(t *testing.T) {
	// given
	runtimeID := uuid.New().String()
	tenantID := uuid.New().String()
	modelRuntime := fixModelRuntime(t, runtimeID, tenantID, "Runtime BCD", "Description for runtime BCD")

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	mockConverter := &automock.EntityConverter{}

	sqlMock.ExpectExec(fmt.Sprintf(`^DELETE FROM public.runtimes WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())).
		WithArgs(tenantID, runtimeID).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	err := pgRepository.Delete(ctx, tenantID, modelRuntime.ID)

	// then
	assert.NoError(t, err)
}
*/

func TestPgRepository_Exist(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Runtime Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.runtimes WHERE id = $1 AND (id IN (SELECT id FROM tenant_runtimes WHERE tenant_id = $2))`),
				Args:     []driver.Value{runtimeID, tenantID},
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
		RepoConstructorFunc: runtime.NewRepository,
		TargetID:            runtimeID,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

func convertIntToBase64String(number int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(number)))
}
