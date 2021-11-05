package repo_test

/*
func TestGetSingle(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	sut := repo.NewSingleGetter(UserType, "users", "tenant_id", []string{"id_col", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col", "tenant_id", "first_name", "last_name", "age"}).AddRow(givenID, givenTenant, "givenFirstName", "givenLastName", 18)
		mock.ExpectQuery(defaultExpectedGetSingleQuery()).WithArgs(givenTenant, givenID).WillReturnRows(rows)
		dest := User{}
		// WHEN
		err := sut.Get(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenID, dest.ID)
		assert.Equal(t, givenTenant, dest.Tenant)
		assert.Equal(t, "givenFirstName", dest.FirstName)
		assert.Equal(t, "givenLastName", dest.LastName)
		assert.Equal(t, 18, dest.Age)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s", fixTenantIsolationSubquery()))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col", "tenant_id", "first_name", "last_name", "age"}).AddRow(givenID, givenTenant, "givenFirstName", "givenLastName", 18)
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant).WillReturnRows(rows)
		dest := User{}
		// WHEN
		err := sut.Get(ctx, givenTenant, repo.Conditions{}, repo.NoOrderBy, &dest)
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
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND first_name = $2 AND last_name = $3", fixTenantIsolationSubquery()))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant, "john", "doe").WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success when IN condition", func(t *testing.T) {
		// GIVEN
		givenTenant := uuidB()
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND first_name IN (SELECT name from names WHERE description = $2 AND id = $3)", fixTenantIsolationSubquery()))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant, "foo", 3).WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx, givenTenant, repo.Conditions{repo.NewInConditionForSubQuery("first_name", "SELECT name from names WHERE description = ? AND id = ?", []interface{}{"foo", 3})}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success when IN condition for values", func(t *testing.T) {
		// GIVEN
		givenTenant := uuidB()
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND first_name IN ($2, $3)", fixTenantIsolationSubquery()))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant, "foo", "bar").WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx, givenTenant, repo.Conditions{repo.NewInConditionForStringValues("first_name", []string{"foo", "bar"})}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with order by params", func(t *testing.T) {
		// GIVEN
		givenTenant := uuidB()
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s ORDER BY first_name ASC", fixTenantIsolationSubquery()))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant).WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx, givenTenant, nil, repo.OrderByParams{repo.NewAscOrderBy("first_name")}, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with multiple order by params", func(t *testing.T) {
		// GIVEN
		givenTenant := uuidB()
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s ORDER BY first_name ASC, last_name DESC", fixTenantIsolationSubquery()))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant).WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx, givenTenant, nil, repo.OrderByParams{repo.NewAscOrderBy("first_name"), repo.NewDescOrderBy("last_name")}, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with conditions and order by params", func(t *testing.T) {
		// GIVEN
		givenTenant := uuidB()
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND first_name = $2 AND last_name = $3 ORDER BY first_name ASC", fixTenantIsolationSubquery()))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant, "john", "doe").WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.Get(ctx,
			givenTenant,
			repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")},
			repo.OrderByParams{repo.NewAscOrderBy("first_name")},
			&dest)
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
		err := sut.Get(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)
		dest := User{}

		err := sut.Get(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &dest)

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns ErrorNotFound if object not found", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		noRows := sqlmock.NewRows([]string{"id_col", "tenant_id", "first_name", "last_name", "age"})
		mock.ExpectQuery(defaultExpectedGetSingleQuery()).WillReturnRows(noRows)
		dest := User{}
		// WHEN
		err := sut.Get(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NotNil(t, err)
		assert.True(t, apperrors.IsNotFoundError(err))
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.Get(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &User{})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		err := sut.Get(context.TODO(), givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, nil)
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})
}

func TestGetSingleGlobal(t *testing.T) {
	givenID := uuidA()
	sut := repo.NewSingleGetterGlobal(UserType, "users", []string{"id_col", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE id_col = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col", "first_name", "last_name", "age"}).AddRow(givenID, "givenFirstName", "givenLastName", 18)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnRows(rows)
		dest := User{}
		// WHEN
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenID, dest.ID)
		assert.Equal(t, "givenFirstName", dest.FirstName)
		assert.Equal(t, "givenLastName", dest.LastName)
		assert.Equal(t, 18, dest.Age)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id_col, tenant_id, first_name, last_name, age FROM users")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col", "first_name", "last_name", "age"}).AddRow(givenID, "givenFirstName", "givenLastName", 18)
		mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
		dest := User{}
		// WHEN
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, givenID, dest.ID)
		assert.Equal(t, "givenFirstName", dest.FirstName)
		assert.Equal(t, "givenLastName", dest.LastName)
		assert.Equal(t, 18, dest.Age)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE first_name = $1 AND last_name = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WithArgs("john", "doe").WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")}, repo.NoOrderBy, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with order by params", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id_col, tenant_id, first_name, last_name, age FROM users ORDER BY first_name ASC")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.GetGlobal(ctx, nil, repo.OrderByParams{repo.NewAscOrderBy("first_name")}, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with multiple order by params", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id_col, tenant_id, first_name, last_name, age FROM users ORDER BY first_name ASC, last_name DESC")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
		mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
		// WHEN
		dest := User{}
		err := sut.GetGlobal(ctx, nil, repo.OrderByParams{repo.NewAscOrderBy("first_name"), repo.NewDescOrderBy("last_name")}, &dest)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success with conditions and order by params", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE first_name = $1 AND last_name = $2 ORDER BY first_name ASC")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{"id_col"}).AddRow(uuidA())
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
		expectedQuery := regexp.QuoteMeta("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE id_col = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WillReturnError(someError())
		dest := User{}
		// WHEN
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns ErrorNotFound if object not found", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE id_col = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		noRows := sqlmock.NewRows([]string{"id_col", "first_name", "last_name", "age"})
		mock.ExpectQuery(expectedQuery).WillReturnRows(noRows)
		dest := User{}
		// WHEN
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &dest)
		// THEN
		require.NotNil(t, err)
		assert.True(t, apperrors.IsNotFoundError(err))
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, &User{})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		err := sut.GetGlobal(context.TODO(), repo.Conditions{repo.NewEqualCondition("id_col", givenID)}, repo.NoOrderBy, nil)
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})
}

func defaultExpectedGetSingleQuery() string {
	givenQuery := fmt.Sprintf("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND id_col = $2", fixTenantIsolationSubquery())
	return regexp.QuoteMeta(givenQuery)
}
*/