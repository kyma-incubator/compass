package bundle_test

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	name := "foo"
	desc := "bar"

	var nilBundleMode *model.Bundle
	bndlModel := fixBundleModel(name, desc)
	bndlEntity := fixEntityBundle(bundleID, name, desc)

	defAuth, err := json.Marshal(bndlModel.DefaultInstanceAuth)
	require.NoError(t, err)

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Bundle",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, appID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.bundles \(.+\) VALUES \(.+\)$`,
				Args:        fixBundleCreateArgs(string(defAuth), *bndlModel.InstanceAuthRequestInputSchema, bndlModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ModelEntity:         bndlModel,
		DBEntity:            bndlEntity,
		NilModelEntity:      nilBundleMode,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.bundles SET name = ?, description = ?, instance_auth_request_json_schema = ?, default_instance_auth = ?, ord_id = ?, short_description = ?, links = ?, labels = ?, credential_exchange_strategies = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ? WHERE id = ? AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID))

	var nilBundleMode *model.Bundle
	bndl := fixBundleModel("foo", "update")
	entity := fixEntityBundle(bundleID, "foo", "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Bundle",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         updateQuery,
				Args:          []driver.Value{entity.Name, entity.Description, entity.InstanceAuthRequestJSONSchema, entity.DefaultInstanceAuth, entity.OrdID, entity.ShortDescription, entity.Links, entity.Labels, entity.CredentialExchangeStrategies, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ModelEntity:         bndl,
		DBEntity:            entity,
		NilModelEntity:      nilBundleMode,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

/*
func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := fmt.Sprintf(`^DELETE FROM public.bundles WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, bundleID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EntityConverter{}
	pgRepository := bundle.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, bundleID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

*/

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Bundle Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.bundles WHERE id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{bundleID, tenantID},
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
		RepoConstructorFunc: bundle.NewRepository,
		TargetID:            bundleID,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	bndlEntity := fixEntityBundle(bundleID, "foo", "bar")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Bundle",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, name, description, instance_auth_request_json_schema, default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error FROM public.bundles WHERE id = $1 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{bundleID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()).
							AddRow(fixBundleRow(bundleID, "placeholder")...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ExpectedModelEntity: fixBundleModel("foo", "bar"),
		ExpectedDBEntity:    bndlEntity,
		MethodArgs:          []interface{}{tenantID, bundleID},
	}

	suite.Run(t)
}

func TestPgRepository_GetForApplication(t *testing.T) {
	bndlEntity := fixEntityBundle(bundleID, "foo", "bar")

	suite := testdb.RepoGetTestSuite{
		Name: "Get Bundle For Application",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, name, description, instance_auth_request_json_schema, default_instance_auth, ord_id, short_description, links, labels, credential_exchange_strategies, ready, created_at, updated_at, deleted_at, error FROM public.bundles WHERE id = $1 AND app_id = $2 AND (id IN (SELECT id FROM bundles_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{bundleID, appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()).
							AddRow(fixBundleRow(bundleID, "placeholder")...),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixBundleColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundle.NewRepository,
		ExpectedModelEntity: fixBundleModel("foo", "bar"),
		ExpectedDBEntity:    bndlEntity,
		MethodArgs:          []interface{}{tenantID, bundleID, appID},
		MethodName:          "GetForApplication",
	}

	suite.Run(t)
}

/*
func TestPgRepository_ListByApplicationIDs(t *testing.T) {
	// GIVEN
	inputCursor := ""
	applicationIDs := []string{appID, appID2}
	firstBndlID := "111111111-1111-1111-1111-111111111111"
	firstBndlEntity := fixEntityBundle(firstBndlID, "foo", "bar")
	secondBndlID := "222222222-2222-2222-2222-222222222222"
	secondBndlEntity := fixEntityBundle(secondBndlID, "foo", "bar")
	secondBndlEntity.ApplicationID = appID2

	selectQuery := fmt.Sprintf(`^\(SELECT (.+) FROM public\.bundles
		WHERE %s AND app_id = \$2 ORDER BY app_id ASC, id ASC LIMIT \$3 OFFSET \$4\) UNION
		\(SELECT (.+) FROM public\.bundles WHERE %s AND app_id = \$6 ORDER BY app_id ASC, id ASC LIMIT \$7 OFFSET \$8\)`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(5))

	countQuery := fmt.Sprintf(`SELECT app_id AS id, COUNT\(\*\) AS total_count FROM public.bundles WHERE %s GROUP BY app_id ORDER BY app_id ASC`, fixTenantIsolationSubquery())

	t.Run("success when there are no more pages", func(t *testing.T) {
		ExpectedLimit := 3
		ExpectedOffset := 0
		inputPageSize := 3

		totalCountForFirstApp := 1
		totalCountForSecondApp := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(firstBndlID, "placeholder")...).
			AddRow(fixBundleRowWithAppID(secondBndlID, appID2)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID, ExpectedLimit, ExpectedOffset, tenantID, appID2, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(appID, totalCountForFirstApp).
				AddRow(appID2, totalCountForSecondApp))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstBndlEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: firstBndlID}}, nil)
		convMock.On("FromEntity", secondBndlEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: secondBndlID}}, nil)
		pgRepository := bundle.NewRepository(convMock)
		// WHEN
		modelBndls, err := pgRepository.ListByApplicationIDs(ctx, tenantID, applicationIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelBndls, 2)
		assert.Equal(t, firstBndlID, modelBndls[0].Data[0].ID)
		assert.Equal(t, secondBndlID, modelBndls[1].Data[0].ID)
		assert.Equal(t, "", modelBndls[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstApp, modelBndls[0].TotalCount)
		assert.False(t, modelBndls[0].PageInfo.HasNextPage)
		assert.Equal(t, "", modelBndls[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondApp, modelBndls[1].TotalCount)
		assert.False(t, modelBndls[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when there is next page", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		inputPageSize := 1

		totalCountForFirstApp := 10
		totalCountForSecondApp := 10

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(firstBndlID, "placeholder")...).
			AddRow(fixBundleRowWithAppID(secondBndlID, appID2)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID, ExpectedLimit, ExpectedOffset, tenantID, appID2, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(appID, totalCountForFirstApp).
				AddRow(appID2, totalCountForSecondApp))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstBndlEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: firstBndlID}}, nil)
		convMock.On("FromEntity", secondBndlEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: secondBndlID}}, nil)
		pgRepository := bundle.NewRepository(convMock)
		// WHEN
		modelBndls, err := pgRepository.ListByApplicationIDs(ctx, tenantID, applicationIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelBndls, 2)
		assert.Equal(t, firstBndlID, modelBndls[0].Data[0].ID)
		assert.Equal(t, secondBndlID, modelBndls[1].Data[0].ID)
		assert.Equal(t, "", modelBndls[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstApp, modelBndls[0].TotalCount)
		assert.True(t, modelBndls[0].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelBndls[0].PageInfo.EndCursor)
		assert.Equal(t, "", modelBndls[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondApp, modelBndls[1].TotalCount)
		assert.True(t, modelBndls[1].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelBndls[1].PageInfo.EndCursor)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when there is next page and it can be traversed", func(t *testing.T) {
		ExpectedLimit := 1
		ExpectedOffset := 0
		ExpectedSecondOffset := 1
		inputPageSize := 1

		totalCountForFirstApp := 2
		totalCountForSecondApp := 2

		thirdBndlID := "333333333-3333-3333-3333-333333333333"
		thirdBndlEntity := fixEntityBundle(thirdBndlID, "foo", "bar")
		fourthBndlID := "444444444-4444-4444-4444-444444444444"
		fourthBndlEntity := fixEntityBundle(fourthBndlID, "foo", "bar")
		fourthBndlEntity.ApplicationID = appID2

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rowsFirstPage := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(firstBndlID, "placeholder")...).
			AddRow(fixBundleRowWithAppID(secondBndlID, appID2)...)

		rowsSecondPage := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(thirdBndlID, "placeholder")...).
			AddRow(fixBundleRowWithAppID(fourthBndlID, appID2)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID, ExpectedLimit, ExpectedOffset, tenantID, appID2, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rowsFirstPage)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(appID, totalCountForFirstApp).
				AddRow(appID2, totalCountForSecondApp))

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID, ExpectedLimit, ExpectedSecondOffset, tenantID, appID2, ExpectedLimit, ExpectedSecondOffset).
			WillReturnRows(rowsSecondPage)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(appID, totalCountForFirstApp).
				AddRow(appID2, totalCountForSecondApp))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstBndlEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: firstBndlID}}, nil)
		convMock.On("FromEntity", secondBndlEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: secondBndlID}}, nil)
		convMock.On("FromEntity", thirdBndlEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: thirdBndlID}}, nil)
		convMock.On("FromEntity", fourthBndlEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: fourthBndlID}}, nil)
		pgRepository := bundle.NewRepository(convMock)
		// WHEN
		modelBndls, err := pgRepository.ListByApplicationIDs(ctx, tenantID, applicationIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelBndls, 2)
		assert.Equal(t, firstBndlID, modelBndls[0].Data[0].ID)
		assert.Equal(t, secondBndlID, modelBndls[1].Data[0].ID)
		assert.Equal(t, totalCountForFirstApp, modelBndls[0].TotalCount)
		assert.True(t, modelBndls[0].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelBndls[0].PageInfo.EndCursor)
		assert.Equal(t, totalCountForSecondApp, modelBndls[1].TotalCount)
		assert.True(t, modelBndls[1].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelBndls[1].PageInfo.EndCursor)
		endCursor := modelBndls[0].PageInfo.EndCursor

		modelBndlsSecondPage, err := pgRepository.ListByApplicationIDs(ctx, tenantID, applicationIDs, inputPageSize, endCursor)

		require.NoError(t, err)
		require.Len(t, modelBndlsSecondPage, 2)
		assert.Equal(t, thirdBndlID, modelBndlsSecondPage[0].Data[0].ID)
		assert.Equal(t, fourthBndlID, modelBndlsSecondPage[1].Data[0].ID)
		assert.Equal(t, totalCountForFirstApp, modelBndlsSecondPage[0].TotalCount)
		assert.False(t, modelBndlsSecondPage[0].PageInfo.HasNextPage)
		assert.Equal(t, totalCountForSecondApp, modelBndlsSecondPage[1].TotalCount)
		assert.False(t, modelBndlsSecondPage[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns empty page", func(t *testing.T) {
		inputPageSize := 1

		totalCountForFirstApp := 0
		totalCountForSecondApp := 0

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleColumns())

		sqlMock.ExpectQuery(selectQuery).WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(appID, totalCountForFirstApp).
				AddRow(appID2, totalCountForSecondApp))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		pgRepository := bundle.NewRepository(convMock)
		// WHEN
		modelBndls, err := pgRepository.ListByApplicationIDs(ctx, tenantID, applicationIDs, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelBndls[0].Data, 0)
		require.Len(t, modelBndls[1].Data, 0)
		assert.Equal(t, totalCountForFirstApp, modelBndls[0].TotalCount)
		assert.False(t, modelBndls[0].PageInfo.HasNextPage)
		assert.Equal(t, totalCountForSecondApp, modelBndls[1].TotalCount)
		assert.False(t, modelBndls[1].PageInfo.HasNextPage)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		inputPageSize := 1
		ExpectedLimit := 1
		ExpectedOffset := 0

		pgRepository := bundle.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID, ExpectedLimit, ExpectedOffset, tenantID, appID2, ExpectedLimit, ExpectedOffset).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelBndls, err := pgRepository.ListByApplicationIDs(ctx, tenantID, applicationIDs, inputPageSize, inputCursor)

		// then
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelBndls)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error when conversion from entity to model failed", func(t *testing.T) {
		ExpectedLimit := 3
		ExpectedOffset := 0
		inputPageSize := 3
		totalCountForFirstApp := 1
		totalCountForSecondApp := 1
		testErr := errors.New("test error")

		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(firstBndlID, "placeholder")...).
			AddRow(fixBundleRowWithAppID(secondBndlID, appID2)...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID, ExpectedLimit, ExpectedOffset, tenantID, appID2, ExpectedLimit, ExpectedOffset).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).
				AddRow(appID, totalCountForFirstApp).
				AddRow(appID2, totalCountForSecondApp))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstBndlEntity).Return(nil, testErr)
		pgRepository := bundle.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.ListByApplicationIDs(ctx, tenantID, applicationIDs, inputPageSize, inputCursor)
		//THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListByApplicationIDNoPaging(t *testing.T) {
	// GIVEN
	totalCount := 2
	firstBundleID := "111111111-1111-1111-1111-111111111111"
	firstBundleEntity := fixEntityBundle(firstBundleID, "foo", "bar")
	secondBundleID := "222222222-2222-2222-2222-222222222222"
	secondBundleEntity := fixEntityBundle(secondBundleID, "foo", "bar")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.bundles
		WHERE %s AND app_id = \$2`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixBundleColumns()).
			AddRow(fixBundleRow(firstBundleID, "placeholder")...).
			AddRow(fixBundleRow(secondBundleID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstBundleEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: firstBundleID}}, nil)
		convMock.On("FromEntity", secondBundleEntity).Return(&model.Bundle{BaseEntity: &model.BaseEntity{ID: secondBundleID}}, nil)
		pgRepository := bundle.NewRepository(convMock)
		// WHEN
		modelBndl, err := pgRepository.ListByApplicationIDNoPaging(ctx, tenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelBndl, totalCount)
		assert.Equal(t, firstBundleID, modelBndl[0].ID)
		assert.Equal(t, secondBundleID, modelBndl[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
*/
