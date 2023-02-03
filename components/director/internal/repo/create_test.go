package repo_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	t.Run("Top Level Entity", func(t *testing.T) {
		creator := repo.NewCreator(appTableName, appColumns)
		resourceType := resource.Application
		m2mTable, ok := resourceType.TenantAccessTable()
		require.True(t, ok)

		t.Run("success", func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id, name, description ) VALUES ( ?, ?, ? )", appTableName))).
				WithArgs(appID, appName, appDescription).WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO %s ( %s, %s, %s ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)", m2mTable, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
				WithArgs(tenantID, appID, true).WillReturnResult(sqlmock.NewResult(1, 1))

			err := creator.Create(ctx, resourceType, tenantID, fixApp)
			require.NoError(t, err)
		})

		t.Run("returns error when operation on db failed", func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id, name, description ) VALUES ( ?, ?, ? )", appTableName))).
				WillReturnError(someError())
			// WHEN
			err := creator.Create(ctx, resourceType, tenantID, fixApp)
			// THEN
			require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		})

		t.Run("context properly canceled", func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			defer mock.AssertExpectations(t)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			ctx = persistence.SaveToContext(ctx, db)
			err := creator.Create(ctx, resourceType, tenantID, fixApp)
			require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
		})

		t.Run("returns non unique error", func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id, name, description ) VALUES ( ?, ?, ? )", appTableName))).
				WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
			// WHEN
			err := creator.Create(ctx, resourceType, tenantID, fixApp)
			// THEN
			require.True(t, apperrors.IsNotUniqueError(err))
		})

		t.Run("returns non unique error if there are matcher columns and the entity already exists", func(t *testing.T) {
			creator := repo.NewCreatorWithMatchingColumns(appTableName, appColumns, []string{"id"})
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id, name, description ) VALUES ( ?, ?, ? )", appTableName))).
				WithArgs(appID, appName, appDescription).WillReturnResult(sqlmock.NewResult(1, 0))
			// WHEN
			err := creator.Create(ctx, resourceType, tenantID, fixApp)
			// THEN
			require.True(t, apperrors.IsNotUniqueError(err))
		})

		t.Run("returns error if missing persistence context", func(t *testing.T) {
			// WHEN
			err := creator.Create(context.TODO(), resourceType, tenantID, fixApp)
			// THEN
			require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
		})

		t.Run("returns error if destination is nil", func(t *testing.T) {
			// WHEN
			err := creator.Create(context.TODO(), resourceType, tenantID, nil)
			// THEN
			require.EqualError(t, err, "Internal Server Error: item cannot be nil")
		})

		t.Run("returns error if id cannot be found", func(t *testing.T) {
			// WHEN
			err := creator.Create(context.TODO(), resourceType, tenantID, struct{}{})
			// THEN
			require.EqualError(t, err, "Internal Server Error: id cannot be empty, check if the entity implements Identifiable")
		})
	})

	t.Run("Child Entity", func(t *testing.T) {
		creator := repo.NewCreator(bundlesTableName, bundleColumns)
		resourceType := resource.Bundle
		m2mTable, ok := resource.Application.TenantAccessTable()
		require.True(t, ok)

		t.Run("success", func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE %s = $1 AND %s = $2 AND %s = $3", m2mTable, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
				WithArgs(tenantID, appID, true).WillReturnRows(testdb.RowWhenObjectExist())
			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id, name, description, app_id ) VALUES ( ?, ?, ?, ? )", bundlesTableName))).
				WithArgs(bundleID, bundleName, bundleDescription, appID).WillReturnResult(sqlmock.NewResult(1, 1))

			err := creator.Create(ctx, resourceType, tenantID, fixBundle)
			require.NoError(t, err)
		})

		t.Run("returns error if tenant does not have access to the parent entity", func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE %s = $1 AND %s = $2 AND %s = $3", m2mTable, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
				WithArgs(tenantID, appID, true).WillReturnRows(testdb.RowWhenObjectDoesNotExist())

			err := creator.Create(ctx, resourceType, tenantID, fixBundle)
			require.Error(t, err)
			require.Contains(t, err.Error(), fmt.Sprintf("tenant %s does not have access to the parent resource %s with ID %s", tenantID, resource.Application, appID))
		})

		t.Run("returns error if checking for parent access fails", func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE %s = $1 AND %s = $2 AND %s = $3", m2mTable, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn, repo.M2MOwnerColumn))).
				WithArgs(tenantID, appID, true).WillReturnError(someError())

			err := creator.Create(ctx, resourceType, tenantID, fixBundle)
			require.Error(t, err)
			require.Contains(t, err.Error(), "while checking for tenant access")
		})

		t.Run("context properly canceled", func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			defer mock.AssertExpectations(t)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			ctx = persistence.SaveToContext(ctx, db)
			err := creator.Create(ctx, resourceType, tenantID, fixBundle)
			require.Error(t, err)
			require.Contains(t, err.Error(), "Maximum processing timeout reached")
		})

		t.Run("returns error if missing parent", func(t *testing.T) {
			// WHEN
			err := creator.Create(context.TODO(), UserType, tenantID, fixUser)
			require.Error(t, err)
			// THEN
			require.Contains(t, err.Error(), fmt.Sprintf("unknown parent for entity type %s", UserType))
		})

		t.Run("returns error if missing parent ID", func(t *testing.T) {
			// WHEN
			err := creator.Create(context.TODO(), resource.API, tenantID, fixUser)
			require.Error(t, err)
			// THEN
			require.Contains(t, err.Error(), fmt.Sprintf("unknown parent for entity type %s", resource.API))
		})

		t.Run("returns error if missing persistence context", func(t *testing.T) {
			// WHEN
			err := creator.Create(context.TODO(), resourceType, tenantID, fixBundle)
			require.Error(t, err)
			// THEN
			require.Contains(t, err.Error(), "unable to fetch database from context")
		})

		t.Run("returns error if destination is nil", func(t *testing.T) {
			// WHEN
			err := creator.Create(context.TODO(), resourceType, tenantID, nil)
			// THEN
			require.EqualError(t, err, "Internal Server Error: item cannot be nil")
		})

		t.Run("returns error if id cannot be found", func(t *testing.T) {
			// WHEN
			err := creator.Create(context.TODO(), resourceType, tenantID, struct{}{})
			// THEN
			require.EqualError(t, err, "Internal Server Error: id cannot be empty, check if the entity implements Identifiable")
		})
	})

	t.Run("Child Entity whose parent is with embedded tenant", func(t *testing.T) {
		creator := repo.NewCreator(webhooksTableName, webhookColumns)
		resourceType := resource.FormationTemplateWebhook
		m2mTable, ok := resource.FormationTemplate.EmbeddedTenantTable()
		require.True(t, ok)

		t.Run("success", func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE %s = $1 AND %s = $2", m2mTable, repo.M2MTenantIDColumn, repo.M2MResourceIDColumn))).
				WithArgs(tenantID, ftID).WillReturnRows(testdb.RowWhenObjectExist())
			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id, formation_template_id ) VALUES ( ?, ? )", webhooksTableName))).
				WithArgs(whID, ftID).WillReturnResult(sqlmock.NewResult(1, 1))

			err := creator.Create(ctx, resourceType, tenantID, fixWebhook)
			require.NoError(t, err)
		})
	})
}

func TestCreateGlobal(t *testing.T) {
	sut := repo.NewCreatorGlobal(UserType, userTableName, []string{"id", "tenant_id", "first_name", "last_name", "age"})
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

		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id, tenant_id, first_name, last_name, age ) VALUES ( ?, ?, ?, ?, ? )", userTableName))).
			WithArgs("given_id", "given_tenant", "given_first_name", "given_last_name", 55).WillReturnResult(sqlmock.NewResult(1, 1))
		// WHEN
		err := sut.Create(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{}

		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id, tenant_id, first_name, last_name, age ) VALUES ( ?, ?, ?, ?, ? )", userTableName))).
			WillReturnError(someError())
		// WHEN
		err := sut.Create(ctx, givenUser)
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
		err := sut.Create(ctx, givenUser)
		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns non unique error", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{}

		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id, tenant_id, first_name, last_name, age ) VALUES ( ?, ?, ?, ?, ? )", userTableName))).
			WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
		// WHEN
		err := sut.Create(ctx, givenUser)
		// THEN
		require.True(t, apperrors.IsNotUniqueError(err))
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		// WHEN
		err := sut.Create(context.TODO(), User{})
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		// WHEN
		err := sut.Create(context.TODO(), nil)
		// THEN
		require.EqualError(t, err, "Internal Server Error: item cannot be nil")
	})
}

func TestCreateWhenWrongConfiguration(t *testing.T) {
	sut := repo.NewCreatorGlobal(UserType, userTableName, []string{"id", "column_does_not_exist"})
	// GIVEN
	db, mock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), db)
	defer mock.AssertExpectations(t)
	// WHEN
	err := sut.Create(ctx, User{})
	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unexpected error while executing SQL query")
}
