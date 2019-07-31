package runtime_test

import (
	"context"
	"fmt"
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

func TestPgRepository_List_ShouldReturnRuntimeModelsForRuntimeEntities(t *testing.T) {
	// given
	runtime1ID := uuid.New().String()
	runtime2ID := uuid.New().String()
	tenantID := uuid.New().String()

	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	sqlxDB, sqlMock := mockDatabase(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth"}).
		AddRow(runtime1ID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, agentAuthStr).
		AddRow(runtime2ID, tenantID, "Runtime XYZ", "Description for runtime XYZ", "INITIAL", timestamp, agentAuthStr)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM "public"."runtimes" WHERE "tenant_id" = \$1$`).
		WithArgs(tenantID).
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	pgRepository := runtime.NewPostgresRepository()

	// when
	modelRuntimePage, err := pgRepository.List(ctx, tenantID, nil, nil, nil)

	//then
	assert.NoError(t, err)
	assert.Equal(t, 2, modelRuntimePage.TotalCount)
	require.NoError(t, sqlMock.ExpectationsWereMet())

	assert.Equal(t, runtime1ID, modelRuntimePage.Data[0].ID)
	assert.Equal(t, tenantID, modelRuntimePage.Data[0].Tenant)

	assert.Equal(t, runtime2ID, modelRuntimePage.Data[1].ID)
	assert.Equal(t, tenantID, modelRuntimePage.Data[1].Tenant)
}

func TestPgRepository_List_WithFiltersShouldReturnRuntimeModelsForRuntimeEntities(t *testing.T) {
	// given
	runtime1ID := uuid.New().String()
	runtime2ID := uuid.New().String()
	tenantID := uuid.New().String()

	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	require.NoError(t, err)

	sqlxDB, sqlMock := mockDatabase(t)

	rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth"}).
		AddRow(runtime1ID, tenantID, "Runtime ABC", "Description for runtime ABC", "INITIAL", timestamp, agentAuthStr).
		AddRow(runtime2ID, tenantID, "Runtime XYZ", "Description for runtime XYZ", "INITIAL", timestamp, agentAuthStr)

	sqlMock.ExpectQuery(`^SELECT (.+) FROM "public"."runtimes" WHERE "tenant_id" = \$1  AND "id" IN \(SELECT "runtime_id" FROM "public"."labels" WHERE "tenant_id" = '` + tenantID + `' AND "key" = 'foo'\)$`).
		WithArgs(tenantID).
		WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	labelFilterFoo := labelfilter.LabelFilter{
		Key: "foo",
	}
	filter := []*labelfilter.LabelFilter{&labelFilterFoo}

	pgRepository := runtime.NewPostgresRepository()

	// when
	modelRuntimePage, err := pgRepository.List(ctx, tenantID, filter, nil, nil)

	//then
	assert.NoError(t, err)
	assert.Equal(t, 2, modelRuntimePage.TotalCount)
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
