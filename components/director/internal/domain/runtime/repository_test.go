package runtime_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_GetByID_ShouldReturnRuntimeModelForRuntimeEntity(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()

	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	mockConverter := &automock.EntityConverter{}

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
		AddRow(runtimeID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, timestamp)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM public.runtimes WHERE tenant_id = \$1 AND id = \$2$`).
		WithArgs(tenantID, runtimeID).
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	modelRuntime, err := pgRepository.GetByID(ctx, tenantID, runtimeID)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeID, modelRuntime.ID)
	assert.Equal(t, tenantID, modelRuntime.Tenant)
}

func TestPgRepository_GetByFiltersAndID_WithoutAdditionalFiltersShouldReturnRuntimeModelForRuntimeEntity(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()

	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	mockConverter := &automock.EntityConverter{}

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
		AddRow(runtimeID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, timestamp)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM public.runtimes WHERE tenant_id = \$1 AND id = \$2$`).
		WithArgs(tenantID, runtimeID).
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	modelRuntime, err := pgRepository.GetByFiltersAndID(ctx, tenantID, runtimeID, nil)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeID, modelRuntime.ID)
	assert.Equal(t, tenantID, modelRuntime.Tenant)
}

func TestPgRepository_GetByFiltersAndID_WithAdditionalFiltersShouldReturnRuntimeModelForRuntimeEntity(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()

	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	mockConverter := &automock.EntityConverter{}

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
		AddRow(runtimeID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, timestamp)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM public.runtimes WHERE tenant_id = \$1 AND id = \$2 AND id IN \(SELECT "runtime_id" FROM public\.labels WHERE "runtime_id" IS NOT NULL AND "tenant_id" = \$3 AND "key" = \$4\)$`).
		WithArgs(tenantID, runtimeID, tenantID, "someKey").
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	filters := []*labelfilter.LabelFilter{labelfilter.NewForKey("someKey")}
	modelRuntime, err := pgRepository.GetByFiltersAndID(ctx, tenantID, runtimeID, filters)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeID, modelRuntime.ID)
	assert.Equal(t, tenantID, modelRuntime.Tenant)
}

func TestPgRepository_GetByFiltersGlobal_ShouldReturnRuntimeModelForRuntimeEntity(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()

	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	mockConverter := &automock.EntityConverter{}

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
		AddRow(runtimeID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, timestamp)

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
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeID, modelRuntime.ID)
}

func TestPgRepository_GetOldestForFilters_ShouldReturnRuntimeModelForRuntimeEntity(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()

	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	mockConverter := &automock.EntityConverter{}

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
		AddRow(runtimeID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, timestamp)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM public.runtimes WHERE tenant_id = \$1 ORDER BY creation_timestamp ASC$`).
		WithArgs(tenantID).
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	modelRuntime, err := pgRepository.GetOldestForFilters(ctx, tenantID, nil)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeID, modelRuntime.ID)
	assert.Equal(t, tenantID, modelRuntime.Tenant)
}

func TestPgRepository_GetOldestForFilters_WithAdditionalFilers_ShouldReturnRuntimeModelForRuntimeEntity(t *testing.T) {
	// given
	tenantID := uuid.New().String()
	runtimeID := uuid.New().String()

	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	mockConverter := &automock.EntityConverter{}

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}).
		AddRow(runtimeID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, timestamp)
	sqlMock.ExpectQuery(`^SELECT (.+) FROM public.runtimes WHERE tenant_id = \$1 AND id IN \(SELECT "runtime_id" FROM public\.labels WHERE "runtime_id" IS NOT NULL AND "tenant_id" = \$2 AND "key" = \$3 AND "value" \?\| array\[\$4\]\) ORDER BY creation_timestamp ASC$`).
		WithArgs(tenantID, tenantID, "scenarios", "DEFAULT").
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	pgRepository := runtime.NewRepository(mockConverter)

	// when
	filters := []*labelfilter.LabelFilter{
		&labelfilter.LabelFilter{
			Key:   model.ScenariosKey,
			Query: str.Ptr(`$[*] ? (@ == "DEFAULT")`),
		},
	}
	modelRuntime, err := pgRepository.GetOldestForFilters(ctx, tenantID, filters)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeID, modelRuntime.ID)
	assert.Equal(t, tenantID, modelRuntime.Tenant)
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

	pageableQuery := `^SELECT (.+) FROM public.runtimes WHERE tenant_id = \$1 ORDER BY name LIMIT %d OFFSET %d$`
	countQuery := regexp.QuoteMeta(`SELECT COUNT(*) FROM public.runtimes WHERE tenant_id = $1`)

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
			expectedQuery := fmt.Sprintf(pageableQuery, testCase.ExpectedLimit, testCase.ExpectedOffset)

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

	pageableQuery := `^SELECT (.+) FROM public.runtimes WHERE tenant_id = \$1$`

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
			expectedQuery := fmt.Sprintf(pageableQuery)

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

	filterQuery := `  AND id IN 
						\(SELECT "runtime_id" FROM public.labels 
							WHERE "runtime_id" IS NOT NULL 
							AND "tenant_id" = \$2 
							AND "key" = \$3\)`
	sqlQuery := fmt.Sprintf(`^SELECT (.+) FROM public.runtimes 
								WHERE tenant_id = \$1 %s ORDER BY name LIMIT %d OFFSET 0`, filterQuery, rowSize)

	sqlMock.ExpectQuery(sqlQuery).
		WithArgs(tenantID, tenantID, "foo").
		WillReturnRows(rows)

	countRows := sqlMock.NewRows([]string{"count"}).AddRow(rowSize)

	countQuery := fmt.Sprintf(`^SELECT COUNT\(\*\) FROM public.runtimes WHERE tenant_id = \$1 %s`, filterQuery)
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

func TestPgRepository_Create_ShouldCreateRuntimeEntityFromValidModel(t *testing.T) {
	// given
	runtimeID := uuid.New().String()
	tenantID := uuid.New().String()
	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)

	description := "Description for runtime BCD"
	modelRuntime := &model.Runtime{
		ID:          runtimeID,
		Tenant:      tenantID,
		Name:        "Runtime XYZ",
		Description: &description,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusConditionInitial,
			Timestamp: timestamp,
		},
		CreationTimestamp: timestamp,
	}

	mockConverter := &automock.EntityConverter{}

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	sqlMock.ExpectExec(`^INSERT INTO public.runtimes \(.+\) VALUES \(.+\)$`).
		WithArgs(modelRuntime.ID, modelRuntime.Tenant, modelRuntime.Name, modelRuntime.Description, modelRuntime.Status.Condition, modelRuntime.Status.Timestamp, modelRuntime.CreationTimestamp).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	err = pgRepository.Create(ctx, modelRuntime)

	// then
	assert.NoError(t, err)
}

func TestPgRepository_Update_ShouldUpdateRuntimeEntityFromValidModel(t *testing.T) {
	// given
	runtimeID := uuid.New().String()
	tenantID := uuid.New().String()
	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)

	description := "Description for runtime BCD"
	modelRuntime := &model.Runtime{
		ID:          runtimeID,
		Tenant:      tenantID,
		Name:        "Runtime XYZ",
		Description: &description,
		Status: &model.RuntimeStatus{
			Condition: model.RuntimeStatusConditionInitial,
			Timestamp: timestamp,
		},
		CreationTimestamp: timestamp,
	}

	mockConverter := &automock.EntityConverter{}

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.runtimes SET name = ?, description = ?, status_condition = ?, status_timestamp = ? WHERE tenant_id = ? AND id = ?`)).
		WithArgs(modelRuntime.Name, modelRuntime.Description, modelRuntime.Status.Condition, modelRuntime.Status.Timestamp, modelRuntime.Tenant, modelRuntime.ID).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	err = pgRepository.Update(ctx, modelRuntime)

	// then
	assert.NoError(t, err)
}

func TestPgRepository_Delete_ShouldDeleteRuntimeEntityUsingValidModel(t *testing.T) {
	// given
	runtimeID := uuid.New().String()
	tenantID := uuid.New().String()
	modelRuntime := fixModelRuntime(t, runtimeID, tenantID, "Runtime BCD", "Description for runtime BCD")

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	mockConverter := &automock.EntityConverter{}

	sqlMock.ExpectExec(fmt.Sprintf(`^DELETE FROM public.runtimes WHERE tenant_id = \$1 AND id = \$2$`)).
		WithArgs(tenantID, runtimeID).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	err := pgRepository.Delete(ctx, tenantID, modelRuntime.ID)

	// then
	assert.NoError(t, err)
}

func TestPgRepository_Exist(t *testing.T) {
	// given
	runtimeID := uuid.New().String()
	tenantID := uuid.New().String()

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	mockConverter := &automock.EntityConverter{}

	sqlMock.ExpectQuery(fmt.Sprintf(`^SELECT 1 FROM public.runtimes WHERE tenant_id = \$1 AND id = \$2$`)).
		WithArgs(tenantID, runtimeID).
		WillReturnRows(testdb.RowWhenObjectExist())

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	ex, err := pgRepository.Exists(ctx, tenantID, runtimeID)

	// then
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestPgRepository_UpdateTenantID_ShouldUpdateTenant(t *testing.T) {
	// given
	runtimeID := uuid.New().String()
	newTenantID := uuid.New().String()
	mockConverter := &automock.EntityConverter{}

	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	sqlMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.runtimes SET tenant_id = ? WHERE id = ?`)).
		WithArgs(newTenantID, runtimeID).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewRepository(mockConverter)

	// when
	err := pgRepository.UpdateTenantID(ctx, runtimeID, newTenantID)

	// then
	assert.NoError(t, err)
}

func convertIntToBase64String(number int) string {
	return string(base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(number))))
}
