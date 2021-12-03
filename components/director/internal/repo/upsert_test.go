package repo_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/lib/pq"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestUpsertGlobal(t *testing.T) {
	expectedQuery := regexp.QuoteMeta(`INSERT INTO users ( id, tenant_id, first_name, last_name, age )
		VALUES ( ?, ?, ?, ?, ? ) ON CONFLICT ( tenant_id, first_name, last_name ) DO UPDATE SET age=EXCLUDED.age`)
	expectedQueryWithTenantCheck := regexp.QuoteMeta(`INSERT INTO users ( id, tenant_id, first_name, last_name, age )
		VALUES ( ?, ?, ?, ?, ? ) ON CONFLICT ( tenant_id, first_name, last_name ) DO UPDATE SET age=EXCLUDED.age WHERE users.tenant_id = ?`)

	sut := repo.NewUpserterGlobal(UserType, "users", []string{"id", "tenant_id", "first_name", "last_name", "age"}, []string{"tenant_id", "first_name", "last_name"}, []string{"age"})
	sutWithEmbededTenant := repo.NewUpserterWithEmbeddedTenant(UserType, "users", []string{"id", "tenant_id", "first_name", "last_name", "age"}, []string{"tenant_id", "first_name", "last_name"}, []string{"age"}, "tenant_id")

	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{
			ID:        "given_id",
			Tenant:    "given_tenant",
			FirstName: "given_first_name",
			LastName:  "given_last_name",
			Age:       55,
		}

		mock.ExpectExec(expectedQuery).
			WithArgs("given_id", "given_tenant", "given_first_name", "given_last_name", 55).WillReturnResult(sqlmock.NewResult(1, 1))
		// WHEN
		err := sut.UpsertGlobal(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{}

		mock.ExpectExec(expectedQuery).
			WillReturnError(someError())
		// WHEN
		err := sut.UpsertGlobal(ctx, givenUser)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)
		givenUser := User{}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)

		err := sut.UpsertGlobal(ctx, givenUser)

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns non unique error", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{}

		mock.ExpectExec(expectedQuery).
			WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
		// WHEN
		err := sut.UpsertGlobal(ctx, givenUser)
		// THEN
		require.True(t, apperrors.IsNotUniqueError(err))
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		// WHEN
		err := sut.UpsertGlobal(context.TODO(), User{})
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		// WHEN
		err := sut.UpsertGlobal(context.TODO(), nil)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})

	t.Run("returns error if there are no affected rows", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{
			ID:        "given_id",
			Tenant:    "given_tenant",
			FirstName: "given_first_name",
			LastName:  "given_last_name",
			Age:       55,
		}

		mock.ExpectExec(expectedQuery).
			WithArgs("given_id", "given_tenant", "given_first_name", "given_last_name", 55).WillReturnResult(sqlmock.NewResult(1, 0))
		// WHEN
		err := sut.UpsertGlobal(ctx, givenUser)
		// THEN
		require.Contains(t, err.Error(), "should upsert single row, but upserted 0 rows")
	})

	t.Run("returns error if there are more than one affected rows", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{
			ID:        "given_id",
			Tenant:    "given_tenant",
			FirstName: "given_first_name",
			LastName:  "given_last_name",
			Age:       55,
		}

		mock.ExpectExec(expectedQuery).
			WithArgs("given_id", "given_tenant", "given_first_name", "given_last_name", 55).WillReturnResult(sqlmock.NewResult(1, 7))
		// WHEN
		err := sut.UpsertGlobal(ctx, givenUser)
		// THEN
		require.Contains(t, err.Error(), "should upsert single row, but upserted 7 rows")
	})

	t.Run("success with embedded tenant", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{
			ID:        "given_id",
			Tenant:    "given_tenant",
			FirstName: "given_first_name",
			LastName:  "given_last_name",
			Age:       55,
		}

		mock.ExpectExec(expectedQueryWithTenantCheck).
			WithArgs("given_id", "given_tenant", "given_first_name", "given_last_name", 55, "given_tenant").WillReturnResult(sqlmock.NewResult(1, 1))
		// WHEN
		err := sutWithEmbededTenant.UpsertGlobal(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error if with embedded tenant there are no affected rows", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{
			ID:        "given_id",
			Tenant:    "given_tenant",
			FirstName: "given_first_name",
			LastName:  "given_last_name",
			Age:       55,
		}

		mock.ExpectExec(expectedQueryWithTenantCheck).
			WithArgs("given_id", "given_tenant", "given_first_name", "given_last_name", 55, "given_tenant").WillReturnResult(sqlmock.NewResult(1, 0))
		// WHEN
		err := sutWithEmbededTenant.UpsertGlobal(ctx, givenUser)
		// THEN
		require.Contains(t, err.Error(), "Owner access is needed for resource modification")
	})

	t.Run("returns error if with embedded tenant there are more than one affected rows", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{
			ID:        "given_id",
			Tenant:    "given_tenant",
			FirstName: "given_first_name",
			LastName:  "given_last_name",
			Age:       55,
		}

		mock.ExpectExec(expectedQueryWithTenantCheck).
			WithArgs("given_id", "given_tenant", "given_first_name", "given_last_name", 55, "given_tenant").WillReturnResult(sqlmock.NewResult(1, 7))
		// WHEN
		err := sutWithEmbededTenant.UpsertGlobal(ctx, givenUser)
		// THEN
		require.Contains(t, err.Error(), "should upsert single row, but upserted 7 rows")
	})
}

func TestUpsert(t *testing.T) {
	expectedQuery := regexp.QuoteMeta(`INSERT INTO apps ( id, name, description )
		VALUES ( ?, ?, ? ) 
		ON CONFLICT ( id ) 
		DO UPDATE SET name=EXCLUDED.name, description=EXCLUDED.description 
		WHERE (apps.id IN (SELECT id FROM tenant_applications WHERE tenant_id = ? AND owner = true))`)

	expectedTenantAccessQuery := regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)", "tenant_applications", repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))

	sut := repo.NewUpserter(appTableName, []string{"id", "name", "description"}, []string{"id"}, []string{"name", "description"})

	resourceType := resource.Application
	tenant := "tenant"

	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(expectedQuery).
			WithArgs(appID, appName, appDescription, tenant).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(expectedTenantAccessQuery).
			WithArgs(tenant, appID, true).WillReturnResult(sqlmock.NewResult(1, 1))
		// WHEN
		err := sut.Upsert(ctx, resourceType, tenant, fixApp)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when upsert operation failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(expectedQuery).
			WithArgs(appID, appName, appDescription, tenant).WillReturnError(someError())
		// WHEN
		err := sut.Upsert(ctx, resourceType, tenant, fixApp)
		// THEN
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error when adding tenant access record failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(expectedQuery).
			WithArgs(appID, appName, appDescription, tenant).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(expectedTenantAccessQuery).
			WithArgs(tenant, appID, true).WillReturnError(someError())
		// WHEN
		err := sut.Upsert(ctx, resourceType, tenant, fixApp)
		// THEN
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)

		err := sut.Upsert(ctx, resourceType, tenant, fixApp)

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		// WHEN
		err := sut.Upsert(context.TODO(), resourceType, tenant, fixApp)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		// WHEN
		err := sut.Upsert(context.TODO(), resourceType, tenant, nil)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})

	t.Run("returns error if does not have owner access", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(expectedQuery).
			WithArgs(appID, appName, appDescription, "other_tenant").WillReturnResult(sqlmock.NewResult(1, 0))
		// WHEN
		err := sut.Upsert(ctx, resourceType, "other_tenant", fixApp)
		// THEN
		require.Contains(t, err.Error(), "Owner access is needed for resource modification")
	})

	t.Run("returns error if there are more than one affected rows", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(expectedQuery).
			WithArgs(appID, appName, appDescription, tenant).WillReturnResult(sqlmock.NewResult(1, 7))
		// WHEN
		err := sut.Upsert(ctx, resourceType, tenant, fixApp)
		// THEN
		require.Contains(t, err.Error(), "should upsert single row, but upserted 7 rows")
	})

	t.Run("returns error if the entity does not have accessTable", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		// WHEN
		err := sut.Upsert(ctx, resource.Tenant, tenant, fixApp)
		// THEN
		require.Contains(t, err.Error(), "entity tenant does not have access table")
	})
}

func TestUpsertGlobalWhenWrongConfiguration(t *testing.T) {
	sut := repo.NewUpserterGlobal("users", "UserType", []string{"id", "tenant_id", "column_does_not_exist"}, []string{"id", "tenant_id"}, []string{"column_does_not_exist"})
	// GIVEN
	db, mock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), db)
	defer mock.AssertExpectations(t)
	// WHEN
	err := sut.UpsertGlobal(ctx, User{})
	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unexpected error while executing SQL query")
}
