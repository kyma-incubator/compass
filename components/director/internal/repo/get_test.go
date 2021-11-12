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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSingle(t *testing.T) {
	sut := repo.NewSingleGetter(appTableName, appColumns)
	resourceType := resource.Application
	m2mTable, ok := resourceType.TenantAccessTable()
	require.True(t, ok)

	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE id = $1 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$2")))
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(appID, appName, appDescription)
		mock.ExpectQuery(expectedQuery).WithArgs(appID, tenantID).WillReturnRows(rows)
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, appID, dest.ID)
		assert.Equal(t, appName, dest.Name)
		assert.Equal(t, appDescription, dest.Description)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(appID, appName, appDescription)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID).WillReturnRows(rows)
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, nil, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, appID, dest.ID)
		assert.Equal(t, appName, dest.Name)
		assert.Equal(t, appDescription, dest.Description)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE id = $1 AND name = $2 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3")))
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(appID, appName, appDescription)
		mock.ExpectQuery(expectedQuery).WithArgs(appID, appName, tenantID).WillReturnRows(rows)
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID), repo.NewEqualCondition("name", appName)}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, appID, dest.ID)
		assert.Equal(t, appName, dest.Name)
		assert.Equal(t, appDescription, dest.Description)
	})

	t.Run("success when IN condition", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE name IN (SELECT name from names WHERE description = $1 AND id = $2) AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3")))
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(appID, appName, appDescription)
		mock.ExpectQuery(expectedQuery).WithArgs("foo", 3, tenantID).WillReturnRows(rows)
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, repo.Conditions{repo.NewInConditionForSubQuery("name", "SELECT name from names WHERE description = ? AND id = ?", []interface{}{"foo", 3})}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, appID, dest.ID)
		assert.Equal(t, appName, dest.Name)
		assert.Equal(t, appDescription, dest.Description)
	})

	t.Run("success when IN condition for values", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE name IN ($1, $2) AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3")))
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(appID, appName, appDescription)
		mock.ExpectQuery(expectedQuery).WithArgs("foo", "bar", tenantID).WillReturnRows(rows)
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, repo.Conditions{repo.NewInConditionForStringValues("name", []string{"foo", "bar"})}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, appID, dest.ID)
		assert.Equal(t, appName, dest.Name)
		assert.Equal(t, appDescription, dest.Description)
	})

	t.Run("success with order by params", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s ORDER BY name ASC", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(appID, appName, appDescription)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID).WillReturnRows(rows)
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, nil, repo.OrderByParams{repo.NewAscOrderBy("name")}, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, appID, dest.ID)
		assert.Equal(t, appName, dest.Name)
		assert.Equal(t, appDescription, dest.Description)
	})

	t.Run("success with multiple order by params", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s ORDER BY name ASC, description DESC", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(appID, appName, appDescription)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID).WillReturnRows(rows)
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, nil, repo.OrderByParams{repo.NewAscOrderBy("name"), repo.NewDescOrderBy("description")}, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, appID, dest.ID)
		assert.Equal(t, appName, dest.Name)
		assert.Equal(t, appDescription, dest.Description)
	})

	t.Run("success with conditions and order by params", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE name = $1 AND description = $2 AND %s ORDER BY name ASC", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3")))
		rows := sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(appID, appName, appDescription)
		mock.ExpectQuery(expectedQuery).WithArgs(appName, appDescription, tenantID).WillReturnRows(rows)
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("name", appName), repo.NewEqualCondition("description", appDescription)}, repo.OrderByParams{repo.NewAscOrderBy("name"), repo.NewDescOrderBy("description")}, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, appID, dest.ID)
		assert.Equal(t, appName, dest.Name)
		assert.Equal(t, appDescription, dest.Description)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE id = $1 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$2")))
		mock.ExpectQuery(expectedQuery).WithArgs(appID, tenantID).WillReturnError(someError())
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)}, repo.NoOrderBy, &dest)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)
		dest := App{}

		err := sut.Get(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)}, repo.NoOrderBy, &dest)

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns ErrorNotFound if object not found", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		noRows := sqlmock.NewRows([]string{"id", "name", "description"})
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE id = $1 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$2")))
		mock.ExpectQuery(expectedQuery).WithArgs(appID, tenantID).WillReturnRows(noRows)
		dest := App{}
		// WHEN
		err := sut.Get(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NotNil(t, err)
		assert.True(t, apperrors.IsNotFoundError(err))
	})

	t.Run("returns error if entity does not have tenant access table", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		err := sut.Get(ctx, UserType, tenantID, nil, repo.NoOrderBy, &User{})
		require.EqualError(t, err, "entity UserType does not have access table")
	})

	t.Run("returns error if empty tenant id", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		err := sut.Get(ctx, resourceType, "", nil, repo.NoOrderBy, &User{})
		require.EqualError(t, err, apperrors.NewTenantRequiredError().Error())
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.Get(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)}, repo.NoOrderBy, &User{})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		err := sut.Get(context.TODO(), resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)}, repo.NoOrderBy, nil)
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})
}

func TestGetSingleWithEmbeddedTenant(t *testing.T) {
	givenID := "id"
	sut := repo.NewSingleGetterWithEmbeddedTenant(userTableName, "tenant_id", []string{"id", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND id = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).AddRow(givenID, "givenFirstName", "givenLastName", 18)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID, givenID).WillReturnRows(rows)
		dest := User{}
		// WHEN
		err := sut.Get(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenID, dest.ID)
		assert.Equal(t, "givenFirstName", dest.FirstName)
		assert.Equal(t, "givenLastName", dest.LastName)
		assert.Equal(t, 18, dest.Age)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).AddRow(givenID, "givenFirstName", "givenLastName", 18)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID).WillReturnRows(rows)
		dest := User{}
		// WHEN
		err := sut.Get(ctx, UserType, tenantID, nil, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenID, dest.ID)
		assert.Equal(t, "givenFirstName", dest.FirstName)
		assert.Equal(t, "givenLastName", dest.LastName)
		assert.Equal(t, 18, dest.Age)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND first_name = $2 AND last_name = $3")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(givenID)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID, "john", "doe").WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with order by params", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 ORDER BY first_name ASC")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(givenID)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID).WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx, UserType, tenantID, nil, repo.OrderByParams{repo.NewAscOrderBy("first_name")}, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with multiple order by params", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 ORDER BY first_name ASC, last_name DESC")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(givenID)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID).WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx, UserType, tenantID, nil, repo.OrderByParams{repo.NewAscOrderBy("first_name"), repo.NewDescOrderBy("last_name")}, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with conditions and order by params", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND first_name = $2 AND last_name = $3 ORDER BY first_name ASC")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(givenID)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID, "john", "doe").WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx,
			UserType,
			tenantID,
			repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")},
			repo.OrderByParams{repo.NewAscOrderBy("first_name")},
			&dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND id = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID, givenID).WillReturnError(someError())
		dest := User{}
		// WHEN
		err := sut.Get(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns ErrorNotFound if object not found", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND id = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		noRows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"})
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID, givenID).WillReturnRows(noRows)
		dest := User{}
		// WHEN
		err := sut.Get(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NotNil(t, err)
		assert.True(t, apperrors.IsNotFoundError(err))
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.Get(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, &User{})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		err := sut.Get(context.TODO(), UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, nil)
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})
}

func TestGetSingleGlobal(t *testing.T) {
	givenID := "id"
	sut := repo.NewSingleGetterGlobal(UserType, "users", []string{"id", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).AddRow(givenID, "givenFirstName", "givenLastName", 18)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnRows(rows)
		dest := User{}
		// WHEN
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenID, dest.ID)
		assert.Equal(t, "givenFirstName", dest.FirstName)
		assert.Equal(t, "givenLastName", dest.LastName)
		assert.Equal(t, 18, dest.Age)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).AddRow(givenID, "givenFirstName", "givenLastName", 18)
		mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
		dest := User{}
		// WHEN
		err := sut.GetGlobal(ctx, nil, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenID, dest.ID)
		assert.Equal(t, "givenFirstName", dest.FirstName)
		assert.Equal(t, "givenLastName", dest.LastName)
		assert.Equal(t, 18, dest.Age)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE first_name = $1 AND last_name = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(givenID)
		mock.ExpectQuery(expectedQuery).WithArgs("john", "doe").WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with order by params", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users ORDER BY first_name ASC")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(givenID)
		mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.GetGlobal(ctx, nil, repo.OrderByParams{repo.NewAscOrderBy("first_name")}, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with multiple order by params", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users ORDER BY first_name ASC, last_name DESC")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(givenID)
		mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.GetGlobal(ctx, nil, repo.OrderByParams{repo.NewAscOrderBy("first_name"), repo.NewDescOrderBy("last_name")}, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with conditions and order by params", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE first_name = $1 AND last_name = $2 ORDER BY first_name ASC")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id"}).AddRow(givenID)
		mock.ExpectQuery(expectedQuery).WithArgs("john", "doe").WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.GetGlobal(ctx,
			repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")},
			repo.OrderByParams{repo.NewAscOrderBy("first_name")},
			&dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnError(someError())
		dest := User{}
		// WHEN
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns ErrorNotFound if object not found", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id, tenant_id, first_name, last_name, age FROM users WHERE id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		noRows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"})
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnRows(noRows)
		dest := User{}
		// WHEN
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NotNil(t, err)
		assert.True(t, apperrors.IsNotFoundError(err))
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, &User{})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		err := sut.GetGlobal(context.TODO(), repo.Conditions{repo.NewEqualCondition("id", givenID)}, repo.NoOrderBy, nil)
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})
}
