package api_test

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

/*
func TestPgRepository_GetByID(t *testing.T) {
	// given
	apiDefEntity := fixFullEntityAPIDefinition(apiDefID, "placeholder")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."api_definitions" WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(apiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, apiDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", apiDefEntity).Return(model.APIDefinition{Tenant: tenantID, BaseEntity: &model.BaseEntity{ID: apiDefID}}, nil).Once()
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDef, err := pgRepository.GetByID(ctx, tenantID, apiDefID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, apiDefID, modelAPIDef.ID)
		assert.Equal(t, tenantID, modelAPIDef.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	// GIVEN
	totalCount := 2
	firstAPIDefID := "111111111-1111-1111-1111-111111111111"
	firstAPIDefEntity := fixFullEntityAPIDefinition(firstAPIDefID, "placeholder")
	secondAPIDefID := "222222222-2222-2222-2222-222222222222"
	secondAPIDefEntity := fixFullEntityAPIDefinition(secondAPIDefID, "placeholder")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."api_definitions"
		WHERE %s AND app_id = \$2`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstAPIDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondAPIDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstAPIDefID}}, nil)
		convMock.On("FromEntity", secondAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondAPIDefID}}, nil)
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDef, err := pgRepository.ListByApplicationID(ctx, tenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDef, totalCount)
		assert.Equal(t, firstAPIDefID, modelAPIDef[0].ID)
		assert.Equal(t, secondAPIDefID, modelAPIDef[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListAllForBundle(t *testing.T) {
	// GIVEN
	inputPageSize := 3
	inputCursor := ""

	firstBndlID := "111111111-1111-1111-1111-111111111111"
	secondBndlID := "222222222-2222-2222-2222-222222222222"
	bundleIDs := []string{firstBndlID, secondBndlID}

	firstAPIDefID := "111111111-1111-1111-1111-111111111111"
	firstAPIDefEntity := fixFullEntityAPIDefinition(firstAPIDefID, "placeholder")
	secondAPIDefID := "222222222-2222-2222-2222-222222222222"
	secondAPIDefEntity := fixFullEntityAPIDefinition(secondAPIDefID, "placeholder")

	firstBundleRef := fixModelBundleReference(firstBndlID, firstAPIDefID)
	secondBundleRef := fixModelBundleReference(secondBndlID, secondAPIDefID)
	bundleRefs := []*model.BundleReference{firstBundleRef, secondBundleRef}

	totalCounts := map[string]int{firstBndlID: 1, secondBndlID: 1}

	selectQuery := fmt.Sprintf(`^SELECT (.+)
		FROM "public"."api_definitions"
		WHERE %s AND id IN \(\$2, \$3\)`, fixTenantIsolationSubquery())

	t.Run("success when there are no more pages", func(t *testing.T) {
		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstAPIDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondAPIDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstAPIDefID, secondAPIDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstAPIDefID}})
		convMock.On("FromEntity", secondAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondAPIDefID}})
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDefs, 2)
		assert.Equal(t, firstAPIDefID, modelAPIDefs[0].Data[0].ID)
		assert.Equal(t, secondAPIDefID, modelAPIDefs[1].Data[0].ID)
		assert.Equal(t, "", modelAPIDefs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelAPIDefs[0].TotalCount)
		assert.False(t, modelAPIDefs[0].PageInfo.HasNextPage)
		assert.Equal(t, "", modelAPIDefs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelAPIDefs[1].TotalCount)
		assert.False(t, modelAPIDefs[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when there is next page", func(t *testing.T) {
		totalCountForFirstBundle := 10
		totalCountForSecondBundle := 10
		totalCounts[firstBndlID] = 10
		totalCounts[secondBndlID] = 10

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstAPIDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondAPIDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstAPIDefID, secondAPIDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstAPIDefID}})
		convMock.On("FromEntity", secondAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondAPIDefID}})
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDefs, 2)
		assert.Equal(t, firstAPIDefID, modelAPIDefs[0].Data[0].ID)
		assert.Equal(t, secondAPIDefID, modelAPIDefs[1].Data[0].ID)
		assert.Equal(t, "", modelAPIDefs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelAPIDefs[0].TotalCount)
		assert.True(t, modelAPIDefs[0].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelAPIDefs[0].PageInfo.EndCursor)
		assert.Equal(t, "", modelAPIDefs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelAPIDefs[1].TotalCount)
		assert.True(t, modelAPIDefs[1].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelAPIDefs[1].PageInfo.EndCursor)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("success when there is next page and it can be traversed", func(t *testing.T) {
		totalCountForFirstBundle := 2
		totalCountForSecondBundle := 2
		totalCounts[firstBndlID] = 2
		totalCounts[secondBndlID] = 2

		thirdAPIDefID := "333333333-3333-3333-3333-333333333333"
		thirdAPIDefEntity := fixFullEntityAPIDefinition(thirdAPIDefID, "placeholder")
		fourthAPIDefID := "444444444-4444-4444-4444-444444444444"
		fourthAPIDefEntity := fixFullEntityAPIDefinition(fourthAPIDefID, "placeholder")

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rowsFirstPage := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstAPIDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondAPIDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstAPIDefID, secondAPIDefID).
			WillReturnRows(rowsFirstPage)

		rowsSecondPage := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(thirdAPIDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(fourthAPIDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, thirdAPIDefID, fourthAPIDefID).
			WillReturnRows(rowsSecondPage)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstAPIDefID}})
		convMock.On("FromEntity", secondAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondAPIDefID}})
		convMock.On("FromEntity", thirdAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: thirdAPIDefID}})
		convMock.On("FromEntity", fourthAPIDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: fourthAPIDefID}})
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDefs, 2)
		assert.Equal(t, firstAPIDefID, modelAPIDefs[0].Data[0].ID)
		assert.Equal(t, secondAPIDefID, modelAPIDefs[1].Data[0].ID)
		assert.Equal(t, "", modelAPIDefs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelAPIDefs[0].TotalCount)
		assert.True(t, modelAPIDefs[0].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelAPIDefs[0].PageInfo.EndCursor)
		assert.Equal(t, "", modelAPIDefs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelAPIDefs[1].TotalCount)
		assert.True(t, modelAPIDefs[1].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelAPIDefs[1].PageInfo.EndCursor)
		endCursor := modelAPIDefs[0].PageInfo.EndCursor

		thirdBundleRef := fixModelBundleReference(firstBndlID, thirdAPIDefID)
		fourthBundleRef := fixModelBundleReference(secondBndlID, fourthAPIDefID)
		bundleRefsSecondPage := []*model.BundleReference{thirdBundleRef, fourthBundleRef}

		modelAPIDefsSecondPage, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefsSecondPage, totalCounts, inputPageSize, endCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDefsSecondPage, 2)
		assert.Equal(t, thirdAPIDefID, modelAPIDefsSecondPage[0].Data[0].ID)
		assert.Equal(t, fourthAPIDefID, modelAPIDefsSecondPage[1].Data[0].ID)
		assert.Equal(t, totalCountForFirstBundle, modelAPIDefsSecondPage[0].TotalCount)
		assert.False(t, modelAPIDefsSecondPage[0].PageInfo.HasNextPage)
		assert.Equal(t, totalCountForSecondBundle, modelAPIDefsSecondPage[1].TotalCount)
		assert.False(t, modelAPIDefsSecondPage[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns empty page", func(t *testing.T) {
		totalCountForFirstBundle := 0
		totalCountForSecondBundle := 0
		totalCounts[firstBndlID] = 0
		totalCounts[secondBndlID] = 0

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns())

		sqlMock.ExpectQuery(selectQuery).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN

		require.NoError(t, err)
		require.Len(t, modelAPIDefs[0].Data, 0)
		require.Len(t, modelAPIDefs[1].Data, 0)
		assert.Equal(t, totalCountForFirstBundle, modelAPIDefs[0].TotalCount)
		assert.False(t, modelAPIDefs[0].PageInfo.HasNextPage)
		assert.Equal(t, totalCountForSecondBundle, modelAPIDefs[1].TotalCount)
		assert.False(t, modelAPIDefs[1].PageInfo.HasNextPage)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		pgRepository := api.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstAPIDefID, secondAPIDefID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelAPIDefs, err := pgRepository.ListByBundleIDs(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)

		// then
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelAPIDefs)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
*/
func TestPgRepository_Create(t *testing.T) {
	var nilApiDefModel *model.APIDefinition
	apiDefModel, _, _ := fixFullAPIDefinitionModel("placeholder")
	apiDefEntity := fixFullEntityAPIDefinition(apiDefID, "placeholder")

	suite := testdb.RepoCreateTestSuite{
		Name: "Create API",
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
				Query:       `^INSERT INTO "public"."api_definitions" \(.+\) VALUES \(.+\)$`,
				Args:        fixAPICreateArgs(apiDefID, &apiDefModel),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc:       api.NewRepository,
		ModelEntity:               &apiDefModel,
		DBEntity:                  &apiDefEntity,
		NilModelEntity:            nilApiDefModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

/*
func TestPgRepository_CreateMany(t *testing.T) {
	insertQuery := `^INSERT INTO "public"."api_definitions" (.+) VALUES (.+)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		first, _, _ := fixFullAPIDefinitionModel("first")
		second, _, _ := fixFullAPIDefinitionModel("second")
		third, _, _ := fixFullAPIDefinitionModel("third")
		items := []*model.APIDefinition{&first, &second, &third}

		convMock := &automock.APIDefinitionConverter{}
		for _, item := range items {
			ent := fixFullEntityAPIDefinition(item.ID, item.Name)
			convMock.On("ToEntity", *item).Return(&ent, nil).Once()
			sqlMock.ExpectExec(insertQuery).
				WithArgs(fixAPICreateArgs(item.ID, item)...).
				WillReturnResult(sqlmock.NewResult(-1, 1))
		}
		pgRepository := api.NewRepository(convMock)
		//WHEN
		err := pgRepository.CreateMany(ctx, items)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
*/

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(fmt.Sprintf(`UPDATE "public"."api_definitions" SET package_id = ?, name = ?, description = ?, group_name = ?, ord_id = ?,
		short_description = ?, system_instance_aware = ?, api_protocol = ?, tags = ?, countries = ?, links = ?, api_resource_links = ?, release_status = ?,
		sunset_date = ?, changelog_entries = ?, labels = ?, visibility = ?, disabled = ?, part_of_products = ?, line_of_business = ?,
		industry = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, implementation_standard = ?, custom_implementation_standard = ?, custom_implementation_standard_description = ?, target_urls = ?, extensible = ?, successors = ?, resource_hash = ?
		WHERE id = ? AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID))

	var nilApiDefModel *model.APIDefinition
	apiModel, _, _ := fixFullAPIDefinitionModel("update")
	entity := fixFullEntityAPIDefinition(apiDefID, "update")
	entity.UpdatedAt = &fixedTimestamp
	entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update API",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query: updateQuery,
				Args: []driver.Value{entity.PackageID, entity.Name, entity.Description, entity.Group,
					entity.OrdID, entity.ShortDescription, entity.SystemInstanceAware, entity.APIProtocol, entity.Tags, entity.Countries,
					entity.Links, entity.APIResourceLinks, entity.ReleaseStatus, entity.SunsetDate, entity.ChangeLogEntries, entity.Labels, entity.Visibility,
					entity.Disabled, entity.PartOfProducts, entity.LineOfBusiness, entity.Industry, entity.Version.Value, entity.Version.Deprecated, entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.ImplementationStandard, entity.CustomImplementationStandard, entity.CustomImplementationStandardDescription, entity.TargetURLs, entity.Extensible, entity.Successors, entity.ResourceHash, entity.ID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc:       api.NewRepository,
		ModelEntity:               &apiModel,
		DBEntity:                  &entity,
		NilModelEntity:            nilApiDefModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

/*
func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := fmt.Sprintf(`^DELETE FROM "public"."api_definitions" WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, apiDefID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.APIDefinitionConverter{}
	pgRepository := api.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, apiDefID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_DeleteAllByBundleID(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := fmt.Sprintf(`DELETE FROM "public"."api_definitions"
		WHERE %s AND id IN \(SELECT (.+) FROM public\.bundle_references WHERE %s AND bundle_id = \$3 AND api_def_id IS NOT NULL\)`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(2))

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, tenantID, bundleID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.APIDefinitionConverter{}
	pgRepository := api.NewRepository(convMock)
	//WHEN
	err := pgRepository.DeleteAllByBundleID(ctx, tenantID, bundleID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

*/
func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "API Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM "public"."api_definitions" WHERE id = $1 AND (id IN (SELECT id FROM api_definitions_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{apiDefID, tenantID},
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
			return &automock.APIDefinitionConverter{}
		},
		RepoConstructorFunc: api.NewRepository,
		TargetID:            apiDefID,
		TenantID:            tenantID,
	}

	suite.Run(t)
}
