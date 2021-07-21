package api_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		modelApiDef, err := pgRepository.GetByID(ctx, tenantID, apiDefID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, apiDefID, modelApiDef.ID)
		assert.Equal(t, tenantID, modelApiDef.Tenant)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

}

func TestPgRepository_ListForBundle(t *testing.T) {
	// GIVEN
	ExpectedLimit := 3
	ExpectedOffset := 0

	inputPageSize := 3
	inputCursor := ""
	totalCount := 2
	firstApiDefID := "111111111-1111-1111-1111-111111111111"
	firstApiDefEntity := fixFullEntityAPIDefinition(firstApiDefID, "placeholder")
	secondApiDefID := "222222222-2222-2222-2222-222222222222"
	secondApiDefEntity := fixFullEntityAPIDefinition(secondApiDefID, "placeholder")

	selectQuery := fmt.Sprintf(`SELECT (.+) FROM "public"."api_definitions"
		WHERE %s AND id IN \(SELECT (.+) FROM public\.bundle_references WHERE %s AND bundle_id = \$3 AND api_def_id IS NOT NULL\) 
		ORDER BY id LIMIT %d OFFSET %d`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(2), ExpectedLimit, ExpectedOffset)

	countQuery := fmt.Sprintf(`SELECT COUNT\(\*\) FROM "public"."api_definitions"
		WHERE %s AND id IN \(SELECT (.+) FROM public\.bundle_references WHERE %s AND bundle_id = \$3 AND api_def_id IS NOT NULL\)`, fixTenantIsolationSubqueryWithArg(1), fixTenantIsolationSubqueryWithArg(2))

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstApiDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, tenantID, bundleID).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(tenantID, tenantID, bundleID).
			WillReturnRows(testdb.RowCount(2))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstApiDefID}}, nil)
		convMock.On("FromEntity", secondApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondApiDefID}}, nil)
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDef, err := pgRepository.ListForBundle(ctx, tenantID, bundleID, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDef.Data, 2)
		assert.Equal(t, firstApiDefID, modelAPIDef.Data[0].ID)
		assert.Equal(t, secondApiDefID, modelAPIDef.Data[1].ID)
		assert.Equal(t, "", modelAPIDef.PageInfo.StartCursor)
		assert.Equal(t, totalCount, modelAPIDef.TotalCount)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	// GIVEN
	totalCount := 2
	firstApiDefID := "111111111-1111-1111-1111-111111111111"
	firstApiDefEntity := fixFullEntityAPIDefinition(firstApiDefID, "placeholder")
	secondApiDefID := "222222222-2222-2222-2222-222222222222"
	secondApiDefEntity := fixFullEntityAPIDefinition(secondApiDefID, "placeholder")

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM "public"."api_definitions" 
		WHERE %s AND app_id = \$2`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstApiDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstApiDefID}}, nil)
		convMock.On("FromEntity", secondApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondApiDefID}}, nil)
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDef, err := pgRepository.ListByApplicationID(ctx, tenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDef, totalCount)
		assert.Equal(t, firstApiDefID, modelAPIDef[0].ID)
		assert.Equal(t, secondApiDefID, modelAPIDef[1].ID)
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

	firstApiDefID := "111111111-1111-1111-1111-111111111111"
	firstApiDefEntity := fixFullEntityAPIDefinition(firstApiDefID, "placeholder")
	secondApiDefID := "222222222-2222-2222-2222-222222222222"
	secondApiDefEntity := fixFullEntityAPIDefinition(secondApiDefID, "placeholder")

	firstBundleRef := fixModelBundleReference(firstBndlID, firstApiDefID)
	secondBundleRef := fixModelBundleReference(secondBndlID, secondApiDefID)
	bundleRefs := []*model.BundleReference{firstBundleRef, secondBundleRef}

	totalCounts := map[string]int{firstBndlID: 1, secondBndlID: 1}

	selectQuery := fmt.Sprintf(`SELECT (.+) 
		FROM "public"."api_definitions" 
		WHERE %s AND id IN \(\$2, \$3\)`, fixTenantIsolationSubquery())

	t.Run("success when there are no more pages", func(t *testing.T) {
		totalCountForFirstBundle := 1
		totalCountForSecondBundle := 1

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstApiDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstApiDefID, secondApiDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstApiDefID}})
		convMock.On("FromEntity", secondApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondApiDefID}})
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDefs, err := pgRepository.ListAllForBundle(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDefs, 2)
		assert.Equal(t, firstApiDefID, modelAPIDefs[0].Data[0].ID)
		assert.Equal(t, secondApiDefID, modelAPIDefs[1].Data[0].ID)
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
			AddRow(fixAPIDefinitionRow(firstApiDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstApiDefID, secondApiDefID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstApiDefID}})
		convMock.On("FromEntity", secondApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondApiDefID}})
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDefs, err := pgRepository.ListAllForBundle(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDefs, 2)
		assert.Equal(t, firstApiDefID, modelAPIDefs[0].Data[0].ID)
		assert.Equal(t, secondApiDefID, modelAPIDefs[1].Data[0].ID)
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

		thirdApiDefID := "333333333-3333-3333-3333-333333333333"
		thirdApiDefEntity := fixFullEntityAPIDefinition(thirdApiDefID, "placeholder")
		fourthApiDefID := "444444444-4444-4444-4444-444444444444"
		fourthApiDefEntity := fixFullEntityAPIDefinition(fourthApiDefID, "placeholder")

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rowsFirstPage := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(firstApiDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(secondApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, firstApiDefID, secondApiDefID).
			WillReturnRows(rowsFirstPage)

		rowsSecondPage := sqlmock.NewRows(fixAPIDefinitionColumns()).
			AddRow(fixAPIDefinitionRow(thirdApiDefID, "placeholder")...).
			AddRow(fixAPIDefinitionRow(fourthApiDefID, "placeholder")...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, thirdApiDefID, fourthApiDefID).
			WillReturnRows(rowsSecondPage)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		convMock.On("FromEntity", firstApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: firstApiDefID}})
		convMock.On("FromEntity", secondApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: secondApiDefID}})
		convMock.On("FromEntity", thirdApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: thirdApiDefID}})
		convMock.On("FromEntity", fourthApiDefEntity).Return(model.APIDefinition{BaseEntity: &model.BaseEntity{ID: fourthApiDefID}})
		pgRepository := api.NewRepository(convMock)
		// WHEN
		modelAPIDefs, err := pgRepository.ListAllForBundle(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDefs, 2)
		assert.Equal(t, firstApiDefID, modelAPIDefs[0].Data[0].ID)
		assert.Equal(t, secondApiDefID, modelAPIDefs[1].Data[0].ID)
		assert.Equal(t, "", modelAPIDefs[0].PageInfo.StartCursor)
		assert.Equal(t, totalCountForFirstBundle, modelAPIDefs[0].TotalCount)
		assert.True(t, modelAPIDefs[0].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelAPIDefs[0].PageInfo.EndCursor)
		assert.Equal(t, "", modelAPIDefs[1].PageInfo.StartCursor)
		assert.Equal(t, totalCountForSecondBundle, modelAPIDefs[1].TotalCount)
		assert.True(t, modelAPIDefs[1].PageInfo.HasNextPage)
		assert.NotEmpty(t, modelAPIDefs[1].PageInfo.EndCursor)
		endCursor := modelAPIDefs[0].PageInfo.EndCursor

		thirdBundleRef := fixModelBundleReference(firstBndlID, thirdApiDefID)
		fourthBundleRef := fixModelBundleReference(secondBndlID, fourthApiDefID)
		bundleRefsSecondPage := []*model.BundleReference{thirdBundleRef, fourthBundleRef}

		modelAPIDefsSecondPage, err := pgRepository.ListAllForBundle(ctx, tenantID, bundleIDs, bundleRefsSecondPage, totalCounts, inputPageSize, endCursor)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelAPIDefsSecondPage, 2)
		assert.Equal(t, thirdApiDefID, modelAPIDefsSecondPage[0].Data[0].ID)
		assert.Equal(t, fourthApiDefID, modelAPIDefsSecondPage[1].Data[0].ID)
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
		modelAPIDefs, err := pgRepository.ListAllForBundle(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)
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
			WithArgs(tenantID, firstApiDefID, secondApiDefID).
			WillReturnError(testError)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelAPIDefs, err := pgRepository.ListAllForBundle(ctx, tenantID, bundleIDs, bundleRefs, totalCounts, inputPageSize, inputCursor)

		// then
		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelAPIDefs)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	apiDefModel, _, _ := fixFullAPIDefinitionModel("placeholder")
	apiDefEntity := fixFullEntityAPIDefinition(apiDefID, "placeholder")
	insertQuery := `^INSERT INTO "public"."api_definitions" \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixAPICreateArgs(apiDefID, &apiDefModel)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.APIDefinitionConverter{}
		convMock.On("ToEntity", apiDefModel).Return(&apiDefEntity, nil).Once()
		pgRepository := api.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, &apiDefModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.APIDefinitionConverter{}
		pgRepository := api.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item cannot be nil")
		convMock.AssertExpectations(t)
	})
}

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

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(fmt.Sprintf(`UPDATE "public"."api_definitions" SET package_id = ?, name = ?, description = ?, group_name = ?, ord_id = ?,
		short_description = ?, system_instance_aware = ?, api_protocol = ?, tags = ?, countries = ?, links = ?, api_resource_links = ?, release_status = ?,
		sunset_date = ?, changelog_entries = ?, labels = ?, visibility = ?, disabled = ?, part_of_products = ?, line_of_business = ?,
		industry = ?, version_value = ?, version_deprecated = ?, version_deprecated_since = ?, version_for_removal = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, implementation_standard = ?, custom_implementation_standard = ?, custom_implementation_standard_description = ?, target_urls = ?, extensible = ?, successors = ?, resource_hash = ? WHERE %s AND id = ?`, fixUpdateTenantIsolationSubquery()))

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		apiModel, _, _ := fixFullAPIDefinitionModel("update")
		entity := fixFullEntityAPIDefinition(apiDefID, "update")
		entity.UpdatedAt = &fixedTimestamp
		entity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

		convMock := &automock.APIDefinitionConverter{}
		convMock.On("ToEntity", apiModel).Return(&entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(entity.PackageID, entity.Name, entity.Description, entity.Group,
				entity.OrdID, entity.ShortDescription, entity.SystemInstanceAware, entity.ApiProtocol, entity.Tags, entity.Countries,
				entity.Links, entity.APIResourceLinks, entity.ReleaseStatus, entity.SunsetDate, entity.ChangeLogEntries, entity.Labels, entity.Visibility,
				entity.Disabled, entity.PartOfProducts, entity.LineOfBusiness, entity.Industry, entity.Version.Value, entity.Version.Deprecated, entity.Version.DeprecatedSince, entity.Version.ForRemoval, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.ImplementationStandard, entity.CustomImplementationStandard, entity.CustomImplementationStandardDescription, entity.TargetURLs, entity.Extensible, entity.Successors, entity.ResourceHash, tenantID, entity.ID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := api.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, &apiModel)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.APIDefinitionConverter{}
		pgRepository := api.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, nil)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "item cannot be nil")
		convMock.AssertExpectations(t)
	})
}

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

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := fmt.Sprintf(`SELECT 1 FROM "public"."api_definitions" WHERE %s AND id = \$2`, fixTenantIsolationSubquery())

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, apiDefID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.APIDefinitionConverter{}
	pgRepository := api.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, apiDefID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}
