package runtime_test

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
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

	sqlxDB, sqlMock := mockDatabase(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth"}).
		AddRow(runtimeID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, agentAuthStr)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM "public"."runtimes" WHERE "id" = \$1 AND "tenant_id" = \$2$`).
		WithArgs(runtimeID, tenantID).
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewPostgresRepository()

	// when
	modelRuntime, err := pgRepository.GetByID(ctx, tenantID, runtimeID)

	//then
	require.NoError(t, err)
	assert.NoError(t, sqlMock.ExpectationsWereMet())
	assert.Equal(t, runtimeID, modelRuntime.ID)
	assert.Equal(t, tenantID, modelRuntime.Tenant)
}

func TestPgRepository_List(t *testing.T) {
	//GIVEN
	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	tenantID := uuid.New().String()
	runtime1ID := uuid.New().String()
	runtime2ID := uuid.New().String()

	limit := 2
	offset := 3

	pageableQuery := `^SELECT (.+) FROM "public"."runtimes" WHERE "tenant_id" = \$1 ORDER BY "id" LIMIT %d OFFSET %d$`
	countQuery := regexp.QuoteMeta(`SELECT COUNT (*) FROM "public"."runtimes" WHERE "tenant_id" = $1`)

	testCases := []struct {
		Name           string
		InputCursor    string
		InputPageSize  int
		ExpectedOffset int
		ExpectedLimit  int
		Rows           *sqlmock.Rows
		TotalCount     int
		ExpectedError  error
	}{
		{
			Name:           "Success getting first page",
			InputPageSize:  2,
			InputCursor:    "",
			ExpectedOffset: 0,
			ExpectedLimit:  limit,
			ExpectedError:  nil,
			Rows: sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth"}).
				AddRow(runtime1ID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, agentAuthStr).
				AddRow(runtime2ID, tenantID, "Runtime XYZ", "Description for runtime XYZ", "INITIAL", timestamp, agentAuthStr),
			TotalCount: 2,
		},
		{
			Name:           "Success getting next page",
			InputPageSize:  2,
			InputCursor:    convertIntToBase64String(offset),
			ExpectedOffset: offset,
			ExpectedLimit:  limit,
			ExpectedError:  nil,
			Rows: sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth"}).
				AddRow(runtime1ID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, agentAuthStr).
				AddRow(runtime2ID, tenantID, "Runtime XYZ", "Description for runtime XYZ", "INITIAL", timestamp, agentAuthStr),
			TotalCount: 2,
		},
		{
			Name:          "Returns error when decoded cursor is non-positive number",
			InputPageSize: 2,
			InputCursor:   convertIntToBase64String(-3),
			ExpectedError: errors.New("cursor is not correct"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sqlxDB, sqlMock := mockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
			pgRepository := runtime.NewPostgresRepository()
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
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedLimit, modelRuntimePage.TotalCount)
				require.NoError(t, sqlMock.ExpectationsWereMet())

				assert.Equal(t, runtime1ID, modelRuntimePage.Data[0].ID)
				assert.Equal(t, tenantID, modelRuntimePage.Data[0].Tenant)

				assert.Equal(t, runtime2ID, modelRuntimePage.Data[1].ID)
				assert.Equal(t, tenantID, modelRuntimePage.Data[1].Tenant)
			}
		})
	}
}

func TestPgRepository_List_WithFiltersShouldReturnRuntimeModelsForRuntimeEntities(t *testing.T) {
	// given
	runtime1ID := uuid.New().String()
	runtime2ID := uuid.New().String()
	tenantID := uuid.New().String()
	rowSize := 2

	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	sqlxDB, sqlMock := mockDatabase(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth"}).
		AddRow(runtime1ID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, agentAuthStr).
		AddRow(runtime2ID, tenantID, "Runtime XYZ", "Description for runtime XYZ", "INITIAL", timestamp, agentAuthStr)

	sqlQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."runtimes" 
								WHERE "tenant_id" = \$1  AND "id" IN
									\(SELECT "runtime_id" FROM "public"."labels" WHERE "runtime_id" IS NOT NULL AND "tenant_id" = '`+tenantID+`' AND "key" = 'foo'\)
									ORDER BY "id" LIMIT %d OFFSET 0$`, rowSize)

	sqlMock.ExpectQuery(sqlQuery).
		WithArgs(tenantID).
		WillReturnRows(rows)

	countRows := sqlMock.NewRows([]string{"count"}).AddRow(rowSize)

	sqlMock.ExpectQuery(`^SELECT COUNT \(\*\) FROM "public"."runtimes" WHERE "tenant_id" = \$1$`).
		WithArgs(tenantID).
		WillReturnRows(countRows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	labelFilterFoo := labelfilter.LabelFilter{
		Key: "foo",
	}
	filter := []*labelfilter.LabelFilter{&labelFilterFoo}

	pgRepository := runtime.NewPostgresRepository()

	// when
	modelRuntimePage, err := pgRepository.List(ctx, tenantID, filter, rowSize, "")

	//then
	assert.NoError(t, err)
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
		AgentAuth: &model.Auth{
			Credential: model.CredentialData{
				Basic: &model.BasicCredentialData{
					Username: "foo",
					Password: "bar",
				},
			},
		},
	}

	sqlxDB, sqlMock := mockDatabase(t)

	sqlMock.ExpectExec(`^INSERT INTO "public"."runtimes" \(.+\) VALUES \(.+\)$`).
		WithArgs(modelRuntime.ID, modelRuntime.Tenant, modelRuntime.Name, modelRuntime.Description, modelRuntime.Status.Condition, modelRuntime.Status.Timestamp, agentAuthStr).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewPostgresRepository()

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
		AgentAuth: &model.Auth{
			Credential: model.CredentialData{
				Basic: &model.BasicCredentialData{
					Username: "foo",
					Password: "bar",
				},
			},
		},
	}

	sqlxDB, sqlMock := mockDatabase(t)

	sqlMock.ExpectExec(fmt.Sprintf(`^UPDATE "public"."runtimes" SET (.+) WHERE "id" = \?$`)).
		WithArgs(modelRuntime.Name, modelRuntime.Description, modelRuntime.Status.Condition, modelRuntime.Status.Timestamp, modelRuntime.ID).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewPostgresRepository()

	// when
	err = pgRepository.Update(ctx, modelRuntime)

	// then
	assert.NoError(t, err)
}

func TestPgRepository_Delete_ShouldDeleteRuntimeEntityUsingValidModel(t *testing.T) {
	// given
	runtimeID := uuid.New().String()
	tenantID := uuid.New().String()
	modelRuntime := fixModelRuntime(runtimeID, tenantID, "Runtime BCD", "Description for runtime BCD")

	sqlxDB, sqlMock := mockDatabase(t)

	sqlMock.ExpectExec(fmt.Sprintf(`^DELETE FROM "public"."runtimes" WHERE "id" = \$1$`)).
		WithArgs(modelRuntime.ID).
		WillReturnResult(sqlmock.NewResult(-1, 1))

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewPostgresRepository()

	// when
	err := pgRepository.Delete(ctx, modelRuntime.ID)

	// then
	assert.NoError(t, err)
}

func mockDatabase(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	sqlDB, sqlMock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(sqlDB, "sqlmock")

	return sqlxDB, sqlMock
}

func convertIntToBase64String(number int) string {
	return string(base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(number))))
}
