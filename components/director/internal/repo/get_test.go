package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSingle(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	sut := repo.NewSingleGetter("users", "tenant_col", []string{"id_col", "tenant_col", "first_name", "last_name", "age"})

	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col", "tenant_col", "first_name", "last_name", "age"}).AddRow(givenID, givenTenant, "givenFirstName", "givenLastName", 18)
		mock.ExpectQuery(defaultExpectedGetSingleQuery()).WithArgs(givenTenant, givenID).WillReturnRows(rows)
		dest := User{}
		// WHEN
		err := sut.Get(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}}, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenID, dest.ID)
		assert.Equal(t, givenTenant, dest.Tenant)
		assert.Equal(t, "givenFirstName", dest.FirstName)
		assert.Equal(t, "givenLastName", dest.LastName)
		assert.Equal(t, 18, dest.Age)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		givenTenant := uuidB()
		expectedQuery := regexp.QuoteMeta("SELECT id_col FROM users WHERE tenant_col = $1 AND first_name = $2 AND last_name = $3")
		sut := repo.NewSingleGetter("users", "tenant_col", []string{"id_col"})
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant, "john", "doe").WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx, givenTenant, repo.Conditions{{Field: "first_name", Val: "john"}, {Field: "last_name", Val: "doe"}}, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(defaultExpectedGetSingleQuery()).WillReturnError(someError())
		dest := User{}
		// WHEN
		err := sut.Get(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}}, &dest)
		// THEN
		require.EqualError(t, err, "while getting object from DB: some error")
	})

	t.Run("returns ErrorNotFound if object not found", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		noRows := sqlmock.NewRows([]string{"id_col", "tenant_col", "first_name", "last_name", "age"})
		mock.ExpectQuery(defaultExpectedGetSingleQuery()).WillReturnRows(noRows)
		dest := User{}
		// WHEN
		err := sut.Get(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}}, &dest)
		// THEN
		require.NotNil(t, err)
		assert.True(t, repo.IsNotFoundError(err))
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.Get(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}}, &User{})
		require.EqualError(t, err, "unable to fetch database from context")
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		err := sut.Get(context.TODO(), givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}}, nil)
		require.EqualError(t, err, "item cannot be nil")
	})
}

func defaultExpectedGetSingleQuery() string {
	givenQuery := "SELECT id_col, tenant_col, first_name, last_name, age FROM users WHERE tenant_col = $1 AND id_col = $2"
	return regexp.QuoteMeta(givenQuery)
}
